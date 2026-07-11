#!/usr/bin/env python3
"""Build lemma-level concordance: root → lemmas → {total_occurrences, [surah:ayah]}"""
import csv, json, os
from collections import defaultdict

BASE = '/Users/saqlain/corpus-quran-java/data/morphology'

# Load corpus-roots
with open(f'{BASE}/corpus-roots.json') as f:
    corpus = json.load(f)

# Build MASAQ index: "surah:ayah:word" → set of stem/lemma forms
word_lemma = {}  # key -> set of lemma forms
with open(f'{BASE}/MASAQ.csv', encoding='utf-8') as f:
    for row in csv.DictReader(f):
        key = f"{row['Sura_No']}:{row['Verse_No']}:{row['Word_No']}"
        mt = row.get('Morph_Type', '')
        if mt == 'Stem' and row['Segmented_Word']:
            word_lemma.setdefault(key, set()).add(row['Segmented_Word'])

print(f'MASAQ word positions: {len(word_lemma)}')

# Build concordance
total = len(corpus)
out_path = f'{BASE}/concordance.jsonl'
with open(out_path, 'w') as out:
    for idx, (root, refs) in enumerate(sorted(corpus.items())):
        lemma_data = defaultdict(list)
        
        for ref in refs:
            parts = ref.split(':')
            if len(parts) < 3:
                continue
            ayah = f'{parts[0]}:{parts[1]}'
            wkey = f'{parts[0]}:{parts[1]}:{parts[2]}'
            
            lemmas = word_lemma.get(wkey, set())
            if lemmas:
                for lemma in sorted(lemmas):
                    lemma_data[lemma].append(ayah)
            else:
                lemma_data[root].append(ayah)
        
        # Deduplicate and build output
        lemmas_out = {}
        for lemma, ayahs in sorted(lemma_data.items()):
            unique = sorted(set(ayahs),
                          key=lambda x: (int(x.split(':')[0]), int(x.split(':')[1])))
            lemmas_out[lemma] = {
                'total_occurrences': len(unique),
                'occurrences': unique,
            }
        
        entry = {'root': root, 'lemmas': lemmas_out}
        out.write(json.dumps(entry, ensure_ascii=False) + '\n')
        
        if (idx + 1) % 400 == 0:
            print(f'  {idx+1}/{total}')

# Verify
with open(out_path) as f:
    lines = [json.loads(line) for line in f if line.strip()]

total_lemmas = sum(len(e['lemmas']) for e in lines)
total_occ = sum(sum(l['total_occurrences'] for l in e['lemmas'].values()) for e in lines)
size = os.path.getsize(out_path)

print(f'\n--- CONCORDANCE ---')
print(f'Roots: {len(lines)}')
print(f'Total lemmas: {total_lemmas}')
print(f'Total ayah occurrences (deduped): {total_occ}')
print(f'File size: {size/1024/1024:.1f} MB')

# Samples
for e in lines:
    r = e['root']
    if r in ['qwl', 'kwn', 'hjd', 'xnq']:
        n_lemmas = len(e['lemmas'])
        print(f'\n[{r}] {n_lemmas} lemmas:')
        items = sorted(e['lemmas'].items(), key=lambda x: -x[1]['total_occurrences'])
        for lemma, data in items[:5]:
            cnt = data['total_occurrences']
            occs = data['occurrences'][:3]
            print(f'  {lemma}: {cnt} occ -> {occs}...')
