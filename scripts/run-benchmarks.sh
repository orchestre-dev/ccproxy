#!/bin/bash

# Run performance benchmarks for CCProxy
# This script runs all benchmarks and saves results for comparison

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RESULTS_DIR="$PROJECT_ROOT/benchmark-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_FILE="$RESULTS_DIR/benchmark_$TIMESTAMP.txt"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}CCProxy Performance Benchmarks${NC}"
echo "================================"
echo "Timestamp: $(date)"
echo ""

# Create results directory if it doesn't exist
mkdir -p "$RESULTS_DIR"

# Function to run benchmarks for a package
run_benchmark() {
    local package=$1
    local name=$2
    
    echo -e "${YELLOW}Running benchmarks for: $name${NC}"
    echo "Package: $package"
    echo ""
    
    # Run benchmarks with various settings
    go test -bench=. -benchmem -benchtime=10s -cpu=1,2,4 "$package" | tee -a "$RESULT_FILE"
    echo "" | tee -a "$RESULT_FILE"
}

# Write header to result file
{
    echo "CCProxy Performance Benchmark Results"
    echo "===================================="
    echo "Date: $(date)"
    echo "Go Version: $(go version)"
    echo "OS: $(uname -s)"
    echo "Architecture: $(uname -m)"
    echo "CPU: $(sysctl -n machdep.cpu.brand_string 2>/dev/null || cat /proc/cpuinfo | grep "model name" | head -1 | cut -d: -f2)"
    echo ""
} > "$RESULT_FILE"

# Run benchmarks for each package
echo "Running benchmarks..."
echo ""

run_benchmark "ccproxy/internal/handlers" "HTTP Handlers"
run_benchmark "ccproxy/internal/converter" "Format Converter"
run_benchmark "ccproxy/internal/provider/common" "Provider Common"

echo -e "${GREEN}Benchmark complete!${NC}"
echo "Results saved to: $RESULT_FILE"
echo ""

# Generate summary statistics
echo "Summary Statistics:"
echo "==================="

# Extract and display key metrics
echo ""
echo "Top 5 Fastest Operations (ns/op):"
grep -E "Benchmark.*-[0-9]+" "$RESULT_FILE" | 
    awk '{print $1, $3}' | 
    sort -k2 -n | 
    head -5 | 
    while read name time; do
        printf "%-50s %s ns/op\n" "$name" "$time"
    done

echo ""
echo "Top 5 Most Memory Efficient (B/op):"
grep -E "B/op" "$RESULT_FILE" | 
    awk '{for(i=1;i<=NF;i++) if($i ~ /B\/op/) print $1, $(i-1)}' | 
    sort -k2 -n | 
    head -5 | 
    while read name bytes; do
        printf "%-50s %s B/op\n" "$name" "$bytes"
    done

echo ""
echo "Operations with Zero Allocations:"
grep -E "0 allocs/op" "$RESULT_FILE" | 
    awk '{print $1}' | 
    while read name; do
        echo "  - $name"
    done

# Optional: Compare with previous benchmark if exists
LATEST_PREVIOUS=$(ls -t "$RESULTS_DIR"/benchmark_*.txt 2>/dev/null | grep -v "$RESULT_FILE" | head -1)
if [ -n "$LATEST_PREVIOUS" ]; then
    echo ""
    echo "Comparison with previous run:"
    echo "============================="
    echo "Previous: $(basename "$LATEST_PREVIOUS")"
    echo ""
    
    # Simple comparison - you can enhance this with benchstat if available
    if command -v benchstat &> /dev/null; then
        benchstat "$LATEST_PREVIOUS" "$RESULT_FILE"
    else
        echo "Install benchstat for detailed comparison:"
        echo "  go install golang.org/x/perf/cmd/benchstat@latest"
    fi
fi

echo ""
echo "To run specific benchmarks:"
echo "  go test -bench=BenchmarkProxyMessages -benchmem ccproxy/internal/handlers"
echo ""
echo "To profile CPU usage:"
echo "  go test -bench=. -cpuprofile=cpu.prof ccproxy/internal/handlers"
echo "  go tool pprof cpu.prof"
echo ""