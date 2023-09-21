package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	// "golang.org/x/tools/go/packages"
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
	comparisonRepoBranchName := flag.String("comparison-branch-name", "", "Name of the source branch to compare the current branch to")

	// Parse input variables
	flag.Parse()

	var err error
	var testDir string

	if *testDirInput != "" {
		// Convert relative path to absolute path from /
		// Increase tolerance to different uses of the tool
		// calling from different starting directories
		testDir, err = filepath.Abs(*testDirInput)
		if err != nil {
			fatal("Bad test directory path: %s\n", err)
		}
	}

	if err := gitChangeGoPath(testDir, *baseImportName, *comparisonRepoBranchName); err != nil {
		fatal("Program Failed: %s", err)

	}
	// Program ran successfully, exit 0.
	fmt.Println("You do not need to run tests")
	os.Exit(0)

}

func gitChangeGoPath(testDir string, baseImportName string, comparisonRepoBranchName string) error {
	var filterTestsOnly bool = true
	initialFileImports, err := getDirectoryImports(testDir, filterTestsOnly, baseImportName)
	fmt.Println("Tests import these files:")
	fmt.Println(initialFileImports)
	if err != nil {
		return err
	}

	// Get full list of imports here, so far just using initial imports
	// Would like to walk the initial file imports and follow their imports.
	fullLocalImportList := initialFileImports

	// Get list of files changed from current branch to master branch
	gitDiffBytesOutput, err := exec.Command("git", "diff", "--name-only", comparisonRepoBranchName).Output()
	if err != nil {
		fatal("Issue Running git diff: %s", err)
	}
	// Parse list of files. 
	gitDiffFiles := strings.Split(string(gitDiffBytesOutput), "\n")
	var gitDiffFilesWithBaseAppended []string
	for _, gitDiffFile := range gitDiffFiles {
		// Path from root of project, where go.mod would exist
		gitDiffFilesWithBaseAppended = append(gitDiffFilesWithBaseAppended, filepath.Join(baseImportName, gitDiffFile))
	}
	fmt.Println("Files from git diff:")
	fmt.Println(gitDiffFilesWithBaseAppended)

	// Will need to update this from the git diff file, to the actual module/package name that the file modified.
	// For all the files changed in the branch, path from go.mod root
	for _, importname := range fullLocalImportList {
		for _, diffname := range gitDiffFilesWithBaseAppended {
		// If a file changed is imported by a test in the test directory specified
			if strings.Contains(diffname, importname) {
				// Exit, because bash: if [gitchangegopath] then exit 0; else run test
				// We want to return false to run tests.
				fmt.Printf("You should run tests!")
				os.Exit(1)
			}
		}
	}

	return nil
}

// Return a list of import names that all go files in the provided directory import.
// if filterTestOnly = true, only considers files that end in test.go
func getDirectoryImports(directory string, filterTestsOnly bool, baseImportFolder string) ([]string, error) {
	if len(directory) == 0 {
		fatal("No directory Provided\n")
		return []string{}, nil
	}
	
	// Collect files in provided directory
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
		// Filter go test files
		if filterTestsOnly && !validTestRegex.MatchString(fileName) {
			fmt.Println("skipping test %s", fileName)
			continue
		} else if !validGoFile.MatchString(fileName) { // Filter go files
			fmt.Println("skipping non-go file %s", fileName)
			continue
		}
		fullImportPath = filepath.Join(directory, file.Name()) // Works, filepath from root
		// Get list of packages that the go file imports
		inportNameIntermediate, err = getImportsFromFile(fullImportPath, baseImportFolder)
		if err != nil {
			fatal("Could not parse imports from %s: %s\n", fullImportPath, err)
		}
		// The import file.
		importNames = append(importNames, fileName)
		// All files after recursing down the import file
		importNames = append(importNames, inportNameIntermediate...)
	}
	
	return unique(importNames), nil
}

// Get list of packages that the go file imports
// Only considers local packages, by checking if the import path has the root folder that the go.mod file lives in
func getImportsFromFile(fullFilePath string, baseImportFolder string) ([]string, error) {
	// represents a set of source files and which we need for the parser
	fset := token.NewFileSet()
	// ParseFile can take a path to file
	// Parse the file into abstract syntax tree
	parsed, err := parser.ParseFile(fset, fullFilePath, nil, parser.ImportsOnly)
	if err != nil {
		return []string{}, err
	}

	var basePath string
	basePathIndex := strings.Index(fullFilePath, baseImportFolder)
	if basePathIndex != -1 {
        // Cut everything up to and including the cutString
        basePath = fullFilePath[:basePathIndex]
    } else {
        fmt.Printf("Can't find base %s in path %s", baseImportFolder, fullFilePath)
    }


	// Helpful info on ast: https://www.zupzup.org/go-ast-traversal/
	imports := parsed.Imports
	checkIfContainsBaseFolder := regexp.MustCompile(".*" + baseImportFolder + ".*")
	var importFileNames []string
	var directoryImports []string

	for _, individualImportStruct := range imports {
		// If the import has the base folder in it
		importString := strings.Trim(individualImportStruct.Path.Value, `"'`)
		if checkIfContainsBaseFolder.MatchString(importString) {
			fileInfo, err := os.Stat(basePath + importString)
			// Walk the directory and import all files it
			if fileInfo.IsDir() {
				directoryImports, err = getDirectoryImports(basePath + importString, false, baseImportFolder)
				if err != nil {
					fmt.Printf("That was weird: %s \n", err)
				}
				importFileNames = append(importFileNames, directoryImports...)
			} else {
				importFileNames = append(importFileNames, importString)
			}
		}
	}
	return importFileNames, nil
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
