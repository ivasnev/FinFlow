#!/bin/sh
set -eu

# Каждый контейнер получает только собственный секрет и публичные ID его зависимостей.
export TVM_SERVICE_ID="$(cat /run/tvm/service.id)"
export TVM_SERVICE_SECRET="$(cat /run/tvm/service.secret)"
if [ -f /run/tvm/id-service.id ]; then
    export ID_TVM_ID="$(cat /run/tvm/id-service.id)"
    export ID_SERVICE_ID="$ID_TVM_ID"
fi
if [ -f /run/tvm/file-service.id ]; then
    export FILE_SERVICE_ID="$(cat /run/tvm/file-service.id)"
fi
exec /app/server "$@"
