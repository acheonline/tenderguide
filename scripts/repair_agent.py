#!/usr/bin/env python3
"""
Repair Agent: анализирует лог упавшего RepoKeeper, переписывает issue через LLM,
публикует diff, закрывает старый issue (not planned) и создаёт новый с меткой agent-todo.
"""

import argparse
import difflib
import json
import os
import re
import subprocess
import sys

import requests

# ---------- GitHub CLI helpers ----------

def gh(*args, check=True, capture=True):
    """Вызов gh CLI. Возвращает stdout."""
    result = subprocess.run(
        ["gh", *args],
        capture_output=capture,
        text=True,
        check=False,
    )
    if check and result.returncode != 0:
        raise RuntimeError(
            f"gh {' '.join(args)} failed (code {result.returncode}):\n{result.stderr}"
        )
    return result.stdout.strip() if capture else ""


def get_issue(repo: str, number: str) -> dict:
    raw = gh("issue", "view", number, "--repo", repo,
             "--json", "title,body,labels,number")
    data = json.loads(raw)
    data["labels"] = [lbl["name"] for lbl in data.get("labels", [])]
    return data


def comment_issue(repo: str, number: str, body: str) -> None:
    path = "/tmp/repair_comment.md"
    with open(path, "w", encoding="utf-8") as f:
        f.write(body)
    gh("issue", "comment", number, "--repo", repo, "--body-file", path)


def close_issue(repo: str, number: str, comment: str = "") -> None:
    """Закрывает issue с причиной 'not planned'. Не требует прав администратора."""
    cmd = ["issue", "close", number, "--repo", repo, "--reason", "not planned"]
    if comment:
        cmd.extend(["--comment", comment])
    gh(*cmd)


def create_issue(repo: str, title: str, body: str, labels: list[str]) -> str:
    path = "/tmp/repair_new_issue.md"
    with open(path, "w", encoding="utf-8") as f:
        f.write(body)
    cmd = ["issue", "create", "--repo", repo,
           "--title", title, "--body-file", path]
    for lbl in dict.fromkeys(labels):
        cmd.extend(["--label", lbl])
    out = gh(*cmd)
    return out.splitlines()[-1] if out else ""


# ---------- LLM ----------

def call_llm(log_text: str, old_issue: dict, api_key: str) -> dict:
    """Возвращает {"title": ..., "body": ..., "reason": ...}."""
    base_url = os.environ.get("LLM_BASE_URL", "https://api.deepseek.com").rstrip("/")
    prompt = f"""Ты — инженер по надёжности. Ниже — лог упавшего запуска GitHub Actions (агент RepoKeeper пытался создать PR по issue, но потерпел неудачу).

Старый issue:
Title: {old_issue['title']}
Body:
{old_issue['body']}

Лог (возможно, обрезан):
{log_text[:8000]}

Задача:
1. Проанализируй, почему PR не был создан.
2. Перепиши текст issue так, чтобы агент смог успешно выполнить задачу. Сохрани исходное намерение, но уточни или исправь то, что привело к сбою.
3. Верни строго JSON: {{"title": "...", "body": "...", "reason": "краткое объяснение изменений"}}.
"""
    resp = requests.post(
        f"{base_url}/chat/completions",
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        json={
            "model": os.environ.get("LLM_MODEL", "deepseek-chat"),
            "messages": [{"role": "user", "content": prompt}],
            "response_format": {"type": "json_object"},
        },
        timeout=120,
    )
    resp.raise_for_status()
    payload = resp.json()
    content = payload["choices"][0]["message"]["content"]
    try:
        return json.loads(content)
    except json.JSONDecodeError as e:
        raise RuntimeError(f"LLM вернул невалидный JSON: {e}\n---\n{content}")


# ---------- Защита от циклов ----------

ATTEMPT_RE = re.compile(r"<!--\s*repair-attempt:\s*(\d+)\s*-->")


def get_attempt(body: str) -> int:
    m = ATTEMPT_RE.search(body or "")
    return int(m.group(1)) if m else 0


def set_attempt(body: str, n: int) -> str:
    marker = f"<!-- repair-attempt: {n} -->"
    if ATTEMPT_RE.search(body or ""):
        return ATTEMPT_RE.sub(marker, body)
    return (body or "") + "\n\n" + marker


# ---------- main ----------

def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--log-file", required=True)
    parser.add_argument("--issue-number", required=True)
    parser.add_argument("--run-id", required=True)
    args = parser.parse_args()

    repo = os.environ["GITHUB_REPOSITORY"]
    api_key = os.environ.get("LLM_API_KEY")
    if not api_key:
        print("LLM_API_KEY не задан", file=sys.stderr)
        return 1

    max_attempts = int(os.environ.get("MAX_REPAIR_ATTEMPTS", "3"))

    # 1. Логи
    try:
        with open(args.log_file, "r", errors="replace") as f:
            log_text = f.read()
    except OSError as e:
        print(f"Не удалось прочитать лог: {e}", file=sys.stderr)
        return 1

    # 2. Старый issue
    try:
        old = get_issue(repo, args.issue_number)
    except RuntimeError as e:
        print(f"Не удалось получить issue #{args.issue_number}: {e}", file=sys.stderr)
        return 1

    attempt = get_attempt(old["body"])

    # 3. Защита от циклов
    if attempt >= max_attempts:
        msg = (
            f"⚠️ Достигнут лимит попыток починки ({attempt}/{max_attempts}). "
            f"Автоматическое пересоздание issue остановлено. Нужен ручной разбор."
        )
        try:
            comment_issue(repo, args.issue_number, msg)
        except RuntimeError:
            pass
        print(msg)
        return 0

    # 4. LLM
    try:
        new = call_llm(log_text, old, api_key)
    except Exception as e:
        print(f"LLM ошибка: {e}", file=sys.stderr)
        return 1

    new_title = (new.get("title") or "").strip()
    new_body = (new.get("body") or "").strip()
    reason = (new.get("reason") or "").strip()
    if not new_title or not new_body:
        print("LLM вернул пустой title или body", file=sys.stderr)
        return 1

    new_body = set_attempt(new_body, attempt + 1)

    # 5. Diff-комментарий
    diff = "\n".join(
        difflib.unified_diff(
            (old["body"] or "").splitlines(),
            new_body.splitlines(),
            fromfile="старый issue",
            tofile="новый issue",
            lineterm="",
        )
    )
    try:
        comment_issue(
            repo,
            args.issue_number,
            f"### Предлагаемое изменение issue (попытка {attempt + 1}/{max_attempts})\n\n"
            f"**Причина:** {reason}\n\n```diff\n{diff[:6000]}\n```",
        )
    except RuntimeError as e:
        print(f"Не удалось опубликовать diff-комментарий: {e}", file=sys.stderr)

    # 6. Закрываем старый issue (not planned) и создаём новый
    try:
        close_issue(
            repo,
            args.issue_number,
            f"Автоматически заменён новым issue после сбоя RepoKeeper "
            f"(попытка {attempt + 1}/{max_attempts}).",
        )
    except RuntimeError as e:
        print(f"Не удалось закрыть issue #{args.issue_number}: {e}", file=sys.stderr)
        return 1

    labels = ["agent-todo"] + [l for l in old["labels"] if l != "agent-todo"]
    try:
        new_url = create_issue(repo, new_title, new_body, labels)
    except RuntimeError as e:
        print(f"Не удалось создать новый issue: {e}", file=sys.stderr)
        return 1

    print(f"Создан новый issue: {new_url}")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception as e:
        print(f"Непредвиденная ошибка: {e}", file=sys.stderr)
        sys.exit(1)