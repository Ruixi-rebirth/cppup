package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var checkNoTests bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run clang-tidy on all source files",
	Run: func(cmd *cobra.Command, args []string) {
		if err := findProjectRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		if !hasTool("clang-tidy") {
			fmt.Fprintf(os.Stderr, "%s✗%s clang-tidy not found in PATH\n", colorBrightRed, colorReset)
			os.Exit(1)
		}

		if _, err := os.Stat("compile_commands.json"); os.IsNotExist(err) {
			if _, err := buildProject(BuildMode{}); err != nil {
				fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
				os.Exit(1)
			}
		}

		excludes := []string{}
		if checkNoTests {
			excludes = append(excludes, "tests")
		}
		files := findCppFiles(".", excludes...)
		if len(files) == 0 {
			fmt.Printf("%snothing to check%s\n", colorDim, colorReset)
			return
		}

		step("Checking", fmt.Sprintf("clang-tidy · %d files", len(files)))
		if err := execCommand("clang-tidy", files...); err != nil {
			fail("Check failed")
			os.Exit(1)
		}
		success("Passed", "")
	},
}

func init() {
	checkCmd.Flags().BoolVar(&checkNoTests, "no-tests", false, "Exclude tests directory from checking")
}
