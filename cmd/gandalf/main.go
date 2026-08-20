package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wellingtonlope/gandalf/internal/generator"
	"github.com/wellingtonlope/gandalf/internal/mellon"
)

// version is injected at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		arg := strings.ToLower(strings.TrimSpace(os.Args[1]))
		if mellon.IsMellon(arg) {
			mellon.Reveal()
			return
		}
		if arg == "help" || arg == "--help" || arg == "-h" {
			printUsage()
			return
		}
		if arg == "version" || arg == "--version" || arg == "-v" {
			fmt.Printf("gandalf %s - A wizard is never late, nor is he early. He generates the password precisely when he means to.\n", version)
			return
		}
	}

	opts := generator.DefaultOptions()
	var count int
	var showHelp bool

	fs := flag.NewFlagSet("gandalf", flag.ExitOnError)
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

	noUpper := fs.Bool("no-upper", false, "exclude uppercase letters")
	noLower := fs.Bool("no-lower", false, "exclude lowercase letters")
	noDigits := fs.Bool("no-digits", false, "exclude digits")
	noSymbols := fs.Bool("no-symbols", false, "exclude symbols")

	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	if showHelp {
		printUsage()
		return
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: try 'gandalf --help' or say 'mellon' 😉\n")
		os.Exit(1)
	}

	for _, p := range passwords {
		fmt.Println(p)
	}
}

func printUsage() {
	fmt.Println(`gandalf - Magical password generator 🧙‍♂️

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
  -h, --help           show this help
  -v, --version        show version

Examples:
  gandalf                         # 16-char password
  gandalf -l 24                   # 24-char password
  gandalf -l 32 --no-symbols      # without symbols
  gandalf -c 5 -l 12              # 5 passwords of 12 chars
  gandalf mellon                  # 👀

"A wizard is never late, nor is he early. He generates the password precisely when he means to."`)
}
