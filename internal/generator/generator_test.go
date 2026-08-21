package generator

import (
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
	tests := map[string]struct {
		opts     Options
		expected string
	}{
		"all enabled": {
			opts:     Options{UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
			expected: LowerLetters + UpperLetters + Digits + Symbols,
		},
		"only lower": {
			opts:     Options{UseLower: true},
			expected: LowerLetters,
		},
		"only upper": {
			opts:     Options{UseUpper: true},
			expected: UpperLetters,
		},
		"only digits": {
			opts:     Options{UseDigits: true},
			expected: Digits,
		},
		"only symbols": {
			opts:     Options{UseSymbols: true},
			expected: Symbols,
		},
		"lower and digits": {
			opts:     Options{UseLower: true, UseDigits: true},
			expected: LowerLetters + Digits,
		},
		"upper and symbols": {
			opts:     Options{UseUpper: true, UseSymbols: true},
			expected: UpperLetters + Symbols,
		},
		"none enabled": {
			opts:     Options{},
			expected: "",
		},
		"lower upper no digits no symbols": {
			opts:     Options{UseLower: true, UseUpper: true},
			expected: LowerLetters + UpperLetters,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.opts.Charset())
		})
	}
}

func TestOptions_Validate(t *testing.T) {
	tests := map[string]struct {
		opts    Options
		wantErr string
	}{
		"valid default": {
			opts: DefaultOptions(),
		},
		"valid custom 32 lower only": {
			opts: Options{Length: 32, UseLower: true},
		},
		"valid length 1": {
			opts: Options{Length: 1, UseLower: true},
		},
		"valid length 256": {
			opts: Options{Length: 256, UseLower: true},
		},
		"invalid length zero": {
			opts:    Options{Length: 0, UseLower: true},
			wantErr: "length must be greater than zero",
		},
		"invalid length negative": {
			opts:    Options{Length: -5, UseLower: true},
			wantErr: "length must be greater than zero",
		},
		"invalid length too large": {
			opts:    Options{Length: 257, UseLower: true},
			wantErr: "maximum length is 256",
		},
		"invalid empty charset": {
			opts:    Options{Length: 16},
			wantErr: "at least one character set must be enabled",
		},
		"invalid zero length and empty charset - length error first": {
			opts:    Options{Length: 0},
			wantErr: "length must be greater than zero",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	tests := map[string]struct {
		opts    Options
		wantErr bool
	}{
		"default options 16 chars": {
			opts: DefaultOptions(),
		},
		"length 1 lower only": {
			opts: Options{Length: 1, UseLower: true},
		},
		"length 32 all charsets": {
			opts: Options{Length: 32, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
		},
		"length 256 upper only": {
			opts: Options{Length: 256, UseUpper: true},
		},
		"only digits length 8": {
			opts: Options{Length: 8, UseDigits: true},
		},
		"only symbols length 12": {
			opts: Options{Length: 12, UseSymbols: true},
		},
		"invalid zero length": {
			opts:    Options{Length: 0, UseLower: true},
			wantErr: true,
		},
		"invalid empty charset": {
			opts:    Options{Length: 10},
			wantErr: true,
		},
		"invalid too long": {
			opts:    Options{Length: 300, UseLower: true},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	tests := map[string]struct {
		opts             Options
		shouldContain    string
		shouldNotContain string
	}{
		"no symbols": {
			opts:             Options{Length: 20, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: false},
			shouldContain:    LowerLetters + UpperLetters + Digits,
			shouldNotContain: Symbols,
		},
		"only lower": {
			opts:             Options{Length: 20, UseLower: true},
			shouldContain:    LowerLetters,
			shouldNotContain: UpperLetters + Digits + Symbols,
		},
		"no digits no symbols": {
			opts:             Options{Length: 20, UseLower: true, UseUpper: true},
			shouldContain:    LowerLetters + UpperLetters,
			shouldNotContain: Digits + Symbols,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for iter := 0; iter < 10; iter++ {
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
	tests := map[string]struct {
		opts    Options
		count   int
		wantErr string
	}{
		"single password": {
			opts:  DefaultOptions(),
			count: 1,
		},
		"5 passwords 12 chars": {
			opts:  Options{Length: 12, UseLower: true, UseUpper: true},
			count: 5,
		},
		"10 passwords digits only": {
			opts:  Options{Length: 8, UseDigits: true},
			count: 10,
		},
		"invalid count zero": {
			opts:    DefaultOptions(),
			count:   0,
			wantErr: "count must be greater than zero",
		},
		"invalid count negative": {
			opts:    DefaultOptions(),
			count:   -3,
			wantErr: "count must be greater than zero",
		},
		"invalid opts empty charset": {
			opts:    Options{Length: 10},
			count:   2,
			wantErr: "at least one character set must be enabled",
		},
		"invalid opts zero length": {
			opts:    Options{Length: 0, UseLower: true},
			count:   1,
			wantErr: "length must be greater than zero",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	// Basic randomness check: two consecutive passwords should rarely be identical.
	opts := Options{Length: 16, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true}
	p1, err1 := Generate(opts)
	p2, err2 := Generate(opts)
	require.NoError(t, err1)
	require.NoError(t, err2)
	if p1 == p2 {
		t.Logf("warning: two consecutive passwords are identical (%q), possible but unlikely", p1)
	}
	assert.NotEqual(t, p1, p2, "two consecutive passwords should not be identical (flaky if equal)")
}

func TestGenerate_GuaranteesCharset(t *testing.T) {
	// This test proves the current algorithm is repetitive: without guarantee,
	// passwords with all charsets enabled often miss digits/symbols, especially
	// with short lengths. After the fix, every password must contain at least
	// one char from each enabled set (and be shuffled).
	tests := map[string]struct {
		opts       Options
		iterations int
	}{
		"length 4 all sets - minimal": {
			opts:       Options{Length: 4, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
			iterations: 20,
		},
		"length 8 all sets": {
			opts:       Options{Length: 8, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
			iterations: 50,
		},
		"length 12 all sets": {
			opts:       Options{Length: 12, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
			iterations: 30,
		},
		"length 10 lower and digits": {
			opts:       Options{Length: 10, UseLower: true, UseDigits: true},
			iterations: 30,
		},
		"length 10 lower, upper, digits": {
			opts:       Options{Length: 10, UseLower: true, UseUpper: true, UseDigits: true},
			iterations: 30,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	tests := map[string]struct {
		opts    Options
		wantErr string
	}{
		"length 2 with 4 sets should fail": {
			opts:    Options{Length: 2, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
			wantErr: "length must be at least 4 to include all selected character sets",
		},
		"length 3 with 4 sets should fail": {
			opts:    Options{Length: 3, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
			wantErr: "length must be at least 4 to include all selected character sets",
		},
		"length 4 with 4 sets should pass": {
			opts: Options{Length: 4, UseLower: true, UseUpper: true, UseDigits: true, UseSymbols: true},
		},
		"length 2 with 2 sets should pass": {
			opts: Options{Length: 2, UseLower: true, UseDigits: true},
		},
		"length 1 with 2 sets should fail": {
			opts:    Options{Length: 1, UseLower: true, UseDigits: true},
			wantErr: "length must be at least 2 to include all selected character sets",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
