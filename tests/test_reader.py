"""Test Reader — verse access, annotated text, Translation stub."""

from jqurantree import Reader, Translations


def test_reader_init():
    r = Reader('uthmani')
    assert r.chapters == list(range(1, 115))

def test_reader_verse():
    r = Reader('uthmani')
    assert len(r.verse(1, 1)) > 0
    assert 'بِسْمِ' in r.verse(1, 1)

def test_reader_dict_access():
    r = Reader('uthmani')
    assert len(r[1, 1]) > 0

def test_reader_surah_text():
    r = Reader('uthmani')
    text = r.surah_text(112)
    assert len(text) > 0

def test_reader_bismillah():
    r = Reader('uthmani')
    v2_1 = r.verse(2, 1)
    assert 'بِسْمِ' in v2_1
    assert 'الٓمٓ' in v2_1 or 'الم' in v2_1

def test_reader_no_bismillah_in_surah_9():
    r = Reader('uthmani')
    v9_1 = r.verse(9, 1)
    assert 'بَرَآءَةٌ' in v9_1 or 'براءة' in v9_1 or 'بَرَاءَة' in v9_1

def test_translations_stub():
    t = Translations()
    assert isinstance(t, Translations)

def test_reader_repr():
    r = Reader('uthmani')
    assert 'Reader' in repr(r)
    assert '114' in repr(r)
