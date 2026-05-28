#!/usr/bin/env python3
"""
Discovers stack services from stack.json files.

Usage:
  python3 scripts/dev-stack-discover.py list          # print service names, one per line
  python3 scripts/dev-stack-discover.py config <name> # print JSON config for service
"""

from __future__ import annotations

import json
import os
import re
import sys
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parent.parent
PG_URL = os.environ.get("PG_URL", "postgresql://postgres:postgres@127.0.0.1:54322")
ENV_EXPR = re.compile(r"\$\{([^}:]+)(?::-([^}]*))?\}")


def walk(directory: Path, files: list[Path]) -> None:
    if not directory.exists():
        return
    try:
        for entry in directory.iterdir():
            if entry.name == "stack.json" and entry.is_file():
                files.append(entry)
            elif entry.is_dir() and entry.name != "node_modules":
                walk(entry, files)
    except Exception:
        # Match the previous implementation's best-effort traversal.
        return


def load_services() -> list[dict[str, Any]]:
    files: list[Path] = []
    walk(ROOT / "services", files)
    walk(ROOT / "apps", files)
    walk(ROOT / "docs", files)

    services: list[dict[str, Any]] = []
    for file_path in files:
        payload = json.loads(file_path.read_text(encoding="utf-8"))
        entries = payload.get("services")
        if isinstance(entries, list):
            for service in entries:
                if not isinstance(service, dict):
                    continue
                item = dict(service)
                cwd = item.get("cwd", "")
                item["cwd"] = str(ROOT / str(cwd))
                services.append(item)
        elif isinstance(payload.get("name"), str):
            item = dict(payload)
            cwd = item.get("cwd", "")
            item["cwd"] = str(ROOT / str(cwd))
            services.append(item)
    return services


def expand(value: Any) -> Any:
    if not isinstance(value, str):
        return value

    def replace(match: re.Match[str]) -> str:
        key = match.group(1)
        default = match.group(2) or ""
        return os.environ.get(key, default)

    return ENV_EXPR.sub(replace, value)


def esc_single_quotes(value: str) -> str:
    return value.replace("'", "'\\''")


def find_service(services: list[dict[str, Any]], name: str) -> dict[str, Any] | None:
    for service in services:
        if service.get("name") == name:
            return service
    return None


def databases(services: list[dict[str, Any]]) -> list[str]:
    seen: set[str] = set()
    values: list[str] = []
    for service in services:
        db = service.get("db")
        if isinstance(db, str) and db and db not in seen:
            seen.add(db)
            values.append(db)
    return values


def service_env(service: dict[str, Any], expand_values: bool) -> dict[str, str]:
    env = dict(service.get("env") or {})
    db = service.get("db")
    if isinstance(db, str) and db:
        env["DATABASE_URL"] = f"{PG_URL}/{db}?sslmode=disable"

    result: dict[str, str] = {}
    for key, value in env.items():
        key_s = str(key)
        value_s = str(expand(value) if expand_values else value)
        result[key_s] = value_s
    return result


def main() -> int:
    services = load_services()
    cmd = sys.argv[1] if len(sys.argv) > 1 else None
    arg = sys.argv[2] if len(sys.argv) > 2 else None

    if cmd == "list":
        for service in services:
            name = service.get("name")
            if isinstance(name, str):
                print(name)
        return 0

    if cmd == "databases":
        for db in databases(services):
            print(db)
        return 0

    if cmd == "config" and arg:
        service = find_service(services, arg)
        if not service:
            return 1
        payload = {
            "name": service.get("name"),
            "cwd": service.get("cwd"),
            "run": service.get("run"),
            "env": service_env(service, expand_values=True),
        }
        print(json.dumps(payload))
        return 0

    if cmd == "config-bash" and arg:
        service = find_service(services, arg)
        if not service:
            return 1
        cwd = str(service.get("cwd", ""))
        run = str(service.get("run", ""))
        print(f"cwd='{esc_single_quotes(cwd)}'")
        print(f"run='{esc_single_quotes(run)}'")
        for key, value in service_env(service, expand_values=True).items():
            print(f"export {key}='{esc_single_quotes(value)}'")
        return 0

    return 1


if __name__ == "__main__":
    raise SystemExit(main())
