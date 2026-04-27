package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [prefix]",
	Short: "Build (release) and install the project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := findProjectRoot(); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		prefix := "/usr/local"
		if len(args) > 0 {
			prefix = args[0]
		}

		if err := checkWritable(prefix); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		mode := BuildMode{Release: true}
		bs, err := buildProject(mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %v\n", colorBrightRed, colorReset, err)
			os.Exit(1)
		}

		dir := mode.dir()
		step("Installing", prefix)

		var installErr error
		switch bs {
		case "cmake":
			installErr = execCommand("cmake", "--install", dir, "--prefix", prefix)
		case "meson":
			installErr = mesonInstall(dir, prefix)
		}

		if installErr != nil {
			fail("Install failed")
			os.Exit(1)
		}
		success("Installed", prefix)
	},
}

func checkWritable(prefix string) error {
	if err := os.MkdirAll(prefix, 0o755); err != nil {
		if prefix == "/usr/local" {
			return fmt.Errorf("cannot write to %s\n  try: sudo cppup install\n   or: cppup install ~/.local", prefix)
		}
		return fmt.Errorf("cannot write to %s", prefix)
	}
	f, err := os.CreateTemp(prefix, ".cppup-write-test-*")
	if err != nil {
		if prefix == "/usr/local" {
			return fmt.Errorf("cannot write to %s\n  try: sudo cppup install\n   or: cppup install ~/.local", prefix)
		}
		return fmt.Errorf("cannot write to %s", prefix)
	}
	os.Remove(f.Name())
	return nil
}

func mesonInstall(buildDir, prefix string) error {
	step("Reconfiguring", fmt.Sprintf("prefix → %s", prefix))
	if err := execCommand("meson", "setup", "--reconfigure", "--prefix", prefix, buildDir); err != nil {
		return err
	}
	if err := execCommand("meson", "compile", "-C", buildDir); err != nil {
		return err
	}
	return execCommand("meson", "install", "-C", buildDir)
}
