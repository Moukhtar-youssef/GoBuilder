package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"text/tabwriter"
)

type GoDist struct {
	GOOS         string `json:"GOOS"`
	GOARCH       string `json:"GOARCH"`
	CgoSupported bool   `json:"CgoSupported"`
	FirstClass   bool   `json:"FirstClass"`
}

type BuildConfig struct {
	ProjectDir  string
	OutputDir   string
	BinaryName  string
	Targets     []GoDist
	Compress    bool
	Concurrency int
	DryRun      bool
}

type buildResult struct {
	platform GoDist
	success  bool
	err      error
	fileSize int64
	output   string
}

func buildForPlatform(config BuildConfig, platform GoDist, mainFilePath string) (string, error) {
	filename := fmt.Sprintf("%s_%s_%s", config.BinaryName, platform.GOOS, platform.GOARCH)
	if platform.GOOS == "windows" {
		filename += ".exe"
	}

	outputPath := filepath.Join(config.OutputDir, filename)

	cmd := exec.Command("go", "build", "-o", outputPath, mainFilePath)
	cmd.Dir = config.ProjectDir
	cmd.Env = append(os.Environ(),
		"GOOS="+platform.GOOS,
		"GOARCH="+platform.GOARCH,
		"CGO_ENABLED=0",
	)

	return outputPath, cmd.Run()
}

func TIERFomrater(firstClass bool) string {
	if firstClass {
		return "first-class"
	} else {
		return "community"
	}
}

func BuildAllPlatforms(config BuildConfig, mainFilePath string, spinner *Spinner) {
	if config.DryRun {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "TARGET\tOUTPUT FILE\tTIER")
		fmt.Fprintln(w, "------\t-----------\t----")
		for _, p := range config.Targets {
			filename := fmt.Sprintf("%s_%s_%s", config.BinaryName, p.GOOS, p.GOARCH)
			if p.GOOS == "windows" {
				filename += ".exe"
			}
			fmt.Fprintf(
				w,
				"%s/%s\t%s\t%s\n",
				p.GOOS,
				p.GOARCH,
				filepath.Join(config.OutputDir, filename),
				TIERFomrater(p.FirstClass),
			)
		}
		spinner.Stop("Dry run — the following targets would be built:")
		w.Flush()
		return
	}
	err := os.MkdirAll(config.OutputDir, 0o755)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Failed to create output directory: %v", err))
		os.Exit(1)
	}

	total := len(config.Targets)

	concurrency := config.Concurrency
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int
	var results []buildResult

	limiterChan := make(chan struct{}, concurrency)

	spinner.SetMessage(fmt.Sprintf("Building 0 / %d...", total))

	for _, platform := range config.Targets {
		wg.Add(1)
		go func(p GoDist) {
			defer wg.Done()
			limiterChan <- struct{}{}
			defer func() { <-limiterChan }()

			outputPath, err := buildForPlatform(config, p, mainFilePath)
			result := buildResult{
				platform: p,
				output:   outputPath,
			}

			if err != nil {
				result.success = false
				result.err = err
			} else {
				result.success = true
				if info, statErr := os.Stat(outputPath); statErr == nil {
					result.fileSize = info.Size()
				}
			}
			mu.Lock()
			results = append(results, result)
			if result.success {
				successCount++
			}
			done := len(results)
			failCount := done - successCount
			spinner.SetMessage(fmt.Sprintf(
				"Building %d / %d  (✓ %d  ✗ %d)",
				done, total, successCount, failCount,
			))
			mu.Unlock()
		}(platform)
	}

	wg.Wait()
	failCount := len(results) - successCount
	summary := fmt.Sprintf("Built %d / %d targets", successCount, total)
	if failCount == 0 {
		spinner.Stop(summary)
	} else {
		spinner.Fail(fmt.Sprintf("%s — %d failed", summary, failCount))
	}
	PrintSummaryTable(results)

	if config.Compress && successCount > 0 {
		compressSpinner := NewSpinner("Compressing binaries...")
		compressSpinner.Start()
		CompressResults(results, compressSpinner)
		compressSpinner.Stop(fmt.Sprintf("Compressed %d binaries", successCount))
	}
}
