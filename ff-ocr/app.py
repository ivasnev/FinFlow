"""Приватный stateless OCR-сервис. Очередью и правами владеет ff-split."""
import asyncio
import hmac
import io
import os
import signal
import time
from contextlib import asynccontextmanager

from fastapi import FastAPI, Header, HTTPException, Request
from fastapi.responses import PlainTextResponse
from PIL import Image, ImageOps, UnidentifiedImageError
from parser import parse_receipt

# Модели загружаются из официального хранилища Paddle, без опроса сторонних зеркал.
os.environ.setdefault('PADDLE_PDX_MODEL_SOURCE', 'BOS')
os.environ.setdefault('PADDLE_PDX_DISABLE_MODEL_SOURCE_CHECK', 'True')

MAX_BYTES = 10 * 1024 * 1024
MAX_PIXELS = 16_000_000
Image.MAX_IMAGE_PIXELS = MAX_PIXELS
TOKEN = os.environ.get('OCR_TOKEN', '')
engine = None
busy = asyncio.Semaphore(1)
running = set()
stats = {'completed': 0, 'failed': 0, 'busy': 0, 'seconds': 0.0}


def prepare_runtime():
    # Paddle replaces native signal handlers during import. Keep termination
    # owned by Uvicorn, including while models load in the background thread.
    handlers = {sig: signal.getsignal(sig) for sig in (signal.SIGTERM, signal.SIGINT)}
    previous_mask = signal.pthread_sigmask(signal.SIG_BLOCK, handlers) if hasattr(signal, 'pthread_sigmask') else None
    try:
        import paddle
        paddle.disable_signal_handler()
    finally:
        try:
            # CPython caches handlers: getsignal() alone cannot detect the
            # native sigaction changes, so reinstall them unconditionally.
            for sig, handler in handlers.items():
                signal.signal(sig, handler)
        finally:
            if previous_mask is not None:
                signal.pthread_sigmask(signal.SIG_SETMASK, previous_mask)


def load_engine():
    from paddleocr import PaddleOCR
    return PaddleOCR(text_detection_model_name='PP-OCRv5_mobile_det', text_recognition_model_name='eslav_PP-OCRv5_mobile_rec',
                     doc_orientation_classify_model_name='PP-LCNet_x1_0_doc_ori', use_doc_orientation_classify=True,
                     use_doc_unwarping=False, use_textline_orientation=False,
                     cpu_threads=int(os.environ.get('OCR_CPU_THREADS', '2')), enable_mkldnn=False)


@asynccontextmanager
async def lifespan(_app):
    global engine
    if len(TOKEN) < 32:
        raise RuntimeError('OCR_TOKEN must contain at least 32 characters')
    prepare_runtime()
    engine = await asyncio.to_thread(load_engine)
    try:
        yield
    finally:
        # Завершаем нативную работу до выгрузки модели при штатном shutdown.
        await asyncio.gather(*(asyncio.shield(task) for task in tuple(running)), return_exceptions=True)
        engine = None


app = FastAPI(lifespan=lifespan, docs_url=None, redoc_url=None, openapi_url=None)


def authorize(authorization):
    if not TOKEN or not hmac.compare_digest((authorization or '').encode(), ('Bearer ' + TOKEN).encode()):
        raise HTTPException(401, 'authentication required')


@app.get('/health/live')
def live():
    return {'status': 'ok'}


@app.get('/health/ready')
def ready():
    if engine is None:
        raise HTTPException(503, 'model loading')
    return {'status': 'ready', 'engine': 'PaddleOCR 3.3.2 / PP-OCRv5 eslav/mobile + orientation fallback'}


def recognize(data):
    import numpy as np
    try:
        with Image.open(io.BytesIO(data)) as source:
            if source.format not in {'JPEG', 'PNG', 'WEBP'} or source.width * source.height > MAX_PIXELS:
                raise ValueError('unsupported or oversized image')
            image = ImageOps.exif_transpose(source).convert('RGB')
            image.thumbnail((1600, 6000))
    except (UnidentifiedImageError, OSError, Image.DecompressionBombError, ValueError) as error:
        raise ValueError('invalid image') from error
    pixels = np.asarray(image)[:, :, ::-1].copy()
    started = time.monotonic()
    result = recognize_text(pixels, orient=False)
    # Общий классификатор иногда ошибается на фоне вокруг чека. Поэтому исправляем
    # ориентацию только после неудачного основного прохода, сохраняя его результат,
    # если повторный проход не дал позиций с согласованным напечатанным итогом.
    if (not result['items'] or result['total_cents'] is None) and time.monotonic() - started < 30:
        try:
            candidate = recognize_text(pixels, orient=True)
            if candidate['items'] and candidate['total_cents'] is not None and sum(item['amount_cents'] for item in candidate['items']) == candidate['total_cents']:
                candidate['warnings'].append('Для распознавания применён автоматический поворот. Сверьте позиции с фотографией.')
                return candidate
        except Exception:
            # Ошибка дополнительного прохода не уничтожает уже полученный черновик.
            pass
    return result


def recognize_text(pixels, orient):
    import numpy as np
    output = list(engine.predict(pixels, use_doc_orientation_classify=orient))
    lines = []
    for page in output:
        texts, scores = page.get('rec_texts', []), page.get('rec_scores', [])
        polygons = page.get('rec_polys', page.get('dt_polys', []))
        for i, text in enumerate(texts[:1200]):
            polygon = polygons[i] if i < len(polygons) else None
            y = float(np.mean(polygon[:, 1])) if polygon is not None else i * 30
            x = float(np.min(polygon[:, 0])) if polygon is not None else 0
            height = float(np.linalg.norm(polygon[3] - polygon[0])) if polygon is not None else 20
            lines.append({'text': str(text), 'confidence': float(scores[i]), 'x': x, 'y': y, 'height': height, 'polygon': polygon})
    # Выравниваем координаты по медианному наклону печатных строк: название и цена
    # на косом фото иначе попадут в соседние покупки.
    import math
    angles = []
    for line in lines:
        polygon = line['polygon']
        if polygon is not None:
            edge = polygon[1] - polygon[0]
            angle = math.atan2(float(edge[1]), float(edge[0]))
            if abs(angle) < math.pi / 4:
                angles.append(angle)
    angle = float(np.median(angles)) if angles else 0.0
    for line in lines:
        x, y = line['x'], line['y']
        polygon = line['polygon']
        if polygon is not None:
            x, y = float(np.mean(polygon[:, 0])), float(np.mean(polygon[:, 1]))
        line['x'], line['y'] = math.cos(angle)*x + math.sin(angle)*y, -math.sin(angle)*x + math.cos(angle)*y
    # OCR часто отдаёт название и цену отдельными блоками одной печатной строки.
    grouped = []
    for line in sorted(lines, key=lambda row: (row['y'], row['x'])):
        match = next((row for row in reversed(grouped[-4:]) if abs(row['y'] - line['y']) <= max(5, min(row['height'], line['height']) * 0.5)), None)
        if match is None:
            grouped.append({**line, 'parts': [line]})
        else:
            match['parts'].append(line)
            match['confidence'] = min(match['confidence'], line['confidence'])
    result_lines = [{'text': ' '.join(part['text'] for part in sorted(row['parts'], key=lambda part: part['x'])), 'confidence': row['confidence']} for row in grouped]
    return {**parse_receipt(result_lines), 'engine': 'PaddleOCR 3.3.2 / PP-OCRv5 eslav/mobile + orientation fallback'}


@app.post('/recognize')
async def scan(request: Request, authorization: str | None = Header(default=None)):
    authorize(authorization)
    if engine is None:
        raise HTTPException(503, 'model not ready')
    if busy.locked():
        stats['busy'] += 1
        raise HTTPException(429, 'OCR worker busy', headers={'Retry-After': '2'})
    slot = busy
    await slot.acquire()
    task = None
    try:
        data = bytearray()
        async for chunk in request.stream():
            data.extend(chunk)
            if len(data) > MAX_BYTES:
                raise HTTPException(413, 'image exceeds 10 MiB')
        started = time.monotonic()
        # Слот принадлежит нативной задаче, а не HTTP-запросу. Даже повторная
        # отмена запроса не допускает параллельный запуск той же модели.
        task = asyncio.create_task(asyncio.to_thread(recognize, bytes(data)))
        running.add(task)

        def completed(finished):
            running.discard(finished)
            slot.release()
            failed = finished.cancelled() or finished.exception() is not None
            stats['failed' if failed else 'completed'] += 1
            stats['seconds'] += time.monotonic() - started

        task.add_done_callback(completed)
        try:
            return await asyncio.shield(task)
        except ValueError:
            raise HTTPException(422, 'invalid image') from None
        except Exception:
            raise HTTPException(503, 'recognition failed') from None
    finally:
        if task is None:
            slot.release()


@app.get('/metrics', response_class=PlainTextResponse)
def metrics(authorization: str | None = Header(default=None)):
    authorize(authorization)
    return '\n'.join(f'finflow_ocr_{name} {value}' for name, value in stats.items()) + '\n'
