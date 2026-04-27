package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	runRelease bool
	runAsan    bool
)

var runCmd = &cobra.Command{
	Use:   "run [-- args...]",
	Short: "Build and run the project",
	Long:  "Build and run. Only available for exe projects.\n\nSupports passing arguments to the binary:\n  cppup run -- arg1 arg2\n  cppup run --release -- arg1 arg2",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := findProjectRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		if !isExe() {
			fmt.Fprintf(os.Stderr, "%s✗%s run requires an exe project\n", colorBrightRed, colorReset)
			os.Exit(1)
		}

		mode := BuildMode{Release: runRelease, Asan: runAsan}
		if _, err := buildProject(mode); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		exePath := filepath.Join(mode.dir(), resolveProjectName())

		c := exec.Command(exePath, args...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		if err := c.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
	},
}

func init() {
	runCmd.Flags().BoolVar(&runRelease, "release", false, "Run the release build")
	runCmd.Flags().BoolVar(&runAsan, "asan", false, "Run the sanitizer build")
}
