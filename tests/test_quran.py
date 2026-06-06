"""Test Quran unified API — gloss, lemmas, roots, POS across variants."""

from jqurantree import Quran


q = Quran("uthmani")
q_simple = Quran("simple")
q_indopak = Quran("indopak")
q_clean = Quran("simple-clean")


def test_variant_property():
    assert q.variant == "uthmani"
    assert q_simple.variant == "simple"

def test_gloss_fatiha():
    gloss = q.gloss(1, 1)
    assert len(gloss) == 4
    assert gloss[0] == "in-(the)-name"
    assert gloss[1] == "(of)-allah"

def test_gloss_ikhlas():
    gloss = q.gloss(112, 1)
    assert len(gloss) >= 3
    assert "say" in gloss

def test_lemmas():
    lemmas = q.lemmas(1, 1)
    assert len(lemmas) > 0

def test_roots():
    roots = q.roots(2, 255)
    assert "Alh" in roots

def test_pos():
    pos = q.pos(1, 1)
    assert "NOUN_ABSTRACT" in pos or "NOUN_PROP" in pos

def test_word_count():
    assert q.word_count(1) > 0
    assert q.word_count(2) > 0

def test_words_object():
    words = q.words(1, 1)
    assert len(words) == 4
    assert words[0].word_index == 1
    assert len(words[0].segments) > 0

def test_verse_text():
    assert len(q.verse_text(1, 1)) > 0
    assert len(q.verse_text(114, 1)) > 0

def test_cross_variant_gloss():
    ref = q.gloss(1, 1)
    assert q_simple.gloss(1, 1) == ref
    assert q_indopak.gloss(1, 1) == ref
    assert q_clean.gloss(1, 1) == ref

def test_cross_variant_roots():
    ref = q.roots(2, 255)
    assert q_simple.roots(2, 255) == ref

def test_unknown_variant():
    try:
        Quran("nonexistent")
        assert False, "Should have raised"
    except KeyError:
        pass
