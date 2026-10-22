"""Real subprocess signals with a native-style handler installed on lazy import."""
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import sys
import time
from concurrent.futures import ThreadPoolExecutor
import urllib.request

import pytest


@pytest.mark.skipif(os.name != 'posix', reason='POSIX native signal regression')
@pytest.mark.parametrize('signum,active', [(signal.SIGTERM, False), (signal.SIGTERM, True), (signal.SIGINT, False)])
def test_uvicorn_retains_shutdown_after_native_import(tmp_path, signum, active):
    # Like Paddle, this module changes sigaction behind CPython's handler cache.
    (tmp_path / 'paddle.py').write_text('''import ctypes, signal
libc = ctypes.CDLL(None)
libc.signal.argtypes = [ctypes.c_int, ctypes.c_void_p]
libc.signal.restype = ctypes.c_void_p
libc.signal(signal.SIGTERM, None)
def disable_signal_handler():
    libc.signal(signal.SIGTERM, None)
''')
    script = tmp_path / 'server.py'
    script.write_text('''import os, sys, time
from pathlib import Path
sys.path.insert(1, sys.argv[1])
os.environ['OCR_TOKEN'] = 'synthetic-local-ocr-token-32-chars'
import app, uvicorn
app.load_engine = lambda: __import__('paddle')
def recognize(data):
    Path(sys.argv[3]).write_text('started')
    time.sleep(.5)
    return {'items': []}
app.recognize = recognize
uvicorn.run(app.app, host='127.0.0.1', port=int(sys.argv[2]), log_level='info')
''')
    with socket.socket() as listener:
        listener.bind(('127.0.0.1', 0))
        port = listener.getsockname()[1]
    marker = tmp_path / 'started'
    log = tmp_path / 'server.log'
    pool = ThreadPoolExecutor(max_workers=1)
    request = None
    with log.open('w') as output:
        process = subprocess.Popen([sys.executable, str(script), str(Path(__file__).resolve().parents[1]), str(port), str(marker)], stdout=output, stderr=subprocess.STDOUT)
    try:
        deadline = time.monotonic() + 15
        while True:
            assert process.poll() is None, log.read_text()
            try:
                with urllib.request.urlopen(f'http://127.0.0.1:{port}/health/ready', timeout=.3) as response:
                    if response.status == 200:
                        break
            except OSError:
                pass
            assert time.monotonic() < deadline, log.read_text()
            time.sleep(.03)
        if active:
            def scan():
                req = urllib.request.Request(f'http://127.0.0.1:{port}/recognize', data=b'test', headers={'Authorization': 'Bearer synthetic-local-ocr-token-32-chars'})
                with urllib.request.urlopen(req, timeout=10) as response:
                    return json.load(response)
            request = pool.submit(scan)
            while not marker.exists():
                assert time.monotonic() < deadline
                time.sleep(.01)
        process.send_signal(signum)
        process.wait(timeout=10)
        assert process.returncode in (0, -signum), log.read_text()
        assert 'Application shutdown complete' in log.read_text(), log.read_text()
        if request:
            assert request.result(timeout=2) == {'items': []}
    finally:
        if process.poll() is None:
            process.kill()
            process.wait(timeout=5)
        pool.shutdown(wait=True)


def test_runtime_restores_handlers_and_mask_when_import_fails(monkeypatch):
    import builtins
    import app
    original_import = builtins.__import__
    handlers = {sig: signal.getsignal(sig) for sig in (signal.SIGTERM, signal.SIGINT)}
    mask = signal.pthread_sigmask(signal.SIG_BLOCK, []) if hasattr(signal, 'pthread_sigmask') else None

    def fail(name, *args, **kwargs):
        if name == 'paddle':
            raise ImportError('synthetic import failure')
        return original_import(name, *args, **kwargs)

    monkeypatch.setattr(builtins, '__import__', fail)
    with pytest.raises(ImportError, match='synthetic import failure'):
        app.prepare_runtime()
    assert {sig: signal.getsignal(sig) for sig in handlers} == handlers
    if mask is not None:
        assert signal.pthread_sigmask(signal.SIG_BLOCK, []) == mask
