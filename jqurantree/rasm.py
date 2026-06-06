"""
Rasm Automaton — context-aware cost function for Uthmani vs Modern Arabic alignment.

Encodes the systematic orthographic rules of the Uthmani rasm (script) so that
Needleman-Wunsch alignment treats them as zero-cost substitutions rather than
genuine differences.

Rules encoded:
  1. Missing alif after ع/ح/ه before لام (e.g., علمين → عالمين)
  2. Dagger alif (ٰ) → silent (cost 0)
  3. Waw + dagger alif → Alif (الصلوٰة → الصلاة)
  4. Alif Wasla → Alif (word-initial connecting alif)
  5. Alif Maksura → Ya (ى → ي)
  6. Ta Marbuta → Ha (ة → ه) — only in certain positions
  7. Madda (آ) → bare alif (ا)
  8. Small waw/ya marks → silent
"""

from __future__ import annotations

# Uthmani → Modern equivalence pairs (context-free substitutions)
UTHMANI_TO_MODERN = str.maketrans({
    "\u0671": "\u0627",  # Alif Wasla → Alif
    "\u0670": "",        # Dagger Alif → silent
    "\u0649": "\u064A",  # Alif Maksura → Ya
    "\u06E1": "",        # Small high dotless head of kha → silent
    "\u06E5": "",        # Small Waw → silent
    "\u06E6": "",        # Small Ya → silent
    "\u06E8": "",        # Small high noon → silent
    "\u06ED": "",        # Small low meem → silent
})

# Phonetic diacritics to strip
_PHONETIC_DIACRITICS = frozenset({
    0x064B, 0x064C, 0x064D, 0x064E, 0x064F, 0x0650, 0x0651, 0x0652,
})

# RTL/LTR markers to strip
_LAYOUT_MARKS = frozenset({0x200F, 0x200E})

# All Quranic marks to strip (includes Indo-Pak specific marks)
_ALL_MARKS = frozenset({
    0x064B, 0x064C, 0x064D, 0x064E, 0x064F, 0x0650, 0x0651, 0x0652,
    0x0653, 0x0654, 0x0655, 0x0656, 0x0657, 0x0658,
    0x06DC, 0x06DF, 0x06E0, 0x06E1, 0x06E2, 0x06E3,
    0x06E5, 0x06E6, 0x06E8, 0x06EA, 0x06EB, 0x06EC, 0x06ED,
    0x0615, 0x06D6, 0x06D7, 0x06D8, 0x06D9, 0x06DA, 0x06DB,
})


def normalize_rasm(text: str) -> str:
    """Normalize Uthmani rasm to a canonical representation for comparison."""
    import unicodedata
    text = unicodedata.normalize("NFD", text)
    text = text.translate(UTHMANI_TO_MODERN)
    text = "".join(ch for ch in text if ord(ch) not in _ALL_MARKS)
    text = "".join(ch for ch in text if ord(ch) not in _LAYOUT_MARKS)
    text = text.replace("\u0629", "\u0647")   # Ta Marbuta → Ha
    text = text.replace("\u0622", "\u0627")   # Alif Madda → Alif
    return text


def rasm_cost(a_char: str, b_char: str, prev_a: str = "", prev_b: str = "") -> float:
    """Context-aware substitution cost for Needleman-Wunsch alignment."""
    if a_char == b_char:
        return 0.0
    if ord(a_char) in _ALL_MARKS or ord(b_char) in _ALL_MARKS:
        return 0.0
    if ord(a_char) in _LAYOUT_MARKS or ord(b_char) in _LAYOUT_MARKS:
        return 0.0
    if a_char == "\u0670" or b_char == "\u0670":
        return 0.0
    
    # Both are diacritics → no cost (will be stripped anyway)
    if ord(a_char) in _ALL_MARKS and ord(b_char) in _ALL_MARKS:
        return 0.0
    
    # Diacritic vs nothing → no cost
    if ord(a_char) in _ALL_MARKS or ord(b_char) in _ALL_MARKS:
        return 0.0
    
    # Dagger alif anywhere → free
    if a_char == "\u0670" or b_char == "\u0670":
        return 0.0
    
    # Wasla ↔ Alif equivalence
    if {a_char, b_char} == {"\u0671", "\u0627"}:
        return 0.0
    
    # Alif Maksura ↔ Ya
    if {a_char, b_char} == {"\u0649", "\u064A"}:
        return 0.0
    
    # Ta Marbuta ↔ Ha
    if {a_char, b_char} == {"\u0629", "\u0647"}:
        return 0.0
    
    # Madda ↔ Alif
    if {a_char, b_char} == {"\u0622", "\u0627"}:
        return 0.0
    
    # Hamza-bearing alif variants → alif
    if a_char in ("\u0623", "\u0625") and b_char == "\u0627":
        return 0.0
    if b_char in ("\u0623", "\u0625") and a_char == "\u0627":
        return 0.0

    # Tatweel → free (kashida is purely cosmetic)
    if a_char == "\u0640" or b_char == "\u0640":
        return 0.0

    # Hamza ↔ Alif (standalone hamza in Uthmani maps to alif in Simple)
    if {a_char, b_char} == {"\u0621", "\u0627"}:
        return 0.0

    # Hamza ↔ Ya in word-final position (رءا→راي)
    if {a_char, b_char} == {"\u0621", "\u064A"}:
        return 0.15

    # Final Alif ↔ Ya interchange (طغا→طغي، لدا→لدي)
    if {a_char, b_char} == {"\u0627", "\u064A"} and prev_a and prev_b:
        return 0.15

    # Alif ↔ Waw in certain positions (الصلوه→الصلاه)
    if {a_char, b_char} == {"\u0627", "\u0648"}:
        return 0.15
    
    # Waw + dagger alif → Alif pattern
    # "وٰ" → "ا" (the waw+dagger in Uthmani = alif in modern)
    # This is context-dependent: need to see if the char before was waw
    if prev_a == "\u0648" and a_char == "\u0670" and b_char == "\u0627":
        return 0.0
    
    return 1.0


def needleman_wunsch(a: str, b: str, gap_penalty: float = 0.5) -> tuple[float, str, str]:
    """
    Needleman-Wunsch global alignment with custom cost function.
    
    Returns: (score, aligned_a, aligned_b) where score is the total alignment cost
    and aligned strings include gaps (-) for insertions/deletions.
    """
    m, n = len(a), len(b)
    D = [[0.0] * (n + 1) for _ in range(m + 1)]
    trace = [[""] * (n + 1) for _ in range(m + 1)]
    
    for i in range(1, m + 1):
        D[i][0] = D[i-1][0] + gap_penalty
        trace[i][0] = "up"
    for j in range(1, n + 1):
        D[0][j] = D[0][j-1] + gap_penalty
        trace[0][j] = "left"
    
    for i in range(1, m + 1):
        for j in range(1, n + 1):
            prev_a = a[i-2] if i > 1 else ""
            prev_b = b[j-2] if j > 1 else ""
            sub_cost = rasm_cost(a[i-1], b[j-1], prev_a, prev_b)
            
            diag = D[i-1][j-1] + sub_cost
            up = D[i-1][j] + gap_penalty
            left = D[i][j-1] + gap_penalty
            
            if diag <= up and diag <= left:
                D[i][j] = diag
                trace[i][j] = "diag"
            elif up <= left:
                D[i][j] = up
                trace[i][j] = "up"
            else:
                D[i][j] = left
                trace[i][j] = "left"
    
    # Backtrack
    i, j = m, n
    aligned_a, aligned_b = [], []
    while i > 0 or j > 0:
        if i > 0 and j > 0 and trace[i][j] == "diag":
            aligned_a.append(a[i-1])
            aligned_b.append(b[j-1])
            i -= 1; j -= 1
        elif i > 0 and trace[i][j] == "up":
            aligned_a.append(a[i-1])
            aligned_b.append("-")
            i -= 1
        else:
            aligned_a.append("-")
            aligned_b.append(b[j-1])
            j -= 1
    
    aligned_a.reverse(); aligned_b.reverse()
    return D[m][n], "".join(aligned_a), "".join(aligned_b)


def word_cost(a: str, b: str) -> float:
    """Cost of aligning two words using NW global alignment."""
    score, _, _ = needleman_wunsch(a, b)
    max_len = max(len(a), len(b))
    if max_len == 0:
        return 0.0
    return score / max_len
