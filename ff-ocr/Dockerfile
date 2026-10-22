FROM python:3.11-slim
WORKDIR /app
ENV PYTHONUNBUFFERED=1 PADDLE_PDX_CACHE_HOME=/models PADDLE_PDX_MODEL_SOURCE=BOS PADDLE_PDX_DISABLE_MODEL_SOURCE_CHECK=True OMP_NUM_THREADS=2
RUN apt-get update && apt-get install -y --no-install-recommends libgomp1 libgl1 libglib2.0-0 && rm -rf /var/lib/apt/lists/*
COPY ff-ocr/requirements.txt ./requirements.txt
RUN python -m pip install --no-cache-dir --upgrade pip==26.2.2 wheel==0.48.0 \
    && python -m pip install --no-cache-dir -r requirements.txt \
    && python -m pip check
COPY ff-ocr/app.py ff-ocr/parser.py ff-ocr/healthcheck.py ./
RUN useradd --system --uid 10001 ocr && mkdir /models && chown ocr /models
USER ocr
EXPOSE 8086
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8086", "--limit-concurrency", "16", "--timeout-keep-alive", "5"]
