"""Test alignment — NFC, rasm automaton, alignment index."""

from jqurantree.alignment import Align, align
from jqurantree.rasm import normalize_rasm, rasm_cost, word_cost, needleman_wunsch


def test_nfc_diacritic_reorder():
    hafs = 'رَبِّكَ'   # Shadda(33)→Kasra(32)
    warsh = 'رَبِّكَ'  # Kasra(32)→Shadda(33)
    assert align(hafs, Align.canonical()) == align(warsh, Align.canonical())

def test_wasla_equivalence():
    cfg = Align.visual()
    assert align('ٱلْحَمْدُ', cfg) == align('الْحَمْدُ', cfg)

def test_dagger_alif_stripped():
    cfg = Align.phonetic()
    assert align('مَٰلِكِ', cfg) == align('مَالِكِ', cfg)

def test_exact_preserves_difference():
    assert align('ٱلْحَمْدُ', Align.exact()) != align('الْحَمْدُ', Align.exact())

def test_rasm_normalize():
    assert normalize_rasm('ٱلْحَمْدُ') == 'الحمد'
    assert normalize_rasm('مَٰلِكِ') == 'ملك'

def test_rasm_cost_identity():
    assert rasm_cost('ا', 'ا') == 0.0
    assert rasm_cost('ب', 'ت') == 1.0

def test_rasm_cost_equivalence():
    assert rasm_cost('\u0671', '\u0627') == 0.0  # Wasla ↔ Alif
    assert rasm_cost('\u0649', '\u064a') == 0.0  # Alif Maksura ↔ Ya

def test_needleman_wunsch():
    score, a, b = needleman_wunsch('العلمين', 'العالمين')
    assert score < 1.0  # Should be low cost (missing alif)

def test_word_cost_identity():
    assert word_cost('بسم', 'بسم') == 0.0

def test_word_cost_equivalent():
    assert word_cost('العلمين', 'العالمين') < 0.2  # Rasm difference

def test_all_presets():
    for preset in [Align.exact(), Align.canonical(), Align.visual(),
                   Align.structural(), Align.phonetic()]:
        result = align('بِسْمِ', preset)
        assert isinstance(result, str)
