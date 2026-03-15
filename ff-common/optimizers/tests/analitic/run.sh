#!/bin/bash
# Корень пакета optimizers (родитель tests/)
OPTIMIZERS_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$OPTIMIZERS_ROOT"
ANALITIC="tests/analitic"

RUN_GREEDY=0
RUN_DINIC=0
RUN_KARP=0
RUN_PUSHRELABEL=0
RUN_INTERACTIVE_DINIC=0
RUN_INTERACTIVE_KARP=0
RUN_INTERACTIVE_PUSHRELABEL=0
GENERATE_CSV=0
GEN_GREEDY=0
GEN_DINIC=0
GEN_KARP=0
GEN_PUSHRELABEL=0
GEN_INTERACTIVE_DINIC=0
GEN_INTERACTIVE_KARP=0
GEN_INTERACTIVE_PUSHRELABEL=0

for arg in "$@"; do
    case "$arg" in
        --bench-greedy)              RUN_GREEDY=1 ;;
        --bench-dinic)               RUN_DINIC=1 ;;
        --bench-karp)                RUN_KARP=1 ;;
        --bench-pushrelabel)         RUN_PUSHRELABEL=1 ;;
        --bench-interactive-dinic)   RUN_INTERACTIVE_DINIC=1 ;;
        --bench-interactive-karp)    RUN_INTERACTIVE_KARP=1 ;;
        --bench-interactive-pushrelabel) RUN_INTERACTIVE_PUSHRELABEL=1 ;;
        --bench-all)
            RUN_GREEDY=1
            RUN_DINIC=1
            RUN_KARP=1
            RUN_PUSHRELABEL=1
            RUN_INTERACTIVE_DINIC=1
            RUN_INTERACTIVE_KARP=1
            RUN_INTERACTIVE_PUSHRELABEL=1
            ;;
        --generate-csv)              GENERATE_CSV=1; GEN_GREEDY=1; GEN_DINIC=1; GEN_KARP=1; GEN_PUSHRELABEL=1; GEN_INTERACTIVE_DINIC=1; GEN_INTERACTIVE_KARP=1; GEN_INTERACTIVE_PUSHRELABEL=1 ;;
        --generate-csv-greedy)       GENERATE_CSV=1; GEN_GREEDY=1 ;;
        --generate-csv-dinic)        GENERATE_CSV=1; GEN_DINIC=1 ;;
        --generate-csv-karp)         GENERATE_CSV=1; GEN_KARP=1 ;;
        --generate-csv-pushrelabel)  GENERATE_CSV=1; GEN_PUSHRELABEL=1 ;;
        --generate-csv-interactive-dinic)    GENERATE_CSV=1; GEN_INTERACTIVE_DINIC=1 ;;
        --generate-csv-interactive-karp)    GENERATE_CSV=1; GEN_INTERACTIVE_KARP=1 ;;
        --generate-csv-interactive-pushrelabel) GENERATE_CSV=1; GEN_INTERACTIVE_PUSHRELABEL=1 ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Benchmarks:"
            echo "  --bench-greedy               Run Greedy benchmarks only"
            echo "  --bench-dinic                Run Dinic benchmarks only"
            echo "  --bench-karp                 Run EdmondsKarp benchmarks only"
            echo "  --bench-pushrelabel          Run PushRelabel (Cherkassky) benchmarks only"
            echo "  --bench-interactive-dinic    Run InteractiveMaxflow+Dinic benchmarks only"
            echo "  --bench-interactive-karp    Run InteractiveMaxflow+EdmondsKarp benchmarks only"
            echo "  --bench-interactive-pushrelabel Run InteractiveMaxflow+PushRelabel benchmarks only"
            echo "  --bench-all                  Run all benchmarks (default)"
            echo "Generate CSV from existing bench_*.json:"
            echo "  --generate-csv                       Generate all dataset_*.csv"
            echo "  --generate-csv-greedy                Generate dataset_greedy.csv"
            echo "  --generate-csv-dinic                 Generate dataset_dinic.csv"
            echo "  --generate-csv-karp                  Generate dataset_karp.csv"
            echo "  --generate-csv-pushrelabel           Generate dataset_pushrelabel.csv"
            echo "  --generate-csv-interactive-dinic     Generate dataset_interactive_dinic.csv"
            echo "  --generate-csv-interactive-karp      Generate dataset_interactive_karp.csv"
            echo "  --generate-csv-interactive-pushrelabel Generate dataset_interactive_pushrelabel.csv"
            exit 0
            ;;
    esac
done

if [[ $GENERATE_CSV -eq 1 ]]; then
    echo "Converting to datasets..."
    [[ $GEN_GREEDY -eq 1 ]]               && python3 "$ANALITIC/bench_to_df.py" --algo greedy
    [[ $GEN_DINIC -eq 1 ]]                && python3 "$ANALITIC/bench_to_df.py" --algo dinic
    [[ $GEN_KARP -eq 1 ]]                 && python3 "$ANALITIC/bench_to_df.py" --algo karp
    [[ $GEN_PUSHRELABEL -eq 1 ]]          && python3 "$ANALITIC/bench_to_df.py" --algo pushrelabel
    [[ $GEN_INTERACTIVE_DINIC -eq 1 ]]    && python3 "$ANALITIC/bench_to_df.py" --algo interactive_dinic
    [[ $GEN_INTERACTIVE_KARP -eq 1 ]]     && python3 "$ANALITIC/bench_to_df.py" --algo interactive_karp
    [[ $GEN_INTERACTIVE_PUSHRELABEL -eq 1 ]] && python3 "$ANALITIC/bench_to_df.py" --algo interactive_pushrelabel
    exit 0
fi

# Если ничего не указано — запускаем всё
if [[ $RUN_GREEDY -eq 0 && $RUN_DINIC -eq 0 && $RUN_KARP -eq 0 && $RUN_PUSHRELABEL -eq 0 && $RUN_INTERACTIVE_DINIC -eq 0 && $RUN_INTERACTIVE_KARP -eq 0 && $RUN_INTERACTIVE_PUSHRELABEL -eq 0 ]]; then
    RUN_GREEDY=1
    RUN_DINIC=1
    RUN_KARP=1
    RUN_PUSHRELABEL=1
    RUN_INTERACTIVE_DINIC=1
    RUN_INTERACTIVE_KARP=1
    RUN_INTERACTIVE_PUSHRELABEL=1
fi

# Список алгоритмов для последующей конвертации в CSV
ALGOS=""
[[ $RUN_GREEDY -eq 1 ]]               && ALGOS="$ALGOS greedy"
[[ $RUN_DINIC -eq 1 ]]                && ALGOS="$ALGOS dinic"
[[ $RUN_KARP -eq 1 ]]                 && ALGOS="$ALGOS karp"
[[ $RUN_PUSHRELABEL -eq 1 ]]          && ALGOS="$ALGOS pushrelabel"
[[ $RUN_INTERACTIVE_DINIC -eq 1 ]]    && ALGOS="$ALGOS interactive_dinic"
[[ $RUN_INTERACTIVE_KARP -eq 1 ]]     && ALGOS="$ALGOS interactive_karp"
[[ $RUN_INTERACTIVE_PUSHRELABEL -eq 1 ]] && ALGOS="$ALGOS interactive_pushrelabel"

# Запуск бенчмарков параллельно (каждый пишет в свой JSON)
echo "Running benchmarks in parallel..."
[[ $RUN_GREEDY -eq 1 ]]               && go test -bench '^BenchmarkRandom_Sweep_Greedy$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_greedy.json" 2>&1 &
[[ $RUN_DINIC -eq 1 ]]                && go test -bench '^BenchmarkRandom_Sweep_Dinic$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_dinic.json" 2>&1 &
[[ $RUN_KARP -eq 1 ]]                 && go test -bench '^BenchmarkRandom_Sweep_EdmondsKarp$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_karp.json" 2>&1 &
[[ $RUN_PUSHRELABEL -eq 1 ]]          && go test -bench '^BenchmarkRandom_Sweep_PushRelabel$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_pushrelabel.json" 2>&1 &
[[ $RUN_INTERACTIVE_DINIC -eq 1 ]]    && go test -bench '^BenchmarkRandom_Sweep_InteractiveMaxflow_Dinic$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_interactive_dinic.json" 2>&1 &
[[ $RUN_INTERACTIVE_KARP -eq 1 ]]     && go test -bench '^BenchmarkRandom_Sweep_InteractiveMaxflow_EdmondsKarp$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_interactive_karp.json" 2>&1 &
[[ $RUN_INTERACTIVE_PUSHRELABEL -eq 1 ]] && go test -bench '^BenchmarkRandom_Sweep_InteractiveMaxflow_PushRelabel$' -benchmem -run=^$ -json ./tests > "$ANALITIC/bench_interactive_pushrelabel.json" 2>&1 &

wait
echo "All benchmarks finished."

echo "Converting to datasets..."
for a in $ALGOS; do
    python3 "$ANALITIC/bench_to_df.py" --algo "$a"
done

