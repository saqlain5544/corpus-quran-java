"""Test morphology — MASAQ queries, glosses, roots, POS."""

from jqurantree import Morphology


m = Morphology()


def test_segment_count():
    assert m.segment_count() == 157676

def test_word_count():
    assert m.word_count() == 77411

def test_bismillah_word():
    segs = m.word(1, 1, 1)
    assert len(segs) == 2  # ب + اسم
    assert segs[0].is_prefix
    assert segs[1].is_stem
    assert segs[0].pos == "PREP"
    assert segs[1].pos == "NOUN_ABSTRACT"

def test_gloss():
    assert m.gloss(1, 1, 1) == "in-(the)-name"
    assert m.gloss(1, 1, 2) == "(of)-allah"

def test_by_pos():
    nouns = m.by_pos("NOUN_PROP")
    assert len(nouns) > 3000

def test_search_gloss():
    results = m.search_gloss("allah")
    assert len(results) > 1000

def test_pos_distribution():
    dist = m.pos_distribution()
    assert len(dist) == 20
    assert dist.get("PREP", 0) > 10000

def test_roots():
    roots = m.roots()
    assert len(roots) == 1642

def test_root_for():
    assert m.root_for(1, 1, 2) == "Alh"  # الله → Alh

def test_verse():
    segs = m.verse(1, 1)
    assert len(segs) == 8  # 4 words × 2 segments each

def test_surah():
    segs = m.surah(112)
    assert len(segs) > 0
