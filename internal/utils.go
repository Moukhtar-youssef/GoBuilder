package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ValidateProjectDirectory(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("project directory does not exist: %s", dir)
	}

	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read project directory: %v", err)
	}

	hasGoFiles := false
	hasGoMod := false

	for _, entry := range entries {
		if !entry.IsDir() {
			if strings.HasSuffix(entry.Name(), ".go") {
				hasGoFiles = true
			}
			if entry.Name() == "go.mod" {
				hasGoMod = true
			}
		}
	}

	if !hasGoFiles {
		return fmt.Errorf("no Go source files found in directory: %s", dir)
	}

	if !hasGoMod {
		fmt.Printf("Warning: no go.mod found in %s - this may not be a proper Go module\n", dir)
	}

	return nil
}

func GetAvilablePlatforms() ([]GoDist, error) {
	cmd := exec.Command("go", "tool", "dist", "list", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get avilable platform list: %w", err)
	}

	var platforms []GoDist
	if err := json.Unmarshal(output, &platforms); err != nil {
		return nil, fmt.Errorf("failed to parse avilable platform list: %w", err)
	}
	return platforms, nil
}

func ParseTargets(targetF string, allPlatforms []GoDist) []GoDist {
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
				targetOS := strings.TrimSpace(parts[0])
				targetArch := strings.TrimSpace(parts[1])

				for _, platform := range allPlatforms {
					if platform.GOOS == targetOS && platform.GOARCH == targetArch {
						selectedPlatforms = append(selectedPlatforms, platform)
						break
					}
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
