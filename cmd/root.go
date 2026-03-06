package cmd

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Moukhtar-youssef/GoBuilder/internal"
	"github.com/spf13/cobra"
)

// go build -ldflags "-X github.com/Moukhtar-youssef/GoBuilder/cmd.version=1.0.0"
var version = "dev"

var (
	projectDir       string
	outputDir        string
	targets          string
	binaryName       string
	specificEntry    string
	concurrencyLimit int
	compress         bool
	firstClassOnly   bool
	dryRun           bool
)

var rootCmd = &cobra.Command{
	Use:     "GoBuilder",
	Version: version,
	Short:   "Cross-compile your Go project for every platform in one command",
	Long: `GoBuilder is a CLI tool that cross-compiles your Go project for multiple
platforms and architectures in parallel.

Instead of manually setting GOOS and GOARCH for every target, GoBuilder
discovers your project's entry point, resolves all available Go toolchain
targets, and builds them concurrently — then prints a summary of every
output with its file size.

Examples:
  # Build for all available platforms
  GoBuilder -p ./myproject

  # Build for specific targets only
  GoBuilder -p ./myproject -t linux/amd64,darwin/arm64,windows/amd64

  # Build first-class targets only with a custom binary name
  GoBuilder -p ./myproject --first-class-only -n mybinary

  # Build and compress output binaries
  GoBuilder -p ./myproject --compress -o ./releases

  # Preview what would be built without building
  GoBuilder -p ./myproject -t linux/amd64,darwin/arm64 --dry-run

  # Point at a specific entry file
  GoBuilder -p ./myproject -e cmd/server/main.go`,

	Run: func(cmd *cobra.Command, args []string) {
		spinner := internal.NewSpinner("Resolving project path...")
		spinner.Start()

		err := internal.CheckGoInstalled()
		if err != nil {
			spinner.Fail(err.Error())
			os.Exit(1)
		}

		absProjectDir, err := filepath.Abs(projectDir)
		if err != nil {
			spinner.Fail("Could not resolve project path: " + err.Error())
			os.Exit(1)
		}

		spinner.SetMessage("Validating project directory...")
		mainFilePath, err := internal.ValidateProjectDirectory(absProjectDir, spinner)
		if err != nil {
			spinner.Fail("Project validation failed: " + err.Error())
			os.Exit(1)
		}

		if binaryName == "" {
			binaryName = filepath.Base(absProjectDir)
		}

		if specificEntry != "" {
			spinner.SetMessage("Validating entry file...")
			if _, err := os.Stat(specificEntry); err != nil {
				spinner.Fail("Entry file not found: " + specificEntry)
				os.Exit(1)
			}
			ok, err := internal.IsMainFile(specificEntry)
			if err != nil {
				spinner.Fail("Error checking entry file: " + err.Error())
				os.Exit(1)
			}
			if !ok {
				spinner.Fail("Entry file has no func main(): " + specificEntry)
				os.Exit(1)
			}
			mainFilePath = specificEntry
		}

		spinner.SetMessage("Resolving build targets...")
		platforms, err := internal.GetAvailablePlatforms()
		if err != nil {
			spinner.Fail("Failed to get available platforms: " + err.Error())
			os.Exit(1)
		}

		selectedTargets := internal.ParseTargets(targets, platforms, spinner)

		if firstClassOnly {
			selectedTargets = internal.FilterFirstClass(selectedTargets)
		}

		if len(selectedTargets) == 0 {
			spinner.Fail("No valid targets found — check your -t flag values")
			os.Exit(1)
		}

		config := internal.BuildConfig{
			ProjectDir:  absProjectDir,
			OutputDir:   outputDir,
			BinaryName:  binaryName,
			Targets:     selectedTargets,
			Compress:    compress,
			Concurrency: concurrencyLimit,
			DryRun:      dryRun,
		}

		internal.BuildAllPlatforms(config, mainFilePath, spinner)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().
		StringVarP(&projectDir, "project", "p", ".", "Path to the Go project directory")
	rootCmd.Flags().
		StringVarP(&outputDir, "output", "o", "./dist", "Directory to write build outputs")
	rootCmd.Flags().
		StringVarP(&targets, "targets", "t", "", "Comma-separated build targets e.g. linux/amd64,darwin/arm64 (default: all platforms)")
	rootCmd.Flags().
		StringVarP(&binaryName, "name", "n", "", "Output binary name (default: project directory name)")
	rootCmd.Flags().
		StringVarP(&specificEntry, "entry", "e", "", "Path to a specific main.go entry file (default: auto-detected)")
	rootCmd.Flags().
		IntVarP(&concurrencyLimit, "concurrency", "c", runtime.NumCPU(), "Max parallel builds (default: number of CPU cores)")
	rootCmd.Flags().
		BoolVarP(&compress, "compress", "z", false, "Compress each binary after building (.zip for Windows, .tar.gz for others)")
	rootCmd.Flags().
		BoolVar(&firstClassOnly, "first-class-only", false, "Only build for first-class supported Go platforms")
	rootCmd.Flags().
		BoolVar(&dryRun, "dry-run", false, "Print targets that would be built without building them")
}
