# Optimizers Benchmarks

Бенчмарки и генерация датасетов для анализа оптимизаторов (Greedy, Dinic, EdmondsKarp).

Запуск из папки `analitic`: `bash run.sh` (скрипт переходит в корень пакета `optimizers` и пишет результаты в `tests/analitic/`).

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

Явно все три:

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

## Справка

```bash
bash run.sh --help
```
