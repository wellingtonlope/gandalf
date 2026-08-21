package generator

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const (
	// LowerLetters contains lowercase ASCII letters.
	LowerLetters = "abcdefghijklmnopqrstuvwxyz"
	// UpperLetters contains uppercase ASCII letters.
	UpperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Digits contains decimal digits.
	Digits = "0123456789"
	// Symbols contains commonly accepted password symbols.
	Symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?/~`"
)

// Options defines the parameters for password generation.
type Options struct {
	Length     int
	UseLower   bool
	UseUpper   bool
	UseDigits  bool
	UseSymbols bool
}

// DefaultOptions returns sensible defaults: 16 chars with all charsets enabled.
func DefaultOptions() Options {
	return Options{
		Length:     16,
		UseLower:   true,
		UseUpper:   true,
		UseDigits:  true,
		UseSymbols: true,
	}
}

// Charset builds the character set based on enabled options.
func (o Options) Charset() string {
	cs := ""
	if o.UseLower {
		cs += LowerLetters
	}
	if o.UseUpper {
		cs += UpperLetters
	}
	if o.UseDigits {
		cs += Digits
	}
	if o.UseSymbols {
		cs += Symbols
	}
	return cs
}

// requiredCount returns how many character sets are enabled.
func (o Options) requiredCount() int {
	n := 0
	if o.UseLower {
		n++
	}
	if o.UseUpper {
		n++
	}
	if o.UseDigits {
		n++
	}
	if o.UseSymbols {
		n++
	}
	return n
}

// Validate checks if options are valid.
func (o Options) Validate() error {
	if o.Length <= 0 {
		return errors.New("length must be greater than zero")
	}
	if o.Length > 256 {
		return errors.New("maximum length is 256")
	}
	if o.Charset() == "" {
		return errors.New("at least one character set must be enabled")
	}
	if o.Length < o.requiredCount() {
		return fmt.Errorf("length must be at least %d to include all selected character sets", o.requiredCount())
	}
	return nil
}

// Generate creates a cryptographically secure random password using crypto/rand.
// It guarantees that each enabled character set appears at least once, then shuffles
// the result with a crypto-secure Fisher-Yates shuffle. This avoids repetitive
// or missing character classes (e.g., passwords without digits).
func Generate(opts Options) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}
	charset := opts.Charset()

	// Pick one guaranteed character from each enabled set to avoid missing classes.
	required := make([]byte, 0, 4)
	if opts.UseLower {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(LowerLetters))))
		if err != nil {
			return "", err
		}
		required = append(required, LowerLetters[n.Int64()])
	}
	if opts.UseUpper {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(UpperLetters))))
		if err != nil {
			return "", err
		}
		required = append(required, UpperLetters[n.Int64()])
	}
	if opts.UseDigits {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(Digits))))
		if err != nil {
			return "", err
		}
		required = append(required, Digits[n.Int64()])
	}
	if opts.UseSymbols {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(Symbols))))
		if err != nil {
			return "", err
		}
		required = append(required, Symbols[n.Int64()])
	}

	// Fill the remaining length with random chars from the full charset.
	remaining := opts.Length - len(required)
	result := make([]byte, 0, opts.Length)
	result = append(result, required...)

	max := big.NewInt(int64(len(charset)))
	for i := 0; i < remaining; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result = append(result, charset[n.Int64()])
	}

	// Shuffle with crypto-secure Fisher-Yates to avoid positional bias.
	for i := len(result) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := int(jBig.Int64())
		result[i], result[j] = result[j], result[i]
	}

	return string(result), nil
}

// GenerateMultiple generates N passwords with the same options.
func GenerateMultiple(opts Options, count int) ([]string, error) {
	if count <= 0 {
		return nil, errors.New("count must be greater than zero")
	}
	out := make([]string, count)
	for i := range out {
		p, err := Generate(opts)
		if err != nil {
			return nil, err
		}
		out[i] = p
	}
	return out, nil
}
