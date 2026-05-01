package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	depGit  string
	depTag  string
	depURL  string
	depName string
)

var addDepCmd = &cobra.Command{
	Use:   "dep [name]",
	Short: "Add a dependency",
	Long:  "Add a dependency via git, archive URL, or meson WrapDB.\n\nExamples:\n  cppup add dep --git https://github.com/fmtlib/fmt.git --tag v11.2.0\n  cppup add dep --url https://example.com/fmt-11.0.2.tar.gz --name fmt\n  cppup add dep fmt  (meson only, uses WrapDB)",
	Args:  cobra.MaximumNArgs(1),
	Run:   runAddDep,
}

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "List project dependencies",
	Run:   runDeps,
}

var removeCmd = &cobra.Command{
	Use:   "remove <component>",
	Short: "Remove a component from the project",
}

var removeDepCmd = &cobra.Command{
	Use:   "dep <name>",
	Short: "Remove a dependency",
	Args:  cobra.ExactArgs(1),
	Run:   runRemoveDep,
}

func init() {
	addDepCmd.Flags().StringVar(&depGit, "git", "", "Git repository URL")
	addDepCmd.Flags().StringVar(&depTag, "tag", "", "Git tag or version (required with --git)")
	addDepCmd.Flags().StringVar(&depURL, "url", "", "Archive URL")
	addDepCmd.Flags().StringVar(&depName, "name", "", "Dependency name (required with --url)")
	addCmd.AddCommand(addDepCmd)
	removeCmd.AddCommand(removeDepCmd)
}

func nameFromGitURL(url string) string {
	base := filepath.Base(url)
	return strings.TrimSuffix(base, ".git")
}

func runAddDep(cmd *cobra.Command, args []string) {
	wrapDB := len(args) == 1 && depGit == "" && depURL == ""

	if !wrapDB {
		if depGit == "" && depURL == "" {
			fmt.Fprintf(os.Stderr, "%s✗%s specify --git, --url, or a package name (meson WrapDB)\n", colorBrightRed, colorReset)
			os.Exit(1)
		}
		if depGit != "" && depURL != "" {
			fmt.Fprintf(os.Stderr, "%s✗%s use --git or --url, not both\n", colorBrightRed, colorReset)
			os.Exit(1)
		}
		if depGit != "" && depTag == "" {
			fmt.Fprintf(os.Stderr, "%s✗%s --tag is required with --git\n", colorBrightRed, colorReset)
			os.Exit(1)
		}
		if depURL != "" && depName == "" {
			fmt.Fprintf(os.Stderr, "%s✗%s --name is required with --url\n", colorBrightRed, colorReset)
			os.Exit(1)
		}
	}

	if err := findProjectRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	meta, err := readMeta()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s not a cppup project (missing %s)\n", colorBrightRed, colorReset, metaFile)
		os.Exit(1)
	}

	if wrapDB {
		if meta.BuildSystem != "meson" {
			fmt.Fprintf(os.Stderr, "%s✗%s WrapDB is only available for meson projects, use --git or --url\n", colorBrightRed, colorReset)
			os.Exit(1)
		}
		name := args[0]
		for _, d := range meta.Deps {
			if d.Name == name {
				fmt.Fprintf(os.Stderr, "%s✗%s dependency %q already exists\n", colorBrightRed, colorReset, name)
				os.Exit(1)
			}
		}
		if err := execCommand("meson", "wrap", "install", name); err != nil {
			fail("Failed to install wrap")
			os.Exit(1)
		}
		meta.Deps = append(meta.Deps, Dep{Name: name})
		if err := writeMetaTo(metaFile, meta); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		appendMesonSubprojectsToGitignore()
		success("Added", name+" (WrapDB)")
		fmt.Printf("  %s- version managed in subprojects/%s.wrap%s\n", colorDim, name, colorReset)
		fmt.Printf("  %s- manually link the dependency in meson.build to use it%s\n", colorDim, colorReset)
		return
	}

	var name string
	if depGit != "" {
		name = nameFromGitURL(depGit)
	} else {
		name = depName
	}

	for _, d := range meta.Deps {
		if d.Name == name {
			fmt.Fprintf(os.Stderr, "%s✗%s dependency %q already exists\n", colorBrightRed, colorReset, name)
			os.Exit(1)
		}
	}

	dep := Dep{Name: name, Git: depGit, Tag: depTag, URL: depURL}

	if depURL != "" {
		step("Fetching", depURL)
		hash, err := fetchSHA256(depURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s failed to fetch: %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}
		dep.Hash = hash
	}

	meta.Deps = append(meta.Deps, dep)

	if err := syncDeps(meta); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	if err := writeMetaTo(metaFile, meta); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	label := name
	if dep.Tag != "" {
		label += " " + dep.Tag
	}
	success("Added", label)

	switch meta.BuildSystem {
	case "cmake":
		fmt.Printf("  %snote: manually add target_link_libraries() in CMakeLists.txt, refer to the library's docs for usage%s\n", colorDim, colorReset)
	case "meson":
		appendMesonSubprojectsToGitignore()
		fmt.Printf("  %snote: manually link this dependency in meson.build, refer to the library's docs for usage%s\n", colorDim, colorReset)
	}
}

func runDeps(cmd *cobra.Command, args []string) {
	if err := findProjectRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	meta, err := readMeta()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s not a cppup project (missing %s)\n", colorBrightRed, colorReset, metaFile)
		os.Exit(1)
	}

	if len(meta.Deps) == 0 {
		fmt.Printf("%sno dependencies%s\n", colorDim, colorReset)
		return
	}

	for _, d := range meta.Deps {
		fmt.Printf("  %s%s%s %s%s%s\n",
			colorBold, d.Name, colorReset,
			colorDim, d.Info(), colorReset)
	}
}

func runRemoveDep(cmd *cobra.Command, args []string) {
	if err := findProjectRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	meta, err := readMeta()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s not a cppup project (missing %s)\n", colorBrightRed, colorReset, metaFile)
		os.Exit(1)
	}

	name := args[0]
	found := false
	var newDeps []Dep
	for _, d := range meta.Deps {
		if d.Name == name {
			found = true
		} else {
			newDeps = append(newDeps, d)
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "%s✗%s dependency %q not found\n", colorBrightRed, colorReset, name)
		os.Exit(1)
	}

	if meta.BuildSystem == "meson" {
		os.Remove(filepath.Join(dirSubproj, name+".wrap"))
	}

	meta.Deps = newDeps

	if err := syncDeps(meta); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	if err := writeMetaTo(metaFile, meta); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
		os.Exit(1)
	}

	success("Removed", name)
}

// syncDeps regenerates dependency files from the current deps list.
func syncDeps(meta ProjectMeta) error {
	switch meta.BuildSystem {
	case "cmake":
		return syncCMakeDeps(meta.Deps)
	case "meson":
		return syncMesonDeps(meta.Deps)
	}
	return nil
}

func syncCMakeDeps(deps []Dep) error {
	path := filepath.Join("cmake", "deps.cmake")

	if len(deps) == 0 {
		os.Remove(path)
		removeCMakeInclude()
		return nil
	}

	if err := os.MkdirAll("cmake", 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(generateCMakeDeps(deps)), 0o644); err != nil {
		return err
	}

	return ensureCMakeInclude()
}

func generateCMakeDeps(deps []Dep) string {
	var b strings.Builder
	b.WriteString("include(FetchContent)\n")
	for _, d := range deps {
		b.WriteString("\nFetchContent_Declare(\n")
		fmt.Fprintf(&b, "  %s\n", d.Name)
		if d.Git != "" {
			fmt.Fprintf(&b, "  GIT_REPOSITORY %s\n", d.Git)
			fmt.Fprintf(&b, "  GIT_TAG        %s\n", d.Tag)
			b.WriteString("  GIT_SHALLOW    TRUE\n")
		} else {
			fmt.Fprintf(&b, "  URL %s\n", d.URL)
			if d.Hash != "" {
				fmt.Fprintf(&b, "  URL_HASH SHA256=%s\n", d.Hash)
			}
		}
		b.WriteString(")\n")
		fmt.Fprintf(&b, "FetchContent_MakeAvailable(%s)\n", d.Name)
	}
	return b.String()
}

const cmakeDepInclude = "include(cmake/deps.cmake)"

func ensureCMakeInclude() error {
	data, err := os.ReadFile("CMakeLists.txt")
	if err != nil {
		return err
	}
	if strings.Contains(string(data), cmakeDepInclude) {
		return nil
	}
	f, err := os.OpenFile("CMakeLists.txt", os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n" + cmakeDepInclude + "\n")
	return err
}

func removeCMakeInclude() {
	data, err := os.ReadFile("CMakeLists.txt")
	if err != nil {
		return
	}
	content := string(data)
	newContent := strings.Replace(content, "\n"+cmakeDepInclude+"\n", "\n", 1)
	if newContent != content {
		_ = os.WriteFile("CMakeLists.txt", []byte(newContent), 0o644)
	}
}

func syncMesonDeps(deps []Dep) error {
	if len(deps) == 0 {
		return nil
	}
	if err := os.MkdirAll(dirSubproj, 0o755); err != nil {
		return err
	}
	for _, d := range deps {
		path := filepath.Join(dirSubproj, d.Name+".wrap")
		if err := os.WriteFile(path, []byte(generateMesonWrap(d)), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func generateMesonWrap(dep Dep) string {
	if dep.Git != "" {
		return fmt.Sprintf("[wrap-git]\nurl = %s\nrevision = %s\ndepth = 1\n", dep.Git, dep.Tag)
	}
	return fmt.Sprintf("[wrap-file]\nsource_url = %s\nsource_filename = %s\nsource_hash = %s\n",
		dep.URL, filepath.Base(dep.URL), dep.Hash)
}

func fetchSHA256(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	h := sha256.New()
	if _, err := io.Copy(h, resp.Body); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
