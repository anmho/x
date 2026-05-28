#!/usr/bin/env python3
"""Materialize declarative platform config snapshots from infra/platform/declarative_spec.py."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

from infra.platform.declarative_spec import materialized_documents  # noqa: E402


def render_document(payload: dict) -> str:
    return json.dumps(payload, indent=2, sort_keys=False) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true", help="fail if generated files are out of date")
    args = parser.parse_args()

    dirty: list[str] = []
    for relative_path, payload in materialized_documents().items():
        target = ROOT / relative_path
        rendered = render_document(payload)
        if args.check:
            current = target.read_text(encoding="utf-8") if target.exists() else None
            if current != rendered:
                dirty.append(relative_path)
            continue

        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(rendered, encoding="utf-8")
        print(f"wrote {relative_path}")

    if dirty:
        for relative_path in dirty:
            print(f"out of date: {relative_path}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
