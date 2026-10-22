"""Независимые проверки выбора и ресурса дополнительного ориентационного прохода."""
import asyncio
import io
import threading
from types import SimpleNamespace
import app
import httpx
import pytest
from PIL import Image


def photo():
    data = io.BytesIO()
    Image.new('RGB', (10, 10)).save(data, format='PNG')
    return data.getvalue()


@pytest.mark.parametrize('mode', ['slow', 'error', 'no_total', 'low_confidence', 'known_mismatch'])
def test_fallback_uses_only_explicit_complete_candidate(monkeypatch, mode):
    calls = []
    first = ['Первый товар 10.00', 'ИТОГО 20.00'] if mode == 'known_mismatch' else ['Первый товар 10.00']
    def predict(_pixels, use_doc_orientation_classify):
        calls.append(use_doc_orientation_classify)
        if not use_doc_orientation_classify:
            texts = first
        elif mode == 'error':
            raise RuntimeError('synthetic model failure')
        else:
            texts = ['Другой товар 20.00']
            if mode != 'no_total':
                texts.append('ИТОГО 20.00')
        return [{'rec_texts': texts, 'rec_scores': [.2 if mode == 'low_confidence' else .99] * len(texts)}]
    class Engine:
        pass
    engine = Engine()
    engine.predict = predict
    monkeypatch.setattr(app, 'engine', engine)
    if mode == 'slow':
        moments = iter([10., 40.])
        monkeypatch.setattr(app, 'time', SimpleNamespace(monotonic=lambda: next(moments)))
    result = app.recognize(photo())
    if mode == 'low_confidence':
        assert result['items'][0]['amount_cents'] == 2000
        assert result['items'][0]['confidence'] == .2
        assert any('низкой уверенностью' in w for w in result['warnings'])
        assert any('поворот' in w for w in result['warnings'])
    else:
        assert result['items'][0]['name'] == 'Первый товар'
        assert result['items'][0]['amount_cents'] == 1000
        assert result['warnings']
    assert calls == ([False] if mode in {'slow', 'known_mismatch'} else [False, True])


def test_cancel_during_orientation_keeps_the_same_native_slot(monkeypatch):
    started, release = threading.Event(), threading.Event()
    calls = []
    class Engine:
        def predict(self, _pixels, use_doc_orientation_classify):
            calls.append(use_doc_orientation_classify)
            if not use_doc_orientation_classify:
                return []
            started.set()
            release.wait(5)
            return [{'rec_texts': ['Покупка 10.00', 'ИТОГО 10.00'], 'rec_scores': [.99, .99]}]
    token = 'synthetic-local-ocr-token-32-chars'
    monkeypatch.setattr(app, 'TOKEN', token)
    monkeypatch.setattr(app, 'engine', Engine())
    monkeypatch.setattr(app, 'busy', asyncio.Semaphore(1))
    monkeypatch.setattr(app, 'running', set())
    async def scenario():
        async with httpx.AsyncClient(transport=httpx.ASGITransport(app=app.app), base_url='http://test') as client:
            headers = {'Authorization': 'Bearer '+token}
            request = asyncio.create_task(client.post('/recognize', content=photo(), headers=headers))
            try:
                assert await asyncio.to_thread(started.wait, 2)
                request.cancel()
                await asyncio.sleep(.01)
                request.cancel()
                await asyncio.gather(request, return_exceptions=True)
                assert (await client.post('/recognize', content=photo(), headers=headers)).status_code == 429
                assert calls == [False, True]
            finally:
                release.set()
                await asyncio.gather(*tuple(app.running), return_exceptions=True)
        assert not app.busy.locked()
    asyncio.run(scenario())
