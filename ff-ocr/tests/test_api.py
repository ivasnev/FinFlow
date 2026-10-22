"""Проверки приватного HTTP и освобождения ограниченного нативного ресурса."""
import asyncio
import io
import threading

import httpx
from PIL import Image
import pytest
import app

TOKEN = 'synthetic-local-ocr-token-32-chars'
AUTH = {'Authorization': 'Bearer ' + TOKEN}


def test_private_api_auth_body_limits_and_invalid_image(monkeypatch):
    monkeypatch.setattr(app, 'TOKEN', TOKEN)
    monkeypatch.setattr(app, 'engine', object())
    monkeypatch.setattr(app, 'busy', asyncio.Semaphore(1))
    monkeypatch.setattr(app, 'MAX_BYTES', 100)

    async def scenario():
        async with httpx.AsyncClient(transport=httpx.ASGITransport(app=app.app), base_url='http://test') as client:
            assert (await client.post('/recognize', content=b'x')).status_code == 401
            assert (await client.get('/metrics')).status_code == 401
            assert (await client.post('/recognize', content=b'x'*101, headers=AUTH)).status_code == 413
            assert (await client.post('/recognize', content=b'invalid', headers=AUTH)).status_code == 422
            assert (await client.get('/health/ready')).status_code == 200
    asyncio.run(scenario())
    with pytest.raises(app.HTTPException) as error:
        app.authorize('Bearer é')
    assert error.value.status_code == 401


def test_busy_and_cancel_do_not_start_another_native_job(monkeypatch):
    started, release = threading.Event(), threading.Event()
    calls = []

    def recognize(_data):
        calls.append(1)
        started.set()
        release.wait(timeout=5)
        return {'items': []}

    monkeypatch.setattr(app, 'TOKEN', TOKEN)
    monkeypatch.setattr(app, 'engine', object())
    monkeypatch.setattr(app, 'busy', asyncio.Semaphore(1))
    monkeypatch.setattr(app, 'recognize', recognize)

    async def scenario():
        async with httpx.AsyncClient(transport=httpx.ASGITransport(app=app.app), base_url='http://test') as client:
            task = asyncio.create_task(client.post('/recognize', content=b'image', headers=AUTH))
            try:
                assert await asyncio.to_thread(started.wait, 2)
                assert (await client.post('/recognize', content=b'image', headers=AUTH)).status_code == 429
                task.cancel()
                await asyncio.sleep(.02)
                assert (await client.post('/recognize', content=b'image', headers=AUTH)).status_code == 429
                assert (await client.get('/health/ready')).status_code == 200
            finally:
                release.set()
                with pytest.raises(asyncio.CancelledError):
                    await task
            await asyncio.gather(*tuple(app.running), return_exceptions=True)
            assert len(calls) == 1
            assert (await client.post('/recognize', content=b'image', headers=AUTH)).status_code == 200
    asyncio.run(scenario())


def test_image_pixel_limit_is_checked_before_model(monkeypatch):
    monkeypatch.setattr(app, 'MAX_PIXELS', 10)
    data = io.BytesIO()
    Image.new('RGB', (4, 4)).save(data, format='PNG')
    with pytest.raises(ValueError, match='invalid image'):
        app.recognize(data.getvalue())


def test_repeated_cancellation_keeps_the_native_slot_owned(monkeypatch):
    started, release = threading.Event(), threading.Event()
    calls = []

    def recognize(_data):
        calls.append(1)
        started.set()
        release.wait(timeout=5)
        return {'items': []}

    monkeypatch.setattr(app, 'TOKEN', TOKEN)
    monkeypatch.setattr(app, 'engine', object())
    monkeypatch.setattr(app, 'busy', asyncio.Semaphore(1))
    monkeypatch.setattr(app, 'recognize', recognize)

    async def scenario():
        async with httpx.AsyncClient(transport=httpx.ASGITransport(app=app.app), base_url='http://test') as client:
            task = asyncio.create_task(client.post('/recognize', content=b'image', headers=AUTH))
            try:
                assert await asyncio.to_thread(started.wait, 2)
                task.cancel()
                await asyncio.sleep(.01)
                task.cancel()
                with pytest.raises(asyncio.CancelledError):
                    await task
                assert app.busy.locked()
                assert (await client.post('/recognize', content=b'image', headers=AUTH)).status_code == 429
                assert len(calls) == 1
            finally:
                release.set()
                await asyncio.gather(*tuple(app.running), return_exceptions=True)
            assert not app.busy.locked()
    asyncio.run(scenario())


def test_orientation_fallback_preserves_first_pass_unless_candidate_is_coherent(monkeypatch):
    class Engine:
        def __init__(self, initial, rotated):
            self.initial, self.rotated, self.calls = initial, rotated, []

        def predict(self, _pixels, use_doc_orientation_classify):
            self.calls.append(use_doc_orientation_classify)
            texts = self.rotated if use_doc_orientation_classify else self.initial
            return [{'rec_texts': texts, 'rec_scores': [.99]*len(texts)}]

    data = io.BytesIO()
    Image.new('RGB', (10, 10)).save(data, format='PNG')
    upright = Engine(['Покупка 10.00', 'ИТОГО 10.00'], [])
    monkeypatch.setattr(app, 'engine', upright)
    assert app.recognize(data.getvalue())['total_cents'] == 1000
    assert upright.calls == [False]
    rotated = Engine([], ['Покупка 10.00', 'ИТОГО 10.00'])
    monkeypatch.setattr(app, 'engine', rotated)
    result = app.recognize(data.getvalue())
    assert len(result['items']) == 1
    assert any('поворот' in text for text in result['warnings'])
    assert rotated.calls == [False, True]
    ambiguous = Engine([], ['Покупка 10.00', 'ИТОГО 20.00'])
    monkeypatch.setattr(app, 'engine', ambiguous)
    assert app.recognize(data.getvalue())['items'] == []
