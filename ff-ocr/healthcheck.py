"""Проверка готовности без shell-кавычек в Docker/Podman healthcheck."""
import urllib.request

with urllib.request.urlopen('http://127.0.0.1:8086/health/ready', timeout=3) as response:
    response.read()
