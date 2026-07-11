#!/usr/bin/env python3
"""
Validate meanings-roots-ai.jsonl against corpus-roots.json.

Eight integrity checks:
  1. No missing roots — every corpus-roots.json root has an entry in meanings-roots-ai.jsonl
  2. No empty en_detailed — zero entries with empty en_detailed
  3. JSONL validity — every line is parseable JSON with required meaning fields
  4. No duplicate roots — no duplicate root values
  5. id uniqueness — no duplicate id values
  6. POS populated — no null/empty pos values
  7. Orphan roots detected — roots in JSONL but not in corpus-roots.json (informational)
  8. Occurrence count mismatch — corpus-roots.json vs JSONL occurrence counts
"""

import json
import os
import sys
from collections import Counter

JSONL_PATH = os.path.join(os.path.dirname(__file__), '..', 'meanings-roots-ai.jsonl')
CORPUS_PATH = os.path.join(os.path.dirname(__file__), '..', 'corpus-roots.json')

REQUIRED_MEANING_KEYS = {'en_short', 'en', 'en_detailed', 'trilateral_meaning', 'ar_definition'}


def load_jsonl(path):
    """Load JSONL file into list of dicts."""
    entries = []
    with open(path, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            entries.append(json.loads(line))
    return entries


def load_json(path):
    with open(path, 'r', encoding='utf-8') as f:
        return json.load(f)


def check_no_missing_roots(corpus_roots, jsonl_root_set):
    """Check 1: Every corpus-roots.json root has an entry in meanings-roots-ai.jsonl."""
    return sorted(set(corpus_roots.keys()) - jsonl_root_set)


def check_no_empty_en_detailed(entries):
    """Check 2: Zero entries with empty en_detailed."""
    return [(e['id'], e['root']) for e in entries
            if not e.get('meaning', {}).get('en_detailed', '').strip()]


def check_meaning_field_completeness(entries):
    """Check 3: Every entry has all required meaning fields (non-empty)."""
    bad = []
    for e in entries:
        meaning = e.get('meaning', {})
        for key in REQUIRED_MEANING_KEYS:
            if not meaning.get(key, '').strip():
                bad.append((e['id'], e['root'], f'missing/empty meaning.{key}'))
    return bad


def check_no_duplicate_roots(entries):
    """Check 4: No duplicate root values."""
    root_counts = Counter(e['root'] for e in entries)
    return {k: v for k, v in root_counts.items() if v > 1}


def check_id_uniqueness(entries):
    """Check 5: No duplicate id values."""
    id_counts = Counter(e['id'] for e in entries)
    return {k: v for k, v in id_counts.items() if v > 1}


def check_pos_populated(entries):
    """Check 6: No null/empty pos values."""
    return [(e['id'], e['root']) for e in entries if not e.get('pos', '').strip()]


def check_no_orphans(corpus_roots, entries):
    """Check 7: Detect roots in JSONL but not in corpus-roots.json (informational)."""
    corpus_keys = set(corpus_roots.keys())
    return [(e['id'], e['root']) for e in entries if e['root'] not in corpus_keys]


def check_occurrence_counts(corpus_roots, entries):
    """Check 8: Cross-reference occurrence counts between corpus-roots.json and JSONL."""
    mismatches = []
    corpus_keys = set(corpus_roots.keys())
    for e in entries:
        if e['root'] not in corpus_keys:
            continue
        corpus_count = len(corpus_roots[e['root']])
        jsonl_count = e.get('occurrences_quran', 0)
        if corpus_count != jsonl_count:
            mismatches.append((e['id'], e['root'], corpus_count, jsonl_count))
    return mismatches


def main():
    jsonl_path = os.path.abspath(JSONL_PATH)
    corpus_path = os.path.abspath(CORPUS_PATH)

    print(f"JSONL file:  {jsonl_path}")
    print(f"Corpus file: {corpus_path}")
    print()

    # Load data
    corpus_roots = load_json(corpus_path)
    entries = load_jsonl(jsonl_path)

    jsonl_root_set = set(e['root'] for e in entries)

    print(f"corpus-roots.json roots:        {len(corpus_roots)}")
    print(f"meanings-roots-ai.jsonl entries: {len(entries)}")
    print(f"Unique roots in JSONL:           {len(jsonl_root_set)}")
    print()

    all_pass = True

    # Check 1: No missing roots
    missing = check_no_missing_roots(corpus_roots, jsonl_root_set)
    if missing:
        print(f"FAIL Check 1 — Missing roots ({len(missing)}): {', '.join(missing[:20])}")
        all_pass = False
    else:
        print("PASS Check 1 — No missing roots")

    # Check 2: No empty en_detailed
    empty_ed = check_no_empty_en_detailed(entries)
    if empty_ed:
        print(f"FAIL Check 2 — Empty en_detailed: {len(empty_ed)}")
        for id_, root in empty_ed[:10]:
            print(f"  id={id_}, root={root}")
        all_pass = False
    else:
        print("PASS Check 2 — No empty en_detailed")

    # Check 3: Meaning field completeness
    bad_fr = check_meaning_field_completeness(entries)
    if bad_fr:
        print(f"FAIL Check 3 — Missing meaning fields: {len(bad_fr)}")
        for id_, root, msg in bad_fr[:10]:
            print(f"  id={id_}, root={root}: {msg}")
        all_pass = False
    else:
        print("PASS Check 3 — All required meaning fields present")

    # Check 4: No duplicate roots
    dups = check_no_duplicate_roots(entries)
    if dups:
        print(f"FAIL Check 4 — Duplicate roots: {len(dups)}")
        for k, v in list(dups.items())[:5]:
            print(f"  {k}: {v} occurrences")
        all_pass = False
    else:
        print("PASS Check 4 — No duplicate roots")

    # Check 5: id uniqueness
    dup_ids = check_id_uniqueness(entries)
    if dup_ids:
        print(f"FAIL Check 5 — Duplicate ids: {len(dup_ids)}")
        for k, v in list(dup_ids.items())[:5]:
            print(f"  id={k}: {v} entries")
        all_pass = False
    else:
        print("PASS Check 5 — All ids unique")

    # Check 6: POS populated
    no_pos = check_pos_populated(entries)
    if no_pos:
        print(f"FAIL Check 6 — Empty POS: {len(no_pos)}")
        for id_, root in no_pos[:10]:
            print(f"  id={id_}, root={root}")
        all_pass = False
    else:
        print("PASS Check 6 — All POS populated")

    # Check 7: Orphan detection (informational)
    orphans = check_no_orphans(corpus_roots, entries)
    if orphans:
        print(f"INFO Check 7 — Orphan roots (in JSONL, not in corpus-roots.json): {len(orphans)}")
        for id_, root in orphans[:10]:
            print(f"  id={id_}, root={root}")
    else:
        print("PASS Check 7 — No orphan roots")

    # Check 8: Occurrence count mismatches
    mismatches = check_occurrence_counts(corpus_roots, entries)
    if mismatches:
        print(f"FAIL Check 8 — Occurrence count mismatches: {len(mismatches)}")
        for id_, root, cc, jc in sorted(mismatches, key=lambda x: abs(x[2]-x[3]), reverse=True)[:15]:
            print(f"  id={id_}, root={root}: corpus={cc}, JSONL={jc}, diff={abs(cc-jc)}")
        all_pass = False
    else:
        print("PASS Check 8 — All occurrence counts match")

    print()
    if all_pass:
        print("\u2713 ALL 8 CHECKS PASSED")
        return 0
    else:
        failed = sum(1 for label in [
            bool(missing), bool(empty_ed), bool(bad_fr),
            bool(dups), bool(dup_ids), bool(no_pos), False, bool(mismatches)
        ] if label)
        print(f"\u2717 {failed} CHECK(S) FAILED")
        return 1


if __name__ == '__main__':
    sys.exit(main())
