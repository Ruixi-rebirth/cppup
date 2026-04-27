package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ruixi-rebirth/cppup/template"
)

// BuildMode captures the three mutually-exclusive build profiles.
type BuildMode struct {
	Release bool
	Asan    bool
}

func (m BuildMode) label() string {
	if m.Asan {
		return "asan"
	}
	if m.Release {
		return "release"
	}
	return "debug"
}

func (m BuildMode) dir() string {
	return filepath.Join(dirBuild, m.label())
}

// wantCompileCommands reports whether this mode should update the
// compile_commands.json symlink (all non-release builds).
func (m BuildMode) wantCompileCommands() bool {
	return !m.Release || m.Asan
}

func findProjectRoot() error {
	for {
		for _, bs := range template.BuildSystems {
			if _, err := os.Stat(bs.BuildFile); err == nil {
				return nil
			}
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			var files []string
			for _, bs := range template.BuildSystems {
				files = append(files, bs.BuildFile)
			}
			return fmt.Errorf("no %s found (searched up from %s)", strings.Join(files, " or "), cwd)
		}
		if err := os.Chdir(parent); err != nil {
			return err
		}
	}
}

func detectBuildSystem() (string, error) {
	for _, bs := range template.BuildSystems {
		if _, err := os.Stat(bs.BuildFile); err == nil {
			return bs.Name, nil
		}
	}
	var files []string
	for _, bs := range template.BuildSystems {
		files = append(files, bs.BuildFile)
	}
	return "", fmt.Errorf("no %s found", strings.Join(files, " or "))
}

func hasTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func hasNinja() bool {
	return hasTool("ninja")
}

func resolveProjectName() string {
	if m, err := readMeta(); err == nil && m.Name != "" {
		return m.Name
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}

func execCommand(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func buildProject(mode BuildMode) (string, error) {
	if err := findProjectRoot(); err != nil {
		return "", err
	}
	bs, err := detectBuildSystem()
	if err != nil {
		return "", err
	}
	dir := mode.dir()
	switch bs {
	case "cmake":
		return bs, buildCMake(dir, mode)
	case "meson":
		return bs, buildMeson(dir, mode)
	}
	return bs, nil
}

func buildCMake(dir string, mode BuildMode) error {
	cmakeBuildType := "Debug"
	if mode.Release && !mode.Asan {
		cmakeBuildType = "Release"
	}

	if _, err := os.Stat(filepath.Join(dir, "CMakeCache.txt")); os.IsNotExist(err) {
		generator := "make"
		configArgs := []string{"-B", dir, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", cmakeBuildType), "-DCMAKE_POLICY_VERSION_MINIMUM=3.10"}
		if mode.wantCompileCommands() {
			configArgs = append(configArgs, "-DCMAKE_EXPORT_COMPILE_COMMANDS=ON")
		}
		if hasNinja() {
			configArgs = append(configArgs, "-G", "Ninja")
			generator = "ninja"
		}
		if mode.Asan {
			sanFlags := "-fsanitize=address,undefined -fno-omit-frame-pointer"
			configArgs = append(configArgs,
				"-DCMAKE_C_FLAGS="+sanFlags,
				"-DCMAKE_CXX_FLAGS="+sanFlags,
				"-DCMAKE_EXE_LINKER_FLAGS=-fsanitize=address,undefined",
			)
		}
		step("Configuring", fmt.Sprintf("cmake · %s · %s", mode.label(), generator))
		if err := execCommand("cmake", configArgs...); err != nil {
			fail("Configure failed")
			return fmt.Errorf("cmake configure: %w", err)
		}
	}

	step("Building", mode.label())
	start := time.Now()
	if err := execCommand("cmake", "--build", dir, "--parallel"); err != nil {
		fail("Build failed")
		return fmt.Errorf("cmake build: %w", err)
	}

	if mode.wantCompileCommands() {
		symlinkCompileCommands(dir)
	}

	artifact := dir
	if m, err := readMeta(); err == nil && m.Type == "exe" {
		artifact = filepath.Join(dir, resolveProjectName())
	}
	success("Finished", fmt.Sprintf("%s  (%.2fs)", artifact, time.Since(start).Seconds()))
	return nil
}

func buildMeson(dir string, mode BuildMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		buildType := mode.label()
		if mode.Asan {
			buildType = "debug"
		}
		setupArgs := []string{"setup", dir, "--buildtype=" + buildType}
		if mode.Asan {
			setupArgs = append(setupArgs, "-Db_sanitize=address,undefined")
		}
		step("Configuring", fmt.Sprintf("meson · %s", mode.label()))
		if err := execCommand("meson", setupArgs...); err != nil {
			fail("Configure failed")
			return fmt.Errorf("meson setup: %w", err)
		}
	}

	step("Building", mode.label())
	start := time.Now()
	if err := execCommand("meson", "compile", "-C", dir); err != nil {
		fail("Build failed")
		return fmt.Errorf("meson compile: %w", err)
	}

	if mode.wantCompileCommands() {
		symlinkCompileCommands(dir)
	}

	artifact := dir
	if m, err := readMeta(); err == nil && m.Type == "exe" {
		artifact = filepath.Join(dir, resolveProjectName())
	}
	success("Finished", fmt.Sprintf("%s  (%.2fs)", artifact, time.Since(start).Seconds()))
	return nil
}

func symlinkCompileCommands(dir string) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	src := filepath.Join(cwd, dir, "compile_commands.json")
	dst := filepath.Join(cwd, "compile_commands.json")
	_ = os.Remove(dst)
	_ = os.Symlink(src, dst)
}
