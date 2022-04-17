package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func fatal(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

// List imported packages in directory
// Must be called from base path where files are imported from

func main() {

	testDirInput := flag.String("test-dir", "", "path to directory of tests")
	baseImportName := flag.String("base-folder-name", "", "base folder name which imports are called from")
	masterRepoBranchName := flag.String("master-repo-branch-name", "", "Name of the source branch to compare the current branch to")


	flag.Parse()

	var err error
	var testDir string

	if *testDirInput != "" {
		testDir, err = filepath.Abs(*testDirInput)
		if err != nil {
			fatal("Bad test directory path: %s\n", err)
		}
	}

	if err := gitChangeGoPath(testDir, *baseImportName, *masterRepoBranchName); err != nil {
		fatal("Program Failed: %s", err)
	}
	os.Exit(0)
}

func gitChangeGoPath(testDir string, baseImportName string, masterRepoBranchName string) error {
	var filterTestsOnly bool = true
	initialFileImports, err := getDirectoryImports(testDir, filterTestsOnly, baseImportName)
	fmt.Println(initialFileImports)
	if err != nil {
		return err
	}

	// Get full list of imports here, so far just using initial imports
	fullLocalImportList := initialFileImports


	gitDiffBytesOutput, err := exec.Command("git", "diff", "--name-only", masterRepoBranchName).Output()
	if err != nil {
		fatal("Issue Running git diff: %s", err)
	}
	gitDiffFiles := strings.Split(string(gitDiffBytesOutput), "\n")
	var gitDiffFilesWithBaseAppended []string
	for _, gitDiffFile := range gitDiffFiles {
		gitDiffFilesWithBaseAppended = append(gitDiffFilesWithBaseAppended, filepath.Join(baseImportName, gitDiffFile))
	}
	fmt.Println(fullLocalImportList)

	importMap := stringSliceToMap(fullLocalImportList)
	
	for _, diffFileName := range gitDiffFilesWithBaseAppended {
		if importMap[diffFileName]{
			os.Exit(1)
		}
	}

	return nil
}

func getDirectoryImports(directory string, filterTestsOnly bool, baseImportFolder string) ([]string, error) {
	if len(directory) == 0 {
		fatal("No directory Provided\n")
		return []string{}, nil
	}
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		fatal("Could not read files in test directory: %s\n", err)
	}
	var importNames []string
	var inportNameIntermediate []string
	var fullImportPath string
	var fileName string
	validTestRegex := regexp.MustCompile(`.*test\.go`)
	validGoFile := regexp.MustCompile(`.*\.go`)
	for _, file := range files {
		fileName = file.Name()
		//fmt.Println(fileName) Works!
		// Filter go test files
		if filterTestsOnly && !validTestRegex.MatchString(fileName) {
			fmt.Println("skipping")
			continue
		} else if !validGoFile.MatchString(fileName) { // Filter go files
			fmt.Println("skipping")
			continue
		}
		fullImportPath = filepath.Join(directory, file.Name()) // Works, filepath from root
		inportNameIntermediate, err = getImportsFromFile(fullImportPath, baseImportFolder)
		fmt.Println(inportNameIntermediate)  // Not working
		if err != nil {
			fatal("Could not parse imports from %s: %s\n", fullImportPath, err)
		}
		importNames = append(importNames, inportNameIntermediate...)
	}
	
	return unique(importNames), nil
}


// Do this by packages instead
func getImportsFromFile(fullFilePath string, baseImportFolder string) ([]string, error) {
	// represents a set of source files and which we need for the parser
	fset := token.NewFileSet()
	// ParseFile can take a path to file
	// Parse the file into abstract syntax tree
	parsed, err := parser.ParseFile(fset, fullFilePath, nil, parser.ImportsOnly)
	if err != nil {
		return []string{}, err
	}

	// Helpful info on ast: https://www.zupzup.org/go-ast-traversal/
	imports := parsed.Imports
	checkIfContainsBaseFolder := regexp.MustCompile(".*" + baseImportFolder + ".*")
	var importNames []string
	for _, individualImportStruct := range imports {
		// If the import has the base folder in it
		if checkIfContainsBaseFolder.MatchString(individualImportStruct.Path.Value) {
			importNames = append(importNames, individualImportStruct.Path.Value)
		}
	}
	return importNames, nil
}

// Returns unique items in a slice
func unique(slice []string) []string {
	// create a map with all the values as key
	uniqMap := make(map[string]struct{})
	for _, v := range slice {
		uniqMap[v] = struct{}{}
	}

	// turn the map keys into a slice
	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		uniqSlice = append(uniqSlice, v)
	}
	return uniqSlice
}

// Returns unique items in a slice
func stringSliceToMap(slice []string) map[string]bool {
	// create a map with all the values as key
	uniqMap := make(map[string]bool)
	for _, v := range slice {
		uniqMap[v] = true
	}
	return uniqMap
}