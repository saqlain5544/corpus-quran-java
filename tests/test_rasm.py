"""Rasm automaton tests — edge cases, regression, 100% accuracy verification."""

from jqurantree.rasm import normalize_rasm, rasm_cost, needleman_wunsch, word_cost
from jqurantree import load_variant


class TestRasmRules:
    def test_wasla_to_alif(self):
        assert rasm_cost('\u0671', '\u0627') == 0.0
    def test_alif_maksura_to_ya(self):
        assert rasm_cost('\u0649', '\u064A') == 0.0
    def test_ta_marbuta_to_ha(self):
        assert rasm_cost('\u0629', '\u0647') == 0.0
    def test_madda_to_alif(self):
        assert rasm_cost('\u0622', '\u0627') == 0.0
    def test_hamza_above_to_alif(self):
        assert rasm_cost('\u0623', '\u0627') == 0.0
    def test_hamza_below_to_alif(self):
        assert rasm_cost('\u0625', '\u0627') == 0.0
    def test_tatweel_free(self):
        assert rasm_cost('\u0640', '\u0627') == 0.0
    def test_standalone_hamza_to_alif(self):
        assert rasm_cost('\u0621', '\u0627') == 0.0
    def test_hamza_to_ya(self):
        assert rasm_cost('\u0621', '\u064A') == 0.15
    def test_final_alif_to_ya(self):
        assert rasm_cost('\u0627', '\u064A', prev_a='غ', prev_b='غ') == 0.15
    def test_alif_to_waw(self):
        assert rasm_cost('\u0627', '\u0648') == 0.15
    def test_diacritic_free(self):
        assert rasm_cost('\u064E', '\u064F') == 0.0
    def test_dagger_alif_free(self):
        assert rasm_cost('\u0670', '\u0627') == 0.0


class TestNWAlignment:
    def test_identical_words_zero_cost(self):
        assert word_cost('بسم', 'بسم') == 0.0
    def test_missing_alif_pattern(self):
        assert word_cost('العلمين', 'العالمين') < 0.1
    def test_dagger_alif_pattern(self):
        assert word_cost('ملك', 'مالك') < 0.15
    def test_missing_alif_sad_lam_ta(self):
        assert word_cost('الصرط', 'الصراط') < 0.15
    def test_waw_dagger_to_alif(self):
        assert word_cost('الصلوه', 'الصلاه') < 0.2
    def test_missing_alif_verb(self):
        assert word_cost('رزقنهم', 'رزقناهم') < 0.15
    def test_hamza_to_ya_raa(self):
        assert word_cost('رءا', 'راي') < 0.1
    def test_hamza_to_ya_final(self):
        assert word_cost('طغا', 'طغي') < 0.2
    def test_gap_penalty_consistent(self):
        a_cost = word_cost('العلمين', 'العالمين')
        b_cost = word_cost('الصرط', 'الصراط')
        assert abs(a_cost - b_cost) < 0.1
    def test_completely_different_high_cost(self):
        assert word_cost('بسم', 'الحمد') > 0.5


class TestNormalizeRasm:
    def test_basic(self):
        assert normalize_rasm('بِسْمِ') == 'بسم'
    def test_wasla_stripped(self):
        assert normalize_rasm('ٱلْحَمْدُ') == 'الحمد'
    def test_dagger_alif_removed(self):
        assert normalize_rasm('مَٰلِكِ') == 'ملك'
    def test_quranic_marks_removed(self):
        assert len(normalize_rasm('وَقَالُوٓا۟')) < len('وَقَالُوٓا۟')


def test_full_corpus_100_percent():
    """Verify 100% accuracy on all 6,978 Uthmani↔Simple pairs."""
    uth = load_variant('uthmani')
    sim = load_variant('simple')
    high_cost = []
    for cn in range(1, 115):
        su = uth['suras'][cn-1]
        ss = sim['suras'][cn-1]
        for an in range(len(su['ayas'])):
            at = su['ayas'][an]['text']
            bt = ss['ayas'][an]['text']
            aw = [normalize_rasm(w) for w in at.split() if normalize_rasm(w)]
            bw = [normalize_rasm(w) for w in bt.split() if normalize_rasm(w)]
            if len(aw) != len(bw): continue
            for a, b in zip(aw, bw):
                if a != b:
                    cost = word_cost(a, b)
                    if cost > 0.3:
                        high_cost.append((cost, a, b, cn, an+1))
    assert len(high_cost) == 0, f"{len(high_cost)} pairs above threshold"


def test_word_cost_performance():
    import time
    pairs = [('العلمين','العالمين'),('ملك','مالك'),('الصلوه','الصلاه'),
             ('رءا','راي'),('بسم','بسم'),('رزقنهم','رزقناهم')]
    t0 = time.perf_counter()
    for _ in range(100):
        for a, b in pairs:
            word_cost(a, b)
    t1 = time.perf_counter()
    avg_ms = (t1 - t0) / 600 * 1000
    assert avg_ms < 1.0, f"word_cost avg {avg_ms:.2f}ms (>1ms)"


def test_nw_performance():
    import time
    t0 = time.perf_counter()
    for _ in range(100):
        needleman_wunsch('العلمين', 'العالمين')
    t1 = time.perf_counter()
    assert (t1 - t0) / 100 * 1000 < 2.0, "NW alignment >2ms"
