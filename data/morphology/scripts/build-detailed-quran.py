#!/usr/bin/env python3
"""
Build detailed-quran.xml, detailed-quran.db, detailed-quran.sql
from all fixed/verified data sources.
"""
import csv, json, sqlite3, xml.etree.ElementTree as ET, os, re

# ── Paths ──────────────────────────────────────────────────────────────
UTHMANI_XML = '/Users/saqlain/corpus-quran-java/data/quran/quran-uthmani.xml'
MASAQ_CSV   = '/Users/saqlain/corpus-quran-java/data/morphology/masaq_fixed.csv'
MASAQ_DB    = '/Users/saqlain/corpus-quran-java/data/morphology/extracted/MASAQ Morphologically-Analyzed and Syntactically-Annotated Quran Dataset/MASAQ.db'
ROOTS_JSON  = '/Users/saqlain/corpus-quran-java/data/morphology/corpus-roots.json'
CONCORDANCE = '/Users/saqlain/corpus-quran-java/data/morphology/concordance.jsonl'
MEANINGS     = '/Users/saqlain/corpus-quran-java/data/morphology/meanings-roots-ai.jsonl'

OUT_DIR = '/Users/saqlain/corpus-quran-java/data/new'
OUT_XML  = f'{OUT_DIR}/detailed-quran.xml'
OUT_DB   = f'{OUT_DIR}/detailed-quran.db'
OUT_SQL  = f'{OUT_DIR}/detailed-quran.sql'

# ── Unicode ranges ─────────────────────────────────────────────────────
DIAC = set(range(0x064B, 0x0653)) | {0x0670} \
       | set(range(0x06D6, 0x06EE)) | {0x06FE, 0x06FF}
WASLA = {'\u0671', '\u0627', '\u0622', '\u0672', '\u0673'}
LATIN = {'W': 'و', 'B': 'ب', 'M': 'م', 'L': 'ل', 'I': 'ا'}

def strip_diac(text):
    return ''.join(c for c in text if ord(c) not in DIAC)

def fix_latin(sw):
    if not sw or sw in ('(null)', 'None'): return ''
    return ''.join(LATIN.get(c, c) for c in sw)

def strip_wasla(s):
    while s and s[0] in WASLA: s = s[1:]
    return s

def norm_align(s):
    return s.replace('ى', 'ي').replace('ٱ', 'ا')

def make_variants(mt):
    vs = [mt]
    if 'ة' in mt: vs.extend([mt.replace('ة','ت'), mt.replace('ة','ه')])
    if 'ت' in mt: vs.extend([mt.replace('ت','ة'), mt.replace('ت','ه')])
    if 'ح' in mt: vs.append(mt.replace('ح','ه'))
    if 'أ' in mt: vs.extend([mt.replace('أ','إ'), mt.replace('أ','ا'), mt.replace('أ','آ')])
    if 'إ' in mt: vs.extend([mt.replace('إ','أ'), mt.replace('إ','ا')])
    return vs

def cons_map(token):
    out, i = [], 0
    while i < len(token):
        ch = token[i]
        if ord(ch) in DIAC:
            if out: c, idx, d = out[-1]; out[-1] = (c, idx, d + ch)
            i += 1
        else:
            d, j = '', i + 1
            while j < len(token) and ord(token[j]) in DIAC: d += token[j]; j += 1
            out.append((ch, i, d)); i = j
    return out

BW2AR = {"'": 'ء','|': 'آ','>': 'أ','<': 'إ','}': 'ئ','&': 'ؤ','A': 'ا','b': 'ب','p': 'ة','t': 'ت','v': 'ث','j': 'ج','H': 'ح','x': 'خ','d': 'د','*': 'ذ','r': 'ر','z': 'ز','s': 'س','$': 'ش','S': 'ص','D': 'ض','T': 'ط','Z': 'ظ','E': 'ع','g': 'غ','f': 'ف','q': 'ق','k': 'ك','l': 'ل','m': 'م','n': 'ن','h': 'ه','w': 'و','y': 'ي','Y': 'ى',}
def bw2ar(s):
    return ''.join([BW2AR.get(c, c) for c in s])

FIELD_MAP = {'Morph_Tag':'morphTag','Morph_Type':'morphType','Syntactic_Role':'syntacticRole','Case_Mood':'caseMood','Case_Mood_Marker':'caseMoodMarker','Invariable_Declinable':'invariableDeclinable','Possessive_Construct':'possessiveConstruct','Phrase':'phrase','Phrasal_Function':'phrasalFunction','Gloss':'gloss'}

# ════════════════════════════════════════════════════════════════════════
# 1. LOAD ALL DATA
# ════════════════════════════════════════════════════════════════════════
print('═══ Loading data sources ═══')

print('  Uthmani XML…', end=' ', flush=True)
uth = ET.parse(UTHMANI_XML).getroot()
print(f'{sum(1 for s in uth for _ in s)} verses')

print('  MASAQ fixed CSV…', end=' ', flush=True)
masaq = {}
with open(MASAQ_CSV, newline='', encoding='utf-8') as f:
    for row in csv.DictReader(f):
        key = (int(row['Sura_No']), int(row['Verse_No']), int(row['Word_No']))
        masaq.setdefault(key, []).append(row)
print(f'{len(masaq)} words, {sum(len(v) for v in masaq.values())} segments')

print('  corpus-roots.json…', end=' ', flush=True)
with open(ROOTS_JSON) as f:
    roots_data = json.load(f)
pos2root = {}; root_freq = {}
for bw_root, positions in roots_data.items():
    root_freq[bw_root] = len(positions)
    for p in positions:
        pos2root[p] = bw_root
print(f'{len(root_freq)} roots')

print('  concordance.jsonl…', end=' ', flush=True)
root_lemmas = {}
with open(CONCORDANCE) as f:
    for line in f:
        d = json.loads(line)
        root_lemmas[d['root']] = {'total_occurrences': d['total_occurrences'], 'lemmas': d['lemmas']}
print(f'{len(root_lemmas)} roots with lemmas')

print('  meanings-roots-ai.jsonl…', end=' ', flush=True)
root_meanings = {}
with open(MEANINGS) as f:
    for line in f:
        d = json.loads(line)
        root_meanings[d['root']] = d
print(f'{len(root_meanings)} roots with meanings')

# ── MASAQ official DB unavailable for reference tables (deprecated) ──
# Previously loaded ref_* lookup tables from MASAQ.db. These tables were
# never queried by the Go runtime and have been removed per normalization
# (2025-06-26). The morphological tag values remain in segments — they just
# no longer have redundant description lookup tables.

# ════════════════════════════════════════════════════════════════════════
# 2. PREPARE DATABASE
# ════════════════════════════════════════════════════════════════════════
print('\n═══ Building detailed-quran.db ═══')

for f in [OUT_DB, f'{OUT_DB}-shm', f'{OUT_DB}-wal']:
    if os.path.exists(f): os.remove(f)

db = sqlite3.connect(OUT_DB)
db.execute('PRAGMA journal_mode=WAL')
db.execute('PRAGMA foreign_keys=ON')

db.executescript('''
CREATE TABLE surahs (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    english_name TEXT,
    english_translation TEXT,
    revelation_type TEXT,
    verse_count INTEGER
);
CREATE TABLE verses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    surah_id INTEGER NOT NULL REFERENCES surahs(id),
    verse_number INTEGER NOT NULL,
    text_uthmani TEXT NOT NULL,
    text_imla_i TEXT,
    UNIQUE(surah_id, verse_number)
);
CREATE TABLE words (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    surah_id INTEGER NOT NULL,
    verse_id INTEGER NOT NULL REFERENCES verses(id),
    word_number INTEGER NOT NULL,
    token_imla_i TEXT NOT NULL,
    translation TEXT,
    root_buckwalter TEXT REFERENCES roots(root_buckwalter),
    UNIQUE(surah_id, verse_id, word_number)
);
CREATE TABLE segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    word_id INTEGER NOT NULL REFERENCES words(id),
    segment_number INTEGER NOT NULL,
    text TEXT,
    without_diacritics TEXT,
    morph_tag TEXT,
    morph_type TEXT,
    syntactic_role TEXT,
    case_mood TEXT,
    case_mood_marker TEXT,
    invariable_declinable TEXT,
    possessive_construct TEXT,
    phrase TEXT,
    phrasal_function TEXT,
    gloss TEXT,
    UNIQUE(word_id, segment_number)
);
CREATE TABLE roots (
    id INTEGER PRIMARY KEY,
    root_buckwalter TEXT NOT NULL UNIQUE,
    root_arabic TEXT,
    root_letters TEXT,
    pos TEXT,
    occurrences_quran INTEGER,
    meaning_en_short TEXT,
    meaning_en TEXT,
    meaning_ar TEXT,
    meaning_en_detailed TEXT,
    meaning_trilateral TEXT,
    meaning_ar_definition TEXT,
    etymology TEXT,
    derived_forms TEXT,
    key_nominals TEXT,
    quranic_analysis TEXT,
    lexicographical_sources TEXT,
    semantic_field TEXT,
    theological_dimensions TEXT,
    hadith_evidence TEXT,
    cross_references TEXT
);
CREATE TABLE root_lemmas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_buckwalter TEXT NOT NULL REFERENCES roots(root_buckwalter),
    lemma_arabic TEXT NOT NULL,
    occurrences INTEGER,
    UNIQUE(root_buckwalter, lemma_arabic)
);
CREATE TABLE lemma_positions (
    lemma_id INTEGER NOT NULL REFERENCES root_lemmas(id),
    verse_id INTEGER NOT NULL REFERENCES verses(id)
);
''')

db.execute('CREATE INDEX idx_verses_surah ON verses(surah_id)')
db.execute('CREATE INDEX idx_words_verse ON words(verse_id)')
db.execute('CREATE INDEX idx_segments_word ON segments(word_id)')
db.execute('CREATE INDEX idx_words_root ON words(root_buckwalter)')

# ════════════════════════════════════════════════════════════════════════
# 3. BUILD XML + POPULATE SQL
# ════════════════════════════════════════════════════════════════════════
print('\nBuilding detailed-quran.xml + populating DB…')

quran = ET.Element('quran')
stats = {'words': 0, 'segments': 0, 'fallback': 0}

for surah in uth:
    si = int(surah.get('index'))
    
    # Surah metadata → DB
    db.execute('INSERT INTO surahs (id,name,english_name,english_translation,revelation_type,verse_count) VALUES (?,?,?,?,?,?)',
        (si, surah.get('name',''), surah.get('englishName',''), surah.get('englishTranslation',''),
         surah.get('revelationType',''), int(surah.get('verseCount',0))))
    
    s_el = ET.SubElement(quran, 'surah')
    for attr in ['index','name','englishName','englishTranslation','revelationType','verseCount']:
        v = surah.get(attr, '')
        if v: s_el.set(attr, str(v))
    s_el.text = '\n    '

    for aya in surah:
        ai = int(aya.get('index'))
        verse_rasm = aya.get('text', '')

        # Verse → DB
        imla_i_parts = []
        for mw in sorted([mw for (ms, mv, mw) in masaq if ms == si and mv == ai]):
            imla_i_parts.append(masaq[(si, ai, mw)][0].get('Word', ''))
        imla_i_text = ' '.join(imla_i_parts) if imla_i_parts else None
        
        cur = db.execute('INSERT INTO verses (surah_id,verse_number,text_uthmani,text_imla_i) VALUES (?,?,?,?)',
            (si, ai, verse_rasm, imla_i_text))
        verse_db_id = cur.lastrowid

        # Verse → XML
        v_el = ET.SubElement(s_el, 'verse')
        v_el.set('index', str(ai))
        v_el.set('text', verse_rasm)
        v_el.text = '\n      '

        words_here = {mw: segs for (ms, mv, mw), segs in masaq.items() if ms == si and mv == ai}
        if not words_here: continue

        for wi in sorted(words_here.keys()):
            seg_rows = words_here[wi]
            stats['words'] += 1
            first = seg_rows[0]
            token = first.get('Word', '')
            wd = strip_diac(token).replace('ـ', '')

            stem = [s for s in seg_rows if s.get('Morph_Type', '') == 'Stem']
            translation = stem[0].get('Gloss', '').strip() if stem else first.get('Gloss', '').strip()

            pos_key = f'{si}:{ai}:{wi}'
            root_bw = pos2root.get(pos_key)
            root_ar = bw2ar(root_bw) if root_bw else None

            # Word → DB
            cur = db.execute('INSERT INTO words (surah_id,verse_id,word_number,token_imla_i,translation,root_buckwalter) VALUES (?,?,?,?,?,?)',
                (si, verse_db_id, wi, token, translation, root_bw))
            word_db_id = cur.lastrowid

            # Word → XML
            w_el = ET.SubElement(v_el, 'word')
            w_el.set('index', str(wi))
            w_el.set('token', token)
            w_el.set('withoutDiacritics', wd)
            w_el.set('translation', translation)
            if root_bw:
                w_el.set('rootBuckwalter', root_bw)
                w_el.set('rootArabic', root_ar or '')
                # Add root meaning if available
                if root_bw in root_meanings:
                    rm = root_meanings[root_bw]
                    w_el.set('rootMeaning', rm.get('meaning',{}).get('en',''))
                if root_bw in root_freq:
                    w_el.set('rootFrequency', str(root_freq[root_bw]))
            w_el.text = '\n        '

            # ── Segment alignment + DB insert ─────────────────────────
            cmap = cons_map(token) if token else []
            cons_off = 0

            for seg_no, seg in enumerate(seg_rows, 1):
                raw_sw = seg.get('Segmented_Word', '').strip()
                sw_norm = fix_latin(raw_sw)

                # Zero-width
                if not sw_norm:
                    seg_xml = ET.SubElement(w_el, 'segment')
                    seg_xml.set('index', str(seg_no))
                    seg_xml.set('text', '')
                    seg_xml.set('withoutDiacritics', '')
                    for f, c in FIELD_MAP.items():
                        v = seg.get(f, '').strip()
                        if v: seg_xml.set(c, v)
                    seg_xml.tail = '\n        '
                    
                    db.execute('INSERT INTO segments (word_id,segment_number,text,without_diacritics,morph_tag,morph_type,syntactic_role,case_mood,case_mood_marker,invariable_declinable,possessive_construct,phrase,phrasal_function,gloss) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)',
                        (word_db_id, seg_no, '', '', seg.get('Morph_Tag',''), seg.get('Morph_Type',''), seg.get('Syntactic_Role',''), seg.get('Case_Mood',''), seg.get('Case_Mood_Marker',''), seg.get('Invariable_Declinable',''), seg.get('Possessive_Construct',''), seg.get('Phrase',''), seg.get('Phrasal_Function',''), seg.get('Gloss','')))
                    stats['segments'] += 1
                    continue

                # Alignment
                mb = strip_wasla(sw_norm)
                wd_n = norm_align(wd)
                mt = norm_align(mb)
                vs = make_variants(mt)
                
                pos = -1
                for swd in [wd_n, wd]:
                    for v in vs:
                        pos = swd.find(v, cons_off)
                        if pos != -1: break
                    if pos != -1: break
                
                if pos == -1:
                    for swd in [wd_n, wd]:
                        for v in vs:
                            if v:
                                pos = swd.find(v[0], cons_off)
                                if pos != -1: break
                        if pos != -1: break
                
                if pos == -1 and seg_no > 1 and cons_off > 0:
                    prev_s = fix_latin(seg_rows[seg_no-2].get('Segmented_Word','').strip())
                    if prev_s and sw_norm:
                        pl = strip_wasla(prev_s)[-1] if strip_wasla(prev_s) else ''
                        tf = strip_wasla(sw_norm)[0] if strip_wasla(sw_norm) else ''
                        if pl == tf:
                            for swd in [wd_n, wd]:
                                for v in vs:
                                    pos = swd.find(v, cons_off - 1)
                                    if pos != -1: break
                                if pos != -1: break
                
                if pos == -1:
                    for swd in [wd_n, wd]:
                        for v in vs:
                            pos = swd.find(v)
                            if pos != -1 and pos >= cons_off: break
                        if pos != -1: break
                
                if pos == -1:
                    n_cons = len(mt)
                    if n_cons > 0 and cons_off + n_cons <= len(cmap):
                        sc = []
                        for i in range(n_cons):
                            if cons_off + i < len(cmap):
                                c, _, d = cmap[cons_off + i]
                                sc.append(c + d)
                        seg_text = ''.join(sc)
                        cons_off += n_cons
                    elif len(seg_rows) == 1 and token:
                        seg_text = token
                    else:
                        seg_text = sw_norm
                    stats['fallback'] += 1
                else:
                    n_cons = 1 if pos == cons_off - 1 else len(mt)
                    if pos + n_cons > len(cmap):
                        n_cons = max(1, len(cmap) - pos)
                    sc = []
                    for i in range(n_cons):
                        if pos + i < len(cmap):
                            c, _, d = cmap[pos + i]
                            sc.append(c + d)
                    seg_text = ''.join(sc)
                    cons_off = pos + n_cons

                # Segment → XML + DB
                seg_xml = ET.SubElement(w_el, 'segment')
                seg_xml.set('index', str(seg_no))
                seg_xml.set('text', seg_text)
                seg_xml.set('withoutDiacritics', sw_norm)
                for f, c in FIELD_MAP.items():
                    v = seg.get(f, '').strip()
                    if v: seg_xml.set(c, v)
                seg_xml.tail = '\n        '
                
                db.execute('INSERT INTO segments (word_id,segment_number,text,without_diacritics,morph_tag,morph_type,syntactic_role,case_mood,case_mood_marker,invariable_declinable,possessive_construct,phrase,phrasal_function,gloss) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)',
                    (word_db_id, seg_no, seg_text, sw_norm, seg.get('Morph_Tag',''), seg.get('Morph_Type',''), seg.get('Syntactic_Role',''), seg.get('Case_Mood',''), seg.get('Case_Mood_Marker',''), seg.get('Invariable_Declinable',''), seg.get('Possessive_Construct',''), seg.get('Phrase',''), seg.get('Phrasal_Function',''), seg.get('Gloss','')))
                stats['segments'] += 1

            # Close word
            if w_el:
                lc = list(w_el)[-1] if list(w_el) else None
                if lc is not None: lc.tail = '\n      '

        lw = list(v_el)[-1] if list(v_el) else None
        if lw is not None: lw.tail = '\n    '

    if si % 10 == 0:
        print(f'  Surah {si}/114…', flush=True)

# ════════════════════════════════════════════════════════════════════════
# 4. POPULATE ROOTS + LEMMAS
# ════════════════════════════════════════════════════════════════════════
print('\nPopulating roots + lemmas…', flush=True)

lemmas_inserted = 0  # counter for lemma_positions rows

for bw_root, freq in sorted(root_freq.items()):
    ar = bw2ar(bw_root)
    row = {
        'root_buckwalter': bw_root,
        'root_arabic': ar,
        'occurrences_quran': freq,
        'meaning_en': '',
        'meaning_ar': '',
        'meaning_ar_definition': '',
        'etymology': None,
        'hadith_evidence': None,
    }
    if bw_root in root_meanings:
        rm = root_meanings[bw_root]
        m = rm.get('meaning', {})
        
        row['id'] = rm.get('id')
        row['pos'] = rm.get('pos', '')
        row['root_letters'] = rm.get('root_letters', '')
        row['meaning_en_short'] = m.get('en_short', '') or ''
        row['meaning_en'] = m.get('en', '') or ''
        row['meaning_ar'] = m.get('ar', '') or ''
        row['meaning_ar_definition'] = m.get('ar_definition', '') or ''
        row['meaning_en_detailed'] = m.get('en_detailed', '') or ''
        row['meaning_trilateral'] = m.get('trilateral_meaning', '') or ''
        
        # Fallbacks
        if not row['meaning_en']:
            row['meaning_en'] = m.get('en_short', '') or ''
        if not row['meaning_ar']:
            row['meaning_ar'] = row['meaning_ar_definition']
        
        # JSON columns — store entire objects (only those kept in schema)
        for col in ('etymology', 'hadith_evidence'):
            val = rm.get(col)
            if val:
                row[col] = json.dumps(val, ensure_ascii=False)

    db.execute('''INSERT OR REPLACE INTO roots 
        (id, root_buckwalter, root_arabic, root_letters, pos, occurrences_quran,
         meaning_en, meaning_ar, meaning_ar_definition,
         etymology, hadith_evidence)
        VALUES (?,?,?,?,?,?,?,?,?,?,?)''',
        (row.get('id'), row['root_buckwalter'], row['root_arabic'], row.get('root_letters'),
         row.get('pos'), row['occurrences_quran'],
         row['meaning_en'], row['meaning_ar'], row['meaning_ar_definition'],
         row['etymology'], row['hadith_evidence']))
    
    # Lemmas from concordance (inserted after roots + verses are populated)
    if bw_root in root_lemmas:
        for lemma_ar, lemma_data in root_lemmas[bw_root]['lemmas'].items():
            cur = db.execute('INSERT OR REPLACE INTO root_lemmas (root_buckwalter,lemma_arabic,occurrences) VALUES (?,?,?)',
                (bw_root, lemma_ar, lemma_data.get('total_occurrences', 0)))
            lemma_id = cur.lastrowid
            for occ in lemma_data.get('occurrences', []):
                parts = occ.split(':')
                if len(parts) == 2:
                    try:
                        s, v = int(parts[0]), int(parts[1])
                        vid = db.execute('SELECT id FROM verses WHERE surah_id=? AND verse_number=?', (s, v)).fetchone()
                        if vid:
                            db.execute('INSERT INTO lemma_positions (lemma_id, verse_id) VALUES (?,?)',
                                (lemma_id, vid[0]))
                            lemmas_inserted += 1
                    except (ValueError, IndexError):
                        pass

# ════════════════════════════════════════════════════════════════════════
# 5. WRITE OUTPUTS
# ════════════════════════════════════════════════════════════════════════
print('\nWriting outputs…')

# XML
print(f'  {OUT_XML}…', end=' ', flush=True)
tree = ET.ElementTree(quran)
tree.write(OUT_XML, encoding='utf-8', xml_declaration=True)
print(f'{os.path.getsize(OUT_XML)/1024/1024:.1f} MB')

# SQLite
db.commit()
db.execute('ANALYZE')
db.close()
print(f'  {OUT_DB}… {os.path.getsize(OUT_DB)/1024/1024:.1f} MB')

# SQL dump
print(f'  {OUT_SQL}…', end=' ', flush=True)
os.system(f'sqlite3 "{OUT_DB}" .dump > "{OUT_SQL}"')
print(f'{os.path.getsize(OUT_SQL)/1024/1024:.1f} MB')

# ════════════════════════════════════════════════════════════════════════
# 6. REPORT
# ════════════════════════════════════════════════════════════════════════
v_count = sum(1 for s in uth for _ in s)
print(f'''
═══ FINAL REPORT ═══════════════════════════════════════════
  Verses (rasm + imla'i):   {v_count:,}
  Words (with roots):       {stats['words']:,}
  Segments (aligned):       {stats['segments']:,}
  Roots (with meanings):    {len(root_freq):,}
  Root lemmas:              {sum(len(rl['lemmas']) for rl in root_lemmas.values()):,}
  Lemma positions:          {lemmas_inserted:,}
  Fallback segments:        {stats['fallback']}
  
  Outputs:
    {os.path.basename(OUT_XML)}  ({os.path.getsize(OUT_XML)/1024/1024:.1f} MB)
    {os.path.basename(OUT_DB)}   ({os.path.getsize(OUT_DB)/1024/1024:.1f} MB)
    {os.path.basename(OUT_SQL)}  ({os.path.getsize(OUT_SQL)/1024/1024:.1f} MB)
═════════════════════════════════════════════════════════════''')
