package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wellingtonlope/gandalf/internal/clipboard"
	"github.com/wellingtonlope/gandalf/internal/generator"
	"github.com/wellingtonlope/gandalf/internal/lore"
	"github.com/wellingtonlope/gandalf/internal/mellon"
)

// version is injected at build time via -ldflags "-X main.version=..."
var version = "dev"

// inputReader is the source for reading passwords in check mode, overrideable in tests.
var inputReader io.Reader = os.Stdin

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
		if arg == "check" {
			return handleCheck(args[1:], stdout, stderr)
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
	fs.BoolVar(&copyToClipboard, "copy", false, "copy passwords to clipboard (one per line)")
	fs.BoolVar(&copyToClipboard, "x", false, "copy passwords to clipboard (shorthand, one per line)")

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
		text := strings.Join(passwords, "\n")
		if err := copyFn(text); err != nil {
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

func handleCheck(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var showHelp bool
	fs.BoolVar(&showHelp, "help", false, "show help")
	fs.BoolVar(&showHelp, "h", false, "show help (shorthand)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if showHelp {
		printCheckUsage(stdout)
		return 0
	}

	var password string
	if fs.NArg() > 0 {
		if fs.NArg() > 1 {
			password = strings.Join(fs.Args(), " ")
		} else {
			password = fs.Arg(0)
		}
	} else {
		data, err := io.ReadAll(inputReader)
		if err != nil {
			fmt.Fprintf(stderr, "Error: failed to read password: %v\n", err)
			return 1
		}
		password = strings.TrimSpace(string(data))
		if password == "" {
			fmt.Fprintln(stderr, "Error: no password provided")
			fmt.Fprintln(stderr, "Usage: gandalf check <password> or echo <password> | gandalf check")
			printCheckUsage(stderr)
			return 1
		}
	}

	res := lore.Evaluate(password)

	emoji := rankEmoji(res.Rank.Name)
	fmt.Fprintf(stdout, "%s %s (%s) - %.1f bits\n", emoji, res.Rank.Title, res.Rank.Name, res.Entropy)
	fmt.Fprintf(stdout, "   \"%s\"\n", res.Rank.Quote)
	fmt.Fprintf(stdout, "   Charset: %s (%d) | Length: %d | Crack: %s\n", res.CharsetDesc, res.CharsetSize, len(res.Password), res.CrackTime)
	if len(res.Warnings) > 0 {
		fmt.Fprintln(stdout, "   Warnings:")
		for _, w := range res.Warnings {
			fmt.Fprintf(stdout, "     ⚠️  %s\n", w)
		}
	} else {
		fmt.Fprintln(stdout, "   Warnings: none - well forged!")
	}
	fmt.Fprintf(stdout, "   Advice: %s\n", res.Rank.Advice)
	return 0
}

func rankEmoji(name string) string {
	switch name {
	case "Hobbit":
		return "🔴"
	case "Orc":
		return "🟠"
	case "Ranger":
		return "🟡"
	case "Elf":
		return "🟢"
	case "Dwarf":
		return "🔵"
	case "Wizard":
		return "🟣"
	case "Sauron":
		return "✨"
	default:
		return "•"
	}
}

func printCheckUsage(w io.Writer) {
	fmt.Fprintln(w, `gandalf check - Evaluate password strength with Middle-earth lore 🧙‍♂️

Usage:
  gandalf check <password>         check strength of a password
  echo <password> | gandalf check  read from stdin
  gandalf check --help             show this help

Ranks (weakest -> strongest):
  Hobbit (0-27 bits)    - Like butter scraped over too much bread
  Orc (28-35 bits)      - Weak as an Orc in sunlight
  Ranger (36-59 bits)   - A Ranger wouldn't trust the road
  Elf (60-79 bits)      - Keen as Elven sight
  Dwarf (80-99 bits)    - Hard as the Doors of Durin
  Wizard (100-127 bits) - You shall not pass!
  Sauron (128+ bits)    - One password to rule them all

Examples:
  gandalf check "password123"
  gandalf check "xK9\$m2PqL8@vB4!"
  echo "mySecret123" | gandalf check

Advice: pipe passwords securely, avoid shell history: echo -n "pass" | gandalf check`)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `gandalf - Magical password generator 🧙‍♂️

Usage:
  gandalf [options]               generate a password
  gandalf check <password>       evaluate password strength (Middle-earth ranks)
  gandalf mellon                 ✨ elven secret
  gandalf friend                 ✨ translation also works

Options:
  -l, --length <n>     password length (default 16)
  -c, --count <n>      how many passwords to generate (default 1)
      --no-upper       exclude uppercase letters
      --no-lower       exclude lowercase letters
      --no-digits      exclude digits
      --no-symbols     exclude symbols
  -x, --copy           copy passwords to clipboard (one per line)
  -h, --help           show this help
  -v, --version        show version

Examples:
  gandalf                         # 16-char password
  gandalf -l 24                   # 24-char password
  gandalf -l 32 --no-symbols      # without symbols
  gandalf -c 5 -l 12              # 5 passwords of 12 chars
  gandalf -x -l 16                # copy to clipboard (shorthand)
  gandalf --copy -l 16            # copy to clipboard
  gandalf check "myPassword123"    # evaluate strength (Hobbit->Sauron)
  echo "secret" | gandalf check    # via stdin
  gandalf mellon                  # 👀

"A wizard is never late, nor is he early. He generates the password precisely when he means to."`)
}
