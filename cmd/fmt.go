package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
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
			changed, err := clangFormatFiles(files)
			if err != nil {
				fail("Format failed")
				os.Exit(1)
			}
			if changed == 0 {
				fmt.Printf("%salready formatted%s\n", colorDim, colorReset)
				return
			}
			success("Formatted", fmt.Sprintf("clang-format · %d/%d files changed", changed, len(files)))
			return
		}
		success("Formatted", "")
	},
}

func clangFormatFiles(files []string) (int, error) {
	changed := 0
	for _, f := range files {
		original, err := os.ReadFile(f)
		if err != nil {
			return 0, err
		}
		out, err := exec.Command("clang-format", f).Output()
		if err != nil {
			return 0, err
		}
		if !bytes.Equal(original, out) {
			if err := os.WriteFile(f, out, 0o644); err != nil {
				return 0, err
			}
			changed++
		}
	}
	return changed, nil
}

func findCppFiles(root string, excludes ...string) []string {
	var files []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == dirBuild || name == dirSubproj || (name != "." && strings.HasPrefix(name, ".")) {
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
