package internal

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CheckGoInstalled() error {
	path, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf(
			"go toolchain not found in PATH — please install Go from https://go.dev/dl",
		)
	}
	_ = path
	return nil
}

func IsMainFile(path string) (bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return false, err
	}
	if node.Name.Name != "main" {
		return false, nil
	}
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.Name == "main" && fn.Recv == nil {
				return true, nil
			}
		}
	}

	return false, nil
}

func ValidateProjectDirectory(dir string, spinner *Spinner) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("project directory does not exist: %s", dir)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("project path is not a directory: %s", dir)
	}

	var hasGoFiles bool
	var hasGoMod bool
	var mainFile string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".go") {
			hasGoFiles = true
			ok, err := IsMainFile(path)
			if err != nil {
				return err
			}

			if ok {
				mainFile = path
			}
		}
		if name == "go.mod" {
			hasGoMod = true
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed walking project directory: %v", err)
	}
	if !hasGoFiles {
		return "", fmt.Errorf("no Go source files found anywhere in directory: %s", dir)
	}
	if !hasGoMod {
		spinner.BufferWarn(
			fmt.Sprintf("no go.mod found in %s — may not be a proper Go module", dir),
		)
	}
	if mainFile == "" {
		return "", fmt.Errorf("no valid main() entrypoint found")
	}

	return mainFile, nil
}

func GetAvailablePlatforms() ([]GoDist, error) {
	cmd := exec.Command("go", "tool", "dist", "list", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get available platform list: %v", err)
	}

	var platforms []GoDist
	if err := json.Unmarshal(output, &platforms); err != nil {
		return nil, fmt.Errorf("failed to parse available platform list: %v", err)
	}
	return platforms, nil
}

func ParseTargets(targetF string, allPlatforms []GoDist, spinner *Spinner) []GoDist {
	if targetF == "" {
		return allPlatforms
	}
	var selectedPlatforms []GoDist
	targets := strings.SplitSeq(targetF, ",")
	for target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		if strings.Contains(target, "/") {
			parts := strings.Split(target, "/")
			if len(parts) == 2 {
				found := false
				targetOS := strings.TrimSpace(parts[0])
				targetArch := strings.TrimSpace(parts[1])

				for _, platform := range allPlatforms {
					if platform.GOOS == targetOS && platform.GOARCH == targetArch {
						selectedPlatforms = append(selectedPlatforms, platform)
						found = true
						break
					}
				}
				if !found {
					spinner.BufferWarn(fmt.Sprintf("unrecognized target %q — skipping", target))
				}
			}
		} else {
			for _, platform := range allPlatforms {
				if platform.GOOS == target {
					selectedPlatforms = append(selectedPlatforms, platform)
				}
			}
		}
	}

	return selectedPlatforms
}

func FilterFirstClass(platforms []GoDist) []GoDist {
	var result []GoDist
	for _, p := range platforms {
		if p.FirstClass {
			result = append(result, p)
		}
	}
	return result
}
