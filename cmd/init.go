package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Scaffold a project into the current directory",
	Args:  cobra.MaximumNArgs(1),
	Run:   runInit,
}

func init() {
	initCmd.Flags().StringVar(&flagType, "type", "", "Project type: "+join(validTypes))
	initCmd.Flags().StringVar(&flagBuild, "build", "", "Build system: "+join(validBuilds))
	initCmd.Flags().StringVar(&flagBuildVersion, "build-version", "", "Minimum build system version (e.g. 3.25 for cmake)")
	initCmd.Flags().StringVar(&flagStd, "std", "", "C++ standard: "+join(validStds))
	initCmd.Flags().StringVar(&flagVersion, "version", "", "Project version")
	initCmd.Flags().StringVar(&flagTests, "tests", "", "Test framework: "+join(validFrameworks))
	initCmd.Flags().StringVar(&flagTestsVersion, "tests-version", "", "Test framework version tag (e.g. 3.7.1)")
	initCmd.Flags().BoolVar(&flagClangFormat, "clang-format", false, "Generate .clang-format")
	initCmd.Flags().BoolVar(&flagClangTidy, "clang-tidy", false, "Generate .clang-tidy")
	initCmd.Flags().BoolVar(&flagNix, "nix", false, "Generate flake.nix and .envrc")
	initCmd.Flags().BoolVar(&flagNoGit, "no-git", false, "Skip git initialization")
}

func runInit(cmd *cobra.Command, args []string) {
	runGenericScaffold(cmd, args, false)
}

func cwdName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "project"
	}
	return filepath.Base(cwd)
}
