/**
 * Translation layer: MASAQ technical terms → learner-friendly grammatical terms.
 * Complete coverage of all 132 Morph_Tag values and all other MASAQ columns.
 *
 * MASAQ Column Reference:
 *   Morph_Tag       — Part of speech / morphological category (132 unique)
 *   Morph_Type      — Morpheme position: Prefix | Stem | Suffix | Other_i3rab
 *   Syntactic_Role  — Grammatical function in the sentence (59 unique)
 *   Case_Mood       — Case (nouns) or Mood (verbs): NOM/GEN/ACC/INV/JUSS
 *   Case_Mood_Marker — Surface realization: DHAMMA, KASRA, FATHA, SUKUN, etc.
 *   Possessive_Construct — CONSTRUCT (مضاف) or NOT_CONSTRUCT
 *   Invariable_Declinable — INVAR (مبني), DECLN (معرب), DEF_ART, etc.
 *   Phrase          — Clause type: NOM_SNT, VERB_SNT, PHRASE, etc.
 *   Phrasal_Function — Function in clause: PRED, APPOS, CIRCUM, etc.
 *   Segmented_Word  — Morphological segment (prefix/stem/suffix text)
 *   Gloss           — English gloss (word-level, shared across segments)
 */

// ===================================================================
// Morph_Tag — Part of Speech / Morphological Category (132 values)
// ===================================================================

const POS_MAP: Record<string, string> = {

  // ── Function words ──────────────────────────────────────────
  PREP: "Preposition (حرف جر) — e.g. bi-, li-, fi-, min, `an",
  DET: "Definite Article (الـ) — the determiner 'al-'",
  CONJ: "Conjunction (حرف عطف) — e.g. wa-, fa-, thumma",
  NEG_PART: "Negation Particle — e.g. lā, mā, lam, lan",
  "NEG_PART.NEG_CAT": "Negation of Category — categorical negation",
  "NEG_PART.NEG_MAA": "Negation with Mā — 'mā' of negation",
  INTERROG: "Interrogative Particle — e.g. hal, a-",
  INTERROG_PART: "Interrogative Particle",
  INTERROG_PRON: "Interrogative Pronoun — e.g. man, mā",
  VOC_PART: "Vocative Particle — 'yā' (O...)",
  FUT_PART: "Future Particle — 'sa-' (near future)",
  FUTURE_PART: "Future Particle — 'sawfa' (distant future)",
  FUTUR_PART: "Future Particle — 'sawfa' (distant future)",
  SUBJUNC_PART: "Subjunctive Particle — 'an', 'lan', 'kay'",
  JUSSIVE_PART: "Jussive Particle — 'lam', 'lammā'",
  CONDITION_PART: "Conditional Particle — 'in', 'law'",
  CERT_PART: "Emphatic Particle — 'qad', 'la-'",
  ANNUL_PART: "Annuller Particle — 'kāna' and its sisters",
  INF_ANNUL_PART: "Annuller Particle — 'inna' and its sisters",
  INF_SUBJUNC_PART: "Subjunctive Particle — 'an'",
  EMPHATIC_NUN: "Emphatic Nūn — 'nūn al-tawkīd'",
  PROTECT_NUN: "Protective Nūn — 'nūn al-wiqāya'",
  NOON_V5: "Nūn of the Five Verbs — 'nūn al-af`āl al-khamsa'",
  PART: "Particle — generic grammatical particle",
  OTHER: "Other Particle",
  "OTHER.OTHER": "Other Particle",

  // ── Nouns ───────────────────────────────────────────────────
  NOUN_ABSTRACT: "Abstract Noun — e.g. 'īmān (faith), `ilm (knowledge)",
  NOUN_CONCRETE: "Concrete Noun — e.g. kitāb (book), bayt (house)",
  NOUN_PROP: "Proper Noun — e.g. Allāh, Mūsā, Ibrāhīm",
  NOUN_ACTIVE_PART: "Active Participle (اسم فاعل) — 'doer' of the action",
  NOUN_PASSIVE_PART: "Passive Participle (اسم مفعول) — receiver of the action",
  NOUN_NUM: "Number Noun — e.g. ithnān (two), thalātha (three)",
  NOUN_FIVE: "One of the Five Nouns — ab, akh, ḥam, fū, dhū",
  NOUN_RELATIVE: "Relative Noun — e.g. alladhī, allatī",
  NOUN_TIME_PLACE: "Noun of Time/Place (اسم زمان/مكان) — e.g. masjid, maw`id",
  NOUN_VERB_LIKE: "Verb-like Noun — nouns that behave like verbs",
  NOUN_PROP_FOREIGN: "Foreign Proper Noun — non-Arabic names",
  EXCEPT_NOUN: "Exception Noun — used in istithnā' constructions",
  GERUND: "Verbal Noun (مصدر) — the abstract action, e.g. ḍarb (hitting)",
  GERUND_INSTANT: "Instant Gerund (اسم مرة) — single occurrence",
  GERUND_MEEM: "Mīm Gerund (مصدر ميمي) — gerund prefixed with mīm",
  GERUND_PROFESSION: "Professional Gerund — e.g. najjār (carpenter)",

  // ── Pronouns ────────────────────────────────────────────────
  PRON: "Pronoun — general pronoun",
  PRON_1S: "1st Person Singular Pronoun — 'I' / 'me' (-ī, -nī)",
  PRON_1P: "1st Person Plural Pronoun — 'we' / 'us' (-nā)",
  PRON_2MS: "2nd Person Masculine Singular — 'you' (-ka)",
  PRON_2MP: "2nd Person Masculine Plural — 'you all' (-kum)",
  PRON_3MS: "3rd Person Masculine Singular — 'he' / 'him' (-hu)",
  PRON_3FS: "3rd Person Feminine Singular — 'she' / 'her' (-hā)",
  PRON_3D: "3rd Person Dual — 'they two' (-humā)",
  PRON_3MP: "3rd Person Masculine Plural — 'they' (-hum)",
  PRON_3FP: "3rd Person Feminine Plural — 'they (f)' (-hunna)",
  SUBJ_PRON: "Subject Pronoun — 'anta', 'huwa', 'nahnu' (independent form)",
  OBJ_PRON: "Object Pronoun — attached pronoun as object",
  POSS_PRON: "Possessive Pronoun — attached pronoun as possessor",
  POSS_PRON_1S: "Possessive Pronoun — 'my' (-ī)",
  POSS_PRON_1P: "Possessive Pronoun — 'our' (-nā)",
  POSS_PRON_3MS: "Possessive Pronoun — 'his' (-hu)",
  POSS_PRON_3FS: "Possessive Pronoun — 'her' (-hā)",
  POSS_PRON_3D: "Possessive Pronoun — 'their' (dual, -humā)",
  POSS_PRON_3MP: "Possessive Pronoun — 'their' (m.pl, -hum)",
  DEM_PRON: "Demonstrative Pronoun — 'this', 'that'",
  DEM_PRON_MS: "Demonstrative — 'hādhā' (this, m.sg)",
  DEM_PRON_FS: "Demonstrative — 'hādhihi' (this, f.sg)",
  DEM_PRON_MP: "Demonstrative — 'hā'ulā'i' (these, m.pl)",
  DEM_PRON_F: "Demonstrative — 'hā'ulā'i' (these, f)",
  REL_PRON: "Relative Pronoun — 'alladhī', 'allatī', etc.",
  REL_ADV: "Relative Adverb — e.g. ḥaythu (where), `inda (when)",

  // ── Verbs ───────────────────────────────────────────────────
  IV: "Imperfect Verb (مضارع) — present/future tense",
  IV1P: "Imperfect Verb 1st Plural — 'naf`alu' (we do)",
  IV3MS: "Imperfect Verb 3rd M.Sg — 'yaf`alu' (he does)",
  IV3FS: "Imperfect Verb 3rd F.Sg — 'taf`alu' (she does)",
  IV3MP: "Imperfect Verb 3rd M.Pl — 'yaf`alūna' (they do)",
  IV_PASS: "Passive Imperfect Verb — 'yuf`alu' (it is done)",
  CV: "Imperative Verb (أمر) — command form, e.g. 'if`al!' (do!)",
  CV_PREF: "Imperative Prefix — the initial 'i-' of the imperative",
  PV: "Perfect Verb (ماض) — past tense, e.g. 'fa`ala' (he did)",
  PV_PASS: "Passive Perfect Verb — 'fu`ila' (it was done)",
  UNINFLECTED_VERB: "Uninflected Verb — frozen form",
  IMPERF_PREF: "Imperfect Prefix — 'ya-', 'ta-', 'a-', 'na-'",

  // ── Subject Suffixes (embedded in verbs) ───────────────────
  CVSUFF_SUBJ_2MP: "Imperative Subject Suffix — 'you (m.pl)' (-ū)",
  CVSUFF_SUBJ_2MS: "Imperative Subject Suffix — 'you (m.sg)'",
  IVSUFF_SUBJ_MP_MOOD_I: "Imperfect Subject Suffix — m.pl indicative (-ūna)",
  IVSUFF_SUBJ_MP_MOOD_SJ: "Imperfect Subject Suffix — m.pl subjunctive (-ū)",
  PVSUFF_SUBJ_1P: "Perfect Subject Suffix — 'we' (-nā)",
  PVSUFF_SUBJ_2MP: "Perfect Subject Suffix — 'you (m.pl)' (-tum)",
  PVSUFF_SUBJ_3FS: "Perfect Subject Suffix — 'she' (-at)",
  PVSUFF_SUBJ_3MP: "Perfect Subject Suffix — 'they (m.pl)' (-ū)",
  PVSUFF_SUBJ_3MS: "Perfect Subject Suffix — 'he' (implicit fatḥa)",

  // ── Adjectives ──────────────────────────────────────────────
  ADJ_COMP: "Comparative Adjective — 'af`alu' pattern, e.g. akbar (greater)",
  ADJ_INTENS: "Intensive Adjective — e.g. ghaffār (oft-forgiving)",
  ADJ_QUALIT: "Qualitative Adjective — e.g. hasan (good), kabīr (big)",

  // ── Noun Suffixes (number/gender/case) ─────────────────────
  NSUFF_FEM_SG: "Feminine Suffix — '-ah / -at' (ـة)",
  NSUFF_FEM_PL: "Feminine Plural Suffix — '-āt' (ـات)",
  NSUFF_FEM_DU_NOM: "Feminine Dual Nominative Suffix — '-atāni' (ـتان)",
  NSUFF_FEM_DU_ACC: "Feminine Dual Acc/Gen Suffix — '-atayni' (ـتين)",
  NSUFF_FEM_DU_GEN: "Feminine Dual Genitive Suffix",
  NSUFF_MASC_DU_NOM: "Masculine Dual Nominative Suffix — '-āni' (ـان)",
  NSUFF_MASC_DU_ACC: "Masculine Dual Acc/Gen Suffix — '-ayni' (ـين)",
  NSUFF_MASC_DU_GEN: "Masculine Dual Genitive Suffix",
  NSUFF_MASC_PL_NOM: "Masculine Plural Nominative Suffix — '-ūna' (ـون)",
  NSUFF_MASC_PL_ACC: "Masculine Plural Acc/Gen Suffix — '-īna' (ـين)",
  NSUFF_MASC_PL_GEN: "Masculine Plural Genitive Suffix",
  SUFF: "Suffix — generic",
  SUFF_FEM_TA: "Feminine Tāʼ Suffix — the '-t' in feminine forms",
  RELATIVE_YA: "Relative Yāʼ — nisba suffix '-ī' (ـي)",

  // ── Case Markers (abstract iʻrāb — no surface Arabic form) ──
  CASE_DEF_ACC: "Definite Accusative Marker — fatḥa (ـَ) on definite nouns",
  CASE_DEF_GEN: "Definite Genitive Marker — kasra (ـِ) on definite nouns",
  CASE_DEF_NOM: "Definite Nominative Marker — ḍamma (ـُ) on definite nouns",
  CASE_INDEF_ACC: "Indefinite Accusative Marker — fatḥatayn (ـً)",
  CASE_INDEF_GEN: "Indefinite Genitive Marker — kasratayn (ـٍ)",
  CASE_INDEF_NOM: "Indefinite Nominative Marker — ḍammatayn (ـٌ)",
  "CASE_INDEF_(ACC_GEN)": "Indefinite Acc/Gen Marker — diptote fatḥa (ـَ)",

  // ── Adverbs ─────────────────────────────────────────────────
  ADV: "Adverb — e.g. hunā (here), abadan (ever)",
  ADV_PLCE: "Adverb of Place (ظرف مكان) — e.g. fawqa (above)",
  ADV_TIME: "Adverb of Time (ظرف زمان) — e.g. ba`da (after)",

  // ── Particles ───────────────────────────────────────────────
  KAAFA_MAKFOUFA: "Preventive Kāffa — 'mā' that prevents `amal",
  EXCEPT_PART: "Exception Particle — 'illā' (except)",
  YES_NO_RESP_PART: "Yes/No Response Particle — 'na`am', 'balā'",
  FOREIGN: "Foreign Word — non-Arabic term",
  "FOREIGN.FOREIGN": "Foreign Word",
  UNKNOWN: "Unclassified — fallback for unrecognized tags",
  "None": "No Tag — unspecified",
};

// ===================================================================
// Syntactic_Role — Grammatical function in the sentence (59 values)
// ===================================================================

const ROLE_MAP: Record<string, string> = {
  SUBJ: "Subject (فاعل / مبتدأ) — the doer or topic of the sentence",
  TOPIC: "Topic (مبتدأ) — the subject of a nominal sentence",
  PRED: "Predicate (خبر) — what is said about the subject",
  OBJ: "Direct Object (مفعول به) — recipient of the action",
  PREP: "Preposition (حرف جر) — governs a genitive noun",
  PREP_OBJ: "Object of Preposition (اسم مجرور) — noun after a preposition",
  ADJ: "Adjective (صفة / نعت) — describes a noun",
  GEN_CONS: "Genitive Construct (مضاف إليه) — second term of iḍāfa",
  APPOS: "Appositive (بدل) — substitutes or clarifies a noun",
  CONJ: "Conjunction (عطف) — links equal elements",
  CONJ_N: "Conjoined Noun — noun linked by conjunction",
  EMPH: "Emphasis (توكيد) — reinforces meaning",
  CIRCUM: "Circumstantial (حال) — describes the state during action",
  VOC: "Vocative (منادى) — the person being called",
  EXCP: "Exception (استثناء) — excluded from a general statement",
  COMPL: "Complement — completes the meaning",
  COGN: "Cognate Accusative (مفعول مطلق) — verbal noun emphasizing the verb",
  PURP: "Purpose (مفعول لأجله) — reason for the action",
  ADV_TIME: "Adverb of Time (ظرف زمان) — temporal adverb",
  ADV_PLCE: "Adverb of Place (ظرف مكان) — spatial adverb",
  ACC_SPECIF: "Specification (تمييز) — clarifies ambiguity",
  INTENCIF: "Intensive — intensifying construction",
  EXPLET: "Expletive — filler word",
  ACRON: "Acronym — abbreviated form",

  CV: "Imperative Verb — command form",
  CV_COP: "Imperative Copula — imperative of kāna",
  IV: "Imperfect Verb — present/future tense",
  IV_COP: "Imperfect Copula — imperfect of kāna",
  IV_PASS: "Passive Imperfect Verb",
  AGNT: "Agent — doer of the action",
  PASS_SUBJ: "Passive Subject (نائب فاعل) — subject of passive verb",
  NON_INFLECT: "Non-Inflected — uninflected word",

  PART_CONDITION: "Conditional Particle — 'in', 'law'",
  PART_COP_PRED: "Copula Predicate Particle — predicate of kāna/inna",
  PART_EXCEPT: "Exception Particle — 'illā'",
  PART_INTERROG: "Interrogative Particle — 'hal', 'a-'",
  PART_JUSSIVE: "Jussive Particle — 'lam', 'lammā'",
  PART_PREV: "Preventive Particle",
  SUBJ_COP_PART: "Subject of Copula Particle — subject of inna",
  SUBJ_COP_V: "Subject of Copula Verb — subject of kāna",
  SUBJ_DELA: "Subject of Delayed Action",
  SUBJ_NEG_CAT: "Subject of Negative Category",
  NEG_CAT: "Negative Category — categorical negation",
  NEG_MAA: "Negative Mā — 'mā' of negation",
  SUBJUNC_PART: "Subjunctive Particle — 'an', 'lan', 'kay'",
  SUBOR_CONJ: "Subordinating Conjunction — introduces dependent clause",
  SUBOR_ANN_CONJ: "Subordinating Conjunction — 'an'",
  SUBS_COG_ACC: "Substitute Cognate Accusative",
  V_COP_PRED: "Predicate of Copula Verb — khabar of kāna",
  INTERJ_CV: "Interjection (imperative form)",
  INTERJ_PV: "Interjection (perfect form)",
  ANNUL_PART: "Annuller Particle — inna/kāna type",
  CERT_PART: "Emphatic Particle — 'qad', 'la-'",
  VOC_PART: "Vocative Particle — 'yā'",
};

// ===================================================================
// Morph_Type — Morpheme position in word (4 values)
// ===================================================================

const MORPH_TYPE_MAP: Record<string, string> = {
  Prefix: "Prefix — grammatical morpheme before the stem (ال, ب, ل, و, ف, س)",
  Stem: "Stem — the lexical root/base carrying the core meaning",
  Suffix: "Suffix — grammatical morpheme after the stem (ون, ين, ة, ت, نا)",
  Other_i3rab: "Iʻrāb Marker — abstract case/mood ending (not a surface morpheme)",
};

// ===================================================================
// Case_Mood — Case (nouns) or Mood (verbs)
// ===================================================================

const CASE_MAP: Record<string, string> = {
  NOMINATIVE: "Nominative (مرفوع) — subject/predicate case (ـُ / ـٌ)",
  GENITIVE: "Genitive (مجرور) — after preposition or in iḍāfa (ـِ / ـٍ)",
  ACCUSATIVE: "Accusative (منصوب) — object/adverbial case (ـَ / ـً)",
  INVARIABLE: "Invariable (مبني) — fixed ending, does not change",
  JUSSIVE: "Jussive (مجزوم) — verb mood after lam/lammā (ـْ)",
};

// ===================================================================
// Case_Mood_Marker — Surface realization of case/mood
// ===================================================================

const MARKER_MAP: Record<string, string> = {
  DHAMMA: "Ḍamma (ـُ) — nominative marker",
  KASRA: "Kasra (ـِ) — genitive marker",
  FATHA: "Fatḥa (ـَ) — accusative marker",
  FATHA_DIPTOTE: "Fatḥa on Diptote (ـَ) — accusative for diptote nouns",
  SUKUN: "Sukūn (ـْ) — jussive marker (no vowel)",
  NUN: "Nūn (ـن) — sound plural / 5 verbs marker",
  WAW: "Wāw (و) — sound masculine plural nominative",
  YAA: "Yāʼ (ي) — sound masc plural acc/gen / dual / nisba",
  ALIF: "Alif (ا) — dual nominative / long vowel",
  DEL_NUN: "Deleted Nūn — nūn dropped in subjunctive/jussive",
  DEL_VOWEL: "Deleted Vowel — vowel dropped in jussive",
  IMP_DHAMMA: "Implied Ḍamma — ḍamma not written (defective noun)",
  IMP_FATHA: "Implied Fatḥa — fatḥa not written",
  IMP_KASRA: "Implied Kasra — kasra not written",
};

// ===================================================================
// Possessive_Construct — Iḍāfa state
// ===================================================================

const CONSTRUCT_MAP: Record<string, string> = {
  CONSTRUCT: "Construct State (مضاف) — 1st term of iḍāfa (no tanwīn, no al-)",
  NOT_CONSTRUCT: "Non-Construct — free-standing noun",
};

// ===================================================================
// Invariable_Declinable — Why a word is fixed or variable
// ===================================================================

const INVARIABLE_MAP: Record<string, string> = {
  INVAR: "Invariable (مبني) — fixed form, does not change with case",
  DECLN: "Declinable (معرب) — changes form with grammatical case",
  DEF_ART: "Definite Article — al- is always invariable",
  DEM_NOUN: "Demonstrative Noun — most demonstratives are invariable",
  REL_PRON: "Relative Pronoun — alladhī, allatī are invariable",
  DISJ_PRON: "Disjoined Pronoun — independent pronoun (huwa, hiya)",
  DISJ_INV_PRON: "Disjoined Invariable Pronoun",
  COND_NOUN: "Conditional Noun — man, mā are invariable",
  INTROG_NOUN: "Interrogative Noun — man, mā interrogative",
  JONT_PRON: "Joined Pronoun — attached pronoun suffix",
  INV_NUM_COMP: "Invariable Number Compound — numbers 11-19",
};

// ===================================================================
// Phrase — Clause/Sentence type
// ===================================================================

const PHRASE_MAP: Record<string, string> = {
  NOM_SNT: "Nominal Sentence (جملة اسمية) — begins with a noun/topic",
  VERB_SNT: "Verbal Sentence (جملة فعلية) — begins with a verb",
  PHRASE: "Phrase (شبه جملة) — prepositional or adverbial phrase",
  NON_GOV_REL_CLS: "Non-Governing Relative Clause (صلة الموصول)",
  DER_GERND_CLS: "Derived Gerund Clause",
  COND_SNT: "Conditional Sentence (جملة شرطية)",
  NON_GOV_SNT: "Non-Governing Sentence",
};

// ===================================================================
// Phrasal_Function — Function within a clause/sentence
// ===================================================================

const PHRASAL_MAP: Record<string, string> = {
  PRED: "Predicate (خبر) — what is said about the subject",
  APPOS: "Appositive (بدل) — renames or clarifies a noun",
  CIRCUM: "Circumstantial (حال) — describes the state/condition",
  PART_COP_PRED: "Predicate of Copula Particle — khabar of inna/annahu/etc.",
  V_COP_PRED: "Predicate of Copula Verb — khabar of kāna",
  AGNT: "Agent (فاعل) — doer of a verbal action",
  PASS_SUBJ: "Passive Subject (نائب فاعل) — subject of passive verb",
  NEG_CAT_PRED: "Predicate of Negative Category",
  ADJ: "Adjective (صفة) — describes a preceding noun",
};

// ===================================================================
// Public API
// ===================================================================

const ALL_MAPS: Record<string, string>[] = [
  POS_MAP, ROLE_MAP, CASE_MAP, MORPH_TYPE_MAP, CONSTRUCT_MAP,
  INVARIABLE_MAP, MARKER_MAP, PHRASE_MAP, PHRASAL_MAP,
];

/** Translate any MASAQ term to learner-friendly English with Arabic */
export function translate(term: string | undefined | null): string {
  if (!term || term === "null" || term === "undefined" || term === "None") return "";
  for (const map of ALL_MAPS) {
    if (map[term]) return map[term];
  }
  return term.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

/** Compact short label for tooltips (1-3 words) */
export function translateShort(term: string | undefined | null): string {
  if (!term) return "";
  const shortForms: Record<string, string> = {
    PREP: "Prep", DET: "al-", CONJ: "Conj", NEG_PART: "Neg",
    INTERROG: "Interr", SUBJUNC_PART: "Subj", JUSSIVE_PART: "Juss",
    NOUN_ABSTRACT: "Abstract N", NOUN_CONCRETE: "Concrete N",
    NOUN_PROP: "Proper N", NOUN_ACTIVE_PART: "Active P",
    NOUN_PASSIVE_PART: "Passive P", GERUND: "Gerund",
    PRON: "Pron", SUBJ_PRON: "Subj Pron", OBJ_PRON: "Obj Pron",
    POSS_PRON: "Poss Pron", DEM_PRON: "Dem Pron", REL_PRON: "Rel Pron",
    IV: "Imperf V", CV: "Imperat V", PV: "Perf V",
    SUBJ: "Subj", TOPIC: "Topic", PRED: "Pred", OBJ: "Obj",
    ADJ: "Adj", GEN_CONS: "Iḍāfa", APPOS: "Appos",
    CIRCUM: "Ḥāl", VOC: "Voc", EXCP: "Exc",
    NOMINATIVE: "Nom", GENITIVE: "Gen", ACCUSATIVE: "Acc", INVARIABLE: "Inv",
    Prefix: "Pref", Stem: "Stem", Suffix: "Suff",
    CONSTRUCT: "Mudāf", NOT_CONSTRUCT: "—",
    INVAR: "Invariable", DECLN: "Declinable", DEF_ART: "Def Art",
    DHAMMA: "Ḍamma", KASRA: "Kasra", FATHA: "Fatḥa", SUKUN: "Sukūn",
    NOM_SNT: "Nominal Sn", VERB_SNT: "Verbal Sn", PHRASE: "Phrase",
    JUSSIVE: "Juss", CONDITION_PART: "Cond",
    IMPERF_PREF: "Imperf Pref", PVSUFF_SUBJ_3MS: "Perf 3MS",
    NOUN_NUM: "Numeral", NSUFF_FEM_SG: "Fem Sg", NSUFF_MASC_PL_GEN: "M.Pl Gen",
    CASE_DEF_ACC: "Def Acc", CASE_DEF_GEN: "Def Gen", CASE_DEF_NOM: "Def Nom",
    DEM_PRON_MS: "Dem M.Sg", REL_ADV: "Rel Adv", ADV_PLCE: "Adv Place",
    ADV_TIME: "Adv Time", KAAFA_MAKFOUFA: "Kāffa", VOC_PART: "Voc Part",
    EMPHATIC_NUN: "Emph Nūn",
  };
  if (shortForms[term]) return shortForms[term];
  const full = translate(term);
  if (full.length > 18) return full.substring(0, 15) + "…";
  return full;
}

/** Determine MASAQ column category from a term value */
export function categorizeTerm(term: string): string {
  if (POS_MAP[term]) return "Morph_Tag (POS)";
  if (ROLE_MAP[term]) return "Syntactic_Role";
  if (CASE_MAP[term]) return "Case_Mood";
  if (MORPH_TYPE_MAP[term]) return "Morph_Type";
  if (MARKER_MAP[term]) return "Case_Mood_Marker";
  if (CONSTRUCT_MAP[term]) return "Possessive_Construct";
  if (INVARIABLE_MAP[term]) return "Invariable_Declinable";
  if (PHRASE_MAP[term]) return "Phrase";
  if (PHRASAL_MAP[term]) return "Phrasal_Function";
  return "";
}
