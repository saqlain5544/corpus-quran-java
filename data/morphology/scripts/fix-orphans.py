#!/usr/bin/env python3
"""
Fix orphan roots: correct Buckwalter encoding errors and Quran references
using only corpus.quran.com and standard Tanzil Quran text as authoritative sources.
"""
import csv, json, os

CSV_PATH = os.path.join(os.path.dirname(__file__), '..', 'root_meaning.csv')

def main():
    with open(CSV_PATH, 'r', newline='', encoding='utf-8') as f:
        rows = list(csv.DictReader(f))

    fieldnames = list(rows[0].keys())
    fixes = []

    for row in rows:
        rid = row['id']
        root = row['root']

        # 1. Ems → Ams (1660): E=ع should be A=أ for yesterday
        if rid == '1660' and root == 'Ems':
            row['root'] = 'Ams'
            row['root_arabic'] = 'أمس'
            # Fix references: Tanzil shows 28:18, 28:19, 28:82
            row['general_meaning'] = (
                'The word أمس (ams) means "yesterday" in Arabic. '
                'In the Quran, it appears in the phrase بِالْأَمْسِ (bil-amsi = "yesterday") '
                'at three verses (28:18, 28:19, 28:82), used to refer to the recent past. '
                'In 28:18-19, Moses is reminded of his action "yesterday" (killing the Egyptian). '
                'In 28:82, those who had envied Qarun "yesterday" now realize their error. '
                'The word functions as a time adverb and is not from a productive triliteral root.'
            )
            # Keep existing full_response but fix root
            row['full_response'] = _reconstruct_full_response(row)
            fixes.append(('Ems', 'Ams', 'E(ع)→A(أ), refs 28:18/19/82'))

        # 2. brxs → brzx (1735): x↔z swapped for barzakh
        if rid == '1735' and root == 'brxs':
            row['root'] = 'brzx'
            row['root_arabic'] = 'برزخ'
            row['general_meaning'] = (
                'The triliteral root B-R-Z-KH (ب-ر-ز-خ) denotes a barrier, partition, '
                'or isthmus between two things. In the Quran, it appears three times: '
                'as the barrier between fresh and salt water (25:53, 55:20), and as the '
                'intermediate realm (barzakh) between death and resurrection (23:100). '
                'The word describes a separating partition that prevents mixing or crossing.'
            )
            row['full_response'] = _reconstruct_full_response(row)
            fixes.append(('brxs', 'brzx', 'x↔z swapped, Arabic برزخ, refs 23:100/25:53/55:20'))

        # 3. zrbE → zrby (1764): E(ع)→y(ي) for zarabi carpets
        if rid == '1764' and root == 'zrbE':
            row['root'] = 'zrby'
            row['root_arabic'] = 'زربي'
            row['general_meaning'] = (
                'The root Z-R-B-Y (ز-ر-ب-ي) refers to luxurious carpets or rugs. '
                'Its broken plural زَرَابِيُّ (zarābiyyu) appears in Quran 88:16 describing '
                'the furnishings of Paradise: "and carpets spread out" (وَزَرَابِيُّ مَبْثُوثَةٌ). '
                'The word denotes richly decorated floor coverings, symbolizing divine hospitality '
                'and comfort in the Hereafter.'
            )
            row['full_response'] = _reconstruct_full_response(row)
            fixes.append(('zrbE', 'zrby', 'E(ع)→y(ي), Arabic زربي, ref 88:16'))

        # 4. qyt → yqwt (1700): root doesn't match word ياقوت (ruby)
        if rid == '1700' and root == 'qyt':
            row['root'] = 'yqwt'
            row['root_arabic'] = 'ياقوت'
            row['general_meaning'] = (
                'The word يَاقُوت (yāqūt) means "ruby" or "hyacinth" (a precious gemstone). '
                'It appears once in the Quran at 55:58 describing the maidens of Paradise: '
                '"as if they were rubies and coral" (كَأَنَّهُنَّ الْيَاقُوتُ وَالْمَرْجَانُ). '
                'The word is of Persian origin and does not derive from a standard Arabic triliteral root.'
            )
            row['full_response'] = _reconstruct_full_response(row)
            fixes.append(('qyt', 'yqwt', 'Root changed to match Arabic ياقوت, ref 55:58'))

        # 5. bErf → bryq (1732): root doesn't match word أباريق (pitchers)
        if rid == '1732' and root == 'bErf':
            row['root'] = 'bryq'
            row['root_arabic'] = 'إبريق'
            row['general_meaning'] = (
                'The word إِبْرِيق (ibrīq) means "pitcher" or "jug with a handle and spout." '
                'Its broken plural أَبَارِيق (abārīq) appears in Quran 56:18 as vessels in Paradise: '
                '"with goblets and pitchers and a cup from a flowing spring" '
                '(بِأَكْوَابٍ وَأَبَارِيقَ وَكَأْسٍ مِّن مَّعِينٍ). '
                'The word is a Persian loanword (from āb-rēz, "water-pourer") and does not have '
                'a standard Arabic triliteral root.'
            )
            row['full_response'] = _reconstruct_full_response(row)
            fixes.append(('bErf', 'bryq', 'Root changed to match Arabic إبريق, ref 56:18'))

    if not fixes:
        print("No fixes needed.")
        return

    with open(CSV_PATH, 'w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, lineterminator='\r\n')
        writer.writeheader()
        writer.writerows(rows)

    print(f"Applied {len(fixes)} fix(es):")
    for old, new, desc in fixes:
        print(f"  {old} → {new}: {desc}")


def _reconstruct_full_response(row):
    """Build full_response from field columns."""
    fr = {}
    for key in ['general_meaning', 'shades_of_meaning', 'hadith_examples',
                'classical_sources', 'idioms_customs', 'word_by_word_shades']:
        val = row[key].strip()
        if key == 'general_meaning':
            fr[key] = val
        else:
            try:
                fr[key] = json.loads(val) if val else []
            except json.JSONDecodeError:
                fr[key] = []
    return json.dumps(fr, ensure_ascii=False)


if __name__ == '__main__':
    main()
