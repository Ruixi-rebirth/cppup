package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove build artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		if err := findProjectRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		if _, err := os.Stat(dirBuild); os.IsNotExist(err) {
			fmt.Printf("%snothing to clean%s\n", colorDim, colorReset)
			return
		}
		if err := os.RemoveAll(dirBuild); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		if err := os.Remove("compile_commands.json"); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		success("Cleaned", dirBuild+"/")
	},
}
