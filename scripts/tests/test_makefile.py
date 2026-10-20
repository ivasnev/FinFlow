"""Проверки корневых команд тестирования без запуска настоящих Go-тестов."""

import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
MODULES = ["ff-auth", "ff-files", "ff-id", "ff-split", "ff-tvm", "ff-common"]


class MakefileTests(unittest.TestCase):
    """Проверяет остановку на ошибке и правильные каталоги всех модулей."""

    def run_target(self, target, failing=None, missing=None):
        """Запускает настоящий make с подставным Go и записывает обращения к нему."""
        with tempfile.TemporaryDirectory(prefix="finflow-make-") as temp:
            directory = Path(temp)
            shutil.copyfile(ROOT / "Makefile", directory / "Makefile")
            for module in MODULES:
                if module != missing:
                    (directory / module).mkdir()
            fake_go = directory / "fake-go"
            fake_go.write_text(
                "#!/usr/bin/env python3\n"
                "import json, os, pathlib, sys\n"
                "module = pathlib.Path.cwd().name\n"
                "with open(os.environ['FINFLOW_MAKE_TRACE'], 'a') as trace:\n"
                "    trace.write(json.dumps({'module': module, 'args': sys.argv[1:]}) + '\\n')\n"
                "sys.exit(7 if module == os.environ.get('FINFLOW_FAIL_MODULE') else 0)\n"
            )
            fake_go.chmod(0o700)
            trace = directory / "trace.jsonl"
            env = os.environ.copy()
            env.update(FINFLOW_MAKE_TRACE=str(trace), FINFLOW_FAIL_MODULE=failing or "")
            result = subprocess.run(
                ["make", target, f"GO={fake_go}"],
                cwd=directory, env=env, capture_output=True, text=True, timeout=10,
            )
            calls = [json.loads(line) for line in trace.read_text().splitlines()] if trace.exists() else []
            return result, calls

    def test_success_covers_all_six_modules(self):
        """Обе команды посещают каждый модуль ровно один раз в его каталоге."""
        for target in ("test", "test-cov"):
            with self.subTest(target=target):
                result, calls = self.run_target(target)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual([call["module"] for call in calls], MODULES)
                for call in calls:
                    self.assertEqual(call["args"][0], "test")
                    self.assertIn("./...", call["args"])
                    self.assertEqual("-coverprofile=coverage.out" in call["args"], target == "test-cov")

    def test_failure_stops_and_propagates(self):
        """Ошибка первого или среднего модуля не маскируется последующим echo."""
        for target in ("test", "test-cov"):
            for failing in ("ff-auth", "ff-id"):
                with self.subTest(target=target, failing=failing):
                    result, calls = self.run_target(target, failing=failing)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertEqual([call["module"] for call in calls], MODULES[:MODULES.index(failing) + 1])
                    self.assertNotIn("All module tests passed.", result.stdout)
                    self.assertNotIn("All coverage tests passed", result.stdout)

    def test_missing_module_fails(self):
        """Отсутствующий каталог тоже завершает команду с ошибкой."""
        for target in ("test", "test-cov"):
            with self.subTest(target=target):
                result, calls = self.run_target(target, missing="ff-id")
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual([call["module"] for call in calls], MODULES[:2])


if __name__ == "__main__":
    unittest.main()
