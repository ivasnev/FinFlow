"""Проверка завершения приватного OCR после отмены HTTP и начала shutdown."""
import asyncio
import threading
import app
import httpx


def test_shutdown_waits_for_cancelled_native_work(monkeypatch):
    started, release = threading.Event(), threading.Event()
    marker = object()
    token = 'synthetic-local-ocr-token-32-chars'
    monkeypatch.setattr(app, 'TOKEN', token)
    monkeypatch.setattr(app, 'busy', asyncio.Semaphore(1))
    monkeypatch.setattr(app, 'running', set())
    monkeypatch.setattr(app, 'prepare_runtime', lambda: None)
    monkeypatch.setattr(app, 'load_engine', lambda: marker)

    def recognize(_):
        started.set()
        release.wait(5)
        assert app.engine is marker, 'Модель выгружена во время нативного вызова'
        return {'items': []}
    monkeypatch.setattr(app, 'recognize', recognize)

    async def scenario():
        lifetime = app.lifespan(app.app)
        await lifetime.__aenter__()
        async with httpx.AsyncClient(transport=httpx.ASGITransport(app=app.app), base_url='http://test') as client:
            request = asyncio.create_task(client.post('/recognize', content=b'image', headers={'Authorization': 'Bearer '+token}))
            shutdown = None
            try:
                assert await asyncio.to_thread(started.wait, 2)
                request.cancel()
                await asyncio.gather(request, return_exceptions=True)
                shutdown = asyncio.create_task(lifetime.__aexit__(None, None, None))
                await asyncio.sleep(.03)
                assert not shutdown.done()
                assert app.engine is marker
                assert app.busy.locked()
            finally:
                release.set()
                await asyncio.gather(request, return_exceptions=True)
                if shutdown:
                    await asyncio.wait_for(shutdown, 2)
                else:
                    await lifetime.__aexit__(None, None, None)
        assert app.engine is None
        assert not app.running
        assert app.busy._value == 1
    asyncio.run(scenario())
