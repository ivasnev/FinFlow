#!/bin/sh
set -eu
# Один пакет атомарен. Ошибка psql останавливает запуск, а не превращается в вечный цикл.
total=0
while :; do
 count=$(psql -X -qAt -v ON_ERROR_STOP=1 -f /stack/purge-events.sql)
 case "$count" in
  -1) echo "Archive cleanup: another batch is running"; exit 0 ;;
  ''|*[!0-9]*) echo "Archive cleanup: unexpected batch result" >&2; exit 1 ;;
 esac
 total=$((total + count))
 [ "$count" -eq 100 ] || break
 sleep 1
done
printf 'Archive cleanup: removed %s events\n' "$total"
