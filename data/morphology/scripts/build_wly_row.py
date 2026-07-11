# -*- coding: utf-8 -*-
import sys
import os
import io

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8')

# Define the 25 fields for the new row
fields = [
    '61',
    'wly',
    'وَلِيّ',
    'و ل ي',
    'N',
    '232',
    '"Guardian, ally, friend, protector, patron, master"',
    '"To be near, to be close, to befriend, to protect, to support, to succeed, to inherit. The root و ل ي expresses the comprehensive Arabic concept of walāyah — nearness, friendship, protection, authority, and succession. The triliteral carries two intertwined semantic nuclei: (1) qurba — nearness/closeness, the spatial-social fact of standing next to someone; (2) mulk/siyāda — mastery and authority, the relational fact of being in charge of another\'s affairs. The verb waliya (Form I, rare in Quran) means to be near, succeed, befriend, inherit; the noun walī (86x, including awliyāʾ) covers friend, protector, ally, guardian, patron, helper, heir, next of kin. Form II wallā (30x) shifts the semantic vector toward direction: to turn toward (and appoint as guardian) or to turn away. Form V tawallā (78x) is the most polemically active verb in this root: to turn away (from the truth), to take as one\'s walī, to follow, to depart, to charge with. The passive-active mawlā (18x, including mawālī) names the master-patron or the client-companion. The Quran deploys this root across the central theological axis of divine and human friendship: Allah as walī of the believers (2:257, 4:75), the awliyāʾ Allāh as those freed from fear and grief (10:62, 7:30), Satan as the walī of the disbelievers (4:76, 7:51), the believers as mutual awliyāʾ (9:71), and mawlā as the master/patron and a name of Allah (47:11)."',
    '"الْقُرْبُ وَالْمُلْكُ — الْوَلَايَةُ: الْمَحَبَّةُ وَالنُّصْرَةُ وَالْقِيَامُ بِالشَّأْنِ. أَصْلُ الْجَذْرِ يَدُلُّ عَلَى الْقُرْبِ مَعَ التَّصَرُّفِ وَالسَّلْطَنَةِ. الْوَلِيُّ: الْقَرِيبُ الْمُتَصَرِّفُ فِي أَمْرِ غَيْرِهِ، وَقَدْ يَكُونُ مَحَبَّةً أَوْ نَسَبًا أَوْ حِلْفًا أَوْ عِتْقًا. وَالْمَوْلَى: الْقَرِيبُ مِنْ جِهَةِ النَّسَبِ أَوِ الْحِلْفِ أَوِ الْعِتْقِ، وَهُوَ السَّيِّدُ تَارَةً وَالْعَبْدُ تَارَةً وَالْوَلِيُّ تَارَةً. وَالتَّوَلِّي: الْإِعْرَاضُ عَنِ الْحَقِّ، أَوْ تَنَاوُلُ الْوَلِيِّ وَاتِّبَاعُهُ. وَالْوِلَايَةُ: الْإِمَارَةُ وَالرِّئَاسَةُ. وَالْوَلَاءُ: حَقُّ الْمَوَدَّةِ وَالْقَرَابَةِ"',
    '"*w-l-y — Pan-Semitic root expressing \'to be near, to be joined, to be a companion\'"',
    '"לוה (l-w-h) — to be a close friend, to be joined, to accompany (archaic Hebrew sense); לוי (Lewi) — the Levite, the priestly tribe named for its \'attachment\' to God (the same semantic core of being-near-to-the-divine)"',
    '"Syriac: ܠܘܐ (l-w-ʾ) — to be a companion, to accompany, to attach oneself to; ܠܘܝܐ (ləwəyā) — companion, fellow; Jewish Palestinian Aramaic: לוי (ləwī) — companion, fellow (Targumic); Babylonian Aramaic: לוואי (ləwāy) — companion, kinsman"',
    '"Ԓወ (läwä) — to be a follower, disciple, companion; Ԓውያን (läwäyān) — followers, disciples (Ge\'ez ecclesiastical usage, preserving the original Semitic sense of attachment)"',
    '"Proto-Semitic *w-l-y carried the sense of \'to be near, to be joined, to be a companion.\' In Arabic the root developed two intertwined cores: (1) qurba (nearness) — the spatial-social fact of being close to someone, hence friendship; (2) tafaḍḍul/walāyah — the relational fact of being a guardian or protector, the one who is near AND in charge. This double sense of \'being-near-and-in-charge\' produced the rich Arabic semantic field of walī: friend, ally, protector, master, patron, client, heir. The Hebrew parallel is striking: the Levites (Ləwiyyīm) take their name from the related root l-w-h, named for their \'attachment\' to God in priestly service — the same semantic core of being-near-to-the-divine. Arabic then extended walāyah into: (1) religious walāyah — friendship with God (2:257, 10:62); (2) political walāyah — master/client relationship (mawlā); (3) hereditary walāyah — succession, inheritance, next-of-kin (waliyy al-mīrāth); (4) the negative pole — Satan as the walī of the disbelievers (4:76, 7:51); (5) the antonymic use — tawalla (Form V, 78x) and walla (Form II, 30x) meaning \'to turn away\' from the truth, exploiting the spatial sense of \'turning one\'s back\' to acquire a moral sense."',
    '"I|وَلِيَ|waliya|"to be near, befriend, protect, succeed, inherit"|3||II|وَلَّى|walla|"to turn away, turn toward, appoint as guardian"|30||III|وَالَى|wālā|"to befriend, show loyalty, be an ally"|6||V|تَوَلَّى|tawallā|"to turn away, take as guardian, follow, charge with"|78||VIII|ٱتَّوْلَى|ittawlā|"to befriend, be loyal (rare)"|1||X|ٱسْتَوْلَى|istawlā|"to gain mastery, dominate, prevail"|2"',
    '"وَلِيّ|walī|"friend, protector, ally, guardian, patron, helper, heir, next of kin — 86x"|أَوْلِيَاء|awliyāʾ|"friends, allies, protectors, helpers (pl. of walī)"|مَوْلَى|mawlā|"master, lord, patron, protector, client, friend, ally, heir — 18x"|مَوَالٍ|mawālin|"masters, patrons, allies, clients (broken pl. of mawlā)"|وَلَايَة|walāyah|"friendship, protection, authority, wilāyah"|وِلَايَة|wilāyah|"authority, charge, governorship, wilāyah"|وَلَاء|walāʾ|"nearness, alliance, clientage, bond"|وُلَاة|wulāt|"governors, protectors, authorities (pl.)"|مُسْتَوْلٍ|mustawlin|"one who has gained mastery, dominant (Form X active part.)"|مُتَوَلٍّ|mutawallin|"one who takes charge, warden, custodian (Form V active part.)"',
    '"Allah as walī of the believers — 2:257 "Allahu waliyyu alladhina amanu yukhrijuhum mina al-ẓulumati ila al-nur"; 4:75 "wa Allāhu waliyyu al-muʾminīn"; the awliyāʾ Allāh freed from fear and grief — 10:62 "awliya Allahi la khawfun ʿalayhim wa la hum yahzanun"; 7:30 (variant); 19:96 (awliyāʾ al-Raḥmān); Satan as walī of disbelievers — 4:76 "al-shayṭānu waliyyu al-kāfirīn"; 7:51; the believers as mutual awliyāʾ — 9:71 "al-muʾminūna wa al-muʾminātu baʿḍuhum awliyāʾu baʿḍ"; mawlā as master/patron and divine attribute — 22:13, 47:11; ūlī al-amr in 4:59; tawalla as turning away from truth (78x, frequent polemical verb); al-mawālī and inheritance law; awliyāʾ al-shar (evil allies); wilāyah as divine succoring and rule — 7:196, 9:74"',
    '"الأصلان عند ابن فارس: القُرْبُ، والمُلْكُ والسُّلْطان. يقال: ولي فلانٌ مكان فلانٍ إذا قام مقامه. والوليُّ: القريب، والمتصرِّف. والوَلايةُ: المُلك. وفي القرآن: ﴿اللهُ وليُّ الذين آمنوا﴾. (معجم مقاييس اللغة لابن فارس)"',
    '"ولاية وولاء كلاهما مصدر لولي بمعنى القرب. والمُوَلاَةُ: المخالفة. والوليُّ: المُحِبُّ والقريب والنصير. والتولّي: الإعراض عن الحق. وولاية الله: تَصَرُّفُهُ في خلقه بهدايتهم وحفظهم. (المفردات في غريب القرآن للراغب الأصبهاني)"',
    '"Walī: next of kin, heir, friend, companion, ally, protector, helper, guardian, patron, client. Mawlā: master, lord, patron, protector, ally, friend, client, kinsman, heir, emancipator. Walāyah: nearness, friendship, authority, guardianship. Walāʾ: nearness, alliance, clientage. Tawallā: to turn away, to take as one\'s walī, to follow, to depart. Wālā: to be a friend, to befriend, to show loyalty. (Lane, Vol.8, pp.2937-2944)"',
    '"Authority / friendship / protection / nearness / allegiance / kinship / inheritance / mastery / support / following / companionship / political loyalty / divine favor / covenant allegiance / clientage"',
    '"The believer\'s walāyah with Allah (2:257) as the central axis of the faith relationship — the most distinctive Quranic deployment of this root; the category of awliyāʾ Allāh (10:62, 7:30) as those freed from fear and grief — central to Quranic piety, Sufi theology, and the doctrine of sainthood (walāyah); mawlā as a name of Allah (47:11) and as the human patron/successor — the focus of major theological disputes (e.g., the Ghadīr Khumm hadith on the succession to the Prophet); the political implications of ūlī al-amr in 4:59; the polemical use of tawalla (78x) to describe disbelievers turning away from the truth; Satan as the walī of the kuffār (4:76); the walāyah doctrine in Sunni and Shi\'i theology with significantly different developments; the mawlā relationship in Islamic law covering manumission (ʿitq), clientship (walāʾ al-ʿitq), and inheritance (al-mawālī); wilāyah as religious authority (wilāyat al-faqīh in Shi\'i doctrine, wilāyat al-amr in Sunnī political theory); the divine attribute al-Walī (The Protecting Friend) as one of the 99 Names; the distinction between awliyāʾ al-Raḥmān (friends of the Most Merciful, 19:96) and awliyāʾ al-shayṭān (allies of Satan)"',
    '"{"ar":"مَنْ عَادَىٰ لِي وَلِيًّا فَقَدْ آذَنْتُهُ بِالْحَرْبِ","en":"Whoever shows enmity to a walī of Mine, I have declared war against him.","source":"Ṣaḥīḥ al-Bukhārī 6502 — ḥadīth qudsī","relevance":"Walāyah as the central divine-human relationship — Allah declares war on the enemies of His walī, identifying friendship with Him as the criterion of faith"}|{"ar":"مَنْ كُنْتُ مَوْلَاهُ فَعَلِيٌّ مَوْلَاهُ","en":"Of whomsoever I am mawlā, ʿAli is his mawlā","source":"Ghadīr Khumm — multi-source ḥadīth; Sunan al-Tirmidhī 3713, Musnad Aḥmad 946, etc.","relevance":"The contested mawlā hadith — central to Shi\'i-Sunni theological divergence on walāyah and succession"}|{"ar":"الْوَلَاءُ لِمَنْ أَعْطَى الثَّمَنَ","en":"The clientage belongs to the one who pays the price","source":"Ṣaḥīḥ al-Bukhārī 2150, 2151","relevance":"Law of mawālāt — mawlā is the one who pays the ransom for manumission"}"',
    'nṣr,qrb,qwm',
    'COMPLETE',
]

# Build the row
row_parts = []
for i, f in enumerate(fields):
    if i < 6 or i == 23 or i == 24:
        # Unquoted fields
        row_parts.append(f)
    else:
        # Text field — wrap in outer quotes and escape internal quotes
        escaped = f.replace('"', '""')
        row_parts.append('"' + escaped + '"')

new_row = ','.join(row_parts)
print(f'New row length: {len(new_row)} chars', file=sys.stderr)
print(f'Field count: {len(fields)}', file=sys.stderr)

# Read existing file
with open('/Users/saqlain/corpus-quran-java/data/morphology/meanings-roots.csv', 'r', encoding='utf-8') as f:
    content = f.read()

lines = content.split('\n')
print(f'File has {len(lines)} lines', file=sys.stderr)
print(f'Old line 62 (first 100): {lines[61][:100]}', file=sys.stderr)

lines[61] = new_row

with open('/Users/saqlain/corpus-quran-java/data/morphology/meanings-roots.csv', 'w', encoding='utf-8') as f:
    f.write('\n'.join(lines))

print('Done', file=sys.stderr)
