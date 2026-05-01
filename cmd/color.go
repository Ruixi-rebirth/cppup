package cmd

import (
	"fmt"
	"os"
)

var noColor = os.Getenv("NO_COLOR") != ""

func c(code string) string {
	if noColor {
		return ""
	}
	return code
}

var (
	colorReset       = c("\033[0m")
	colorBold        = c("\033[1m")
	colorDim         = c("\033[2m")
	colorBrightGreen = c("\033[92m")
	colorBrightCyan  = c("\033[96m")
	colorBrightRed   = c("\033[91m")
	colorBrightBlue  = c("\033[94m")
)

func step(label, detail string) {
	fmt.Printf("%s%s→%s %-14s%s%s%s\n",
		colorBold, colorBrightBlue, colorReset,
		label,
		colorDim, detail, colorReset,
	)
}

func success(label, detail string) {
	fmt.Printf("%s%s✓%s %-14s%s%s%s\n",
		colorBold, colorBrightGreen, colorReset,
		label,
		colorBrightCyan, detail, colorReset,
	)
}

func fail(label string) {
	fmt.Printf("%s%s✗%s %s\n",
		colorBold, colorBrightRed, colorReset,
		label,
	)
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s✗%s %s\n", colorBrightRed, colorReset, fmt.Sprintf(format, a...))
	os.Exit(1)
}
