#!/bin/sh
# Runs the parallel evaluation benchmarks (evaluator/concurrency_bench_test.go) and prints a Markdown table.
#
#   ./bench.sh                 run and print the table; the raw output is saved to bench-latest.txt
#   ./bench.sh old.txt         also compare with a saved raw output (e.g. a copy of an earlier bench-latest.txt)
#
# Environment: CPUS (GOMAXPROCS list, default 1,2,4,8,16), COUNT (runs per cell, the median is used, default 5),
# BENCHTIME (default 1s).
set -e -u

cd "$(dirname "$0")"
baseline=${1:-}
raw=bench-latest.txt

go test -run '^$' -bench 'Parallel' -cpu "${CPUS:-1,2,4,8,16}" -count "${COUNT:-5}" \
    -benchtime "${BENCHTIME:-1s}" ./evaluator > "$raw"
grep '^cpu:' "$raw" || true

awk -v baseline="$baseline" '
function median(list,    n, a, i, j, t) {
    n = split(list, a, " ")
    for (i = 2; i <= n; i++) for (j = i; j > 1 && a[j-1] + 0 > a[j] + 0; j--) { t = a[j]; a[j] = a[j-1]; a[j-1] = t }
    return n % 2 ? a[(n + 1) / 2] : (a[n / 2] + a[n / 2 + 1]) / 2
}
/^Benchmark.*ns\/op/ {
    name = $1; procs = 1
    if (match(name, /-[0-9]+$/)) { procs = substr(name, RSTART + 1) + 0; name = substr(name, 1, RSTART - 1) }
    sub(/^BenchmarkParallel/, "", name)
    for (i = 2; i < NF; i++) if ($(i + 1) == "ns/op") ns = $i
    key = name SUBSEP procs
    if (FILENAME == baseline) old[key] = old[key] " " ns
    else {
        if (!(key in cur)) { if (!(name in seen)) { seen[name] = 1; names[++nn] = name }
                             if (!(procs in pseen)) { pseen[procs] = 1; plist[++np] = procs } }
        cur[key] = cur[key] " " ns
    }
}
END {
    printf "| Workload | Goroutines | ns/eval | M evals/s | Scaling vs 1 |"
    if (baseline != "") printf " Baseline ns/eval | Speedup vs baseline |"
    printf "\n|---|--:|--:|--:|--:|"
    if (baseline != "") printf "--:|--:|"
    printf "\n"
    for (i = 1; i <= nn; i++) {
        first = ""
        for (j = 1; j <= np; j++) {
            key = names[i] SUBSEP plist[j]
            if (!(key in cur)) continue
            ns = median(cur[key]); if (first == "") first = ns
            printf "| %s | %d | %.1f | %.2f | %.2fx |", names[i], plist[j], ns, 1000 / ns, first / ns
            if (baseline != "") {
                if (key in old) { o = median(old[key]); printf " %.1f | %.2fx |", o, o / ns }
                else printf " – | – |"
            }
            printf "\n"
        }
    }
}' $baseline "$raw"
