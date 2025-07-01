package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TestStatus struct {
	Name      string
	ColorCode string
}

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Orange = "\033[38;5;208m"
	Gray   = "\033[37m"
	Purple = "\033[35m"
)

var (
	Passed  = TestStatus{"PASSED", Green}
	Skipped = TestStatus{"SKIPPED", Yellow}
	Failed  = TestStatus{"FAILED", Red}
	Updated = TestStatus{"UPDATED", Orange}
)

type NonemptyDiffError struct {
	diff string
}

func (e NonemptyDiffError) Error() string {
	return e.diff
}

func NewNonemptyDiffError(diff string) NonemptyDiffError {
	return NonemptyDiffError{diff}
}

type SuiteResult struct {
	Results map[TestStatus]int
}

func NewSuiteResult() *SuiteResult {
	return &SuiteResult{Results: make(map[TestStatus]int)}
}

func (sr *SuiteResult) Add(result TestStatus) {
	sr.Results[result]++
}

func (sr *SuiteResult) Summary() string {
	var parts []string
	for key, value := range sr.Results {
		parts = append(parts, fmt.Sprintf("%d %s", value, strings.ToLower(key.Name)))
	}
	return strings.Join(parts, ", ")
}

func (sr *SuiteResult) ExitCode() int {
	if sr.Results[Failed] > 0 {
		return 1
	}
	return 0
}

// Returns true if `path` is executable by user, group and other.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

// Scans `paths` for executable files, recursing into directories. Returns a
// slice of absolute file paths.
func parsePaths(paths []string) ([]string, error) {
	var out []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() && isExecutable(path) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}
			out = append(out, absPath)
		} else if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil, err
			}
			subPaths := make([]string, 0, len(entries))
			for _, entry := range entries {
				fullPath := filepath.Join(path, entry.Name())
				subPaths = append(subPaths, fullPath)
			}
			parsedPaths, err := parsePaths(subPaths)
			if err != nil {
				return nil, err
			}
			for _, parsedPath := range parsedPaths {
				out = append(out, parsedPath)
			}
		}
	}
	return out, nil
}

func getDiff(path1, path2 string) (string, error) {
	cmd := exec.Command("diff", path1, path2)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func colorString(str string, colorCode string) string {
	return string(colorCode) + str + Reset
}

func formatDuration(d time.Duration) string {
	if d > time.Minute {
		return d.Round(time.Second).String()
	}
	for unit := time.Second; unit >= time.Microsecond; unit /= 10 {
		if d > unit {
			d = d.Round(unit / 100)
			break
		}
	}
	return d.String()
}

func printName(path string, maxwidth int) {
	name := filepath.Base(path)
	padding := strings.Repeat(" ", maxwidth-len(name)+8)
	fmt.Printf("%s%s", name, padding)
}

func printStatus(status TestStatus, duration time.Duration) {
	fmt.Printf("%s\t%s\n",
		colorString(status.Name, status.ColorCode),
		colorString(formatDuration(duration), Orange))
}

func printError(err error, quiet bool) {
	if nderr, ok := err.(NonemptyDiffError); ok {
		if !quiet {
			fmt.Fprintf(os.Stderr, nderr.diff)
		}
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func runTestCase(path string, update bool) (res TestStatus, dur time.Duration, err error) {
	res = Skipped
	snapPath := path + ".snapshot"
	var diff string
	var cmdResult *exec.Cmd
	var output []byte

	if _, err = os.Stat(snapPath); err == nil || update {
		start := time.Now()
		cmdResult = exec.Command(path)
		output, err = cmdResult.CombinedOutput()
		dur = time.Since(start)
		if err != nil {
			return Failed, dur, err
		}
		tmpfile, err := os.CreateTemp("", "snapshot")
		if err != nil {
			return Failed, dur, fmt.Errorf("creating temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write(output); err != nil {
			return Failed, dur, fmt.Errorf("writing to temp file: %v", err)
		}
		tmpfile.Close()

		if update {
			err := os.Rename(tmpfile.Name(), snapPath)
			if err != nil {
				return Failed, dur, fmt.Errorf("updating snapshot: %v", err)
			}
			res = Updated
		} else {
			diff, err = getDiff(snapPath, tmpfile.Name())
			if err != nil {
				if diff == "" {
					return Failed, dur, fmt.Errorf("diff failed: %v", err)
				}
				return Failed, dur, NewNonemptyDiffError(diff)
			}
			return Passed, dur, nil
		}
	}

	return res, dur, nil
}

func getMaxWidth(paths []string) int {
	var maxwidth int
	for _, path := range paths {
		name := filepath.Base(path)
		if len(name) > maxwidth {
			maxwidth = len(name)
		}
	}
	return maxwidth
}

func runTestCases(paths []string, update bool, quiet bool) *SuiteResult {
	suiteResult := NewSuiteResult()
	maxwidth := getMaxWidth(paths)
	for _, path := range paths {
		printName(path, maxwidth)
		status, duration, err := runTestCase(path, update)
		printStatus(status, duration)
		printError(err, quiet)
		suiteResult.Add(status)
	}
	return suiteResult
}

func main() {
	updateFlag := flag.Bool("u", false, "Update snapshots")
	quietFlag := flag.Bool("q", false, "Suppress diff output")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [flags] [test cases]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	cases := flag.Args()

	paths, err := parsePaths(cases)
	if err != nil {
		fmt.Println("Error parsing paths:", err)
		os.Exit(1)
	}

	if len(paths) == 0 {
		fmt.Println("No test cases found")
		os.Exit(1)
	}

	suiteResult := runTestCases(paths, *updateFlag, *quietFlag)
	fmt.Println(colorString(suiteResult.Summary(), Purple))

	os.Exit(suiteResult.ExitCode())
}
