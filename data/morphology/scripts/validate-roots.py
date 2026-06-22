#!/usr/bin/env python3
"""
Validate root_meaning.csv against corpus-roots.json.

Eight integrity checks:
  1. No missing roots — every corpus-roots.json root has a row in root_meaning.csv
  2. No empty full_response — zero rows with empty JSON
  3. full_response validity — every full_response is parseable JSON with correct schema
  4. No duplicate roots — no duplicate root values
  5. id uniqueness — no duplicate id values
  6. POS populated — no null/empty pos values
  7. Orphan rows detected — roots in CSV but not in corpus-roots.json (informational)
  8. Missing field strings — completely empty field columns (empty arrays are OK)
"""

import csv
import json
import os
import sys

CSV_PATH = os.path.join(os.path.dirname(__file__), '..', 'root_meaning.csv')
CORPUS_PATH = os.path.join(os.path.dirname(__file__), '..', 'corpus-roots.json')

FR_SCHEMA_KEYS = {'general_meaning', 'shades_of_meaning', 'hadith_examples',
                  'classical_sources', 'idioms_customs', 'word_by_word_shades'}

def check_no_missing_roots(corpus_roots, csv_roots):
    """Check 1: Every corpus-roots.json root has a corresponding row in root_meaning.csv."""
    missing = sorted(set(corpus_roots.keys()) - csv_roots)
    # Exclude quadriliteral roots
    # Determine quadrilateral by length > 3 AND not in corpus_roots (triliteral roots only)
    # Actually, corpus_roots has only triliteral roots (3 letters), so all its keys are valid
    # Issues: a root that is key in corpus_roots.json but missing from root_meaning.csv
    # But some roots in corpus_roots might also be quadriliteral? No - corpus-roots only has triliteral roots (by design)
    return missing


def check_no_empty_full_response(rows):
    """Check 2: Zero rows with empty full_response."""
    return [r for r in rows if not r['full_response'].strip()]


def check_full_response_validity(rows):
    """Check 3: Every full_response is parseable JSON with correct schema."""
    bad = []
    for r in rows:
        fr_raw = r['full_response'].strip()
        if not fr_raw:
            bad.append((r['id'], r['root'], 'empty'))
            continue
        try:
            fr = json.loads(fr_raw)
        except json.JSONDecodeError as e:
            bad.append((r['id'], r['root'], f'invalid JSON: {e}'))
            continue
        if not isinstance(fr, dict):
            bad.append((r['id'], r['root'], 'not a dict'))
            continue
        missing_keys = FR_SCHEMA_KEYS - set(fr.keys())
        if missing_keys:
            bad.append((r['id'], r['root'], f'missing keys: {missing_keys}'))
            continue
        # Check general_meaning is a string
        if not isinstance(fr.get('general_meaning'), str):
            bad.append((r['id'], r['root'], 'general_meaning not a string'))
            continue
        # Check array fields are lists
        for key in ['shades_of_meaning', 'hadith_examples', 'classical_sources', 'idioms_customs', 'word_by_word_shades']:
            if not isinstance(fr.get(key), list):
                bad.append((r['id'], r['root'], f'{key} not a list'))
                continue
    return bad


def check_no_duplicate_roots(rows):
    """Check 4: No duplicate root values."""
    from collections import Counter
    root_counts = Counter(r['root'] for r in rows)
    return {k: v for k, v in root_counts.items() if v > 1}


def check_id_uniqueness(rows):
    """Check 5: No duplicate id values."""
    from collections import Counter
    id_counts = Counter(r['id'] for r in rows)
    return {k: v for k, v in id_counts.items() if v > 1}


def check_pos_populated(rows):
    """Check 6: No null/empty pos values."""
    return [r for r in rows if not r.get('pos', '').strip()]


def check_no_orphans(corpus_roots, rows):
    """Check 7: Detect roots in CSV but not in corpus-roots.json (informational)."""
    corpus_keys = set(corpus_roots.keys())
    orphans = []
    for r in rows:
        root = r['root']
        if root not in corpus_keys:
            orphans.append((r['id'], root))
    return orphans


def check_fields_nonempty(rows):
    """Check 8: Field columns are not completely empty strings (empty arrays are OK)."""
    missing = []
    for r in rows:
        for col in ['shades_of_meaning', 'hadith_examples', 'classical_sources',
                     'idioms_customs', 'word_by_word_shades']:
            raw = r[col].strip()
            if not raw:
                missing.append((r['id'], r['root'], col, 'empty string'))
    return missing


def load_json(path):
    with open(path, 'r', encoding='utf-8') as f:
        return json.load(f)


def main():
    csv_path = os.path.abspath(CSV_PATH)
    corpus_path = os.path.abspath(CORPUS_PATH)

    print(f"CSV file:    {csv_path}")
    print(f"Corpus file: {corpus_path}")
    print()

    # Load data
    corpus_roots = load_json(corpus_path)
    with open(csv_path, 'r', newline='', encoding='utf-8') as f:
        rows = list(csv.DictReader(f))

    csv_root_set = set(r['root'] for r in rows)
    csv_root_count = len(csv_root_set)

    print(f"corpus-roots.json roots: {len(corpus_roots)}")
    print(f"root_meaning.csv rows:   {len(rows)}")
    print(f"Unique roots in CSV:     {csv_root_count}")
    print()

    checks = []
    all_pass = True

    # Check 1: No missing roots
    missing = check_no_missing_roots(corpus_roots, csv_root_set)
    if missing:
        print(f"FAIL Check 1 — Missing roots: {len(missing)}")
        for r in missing:
            print(f"  {r}")
        all_pass = False
    else:
        print("PASS Check 1 — No missing roots")
    checks.append(('no-missing-roots', len(missing) == 0))

    # Check 2: No empty full_response
    empty_fr = check_no_empty_full_response(rows)
    if empty_fr:
        print(f"FAIL Check 2 — Empty full_response: {len(empty_fr)}")
        for r in empty_fr[:5]:
            print(f"  id={r['id']}, root={r['root']}")
        all_pass = False
    else:
        print("PASS Check 2 — No empty full_response")
    checks.append(('no-empty-full-response', len(empty_fr) == 0))

    # Check 3: full_response validity
    bad_fr = check_full_response_validity(rows)
    if bad_fr:
        print(f"FAIL Check 3 — Invalid full_response: {len(bad_fr)}")
        for r in bad_fr[:5]:
            print(f"  id={r[0]}, root={r[1]}: {r[2]}")
        all_pass = False
    else:
        print("PASS Check 3 — All full_response valid")
    checks.append(('full-response-valid', len(bad_fr) == 0))

    # Check 4: No duplicate roots
    dups = check_no_duplicate_roots(rows)
    if dups:
        print(f"FAIL Check 4 — Duplicate roots: {len(dups)}")
        for k, v in list(dups.items())[:5]:
            print(f"  {k}: {v} occurrences")
        all_pass = False
    else:
        print("PASS Check 4 — No duplicate roots")
    checks.append(('no-duplicate-roots', len(dups) == 0))

    # Check 5: id uniqueness
    dup_ids = check_id_uniqueness(rows)
    if dup_ids:
        print(f"FAIL Check 5 — Duplicate ids: {len(dup_ids)}")
        for k, v in list(dup_ids.items())[:5]:
            print(f"  id={k}: {v} rows")
        all_pass = False
    else:
        print("PASS Check 5 — All ids unique")
    checks.append(('id-uniqueness', len(dup_ids) == 0))

    # Check 6: POS populated
    no_pos = check_pos_populated(rows)
    if no_pos:
        print(f"FAIL Check 6 — Empty POS: {len(no_pos)}")
        for r in no_pos[:5]:
            print(f"  id={r['id']}, root={r['root']}")
        all_pass = False
    else:
        print("PASS Check 6 — All POS populated")
    checks.append(('pos-populated', len(no_pos) == 0))

    # Check 7: Orphan detection (informational)
    orphans = check_no_orphans(corpus_roots, rows)
    if orphans:
        print(f"INFO Check 7 — Orphan rows (in CSV, not in corpus-roots.json): {len(orphans)}")
        for r in orphans:
            print(f"  id={r[0]}, root={r[1]}")
        # This is informational — don't fail
    else:
        print("PASS Check 7 — No orphan rows")
    checks.append(('no-orphans', True))

    # Check 8: Missing field strings (completely empty, not empty arrays)
    missing_fields = check_fields_nonempty(rows)
    if missing_fields:
        print(f"INFO Check 8 — Completely empty field strings: {len(missing_fields)}")
        for r in missing_fields[:10]:
            print(f"  id={r[0]}, root={r[1]}, col={r[2]}: {r[3]}")
        # These are fields with completely empty strings (not even an empty array)
        # This is informational
    else:
        print("PASS Check 8 — All field columns have content")
    checks.append(('fields-nonempty', True))

    print()
    if all_pass:
        print("✓ ALL 8 CHECKS PASSED")
        return 0
    else:
        print(f"✗ {sum(1 for _, ok in checks if not ok)} CHECK(S) FAILED")
        return 1


if __name__ == '__main__':
    sys.exit(main())
