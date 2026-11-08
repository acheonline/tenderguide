#!/usr/bin/env bash
set -euo pipefail

THRESHOLD=50

# Генерируем профиль покрытия
go test -coverprofile=/tmp/coverage.out -covermode=atomic ./... > /dev/null

# Извлекаем общий процент
TOTAL=$(go tool cover -func=/tmp/coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')

echo "Total coverage: ${TOTAL}% (threshold: ${THRESHOLD}%)"

# Сравниваем
awk -v c="$TOTAL" -v t="$THRESHOLD" 'BEGIN { if (c+0 < t+0) exit 1 }' || {
  echo "FAIL: coverage ${TOTAL}% is below threshold ${THRESHOLD}%"
  exit 1
}

echo "OK: coverage meets threshold."