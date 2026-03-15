import argparse
import json
import os
import re
import pandas as pd

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))

# Маппинг: тег -> (bench_file, algo_name в бенчмарке)
ALGO_CONFIG = {
    "greedy": ("bench_greedy.json", "Random_Sweep_Greedy"),
    "dinic": ("bench_dinic.json", "Random_Sweep_Dinic"),
    "karp": ("bench_karp.json", "Random_Sweep_EdmondsKarp"),
    "pushrelabel": ("bench_pushrelabel.json", "Random_Sweep_PushRelabel"),
    "interactive_dinic": ("bench_interactive_dinic.json", "Random_Sweep_InteractiveMaxflow_Dinic"),
    "interactive_karp": ("bench_interactive_karp.json", "Random_Sweep_InteractiveMaxflow_EdmondsKarp"),
    "interactive_pushrelabel": ("bench_interactive_pushrelabel.json", "Random_Sweep_InteractiveMaxflow_PushRelabel"),
}

# Порядок в выводе Go: метрики по алфавиту; между полями могут быть лишние пробелы
bench_regex = re.compile(
    r'Benchmark(?P<algo>[^/]+)/nodes=(?P<nodes>\d+)_trans=(?P<trans>\d+)-\d+\s+'
    r'\d+\s+'
    r'(?P<ns>\d+)\s+ns/op\s+'
    r'\s*(?P<collapse_ns>[\d.]+)\s+collapse_ns_per_op\s+'
    r'\s*(?P<collapse>[\d.]+)\s+collapse_pct\s+'
    r'\s*(?P<input_after>[\d.]+)\s+input_after_collapse\s+'
    r'\s*(?P<optimize_ns>[\d.]+)\s+optimize_ns_per_op\s+'
    r'\s*(?P<reduce_success>[\d.]+)\s+reduce_success_pct\s+'
    r'\s*(?P<reduction_len>-?[\d.]+)\s+reduction_len_pct\s+'
    r'\s*(?P<valid>[\d.]+)\s+valid_pct\s+'
    r'\s*(?P<bytes>\d+)\s+B/op\s+'
    r'\s*(?P<allocs>\d+)\s+allocs/op'
)
# Когда Go переносит метрики на следующую строку: только числа и метрики (nodes/trans из Test)
bench_continuation_regex = re.compile(
    r'^\s*\d+\s+'
    r'(?P<ns>\d+)\s+ns/op\s+'
    r'\s*(?P<collapse_ns>[\d.]+)\s+collapse_ns_per_op\s+'
    r'\s*(?P<collapse>[\d.]+)\s+collapse_pct\s+'
    r'\s*(?P<input_after>[\d.]+)\s+input_after_collapse\s+'
    r'\s*(?P<optimize_ns>[\d.]+)\s+optimize_ns_per_op\s+'
    r'\s*(?P<reduce_success>[\d.]+)\s+reduce_success_pct\s+'
    r'\s*(?P<reduction_len>-?[\d.]+)\s+reduction_len_pct\s+'
    r'\s*(?P<valid>[\d.]+)\s+valid_pct\s+'
    r'\s*(?P<bytes>\d+)\s+B/op\s+'
    r'\s*(?P<allocs>\d+)\s+allocs/op'
)
test_name_regex = re.compile(r'/nodes=(?P<nodes>\d+)_trans=(?P<trans>\d+)')


def process_bench_file(bench_path: str, algo_display: str) -> pd.DataFrame:
    rows = []
    with open(bench_path, "r") as f:
        for line in f:
            try:
                event = json.loads(line)
            except Exception:
                continue

            if event.get("Action") != "output":
                continue

            output = event.get("Output", "").strip()
            match = bench_regex.search(output)
            if match:
                data = match.groupdict()
                nodes = int(data["nodes"])
                trans = int(data["trans"])
            else:
                # Вывод Go иногда разбит на две строки: имя в первой, метрики во второй
                match = bench_continuation_regex.search(output)
                if not match:
                    continue
                test_name = event.get("Test", "")
                nm = test_name_regex.search(test_name)
                if not nm:
                    continue
                data = match.groupdict()
                nodes = int(nm.group("nodes"))
                trans = int(nm.group("trans"))

            rows.append({
                "algorithm": algo_display,
                "nodes": nodes,
                "transactions": trans,
                "ns_per_op": int(data["ns"]),
                "collapse_pct": float(data["collapse"]),
                "input_after_collapse": float(data["input_after"]),
                "collapse_ns_per_op": float(data["collapse_ns"]),
                "optimize_ns_per_op": float(data["optimize_ns"]),
                "reduce_success_pct": float(data["reduce_success"]),
                "reduction_len_pct": float(data["reduction_len"]),
                "valid_pct": float(data["valid"]),
                "bytes_per_op": int(data["bytes"]),
                "allocs_per_op": int(data["allocs"]),
                "density": trans / (nodes * nodes),
            })

    df = pd.DataFrame(rows)
    if df.empty:
        return df
    return df.groupby(["algorithm", "nodes", "transactions"], as_index=False).mean()


def main():
    parser = argparse.ArgumentParser(description="Convert Go benchmark JSON to CSV datasets")
    parser.add_argument(
        "--algo",
        choices=["greedy", "dinic", "karp", "pushrelabel", "interactive_dinic", "interactive_karp", "interactive_pushrelabel"],
        help="Process only this algorithm",
    )
    parser.add_argument(
        "--all",
        action="store_true",
        help="Process all 3 algorithms (default)",
    )
    args = parser.parse_args()

    if args.algo:
        algos_to_run = [args.algo]
    else:
        algos_to_run = list(ALGO_CONFIG.keys())

    algo_display_map = {
        "greedy": "Greedy",
        "dinic": "Dinic",
        "karp": "EdmondsKarp",
        "pushrelabel": "PushRelabel",
        "interactive_dinic": "Interactive Dinic",
        "interactive_karp": "Interactive EdmondsKarp",
        "interactive_pushrelabel": "Interactive PushRelabel",
    }

    for tag in algos_to_run:
        bench_file, _ = ALGO_CONFIG[tag]
        bench_path = os.path.join(SCRIPT_DIR, bench_file)

        if not os.path.exists(bench_path):
            print(f"Warning: {bench_path} not found, skipping {tag}")
            continue

        df = process_bench_file(bench_path, algo_display_map[tag])
        if df.empty:
            print(f"No benchmark data in {bench_file}")
            continue

        out_csv = os.path.join(SCRIPT_DIR, f"dataset_{tag}.csv")
        df.to_csv(out_csv, index=False)
        print(f"Saved {out_csv} ({len(df)} rows)")
        print(df.head())


if __name__ == "__main__":
    main()