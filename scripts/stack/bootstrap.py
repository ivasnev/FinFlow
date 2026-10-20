"""Подготовка TVM-учётных данных для локального Compose без ключей в репозитории."""

import json
import os
from pathlib import Path
import tempfile
import time
import urllib.error
import urllib.request


SERVICES = ("ff-split", "ff-id", "ff-files", "ff-auth")
ACCESS = (("ff-auth", "ff-id"), ("ff-split", "ff-id"), ("ff-split", "ff-files"), ("ff-id", "ff-files"))


def request_json(base, endpoint, token, payload=None):
    """Обращается к TVM, не записывая ключи и административный токен в журнал."""
    data = None if payload is None else json.dumps(payload).encode()
    headers = {"Authorization": "Bearer " + token, "Content-Type": "application/json"}
    request = urllib.request.Request(base + endpoint, data=data, headers=headers)
    with urllib.request.urlopen(request, timeout=5) as response:
        return json.load(response)


def wait_ready(url, allowed=(200,), timeout=60):
    """Ожидает HTTP-сервис с ограниченным общим сроком запуска."""
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        try:
            with urllib.request.urlopen(url, timeout=2) as response:
                if response.status in allowed:
                    return
        except urllib.error.HTTPError as error:
            try:
                if error.code in allowed:
                    return
            finally:
                error.close()
        except (OSError, TimeoutError):
            pass
        time.sleep(0.25)
    raise RuntimeError("service readiness timeout")


def atomic_write(path, data):
    """Сохраняет состояние заменой файла с доступом только владельцу."""
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
    name = None
    try:
        with tempfile.NamedTemporaryFile(dir=path.parent, delete=False) as file:
            name = file.name
            os.fchmod(file.fileno(), 0o600)
            file.write(data.encode())
            file.flush()
            os.fsync(file.fileno())
        os.replace(name, path)
    finally:
        if name and os.path.exists(name):
            os.unlink(name)


def configure(base, token, state):
    """Создаёт недостающие сервисы, проверяет сохранённые ключи и выдаёт минимальные связи доступа."""
    if not token:
        raise RuntimeError("TVM_ADMIN_TOKEN is required")
    state = Path(state)
    manifest_path = state / "manifest" / "tvm.json"
    manifest = json.loads(manifest_path.read_text()) if manifest_path.exists() else {}
    for name in SERVICES:
        if name not in manifest:
            service = request_json(base, "/service", token, {"name": name})
            if not isinstance(service.get("id"), int) or service["id"] <= 0 or not service.get("private_key") or not service.get("public_key"):
                raise RuntimeError("TVM returned invalid service metadata")
            manifest[name] = service
            # Запись после каждого ответа позволяет продолжить прерванную настройку.
            atomic_write(manifest_path, json.dumps(manifest))
        service = manifest[name]
        public = request_json(base, f"/service/{service['id']}/key", token)
        if public.get("public_key") != service["public_key"]:
            raise RuntimeError("saved credentials do not match TVM database; restore matching volumes")
    for source, destination in ACCESS:
        request_json(base, "/access/grant", token, {"from": manifest[source]["id"], "to": manifest[destination]["id"]})
    for name in SERVICES:
        directory = state / name
        atomic_write(directory / "service.id", str(manifest[name]["id"]))
        atomic_write(directory / "service.secret", manifest[name]["private_key"])
        if name in ("ff-auth", "ff-split"):
            atomic_write(directory / "id-service.id", str(manifest["ff-id"]["id"]))
        if name in ("ff-id", "ff-split"):
            atomic_write(directory / "file-service.id", str(manifest["ff-files"]["id"]))
    return manifest


if __name__ == "__main__":
    tvm = os.environ.get("TVM_BASE_URL", "http://ff-tvm:8081")
    wait_ready(tvm + "/service/0/key", allowed=(404,))
    wait_ready(os.environ.get("MINIO_HEALTH_URL", "http://minio:9000/minio/health/live"))
    configure(tvm, os.environ["TVM_ADMIN_TOKEN"], os.environ.get("TVM_STATE_DIR", "/state"))
    print("Configured four TVM services and access rules.")
