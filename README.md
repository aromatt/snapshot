# snapshot
## What
This is a CLI tool for text-based snapshot tests.

In the snapshot testing paradigm, a test case is an executable program which writes
meaningful output to stdout. This output is recorded in a "snapshot" file and should
be committed to source control alongside the test case. Later, if the test's output
changes, the `snapshot` tool will report it as a failure.

## Installation
```
$ go install github.com/aromatt/snapshot
```
## Usage
```
Usage: snapshot [flags] [test cases]
  -q    Suppress diff output
  -u    Update snapshots
```

## Example output
```
$ snapshot example/
example/fail.sh        FAILED
1c1
< bar
---
> foo

example/pass.sh        PASSED
1 failed, 1 passed
```
