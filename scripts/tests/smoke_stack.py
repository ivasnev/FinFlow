"""HTTP-проверка полного локального стека и сохранности данных после перезапуска."""

import argparse
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.error
import urllib.request
import uuid


def endpoint(service):
    """Возвращает опубликованный адрес сервиса с учётом переменных Compose."""
    ports = {"AUTH": 8084, "ID": 8083, "SPLIT": 8085, "TVM": 8081, "FILES": 8082}
    base = "http://localhost:" + os.environ.get(service + "_PORT", str(ports[service]))
    return base + ("/api/v1" if service == "AUTH" else "")


def request(service, path, payload=None, token=None, method=None, headers=None):
    """Выполняет запрос, не выводя пароли, токены и подписанные ссылки."""
    actual_headers = dict(headers or {})
    if token:
        actual_headers["Authorization"] = "Bearer " + token
    data = payload
    if isinstance(payload, dict):
        data = json.dumps(payload).encode()
        actual_headers["Content-Type"] = "application/json"
    req = urllib.request.Request(endpoint(service) + path, data=data, headers=actual_headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=15) as response:
            return json.load(response)
    except urllib.error.HTTPError as error:
        error.close()
        raise RuntimeError(f"{service} {method or ('POST' if data else 'GET')} {path}: HTTP {error.code}") from None


def wait_ready():
    """Ожидает готовность всех приложений перед изменением тестовых данных."""
    probes = (("AUTH", "/auth/public-key"), ("ID", "/api/v1/users/me/friend-requests"),
              ("SPLIT", "/api/v1/event"), ("FILES", "/api/v1/files/00000000-0000-0000-0000-000000000000"))
    deadline = time.monotonic() + 90
    for service, path in probes:
        while True:
            try:
                with urllib.request.urlopen(endpoint(service) + path, timeout=2):
                    pass
                break
            except urllib.error.HTTPError as error:
                error.close()
                if error.code == 401:
                    break
                if time.monotonic() >= deadline:
                    raise
            except (OSError, TimeoutError):
                if time.monotonic() >= deadline:
                    raise RuntimeError("stack readiness timeout") from None
            time.sleep(0.25)


def files_ticket():
    """Получает билет от имени split, читая только его собственный том ключей."""
    completed = subprocess.run(
        ["docker", "compose", "exec", "-T", "ff-split", "sh", "-c",
         "cat /run/tvm/service.id; printf '\\n'; cat /run/tvm/service.secret; printf '\\n'; cat /run/tvm/file-service.id"],
        check=True, capture_output=True, text=True,
    )
    source, secret, destination = completed.stdout.strip().splitlines()
    ticket = request("TVM", "/ticket", {"from": int(source), "to": int(destination), "secret": secret})
    return {"X-FF-Service-Ticket": ticket["ticket"]}


def verify(state):
    """Проверяет вход, профиль, расход, оптимизированные долги и ранее загруженный файл."""
    auth = request("AUTH", "/auth/login", {"login": state["email"], "password": state["password"]})
    token = auth["access_token"]
    assert auth["user"]["id"] == state["user_id"]
    profile = request("ID", "/api/v1/users/" + auth["user"]["nickname"])
    assert profile["id"] == state["user_id"] and profile["name"] == "Compose profile"
    event = "/api/v1/event/" + str(state["event_id"])
    assert request("SPLIT", event, token=token)["id"] == state["event_id"]
    debts = request("SPLIT", event + "/optimized-debts", token=token)["optimized_debts"]
    assert len(debts) == 1 and debts[0]["amount"] == 12.34, "debt was not preserved"
    result = request("FILES", "/api/v1/files/" + state["object_id"], headers=files_ticket())
    with urllib.request.urlopen(result["data"], timeout=10) as response:
        assert response.read() == b"FinFlow Compose smoke\n", "file content mismatch"


def create():
    """Создаёт тестовые данные через реальные HTTP-интерфейсы всех сервисов."""
    unique = uuid.uuid4().hex
    state = {"email": unique + "@example.com", "password": uuid.uuid4().hex}
    auth = request("AUTH", "/auth/register", {**state, "nickname": unique, "name": "Compose Smoke"})
    state["user_id"] = auth["user"]["id"]
    token = auth["access_token"]
    request("ID", "/api/v1/users/me", {"name": "Compose profile"}, token, method="PATCH")
    event = request("SPLIT", "/api/v1/event", {"name": "Compose smoke", "members": {"dummies_names": ["Guest"]}}, token)
    state["event_id"] = event["id"]
    path = "/api/v1/event/" + str(event["id"])
    users = request("SPLIT", path + "/user", token=token)["users"]
    payer = next(user["internal_id"] for user in users if user.get("user_id") == state["user_id"])
    guest = next(user["internal_id"] for user in users if user.get("user_id") != state["user_id"])
    request("SPLIT", path + "/transaction", {"name": "Smoke expense", "amount": 12.34,
            "from_user": payer, "type": "units", "users": [guest]}, token)
    boundary = "finflow-" + unique
    body = (f"--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"smoke.txt\"\r\n"
            "Content-Type: text/plain\r\n\r\nFinFlow Compose smoke\n" + f"\r\n--{boundary}--\r\n").encode()
    headers = {**files_ticket(), "Content-Type": "multipart/form-data; boundary=" + boundary}
    uploaded = request("FILES", "/api/v1/files", body, headers=headers)
    state["object_id"] = uploaded["data"]["object_id"]
    verify(state)
    return state


def main():
    """Запускает сценарий создания либо проверку сохранённого состояния."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--state", type=Path)
    parser.add_argument("--verify", action="store_true")
    args = parser.parse_args()
    if args.verify and not args.state:
        parser.error("--verify requires --state")
    wait_ready()
    if args.verify:
        verify(json.loads(args.state.read_text()))
    else:
        state = create()
        if args.state:
            descriptor = os.open(args.state, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
            with os.fdopen(descriptor, "w") as output:
                os.fchmod(output.fileno(), 0o600)
                json.dump(state, output)
    print("PASS: registration, login, profile, event, expense, optimized debt, file content")


if __name__ == "__main__":
    main()
