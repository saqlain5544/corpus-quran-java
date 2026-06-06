"""
Word alignment engine — maps any intra-Hafs variant's word positions 
to the reference Uthmani Hafs positions used by MASAQ morphology data.

Enables unified queries: Quran("indopak").roots(2, 255) works correctly
even though MASAQ is keyed to Uthmani Hafs token positions.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from functools import lru_cache
from .rasm import normalize_rasm
from .variants import load_variant


@dataclass(slots=True)
class WordMap:
    """Maps variant word indices → reference (Uthmani) word indices per verse."""
    ref_to_variant: dict[int, int] = field(default_factory=dict)
    variant_to_ref: dict[int, int] = field(default_factory=dict)


class Alignment:
    """Word-to-word alignment between any variant and the reference Uthmani Hafs."""

    def __init__(self, variant_key: str = "uthmani"):
        self._key = variant_key
        self._ref = load_variant("uthmani")
        self._var = load_variant(variant_key)
        self._cache: dict[tuple[int, int], WordMap] = {}

    def map(self, sura: int, verse: int) -> WordMap:
        key = (sura, verse)
        if key in self._cache:
            return self._cache[key]

        a_words, b_words = self._tokenize(sura, verse)

        wm = WordMap()
        # Direct 1:1 mapping for matching word counts
        if len(a_words) == len(b_words):
            for i in range(len(a_words)):
                wm.ref_to_variant[i + 1] = i + 1
                wm.variant_to_ref[i + 1] = i + 1
            self._cache[key] = wm
            return wm

        # NW alignment for different word counts
        self._align(a_words, b_words, wm)
        self._cache[key] = wm
        return wm

    def ref_position(self, sura: int, verse: int, variant_word: int) -> int:
        """Given a variant word index, return the reference (Uthmani) word index."""
        wm = self.map(sura, verse)
        return wm.variant_to_ref.get(variant_word, variant_word)

    def variant_position(self, sura: int, verse: int, ref_word: int) -> int:
        """Given a reference word index, return the variant's word index."""
        wm = self.map(sura, verse)
        return wm.ref_to_variant.get(ref_word, ref_word)

    def _tokenize(self, sura: int, verse: int) -> tuple[list[str], list[str]]:
        su_uth = self._ref["suras"][sura - 1]
        su_var = self._var["suras"][sura - 1]
        a_text = su_uth["ayas"][verse - 1]["text"]
        b_text = su_var["ayas"][verse - 1]["text"]
        a_words = [normalize_rasm(w) for w in a_text.split() if normalize_rasm(w)]
        b_words = [normalize_rasm(w) for w in b_text.split() if normalize_rasm(w)]
        return a_words, b_words

    def _align(self, a: list[str], b: list[str], wm: WordMap) -> None:
        m, n = len(a), len(b)
        D = [[0.0] * (n + 1) for _ in range(m + 1)]
        trace = [[""] * (n + 1) for _ in range(m + 1)]

        for i in range(1, m + 1):
            D[i][0] = D[i - 1][0] + 1.0
            trace[i][0] = "up"
        for j in range(1, n + 1):
            D[0][j] = D[0][j - 1] + 1.0
            trace[0][j] = "left"

        for i in range(1, m + 1):
            for j in range(1, n + 1):
                cost = 0.0 if a[i - 1] == b[j - 1] else self._word_sub_cost(a[i - 1], b[j - 1])
                diag = D[i - 1][j - 1] + cost
                up = D[i - 1][j] + 0.5
                left = D[i][j - 1] + 0.5
                if diag <= up and diag <= left:
                    D[i][j] = diag
                    trace[i][j] = "diag"
                elif up <= left:
                    D[i][j] = up
                    trace[i][j] = "up"
                else:
                    D[i][j] = left
                    trace[i][j] = "left"

        i, j = m, n
        ref_idx, var_idx = m, n
        while i > 0 or j > 0:
            if i > 0 and j > 0 and trace[i][j] == "diag":
                wm.ref_to_variant[ref_idx] = var_idx
                wm.variant_to_ref[var_idx] = ref_idx
                ref_idx -= 1
                var_idx -= 1
                i -= 1
                j -= 1
            elif i > 0 and trace[i][j] == "up":
                ref_idx -= 1
                i -= 1
            else:
                var_idx -= 1
                j -= 1

    @staticmethod
    def _word_sub_cost(a: str, b: str) -> float:
        if a == b:
            return 0.0
        if a and b and (a in b or b in a):
            return 0.1
        if a and b and a[0] == b[0]:
            return 0.3
        return 1.0
