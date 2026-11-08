#!/usr/bin/env python3
"""
Сверяет diff PR с docs/requirements.md через LLM.
Формирует таблицу с цветными галочками и сохраняет в /tmp/requirements_report.md.
"""

import argparse
import json
import os
import re
import subprocess
import sys

import requests

REQ_RE = re.compile(r"^##\s+(REQ-[A-Z0-9-]+):\s*(.+)$", re.MULTILINE)
ICON = {"implemented": "✅", "partial": "🟡", "missing": "❌"}


def gh(*args: str) -> str:
    out = subprocess.run(["gh", *args], capture_output=True, text=True, check=False)
    if out.returncode != 0:
        raise RuntimeError(f"gh {' '.join(args)} failed:\n{out.stderr}")
    return out.stdout.strip()


def load_requirements(path: str) -> list[dict]:
    with open(path, "r", encoding="utf-8") as f:
        text = f.read()
    return [
        {"id": m.group(1), "title": m.group(2).strip()}
        for m in REQ_RE.finditer(text)
    ]


def call_llm(diff: str, reqs: list[dict], api_key: str) -> dict:
    base_url = (os.environ.get("LLM_BASE_URL") or "https://api.deepseek.com").rstrip("/")
    req_list = "\n".join(f"- {r['id']}: {r['title']}" for r in reqs)
    prompt = f"""Ты — инженер по требованиям. Сравни diff pull request с требованиями.

Требования:
{req_list}

Diff PR (может быть обрезан):
{diff[:12000]}

Для каждого требования верни JSON-объект:
{{
  "results": [
    {{"id": "REQ-XXX", "status": "implemented|partial|missing", "note": "краткое пояснение"}}
  ]
}}

Правила:
- "implemented" — в diff есть код/документация, реализующая требование.
- "partial" — реализовано частично.
- "missing" — в diff нет ничего, что реализует требование.
- Если требование не относится к коду (например, про обучение агента), всё равно верни "missing".
"""
    resp = requests.post(
        f"{base_url}/chat/completions",
        headers={"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"},
        json={
            "model": os.environ.get("LLM_MODEL", "deepseek-chat"),
            "messages": [{"role": "user", "content": prompt}],
            "response_format": {"type": "json_object"},
        },
        timeout=120,
    )
    resp.raise_for_status()
    return json.loads(resp.json()["choices"][0]["message"]["content"])


def render_table(results: list[dict], reqs: list[dict]) -> str:
    by_id = {r["id"]: r for r in reqs}
    lines = [
        "## Requirements coverage",
        "",
        "| Требование | Статус | Пояснение |",
        "|---|---|---|",
    ]
    counts = {"implemented": 0, "partial": 0, "missing": 0}
    for item in results:
        rid = item.get("id", "?")
        status = item.get("status", "missing")
        counts[status] = counts.get(status, 0) + 1
        title = by_id.get(rid, {}).get("title", "")
        icon = ICON.get(status, "❓")
        note = (item.get("note") or "").replace("|", "\\|")
        lines.append(f"| `{rid}` {title} | {icon} {status} | {note} |")
    lines.append("")
    lines.append(
        f"**Итого:** ✅ {counts['implemented']} · 🟡 {counts['partial']} · ❌ {counts['missing']}"
    )
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", required=True)
    parser.add_argument("--pr", required=True)
    parser.add_argument("--requirements", required=True)
    args = parser.parse_args()

    api_key = os.environ.get("DEEPSEEK_API_KEY")
    if not api_key:
        print("DEEPSEEK_API_KEY не задан", file=sys.stderr)
        return 1

    reqs = load_requirements(args.requirements)
    if not reqs:
        print("В requirements.md не найдено ни одного REQ-*", file=sys.stderr)
        return 1

    diff = gh("pr", "diff", args.pr, "--repo", args.repo)
    result = call_llm(diff, reqs, api_key)
    table = render_table(result.get("results", []), reqs)

    with open("/tmp/requirements_report.md", "w", encoding="utf-8") as f:
        f.write(table)

    print(table)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception as e:
        print(f"Ошибка requirements check: {e}", file=sys.stderr)
        sys.exit(1)