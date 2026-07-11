#!/usr/bin/env python3
"""Apply normalization to detailed-quran.db — run once to migrate existing DB."""

import sqlite3, os, sys

DB = '/Users/saqlain/corpus-quran-java/data/new/detailed-quran.db'
BACKUP = DB + '.pre-norm-backup'

if not os.path.exists(DB):
    print(f"ERROR: {DB} not found", file=sys.stderr)
    sys.exit(1)

# Backup
print(f"Backing up to {BACKUP}…")
os.system(f'cp "{DB}" "{BACKUP}"')

db = sqlite3.connect(DB)
db.execute('PRAGMA journal_mode=WAL')
db.execute('PRAGMA foreign_keys=OFF')
db.execute('BEGIN')

try:
    # ── 1. Drop 6 ref_* tables ──────────────────────────────────────────
    print('Dropping ref_* tables…')
    for t in ['ref_case_mood','ref_syntactic_role','ref_invariable_declinable',
              'ref_possessive_construct','ref_phrase','ref_phrasal_function',
              'ref_Case_Mood_Marker']:
        db.execute(f'DROP TABLE IF EXISTS {t}')
    print('  Done.')

    # ── 2. Recreate lemma_positions with verse_id FK ────────────────────
    print('Normalizing lemma_positions (surah_id,verse_number → verse_id)…')
    db.execute('''
        CREATE TABLE lemma_positions_new (
            lemma_id INTEGER NOT NULL REFERENCES root_lemmas(id),
            verse_id INTEGER NOT NULL REFERENCES verses(id)
        )
    ''')
    db.execute('''
        INSERT INTO lemma_positions_new (lemma_id, verse_id)
        SELECT lp.lemma_id, v.id
        FROM lemma_positions lp
        JOIN verses v ON lp.surah_id = v.surah_id AND lp.verse_number = v.verse_number
    ''')
    db.execute('DROP TABLE lemma_positions')
    db.execute('ALTER TABLE lemma_positions_new RENAME TO lemma_positions')
    db.execute('CREATE INDEX idx_lemma_positions_verse ON lemma_positions(verse_id)')
    db.execute('CREATE INDEX idx_lemma_positions_lemma ON lemma_positions(lemma_id)')
    count = db.execute('SELECT COUNT(*) FROM lemma_positions').fetchone()[0]
    print(f'  {count:,} rows migrated.')

    # ── 3. Recreate roots — drop 11 unused columns ──────────────────────
    print('Dropping 11 unused columns from roots…')
    db.execute('''
        CREATE TABLE roots_new (
            id INTEGER PRIMARY KEY,
            root_buckwalter TEXT NOT NULL UNIQUE,
            root_arabic TEXT,
            root_letters TEXT,
            pos TEXT,
            occurrences_quran INTEGER,
            meaning_en TEXT,
            meaning_ar TEXT,
            meaning_ar_definition TEXT,
            etymology TEXT,
            hadith_evidence TEXT
        )
    ''')
    db.execute('''
        INSERT INTO roots_new 
            (id, root_buckwalter, root_arabic, root_letters, pos,
             occurrences_quran, meaning_en, meaning_ar,
             meaning_ar_definition, etymology, hadith_evidence)
        SELECT id, root_buckwalter, root_arabic, root_letters, pos,
               occurrences_quran, meaning_en, meaning_ar,
               meaning_ar_definition, etymology, hadith_evidence
        FROM roots
    ''')
    db.execute('DROP TABLE roots')
    db.execute('ALTER TABLE roots_new RENAME TO roots')
    count = db.execute('SELECT COUNT(*) FROM roots').fetchone()[0]
    print(f'  {count:,} rows — 11 columns dropped (kept 10 + id).')

    # ── 4. Recreate root_lemmas (FK to new roots table) ─────────────────
    print('Recreating root_lemmas (FK refresh)…')
    db.execute('''
        CREATE TABLE root_lemmas_new (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            root_buckwalter TEXT NOT NULL REFERENCES roots(root_buckwalter),
            lemma_arabic TEXT NOT NULL,
            occurrences INTEGER,
            UNIQUE(root_buckwalter, lemma_arabic)
        )
    ''')
    db.execute('INSERT INTO root_lemmas_new SELECT * FROM root_lemmas')
    db.execute('DROP TABLE root_lemmas')
    db.execute('ALTER TABLE root_lemmas_new RENAME TO root_lemmas')
    count = db.execute('SELECT COUNT(*) FROM root_lemmas').fetchone()[0]
    print(f'  {count:,} rows migrated.')

    # ── 5. Recreate words — drop root_arabic, without_diacritics; add FK ─
    print('Normalizing words (drop root_arabic, without_diacritics; FK to roots)…')
    db.execute('''
        CREATE TABLE words_new (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            surah_id INTEGER NOT NULL,
            verse_id INTEGER NOT NULL REFERENCES verses(id),
            word_number INTEGER NOT NULL,
            token_imla_i TEXT NOT NULL,
            translation TEXT,
            root_buckwalter TEXT REFERENCES roots(root_buckwalter),
            UNIQUE(surah_id, verse_id, word_number)
        )
    ''')
    db.execute('''
        INSERT INTO words_new (id, surah_id, verse_id, word_number, token_imla_i, translation, root_buckwalter)
        SELECT id, surah_id, verse_id, word_number, token_imla_i, translation,
               CASE WHEN root_buckwalter IS NULL OR root_buckwalter = '' OR root_buckwalter = 'None'
                    THEN NULL ELSE root_buckwalter END
        FROM words
    ''')
    db.execute('DROP TABLE words')
    db.execute('ALTER TABLE words_new RENAME TO words')

    # Recreate indexes
    db.execute('CREATE INDEX idx_words_verse ON words(verse_id)')
    db.execute('CREATE INDEX idx_words_root ON words(root_buckwalter)')
    count = db.execute('SELECT COUNT(*) FROM words').fetchone()[0]
    print(f'  {count:,} rows — 2 columns dropped, FK added.')

    # ── 6. Recreate segments (refresh FK to new words table) ────────────
    print('Recreating segments (FK refresh)…')
    db.execute('''
        CREATE TABLE segments_new (
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
        )
    ''')
    db.execute('INSERT INTO segments_new SELECT * FROM segments')
    db.execute('DROP TABLE segments')
    db.execute('ALTER TABLE segments_new RENAME TO segments')
    db.execute('CREATE INDEX idx_segments_word ON segments(word_id)')
    count = db.execute('SELECT COUNT(*) FROM segments').fetchone()[0]
    print(f'  {count:,} rows migrated.')

    # ── Commit & verify ─────────────────────────────────────────────────
    db.commit()
    db.execute('PRAGMA foreign_keys=ON')

    fk_check = db.execute('PRAGMA foreign_key_check').fetchall()
    if fk_check:
        print(f'ERROR: FK violations after migration: {len(fk_check)} rows')
        for row in fk_check[:5]:
            print(f'  {row}')
        sys.exit(1)
    else:
        print('FK check: PASSED (zero violations)')

    integrity = db.execute('PRAGMA integrity_check').fetchone()[0]
    print(f'Integrity check: {integrity}')

    db.execute('ANALYZE')
    db.close()

    print(f'\n═══ Normalization complete ═══')
    print(f'Backup saved: {BACKUP}')

except Exception as e:
    db.rollback()
    db.close()
    print(f'ERROR: {e}', file=sys.stderr)
    print(f'Restored from backup. DB unchanged.')
    sys.exit(1)
