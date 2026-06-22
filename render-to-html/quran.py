import xml.etree.ElementTree as ET
from dataclasses import dataclass, field
from functools import lru_cache

BISMILLAH_TEXT = "بِسْمِ ٱللَّهِ ٱلرَّحْمَـٰنِ ٱلرَّحِيمِ"


# ──────────────────────────────────────────────────────────
# Utility Functions
# ──────────────────────────────────────────────────────────


def normalize_text(text: str | None) -> str:
    """Normalize XML text values."""
    return (text or "").strip()


def is_bismillah(text: str | None) -> bool:
    """Check whether text is the standard Bismillah."""
    return normalize_text(text) == BISMILLAH_TEXT


def validate_surah_number(number: int) -> None:
    if not 1 <= number <= 114:
        raise ValueError(f"Invalid surah number: {number}")


def validate_ayah_number(number: int) -> None:
    if number < 1:
        raise ValueError(f"Invalid ayah number: {number}")


# ──────────────────────────────────────────────────────────
# Data Classes
# ──────────────────────────────────────────────────────────


@dataclass(slots=True)
class Verse:
    number: int
    text: str

    def words(self) -> list[str]:
        return self.text.split()

    def __str__(self) -> str:
        return self.text


@dataclass(slots=True)
class Surah:
    number: int
    name: str
    bismillah: str | None
    verses: dict[int, Verse] = field(default_factory=dict)

    def verse(self, number: int) -> Verse:
        validate_ayah_number(number)
        return self.verses[number]

    def verse_text(self, number: int) -> str:
        return self.verse(number).text

    def verse_count(self) -> int:
        return len(self.verses)

    def words(self):
        for verse in self.verses.values():
            yield from verse.words()

    def __iter__(self):
        return iter(self.verses.values())

    def __len__(self):
        return len(self.verses)

    def __getitem__(self, verse_number):
        return self.verse(verse_number)


@dataclass(slots=True)
class Quran:
    surahs: dict[int, Surah] = field(default_factory=dict)

    # ---------- Lookup ----------

    def surah(self, number: int) -> Surah:
        validate_surah_number(number)
        return self.surahs[number]

    def verse(
        self,
        surah_number: int,
        ayah_number: int,
    ) -> Verse:

        return self.surah(surah_number).verse(ayah_number)

    def verse_text(
        self,
        surah_number: int,
        ayah_number: int,
    ) -> str:

        return self.verse(
            surah_number,
            ayah_number,
        ).text

    def bismillah(
        self,
        surah_number: int,
    ) -> str | None:

        return self.surah(surah_number).bismillah

    def surah_name(
        self,
        surah_number: int,
    ) -> str:

        return self.surah(surah_number).name

    # ---------- Search ----------

    def search(self, query: str):

        results = []

        for surah in self:
            for verse in surah:
                if query in verse.text:
                    results.append(
                        (
                            surah.number,
                            verse.number,
                            verse.text,
                        )
                    )

        return results

    # ---------- Statistics ----------

    def total_surahs(self) -> int:
        return len(self.surahs)

    def total_verses(self) -> int:

        return sum(len(surah) for surah in self)

    def total_words(self) -> int:

        return sum(len(verse.words()) for surah in self for verse in surah)

    def longest_surah(self) -> Surah:

        return max(
            self,
            key=len,
        )

    # ---------- Iteration ----------

    def iter_verses(self):

        for surah in self:
            for verse in surah:
                yield (
                    surah.number,
                    verse.number,
                    verse,
                )

    # ---------- Dunder ----------

    def __getitem__(self, surah_number):
        return self.surah(surah_number)

    def __iter__(self):
        return iter(self.surahs.values())

    def __len__(self):
        return len(self.surahs)


# ──────────────────────────────────────────────────────────
# Parser
# ──────────────────────────────────────────────────────────


@lru_cache(maxsize=4)
def parse_quran(xml_path: str) -> Quran:
    """
    Parse Quran XML.

    Rules:
      - Surah 1: Bismillah is verse 1.
      - Surah 9: No Bismillah.
      - Other Surahs:
          Bismillah stored separately and
          not counted as a verse.
    """

    tree = ET.parse(xml_path)
    root = tree.getroot()

    quran = Quran()

    for sura_xml in root.findall("sura"):
        surah_number = int(sura_xml.attrib["index"])

        validate_surah_number(surah_number)

        if surah_number in quran.surahs:
            raise ValueError(f"Duplicate surah: {surah_number}")

        surah_name = normalize_text(sura_xml.attrib.get("name"))

        verses = {}
        bismillah = None

        expected_ayah = 1

        for aya_xml in sura_xml.findall("aya"):
            ayah_number = int(aya_xml.attrib["index"])

            validate_ayah_number(ayah_number)

            if ayah_number != expected_ayah:
                raise ValueError(
                    f"Surah {surah_number}: "
                    f"expected ayah "
                    f"{expected_ayah}, "
                    f"got {ayah_number}"
                )

            expected_ayah += 1

            ayah_text = normalize_text(aya_xml.attrib.get("text"))

            ayah_bismillah = normalize_text(aya_xml.attrib.get("bismillah"))

            # ──────────────────────
            # Surah 1
            # ──────────────────────

            if surah_number == 1:
                if ayah_number == 1:
                    bismillah = ayah_text

                verses[ayah_number] = Verse(
                    ayah_number,
                    ayah_text,
                )

                continue

            # ──────────────────────
            # Surah 9
            # ──────────────────────

            if surah_number == 9:
                verses[ayah_number] = Verse(
                    ayah_number,
                    ayah_text,
                )

                continue

            # ──────────────────────
            # Other Surahs
            # ──────────────────────

            if ayah_number == 1:
                if ayah_bismillah:
                    bismillah = ayah_bismillah

                elif is_bismillah(ayah_text):
                    bismillah = ayah_text

                # Skip standalone Bismillah
                if is_bismillah(ayah_text):
                    continue

            verses[ayah_number] = Verse(
                ayah_number,
                ayah_text,
            )

        quran.surahs[surah_number] = Surah(
            number=surah_number,
            name=surah_name,
            bismillah=bismillah,
            verses=verses,
        )

    if len(quran) != 114:
        raise ValueError(f"Expected 114 surahs, found {len(quran)}")

    return quran


# ──────────────────────────────────────────────────────────
# Standalone Helpers
# ──────────────────────────────────────────────────────────


def iter_surahs(quran: Quran):
    yield from quran


def iter_verses(quran: Quran):
    yield from quran.iter_verses()


def verse_words(
    quran: Quran,
    surah: int,
    ayah: int,
):
    return quran.verse_text(
        surah,
        ayah,
    ).split()


def find_text(
    quran: Quran,
    query: str,
):
    return quran.search(query)


def total_surahs(
    quran: Quran,
):
    return quran.total_surahs()


def total_verses(
    quran: Quran,
):
    return quran.total_verses()


def total_words(
    quran: Quran,
):
    return quran.total_words()
