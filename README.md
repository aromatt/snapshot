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
error.sh              FAILED    6.17ms
exit status 1
fail.sh               FAILED    4.13ms
1c1
< bar
---
> foo
pass.sh               PASSED    4.44ms
python_pass.py        PASSED    48.6ms
skip.sh               SKIPPED   0s
sleep.sh              PASSED    1.01s
2 failed, 3 passed, 1 skipped
```
