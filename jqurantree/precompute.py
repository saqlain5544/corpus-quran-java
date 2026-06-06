"""
Precomputation & 100% cross-verification of MASAQ data across all variants.
Uses parallel processing for full-corpus alignment and validation.
"""

from __future__ import annotations

import sys, time, traceback
from collections import defaultdict
from concurrent.futures import ProcessPoolExecutor, as_completed
from dataclasses import dataclass, field
from pathlib import Path

from jqurantree import VARIANTS, load_variant
from jqurantree.morphology import Morphology, Segment
from jqurantree.alignment_index import Alignment
from jqurantree.rasm import normalize_rasm
from jqurantree.constants import VERSE_COUNTS, surah_name
from jqurantree.parallel import cpu_count


@dataclass(slots=True)
class VariantVerification:
    variant: str
    total_verses: int = 0
    total_words: int = 0
    aligned_words: int = 0
    masaq_matches: int = 0
    masaq_misses: int = 0
    root_matches: int = 0
    gloss_diffs_vs_ref: dict[str, int] = field(default_factory=dict)
    errors: list[str] = field(default_factory=list)

    @property
    def alignment_pct(self) -> float:
        return self.aligned_words / max(self.total_words, 1) * 100

    @property
    def masaq_pct(self) -> float:
        return self.masaq_matches / max(self.total_words, 1) * 100

    def summary(self) -> str:
        lines = [
            f"  {self.variant:<16s}",
            f"    Words:     {self.total_words:>6,d}",
            f"    Aligned:   {self.aligned_words:>6,d} ({self.alignment_pct:.2f}%)",
            f"    MASAQ ok:  {self.masaq_matches:>6,d} ({self.masaq_pct:.2f}%)",
            f"    MASAQ miss:{self.masaq_misses:>6,d}",
            f"    Root hits: {self.root_matches:>6,d}",
        ]
        if self.errors:
            lines.append(f"    Errors:    {len(self.errors)}")
            for e in self.errors[:3]:
                lines.append(f"      {e}")
        return "\n".join(lines)


def _verify_variant(vk: str) -> VariantVerification:
    """Process one variant — called via ProcessPoolExecutor."""
    m = Morphology()
    v = VariantVerification(variant=vk)
    try:
        data = load_variant(vk)
        alignment = Alignment(vk)

        for cn in range(1, 115):
            su = data["suras"][cn - 1]
            for an in range(1, len(su["ayas"]) + 1):
                aya = su["ayas"][an - 1]
                text = aya["text"]
                if not text.strip():
                    continue
                v.total_verses += 1
                tokens = text.split()
                wm = alignment.map(cn, an)

                for vi, token in enumerate(tokens, start=1):
                    v.total_words += 1
                    ref_wi = wm.variant_to_ref.get(vi, vi)
                    if ref_wi != vi or vi in wm.variant_to_ref:
                        v.aligned_words += 1

                    segs = m.word(cn, an, ref_wi)
                    if segs:
                        v.masaq_matches += 1
                        root = m.root_for(cn, an, ref_wi)
                        if root:
                            v.root_matches += 1
                    else:
                        v.masaq_misses += 1
                        if v.masaq_misses <= 5:
                            v.errors.append(
                                f"S{cn}:{an} w{vi}='{token}' → ref_w{ref_wi} no MASAQ"
                            )

    except Exception as e:
        v.errors.append(f"CRASH: {e}")

    return v


def verify_all(variants: list[str] | None = None, parallel: bool = True):
    """Precompute and verify all variants. Returns dict of VariantVerification."""
    if variants is None:
        variants = ["uthmani", "simple", "simple-clean", "indopak"]

    print("=" * 70)
    print("  FULL-CORPUS PRECOMPUTATION & VERIFICATION")
    print("=" * 70)
    print()

    m = Morphology()
    print(f"  MASAQ: {m.segment_count():,} segments, {m.word_count():,} words")
    print(f"  Roots: {len(m.roots()):,} roots")
    print()

    results: dict[str, VariantVerification] = {}
    t0 = time.perf_counter()

    if parallel and len(variants) > 1:
        n_workers = min(cpu_count(), len(variants))
        print(f"  Parallel mode: {n_workers} workers")
        with ProcessPoolExecutor(max_workers=n_workers) as ex:
            futures = {ex.submit(_verify_variant, vk): vk for vk in variants}
            for future in as_completed(futures):
                vk = futures[future]
                results[vk] = future.result()
                print(results[vk].summary())
                print()
    else:
        for vk in variants:
            results[vk] = _verify_variant(vk)
            print(results[vk].summary())
            print()

    t1 = time.perf_counter()
    print(f"  Completed in {(t1 - t0):.1f}s")
    print()

    # Cross-variant gloss consistency
    if "uthmani" in results and len(variants) > 1:
        print("  Cross-variant consistency:")
        ref_v = results["uthmani"]
        for vk in variants[1:]:
            v = results[vk]
            diff = v.masaq_matches - ref_v.masaq_matches
            print(f"    {vk:<16s} MASAQ delta: {diff:+d} from reference")
        print()

    # Summary
    total_w = sum(r.total_words for r in results.values())
    total_m = sum(r.masaq_matches for r in results.values())
    total_r = sum(r.root_matches for r in results.values())
    total_miss = sum(r.masaq_misses for r in results.values())

    print("=" * 70)
    print(f"  TOTAL: {total_w:,} words, {total_m:,} MASAQ hits, "
          f"{total_r:,} root hits, {total_miss} misses")
    print(f"  ACCURACY: {total_m/max(total_w,1)*100:.4f}%")
    print("=" * 70)

    return results


if __name__ == "__main__":
    verify_all()
