"""
Search — TokenSearch (token/substring/phrase) and AnalysisTable.
"""

from __future__ import annotations

import csv, io
from pathlib import Path

from .types import CharacterType, DiacriticTypes, ArabicCharacter
from .model import Document, Location
from .text import _CHAR_WIDTH
from .encoding import decode as _unicode_decode


class TokenSearch:
    __slots__ = ('_doc', '_criteria')

    def __init__(self):
        self._doc = Document.get()
        self._criteria: list[tuple[str, bool, bool]] = []

    def find_token(self, text: str, remove_diacritics: bool = False):
        self._criteria.append((text, False, remove_diacritics))
        return self

    def find_substring(self, text: str, remove_diacritics: bool = False):
        self._criteria.append((text, True, remove_diacritics))
        return self

    def get_results(self):
        table = AnalysisTable("Chapter", "Verse", "Token", "Text")
        seen: set[Location] = set()
        for query_text, is_substr, rm_diac in self._criteria:
            self._execute(query_text, is_substr, rm_diac, table, seen)
        return table

    def _execute(self, query: str, is_substr: bool, rm_diac: bool,
                 table, seen: set[Location]):
        query_parts = query.split()
        if len(query_parts) > 1:
            self._phrase(query_parts, rm_diac, table, seen)
            return

        q_chars = _unicode_decode(query)
        if rm_diac:
            q_chars = [ArabicCharacter(c.char_type, DiacriticTypes.NONE) for c in q_chars]
        if not q_chars: return
        q_pairs = [(int(c.char_type), int(c.diacritics)) for c in q_chars]
        q_len = len(q_pairs)

        for ch in self._doc.chapters:
            for verse in ch:
                for token in verse:
                    loc = token.location
                    if loc in seen:
                        continue
                    buf = token.buffer
                    off = token.offset
                    n = token.get_length()
                    tok_pairs = [
                        (buf[off + i * _CHAR_WIDTH],
                         0 if rm_diac else buf[off + i * _CHAR_WIDTH + 1])
                        for i in range(n)
                    ]
                    if is_substr:
                        for start in range(len(tok_pairs) - q_len + 1):
                            if all(tok_pairs[start + i] == q_pairs[i] for i in range(q_len)):
                                seen.add(loc)
                                table.add(loc.chapter, loc.verse, loc.token, str(token))
                                break
                    else:
                        if len(tok_pairs) == q_len and all(tok_pairs[i] == q_pairs[i] for i in range(q_len)):
                            seen.add(loc)
                            table.add(loc.chapter, loc.verse, loc.token, str(token))

    def _phrase(self, query_parts: list[str], rm_diac: bool, table, seen: set[Location]):
        term_pairs: list[list[tuple[int, int]]] = []
        for part in query_parts:
            qc = _unicode_decode(part)
            if rm_diac:
                qc = [ArabicCharacter(c.char_type, DiacriticTypes.NONE) for c in qc]
            if not qc: return
            term_pairs.append([(int(c.char_type), int(c.diacritics)) for c in qc])

        for ch in self._doc.chapters:
            for verse in ch:
                tokens = verse.tokens
                for ti in range(len(tokens)):
                    if ti + len(term_pairs) > len(tokens): break
                    match = True
                    matched = []
                    for qi in range(len(term_pairs)):
                        token = tokens[ti + qi]
                        buf = token.buffer
                        off = token.offset
                        n = token.get_length()
                        tp = [(buf[off + i * _CHAR_WIDTH],
                               0 if rm_diac else buf[off + i * _CHAR_WIDTH + 1])
                              for i in range(n)]
                        if len(tp) != len(term_pairs[qi]) or any(
                            tp[i] != term_pairs[qi][i] for i in range(len(tp))
                        ):
                            match = False; break
                        matched.append(token)
                    if match and matched:
                        loc = matched[0].location
                        if loc not in seen:
                            seen.add(loc)
                            table.add(loc.chapter, loc.verse, loc.token, str(matched[0]))


class AnalysisTable:
    __slots__ = ('_cols', '_col_idx', '_rows')

    def __init__(self, *cols: str):
        self._cols = list(cols)
        self._col_idx = {n: i for i, n in enumerate(cols)}
        self._rows: list[list] = []

    def add(self, *values):
        self._rows.append(list(values))

    def get_row_count(self) -> int: return len(self._rows)
    def get_column_count(self) -> int: return len(self._cols)

    def get_value(self, row: int, col: int | str):
        ci = col if isinstance(col, int) else self._col_idx[col]
        return self._rows[row][ci]

    def get_integer(self, row: int, col: int | str) -> int:
        return int(self.get_value(row, col))

    def get_string(self, row: int, col: int | str) -> str:
        return str(self.get_value(row, col))

    def sort(self, col_name: str, reverse: bool = False):
        ci = self._col_idx[col_name]
        self._rows.sort(key=lambda r: self._sort_key(r[ci]), reverse=reverse)
        return self

    @staticmethod
    def _sort_key(v):
        if isinstance(v, str):
            try: return int(v)
            except ValueError: pass
        return v

    def group(self, *col_names: str):
        from collections import Counter
        indices = [self._col_idx[c] for c in col_names]
        grouped = Counter()
        for row in self._rows:
            grouped[tuple(row[i] for i in indices)] += 1
        result = AnalysisTable(*col_names, "Count")
        for key in sorted(grouped):
            result.add(*key, grouped[key])
        return result

    def to_string(self, row_count: int | None = None) -> str:
        if not self._rows: return "(empty)"
        n = min(row_count or len(self._rows), len(self._rows))
        widths = [max(len(str(self._cols[i])), max((len(str(r[i])) for r in self._rows[:n]), default=0))
                  for i in range(len(self._cols))]
        lines = [" | ".join(str(self._cols[i]).ljust(widths[i]) for i in range(len(self._cols)))]
        lines.append("-" * len(lines[0]))
        for row in self._rows[:n]:
            lines.append(" | ".join(str(row[i]).ljust(widths[i]) for i in range(len(self._cols))))
        return "\n".join(lines)

    def write_csv(self, path: str | Path):
        with open(str(path), "w", newline="", encoding="utf-8") as f:
            w = csv.writer(f)
            w.writerow(self._cols)
            for row in self._rows: w.writerow(row)

    def __len__(self): return len(self._rows)
    def __iter__(self): return iter(self._rows)
    def __repr__(self): return f"AnalysisTable(rows={len(self._rows)}, cols={len(self._cols)})"
