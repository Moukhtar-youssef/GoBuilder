/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Moukhtar-youssef/GoBuilder/internal"
	"github.com/spf13/cobra"
)

var (
	projectDir string
	outputDir  string
	targets    string
	BinaryName string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "GoBuilder",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		absProjectDir, err := filepath.Abs(projectDir)
		if err != nil {
			log.Fatal("Error resolving project directory: %w\n", err)
		}
		fmt.Println("Validating project dir")
		err = internal.ValidateProjectDirectory(absProjectDir)
		if err != nil {
			log.Fatal("Error: %w\n", err)
		}
		if BinaryName == "" {
			BinaryName = filepath.Base(absProjectDir)
		}
		fmt.Println("Grabbing all avilable platforms")
		platforms, err := internal.GetAvilablePlatforms()
		if err != nil {
			log.Fatal("Error getting all avilable platforms: %w", err)
		}
		SelectedTarget := internal.ParseTargets(targets, platforms)
		config := internal.BuildConfig{
			ProjectDir: absProjectDir,
			OutputDir:  outputDir,
			BinaryName: BinaryName,
			Targets:    SelectedTarget,
		}
		internal.BuildAllPlatforms(config)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.GoBuilder.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().
		StringVarP(&projectDir, "project", "p", ".", "Specify the project dir")
	rootCmd.Flags().
		StringVarP(&outputDir, "output", "o", "./dist", "Specify where the build files are stored")
	rootCmd.Flags().
		StringVarP(&targets, "targets", "t", "", "Specify which target you want to build to if left empty every avilable build profile will run")
	rootCmd.Flags().
		StringVarP(&BinaryName, "name", "n", "", "Specify the name of the build file if left empty the name will default to the folder name of the project")
}
