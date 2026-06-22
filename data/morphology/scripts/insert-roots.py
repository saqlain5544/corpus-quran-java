#!/usr/bin/env python3
"""
Insert roots into root_meaning.csv.

Two modes:
  --backfill-full-response  Reconstruct full_response JSON from field columns
                             for rows where it is empty or invalid JSON.
  --add-missing             Insert rows for the 8 roots present in
                             corpus-roots.json but missing from root_meaning.csv.
"""

import csv
import json
import sys
import os

CSV_PATH = os.path.join(os.path.dirname(__file__), '..', 'root_meaning.csv')
CORPUS_PATH = os.path.join(os.path.dirname(__file__), '..', 'corpus-roots.json')

FIELD_COLS = [
    'general_meaning',
    'shades_of_meaning',
    'hadith_examples',
    'classical_sources',
    'idioms_customs',
    'word_by_word_shades',
]
FR_KEYS = FIELD_COLS  # full_response has the same 6 keys

# Roots missing from root_meaning.csv but present in corpus-roots.json
MISSING_ROOTS = {
    'SbA': {
        'root_arabic': 'صبأ',
        'general_meaning': 'The triliteral root Ṣ-B-A (ص-ب-أ) refers to the Sabians (al-Ṣābi\u2019ūn), a religious group mentioned in the Quran alongside believers, Jews, Christians, and Magians. The Sabians are traditionally identified as a monotheistic or semi-monotheistic community, possibly the Mandaeans of Mesopotamia or a related group. The root is exclusively used in the Quran to designate this religious community, and God promises that those among them who believe in Him and the Last Day and do righteousness shall have their reward.',
        'shades_of_meaning': json.dumps([
            {"shade": "Sabians among the religious communities", "reference": "2:62:7", "arabic_word": "وَالصَّابِئِينَ", "explanation": "The Sabians are listed among believers, Jews, and Christians \u2014 whoever believes in God and the Last Day and does righteousness will have their reward with their Lord."},
            {"shade": "Sabians in the assurance of reward", "reference": "5:69:6", "arabic_word": "وَالصَّابِئُونَ", "explanation": "Repeats the same divine assurance: believers, Jews, Sabians, and Christians \u2014 all who believe and act righteously need not fear."},
            {"shade": "Sabians among the judged communities", "reference": "22:17:8", "arabic_word": "وَالْمَجُوسَ", "explanation": "God will judge between believers, Jews, Sabians, Christians, Magians, and polytheists on the Day of Resurrection \u2014 here in the context of Sabians alongside Magians."},
        ], ensure_ascii=False),
        'hadith_examples': json.dumps([
            {"arabic": "عَنِ ابْنِ عَبَّاسٍ رَضِيَ اللَّهُ عَنْهُمَا فِي قَوْلِهِ: «الصَّابِئِينَ» قَالَ: هُمْ قَوْمٌ مَا بَيْنَ الْيَهُودِ وَالْمَجُوسِ", "translation": "Ibn Abbas said regarding 'the Sabians': They are a people between the Jews and the Magians."},
        ], ensure_ascii=False),
        'classical_sources': json.dumps([
            {"arabic": "قَالَ ابْنُ فَارِسٍ: الصَّادُ وَالْبَاءُ وَالْهَمْزَةُ أَصْلٌ، وَالصَّابِئُ: الَّذِي يَخْرُجُ مِنْ دِينٍ إِلَى دِينٍ", "translation": "Ibn Faris said: The letters ṣād, bā\u2019, and hamzah: al-ṣābi\u2019 is one who leaves one religion for another."},
            {"arabic": "قَالَ الْفَرَّاءُ: الصَّابِئُونَ قَوْمٌ يُشْبِهُونَ النَّصَارَى فِي شَيْءٍ وَالْيَهُودَ فِي شَيْءٍ", "translation": "Al-Farrā\u2019 said: The Sabians are a people who resemble Christians in some ways and Jews in others."},
        ], ensure_ascii=False),
        'idioms_customs': json.dumps([
            {"arabic": "الصَّابِئُونَ / الصَّابِئِينَ", "translation": "The Sabians \u2014 a religious group, possibly Mandaeans or a gnostic monotheistic community."},
            {"arabic": "صَبَأَ", "translation": "To apostatize or change religion; also said of a person who adopts the Sabian faith."},
        ], ensure_ascii=False),
        'word_by_word_shades': json.dumps([
            {"shade": "Gentilic plural, genitive/accusative (\u0671\u0644\u0635\u064E\u0651\u0627\u0628\u0650\u0650\u0654\u064A\u0646\u064E) \u2014 the Sabians", "reference": "2:62:7", "arabic_word": "\u0671\u0644\u0635\u064E\u0651\u0627\u0628\u0650\u0650\u0654\u064A\u0646\u064E", "explanation": "Definite plural noun, inflected form, referring to the Sabian community."},
            {"shade": "Gentilic plural, nominative (\u0671\u0644\u0635\u064E\u0651\u0627\u0628\u0650\u0650\u0654\u0648\u0646\u064E)", "reference": "5:69:6", "arabic_word": "\u0671\u0644\u0635\u064E\u0651\u0627\u0628\u0650\u0650\u0654\u0648\u0646\u064E", "explanation": "Nominative case form of the same gentilic plural."},
        ], ensure_ascii=False),
        'pos': 'N',
    },
    'zqm': {
        'root_arabic': 'زقم',
        'general_meaning': 'The quadriliteral root Z-Q-M (ز-ق-م) specifically denotes the Zaqqum tree, a horrid tree that grows in the depths of Hellfire, mentioned in the Quran as the food of the sinful. Its fruit is described as being like the heads of devils, and it boils in the bellies like molten brass. The root is exclusively used in the Quran for this infernal tree, serving as a powerful symbol of divine punishment.',
        'shades_of_meaning': json.dumps([
            {"shade": "The tree of Zaqqum as the food of sinners", "reference": "37:62:6", "arabic_word": "زَقُّومٍ", "explanation": "Allah challenges: 'Is that [Paradise] a better accommodation or the tree of Zaqqum?' — contrasting the bliss of Paradise with the punishment of Zaqqum."},
            {"shade": "Zaqqum described as a tree in Hellfire", "reference": "44:43:3", "arabic_word": "زَقُّومٍ", "explanation": "Emphatic declaration: 'Indeed, the tree of Zaqqum is the food of the sinful.'"},
            {"shade": "Eating from Zaqqum to fill bellies", "reference": "56:52:5", "arabic_word": "زَقُّومٍ", "explanation": "The disbelievers will 'be eating from the tree of Zaqqum, filling their bellies from it' — describing the forced consumption."},
        ], ensure_ascii=False),
        'hadith_examples': json.dumps([
            {"arabic": "عَنِ ابْنِ عَبَّاسٍ رَضِيَ اللَّهُ عَنْهُمَا قَالَ: قَالَ رَسُولُ اللَّهِ صَلَّى اللَّهُ عَلَيْهِ وَسَلَّمَ: «لَوْ أَنَّ قَطْرَةً مِنْ زَقُّومِ جَهَنَّمَ قُطِرَتْ فِي دُنْيَا لَأَفْسَدَتْ عَلَى أَهْلِ الدُّنْيَا مَعَايِشَهُمْ»", "translation": "Ibn Abbas narrated: The Prophet ﷺ said: 'If a single drop of the Zaqqum of Hell were to drip into this world, it would spoil all the livelihoods of the people of this world.' (Tirmidhi)"},
        ], ensure_ascii=False),
        'classical_sources': json.dumps([
            {"arabic": "قَالَ ابْنُ فَارِسٍ: الزَّايُ وَالْقَافُ وَالْمِيمُ لَيْسَ بِأَصْلٍ، وَالزَّقُّومُ: شَجَرَةٌ ذَاتُ سُمٍّ", "translation": "Ibn Faris said: The letters zāy, qāf, and mīm do not form a proper root; Zaqqum is a poisonous tree."},
            {"arabic": "قَالَ الْقُرْطُبِيُّ: الزَّقُّومُ شَجَرَةٌ خَبِيثَةٌ تَنْبُتُ فِي النَّارِ", "translation": "Al-Qurtubi said: Zaqqum is an evil tree that grows in the fire."},
        ], ensure_ascii=False),
        'idioms_customs': json.dumps([
            {"arabic": "زَقُّومٌ", "translation": "Zaqqum — the tree of Hell, a byword for extreme suffering and punishment."},
            {"arabic": "أَزْقَمَ", "translation": "A derived form meaning 'to gulp or swallow hastily' with difficulty, related to the root."},
        ], ensure_ascii=False),
        'word_by_word_shades': json.dumps([
            {"shade": "Noun, genitive (زَقُّومٍ) — tree of Zaqqum", "reference": "37:62:6", "arabic_word": "زَقُّومٍ", "explanation": "Indefinite genitive noun referring to the infernal tree."},
            {"shade": "Noun, genitive (زَقُّومٍ)", "reference": "44:43:3", "arabic_word": "زَقُّومٍ", "explanation": "Same noun in the construct phrase 'shajarata zaqqūm' (tree of Zaqqum)."},
            {"shade": "Noun, genitive (زَقُّومٍ)", "reference": "56:52:5", "arabic_word": "زَقُّومٍ", "explanation": "Part of the phrase 'min shajarati min zaqqūm' (from a tree of Zaqqum)."},
        ], ensure_ascii=False),
        'pos': 'N',
    },
    'Hqf': {
        'root_arabic': 'حقف',
        'general_meaning': 'The triliteral root Ḥ-Q-F (ح-ق-ف) refers to a curved sand dune or a winding tract of sand. In the Quran, it appears in the plural form al-Aḥqāf (الأحقاف), the name of Surah 46, which refers to the sand dunes region where the prophet Hūd was sent to the people of ʿĀd. The root evokes a desolate, wind-swept desert landscape.',
        'shades_of_meaning': json.dumps([
            {"shade": "Al-Ahqaf — the sand dunes, a geographic region", "reference": "46:21:7", "arabic_word": "ٱلْأَحْقَافِ", "explanation": "'And mention, [O Muḥammad], the brother of ʿĀd, when he warned his people in [the region of] al-Aḥqāf' — the name of the valley of sand dunes where the ʿĀd tribe lived."},
        ], ensure_ascii=False),
        'hadith_examples': json.dumps([
            {"arabic": "عَنِ ابْنِ عَبَّاسٍ رَضِيَ اللَّهُ عَنْهُمَا فِي قَوْلِهِ تَعَالَى: «وَاذْكُرْ أَخَا عَادٍ إِذْ أَنْذَرَ قَوْمَهُ بِالْأَحْقَافِ» قَالَ: الْأَحْقَافُ: جِبَالُ رَمْلٍ بِالشِّحْرِ", "translation": "Ibn Abbas commented on the verse: 'Al-Ahqaf is mountains of sand in al-Shihr (a coastal region in Yemen).'"},
        ], ensure_ascii=False),
        'classical_sources': json.dumps([
            {"arabic": "قَالَ ابْنُ فَارِسٍ: الْحَاءُ وَالْقَافُ وَالْفَاءُ أَصْلٌ وَاحِدٌ يَدُلُّ عَلَى اسْتِطَالَةٍ فِي انْعِرَاجٍ", "translation": "Ibn Faris said: The letters ḥāʾ, qāf, and fāʾ have one core meaning: elongation with curvature (i.e., long winding sand dunes)."},
            {"arabic": "قَالَ الْفَرَّاءُ: الْأَحْقَافُ: رَمْلٌ كَثِيرٌ مُسْتَطِيلٌ", "translation": "Al-Farrāʾ said: Al-Aḥqāf is abundant sand that extends at length."},
        ], ensure_ascii=False),
        'idioms_customs': json.dumps([
            {"arabic": "الْأَحْقَافُ", "translation": "Al-Ahqaf — the curved sand dunes; also the name of the 46th surah of the Quran."},
            {"arabic": "حِقْفٌ", "translation": "A single curved sand dune."},
        ], ensure_ascii=False),
        'word_by_word_shades': json.dumps([
            {"shade": "Definite plural noun (ٱلْأَحْقَافِ) — the sand dunes", "reference": "46:21:7", "arabic_word": "ٱلْأَحْقَافِ", "explanation": "Genitive plural definite noun, referring to the specific region of sand dunes where ʿĀd dwelt."},
        ], ensure_ascii=False),
        'pos': 'N',
    },
    'hdhd': {
        'root_arabic': 'هدهد',
        'general_meaning': 'The quadriliteral root H-D-H-D (ه-د-ه-د) denotes the hoopoe bird (Upupa epops), a distinctive crested bird known in Arabic folklore. In the Quran, the hoopoe appears in the story of Prophet Solomon (Sulaymān), serving as a scout who brings news of the kingdom of Sheba and its queen Bilqīs. The bird is portrayed as intelligent, observant, and faithful in service.',
        'shades_of_meaning': json.dumps([
            {"shade": "The hoopoe as Solomon's messenger and scout", "reference": "27:20:8", "arabic_word": "ٱلْهُدْهُدَ", "explanation": "Solomon inspected the birds and said: 'Why do I not see the hoopoe? Or is he among the absent?' — showing the hoopoe was part of his retinue."},
        ], ensure_ascii=False),
        'hadith_examples': json.dumps([
            {"arabic": "عَنِ ابْنِ عَبَّاسٍ قَالَ: نَهَى رَسُولُ اللَّهِ صَلَّى اللَّهُ عَلَيْهِ وَسَلَّمَ عَنْ قَتْلِ الْهُدْهُدِ", "translation": "Ibn Abbas narrated: The Messenger of Allah ﷺ forbade killing the hoopoe. (Abu Dawud)"},
        ], ensure_ascii=False),
        'classical_sources': json.dumps([
            {"arabic": "قَالَ ابْنُ فَارِسٍ: الْهَاءُ وَالدَّالُ وَالْهَاءُ وَالدَّالُ: الْهُدْهُدُ طَائِرٌ مَعْرُوفٌ", "translation": "Ibn Faris said: Hāʾ, dāl, hāʾ, dāl: the hoopoe is a well-known bird."},
        ], ensure_ascii=False),
        'idioms_customs': json.dumps([
            {"arabic": "الْهُدْهُدُ", "translation": "The hoopoe — a crested bird known for its intelligence and role in Solomon's court."},
        ], ensure_ascii=False),
        'word_by_word_shades': json.dumps([
            {"shade": "Definite noun, accusative (ٱلْهُدْهُدَ) — the hoopoe", "reference": "27:20:8", "arabic_word": "ٱلْهُدْهُدَ", "explanation": "Accusative singular definite, object of Solomon's search."},
        ], ensure_ascii=False),
        'pos': 'N',
    },
    'lwt': {
        'root_arabic': 'لوت',
        'general_meaning': 'The triliteral root L-W-T (ل-و-ت) in the Quran is associated with the word \u0644\u0627\u062a (l\u0101ta), a rare negative particle meaning "not" or "not at all", appearing in the phrase \u0648\u064e\u0644\u064e\u0627\u062a\u064e \u062d\u0650\u064a\u0646\u064e \u0645\u064e\u0646\u064e\u0627\u0635\u064d ("but there was no time for escape") at 38:3. Classical grammarians differ on its derivation \u2014 some analyze it as a combination of \u0644\u064e\u0627 (negation) and \u062a\u064e (an emphatic suffix), while others trace it to a root \u0644-\u0648-\u062a with the sense of deficiency or cessation.',
        'shades_of_meaning': json.dumps([
            {"shade": "Negative particle of time \u2014 'not'", "reference": "38:3:8", "arabic_word": "\u0648\u064e\u0644\u064e\u0627\u062a\u064e", "explanation": "In the phrase '\u0648\u064e\u0644\u064e\u0627\u062a\u064e \u062d\u0650\u064a\u0646\u064e \u0645\u064e\u0646\u064e\u0627\u0635\u064d' \u2014 'but there was no time for escape' \u2014 the word \u0644\u0627\u062a serves as a specialized negation operating specifically on temporal nouns."},
        ], ensure_ascii=False),
        'hadith_examples': json.dumps([
            {"arabic": "\u0639\u064e\u0646\u0650 \u0627\u0628\u0652\u0646\u0650 \u0639\u064e\u0628\u064e\u0651\u0627\u0633\u064d \u0641\u0650\u064a \u0642\u064e\u0648\u0652\u0644\u0650\u0647\u0650 \u062a\u064e\u0639\u064e\u0627\u0644\u064e\u0649: \u00ab\u0648\u064e\u0644\u064e\u0627\u062a\u064e \u062d\u0650\u064a\u0646\u064e \u0645\u064e\u0646\u064e\u0627\u0635\u064d\u00bb \u0642\u064e\u0627\u0644\u064e: \u0644\u064e\u064a\u0652\u0633\u064e \u062d\u0650\u064a\u0646\u064e \u0641\u0650\u0631\u064e\u0627\u0631\u064d", "translation": "Ibn Abbas said regarding 'there was no time for escape': It is not the time for fleeing."},
        ], ensure_ascii=False),
        'classical_sources': json.dumps([
            {"arabic": "\u0642\u064e\u0627\u0644\u064e \u0633\u0650\u064a\u0628\u064e\u0648\u064e\u064a\u0652\u0647\u0650: \u0644\u064e\u0627\u062a\u064e \u0645\u064e\u0639\u0652\u0646\u064e\u0627\u0647\u064e\u0627 \u0644\u064e\u064a\u0652\u0633\u064e\u060c \u0648\u064e\u0647\u0650\u064a\u064e \u0644\u064e\u0627 \u0632\u0650\u064a\u062f\u064e\u062a\u0652 \u0641\u0650\u064a\u0647\u064e\u0627 \u0627\u0644\u062a\u064e\u0651\u0627\u0621\u064f", "translation": "Sibawayh said: L\u0101ta means laysa (is not), and it is l\u0101 with an extra t\u0101\u2019."},
            {"arabic": "\u0642\u064e\u0627\u0644\u064e \u0627\u0644\u0632\u064e\u0651\u062c\u064e\u0627\u062c\u064f: \u0644\u064e\u0627\u062a\u064e \u0623\u064e\u0635\u0652\u0644\u064f\u0647\u064e\u0627 \u0644\u064e\u0627 \u0648\u064e\u0627\u0644\u062a\u064e\u0651\u0627\u0621\u064f \u062a\u064e\u0623\u0652\u0643\u0650\u064a\u062f\u064c", "translation": "Al-Zajj\u0101j said: The origin of l\u0101ta is l\u0101 and the t\u0101\u2019 is for emphasis."},
        ], ensure_ascii=False),
        'idioms_customs': json.dumps([
            {"arabic": "\u0648\u064e\u0644\u064e\u0627\u062a\u064e", "translation": "wa-l\u0101ta \u2014 'but not', 'and there was no' \u2014 a compound negation used before time words."},
        ], ensure_ascii=False),
        'word_by_word_shades': json.dumps([
            {"shade": "Negative particle (\u0644\u064e\u0627\u062a\u064e) \u2014 'not'", "reference": "38:3:8", "arabic_word": "\u0648\u064e\u0644\u064e\u0627\u062a\u064e", "explanation": "A specialized negative particle used before temporal nouns; prefixed with wa- (and) in this occurrence."},
        ], ensure_ascii=False),
        'pos': 'N',
    },
    'mjs': {
        'root_arabic': 'مجس',
        'general_meaning': 'The triliteral root M-J-S (م-ج-س) refers to the Magians (al-Majūs), commonly identified as Zoroastrians. In the Quran, they are listed in one verse (22:17) among the religious communities — alongside believers, Jews, Sabians, Christians, and polytheists — whom God will judge on the Day of Resurrection. The term was used in pre-Islamic Arabia and early Islam for the Persian Zoroastrian priestly class.',
        'shades_of_meaning': json.dumps([
            {"shade": "Magians mentioned among religious groups", "reference": "22:17:8", "arabic_word": "ٱلْمَجُوسَ", "explanation": "'Indeed, those who have believed and those who were Jews and the Sabians and the Christians and the Magians and those who associated with Allah — Allah will judge between them on the Day of Resurrection.' — the only Quranic mention."},
        ], ensure_ascii=False),
        'hadith_examples': json.dumps([
            {"arabic": "عَنْ رَسُولِ اللَّهِ صَلَّى اللَّهُ عَلَيْهِ وَسَلَّمَ قَالَ: «سُنُّوا بِهِمْ سُنَّةَ أَهْلِ الْكِتَابِ»", "translation": "The Prophet ﷺ said regarding the Magians: 'Treat them as you treat the People of the Book.' (Muwatta Malik — indicating Magians were given protected status.)"},
        ], ensure_ascii=False),
        'classical_sources': json.dumps([
            {"arabic": "قَالَ ابْنُ فَارِسٍ: الْمِيمُ وَالْجِيمُ وَالسِّينُ: الْمَجُوسُ جِنْسٌ مِنَ الْكُفَّارِ", "translation": "Ibn Faris said: The letters mīm, jīm, and sīn: al-Majūs is a class of disbelievers (Zoroastrians)."},
        ], ensure_ascii=False),
        'idioms_customs': json.dumps([
            {"arabic": "الْمَجُوسُ", "translation": "The Magians / Zoroastrians — followers of the Persian religion founded by Zoroaster."},
            {"arabic": "مَجُوسِيٌّ", "translation": "A Magian / Zoroastrian (individual)."},
        ], ensure_ascii=False),
        'word_by_word_shades': json.dumps([
            {"shade": "Definite plural noun, accusative (ٱلْمَجُوسَ) — the Magians", "reference": "22:17:8", "arabic_word": "ٱلْمَجُوسَ", "explanation": "Accusative plural definite, listed among the groups subject to divine judgment."},
        ], ensure_ascii=False),
        'pos': 'N',
    },
    'nsr': {
        'root_arabic': 'نسر',
        'general_meaning': 'The triliteral root N-S-R (ن-س-ر) in the Quranic context refers to the idol Naser (نَسْر), one of the five idols worshipped by the people of Noah (Nūḥ). These idols — Wadd, Suwāʿ, Yaghūth, Yaʿūq, and Nasr — were originally righteous people whose images were eventually deified. The root also means \u2018vulture\u2019 in Arabic, as the idol likely took the form of a vulture or eagle.',
        'shades_of_meaning': json.dumps([
            {"shade": "The idol Nasr worshipped in Noah's time", "reference": "71:23:13", "arabic_word": "نَسْرًا", "explanation": "The people of Noah said: 'And do not leave Wadd or Suwāʿ or Yaghūth and Yaʿūq and Nasr' — listing five idols that they insisted on worshipping."},
        ], ensure_ascii=False),
        'hadith_examples': json.dumps([
            {"arabic": "عَنِ ابْنِ عَبَّاسٍ رَضِيَ اللَّهُ عَنْهُمَا قَالَ: «صَارَتِ الْأَوْثَانُ الَّتِي كَانَتْ فِي قَوْمِ نُوحٍ فِي الْعَرَبِ بَعْدُ»", "translation": "Ibn Abbas said: 'The idols that were among the people of Noah later appeared among the Arabs.' — explaining how the worship of these idols spread after the flood."},
        ], ensure_ascii=False),
        'classical_sources': json.dumps([
            {"arabic": "قَالَ ابْنُ فَارِسٍ: النُّونُ وَالسِّينُ وَالرَّاءُ: النَّسْرُ طَائِرٌ مَعْرُوفٌ", "translation": "Ibn Faris said: The letters nūn, sīn, and rāʾ: al-nasr (vulture) is a well-known bird."},
            {"arabic": "قَالَ مُجَاهِدٌ: نَسْرٌ: صَنَمٌ كَانَ لِقَوْمِ نُوحٍ", "translation": "Mujāhid said: Nasr was an idol of the people of Noah."},
        ], ensure_ascii=False),
        'idioms_customs': json.dumps([
            {"arabic": "نَسْرٌ", "translation": "Nasr — a vulture-shaped idol worshipped by Noah's people; also the Arabic word for vulture/eagle."},
            {"arabic": "النَّسْرُ الطَّائِرُ", "translation": "The vulture bird — a large bird of prey."},
        ], ensure_ascii=False),
        'word_by_word_shades': json.dumps([
            {"shade": "Proper noun, accusative (نَسْرًا) — the idol Nasr", "reference": "71:23:13", "arabic_word": "نَسْرًا", "explanation": "Indefinite accusative, the last of five idols mentioned in Noah's story."},
        ], ensure_ascii=False),
        'pos': 'N',
    },
    'xrTm': {
        'root_arabic': 'خرطم',
        'general_meaning': 'The quadriliteral root KH-R-Ṭ-M (خ-ر-ط-م) denotes the snout or trunk, particularly of an elephant or pig. In the Quran, it appears in Surah Al-Qalam (68:16) as a vivid metaphorical expression: \u2018We will brand him upon the snout\u2019 (سَنَسِمُهُ عَلَى الْخُرْطُومِ). This refers to a prominent disbeliever (al-Walīd ibn al-Mughīrah or similar) who will be disfigured or marked for disgrace on the Day of Judgment — the snout being the most prominent and sensitive part of the face.',
        'shades_of_meaning': json.dumps([
            {"shade": "Metaphorical branding on the snout as divine punishment", "reference": "68:16:3", "arabic_word": "ٱلْخُرْطُومِ", "explanation": "'We will brand him upon the snout' — a powerful metaphor for public humiliation and divine marking of the guilty, using the animalistic term 'snout' to degrade the arrogant opponent."},
        ], ensure_ascii=False),
        'hadith_examples': json.dumps([
            {"arabic": "عَنِ ابْنِ عَبَّاسٍ رَضِيَ اللَّهُ عَنْهُمَا قَالَ: نَزَلَتْ فِي الْوَلِيدِ بْنِ الْمُغِيرَةِ", "translation": "Ibn Abbas said: 'This verse was revealed concerning al-Walīd ibn al-Mughīrah.' (Tafsir sources — referring to 68:16.)"},
        ], ensure_ascii=False),
        'classical_sources': json.dumps([
            {"arabic": "قَالَ ابْنُ فَارِسٍ: الْخَاءُ وَالرَّاءُ وَالطَّاءُ وَالْمِيمُ لَيْسَ بِأَصْلٍ، وَالْخُرْطُومُ: طَرَفُ الْأَنْفِ", "translation": "Ibn Faris said: These four letters do not form a proper root; al-khurṭūm is the tip/extremity of the nose."},
        ], ensure_ascii=False),
        'idioms_customs': json.dumps([
            {"arabic": "الْخُرْطُومُ", "translation": "The snout/trunk — the protruding nose or muzzle of an animal; used metaphorically in the Quran for human degradation."},
            {"arabic": "سَنَسِمُهُ عَلَى الْخُرْطُومِ", "translation": "'We will brand him on the snout' — a Quranic idiom for irreversible public disgrace."},
        ], ensure_ascii=False),
        'word_by_word_shades': json.dumps([
            {"shade": "Definite noun, genitive (ٱلْخُرْطُومِ) — the snout", "reference": "68:16:3", "arabic_word": "ٱلْخُرْطُومِ", "explanation": "Genitive singular definite noun, preceded by 'alā (upon), forming a vivid metaphorical expression of divine punishment."},
        ], ensure_ascii=False),
        'pos': 'N',
    },
}


def read_csv_rows():
    """Read all rows from root_meaning.csv as a list of dicts."""
    with open(CSV_PATH, 'r', newline='', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        return list(reader)


def write_csv_rows(rows, fieldnames):
    """Write rows back to root_meaning.csv."""
    with open(CSV_PATH, 'w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, lineterminator='\r\n')
        writer.writeheader()
        writer.writerows(rows)


def build_full_response(row):
    """Build full_response JSON dict from the 6 field columns."""
    fr = {}
    for key in FR_KEYS:
        val = row[key].strip()
        if key == 'general_meaning':
            fr[key] = val
        else:
            # These are JSON arrays stored as strings
            try:
                fr[key] = json.loads(val) if val else []
            except json.JSONDecodeError:
                fr[key] = []
    return fr


def backfill_full_response(rows):
    """Reconstruct full_response for rows where it's empty or invalid JSON."""
    updated = 0
    skipped_no_data = 0
    skipped_already = 0

    for row in rows:
        fr_raw = row['full_response'].strip()
        needs_backfill = False

        if not fr_raw:
            needs_backfill = True
        else:
            try:
                json.loads(fr_raw)
            except json.JSONDecodeError:
                needs_backfill = True

        if not needs_backfill:
            skipped_already += 1
            continue

        # Build full_response from field columns
        fr = build_full_response(row)

        # Check if field columns have any actual data
        has_data = bool(fr['general_meaning']) or any(fr[k] for k in FR_KEYS[1:])
        if not has_data:
            skipped_no_data += 1
            continue

        row['full_response'] = json.dumps(fr, ensure_ascii=False)
        updated += 1

    return updated, skipped_no_data, skipped_already


def add_missing(rows):
    """Add rows for the 8 missing roots."""
    existing_roots = set(r['root'] for r in rows)
    added = 0

    max_id = max(int(r['id']) for r in rows)
    next_id = max_id + 1

    fieldnames = list(rows[0].keys())

    for root, data in sorted(MISSING_ROOTS.items()):
        if root in existing_roots:
            continue

        fr = build_full_response_from_data(data)
        new_row = {
            'id': str(next_id),
            'root': root,
            'root_arabic': data['root_arabic'],
            'general_meaning': data['general_meaning'],
            'shades_of_meaning': data['shades_of_meaning'],
            'hadith_examples': data['hadith_examples'],
            'classical_sources': data['classical_sources'],
            'idioms_customs': data['idioms_customs'],
            'word_by_word_shades': data['word_by_word_shades'],
            'full_response': json.dumps(fr, ensure_ascii=False),
            'model_used': 'deepseek-v4-flash',
            'pos': data['pos'],
        }
        rows.append(new_row)
        existing_roots.add(root)
        next_id += 1
        added += 1

    return added


def build_full_response_from_data(data):
    """Build full_response dict from the MISSING_ROOTS data structure."""
    return {
        'general_meaning': data['general_meaning'],
        'shades_of_meaning': json.loads(data['shades_of_meaning']),
        'hadith_examples': json.loads(data['hadith_examples']),
        'classical_sources': json.loads(data['classical_sources']),
        'idioms_customs': json.loads(data['idioms_customs']),
        'word_by_word_shades': json.loads(data['word_by_word_shades']),
    }


def main():
    if len(sys.argv) < 2:
        print("Usage: python insert-roots.py [--backfill-full-response | --add-missing]")
        sys.exit(1)

    mode = sys.argv[1]

    if mode == '--backfill-full-response':
        rows = read_csv_rows()
        original_len = len(rows)
        fieldnames = list(rows[0].keys())

        updated, skipped_no_data, skipped_already = backfill_full_response(rows)

        write_csv_rows(rows, fieldnames)
        print(f"Backfill complete.")
        print(f"  Total rows:              {original_len}")
        print(f"  Already valid:           {skipped_already}")
        print(f"  Updated (reconstructed): {updated}")
        print(f"  Skipped (no field data): {skipped_no_data}")

    elif mode == '--add-missing':
        rows = read_csv_rows()
        original_len = len(rows)
        fieldnames = list(rows[0].keys())

        added = add_missing(rows)

        write_csv_rows(rows, fieldnames)
        print(f"Add-missing complete.")
        print(f"  Original rows: {original_len}")
        print(f"  Added rows:    {added}")
        print(f"  Total rows:    {len(rows)}")

    else:
        print(f"Unknown mode: {mode}")
        print("Usage: python insert-roots.py [--backfill-full-response | --add-missing]")
        sys.exit(1)


if __name__ == '__main__':
    main()
