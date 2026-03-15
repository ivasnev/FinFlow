# Optimizers Benchmarks

Бенчмарки и генерация датасетов для анализа оптимизаторов (Greedy, Dinic, EdmondsKarp, PushRelabel, InteractiveMaxflow+Dinic, InteractiveMaxflow+EdmondsKarp, InteractiveMaxflow+PushRelabel).

Запуск из папки `analitic`: `bash run.sh` (скрипт переходит в корень пакета `optimizers` и пишет результаты в `tests/analitic/`).

Для конвертации бенчмарков в CSV и ноутбука нужен Python 3 и pandas:

```bash
pip install -r tests/analitic/requirements.txt
```

или `pip3 install pandas`.

## Бенчмарки

Запуск всех бенчмарков (по умолчанию):

```bash
cd tests/analitic && bash run.sh
```

Только Greedy:

```bash
bash run.sh --bench-greedy
```

Только Dinic:

```bash
bash run.sh --bench-dinic
```

Только EdmondsKarp:

```bash
bash run.sh --bench-karp
```

Только PushRelabel (Черкасский):

```bash
bash run.sh --bench-pushrelabel
```

InteractiveMaxflow с разными maxflow-солверами:

```bash
bash run.sh --bench-interactive-dinic
bash run.sh --bench-interactive-karp
bash run.sh --bench-interactive-pushrelabel
```

Явно все:

```bash
bash run.sh --bench-all
```

## Генерация CSV

Сгенерировать все датасеты из существующих `bench_*.json`:

```bash
bash run.sh --generate-csv
```

Только Greedy:

```bash
bash run.sh --generate-csv-greedy
```

Только Dinic:

```bash
bash run.sh --generate-csv-dinic
```

Только EdmondsKarp:

```bash
bash run.sh --generate-csv-karp
```

Только PushRelabel:

```bash
bash run.sh --generate-csv-pushrelabel
```

Только Interactive датасеты:

```bash
bash run.sh --generate-csv-interactive-dinic
bash run.sh --generate-csv-interactive-karp
bash run.sh --generate-csv-interactive-pushrelabel
```

## Python напрямую

```bash
python3 bench_to_df.py --all
```

```bash
python3 bench_to_df.py --algo greedy
```

```bash
python3 bench_to_df.py --algo dinic
```

```bash
python3 bench_to_df.py --algo karp
```

```bash
python3 bench_to_df.py --algo pushrelabel
```

```bash
python3 bench_to_df.py --algo interactive_dinic
python3 bench_to_df.py --algo interactive_karp
python3 bench_to_df.py --algo interactive_pushrelabel
```

## Справка

```bash
bash run.sh --help
```
