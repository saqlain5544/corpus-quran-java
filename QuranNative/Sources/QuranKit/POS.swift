import Foundation

public enum POSType: String, CaseIterable, Sendable {
    case noun, verb, particle, pronoun, other
}

private let NOUN_TAGS: Set<String> = [
    "NOUN", "NOUN_ABSTRACT", "NOUN_ACTIVE_PART", "NOUN_PASSIVE_PART", "NOUN_VERBAL",
    "NOUN_PROP", "NOUN_CONCRETE", "NOUN_NUM", "NOUN_ADJECT", "NOUN_QUANT",
    "NOUN_ADVERB", "NOUN_PN", "NOUN_DIMINUTIVE", "NOUN_FIVE", "NOUN_INSTRUMENT",
    "NOUN_RELATIVE", "NOUN_TIME_PLACE", "NOUN_VERB_LIKE",
]

private let VERB_TAGS: Set<String> = [
    "IV", "PV", "CV", "IMPF_V", "PERF_V", "IMP_V",
    "PV_PASS", "IV_PASS", "IMPERATIVE", "IMPERATIVE_VERB",
    "UNINFLECTED_VERB", "CV_PREF",
]

private let PARTICLE_TAGS: Set<String> = [
    "PREP", "CONJ", "DET", "NEG_PART", "EMPHATIC_PART", "INTERROG_PART",
    "INTERJ", "INTERJ_PART", "CONDITION_PART", "EXCEPT_PART",
    "FUT_PART", "FUTUR_PART", "FUTURE_PART", "JUSSIVE_PART",
    "ANNUL_PART", "INF_ANNUL_PART", "INF_SUBJUNC_PART", "SUBJUNC_PART",
    "CERT_PART", "VOC_PART", "INTERROG", "INTERROG_PRON",
    "SURPRISE_PART", "RESULT_PART", "YES_NO_RESP_PART",
]

private let PRONOUN_TAGS: Set<String> = [
    "PRON", "PRON_DEM", "PRON_REL", "PRON_INTER", "PRON_EXCL",
    "PRON_1P", "PRON_2P", "PRON_2F", "PRON_2M", "PRON_3P", "PRON_3F", "PRON_3M",
    "PRON_3MS", "PRON_3D", "PRON_3FD", "PRON_3FS", "PRON_3MD",
    "PRON_3MP", "PRON_3FP", "PRON_1S", "PRON_2MP", "PRON_2MS",
    "PRON_1P_DUAL", "PRON_2D", "PRON_2MD", "PRON_2FD",
    "SUBJ_PRON", "OBJ_PRON", "POSS_PRON",
    "REL_PRON", "DEM_PRON", "DEM_PRON_F", "DEM_PRON_FS",
    "DEM_PRON_MP", "DEM_PRON_MS", "DEM_PRON_M", "DEM_PRON_MD", "DEM_PRON_FD",
]

public func isNoun(_ pos: String) -> Bool { NOUN_TAGS.contains(pos) || pos.hasPrefix("NOUN_") }
public func isVerb(_ pos: String) -> Bool { VERB_TAGS.contains(pos) }
public func isParticle(_ pos: String) -> Bool { PARTICLE_TAGS.contains(pos) }
public func isPronoun(_ pos: String) -> Bool { PRONOUN_TAGS.contains(pos) }

public func categorize(_ pos: String) -> POSType {
    if isNoun(pos) { return .noun }
    if isVerb(pos) { return .verb }
    if isParticle(pos) { return .particle }
    if isPronoun(pos) { return .pronoun }
    return .other
}


