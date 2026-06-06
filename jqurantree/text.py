"""
Text — ArabicText (byte-buffer), ArabicTextBuilder, ByteFormat.
"""

from __future__ import annotations

from .types import CharacterType, DiacriticTypes, ArabicCharacter

_CHAR_WIDTH = 3


class ByteFormat:
    @staticmethod
    def encode(ct: int, dc: int) -> bytes:
        return bytes([ct & 0xFF, dc & 0xFF, 0])


class ArabicText:
    __slots__ = ('_buffer', '_offset', '_character_count')

    def __init__(self, buffer: bytes = b"", offset: int = 0, character_count: int = 0):
        self._buffer = buffer
        self._offset = offset
        self._character_count = character_count

    @property
    def buffer(self) -> bytes: return self._buffer
    @property
    def offset(self) -> int: return self._offset

    def get_length(self) -> int: return self._character_count

    def get_character(self, index: int) -> ArabicCharacter:
        bo = self._offset + index * _CHAR_WIDTH
        return ArabicCharacter(CharacterType(self._buffer[bo]), DiacriticTypes(self._buffer[bo + 1]))

    def __len__(self) -> int: return self._character_count

    def __iter__(self):
        for i in range(self._character_count):
            yield self.get_character(i)

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, ArabicText): return NotImplemented
        if self._character_count != other._character_count: return False
        for i in range(self._character_count):
            si = self._offset + i * _CHAR_WIDTH
            oi = other._offset + i * _CHAR_WIDTH
            if self._buffer[si] != other._buffer[oi] or self._buffer[si + 1] != other._buffer[oi + 1]:
                return False
        return True

    def __hash__(self) -> int:
        return hash(self._buffer[self._offset:self._offset + self._character_count * _CHAR_WIDTH])

    def to_unicode(self) -> str:
        from .encoding import encode
        return encode(list(self))

    def __str__(self) -> str: return self.to_unicode()
    def __repr__(self) -> str: return f"ArabicText(length={self._character_count})"

    @staticmethod
    def from_unicode(text: str) -> ArabicText:
        from .encoding import decode
        builder = ArabicTextBuilder()
        for ac in decode(text):
            builder.add(ac)
        return builder.to_arabic_text()


class ArabicTextBuilder:
    __slots__ = ('_chars',)

    def __init__(self): self._chars: list[ArabicCharacter] = []

    def add(self, character: ArabicCharacter): self._chars.append(character); return self

    def to_arabic_text(self) -> ArabicText:
        if not self._chars:
            return ArabicText(b"", 0, 0)
        buf = bytearray(len(self._chars) * _CHAR_WIDTH)
        for i, ac in enumerate(self._chars):
            off = i * _CHAR_WIDTH
            buf[off] = ac.char_type.value & 0xFF
            buf[off + 1] = int(ac.diacritics) & 0xFF
        result = ArabicText.__new__(ArabicText)
        result._buffer = bytes(buf)
        result._offset = 0
        result._character_count = len(self._chars)
        return result

    def __len__(self) -> int: return len(self._chars)
