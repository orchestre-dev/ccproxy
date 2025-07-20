#!/usr/bin/env bash
# CCProxy comprehensive test runner

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
RESULTS_DIR="$PROJECT_ROOT/test-results"

# Test modes
MODE="${1:-all}"
VERBOSE="${VERBOSE:-false}"

# Print colored message
print_msg() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# Print info message
info() {
    print_msg "$GREEN" "ℹ️  $*"
}

# Print error message
error() {
    print_msg "$RED" "❌ $*"
}

# Print warning message
warn() {
    print_msg "$YELLOW" "⚠️  $*"
}

# Print step message
step() {
    print_msg "$BLUE" "▶️  $*"
}

# Create directories
setup_directories() {
    mkdir -p "$COVERAGE_DIR"
    mkdir -p "$RESULTS_DIR"
}

# Clean previous results
clean_results() {
    step "Cleaning previous test results..."
    rm -rf "$COVERAGE_DIR"/*
    rm -rf "$RESULTS_DIR"/*
}

# Run unit tests
run_unit_tests() {
    step "Running unit tests..."
    
    local test_args="-v -race -coverprofile=$COVERAGE_DIR/unit.out"
    if [ "$VERBOSE" = "true" ]; then
        test_args="$test_args -v"
    fi
    
    go test $test_args ./internal/... | tee "$RESULTS_DIR/unit-tests.log"
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        info "Unit tests passed ✅"
    else
        error "Unit tests failed"
        return $exit_code
    fi
}

# Run integration tests
run_integration_tests() {
    step "Running integration tests..."
    
    local test_args="-v -coverprofile=$COVERAGE_DIR/integration.out"
    if [ "$VERBOSE" = "true" ]; then
        test_args="$test_args -v"
    fi
    
    go test $test_args ./tests/integration/... | tee "$RESULTS_DIR/integration-tests.log"
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        info "Integration tests passed ✅"
    else
        error "Integration tests failed"
        return $exit_code
    fi
}

# Run benchmark tests
run_benchmark_tests() {
    step "Running benchmark tests..."
    
    local bench_args="-bench=. -benchmem -benchtime=10s"
    if [ "$VERBOSE" = "true" ]; then
        bench_args="$bench_args -v"
    fi
    
    go test $bench_args ./tests/benchmark/... | tee "$RESULTS_DIR/benchmark-tests.log"
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        info "Benchmark tests completed ✅"
        
        # Extract key metrics
        echo ""
        info "Benchmark Summary:"
        grep -E "Benchmark.*ns/op|Benchmark.*B/op" "$RESULTS_DIR/benchmark-tests.log" | tail -10
    else
        error "Benchmark tests failed"
        return $exit_code
    fi
}

# Run load tests
run_load_tests() {
    step "Running load tests..."
    
    local test_args="-v -timeout=10m"
    if [ "$VERBOSE" = "true" ]; then
        test_args="$test_args -v"
    fi
    
    go test $test_args ./tests/load/... | tee "$RESULTS_DIR/load-tests.log"
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        info "Load tests passed ✅"
        
        # Extract key metrics
        echo ""
        info "Load Test Summary:"
        grep -E "Requests/sec:|Success Rate:|Error Rate:" "$RESULTS_DIR/load-tests.log" | tail -10
    else
        error "Load tests failed"
        return $exit_code
    fi
}

# Run specific package tests
run_package_tests() {
    local package=$1
    step "Running tests for package: $package"
    
    local test_args="-v -race -coverprofile=$COVERAGE_DIR/package-$package.out"
    if [ "$VERBOSE" = "true" ]; then
        test_args="$test_args -v"
    fi
    
    go test $test_args ./$package/... | tee "$RESULTS_DIR/package-$package-tests.log"
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        info "Package tests passed ✅"
    else
        error "Package tests failed"
        return $exit_code
    fi
}

# Merge coverage files
merge_coverage() {
    step "Merging coverage reports..."
    
    # Find all coverage files
    local coverage_files=$(find "$COVERAGE_DIR" -name "*.out" -type f)
    
    if [ -z "$coverage_files" ]; then
        warn "No coverage files found"
        return
    fi
    
    # Merge coverage files
    echo "mode: atomic" > "$COVERAGE_DIR/coverage.out"
    for file in $coverage_files; do
        if [ "$file" != "$COVERAGE_DIR/coverage.out" ]; then
            tail -n +2 "$file" >> "$COVERAGE_DIR/coverage.out"
        fi
    done
    
    info "Coverage reports merged"
}

# Generate coverage report
generate_coverage_report() {
    step "Generating coverage report..."
    
    if [ ! -f "$COVERAGE_DIR/coverage.out" ]; then
        warn "No coverage data found"
        return
    fi
    
    # Generate HTML report
    go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"
    
    # Calculate total coverage
    local total_coverage=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | grep total | awk '{print $3}')
    
    info "Total coverage: $total_coverage"
    info "Coverage report: $COVERAGE_DIR/coverage.html"
    
    # Show coverage by package
    echo ""
    info "Coverage by package:"
    go tool cover -func="$COVERAGE_DIR/coverage.out" | grep -E "github.com/orchestre-dev/ccproxy" | sort -k3 -nr | head -20
}

# Run race detection tests
run_race_tests() {
    step "Running race detection tests..."
    
    go test -race -short ./... 2>&1 | tee "$RESULTS_DIR/race-tests.log"
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        info "No race conditions detected ✅"
    else
        error "Race conditions detected"
        return $exit_code
    fi
}

# Run short tests (for CI)
run_short_tests() {
    step "Running short tests..."
    
    go test -short -v ./... | tee "$RESULTS_DIR/short-tests.log"
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        info "Short tests passed ✅"
    else
        error "Short tests failed"
        return $exit_code
    fi
}

# Main execution
main() {
    info "CCProxy Test Runner"
    info "Mode: $MODE"
    echo ""
    
    # Setup
    setup_directories
    
    # Run tests based on mode
    case "$MODE" in
        all)
            clean_results
            run_unit_tests
            run_integration_tests
            run_benchmark_tests
            run_load_tests
            merge_coverage
            generate_coverage_report
            ;;
        unit)
            run_unit_tests
            generate_coverage_report
            ;;
        integration)
            run_integration_tests
            generate_coverage_report
            ;;
        benchmark)
            run_benchmark_tests
            ;;
        load)
            run_load_tests
            ;;
        race)
            run_race_tests
            ;;
        short)
            run_short_tests
            ;;
        coverage)
            clean_results
            run_unit_tests
            run_integration_tests
            merge_coverage
            generate_coverage_report
            ;;
        package)
            if [ -z "${2:-}" ]; then
                error "Package name required"
                echo "Usage: $0 package <package-name>"
                exit 1
            fi
            run_package_tests "$2"
            generate_coverage_report
            ;;
        *)
            error "Unknown mode: $MODE"
            echo "Available modes: all, unit, integration, benchmark, load, race, short, coverage, package"
            exit 1
            ;;
    esac
    
    echo ""
    step "Test run completed! 🎉"
    
    # Show summary
    if [ -d "$RESULTS_DIR" ] && [ "$(ls -A $RESULTS_DIR)" ]; then
        echo ""
        info "Test results saved in: $RESULTS_DIR/"
        ls -la "$RESULTS_DIR/"
    fi
}

# Handle script arguments
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    cat << EOF
Usage: $0 [mode] [options]

Modes:
  all         Run all tests (default)
  unit        Run unit tests only
  integration Run integration tests only
  benchmark   Run benchmark tests only
  load        Run load tests only
  race        Run race detection tests
  short       Run short tests (for CI)
  coverage    Run tests with coverage analysis
  package     Run tests for specific package

Options:
  VERBOSE=true  Enable verbose output

Examples:
  $0                    # Run all tests
  $0 unit               # Run unit tests only
  $0 package internal/router  # Run tests for router package
  VERBOSE=true $0 integration  # Run integration tests with verbose output
EOF
    exit 0
fi

# Run main function
main "$@"