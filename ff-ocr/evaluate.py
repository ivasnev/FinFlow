"""Локальный прогон модели по заранее подготовленному корпусу, без скачивания фотографий."""
import argparse
import hashlib
import json
from pathlib import Path
import time

import app


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('manifest', type=Path, help='JSON-массив с file и необязательными source/license/sha256')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    root = args.manifest.resolve().parent
    entries = json.loads(args.manifest.read_text())
    if not isinstance(entries, list) or not 1 <= len(entries) <= 1000:
        parser.error('manifest must contain 1..1000 entries')
    app.engine = app.load_engine()
    results = []
    for entry in entries:
        image = (root / entry['file']).resolve()
        if not image.is_relative_to(root):
            parser.error('image must be inside the corpus directory')
        data = image.read_bytes()
        digest = hashlib.sha256(data).hexdigest()
        if entry.get('sha256') and digest != entry['sha256']:
            parser.error('image checksum mismatch: ' + image.name)
        started = time.perf_counter()
        result = app.recognize(data)
        elapsed = time.perf_counter() - started
        results.append({'file': entry['file'], 'source': entry.get('source'), 'license': entry.get('license'),
                        'sha256': digest, 'elapsed_seconds': elapsed, 'actual': result})
        print(f'{image.name}: {elapsed:.3f}s, items={len(result["items"])}, warnings={len(result["warnings"])}', flush=True)
    args.output.write_text(json.dumps(results, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
