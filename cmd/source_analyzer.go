package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type SourceAnalyzer struct {
	importPathPrefix     string
	rootDir              string
	excludedChangedFiles []string
	excludedRootFolders  []string
}

func NewSourceAnalyzer(rootDir, sourcePathPrefix string) *SourceAnalyzer {
	sourceAnalyzer := &SourceAnalyzer{
		importPathPrefix: sourcePathPrefix,
		rootDir:          rootDir,
	}

	sourceAnalyzer.excludedChangedFiles = []string{"test"}
	sourceAnalyzer.excludedRootFolders = []string{"vendor", "third_party", "sql", "bin"}

	return sourceAnalyzer
}

func (sa *SourceAnalyzer) GetChangedAndDependentSources() ([]string, []string) {
	// 1 - get changed files list
	changedFilesList := sa.getChangedFiles()

	// 2 - remove file-name.go from files list
	changedPackages := sa.getChangedPackages(changedFilesList)

	// 3 - iterate all source files of backend
	goFiles := sa.getSourceGoFiles()

	// 4 - check import - is current changed package used/imported there ?
	dependentPackages := sa.getDependentPackages(changedPackages, goFiles)

	return changedPackages, dependentPackages
}

func (sa *SourceAnalyzer) getChangedFiles() []string {
	var changedFilesListRaw []byte
	var err error

	cmd := exec.Command("git", "diff", "--name-only", "master...")
	cmd.Dir = rootDir
	if changedFilesListRaw, err = cmd.Output(); err != nil {
		log.Fatalf("error, check if valid git repository. %v", err)
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

func (sa *SourceAnalyzer) getPackageBySourceFile(sourceFile string) string {
	sourceFile = strings.TrimPrefix(sourceFile, sa.rootDir)
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

func (sa *SourceAnalyzer) getChangedPackages(changedFiles []string) []string {
	changedPackages := make(map[string]struct{})
	for _, f := range changedFiles {
		changedPackages[sa.getPackageBySourceFile(f)] = struct{}{}
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

func (sa *SourceAnalyzer) getSourceGoFiles() []string {
	var goFiles []string
	err := filepath.Walk(sa.rootDir, func(path string, info os.FileInfo, err error) error {
		if sa.isRootFolderExcluded(path) {
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

func (sa *SourceAnalyzer) isChangedFileExcluded(file string) bool {
	for _, exFile := range sa.excludedChangedFiles {
		if exFile == file {
			return true
		}
	}
	return false
}

func (sa *SourceAnalyzer) isRootFolderExcluded(rootFolderPath string) bool {
	for _, f := range sa.excludedRootFolders {
		if strings.HasPrefix(rootFolderPath, sa.rootDir+"/"+f) {
			return true
		}
	}
	return false
}

func (sa *SourceAnalyzer) getDependentPackages(changedPackages []string, goSources []string) []string {
	var changedPackagesFiltered []string
	for _, p := range changedPackages {
		if !sa.isChangedFileExcluded(p) {
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

		if sourceImported := sa.nodeContainsAnyImport(node, changedPackagesFiltered); sourceImported {
			sourcePackage := sa.getPackageBySourceFile(sourcePath)
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

func (sa *SourceAnalyzer) GetRootFolders(packages []string) []string {
	rootFolders := make(map[string]struct{})
	for _, p := range packages {
		rootFolders[strings.Split(p, "/")[0]] = struct{}{}
	}
	var rootFoldersList []string
	for rf := range rootFolders {
		rootFoldersList = append(rootFoldersList, rf)
	}
	sort.Strings(rootFoldersList)
	return rootFoldersList
}

func (sa *SourceAnalyzer) nodeContainsAnyImport(node *ast.File, changedPackages []string) bool {
	for _, i := range node.Imports {
		for _, changedPackage := range changedPackages {
			if strings.Contains(i.Path.Value, changedPackage) && strings.HasPrefix(i.Path.Value, `"`+sa.importPathPrefix) {
				return true
			}
		}
	}
	return false
}
