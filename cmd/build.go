package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	buildRelease bool
	buildAsan    bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := buildProject(BuildMode{Release: buildRelease, Asan: buildAsan}); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
	},
}

func init() {
	buildCmd.Flags().BoolVar(&buildRelease, "release", false, "Build in release mode")
	buildCmd.Flags().BoolVar(&buildAsan, "asan", false, "Build with AddressSanitizer + UBSan")
}
