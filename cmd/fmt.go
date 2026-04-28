package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format all source files",
	Run: func(cmd *cobra.Command, args []string) {
		if err := findProjectRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		if _, err := os.Stat("flake.nix"); err == nil {
			step("Formatting", "nix fmt")
			if err := execCommand("nix", "fmt"); err != nil {
				fail("Format failed")
				os.Exit(1)
			}
		} else {
			if !hasTool("clang-format") {
				fmt.Fprintf(os.Stderr, "%s✗%s clang-format not found in PATH\n", colorBrightRed, colorReset)
				os.Exit(1)
			}
			files := findCppFiles(".")
			if len(files) == 0 {
				fmt.Printf("%snothing to format%s\n", colorDim, colorReset)
				return
			}
			step("Formatting", fmt.Sprintf("clang-format · %d files", len(files)))
			if err := execCommand("clang-format", append([]string{"-i"}, files...)...); err != nil {
				fail("Format failed")
				os.Exit(1)
			}
		}
		success("Formatted", "")
	},
}

func findCppFiles(root string, excludes ...string) []string {
	var files []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == dirBuild || name == dirSubproj || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			for _, ex := range excludes {
				if name == ex {
					return filepath.SkipDir
				}
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".cpp", ".hpp", ".h", ".cc", ".cxx", ".hxx":
			files = append(files, path)
		}
		return nil
	})
	return files
}
