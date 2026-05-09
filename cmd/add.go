package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ruixi-rebirth/cppup/template"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var (
	addFlagFramework        string
	addFlagFrameworkVersion string
)

var addCmd = &cobra.Command{
	Use:   "add <component>",
	Short: "Add a component to an existing project",
}

var addTestsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Add a test framework to the project",
	Run:   runAddTests,
}

var addClangFormatCmd = &cobra.Command{
	Use:   "clang-format",
	Short: "Add .clang-format to the project",
	Run: func(cmd *cobra.Command, args []string) {
		content, err := template.ClangFormat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		if err := addSingleFile(".clang-format", content); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		success("Added", ".clang-format")
	},
}

var addClangTidyCmd = &cobra.Command{
	Use:   "clang-tidy",
	Short: "Add .clang-tidy to the project",
	Run: func(cmd *cobra.Command, args []string) {
		content, err := template.ClangTidy()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		if err := addSingleFile(".clang-tidy", content); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		success("Added", ".clang-tidy")
	},
}

var addNixCmd = &cobra.Command{
	Use:   "nix",
	Short: "Add flake.nix and .envrc to the project",
	Run:   runAddNix,
}

func init() {
	addTestsCmd.Flags().StringVar(&addFlagFramework, "framework", "", "Test framework: "+join(validFrameworks))
	addTestsCmd.Flags().StringVar(&addFlagFrameworkVersion, "version", "", "Test framework version tag (default per framework)")
	addCmd.AddCommand(addTestsCmd, addClangFormatCmd, addClangTidyCmd, addNixCmd)
}

// addSingleFile finds the project root and writes a file, rejecting if it already exists.
func addSingleFile(path, content string) error {
	if err := findProjectRoot(); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%q already exists", path)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func runAddTests(cmd *cobra.Command, args []string) {
	if err := findProjectRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	meta, err := readMeta()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s not a cppup project (missing %s)\n", colorBrightRed, colorReset, metaFile)
		os.Exit(1)
	}
	if meta.TestFramework != "" {
		fmt.Fprintf(os.Stderr, "%s✗%s tests already configured: %s\n", colorBrightRed, colorReset, meta.TestFramework)
		os.Exit(1)
	}

	framework := addFlagFramework
	version := addFlagFrameworkVersion

	if isTTY() {
		rl, err := readline.NewEx(&readline.Config{
			AutoComplete: &completer{},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		defer rl.Close()

		if framework == "" {
			// Set Tab completion candidates
			if rl.Config.AutoComplete != nil {
				if c, ok := rl.Config.AutoComplete.(*completer); ok {
					c.candidates = validFrameworks
				}
			}
			framework, err = promptOr(rl, "Test framework ("+join(validFrameworks)+")", "catch2", "", validFrameworks...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
				os.Exit(1)
			}
		}
		framework = strings.ToLower(framework)
		if version == "" {
			fw, _ := template.TFrameworkByName(framework)
			v, err := prompt(rl, "Test framework version", fw.DefaultTag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
				os.Exit(1)
			}
			version = normalizeTag(v)
		}
	} else {
		framework = strings.ToLower(framework)
		if framework == "" {
			framework = "catch2"
		}
	}
	version = normalizeTag(version)

	if !contains(validFrameworks, framework) {
		fmt.Fprintf(os.Stderr, "%s✗%s invalid test framework %q: choose %s\n", colorBrightRed, colorReset, framework, join(validFrameworks))
		os.Exit(1)
	}
	if version == "" {
		if fw, ok := template.TFrameworkByName(framework); ok {
			version = fw.DefaultTag
		}
	}
	if !validTagFormat(version) {
		fmt.Fprintf(os.Stderr, "%s✗%s invalid version %q: expected vX.Y.Z\n", colorBrightRed, colorReset, version)
		os.Exit(1)
	}

	if err := os.MkdirAll(dirTests, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	files := map[string]string{}
	var renderErr error

	if files[filepath.Join(dirTests, "test_main.cpp")], renderErr = template.TestSourceCpp(framework); renderErr != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, renderErr)
		os.Exit(1)
	}

	switch meta.BuildSystem {
	case "cmake":
		if files[filepath.Join(dirTests, "CMakeLists.txt")], renderErr = template.TestCMakeLists(meta.Name, meta.Type, framework, version); renderErr != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, renderErr)
			os.Exit(1)
		}
	case "meson":
		if files[filepath.Join(dirTests, "meson.build")], renderErr = template.TestMesonBuild(meta.Name, meta.Type, framework); renderErr != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, renderErr)
			os.Exit(1)
		}
		if err := os.MkdirAll(dirSubproj, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		if files[filepath.Join(dirSubproj, template.MesonWrapFilename(framework))], renderErr = template.MesonWrap(framework, version); renderErr != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, renderErr)
			os.Exit(1)
		}
	}

	for path, content := range files {
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %q already exists\n", colorBrightRed, colorReset, path)
			os.Exit(1)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
	}

	if err := appendTestsToRootBuild(meta.BuildSystem); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	if meta.BuildSystem == "meson" {
		appendMesonSubprojectsToGitignore()
	}

	meta.TestFramework = framework
	if err := writeMetaTo(metaFile, meta); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	success("Added", framework+" "+version)
}

func appendTestsToRootBuild(buildSystem string) error {
	bs, ok := template.BuildSystemByName(buildSystem)
	if !ok {
		return fmt.Errorf("unknown build system %q", buildSystem)
	}
	var addition, marker string
	switch buildSystem {
	case "cmake":
		addition = "\nenable_testing()\nadd_subdirectory(tests)\n"
		marker = "add_subdirectory(tests)"
	case "meson":
		addition = "\nsubdir('tests')\n"
		marker = "subdir('tests')"
	default:
		return fmt.Errorf("no test subdirectory snippet defined for build system %q", buildSystem)
	}
	existing, err := os.ReadFile(bs.BuildFile)
	if err != nil {
		return err
	}
	if strings.Contains(string(existing), marker) {
		return fmt.Errorf("%s already contains %q", bs.BuildFile, marker)
	}
	f, err := os.OpenFile(bs.BuildFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(addition)
	return err
}

func runAddNix(cmd *cobra.Command, args []string) {
	if err := findProjectRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	meta, err := readMeta()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s not a cppup project (missing %s)\n", colorBrightRed, colorReset, metaFile)
		os.Exit(1)
	}

	if _, err := os.Stat("flake.nix"); err == nil {
		fmt.Fprintf(os.Stderr, "%s✗%s flake.nix already exists\n", colorBrightRed, colorReset)
		os.Exit(1)
	}

	version := meta.Version
	if version == "" {
		version = defaultVersion
		if isTTY() {
			rl, err := readline.NewEx(&readline.Config{
				AutoComplete: &completer{},
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
				os.Exit(1)
			}
			defer rl.Close()
			v, err := prompt(rl, "Project version", version)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
				os.Exit(1)
			}
			version = v
			meta.Version = version
			_ = writeMetaTo(metaFile, meta)
		}
	}

	flakeContent, err := template.FlakeNix(meta.Name, meta.BuildSystem, version, meta.TestFramework)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}
	if err := os.WriteFile("flake.nix", []byte(flakeContent), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	treefmtContent, err := template.TreefmtNix()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}
	if err := os.WriteFile("treefmt.nix", []byte(treefmtContent), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	envrcContent, err := template.Envrc()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}
	if err := os.WriteFile(".envrc", []byte(envrcContent), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}
	appendNixToGitignore()

	success("Added", "flake.nix, treefmt.nix, .envrc")
	fmt.Printf("  %s$%s direnv allow%s  %s# requires direnv + nix-direnv%s\n", colorDim, colorReset, colorDim, colorDim, colorReset)
}

func appendMesonSubprojectsToGitignore() {
	const entries = "\nsubprojects/*\n!subprojects/*.wrap\n"
	content, err := os.ReadFile(".gitignore")
	if err != nil {
		return
	}
	if strings.Contains(string(content), "subprojects/*") {
		return
	}
	f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(entries)
}

func appendNixToGitignore() {
	const nixEntries = "\n.direnv/\nresult\nresult-*\n"
	content, err := os.ReadFile(".gitignore")
	if err != nil {
		return
	}
	if strings.Contains(string(content), ".direnv/") {
		return
	}
	f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(nixEntries)
}
