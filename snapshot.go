package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TestResult struct {
	Name  string
	Color int
}

var (
	Passed  = TestResult{"PASSED", 2}
	Skipped = TestResult{"SKIPPED", 15}
	Failed  = TestResult{"FAILED", 9}
	Updated = TestResult{"UPDATED", 6}
)

type SuiteResult struct {
	Results map[TestResult]int
}

func NewSuiteResult() *SuiteResult {
	return &SuiteResult{Results: make(map[TestResult]int)}
}

func (sr *SuiteResult) Add(result TestResult) {
	sr.Results[result]++
}

func (sr *SuiteResult) Summary() string {
	var parts []string
	for key, value := range sr.Results {
		parts = append(parts, fmt.Sprintf("%d %s", value, strings.ToLower(key.Name)))
	}
	return strings.Join(parts, ", ")
}

func (sr *SuiteResult) ExitCode() bool {
	return sr.Results[Failed] > 0
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

func parsePaths(paths []string) ([]string, error) {
	var out []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() && isExecutable(path) {
			out = append(out, path)
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

func colorString(str string, num int) string {
	return fmt.Sprintf("\033[38;5;%dm%s\033[0;0m", num, str)
}

func printStatus(path string, status TestResult, maxwidth int) {
	name := path
	padding := strings.Repeat(" ", maxwidth-len(name)+8)
	fmt.Printf("%s%s%s\n", name, padding, colorString(status.Name, status.Color))
}

func runTestCase(path string, update bool, quiet bool, maxwidth int) TestResult {
	snapPath := path + ".snapshot"
	status := Skipped
	var diff string
	var result *exec.Cmd

	if _, err := os.Stat(snapPath); err == nil || update {
		result = exec.Command("sh", path)
		output, err := result.CombinedOutput()
		if err == nil {
			tmpfile, err := os.CreateTemp("", "snapshot")
			if err != nil {
				fmt.Println("Error creating temp file:", err)
				return Failed
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write(output); err != nil {
				fmt.Println("Error writing to temp file:", err)
				return Failed
			}
			tmpfile.Close()

			if update {
				err := os.Rename(tmpfile.Name(), snapPath)
				if err != nil {
					fmt.Println("Error updating snapshot:", err)
					return Failed
				}
				status = Updated
			} else {
				diff, err = getDiff(snapPath, tmpfile.Name())
				if err != nil {
					status = Failed
				} else if diff == "" {
					status = Passed
				} else {
					status = Failed
				}
			}
		} else {
			status = Failed
			fmt.Println("Error:", err)
		}
	}
	printStatus(path, status, maxwidth)
	if status == Failed && diff != "" && !quiet {
		fmt.Println(diff)
	}
	return status
}

func runTestCases(paths []string, update bool, quiet bool) *SuiteResult {
	suiteResult := NewSuiteResult()
	var maxwidth int
	// maxwidth is the length of the longest path
	for _, path := range paths {
		if len(path) > maxwidth {
			maxwidth = len(path)
		}
	}
	for _, path := range paths {
		suiteResult.Add(runTestCase(path, update, quiet, maxwidth))
	}
	return suiteResult
}

func main() {
	updateFlag := flag.Bool("u", false, "Update snapshots")
	quietFlag := flag.Bool("q", false, "Suppress diff output")
	// add usage banner
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] [test cases]\n", os.Args[0])
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
	fmt.Println(suiteResult.Summary())
	if suiteResult.ExitCode() {
		os.Exit(1)
	}
}
