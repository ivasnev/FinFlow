#!/bin/sh
set -eu
# pg_isready недостаточно: существующий том должен получить миграцию архива.
until [ "$(psql -X -qAt -v ON_ERROR_STOP=1 -c "SELECT to_regprocedure('event_purge_at(timestamptz)') IS NOT NULL" 2>/dev/null)" = t ]; do
 echo "Archive cleanup: waiting for archive migration"
 sleep 5
done
mkdir -p /var/spool/cron/crontabs
# 00:00 UTC соответствует 03:00 Europe/Moscow. Cron работает без зависимости от локального TZ хоста.
printf '0 0 * * * /bin/sh /stack/cleanup-events.sh\n' > /var/spool/cron/crontabs/root
chmod 600 /var/spool/cron/crontabs/root
exec busybox crond -f -d 8 -c /var/spool/cron/crontabs
