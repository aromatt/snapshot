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

This repo has an [example](./example) directory full of test cases.

## Output
The tool outputs a summary of the test run, including diffs for failed test cases
(unless `-q`), and returns an exit code of 1 if any tests failed.

Example output:
```
$ snapshot example/
diff.sh               FAILED    4.36ms
1c1
< bar
---
> foo
error.sh              FAILED    2.88ms
exit status 1
pass.sh               PASSED    2.94ms
python_pass.py        PASSED    34.7ms
skip.sh               SKIPPED   0s
sleep.sh              PASSED    1.01s
stderr.sh             PASSED    5.52ms
2 failed, 4 passed, 1 skipped
```
