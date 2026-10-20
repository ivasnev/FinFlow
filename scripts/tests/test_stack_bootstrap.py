"""Проверки возобновления настройки TVM и разделения секретов сервисов."""

import importlib.util
import json
from pathlib import Path
import stat
import tempfile
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import unittest
import urllib.error


SPEC = importlib.util.spec_from_file_location("bootstrap", Path(__file__).resolve().parents[1] / "stack" / "bootstrap.py")
bootstrap = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(bootstrap)


class BootstrapTests(unittest.TestCase):
    """Проверяет состояние TVM через настоящий локальный HTTP-сервер."""

    def setUp(self):
        """Создаёт независимые состояние, каталог и HTTP-обработчик."""
        self.services = {}
        self.access = set()
        self.fail_name = None
        self.created = 0
        owner = self

        class Handler(BaseHTTPRequestHandler):
            """Предоставляет минимальный контракт административного API TVM."""

            def log_message(self, *args):
                """Подавляет служебный журнал тестового сервера."""

            def respond(self, status, data):
                """Возвращает JSON-ответ."""
                self.send_response(status)
                self.end_headers()
                self.wfile.write(json.dumps(data).encode())

            def do_GET(self):
                """Возвращает ключ существующего сервиса."""
                service_id = int(self.path.split("/")[2])
                service = owner.services.get(service_id)
                self.respond(200 if service else 404, {"public_key": service["public_key"]} if service else {})

            def do_POST(self):
                """Создаёт сервис или выдаёт доступ при правильном токене."""
                if self.headers.get("Authorization") != "Bearer test-admin":
                    self.respond(401, {})
                    return
                request = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
                if self.path == "/service":
                    if request["name"] == owner.fail_name:
                        self.respond(500, {})
                        return
                    owner.created += 1
                    service_id = 100 + owner.created
                    service = dict(id=service_id, name=request["name"], public_key=f"public-{service_id}", private_key=f"private-{service_id}")
                    owner.services[service_id] = service
                    self.respond(201, service)
                else:
                    owner.access.add((request["from"], request["to"]))
                    self.respond(200, {})

        server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        self.addCleanup(server.server_close)
        self.addCleanup(thread.join)
        self.addCleanup(server.shutdown)
        self.base = f"http://127.0.0.1:{server.server_port}"
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.state = Path(self.temp.name)

    def test_restart_keeps_keys_and_separates_credentials(self):
        """Повторная настройка сохраняет ключи, а контейнеру выдаётся только его собственный секрет."""
        first = bootstrap.configure(self.base, "test-admin", self.state)
        second = bootstrap.configure(self.base, "test-admin", self.state)
        self.assertEqual(first, second)
        self.assertEqual(self.created, 4)
        self.assertEqual(len(self.access), 4)
        for name in bootstrap.SERVICES:
            directory = self.state / name
            self.assertEqual((directory / "service.secret").read_text(), first[name]["private_key"])
            self.assertEqual((directory / "service.id").read_text(), str(first[name]["id"]))
            for path in directory.iterdir():
                self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)
                self.assertNotEqual(path.read_text(), first["ff-auth"]["private_key"] if name != "ff-auth" else first["ff-id"]["private_key"])
        self.assertEqual((self.state / "ff-auth" / "id-service.id").read_text(), str(first["ff-id"]["id"]))

    def test_partial_setup_resumes(self):
        """Сохранённые ключи первых сервисов не теряются при сбое последующего POST."""
        self.fail_name = "ff-files"
        with self.assertRaises(urllib.error.HTTPError):
            bootstrap.configure(self.base, "test-admin", self.state)
        partial = json.loads((self.state / "manifest" / "tvm.json").read_text())
        self.assertEqual(len(partial), 2)
        self.fail_name = None
        complete = bootstrap.configure(self.base, "test-admin", self.state)
        self.assertEqual(self.created, 4)
        for name in partial:
            self.assertEqual(partial[name], complete[name])

    def test_mismatched_database_fails_without_rotating_keys(self):
        """Потеря БД при сохранённом томе ключей не вызывает скрытую перерегистрацию."""
        bootstrap.configure(self.base, "test-admin", self.state)
        self.services.clear()
        with self.assertRaises(urllib.error.HTTPError):
            bootstrap.configure(self.base, "test-admin", self.state)
        self.assertEqual(self.created, 4)

    def test_empty_admin_token_is_rejected(self):
        """Настройка без административного токена завершается до HTTP-вызовов."""
        with self.assertRaises(RuntimeError):
            bootstrap.configure(self.base, "", self.state)
        self.assertEqual(self.created, 0)


if __name__ == "__main__":
    unittest.main()
