# snapshot
## What
This is a CLI tool for text-based snapshot tests.

In the snapshot testing paradigm, a test case is an executable program which writes
meaningful output to stdout. This output is recorded in a "snapshot" file and should
be committed to version control alongside the test case. Later, if the test's output
changes, the `snapshot` tool will report it as a failure.

## Installation
```
$ go install github.com/aromatt/snapshot@latest
```
## Usage
```
Usage: snapshot [flags] [test cases]
  -q    Suppress diff output
  -u    Update snapshots
```
The `[test cases]` argument can be files or directories.

The tool will look for executable files among the arguments you provide, and will
execute all that have matching `.snapshot` files.

This repo has an [example](./example) directory full of test cases:
```
$ tree example/
example/
├── fail.sh
├── fail.sh.snapshot
├── pass.sh
├── pass.sh.snapshot
├── python_pass.py
├── python_pass.py.snapshot
└── skip.sh
```

## Output
The tool outputs a summary of the test run, including diffs for failed test cases
(unless `-q`), and returns an exit code of 1 if any tests failed.

Example output:
```
$ snapshot example/
fail.sh               FAILED    2.77ms
1c1
< bar
---
> foo

pass.sh               PASSED    1.98ms
python_pass.py        PASSED    30.4ms
skip.sh               SKIPPED   0s
sleep.sh              PASSED    1.01s
1 skipped, 1 failed, 3 passed
```
