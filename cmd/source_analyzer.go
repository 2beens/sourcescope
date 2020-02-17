package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// TODO: make configurable
var mainSourcePathPrefix = "github.com/adjust/backend"
var excludedChangedFiles []string

func getChangedAndDependentSources() ([]string, []string) {
	// TODO: make configurable
	excludedChangedFiles = append(excludedChangedFiles, "test")

	// 1 - get changed files list
	changedFilesList := getChangedFiles()

	// 2 - remove file-name.go from files list
	changedPackages := getChangedPackages(changedFilesList)

	// 3 - iterate all source files of backend
	goFiles := getSourceGoFiles()

	// 4 - check import - is current changed package used/imported there ?
	dependentPackages := getDependentPackages(changedPackages, goFiles)

	return changedPackages, dependentPackages
}


func getChangedFiles() []string {
	var changedFilesListRaw []byte
	var err error

	cmd := exec.Command( "git", "diff", "--name-only", "master...")
	if changedFilesListRaw, err = cmd.Output(); err != nil {
		panic(err.Error())
	}

	changedFiles := strings.Split(string(changedFilesListRaw), "\n")
	var changedFilesFiltered []string
	for _, f := range changedFiles {
		if len(f) > 0 && strings.HasSuffix(f, ".go") {
			changedFilesFiltered = append(changedFilesFiltered, f)
		}
	}

	return changedFilesFiltered
}

func getPackageBySourceFile(sourceFile string) string {
	sourceFileParts := strings.Split(sourceFile, "/")
	var sb strings.Builder
	for _, p := range sourceFileParts {
		if strings.HasSuffix(p, ".go") {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("/")
		}
		sb.WriteString(p)
	}
	return sb.String()
}

func getChangedPackages(changedFiles []string) []string {
	changedPackages := make(map[string]struct{})
	for _, f := range changedFiles {
		changedPackages[getPackageBySourceFile(f)] = struct{}{}
	}

	var changedPackagesList []string
	for p := range changedPackages {
		//if isChangedFileExcluded(p) {
		//	continue
		//}
		changedPackagesList = append(changedPackagesList, p)
	}

	sort.Strings(changedPackagesList)
	return changedPackagesList
}

func isChangedFileExcluded(file string) bool {
	for _, exFile := range excludedChangedFiles {
		if exFile == file {
			return true
		}
	}
	return false
}

func getSourceGoFiles() []string {
	var goFiles []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(path,"vendor") {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		goFiles = append(goFiles, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return goFiles
}

func getDependentPackages(changedPackages []string, goSources []string) []string {
	var changedPackagesFiltered []string
	for _, p := range changedPackages {
		if !isChangedFileExcluded(p) {
			changedPackagesFiltered = append(changedPackagesFiltered, p)
		}
	}

	dependentPackages := make(map[string]struct{})
	fset := token.NewFileSet()
	for _, sourcePath := range goSources {
		node, err := parser.ParseFile(fset, sourcePath, nil, parser.ImportsOnly)
		if err != nil {
			panic(err)
		}

		if sourceImported := nodeContainsAnyImport(node, changedPackagesFiltered); sourceImported {
			sourcePackage := getPackageBySourceFile(sourcePath)
			dependentPackages[sourcePackage] = struct{}{}
		}
	}

	var dependentPackagesList []string
	for ds := range dependentPackages {
		dependentPackagesList = append(dependentPackagesList, ds)
	}

	sort.Strings(dependentPackagesList)
	return dependentPackagesList
}

func nodeContainsAnyImport(node *ast.File, changedPackages []string) bool {
	for _, i := range node.Imports {
		for _, changedPackage := range changedPackages {
			if strings.Contains(i.Path.Value, changedPackage) && strings.HasPrefix(i.Path.Value, `"` + mainSourcePathPrefix) {
				return true
			}
		}
	}
	return false
}
