#!/bin/bash
# Корень пакета optimizers (родитель tests/)
OPTIMIZERS_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$OPTIMIZERS_ROOT"
ANALITIC="tests/analitic"

RUN_GREEDY=0
RUN_DINIC=0
RUN_KARP=0
GENERATE_CSV=0
GEN_GREEDY=0
GEN_DINIC=0
GEN_KARP=0

for arg in "$@"; do
    case "$arg" in
        --bench-greedy)       RUN_GREEDY=1 ;;
        --bench-dinic)        RUN_DINIC=1 ;;
        --bench-karp)         RUN_KARP=1 ;;
        --bench-all)
            RUN_GREEDY=1
            RUN_DINIC=1
            RUN_KARP=1
            ;;
        --generate-csv)       GENERATE_CSV=1; GEN_GREEDY=1; GEN_DINIC=1; GEN_KARP=1 ;;
        --generate-csv-greedy) GENERATE_CSV=1; GEN_GREEDY=1 ;;
        --generate-csv-dinic)  GENERATE_CSV=1; GEN_DINIC=1 ;;
        --generate-csv-karp)   GENERATE_CSV=1; GEN_KARP=1 ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Benchmarks:"
            echo "  --bench-greedy       Run Greedy benchmarks only"
            echo "  --bench-dinic        Run Dinic benchmarks only"
            echo "  --bench-karp         Run EdmondsKarp benchmarks only"
            echo "  --bench-all          Run all benchmarks (default)"
            echo "Generate CSV from existing bench_*.json:"
            echo "  --generate-csv           Generate all dataset_*.csv"
            echo "  --generate-csv-greedy    Generate dataset_greedy.csv"
            echo "  --generate-csv-dinic     Generate dataset_dinic.csv"
            echo "  --generate-csv-karp      Generate dataset_karp.csv"
            exit 0
            ;;
    esac
done

if [[ $GENERATE_CSV -eq 1 ]]; then
    echo "Converting to datasets..."
    [[ $GEN_GREEDY -eq 1 ]] && python3 "$ANALITIC/bench_to_df.py" --algo greedy
    [[ $GEN_DINIC -eq 1 ]]  && python3 "$ANALITIC/bench_to_df.py" --algo dinic
    [[ $GEN_KARP -eq 1 ]]   && python3 "$ANALITIC/bench_to_df.py" --algo karp
    exit 0
fi

# Если ничего не указано — запускаем всё
if [[ $RUN_GREEDY -eq 0 && $RUN_DINIC -eq 0 && $RUN_KARP -eq 0 ]]; then
    RUN_GREEDY=1
    RUN_DINIC=1
    RUN_KARP=1
fi

ALGOS=""
[[ $RUN_GREEDY -eq 1 ]] && { echo "Running Greedy benchmarks..."; go test -bench '^BenchmarkRandom_Sweep_Greedy$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_greedy.json" 2>&1; ALGOS="$ALGOS greedy"; }
[[ $RUN_DINIC -eq 1 ]]  && { echo "Running Dinic benchmarks..."; go test -bench '^BenchmarkRandom_Sweep_Dinic$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_dinic.json" 2>&1; ALGOS="$ALGOS dinic"; }
[[ $RUN_KARP -eq 1 ]]   && { echo "Running EdmondsKarp benchmarks..."; go test -bench '^BenchmarkRandom_Sweep_EdmondsKarp$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_karp.json" 2>&1; ALGOS="$ALGOS karp"; }

echo "Converting to datasets..."
for a in $ALGOS; do python3 "$ANALITIC/bench_to_df.py" --algo "$a"; done

