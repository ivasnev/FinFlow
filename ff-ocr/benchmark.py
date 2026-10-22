"""Локальные нагрузочные проверки на новых синтетических событиях, без создания расходов."""
import argparse
import asyncio
import collections
import hashlib
import json
import math
import os
from pathlib import Path
import secrets
import statistics
import time
from urllib.parse import urlparse

import httpx


def local_url(value):
    parsed = urlparse(value)
    if parsed.scheme != 'http' or parsed.hostname not in {'127.0.0.1', 'localhost', '::1'} or parsed.username or parsed.query or parsed.fragment:
        raise argparse.ArgumentTypeError('benchmark accepts only a local HTTP endpoint')
    return value.rstrip('/')


def samples(path, limit):
    root = path.resolve().parent
    entries = json.loads(path.read_text())[:limit]
    images, hashes = [], set()
    for entry in entries:
        image = (root / entry['file']).resolve()
        if not image.is_relative_to(root):
            raise ValueError('image must remain inside manifest directory')
        with image.open("rb") as source:
            data = source.read(10 * 1024 * 1024 + 1)
        digest = hashlib.sha256(data).hexdigest()
        if digest != entry['sha256'] or digest in hashes or not 0 < len(data) <= 10 * 1024 * 1024:
            raise ValueError('samples must have unique verified hashes and fit the upload limit')
        hashes.add(digest)
        images.append((image.name, data, digest))
    if not images:
        raise ValueError('empty manifest')
    return images


def latency(rows):
    values = sorted(row['seconds'] * 1000 for row in rows)
    return {'p50_ms': round(statistics.median(values), 2), 'p95_ms': round(values[math.ceil(len(values) * .95) - 1], 2)} if values else {}


async def main(args):
    images = samples(args.manifest, args.limit)
    semaphore = asyncio.Semaphore(args.concurrency)
    sessions, jobs, rows = [], [], []
    output = {'mode': args.mode, 'samples': len(images), 'concurrency': args.concurrency, 'payload_bytes': sum(len(data) for _, data, _ in images)}
    async with httpx.AsyncClient(timeout=100, limits=httpx.Limits(max_connections=32)) as client:
        try:
            if args.mode == 'admission':
                # Каждый запуск регистрирует собственных пользователей и новые события:
                # дедупликация и уже готовые задания предыдущего запуска не помогают API.
                prefix = 'ocrbench_' + secrets.token_hex(5)
                for index in range(math.ceil(len(images) / 3)):
                    nickname = f'{prefix}_{index}'
                    response = await client.post(args.auth_url + '/auth/register', json={'nickname': nickname, 'email': nickname + '@example.com', 'password': secrets.token_urlsafe(24), 'name': 'OCR synthetic load'})
                    response.raise_for_status()
                    session = response.json()
                    headers = {'Authorization': 'Bearer ' + session['access_token']}
                    session['headers'] = headers
                    sessions.append(session)
                    response = await client.post(args.split_url + '/event', headers=headers, json={'name': f'OCR synthetic load {index}', 'description': 'Disposable local benchmark; no expenses created.'})
                    response.raise_for_status()
                    session['event_id'] = response.json()['id']
                for index, sample in enumerate(images):
                    jobs.append((sample, sessions[index % len(sessions)]))
            else:
                token = os.environ.get('OCR_TOKEN', '')
                if len(token) < 32:
                    raise ValueError('OCR_TOKEN is required for private OCR benchmark')
                jobs = [(sample, None) for sample in images]

            async def upload(job):
                (filename, data, digest), session = job
                async with semaphore:
                    start = time.perf_counter()
                    try:
                        if session:
                            response = await client.post(f'{args.split_url}/event/{session["event_id"]}/receipts', headers={**session['headers'], 'Content-Type': 'application/octet-stream'}, content=data)
                        else:
                            response = await client.post(args.ocr_url + '/recognize', headers={'Authorization': 'Bearer ' + token, 'Content-Type': 'application/octet-stream'}, content=data)
                    except httpx.HTTPError as error:
                        return {'file': filename, 'sha256': digest, 'status': 0, 'transport_error': type(error).__name__, 'seconds': round(time.perf_counter() - start, 6)}
                    row = {'file': filename, 'sha256': digest, 'status': response.status_code, 'seconds': round(time.perf_counter() - start, 6)}
                    if session and response.status_code == 202:
                        row['receipt_id'] = response.json()['id']
                        session.setdefault('receipts', []).append(row['receipt_id'])
                    if not session and response.status_code == 200:
                        row['items'] = len(response.json()['items'])
                    return row

            start = time.perf_counter()
            rows = await asyncio.gather(*(upload(job) for job in jobs))
            duration = time.perf_counter() - start
            accepted = [row for row in rows if row['status'] == (202 if args.mode == 'admission' else 200)]
            output.update(seconds=round(duration, 3), responses=dict(collections.Counter(row['status'] for row in rows if row['status'])), transport_errors=sum(row['status'] == 0 for row in rows), all_completed=len(accepted) == len(rows), accepted_per_second=round(len(accepted) / duration, 3), accepted_latency=latency(accepted), all_latency=latency(rows), results=rows)
            if args.mode == 'admission':
                ids = [row['receipt_id'] for row in accepted]
                if len(set(ids)) != len(ids):
                    raise AssertionError('benchmark hit receipt deduplication')
                output['new_receipt_ids'] = len(set(ids))
                output['distinct_image_hashes'] = len(set(row['sha256'] for row in accepted))
                print('Admission:', json.dumps({key: value for key, value in output.items() if key != 'results'}), flush=True)
                deadline = time.monotonic() + args.wait_seconds
                terminal = {}
                while time.monotonic() < deadline:
                    for session in sessions:
                        for receipt_id in session.get('receipts', []):
                            if receipt_id in terminal:
                                continue
                            response = await client.get(f'{args.split_url}/event/{session["event_id"]}/receipts/{receipt_id}', headers=session['headers'])
                            response.raise_for_status()
                            row = response.json()
                            if row['status'] in {'ready', 'failed', 'confirmed'}:
                                terminal[receipt_id] = row['status']
                    if len(terminal) == len(accepted):
                        break
                    await asyncio.sleep(2)
                output['terminal'] = dict(collections.Counter(terminal.values()))
                output['unfinished'] = len(accepted) - len(terminal)
                output['end_to_end_seconds'] = round(time.perf_counter() - start, 3)
                for session in sessions:
                    response = await client.get(f'{args.split_url}/event/{session["event_id"]}/transaction', headers=session['headers'])
                    response.raise_for_status()
                    if response.json().get('transactions'):
                        raise AssertionError('OCR benchmark unexpectedly created expenses')
                output['expenses_created'] = 0
        finally:
            # Отзываются только сессии, созданные этим запуском; токены не попадают в отчёт.
            revoked, revocation_errors = 0, []
            for session in sessions:
                try:
                    response = await client.post(args.auth_url + '/auth/logout', headers=session['headers'], json={'refresh_token': session['refresh_token']})
                    if response.is_success:
                        revoked += 1
                    else:
                        revocation_errors.append('HTTP ' + str(response.status_code))
                except httpx.HTTPError as error:
                    revocation_errors.append(type(error).__name__)
            output['sessions_revoked'] = revoked
            output['session_revocation_errors'] = revocation_errors
            args.output.write_text(json.dumps(output, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({key: value for key, value in output.items() if key != 'results'}, ensure_ascii=False), flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['admission', 'ocr'])
    parser.add_argument('manifest', type=Path)
    parser.add_argument('--auth-url', type=local_url, default='http://127.0.0.1:65170/api/v1')
    parser.add_argument('--split-url', type=local_url, default='http://127.0.0.1:65171/api/v1')
    parser.add_argument('--ocr-url', type=local_url, default='http://127.0.0.1:65172')
    parser.add_argument('--limit', type=int, default=24)
    parser.add_argument('--concurrency', type=int, default=4)
    parser.add_argument('--wait-seconds', type=int, default=900)
    parser.add_argument('--output', type=Path, required=True)
    options = parser.parse_args()
    if not 1 <= options.limit <= 100 or not 1 <= options.concurrency <= 32 or not 1 <= options.wait_seconds <= 1800:
        parser.error('invalid sample/concurrency/wait bound')
    asyncio.run(main(options))
