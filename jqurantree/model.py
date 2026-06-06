"""
Model — Document, Chapter, Verse, Token, Location.
"""

from __future__ import annotations

from .types import ArabicCharacter
from .text import ArabicText


class Location:
    __slots__ = ('_c', '_v', '_t')

    def __init__(self, chapter: int, verse: int, token: int):
        self._c = chapter; self._v = verse; self._t = token

    @property
    def chapter(self) -> int: return self._c
    @property
    def verse(self) -> int: return self._v
    @property
    def token(self) -> int: return self._t

    def __hash__(self): return hash((self._c, self._v, self._t))
    def __eq__(self, o): return isinstance(o, Location) and self._c == o._c and self._v == o._v and self._t == o._t
    def __lt__(self, o): return (self._c, self._v, self._t) < (o._c, o._v, o._t) if isinstance(o, Location) else NotImplemented
    def __repr__(self): return f"({self._c}:{self._v}:{self._t})"


class Token(ArabicText):
    __slots__ = ('_location',)

    def __init__(self, buffer: bytes = b"", offset: int = 0, char_count: int = 0,
                 location: Location | None = None):
        super().__init__(buffer, offset, char_count)
        self._location = location or Location(0, 0, 0)

    @property
    def location(self) -> Location: return self._location

    @staticmethod
    def from_characters(chars: list[ArabicCharacter], location: Location) -> "Token":
        from .text import ArabicTextBuilder
        builder = ArabicTextBuilder()
        for ac in chars: builder.add(ac)
        base = builder.to_arabic_text()
        t = Token.__new__(Token)
        t._buffer = base.buffer
        t._offset = base.offset
        t._character_count = base.get_length()
        t._location = location
        return t

    def __repr__(self): return f"Token({self._location}, len={self._character_count})"


class Verse:
    __slots__ = ('_tokens', '_location')

    def __init__(self, tokens: list[Token], location: Location):
        self._tokens = tuple(tokens)
        self._location = location

    @property
    def tokens(self) -> tuple[Token, ...]: return self._tokens
    @property
    def location(self) -> Location: return self._location
    @property
    def verse_number(self) -> int: return self._location.verse
    @property
    def token_count(self) -> int: return len(self._tokens)

    def __len__(self): return len(self._tokens)
    def __getitem__(self, i): return self._tokens[i]
    def __iter__(self): return iter(self._tokens)

    def __repr__(self): return f"Verse({self._location}, tokens={len(self._tokens)})"


class Chapter:
    __slots__ = ('_verses', '_number', '_name', '_bismillah')

    def __init__(self, number: int, name: ArabicText, bismillah: ArabicText | None, verses: list[Verse]):
        self._verses = tuple(verses); self._number = number
        self._name = name; self._bismillah = bismillah

    @property
    def chapter_number(self) -> int: return self._number
    @property
    def verses(self) -> tuple[Verse, ...]: return self._verses
    @property
    def name(self) -> ArabicText: return self._name
    @property
    def bismillah(self) -> ArabicText | None: return self._bismillah
    @property
    def verse_count(self) -> int: return len(self._verses)
    @property
    def token_count(self) -> int: return sum(v.token_count for v in self._verses)

    def get_verse(self, n: int) -> Verse:
        if n < 1 or n > len(self._verses): raise IndexError(f"Verse {n}")
        return self._verses[n - 1]

    def __len__(self): return len(self._verses)
    def __getitem__(self, i): return self._verses[i]
    def __iter__(self): return iter(self._verses)
    def __repr__(self): return f"Chapter({self._number}, verses={len(self._verses)})"


class Document:
    __slots__ = ('_chapters', '_init')
    _instance: "Document | None" = None

    def __init__(self):
        self._chapters: tuple[Chapter, ...] = tuple([None] * 114)  # type: ignore
        self._init = False

    @classmethod
    def get(cls) -> "Document":
        if cls._instance is None: cls._instance = Document()
        return cls._instance

    @classmethod
    def reset(cls):
        cls._instance = None

    @property
    def chapters(self) -> tuple[Chapter, ...]:
        if not self._init: raise RuntimeError("Document not initialized")
        return self._chapters

    def _ensure(self):
        if not self._init: raise RuntimeError("Document not initialized")

    @staticmethod
    def chapter_count() -> int: return 114

    @staticmethod
    def verse_count() -> int: return 6236

    def token_count(self) -> int:
        self._ensure()
        return sum(c.token_count for c in self._chapters)

    def get_chapter(self, n: int) -> Chapter:
        self._ensure()
        return self._chapters[n - 1]

    def set_chapters(self, chapters: list[Chapter]):
        arr: list[Chapter | None] = [None] * 114
        for ch in chapters:
            cn = ch.chapter_number
            if arr[cn - 1] is not None: raise ValueError(f"Duplicate chapter {cn}")
            arr[cn - 1] = ch
        for i, ch in enumerate(arr):
            if ch is None: raise ValueError(f"Missing chapter {i + 1}")
        self._chapters = tuple(arr)  # type: ignore
        self._init = True

    def __repr__(self):
        return f"Document(ch=114, v=6236)" if self._init else "Document(uninitialized)"
