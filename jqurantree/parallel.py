"""
Parallel executor — adaptive ProcessPool/ThreadPool dispatch for
CPU-bound and I/O-bound tasks across all modules.
"""

from __future__ import annotations

import os, sysconfig
from concurrent.futures import ProcessPoolExecutor, ThreadPoolExecutor, as_completed
from functools import partial
from typing import Callable, TypeVar, Iterator

T = TypeVar("T")


def _is_free_threaded() -> bool:
    return sysconfig.get_config_var("Py_GIL_DISABLED") is not None


def cpu_count() -> int:
    try:
        return os.process_cpu_count() or os.cpu_count() or 4
    except AttributeError:
        return os.cpu_count() or 4


def pmap(func: Callable, items: list, max_workers: int | None = None,
         chunksize: int = 1, unordered: bool = True) -> Iterator:
    """Parallel map — auto-selects ThreadPool (free-threaded) or ProcessPool."""
    n = max_workers or min(cpu_count(), len(items))
    if n <= 1:
        yield from map(func, items)
        return

    if _is_free_threaded():
        executor_cls = ThreadPoolExecutor
    else:
        executor_cls = ProcessPoolExecutor

    with executor_cls(max_workers=n) as ex:
        if chunksize > 1:
            futures = []
            for i in range(0, len(items), chunksize):
                chunk = items[i:i + chunksize]
                futures.append(ex.submit(_process_chunk, func, chunk))
            if unordered:
                for f in as_completed(futures):
                    yield from f.result()
            else:
                for f in futures:
                    yield from f.result()
        else:
            futures = {ex.submit(func, item): item for item in items}
            if unordered:
                for f in as_completed(futures):
                    yield f.result()
            else:
                for item in items:
                    pass


def _process_chunk(func, chunk):
    return [func(item) for item in chunk]


def parallel_surahs(func: Callable, max_workers: int | None = None) -> list:
    """Apply func(surah_number) to all 114 chapters in parallel."""
    items = list(range(1, 115))
    n = max_workers or cpu_count()
    n = min(n, len(items))
    if n <= 1:
        return [func(i) for i in items]
    with ProcessPoolExecutor(max_workers=n) as ex:
        futures = {ex.submit(func, i): i for i in items}
        results = {}
        for f in as_completed(futures):
            i = futures[f]
            results[i] = f.result()
    return [results[i] for i in range(1, 115)]
