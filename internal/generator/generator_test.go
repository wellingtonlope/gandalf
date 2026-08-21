package generator

import (
	"errors"
	"io"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	assert.Equal(t, 16, opts.Length)
	assert.True(t, opts.UseLower)
	assert.True(t, opts.UseUpper)
	assert.True(t, opts.UseDigits)
	assert.True(t, opts.UseSymbols)

	expectedLen := len(LowerLetters) + len(UpperLetters) + len(Digits) + len(Symbols)
	assert.Len(t, opts.Charset(), expectedLen)
}

func TestOptions_Charset(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		expected string
	}{
		{name: "all enabled", opts: Options{UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}, expected: LowerLetters + UpperLetters + Digits + Symbols},
		{name: "only lower", opts: Options{UseLower: true}, expected: LowerLetters},
		{name: "only upper", opts: Options{UseUpper: true}, expected: UpperLetters},
		{name: "only digits", opts: Options{UseDigits: true}, expected: Digits},
		{name: "only symbols", opts: Options{UseSymbols: true}, expected: Symbols},
		{name: "lower and digits", opts: Options{UseLower: true, UseDigits: true}, expected: LowerLetters + Digits},
		{name: "upper and symbols", opts: Options{UseUpper: true, UseSymbols: true}, expected: UpperLetters + Symbols},
		{name: "none enabled", opts: Options{}, expected: ""},
		{name: "lower upper no digits no symbols", opts: Options{UseLower: true, UseUpper: true}, expected: LowerLetters + UpperLetters},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.opts.Charset())
		})
	}
}

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr string
	}{
		{name: "valid default", opts: DefaultOptions()},
		{name: "valid custom 32 lower only", opts: Options{Length: 32, UseLower: true}},
		{name: "valid length 1", opts: Options{Length: 1, UseLower: true}},
		{name: "valid length 256", opts: Options{Length: 256, UseLower: true}},
		{name: "invalid length zero", opts: Options{Length: 0, UseLower: true}, wantErr: "length must be greater than zero"},
		{name: "invalid length negative", opts: Options{Length: -5, UseLower: true}, wantErr: "length must be greater than zero"},
		{name: "invalid length too large", opts: Options{Length: 257, UseLower: true}, wantErr: "maximum length is 256"},
		{name: "invalid empty charset", opts: Options{Length: 16}, wantErr: "at least one character set must be enabled"},
		{name: "invalid zero length and empty charset - length error first", opts: Options{Length: 0}, wantErr: "length must be greater than zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{name: "default options 16 chars", opts: DefaultOptions()},
		{name: "length 1 lower only", opts: Options{Length: 1, UseLower: true}},
		{name: "length 32 all charsets", opts: Options{Length: 32, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}},
		{name: "length 256 upper only", opts: Options{Length: 256, UseUpper: true}},
		{name: "only digits length 8", opts: Options{Length: 8, UseDigits: true}},
		{name: "only symbols length 12", opts: Options{Length: 12, UseSymbols: true}},
		{name: "invalid zero length", opts: Options{Length: 0, UseLower: true}, wantErr: true},
		{name: "invalid empty charset", opts: Options{Length: 10}, wantErr: true},
		{name: "invalid too long", opts: Options{Length: 300, UseLower: true}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.opts.Length)

			charset := tt.opts.Charset()
			for i, ch := range got {
				assert.True(t, strings.ContainsRune(charset, ch), "char at %d = %q not in charset %q", i, ch, charset)
			}
		})
	}
}

func TestGenerate_CharsetFiltering(t *testing.T) {
	tests := []struct {
		name             string
		opts             Options
		shouldContain    string
		shouldNotContain string
	}{
		{name: "no symbols", opts: Options{Length: 20, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: false}, shouldContain: LowerLetters + UpperLetters + Digits, shouldNotContain: Symbols},
		{name: "only lower", opts: Options{Length: 20, UseLower: true}, shouldContain: LowerLetters, shouldNotContain: UpperLetters + Digits + Symbols},
		{name: "no digits no symbols", opts: Options{Length: 20, UseLower: true, UseUpper: true}, shouldContain: LowerLetters + UpperLetters, shouldNotContain: Digits + Symbols},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for iter := 0; iter < 50; iter++ {
				got, err := Generate(tt.opts)
				require.NoError(t, err)
				for _, ch := range got {
					assert.False(t, strings.ContainsRune(tt.shouldNotContain, ch), "char %q should not be in charset, excluded set %q", ch, tt.shouldNotContain)
					assert.True(t, strings.ContainsRune(tt.shouldContain, ch), "char %q not in expected charset %q", ch, tt.shouldContain)
				}
			}
		})
	}
}

func TestGenerateMultiple(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		count   int
		wantErr string
	}{
		{name: "single password", opts: DefaultOptions(), count: 1},
		{name: "5 passwords 12 chars", opts: Options{Length: 12, UseLower: true, UseUpper: true}, count: 5},
		{name: "10 passwords digits only", opts: Options{Length: 8, UseDigits: true}, count: 10},
		{name: "invalid count zero", opts: DefaultOptions(), count: 0, wantErr: "count must be greater than zero"},
		{name: "invalid count negative", opts: DefaultOptions(), count: -3, wantErr: "count must be greater than zero"},
		{name: "invalid opts empty charset", opts: Options{Length: 10}, count: 2, wantErr: "at least one character set must be enabled"},
		{name: "invalid opts zero length", opts: Options{Length: 0, UseLower: true}, count: 1, wantErr: "length must be greater than zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateMultiple(tt.opts, tt.count)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.count)

			charset := tt.opts.Charset()
			for idx, pwd := range got {
				assert.Len(t, pwd, tt.opts.Length, "password[%d] length", idx)
				for _, ch := range pwd {
					assert.True(t, strings.ContainsRune(charset, ch), "password[%d] char %q not in charset", idx, ch)
				}
			}
		})
	}
}

func TestGenerate_Randomness(t *testing.T) {
	// Statistical randomness check: generate 20 passwords; at least 18 should be unique.
	// This avoids flakiness of asserting two consecutive passwords differ while still
	// catching a broken RNG that returns deterministic output.
	opts := Options{Length: 16, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}
	seen := make(map[string]struct{})
	const n = 20
	for i := 0; i < n; i++ {
		pwd, err := Generate(opts)
		require.NoError(t, err, "iteration %d", i)
		seen[pwd] = struct{}{}
	}
	// Allow at most 2 collisions out of 20 (extremely unlikely with secure RNG).
	assert.GreaterOrEqual(t, len(seen), 18, "expected at least 18 unique passwords out of 20, RNG may be broken")
}

func TestGenerate_GuaranteesCharset(t *testing.T) {
	tests := []struct {
		name       string
		opts       Options
		iterations int
	}{
		{name: "length 4 all sets - minimal", opts: Options{Length: 4, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}, iterations: 20},
		{name: "length 8 all sets", opts: Options{Length: 8, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}, iterations: 50},
		{name: "length 12 all sets", opts: Options{Length: 12, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}, iterations: 30},
		{name: "length 10 lower and digits", opts: Options{Length: 10, UseLower: true, UseDigits: true}, iterations: 30},
		{name: "length 10 lower, upper, digits", opts: Options{Length: 10, UseLower: true, UseUpper: true, UseDigits: true}, iterations: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.iterations; i++ {
				pwd, err := Generate(tt.opts)
				require.NoError(t, err, "iteration %d", i)
				if tt.opts.UseLower {
					assert.True(t, strings.ContainsAny(pwd, LowerLetters), "should contain lower, pwd %q iter %d", pwd, i)
				}
				if tt.opts.UseUpper {
					assert.True(t, strings.ContainsAny(pwd, UpperLetters), "should contain upper, pwd %q iter %d", pwd, i)
				}
				if tt.opts.UseDigits {
					assert.True(t, strings.ContainsAny(pwd, Digits), "should contain digit, pwd %q iter %d - repetitive numbers issue", pwd, i)
				}
				if tt.opts.UseSymbols {
					assert.True(t, strings.ContainsAny(pwd, Symbols), "should contain symbol, pwd %q iter %d", pwd, i)
				}
			}
		})
	}
}

func TestGenerate_ValidateRequiresLengthForCharsets(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr string
	}{
		{name: "length 2 with 4 sets should fail", opts: Options{Length: 2, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}, wantErr: "length must be at least 4 to include all selected character sets"},
		{name: "length 3 with 4 sets should fail", opts: Options{Length: 3, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}, wantErr: "length must be at least 4 to include all selected character sets"},
		{name: "length 4 with 4 sets should pass", opts: Options{Length: 4, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}},
		{name: "length 2 with 2 sets should pass", opts: Options{Length: 2, UseLower: true, UseDigits: true}},
		{name: "length 1 with 2 sets should fail", opts: Options{Length: 1, UseLower: true, UseDigits: true}, wantErr: "length must be at least 2 to include all selected character sets"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestGenerate_RandErrors(t *testing.T) {
	origReader := randReader
	origRandInt := randInt
	t.Cleanup(func() {
		randReader = origReader
		randInt = origRandInt
	})

	// Helper reader that always fails.
	failingReader := errorReader{err: errors.New("simulated rand failure")}
	// Helper that returns deterministic 0 but never errors, used to count calls.
	simulateErrorOnNthCall := func(n int, err error) func(io.Reader, *big.Int) (*big.Int, error) {
		calls := 0
		return func(r io.Reader, max *big.Int) (*big.Int, error) {
			calls++
			if calls == n {
				return nil, err
			}
			return big.NewInt(0), nil
		}
	}

	tests := []struct {
		name    string
		opts    Options
		setup   func()
		wantErr string
	}{
		{
			name:    "rand failure via reader - lower",
			opts:    Options{Length: 8, UseLower: true},
			setup:   func() { randReader = failingReader; randInt = origRandInt },
			wantErr: "simulated rand failure",
		},
		{
			name:    "rand failure via reader - upper only",
			opts:    Options{Length: 8, UseUpper: true},
			setup:   func() { randReader = failingReader; randInt = origRandInt },
			wantErr: "simulated rand failure",
		},
		{
			name:    "rand failure via reader - digits only",
			opts:    Options{Length: 8, UseDigits: true},
			setup:   func() { randReader = failingReader; randInt = origRandInt },
			wantErr: "simulated rand failure",
		},
		{
			name:    "rand failure via reader - symbols only",
			opts:    Options{Length: 8, UseSymbols: true},
			setup:   func() { randReader = failingReader; randInt = origRandInt },
			wantErr: "simulated rand failure",
		},
		{
			name: "rand failure on fill phase",
			opts: Options{Length: 10, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
			setup: func() {
				randReader = origReader
				// 4 guaranteed chars + 1st fill char = 5th call fails
				randInt = simulateErrorOnNthCall(5, errors.New("fill phase failure"))
			},
			wantErr: "fill phase failure",
		},
		{
			name: "rand failure on shuffle phase",
			opts: Options{Length: 10, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
			setup: func() {
				randReader = origReader
				// 4 guaranteed + 6 fill =10, shuffle first iteration = 11th call
				randInt = simulateErrorOnNthCall(11, errors.New("shuffle phase failure"))
			},
			wantErr: "shuffle phase failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := Generate(tt.opts)
			require.Error(t, err)
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestGenerateMultiple_RandError(t *testing.T) {
	origRandInt := randInt
	t.Cleanup(func() { randInt = origRandInt })

	// Fail on second password generation.
	calls := 0
	randInt = func(r io.Reader, max *big.Int) (*big.Int, error) {
		calls++
		// First password needs ~16 + shuffle calls; fail after ~20 calls to ensure second password fails.
		if calls > 20 {
			return nil, errors.New("second password rand failure")
		}
		return big.NewInt(0), nil
	}

	_, err := GenerateMultiple(Options{Length: 10, UseLower: true}, 2)
	require.Error(t, err)
	assert.EqualError(t, err, "second password rand failure")
}

func TestGenerate_RequiredCount(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want int
	}{
		{name: "none", opts: Options{}, want: 0},
		{name: "only lower", opts: Options{UseLower: true}, want: 1},
		{name: "lower upper", opts: Options{UseLower: true, UseUpper: true}, want: 2},
		{name: "all four", opts: Options{UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}, want: 4},
		{name: "digits symbols", opts: Options{UseDigits: true, UseSymbols: true}, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.opts.requiredCount())
		})
	}
}

// errorReader is an io.Reader that always fails, used to test error propagation.
type errorReader struct {
	err error
}

func (e errorReader) Read(p []byte) (int, error) {
	return 0, e.err
}
