package generator

import (
	"crypto/rand"
	"errors"
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
	return nil
}

// Generate creates a cryptographically secure random password using crypto/rand.
func Generate(opts Options) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}
	charset := opts.Charset()
	result := make([]byte, opts.Length)
	max := big.NewInt(int64(len(charset)))
	for i := 0; i < opts.Length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
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
