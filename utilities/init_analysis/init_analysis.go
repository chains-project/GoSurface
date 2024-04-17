package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

var (
	initFunctions []*initFunction
)

type initFunction struct {
	PackageName string
	FilePath    string
	LineNumber  int
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go /path/to/project")
		return
	}

	projectPath := os.Args[1]

	// Get project dependencies
	dependencies, err := getDependencies(projectPath)
	if err != nil {
		fmt.Printf("Error getting dependencies: %v\n", err)
		return
	}

	// Analyze project and dependencies for init functions
	analyzeProject(projectPath)
	for _, dep := range dependencies {
		analyzeProject(dep)
	}

	// Print results
	fmt.Println("Occurrences of init() function:")
	for _, fn := range initFunctions {
		fmt.Printf("- Package: %s, File: %s, Line: %d\n", fn.PackageName, fn.FilePath, fn.LineNumber)
	}

	// Count unique occurrences of init() functions
    	uniqueCount := countUniqueOccurrences(initFunctions)

    	// Print results
   	fmt.Println("Occurrences of init() function:")
    	for _, fn := range initFunctions {
        	fmt.Printf("- Package: %s, File: %s, Line: %d\n", fn.PackageName, fn.FilePath, fn.LineNumber)
    	}
    	fmt.Printf("Total unique occurrences of init() function: %d\n", uniqueCount)
}

func analyzeProject(path string) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing file %s: %v\n", path, err)
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		parseFile(path)
		return nil
	})
}

func parseFile(path string) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", path, err)
		return
	}

	for _, decl := range node.Decls {
    		fn, ok := decl.(*ast.FuncDecl)
    		if !ok || fn.Name.Name != "init" {
        		// Skip if it's not an init function
        		continue
    		}

		initFunctions = append(initFunctions, &initFunction{
			PackageName: packageName(path),
			FilePath:    path,
			LineNumber:  fset.Position(fn.Pos()).Line,
		})
	}
}

func packageName(filePath string) string {
	dir, _ := filepath.Split(filePath)
	goModPath := filepath.Join(dir, "go.mod")

	data, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}

	moduleName, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return ""
	}

	return moduleName.Module.Mod.Path
}

func getDependencies(projectPath string) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
	}

	pkgs, err := packages.Load(cfg, projectPath)
	if err != nil {
		return nil, err
	}

	var dependencies []string
	for _, pkg := range pkgs {
		dependencies = append(dependencies, pkg.PkgPath)
	}

	return dependencies, nil
}

func countUniqueOccurrences(initFunctions []*initFunction) int {
    uniqueOccurrences := make(map[string]struct{})
    for _, fn := range initFunctions {
        key := fmt.Sprintf("%s:%s:%d", fn.PackageName, fn.FilePath, fn.LineNumber)
        uniqueOccurrences[key] = struct{}{}
    }
    return len(uniqueOccurrences)
}
