package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Ruixi-rebirth/cppup/template"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

const (
	defaultType    = "exe"
	defaultBuild   = "cmake"
	defaultStd     = "17"
	defaultVersion = "0.1.0"
)

var (
	validTypes      = []string{"exe", "lib-static", "lib-shared", "lib-header"}
	validStds       = []string{"11", "14", "17", "20", "23"}
	validBuilds     = buildSystemNames()
	validFrameworks = frameworkNames()
)

func buildSystemNames() []string {
	s := make([]string, len(template.BuildSystems))
	for i, b := range template.BuildSystems {
		s[i] = b.Name
	}
	return s
}

func frameworkNames() []string {
	s := make([]string, len(template.TFrameworks))
	for i, f := range template.TFrameworks {
		s[i] = f.Name
	}
	return s
}

// ProjectConfig holds all settings for a project.
type ProjectConfig struct {
	Name         string
	Type         string
	Build        string
	BuildVersion string
	Std          string
	Version      string
	Tests        string
	TestsVersion string
	ClangFormat  bool
	ClangTidy    bool
	Nix          bool
	NoGit        bool
}

var (
	flagType         string
	flagBuild        string
	flagBuildVersion string
	flagStd          string
	flagVersion      string
	flagTests        string
	flagTestsVersion string
	flagClangFormat  bool
	flagClangTidy    bool
	flagNix          bool
	flagNoGit        bool
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Scaffold a project into a new directory",
	Args:  cobra.MaximumNArgs(1),
	Run:   runNew,
}

func init() {
	newCmd.Flags().StringVar(&flagType, "type", "", "Project type: "+join(validTypes))
	newCmd.Flags().StringVar(&flagBuild, "build", "", "Build system: "+join(validBuilds))
	newCmd.Flags().StringVar(&flagBuildVersion, "build-version", "", "Minimum build system version (e.g. 3.25 for cmake)")
	newCmd.Flags().StringVar(&flagStd, "std", "", "C++ standard: "+join(validStds))
	newCmd.Flags().StringVar(&flagVersion, "version", "", "Project version")
	newCmd.Flags().StringVar(&flagTests, "tests", "", "Test framework: "+join(validFrameworks))
	newCmd.Flags().StringVar(&flagTestsVersion, "tests-version", "", "Test framework version tag (e.g. 3.7.1)")
	newCmd.Flags().BoolVar(&flagClangFormat, "clang-format", false, "Generate .clang-format")
	newCmd.Flags().BoolVar(&flagClangTidy, "clang-tidy", false, "Generate .clang-tidy")
	newCmd.Flags().BoolVar(&flagNix, "nix", false, "Generate flake.nix and .envrc")
	newCmd.Flags().BoolVar(&flagNoGit, "no-git", false, "Skip git initialization")
}

func join(vals []string) string { return strings.Join(vals, "/") }

func normalizeTag(v string) string {
	if v != "" && !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// queryInstalledVersion returns the version string of an installed tool (e.g. "cmake", "meson").
func queryInstalledVersion(tool string) string {
	out, err := exec.Command(tool, "--version").Output()
	if err != nil {
		return ""
	}
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	for _, field := range strings.Fields(line) {
		if len(field) > 0 && field[0] >= '0' && field[0] <= '9' {
			return field
		}
	}
	return ""
}

// validTagFormat checks that a git tag matches vX.Y or vX.Y.Z.
func validTagFormat(v string) bool {
	if !strings.HasPrefix(v, "v") {
		return false
	}
	return validVersionFormat(v[1:])
}

// validVersionFormat checks that v matches X.Y or X.Y.Z (digits only).
func validVersionFormat(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

type completer struct {
	candidates []string
}

func (c *completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	var matches [][]rune
	lineStr := string(line[:pos])
	for _, cand := range c.candidates {
		if strings.HasPrefix(cand, lineStr) {
			matches = append(matches, []rune(cand[pos:]))
		}
	}
	return matches, pos
}

func promptOr(rl *readline.Instance, question, defaultVal, flagVal string, validOptions ...string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}

	// Set Tab completion candidates if available
	if rl.Config.AutoComplete != nil {
		if c, ok := rl.Config.AutoComplete.(*completer); ok {
			c.candidates = validOptions
		}
	}

	for {
		input, err := prompt(rl, question, defaultVal)
		if err != nil {
			return "", err
		}
		if len(validOptions) == 0 {
			return input, nil
		}
		input = strings.ToLower(input)
		// Exact match
		if contains(validOptions, input) {
			return input, nil
		}
		// First letter match (only if unambiguous)
		var matches []string
		for _, opt := range validOptions {
			if strings.HasPrefix(opt, input) {
				matches = append(matches, opt)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		fmt.Printf("%s✗%s invalid choice %q: choose %s\n", colorBrightRed, colorReset, input, join(validOptions))
	}
}

func prompt(rl *readline.Instance, question, defaultVal string) (string, error) {
	label := fmt.Sprintf("%s%s:%s", colorBold+colorBrightCyan, question, colorReset)
	if defaultVal != "" {
		rl.SetPrompt(fmt.Sprintf("%s %s[%s]%s ", label, colorDim, defaultVal, colorReset))
	} else {
		rl.SetPrompt(fmt.Sprintf("%s ", label))
	}
	input, err := rl.Readline()
	if err == io.EOF {
		return defaultVal, nil
	}
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	input = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, input)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

func promptBool(rl *readline.Instance, question string, defaultVal bool) (bool, error) {
	def := "n"
	if defaultVal {
		def = "y"
	}
	answer, err := prompt(rl, question+" (y/n)", def)
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(answer)
	return answer == "y" || answer == "yes", nil
}

func isTTY() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && (fi.Mode()&os.ModeCharDevice != 0)
}

func collectSettings(rl *readline.Instance, defaultName string) (ProjectConfig, error) {
	var cfg ProjectConfig

	name, err := prompt(rl, "Project name", defaultName)
	if err != nil {
		return cfg, err
	}
	if name == "" {
		return cfg, fmt.Errorf("project name cannot be empty")
	}
	if strings.ContainsAny(name, " /\\") {
		return cfg, fmt.Errorf("project name must not contain spaces or slashes")
	}
	cfg.Name = name

	projType, err := promptOr(rl, "Project type ("+join(validTypes)+")", defaultType, flagType, validTypes...)
	if err != nil {
		return cfg, err
	}
	cfg.Type = projType

	bs, err := promptOr(rl, "Build system ("+join(validBuilds)+")", defaultBuild, flagBuild, validBuilds...)
	if err != nil {
		return cfg, err
	}
	cfg.Build = bs

	buildVersion := flagBuildVersion
	if buildVersion == "" {
		bsInfo, _ := template.BuildSystemByName(bs)
		label := fmt.Sprintf("Min %s version", bs)
		if installed := queryInstalledVersion(bs); installed != "" {
			label = fmt.Sprintf("Min %s version (installed: %s)", bs, installed)
		}
		bv, err := prompt(rl, label, bsInfo.MinVersion)
		if err != nil {
			return cfg, err
		}
		buildVersion = bv
	}
	if !validVersionFormat(buildVersion) {
		return cfg, fmt.Errorf("invalid build system version %q: expected X.Y or X.Y.Z", buildVersion)
	}
	cfg.BuildVersion = buildVersion

	std, err := promptOr(rl, "C++ standard ("+join(validStds)+")", defaultStd, flagStd, validStds...)
	if err != nil {
		return cfg, err
	}
	cfg.Std = std

	version, err := promptOr(rl, "Version", defaultVersion, flagVersion)
	if err != nil {
		return cfg, err
	}
	cfg.Version = version

	tests := flagTests
	testsVersion := flagTestsVersion
	if tests == "" {
		useTests, err := promptBool(rl, "Enable tests", false)
		if err != nil {
			return cfg, err
		}
		if useTests {
			tests, err = promptOr(rl, "Test framework ("+join(validFrameworks)+")", "catch2", "", validFrameworks...)
			if err != nil {
				return cfg, err
			}
		}
	} else {
		tests = strings.ToLower(tests)
		if !contains(validFrameworks, tests) {
			return cfg, fmt.Errorf("invalid test framework %q: choose %s", tests, join(validFrameworks))
		}
	}
	if tests != "" && testsVersion == "" {
		fw, _ := template.TFrameworkByName(tests)
		v, err := prompt(rl, "Test framework version", fw.DefaultTag)
		if err != nil {
			return cfg, err
		}
		testsVersion = normalizeTag(v)
	}
	testsVersion = normalizeTag(testsVersion)
	if testsVersion != "" && !validTagFormat(testsVersion) {
		return cfg, fmt.Errorf("invalid test framework version %q: expected vX.Y.Z", testsVersion)
	}
	cfg.Tests = tests
	cfg.TestsVersion = testsVersion

	clangFormat := flagClangFormat
	if !clangFormat {
		if clangFormat, err = promptBool(rl, "Enable clang-format", false); err != nil {
			return cfg, err
		}
	}
	cfg.ClangFormat = clangFormat

	clangTidy := flagClangTidy
	if !clangTidy {
		if clangTidy, err = promptBool(rl, "Enable clang-tidy", false); err != nil {
			return cfg, err
		}
	}
	cfg.ClangTidy = clangTidy

	nix := flagNix
	if !nix {
		if nix, err = promptBool(rl, "Enable Nix flake", false); err != nil {
			return cfg, err
		}
	}
	cfg.Nix = nix

	noGit := flagNoGit
	if !noGit {
		if noGit, err = promptBool(rl, "Skip git init", false); err != nil {
			return cfg, err
		}
	}
	cfg.NoGit = noGit

	return cfg, nil
}

func configFromFlags(name string) (ProjectConfig, error) {
	if strings.ContainsAny(name, " /\\") {
		return ProjectConfig{}, fmt.Errorf("project name must not contain spaces or slashes")
	}
	cfg := ProjectConfig{
		Name:         name,
		Type:         flagType,
		Build:        flagBuild,
		BuildVersion: flagBuildVersion,
		Std:          flagStd,
		Version:      flagVersion,
		Tests:        flagTests,
		TestsVersion: flagTestsVersion,
		ClangFormat:  flagClangFormat,
		ClangTidy:    flagClangTidy,
		Nix:          flagNix,
		NoGit:        flagNoGit,
	}
	if cfg.Type == "" {
		cfg.Type = defaultType
	}
	if cfg.Build == "" {
		cfg.Build = defaultBuild
	}
	if cfg.Std == "" {
		cfg.Std = defaultStd
	}
	if cfg.Version == "" {
		cfg.Version = defaultVersion
	}
	if cfg.BuildVersion == "" {
		if bs, ok := template.BuildSystemByName(cfg.Build); ok {
			cfg.BuildVersion = bs.MinVersion
		}
	}
	if !validVersionFormat(cfg.BuildVersion) {
		return cfg, fmt.Errorf("invalid build system version %q: expected X.Y or X.Y.Z", cfg.BuildVersion)
	}
	cfg.Tests = strings.ToLower(cfg.Tests)
	cfg.TestsVersion = normalizeTag(cfg.TestsVersion)
	if cfg.Tests != "" && cfg.TestsVersion == "" {
		if fw, ok := template.TFrameworkByName(cfg.Tests); ok {
			cfg.TestsVersion = fw.DefaultTag
		}
	}
	if cfg.TestsVersion != "" && !validTagFormat(cfg.TestsVersion) {
		return cfg, fmt.Errorf("invalid test framework version %q: expected vX.Y.Z", cfg.TestsVersion)
	}
	return cfg, validateSettings(cfg)
}

func printSummary(cfg ProjectConfig) {
	fmt.Printf("\n%s  Project summary%s\n", colorBold, colorReset)
	fmt.Printf("  %sname%s       %s\n", colorDim, colorReset, cfg.Name)
	fmt.Printf("  %slang%s       C++%s\n", colorDim, colorReset, cfg.Std)
	fmt.Printf("  %stype%s       %s\n", colorDim, colorReset, cfg.Type)
	fmt.Printf("  %sbuild%s      %s %s(%s)%s\n", colorDim, colorReset, cfg.Build, colorDim, cfg.BuildVersion, colorReset)
	fmt.Printf("  %sversion%s    %s\n", colorDim, colorReset, cfg.Version)
	if cfg.Tests != "" {
		fmt.Printf("  %stests%s      %s %s(%s)%s\n", colorDim, colorReset, cfg.Tests, colorDim, cfg.TestsVersion, colorReset)
	}
	var extras []string
	if cfg.ClangFormat {
		extras = append(extras, "clang-format")
	}
	if cfg.ClangTidy {
		extras = append(extras, "clang-tidy")
	}
	if cfg.Nix {
		extras = append(extras, "nix flake")
	}
	if len(extras) > 0 {
		fmt.Printf("  %sextras%s     %s\n", colorDim, colorReset, strings.Join(extras, ", "))
	}
	fmt.Println()
}

func printCreated(verb, dir string, cfg ProjectConfig) {
	label := fmt.Sprintf("%s · C++%s · %s · %s", cfg.Name, cfg.Std, cfg.Type, cfg.Build)
	if cfg.Tests != "" {
		label += fmt.Sprintf(" · %s", cfg.Tests)
	}
	success(verb, label)

	buildFile := ""
	if bs, ok := template.BuildSystemByName(cfg.Build); ok {
		buildFile = bs.BuildFile
	}

	fmt.Printf("%s  %s/%s\n", colorDim, dir, buildFile)
	switch cfg.Type {
	case "exe":
		fmt.Printf("  %s/src/main.cpp\n", dir)
		fmt.Printf("  %s/include/%s/\n", dir, cfg.Name)
	case "lib-static", "lib-shared":
		fmt.Printf("  %s/src/%s.cpp\n", dir, cfg.Name)
		fmt.Printf("  %s/include/%s/%s.hpp\n", dir, cfg.Name, cfg.Name)
	case "lib-header":
		fmt.Printf("  %s/include/%s/%s.hpp\n", dir, cfg.Name, cfg.Name)
	}
	if cfg.Tests != "" {
		fmt.Printf("  %s/tests/test_main.cpp\n", dir)
	}
	if cfg.ClangFormat {
		fmt.Printf("  %s/.clang-format\n", dir)
	}
	if cfg.ClangTidy {
		fmt.Printf("  %s/.clang-tidy\n", dir)
	}
	if cfg.Nix {
		fmt.Printf("  %s/flake.nix\n", dir)
	}
	fmt.Printf("%s\n", colorReset)

	if dir != "." {
		fmt.Printf("  %s$%s cd %s%s%s\n", colorDim, colorReset, colorBold, dir, colorReset)
	}
	if cfg.Nix {
		fmt.Printf("  %s$%s nix develop\n", colorDim, colorReset)
	}
	fmt.Printf("  %s$%s cppup build\n", colorDim, colorReset)
	if cfg.Type == "exe" {
		fmt.Printf("  %s$%s cppup run\n", colorDim, colorReset)
	}
}

func runGenericScaffold(cmd *cobra.Command, args []string, isNew bool) {
	defaultName := ""
	if isNew {
		if len(args) > 0 {
			defaultName = args[0]
		}
	} else {
		defaultName = cwdName()
		if len(args) > 0 {
			defaultName = args[0]
		}
	}

	if !isTTY() {
		name := defaultName
		if isNew && name == "" {
			fmt.Fprintf(os.Stderr, "%s✗%s project name required (pass as argument)\n", colorBrightRed, colorReset)
			os.Exit(1)
		}
		cfg, err := configFromFlags(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		if isNew {
			if err := createProject(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
				os.Exit(1)
			}
			printCreated("Created", cfg.Name, cfg)
		} else {
			if err := initProject(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
				os.Exit(1)
			}
			printCreated("Initialized", ".", cfg)
		}
		return
	}

	rl, err := readline.NewEx(&readline.Config{
		AutoComplete: &completer{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}
	defer rl.Close()

	cfg, err := collectSettings(rl, defaultName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	printSummary(cfg)

	action := "Create project"
	if !isNew {
		action = "Initialize project"
	}
	confirm, err := promptBool(rl, action, true)
	if err != nil || !confirm {
		fmt.Printf("%sCancelled%s\n", colorDim, colorReset)
		return
	}

	if isNew {
		if err := createProject(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		printCreated("Created", cfg.Name, cfg)
	} else {
		if err := initProject(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		printCreated("Initialized", ".", cfg)
	}
}

func runNew(cmd *cobra.Command, args []string) {
	runGenericScaffold(cmd, args, true)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func validateSettings(cfg ProjectConfig) error {
	if !contains(validTypes, cfg.Type) {
		return fmt.Errorf("invalid project type %q: choose %s", cfg.Type, join(validTypes))
	}
	if !contains(validBuilds, cfg.Build) {
		return fmt.Errorf("invalid build system %q: choose %s", cfg.Build, join(validBuilds))
	}
	if !contains(validStds, cfg.Std) {
		return fmt.Errorf("invalid standard %q: choose %s", cfg.Std, join(validStds))
	}
	if cfg.Tests != "" && !contains(validFrameworks, cfg.Tests) {
		return fmt.Errorf("invalid test framework %q: choose %s", cfg.Tests, join(validFrameworks))
	}
	return nil
}

func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	displayPath := path
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, path); err == nil {
			displayPath = rel
		}
	}
	fmt.Printf("  %s%screating%s %s\n", colorBold, colorBrightBlue, colorReset, displayPath)
	return os.WriteFile(path, []byte(content), 0o644)
}

func scaffoldFiles(root string, cfg ProjectConfig) error {
	step("Scaffolding", cfg.Name)

	withTests := cfg.Tests != ""

	if err := scaffoldBuildSystem(root, cfg, withTests); err != nil {
		return err
	}
	if err := scaffoldSources(root, cfg); err != nil {
		return err
	}
	if err := scaffoldTests(root, cfg, withTests); err != nil {
		return err
	}
	if err := scaffoldExtras(root, cfg); err != nil {
		return err
	}

	meta := ProjectMeta{
		Name:          cfg.Name,
		Type:          cfg.Type,
		BuildSystem:   cfg.Build,
		Version:       cfg.Version,
		Std:           cfg.Std,
		TestFramework: cfg.Tests,
	}
	return writeMetaTo(filepath.Join(root, metaFile), meta)
}

func scaffoldBuildSystem(root string, cfg ProjectConfig, withTests bool) error {
	var err error
	switch cfg.Build {
	case "cmake":
		content := ""
		if cfg.Type == "lib-header" {
			content, err = template.CMakeListsHeaderOnly(cfg.Name, cfg.Version, cfg.BuildVersion, withTests)
		} else if cfg.Type == "exe" {
			content, err = template.CMakeListsExe(cfg.Name, cfg.Version, cfg.Std, cfg.BuildVersion, withTests)
		} else {
			libType := "STATIC"
			if cfg.Type == "lib-shared" {
				libType = "SHARED"
			}
			content, err = template.CMakeListsLib(cfg.Name, cfg.Version, cfg.Std, libType, cfg.BuildVersion, withTests)
		}
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(root, "CMakeLists.txt"), content)
	case "meson":
		content := ""
		if cfg.Type == "lib-header" {
			content, err = template.MesonBuildHeaderOnly(cfg.Name, cfg.Version, cfg.Std, cfg.BuildVersion, withTests)
		} else if cfg.Type == "exe" {
			content, err = template.MesonBuildExe(cfg.Name, cfg.Version, cfg.Std, cfg.BuildVersion, withTests)
		} else {
			libType := "static"
			if cfg.Type == "lib-shared" {
				libType = "shared"
			}
			content, err = template.MesonBuildLib(cfg.Name, cfg.Version, cfg.Std, libType, cfg.BuildVersion, withTests)
		}
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(root, "meson.build"), content)
	}
	return nil
}

func scaffoldSources(root string, cfg ProjectConfig) error {
	switch cfg.Type {
	case "exe":
		content, err := template.MainCpp()
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(root, dirSrc, "main.cpp"), content)
	case "lib-static", "lib-shared":
		hpp, err := template.LibHpp(cfg.Name)
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, dirInclude, cfg.Name, cfg.Name+".hpp"), hpp); err != nil {
			return err
		}
		cpp, err := template.LibCpp(cfg.Name)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(root, dirSrc, cfg.Name+".cpp"), cpp)
	case "lib-header":
		hpp, err := template.HeaderOnlyHpp(cfg.Name)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(root, dirInclude, cfg.Name, cfg.Name+".hpp"), hpp)
	}
	return nil
}

func scaffoldTests(root string, cfg ProjectConfig, withTests bool) error {
	if !withTests {
		return nil
	}
	testMain, err := template.TestSourceCpp(cfg.Tests)
	if err != nil {
		return err
	}
	if err := writeFile(filepath.Join(root, dirTests, "test_main.cpp"), testMain); err != nil {
		return err
	}
	switch cfg.Build {
	case "cmake":
		content, err := template.TestCMakeLists(cfg.Name, cfg.Type, cfg.Tests, cfg.TestsVersion)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(root, dirTests, "CMakeLists.txt"), content)
	case "meson":
		content, err := template.TestMesonBuild(cfg.Name, cfg.Type, cfg.Tests)
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, dirTests, "meson.build"), content); err != nil {
			return err
		}
		wrap, err := template.MesonWrap(cfg.Tests, cfg.TestsVersion)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(root, dirSubproj, template.MesonWrapFilename(cfg.Tests)), wrap)
	}
	return nil
}

func scaffoldExtras(root string, cfg ProjectConfig) error {
	if !cfg.NoGit {
		ignore, err := template.Gitignore(cfg.Nix, cfg.Build == "meson")
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, ".gitignore"), ignore); err != nil {
			return err
		}
	}

	if cfg.ClangFormat {
		content, err := template.ClangFormat()
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, ".clang-format"), content); err != nil {
			return err
		}
	}
	if cfg.ClangTidy {
		content, err := template.ClangTidy()
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, ".clang-tidy"), content); err != nil {
			return err
		}
	}
	if cfg.Nix {
		flake, err := template.FlakeNix(cfg.Name, cfg.Build, cfg.Version, cfg.Tests)
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, "flake.nix"), flake); err != nil {
			return err
		}
		treefmt, err := template.TreefmtNix()
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, "treefmt.nix"), treefmt); err != nil {
			return err
		}
		envrc, err := template.Envrc()
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, ".envrc"), envrc); err != nil {
			return err
		}
	}
	return nil
}

func createProject(cfg ProjectConfig) error {
	if _, err := os.Stat(cfg.Name); !os.IsNotExist(err) {
		return fmt.Errorf("directory %q already exists", cfg.Name)
	}
	if err := scaffoldFiles(cfg.Name, cfg); err != nil {
		return err
	}
	if cfg.NoGit {
		return nil
	}
	gitCmd := exec.Command("git", "init", "-b", "main", "-q", cfg.Name)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	return gitCmd.Run()
}

func initProject(cfg ProjectConfig) error {
	if _, err := os.Stat(metaFile); err == nil {
		return fmt.Errorf("project already initialized: %q exists", metaFile)
	}
	if err := scaffoldFiles(".", cfg); err != nil {
		return err
	}
	if cfg.NoGit {
		return nil
	}
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		gitCmd := exec.Command("git", "init", "-b", "main", "-q")
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr
		return gitCmd.Run()
	}
	return nil
}

func writeMetaTo(path string, m ProjectMeta) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
