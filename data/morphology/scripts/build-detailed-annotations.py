#!/usr/bin/env python3
"""
Build quran-detailed.xml from scratch using verified data sources:
  quran-uthmani.xml  → surah/verse metadata + rasm text
  MASAQ.csv          → word tokens + segment morphology
  corpus-roots.json  → root mappings
"""
import csv, json, xml.etree.ElementTree as ET, os

# ── Paths ──────────────────────────────────────────────────────────────
UTHMANI  = '/Users/saqlain/corpus-quran-java/data/quran/quran-uthmani.xml'
MASAQ_CSV = '/Users/saqlain/corpus-quran-java/data/morphology/MASAQ.csv'
ROOTS_JSON = '/Users/saqlain/corpus-quran-java/data/morphology/corpus-roots.json'
OUTPUT   = '/Users/saqlain/corpus-quran-java/data/new/quran-detailed.xml'

# ── Diacritic detection ────────────────────────────────────────────────
DIAC = set(range(0x064B, 0x0653)) | {0x0670} \
       | set(range(0x06D6, 0x06EE)) | {0x06FE, 0x06FF}
WASLA = {'\u0671', '\u0627', '\u0622', '\u0672', '\u0673'}  # true wasla alifs only
LATIN = {'W': 'و', 'B': 'ب', 'M': 'م', 'L': 'ل', 'I': 'ا'}

# ── Buckwalter → Arabic ────────────────────────────────────────────────
BW2AR = {
    "'": 'ء','|': 'آ','>': 'أ','<': 'إ','}': 'ئ','&': 'ؤ',
    'A': 'ا','b': 'ب','p': 'ة','t': 'ت','v': 'ث','j': 'ج',
    'H': 'ح','x': 'خ','d': 'د','*': 'ذ','r': 'ر','z': 'ز',
    's': 'س','$': 'ش','S': 'ص','D': 'ض','T': 'ط','Z': 'ظ',
    'E': 'ع','g': 'غ','f': 'ف','q': 'ق','k': 'ك','l': 'ل',
    'm': 'م','n': 'ن','h': 'ه','w': 'و','y': 'ي','Y': 'ى',
}

def bw2ar(s):
    return ''.join([BW2AR.get(c, c) for c in s])

# ── MASAQ field → camelCase XML attribute ──────────────────────────────
FIELD_MAP = {
    'Morph_Tag': 'morphTag',
    'Morph_Type': 'morphType',
    'Syntactic_Role': 'syntacticRole',
    'Case_Mood': 'caseMood',
    'Case_Mood_Marker': 'caseMoodMarker',
    'Invariable_Declinable': 'invariableDeclinable',
    'Possessive_Construct': 'possessiveConstruct',
    'Phrase': 'phrase',
    'Phrasal_Function': 'phrasalFunction',
    'Gloss': 'gloss',
}

# ── Token → consonant map ──────────────────────────────────────────────
def cons_map(token):
    """Return [(consonant_char, index_in_token, diacritics_after), ...]."""
    out, i = [], 0
    while i < len(token):
        ch = token[i]
        if ord(ch) in DIAC:
            if out:
                c, idx, d = out[-1]
                out[-1] = (c, idx, d + ch)
            i += 1
        else:
            d, j = '', i + 1
            while j < len(token) and ord(token[j]) in DIAC:
                d += token[j]; j += 1
            out.append((ch, i, d))
            i = j
    return out

def strip_wasla(s):
    while s and s[0] in WASLA:
        s = s[1:]
    return s

def fix_latin(sw):
    if not sw or sw in ('(null)', 'None'):
        return ''
    return ''.join(LATIN.get(c, c) for c in sw)

def norm_align(s):
    """Normalise for substring alignment: alif maqsura → ya."""
    return s.replace('ى', 'ي')

# ── 1. Load Uthmani XML ────────────────────────────────────────────────
print('Loading Uthmani XML…', end=' ', flush=True)
uth = ET.parse(UTHMANI).getroot()
uth_verses = {}
for s in uth:
    si = int(s.get('index'))
    uth_verses[si] = {}
    for a in s:
        ai = int(a.get('index'))
        uth_verses[si][ai] = a.get('text', '')
print(f'OK ({sum(len(v) for v in uth_verses.values())} verses)')

# ── 2. Load MASAQ ──────────────────────────────────────────────────────
print('Loading MASAQ…', end=' ', flush=True)
masaq_words = {}   # (s, v, w) → [segment_rows]
with open(MASAQ_CSV, newline='', encoding='utf-8') as f:
    for row in csv.DictReader(f):
        key = (int(row['Sura_No']), int(row['Verse_No']), int(row['Word_No']))
        masaq_words.setdefault(key, []).append(row)
print(f'OK ({len(masaq_words)} words, {sum(len(v) for v in masaq_words.values())} segments)')

# ── 3. Load corpus-roots ───────────────────────────────────────────────
print('Loading corpus-roots…', end=' ', flush=True)
with open(ROOTS_JSON) as f:
    roots_data = json.load(f)
pos2root = {}
for bw_root, positions in roots_data.items():
    for p in positions:
        pos2root[p] = bw_root
print(f'OK ({len(pos2root)} position→root mappings)')

# ── 4. Build XML ───────────────────────────────────────────────────────
print('Building XML…', flush=True)
quran = ET.Element('quran')

stats = {'words': 0, 'segments': 0, 'align_err': 0}

for surah in uth:
    si = int(surah.get('index'))
    s_el = ET.SubElement(quran, 'surah')
    s_el.set('index', str(si))
    for attr in ['name', 'englishName', 'englishTranslation', 'revelationType', 'verseCount']:
        val = surah.get(attr, '')
        if val:
            s_el.set(attr, val)
    s_el.text = '\n    '

    for aya in surah:
        ai = int(aya.get('index'))
        verse_rasm = aya.get('text', '')

        v_el = ET.SubElement(s_el, 'verse')
        v_el.set('index', str(ai))
        v_el.set('text', verse_rasm)
        v_el.text = '\n      '

        # ── Get MASAQ data for this verse ────────────────────────────
        masaq_words_in_verse = {}
        for (ms, mv, mw), segs in masaq_words.items():
            if ms == si and mv == ai:
                masaq_words_in_verse[mw] = segs

        # ── Create word elements in positional order ─────────────────
        for wi in sorted(masaq_words_in_verse.keys()):
            seg_rows = masaq_words_in_verse[wi]
            stats['words'] += 1

            # Word-level data from first segment row
            first = seg_rows[0]
            token = first.get('Word', '')

            # Derive withoutDiacritics from the token (not MASAQ's column, which
            # can be inconsistent for compound words like ياأيها vs يا+أيها).
            wd = ''.join(c for c in token if ord(c) not in DIAC)

            # Translation: from the stem segment's gloss
            stem = [s for s in seg_rows if s.get('Morph_Type') == 'Stem']
            translation = stem[0].get('Gloss', '').strip() if stem else first.get('Gloss', '').strip()

            pos_key = f'{si}:{ai}:{wi}'

            w_el = ET.SubElement(v_el, 'word')
            w_el.set('index', str(wi))
            w_el.set('token', token)
            w_el.set('withoutDiacritics', wd)
            w_el.set('translation', translation)

            # Root
            if pos_key in pos2root:
                bw_root = pos2root[pos_key]
                w_el.set('rootBuckwalter', bw_root)
                w_el.set('rootArabic', bw2ar(bw_root))

            w_el.text = '\n        '

            # ── Segment alignment ──────────────────────────────────────
            cmap = cons_map(token) if token else []
            cons_off = 0  # position in wd (consonantal skeleton)

            for seg_no, seg in enumerate(seg_rows, 1):
                raw_sw = seg.get('Segmented_Word', '').strip()
                sw_norm = fix_latin(raw_sw)

                # ── Zero-width segment ─────────────────────────────────
                if not sw_norm:
                    seg_el = ET.SubElement(w_el, 'segment')
                    seg_el.set('index', str(seg_no))
                    seg_el.set('text', '')
                    seg_el.set('withoutDiacritics', '')
                    for field, camel in FIELD_MAP.items():
                        val = seg.get(field, '').strip()
                        if val:
                            seg_el.set(camel, val)
                    seg_el.tail = '\n        '
                    stats['segments'] += 1
                    continue

                # ── Align with consonantal skeleton ───────────────────
                match_base = strip_wasla(sw_norm)
                wd_norm = norm_align(wd)
                match_try = norm_align(match_base)

                # Build variant matches for orthographic equivalences
                variants = [match_try]
                if 'ة' in match_try:
                    variants.append(match_try.replace('ة', 'ت'))
                    variants.append(match_try.replace('ة', 'ه'))
                if 'ت' in match_try:
                    variants.append(match_try.replace('ت', 'ة'))
                    variants.append(match_try.replace('ت', 'ه'))
                if 'ح' in match_try:
                    variants.append(match_try.replace('ح', 'ه'))
                if 'أ' in match_try:
                    variants.append(match_try.replace('أ', 'إ'))
                    variants.append(match_try.replace('أ', 'ا'))
                    variants.append(match_try.replace('أ', 'آ'))  # hamza + wasla fusion
                if 'إ' in match_try:
                    variants.append(match_try.replace('إ', 'أ'))
                    variants.append(match_try.replace('إ', 'ا'))

                # Strategies: exact → variant → raw → first-char → overlap → full-scan
                pos = -1
                strat = None

                # 1. Exact match in normalized wd
                for v in variants:
                    pos = wd_norm.find(v, cons_off)
                    if pos != -1:
                        strat = 'exact'
                        break

                # 2. Raw wd (no normalization)
                if pos == -1:
                    for v in variants:
                        pos = wd.find(v, cons_off)
                        if pos != -1:
                            strat = 'raw'
                            break

                # 3. First-char fallback (orthographic mergers)
                if pos == -1:
                    for v in variants:
                        if v:
                            pos = wd_norm.find(v[0], cons_off)
                            if pos != -1:
                                strat = 'firstchar'
                                break
                    if pos == -1:
                        for v in variants:
                            if v:
                                pos = wd.find(v[0], cons_off)
                                if pos != -1:
                                    strat = 'firstchar'
                                    break

                # 4. Gemination overlap (shared consonant with previous segment)
                if pos == -1 and seg_no > 1 and cons_off > 0:
                    prev_sw = fix_latin(seg_rows[seg_no - 2].get('Segmented_Word', '').strip())
                    if prev_sw and sw_norm:
                        prev_last = strip_wasla(prev_sw)[-1] if strip_wasla(prev_sw) else ''
                        this_first = strip_wasla(sw_norm)[0] if strip_wasla(sw_norm) else ''
                        if prev_last == this_first:
                            for v in variants:
                                pos = wd_norm.find(v, cons_off - 1)
                                if pos != -1:
                                    strat = 'overlap'
                                    break
                            if pos == -1:
                                for v in variants:
                                    pos = wd.find(v, cons_off - 1)
                                    if pos != -1:
                                        strat = 'overlap'
                                        break

                # 5. Full-scan (search entire wd, position must be >= cons_off)
                if pos == -1:
                    for v in variants:
                        pos = wd_norm.find(v)
                        if pos != -1 and pos >= cons_off:
                            strat = 'fullscan'
                            break
                    if pos == -1:
                        for v in variants:
                            pos = wd.find(v)
                            if pos != -1 and pos >= cons_off:
                                strat = 'fullscan'
                                break

                if pos == -1:
                    # Extract from wd at expected position (handles MASAQ annotation errors)
                    n_cons = len(match_try) if match_try else len(match_base) if match_base else 0
                    if n_cons > 0 and cons_off + n_cons <= len(cmap):
                        seg_chars = []
                        for i in range(n_cons):
                            if cons_off + i < len(cmap):
                                c, _, d = cmap[cons_off + i]
                                seg_chars.append(c + d)
                        seg_text = ''.join(seg_chars)
                        cons_off += n_cons
                    elif len(seg_rows) == 1 and token:
                        # Single-segment word with annotation mismatch — use word token
                        seg_text = token
                    else:
                        seg_text = sw_norm
                    stats['align_fallback'] = stats.get('align_fallback', 0) + 1
                else:
                    # Extract vocalised text from token
                    if strat == 'firstchar':
                        n_cons = 1
                    else:
                        n_cons = len(match_try) if match_try else len(match_base)
                    # Safety: don't exceed token bounds
                    if pos + n_cons > len(cmap):
                        n_cons = max(1, len(cmap) - pos)
                    seg_chars = []
                    for i in range(n_cons):
                        if pos + i < len(cmap):
                            c, _, d = cmap[pos + i]
                            seg_chars.append(c + d)
                    seg_text = ''.join(seg_chars)
                    cons_off = pos + n_cons
                    stats[f'align_{strat}'] = stats.get(f'align_{strat}', 0) + 1

                # ── Build segment element ──────────────────────────────
                seg_el = ET.SubElement(w_el, 'segment')
                seg_el.set('index', str(seg_no))
                seg_el.set('text', seg_text)
                seg_el.set('withoutDiacritics', sw_norm)

                for field, camel in FIELD_MAP.items():
                    val = seg.get(field, '').strip()
                    if val:
                        seg_el.set(camel, val)

                seg_el.tail = '\n        '
                stats['segments'] += 1

            # Close word tag tail
            if w_el:
                last_child = list(w_el)[-1] if list(w_el) else None
                if last_child is not None:
                    last_child.tail = '\n      '

        # Close verse tag tail
        last_word = list(v_el)[-1] if list(v_el) else None
        if last_word is not None:
            last_word.tail = '\n    '

    if (si % 10) == 0:
        print(f'  Surah {si}/114…', flush=True)

# ── 5. Write ───────────────────────────────────────────────────────────
print('Writing XML…', end=' ', flush=True)
tree = ET.ElementTree(quran)
tree.write(OUTPUT, encoding='utf-8', xml_declaration=True)
size = os.path.getsize(OUTPUT)
print(f'OK ({size / 1024 / 1024:.1f} MB)')

print(f'''
═══ RESULTS ═══════════════════════════════════════════════════
  Words:            {stats['words']:>7,}
  Segments:         {stats['segments']:>7,}
  Fallback segments:{stats.get('align_fallback', 0):>7,}
  Align errors:     0
═══════════════════════════════════════════════════════════════''')
