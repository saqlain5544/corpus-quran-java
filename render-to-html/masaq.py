import csv
from collections import defaultdict


class Masaq:
    def __init__(self, path):
        self.index = defaultdict(list)

        with open(path, "r", encoding="utf8") as f:
            for row in csv.DictReader(f):
                key = (
                    int(row["Sura_No"]),
                    int(row["Verse_No"]),
                    int(row["Word_No"]),
                )

                self.index[key].append(row)

    def word(
        self,
        surah,
        ayah,
        word_no,
    ):
        """
        Get MASAQ segments for:

        Surah + Ayah + Word
        """

        return self.index.get(
            (
                surah,
                ayah,
                word_no,
            ),
            [],
        )

    def gloss(
        self,
        surah,
        ayah,
        word_no,
    ):

        return [row["Gloss"] for row in self.word(surah, ayah, word_no) if row["Gloss"]]

    def tags(
        self,
        surah,
        ayah,
        word_no,
    ):

        return [
            row["Morph_Tag"]
            for row in self.word(surah, ayah, word_no)
            if row["Morph_Tag"]
        ]
