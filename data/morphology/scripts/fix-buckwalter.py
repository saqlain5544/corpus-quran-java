#!/usr/bin/env python3
"""
Fix Buckwalter/Arabic encoding errors found during cross-validation.

Errors found:
  id=1785, root='SbA',  root_arabic='سبأ' → root='SbA' is ص-ب-أ (Sabians), not س-ب-أ (Sheba)
  id=1787, root='lwt',  root_arabic='لوط' → root='lwt' is ل-و-ت (neg. particle), not ل-و-ط (Lot)
  id=1791, root='zqm',  root_arabic='زقوم' → root='zqm' is ز-ق-م (3-letter root), Arabic should be 'زقم'
  id=1778, root='qTyn', root_arabic='يقطن' → Buckwalter should be 'yqTn' (ي-ق-ط-ن), not 'qTyn' (ق-ط-ي-ن)
"""

import csv
import json
import os

CSV_PATH = os.path.join(os.path.dirname(__file__), '..', 'root_meaning.csv')


def fix_sba(row):
    """Fix SbA: ص-ب-أ (Sabians), not س-ب-أ (Sheba)."""
    row['root_arabic'] = 'صبأ'
    row['general_meaning'] = (
        'The triliteral root Ṣ-B-A (ص-ب-أ) refers to the Sabians (al-Ṣābi\'ūn), '
        'a religious group mentioned in the Quran alongside believers, Jews, Christians, '
        'and Magians. The Sabians are traditionally identified as a monotheistic or '
        'semi-monotheistic community, possibly the Mandaeans of Mesopotamia or a related '
        'group. The root is exclusively used in the Quran to designate this religious community, '
        'and God promises that those among them who believe in Him and the Last Day and do '
        'righteousness shall have their reward.'
    )
    row['shades_of_meaning'] = json.dumps([
        {"shade": "Sabians among the religious communities", "reference": "2:62:7", "arabic_word": "وَالصَّابِئِينَ", "explanation": "The Sabians are listed among believers, Jews, and Christians — whoever believes in God and the Last Day and does righteousness will have their reward with their Lord."},
        {"shade": "Sabians in the assurance of reward", "reference": "5:69:6", "arabic_word": "وَالصَّابِئُونَ", "explanation": "Repeats the same divine assurance: believers, Jews, Sabians, and Christians — all who believe and act righteously need not fear."},
        {"shade": "Sabians among the judged communities", "reference": "22:17:8", "arabic_word": "وَالْمَجُوسَ", "explanation": "God will judge between believers, Jews, Sabians, Christians, Magians, and polytheists on the Day of Resurrection — here in the context of Sabians alongside Magians."},
    ], ensure_ascii=False)
    row['hadith_examples'] = json.dumps([
        {"arabic": "عَنِ ابْنِ عَبَّاسٍ رَضِيَ اللَّهُ عَنْهُمَا فِي قَوْلِهِ: «الصَّابِئِينَ» قَالَ: هُمْ قَوْمٌ مَا بَيْنَ الْيَهُودِ وَالْمَجُوسِ", "translation": "Ibn Abbas said regarding 'the Sabians': They are a people between the Jews and the Magians."},
    ], ensure_ascii=False)
    row['classical_sources'] = json.dumps([
        {"arabic": "قَالَ ابْنُ فَارِسٍ: الصَّادُ وَالْبَاءُ وَالْهَمْزَةُ أَصْلٌ، وَالصَّابِئُ: الَّذِي يَخْرُجُ مِنْ دِينٍ إِلَى دِينٍ", "translation": "Ibn Faris said: The letters ṣād, bā\', and hamzah: al-ṣābi\' is one who leaves one religion for another."},
        {"arabic": "قَالَ الْفَرَّاءُ: الصَّابِئُونَ قَوْمٌ يُشْبِهُونَ النَّصَارَى فِي شَيْءٍ وَالْيَهُودَ فِي شَيْءٍ", "translation": "Al-Farrāʾ said: The Sabians are a people who resemble Christians in some ways and Jews in others."},
    ], ensure_ascii=False)
    row['idioms_customs'] = json.dumps([
        {"arabic": "الصَّابِئُونَ / الصَّابِئِينَ", "translation": "The Sabians — a religious group, possibly Mandaeans or a gnostic monotheistic community."},
        {"arabic": "صَبَأَ", "translation": "To apostatize or change religion; also said of a person who adopts the Sabian faith."},
    ], ensure_ascii=False)
    row['word_by_word_shades'] = json.dumps([
        {"shade": "Gentilic plural, genitive/accusative (ٱلصَّـٰبِـِٔينَ) — the Sabians", "reference": "2:62:7", "arabic_word": "ٱلصَّـٰبِـِٔينَ", "explanation": "Definite plural noun, inflected form, referring to the Sabian community."},
        {"shade": "Gentilic plural, nominative (ٱلصَّـٰبِـُٔونَ)", "reference": "5:69:6", "arabic_word": "ٱلصَّـٰبِـُٔونَ", "explanation": "Nominative case form of the same gentilic plural."},
    ], ensure_ascii=False)
    row['full_response'] = reconstruct_full_response(row)
    return row


def fix_lwt(row):
    """Fix lwt: ل-و-ت (the negative particle 'لات'), not ل-و-ط (prophet Lot)."""
    row['root_arabic'] = 'لوت'
    row['general_meaning'] = (
        'The triliteral root L-W-T (ل-و-ت) in the Quran is associated with the word '
        'لات (lāta), a rare negative particle meaning "not" or "not at all", appearing '
        'in the phrase وَلَاتَ حِينَ مَنَاصٍ ("but there was no time for escape") at '
        '38:3. Classical grammarians differ on its derivation — some analyze it as a '
        'combination of لَا (negation) and تَ (an emphatic suffix), while others trace '
        'it to a root ل-و-ت with the sense of deficiency or cessation.'
    )
    row['shades_of_meaning'] = json.dumps([
        {"shade": "Negative particle of time — 'not'", "reference": "38:3:8", "arabic_word": "وَلَاتَ", "explanation": "In the phrase 'وَلَاتَ حِينَ مَنَاصٍ' — 'but there was no time for escape' — the word لات serves as a specialized negation operating specifically on temporal nouns."},
    ], ensure_ascii=False)
    row['hadith_examples'] = json.dumps([
        {"arabic": "عَنِ ابْنِ عَبَّاسٍ فِي قَوْلِهِ تَعَالَى: «وَلَاتَ حِينَ مَنَاصٍ» قَالَ: لَيْسَ حِينَ فِرَارٍ", "translation": "Ibn Abbas said regarding 'there was no time for escape': It is not the time for fleeing."},
    ], ensure_ascii=False)
    row['classical_sources'] = json.dumps([
        {"arabic": "قَالَ سِيبَوَيْهِ: لَاتَ مَعْنَاهَا لَيْسَ، وَهِيَ لَا زِيدَتْ فِيهَا التَّاءُ", "translation": "Sibawayh said: Lāta means laysa (is not), and it is lā with an extra tā\'."},
        {"arabic": "قَالَ الزَّجَّاجُ: لَاتَ أَصْلُهَا لَا وَالتَّاءُ تَأْكِيدٌ", "translation": "Al-Zajjāj said: The origin of lāta is lā and the tā\' is for emphasis."},
    ], ensure_ascii=False)
    row['idioms_customs'] = json.dumps([
        {"arabic": "وَلَاتَ", "translation": "wa-lāta — 'but not', 'and there was no' — a compound negation used before time words."},
    ], ensure_ascii=False)
    row['word_by_word_shades'] = json.dumps([
        {"shade": "Negative particle (لَاتَ) — 'not'", "reference": "38:3:8", "arabic_word": "وَلَاتَ", "explanation": "A specialized negative particle used before temporal nouns; prefixed with wa- (and) in this occurrence."},
    ], ensure_ascii=False)
    row['full_response'] = reconstruct_full_response(row)
    return row


def fix_zqm(row):
    """Fix zqm: Arabic root is زقم (3-letter), not زقوم (derived word)."""
    row['root_arabic'] = 'زقم'
    # Content stays the same — it was already correct about Zaqqum
    row['full_response'] = reconstruct_full_response(row)
    return row


def fix_qtyn(row):
    """Fix qTyn: Buckwalter should be yqTn (يقطن), not qTyn (قطين)."""
    row['root'] = 'yqTn'
    row['full_response'] = reconstruct_full_response(row)
    return row


def reconstruct_full_response(row):
    """Build full_response JSON from the 6 field columns."""
    fr = {}
    keys = ['general_meaning', 'shades_of_meaning', 'hadith_examples',
            'classical_sources', 'idioms_customs', 'word_by_word_shades']
    for key in keys:
        val = row[key].strip()
        if key == 'general_meaning':
            fr[key] = val
        else:
            try:
                fr[key] = json.loads(val) if val else []
            except json.JSONDecodeError:
                fr[key] = []
    return json.dumps(fr, ensure_ascii=False)


def main():
    print(f"Reading: {CSV_PATH}")
    with open(CSV_PATH, 'r', newline='', encoding='utf-8') as f:
        rows = list(csv.DictReader(f))

    fieldnames = list(rows[0].keys())
    fixes = []

    for row in rows:
        root = row['root']
        rid = row['id']

        if root == 'SbA' and rid == '1785':
            fix_sba(row)
            fixes.append((rid, root, 'SbA → fixed Sabians root'))
        elif root == 'lwt' and rid == '1787':
            fix_lwt(row)
            fixes.append((rid, root, 'lwt → fixed negative particle root'))
        elif root == 'zqm' and rid == '1791':
            fix_zqm(row)
            fixes.append((rid, root, 'zqm → fixed root_arabic to زقم'))
        elif root == 'qTyn' and rid == '1778':
            old_root = row['root']
            fix_qtyn(row)
            fixes.append((rid, old_root, f'qTyn → {row["root"]}'))

    if not fixes:
        print("No fixes needed.")
        return

    print(f"\nWriting {len(fixes)} fix(es) to CSV...")
    with open(CSV_PATH, 'w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, lineterminator='\r\n')
        writer.writeheader()
        writer.writerows(rows)

    print("\n=== Applied fixes ===")
    for rid, old_root, desc in fixes:
        print(f"  id={rid}, root={old_root}: {desc}")

    print("\nDone!")


if __name__ == '__main__':
    main()
