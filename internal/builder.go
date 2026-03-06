package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

type GoDist struct {
	GOOS         string `json:"GOOS"`
	GOARCH       string `json:"GOARCH"`
	CgoSupported bool   `json:"CgoSupported"`
	FirstClass   bool   `json:"FirstClass"`
}

type BuildConfig struct {
	ProjectDir string
	OutputDir  string
	BinaryName string
	Targets    []GoDist
}

func buildForPlatform(config BuildConfig, platform GoDist) error {
	filename := fmt.Sprintf("%s_%s_%s", config.BinaryName, platform.GOOS, platform.GOARCH)
	if platform.GOOS == "windows" {
		filename += ".exe"
	}

	outputPath := filepath.Join(config.OutputDir, filename)

	cmd := exec.Command("go", "build", "-o", outputPath)
	cmd.Dir = config.ProjectDir
	cmd.Env = append(os.Environ(),
		"GOOS="+platform.GOOS,
		"GOARCH="+platform.GOARCH,
		"CGO_ENABLED=0",
	)

	return cmd.Run()
}

func BuildAllPlatforms(config BuildConfig) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int
	var failCount int
	limiterChan := make(chan struct{}, runtime.NumCPU())

	for _, platform := range config.Targets {
		wg.Add(1)
		go func(p GoDist) {
			defer wg.Done()
			limiterChan <- struct{}{}
			defer func() { <-limiterChan }()

			err := buildForPlatform(config, p)
			mu.Lock()
			if err != nil {
				failCount++
				fmt.Printf("Failed to build %s/%s: %v\n", p.GOOS, p.GOARCH, err)
			} else {
				successCount++
			}
			mu.Unlock()
		}(platform)
	}

	wg.Wait()
}
