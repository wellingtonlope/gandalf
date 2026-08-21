package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wellingtonlope/gandalf/internal/clipboard"
	"github.com/wellingtonlope/gandalf/internal/generator"
	"github.com/wellingtonlope/gandalf/internal/mellon"
)

// version is injected at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, clipboard.Copy))
}

// run executes the CLI logic.
// args are the command-line arguments without the program name.
// stdout and stderr are used for output.
// copyFn is called when --copy is requested; if nil, clipboard is ignored.
// It returns an exit code (0 for success, 1 for error).
func run(args []string, stdout, stderr io.Writer, copyFn func(string) error) int {
	if len(args) > 0 {
		arg := strings.ToLower(strings.TrimSpace(args[0]))
		if mellon.IsMellon(arg) {
			mellon.Reveal()
			return 0
		}
		if arg == "help" || arg == "--help" || arg == "-h" {
			printUsage(stdout)
			return 0
		}
		if arg == "version" || arg == "--version" || arg == "-v" {
			fmt.Fprintf(stdout, "gandalf %s - A wizard is never late, nor is he early. He generates the password precisely when he means to.\n", version)
			return 0
		}
	}

	opts := generator.DefaultOptions()
	var count int
	var showHelp bool
	var copyToClipboard bool

	fs := flag.NewFlagSet("gandalf", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.IntVar(&opts.Length, "length", opts.Length, "password length")
	fs.IntVar(&opts.Length, "l", opts.Length, "password length (shorthand)")
	fs.BoolVar(&opts.UseLower, "lower", true, "include lowercase letters")
	fs.BoolVar(&opts.UseUpper, "upper", true, "include uppercase letters")
	fs.BoolVar(&opts.UseDigits, "digits", true, "include digits")
	fs.BoolVar(&opts.UseSymbols, "symbols", true, "include symbols")
	fs.IntVar(&count, "count", 1, "number of passwords to generate")
	fs.IntVar(&count, "c", 1, "number of passwords (shorthand)")
	fs.BoolVar(&showHelp, "help", false, "show help")
	fs.BoolVar(&showHelp, "h", false, "show help (shorthand)")
	fs.BoolVar(&copyToClipboard, "copy", false, "copy first password to clipboard")
	fs.BoolVar(&copyToClipboard, "x", false, "copy first password to clipboard (shorthand)")

	noUpper := fs.Bool("no-upper", false, "exclude uppercase letters")
	noLower := fs.Bool("no-lower", false, "exclude lowercase letters")
	noDigits := fs.Bool("no-digits", false, "exclude digits")
	noSymbols := fs.Bool("no-symbols", false, "exclude symbols")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if showHelp {
		printUsage(stdout)
		return 0
	}

	if *noUpper {
		opts.UseUpper = false
	}
	if *noLower {
		opts.UseLower = false
	}
	if *noDigits {
		opts.UseDigits = false
	}
	if *noSymbols {
		opts.UseSymbols = false
	}

	passwords, err := generator.GenerateMultiple(opts, count)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		fmt.Fprintf(stderr, "Hint: try 'gandalf --help' or say 'mellon' 😉\n")
		return 1
	}

	if copyToClipboard && len(passwords) > 0 && copyFn != nil {
		if err := copyFn(passwords[0]); err != nil {
			fmt.Fprintf(stderr, "Warning: could not copy to clipboard: %v\n", err)
		} else {
			fmt.Fprintln(stderr, "Copied to clipboard")
		}
	}

	for _, p := range passwords {
		fmt.Fprintln(stdout, p)
	}
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `gandalf - Magical password generator 🧙‍♂️

Usage:
  gandalf [options]               generate a password
  gandalf mellon                 ✨ elven secret
  gandalf friend                 ✨ translation also works

Options:
  -l, --length <n>     password length (default 16)
  -c, --count <n>      how many passwords to generate (default 1)
      --no-upper       exclude uppercase letters
      --no-lower       exclude lowercase letters
      --no-digits      exclude digits
      --no-symbols     exclude symbols
  -x, --copy           copy first password to clipboard
  -h, --help           show this help
  -v, --version        show version

Examples:
  gandalf                         # 16-char password
  gandalf -l 24                   # 24-char password
  gandalf -l 32 --no-symbols      # without symbols
  gandalf -c 5 -l 12              # 5 passwords of 12 chars
  gandalf -x -l 16                # copy to clipboard (shorthand)
  gandalf --copy -l 16            # copy to clipboard
  gandalf mellon                  # 👀

"A wizard is never late, nor is he early. He generates the password precisely when he means to."`)
}
