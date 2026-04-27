package cmd

import (
	"github.com/spf13/cobra"
)

var version = "git"

const (
	dirInclude = "include"
	dirSrc     = "src"
	dirTests   = "tests"
	dirSubproj = "subprojects"
	dirBuild   = "build"
)

var rootCmd = &cobra.Command{
	Use:     "cppup",
	Short:   "A scaffold tool for C++ projects",
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(fmtCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(addCmd)
}
