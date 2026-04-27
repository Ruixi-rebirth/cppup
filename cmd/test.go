package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	testRelease bool
	testAsan    bool
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Build and run tests",
	Run: func(cmd *cobra.Command, args []string) {
		if err := findProjectRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		if !hasTests() {
			fmt.Fprintf(os.Stderr, "%s✗%s no test framework configured, use --tests when creating or cppup add tests\n",
				colorBrightRed, colorReset)
			os.Exit(1)
		}

		mode := BuildMode{Release: testRelease, Asan: testAsan}
		bs, err := buildProject(mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		dir := mode.dir()
		step("Testing", "")

		var testErr error
		switch bs {
		case "cmake":
			testErr = execCommand("ctest", "--test-dir", dir, "--output-on-failure")
		case "meson":
			testErr = execCommand("meson", "test", "-C", dir)
		}

		if testErr != nil {
			fail("Tests failed")
			os.Exit(1)
		}
		success("Passed", "")
	},
}

func init() {
	testCmd.Flags().BoolVar(&testRelease, "release", false, "Test the release build")
	testCmd.Flags().BoolVar(&testAsan, "asan", false, "Test the sanitizer build")
}
