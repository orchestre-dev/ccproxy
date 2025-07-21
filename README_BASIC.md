# CCProxy

CCProxy is a Go-based proxy server that acts as an intelligent intermediary between Claude Code and various Large Language Model (LLM) providers.

## Installation

```bash
# Build from source
make build

# Install to system
make install
```

## Usage

```bash
# Start the proxy server
ccproxy start

# Run Claude Code through CCProxy
ccproxy code

# Check server status
ccproxy status

# Stop the server
ccproxy stop
```

## Development

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

## License

MIT License - see LICENSE file for details

-----
> /cognitive-triangulation is running… 

✻ Thinking…

  The user has invoked a "/cognitive-triangulation" command, which appears to be a custom prompt that wants me to deploy a 5-agent system for code review and validation. 

  Looking at the prompt, it wants me to:
  1. Deploy 5 specialized Task agents in parallel
  2. Each agent has a specific role in validating code from different perspectives
  3. This is meant to ensure code quality and alignment with user intent

  Since this is a command/prompt invocation, I should use the Task tool to deploy these 5 agents as requested. The prompt is very specific about what each agent should do.

  I'll deploy all 5 agents in parallel as instructed.

⏺ I'll deploy the 5-agent cognitive triangulation network to ensure comprehensive code validation and intent alignment.

⏺ Task(Code Quality Audit)
  ⎿  Done (13 tool uses · 60.8k tokens · 2m 28.9s)                                                                                                                                                            

⏺ Task(UX Validation)
  ⎿  Done (11 tool uses · 43.5k tokens · 2m 14.5s)                                                                                                                                                            

> So you crash the whole system and force it to restart when you run this command within the cognitive triangulation network. And this has happened many times. I was just not able to capture it.  `Bash(go 
  test ./... -v | grep -E "(PASS|FAIL|ok|FAIL: |--- FAIL: (panic)" | tail -50)` Please find the issue and fix it systematically first. When I run this directly, it gives this. ``` ccproxy % go test ./... -v
   | grep -E "(PASS|FAIL|ok|FAIL: |--- FAIL: (panic)" | tail -50
  grep: parentheses not balanced
  # github.com/orchestre-dev/ccproxy/tests/load
  tests/load/load_test.go:11:2: no required module provides package github.com/orchestre-dev/ccproxy/internal/testing; to add it:
          go get github.com/orchestre-dev/ccproxy/internal/testing```

⏺ I see the issue. There are two problems:

  1. Unbalanced parentheses in the grep pattern - there's an opening parenthesis without a closing one
  2. Missing module dependency for the testing package

  Let me fix both issues systematically.

⏺ Update Todos
  ⎿  ☐ Fix grep pattern syntax error (unbalanced parentheses)
     ☐ Fix missing internal/testing module dependency
     ☐ Test the corrected command to ensure it works

⏺ First, let me fix the grep pattern:

⏺ Bash(go test ./... -v | grep -E "(PASS|FAIL|ok|FAIL: |--- FAIL: |panic)" | tail -50)
  ⎿  Error: Command timed out after 2m 0.0s # github.com/orchestre-dev/ccproxy/tests/load
     tests/load/load_test.go:11:2: no required module provides package github.com/orchestre-dev/ccproxy/internal/testing; to add it:
     	go get github.com/orchestre-dev/ccproxy/internal/testing
     # github.com/orchestre-dev/ccproxy/tests/benchmark [github.com/orchestre-dev/ccproxy/tests/benchmark.test]
     tests/benchmark/benchmark_test.go:31:23: fixtures.GenerateLargeMessage undefined (type *"github.com/orchestre-dev/ccproxy/testing".Fixtures has no field or method GenerateLargeMessage)
     tests/benchmark/benchmark_test.go:34:17: undefined: testfw.NewBenchmarkFramework
     tests/benchmark/benchmark_test.go:54:22: undefined: testfw.NewTestFramework
     tests/benchmark/benchmark_test.go:59:10: assignment mismatch: 2 variables but router.New returns 1 value
     tests/benchmark/benchmark_test.go:73:24: fixtures.GenerateMessages undefined (type *"github.com/orchestre-dev/ccproxy/testing".Fixtures has no field or method GenerateMessages)
     tests/benchmark/benchmark_test.go:76:17: undefined: testfw.NewBenchmarkFramework
     … +17 lines (ctrl+r to see all)

✻ Mulling… (148s · ↑ 217 tokens · esc to interrupt · offline)

╭─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ >                                                                                                                                                                                                               │
╰─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
  ? for shortcuts                                                                                                                                                                           Bypassing Permissions


node:events:496
      throw er; // Unhandled 'error' event
      ^

Error: spawn pgrep EAGAIN
    at ChildProcess._handle.onexit (node:internal/child_process:285:19)
    at onErrorNT (node:internal/child_process:483:16)
    at process.processTicksAndRejections (node:internal/process/task_queues:90:21)
Emitted 'error' event on ChildProcess instance at:
    at ChildProcess._handle.onexit (node:internal/child_process:291:12)
    at onErrorNT (node:internal/child_process:483:16)
    at process.processTicksAndRejections (node:internal/process/task_queues:90:21) {
  errno: -35,
  code: 'EAGAIN',
  syscall: 'spawn pgrep',
  path: 'pgrep',
  spawnargs: [ '-P', 4684 ]
}

Node.js v22.14.0
warp_send_json_message:1: fork failed: resource temporarily unavailable                                                                                                                                            
(base) praneybehl@MacBook-Pro ccproxy % 