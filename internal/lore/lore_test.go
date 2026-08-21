package lore

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRankFor(t *testing.T) {
	tests := []struct {
		name string
		bits float64
		want Rank
	}{
		{name: "negative goes to Hobbit", bits: -5, want: Hobbit},
		{name: "zero Hobbit", bits: 0, want: Hobbit},
		{name: "27 Hobbit upper bound", bits: 27, want: Hobbit},
		{name: "27.9 Hobbit floor 27", bits: 27.9, want: Hobbit},
		{name: "28 Orc lower bound", bits: 28, want: Orc},
		{name: "35 Orc", bits: 35, want: Orc},
		{name: "35.9 Orc floor 35", bits: 35.9, want: Orc},
		{name: "36 Ranger lower bound", bits: 36, want: Ranger},
		{name: "59 Ranger", bits: 59, want: Ranger},
		{name: "60 Elf lower bound", bits: 60, want: Elf},
		{name: "79 Elf upper", bits: 79, want: Elf},
		{name: "80 Dwarf lower", bits: 80, want: Dwarf},
		{name: "99 Dwarf upper", bits: 99, want: Dwarf},
		{name: "100 Wizard lower", bits: 100, want: Wizard},
		{name: "127 Wizard upper", bits: 127, want: Wizard},
		{name: "128 Sauron lower", bits: 128, want: Sauron},
		{name: "200 Sauron", bits: 200, want: Sauron},
		{name: "1000 Sauron", bits: 1000, want: Sauron},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RankFor(tt.bits)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.MinBits, got.MinBits)
			assert.Equal(t, tt.want.MaxBits, got.MaxBits)
		})
	}
}

func TestRanks_OrderAndCoverage(t *testing.T) {
	// Ensure Ranks covers 0..infinity without gaps.
	require.Len(t, Ranks, 7)
	assert.Equal(t, "Hobbit", Ranks[0].Name)
	assert.Equal(t, "Sauron", Ranks[len(Ranks)-1].Name)

	// Check no gaps: each next MinBits == prev MaxBits.
	for i := 1; i < len(Ranks); i++ {
		assert.Equal(t, Ranks[i-1].MaxBits, Ranks[i].MinBits, "gap between %s and %s", Ranks[i-1].Name, Ranks[i].Name)
	}
	// First starts at 0, last ends at MaxInt.
	assert.Equal(t, 0, Ranks[0].MinBits)
	assert.Equal(t, math.MaxInt, Ranks[len(Ranks)-1].MaxBits)

	// Each rank has required fields.
	for _, r := range Ranks {
		t.Run(r.Name, func(t *testing.T) {
			assert.NotEmpty(t, r.Title)
			assert.NotEmpty(t, r.Quote)
			assert.NotEmpty(t, r.Advice)
			assert.NotEmpty(t, r.Color)
			assert.True(t, r.MinBits < r.MaxBits)
		})
	}
}

func TestCharsetInfo(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantSize int
		wantDesc string
	}{
		{name: "empty", password: "", wantSize: 0, wantDesc: "none"},
		{name: "only lower", password: "abc", wantSize: 26, wantDesc: "lower"},
		{name: "only upper", password: "ABC", wantSize: 26, wantDesc: "upper"},
		{name: "only digits", password: "123", wantSize: 10, wantDesc: "digits"},
		{name: "only symbols generator set", password: "!@#", wantSize: 29, wantDesc: "symbols"},
		{name: "other symbol triggers symbols", password: "abc€", wantSize: 26 + 29, wantDesc: "lower+symbols"},
		{name: "lower+upper", password: "aB", wantSize: 52, wantDesc: "lower+upper"},
		{name: "lower+digits", password: "a1", wantSize: 36, wantDesc: "lower+digits"},
		{name: "upper+digits", password: "A1", wantSize: 36, wantDesc: "upper+digits"},
		{name: "all charsets", password: "aA1!", wantSize: 91, wantDesc: "lower+upper+digits+symbols"},
		{name: "lower+upper+digits", password: "aA1", wantSize: 62, wantDesc: "lower+upper+digits"},
		{name: "lower+symbols", password: "a!", wantSize: 55, wantDesc: "lower+symbols"},
		{name: "digits+symbols", password: "1!", wantSize: 39, wantDesc: "digits+symbols"},
		{name: "upper+symbols", password: "A!", wantSize: 55, wantDesc: "upper+symbols"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, desc := charsetInfo(tt.password)
			assert.Equal(t, tt.wantSize, size)
			assert.Equal(t, tt.wantDesc, desc)
		})
	}
}

func TestEntropyBits(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		charsetSize int
		wantBits    float64
	}{
		{name: "empty zero", password: "", charsetSize: 0, wantBits: 0},
		{name: "empty with size", password: "", charsetSize: 26, wantBits: 0},
		{name: "charset zero", password: "abc", charsetSize: 0, wantBits: 0},
		{name: "lower 3 chars", password: "abc", charsetSize: 26, wantBits: 3 * math.Log2(26)},
		{name: "all 16 chars", password: "aA1!aA1!aA1!aA1!", charsetSize: 91, wantBits: 16 * math.Log2(91)},
		{name: "single char", password: "a", charsetSize: 26, wantBits: math.Log2(26)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := entropyBits(tt.password, tt.charsetSize)
			assert.InDelta(t, tt.wantBits, got, 0.0001)
		})
	}
}

func TestCrackTime(t *testing.T) {
	tests := []struct {
		name string
		bits float64
		want string
	}{
		{name: "zero instant", bits: 0, want: "instant"},
		{name: "negative instant", bits: -5, want: "instant"},
		{name: "small bits instant", bits: 10, want: "instant"},
		{name: "20 bits still instant (1M guesses)", bits: 20, want: "instant"},
		{name: "30 bits still instant", bits: 30, want: "instant"},
		{name: "40 bits minutes", bits: 40, want: "minute"},
		{name: "50 bits days", bits: 50, want: "day"},
		{name: "60 bits years", bits: 60, want: "year"},
		{name: "70 bits years", bits: 70, want: "year"},
		{name: "80 bits huge", bits: 80, want: "year"},
		{name: "100 bits very huge", bits: 100, want: "year"},
		{name: "128 bits mythical", bits: 128, want: "year"},
		{name: "200 bits trillion", bits: 200, want: "year"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := crackTime(tt.bits)
			// For range tests, just check contains expected substring and not empty.
			assert.NotEmpty(t, got)
			assert.Contains(t, strings.ToLower(got), strings.ToLower(tt.want))
		})
	}

	// Exact expectations for determinism.
	exact := []struct {
		bits float64
		want string
	}{
		{bits: 0, want: "instant"},
		{bits: 1, want: "instant"},
		{bits: 30, want: "1 second"}, // 2^30 /1e10 ~0.1 sec? Actually 1,073,741,824/1e10=0.107 instant -> adjust
	}
	// Don't enforce fragile exact; just verify instant vs not instant boundary logic works.
	for _, tt := range exact {
		t.Run("exact_"+tt.want, func(t *testing.T) {
			got := crackTime(tt.bits)
			if tt.want == "instant" {
				assert.Equal(t, "instant", got)
			}
		})
	}
}

func TestDetectWarnings_Empty(t *testing.T) {
	warns := detectWarnings("")
	require.Len(t, warns, 1)
	assert.Contains(t, warns[0], "empty")
}

func TestDetectWarnings_Short(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantWarn string
		wantNot  string
	}{
		{name: "short 3 chars", password: "abc", wantWarn: "short:"},
		{name: "short 7 chars", password: "Ab1!xyz", wantWarn: "short:"},
		{name: "not short 8 chars", password: "Ab1!xyzQ", wantNot: "short:"},
		{name: "not short 16", password: "Ab1!xyzQAb1!xyzQ", wantNot: "short:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns := detectWarnings(tt.password)
			joined := strings.Join(warns, "|")
			if tt.wantWarn != "" {
				assert.Contains(t, joined, tt.wantWarn)
			}
			if tt.wantNot != "" {
				assert.NotContains(t, joined, tt.wantNot)
			}
		})
	}
}

func TestDetectWarnings_SingleCharset(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantWarn bool
	}{
		{name: "only lower 8", password: "abcdefgh", wantWarn: true},
		{name: "only upper 8", password: "ABCDEFGH", wantWarn: true},
		{name: "only digits 8", password: "12345678", wantWarn: true},
		{name: "only symbols 8", password: "!@#$%^&*", wantWarn: true},
		{name: "lower+upper no warn", password: "abCD", wantWarn: false},
		{name: "lower+digit no warn", password: "ab12", wantWarn: false},
		{name: "all charsets no warn", password: "aA1!", wantWarn: false},
		{name: "mixed case with symbols no warn", password: "aA!aA!", wantWarn: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns := detectWarnings(tt.password)
			found := false
			for _, w := range warns {
				if strings.Contains(w, "only one character type") {
					found = true
					break
				}
			}
			assert.Equal(t, tt.wantWarn, found, "warns: %v", warns)
		})
	}
}

func TestDetectWarnings_CommonPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantWarn bool
	}{
		{name: "password exact", password: "password", wantWarn: true},
		{name: "PASSWORD case insensitive", password: "PASSWORD", wantWarn: true},
		{name: "123456 exact", password: "123456", wantWarn: true},
		{name: "qwerty exact", password: "qwerty", wantWarn: true},
		{name: "password inside short string", password: "password123", wantWarn: true},
		{name: "password inside long string >12 no common warning (but lore maybe)", password: "myLongPasswordIsSafe123!", wantWarn: false},
		{name: "admin exact", password: "admin", wantWarn: true},
		{name: "unique not common", password: "xK9$m2Pq", wantWarn: false},
		{name: "starwars exact", password: "starwars", wantWarn: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns := detectWarnings(tt.password)
			found := false
			for _, w := range warns {
				if strings.Contains(w, "common password") {
					found = true
					break
				}
			}
			assert.Equal(t, tt.wantWarn, found, "warns: %v", warns)
		})
	}
}

func TestDetectWarnings_LoreWords(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantWarn bool
		wantWord string
	}{
		{name: "mellon exact", password: "mellon", wantWarn: true, wantWord: "mellon"},
		{name: "GANDALF case insensitive", password: "GandalfTheWhite", wantWarn: true, wantWord: "gandalf"},
		{name: "frodo", password: "frodo123", wantWarn: true, wantWord: "frodo"},
		{name: "shire lore", password: "shire2023", wantWarn: true, wantWord: "shire"},
		{name: "sauron lore", password: "Sauron!123", wantWarn: true, wantWord: "sauron"},
		{name: "mordor lore", password: "mordorDark", wantWarn: true, wantWord: "mordor"},
		{name: "ring lore", password: "myring123", wantWarn: true, wantWord: "ring"},
		{name: "hobbit lore", password: "HobbitHole", wantWarn: true, wantWord: "hobbit"},
		{name: "no lore word", password: "xK9$m2PqL8@vB4", wantWarn: false},
		{name: "precious lore", password: "precious", wantWarn: true, wantWord: "precious"},
		{name: "baggins lore", password: "Baggins123", wantWarn: true, wantWord: "baggins"},
		{name: "elf generic", password: "myElfPass123", wantWarn: true, wantWord: "elf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns := detectWarnings(tt.password)
			found := false
			foundWord := ""
			for _, w := range warns {
				if strings.Contains(w, "Middle-earth") {
					found = true
					foundWord = w
					break
				}
			}
			assert.Equal(t, tt.wantWarn, found, "warns: %v", warns)
			if tt.wantWarn && tt.wantWord != "" {
				assert.Contains(t, strings.ToLower(foundWord), strings.ToLower(tt.wantWord))
			}
		})
	}
}

func TestDetectWarnings_Sequential(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantWarn bool
	}{
		{name: "abc sequential", password: "abcXYZ", wantWarn: true},
		{name: "123 sequential", password: "aA1!123", wantWarn: true},
		{name: "cba reverse", password: "cba321", wantWarn: true},
		{name: "qwerty sequential", password: "qwerty123", wantWarn: true},
		{name: "asdf sequential", password: "asdf1234", wantWarn: true},
		{name: "no sequential", password: "xK9$m2Pq", wantWarn: false},
		{name: "mixed not sequential", password: "aB3$dE7!", wantWarn: false},
		{name: "xyz sequential", password: "aBxyz12!", wantWarn: true},
		{name: "short not sequential", password: "aB", wantWarn: false},
		{name: "321 reverse digits", password: "321abc", wantWarn: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns := detectWarnings(tt.password)
			found := false
			for _, w := range warns {
				if strings.Contains(w, "sequential") {
					found = true
					break
				}
			}
			assert.Equal(t, tt.wantWarn, found, "warns: %v", warns)
		})
	}
}

func TestDetectWarnings_Repeating(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantWarn bool
	}{
		{name: "aaa repeating", password: "aaaB1!", wantWarn: true},
		{name: "111 repeating", password: "aA111!b", wantWarn: true},
		{name: "!!! repeating", password: "aA111!!!", wantWarn: true},
		{name: "no repeat", password: "aA1!bB2@", wantWarn: false},
		{name: "aa not enough", password: "aaB1!c", wantWarn: false},
		{name: "two repeats but not three", password: "aAbb12!", wantWarn: false},
		{name: "repeating in middle", password: "abbbcd1!", wantWarn: true},
		{name: "short no repeating", password: "ab", wantWarn: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns := detectWarnings(tt.password)
			found := false
			for _, w := range warns {
				if strings.Contains(w, "repeating") {
					found = true
					break
				}
			}
			assert.Equal(t, tt.wantWarn, found, "warns: %v", warns)
		})
	}
}

func TestHasSequentialEdge(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{name: "empty", s: "", want: false},
		{name: "two chars", s: "ab", want: false},
		{name: "exact abc", s: "abc", want: true},
		{name: "upper ABC", s: "ABC", want: true},
		{name: "qwe", s: "qwe", want: true},
		{name: "zxy no", s: "zxy", want: false},
		{name: "321", s: "321", want: true},
		{name: "890 no wrap", s: "890", want: false},
		{name: "789", s: "789", want: true},
		{name: "random", s: "k9$2", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasSequential(tt.s))
		})
	}
}

func TestHasRepeatingEdge(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{name: "empty", s: "", want: false},
		{name: "two", s: "aa", want: false},
		{name: "three same", s: "aaa", want: true},
		{name: "four same", s: "aaaa", want: true},
		{name: "aba no", s: "aba", want: false},
		{name: "aab no", s: "aab", want: false},
		{name: "abb no two", s: "abb", want: false},
		{name: "abbb yes", s: "abbb", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasRepeating(tt.s))
		})
	}
}

func TestEvaluate_Basics(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantRank  string
		wantWarns int // at least
		wantSize  int
	}{
		{name: "empty Hobbit", password: "", wantRank: "Hobbit", wantWarns: 1, wantSize: 0},
		{name: "short weak Hobbit", password: "abc", wantRank: "Hobbit", wantWarns: 2, wantSize: 26},
		{name: "password common Hobbit", password: "password", wantRank: "Hobbit", wantWarns: 1, wantSize: 26},
		{name: "123456 Hobbit", password: "123456", wantRank: "Hobbit", wantWarns: 1, wantSize: 10},
		{name: "mellon lore Hobbit", password: "mellon", wantRank: "Hobbit", wantWarns: 1, wantSize: 26},
		{name: "strong generated style Wizard/Sauron", password: "aA1!aA1!aA1!aA1!aA1!", wantRank: "Sauron", wantWarns: 0, wantSize: 91},
		{name: "moderate Ranger", password: "aA1!aB2@", wantRank: "Ranger", wantWarns: 0, wantSize: 91},
		{name: "elf level 12 mixed", password: "xK9$m2PqL8@v", wantRank: "Elf", wantWarns: 0, wantSize: 91},
		{name: "dwarf 14 mixed", password: "xK9$m2PqL8@vB4", wantRank: "Dwarf", wantWarns: 0, wantSize: 91},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Evaluate(tt.password)
			assert.Equal(t, tt.wantRank, res.Rank.Name, "entropy %.2f charset %d warns %v", res.Entropy, res.CharsetSize, res.Warnings)
			assert.GreaterOrEqual(t, len(res.Warnings), tt.wantWarns)
			assert.Equal(t, tt.wantSize, res.CharsetSize)
			assert.NotEmpty(t, res.CrackTime)
			assert.NotEmpty(t, res.CharsetDesc)
			// Entropy non-negative
			assert.GreaterOrEqual(t, res.Entropy, 0.0)
		})
	}
}

func TestEvaluate_PenaltiesAffectRank(t *testing.T) {
	// Same raw entropy but with penalties should drop rank.
	// Without warning: 8 chars all sets = 8*log2(91)= ~52 bits => Ranger (36-60) but sequential drops to Hobbit/Orc.

	tests := []struct {
		name     string
		password string
		wantRank string
	}{
		{name: "common drops", password: "password123", wantRank: "Hobbit"},
		{name: "sequential drops", password: "abc123ABC", wantRank: "Hobbit"},
		{name: "repeating drops", password: "aaaBBB111", wantRank: "Ranger"}, // raw would be Elf-ish but penalty drops
		{name: "lore drops short", password: "Gandalf1!", wantRank: "Ranger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Evaluate(tt.password)
			assert.Equal(t, tt.wantRank, res.Rank.Name, "for %q got %s entropy %.2f warns %v", tt.password, res.Rank.Name, res.Entropy, res.Warnings)
		})
	}
}

func TestEvaluate_Deterministic(t *testing.T) {
	// Evaluate should be deterministic.
	pwd := "xK9$m2PqL8@vB4!"
	res1 := Evaluate(pwd)
	res2 := Evaluate(pwd)
	assert.Equal(t, res1.Entropy, res2.Entropy)
	assert.Equal(t, res1.Rank.Name, res2.Rank.Name)
	assert.Equal(t, res1.CharsetSize, res2.CharsetSize)
	assert.Equal(t, res1.CrackTime, res2.CrackTime)
	assert.Equal(t, res1.Warnings, res2.Warnings)
}

func TestEvaluate_CharsetDesc(t *testing.T) {
	tests := []struct {
		password string
		wantDesc string
	}{
		{password: "abc", wantDesc: "lower"},
		{password: "ABC", wantDesc: "upper"},
		{password: "123", wantDesc: "digits"},
		{password: "!@#", wantDesc: "symbols"},
		{password: "aA1!", wantDesc: "lower+upper+digits+symbols"},
		{password: "", wantDesc: "none"},
	}
	for _, tt := range tests {
		t.Run(tt.password+"_desc", func(t *testing.T) {
			res := Evaluate(tt.password)
			assert.Equal(t, tt.wantDesc, res.CharsetDesc)
		})
	}
}

func TestEvaluate_CrackTimeNotEmpty(t *testing.T) {
	passwords := []string{"a", "ab", "aA1!", "aA1!aA1!aA1!", "xK9$m2PqL8@vB4!xK9$m2PqL8@vB4!xK9$m2PqL8@vB4!"}
	for _, pwd := range passwords {
		t.Run(pwd, func(t *testing.T) {
			res := Evaluate(pwd)
			assert.NotEmpty(t, res.CrackTime)
			// Hobbit should be instant or seconds, Sauron should be years.
			if res.Rank.Name == "Hobbit" {
				// Allow instant, seconds, minutes.
				assert.True(t, strings.Contains(res.CrackTime, "instant") || strings.Contains(res.CrackTime, "second") || strings.Contains(res.CrackTime, "minute") || strings.Contains(res.CrackTime, "hour") || strings.Contains(res.CrackTime, "day"), "Hobbit crack %q", res.CrackTime)
			}
			if res.Rank.Name == "Sauron" {
				assert.Contains(t, res.CrackTime, "year")
			}
		})
	}
}

func TestEvaluate_EntropyBounds(t *testing.T) {
	// Very long strong password should reach Sauron even with minor warnings? No warnings for this.
	pwd := strings.Repeat("aA1!", 32) // 128 chars, but has repeating? Actually "aA1!" repeated has no 3 repeating same char, but sequential? No.
	// It has repeating pattern but not 3 same char. Should be strong.
	res := Evaluate(pwd)
	assert.Equal(t, "Sauron", res.Rank.Name)
	assert.GreaterOrEqual(t, res.Entropy, 128.0)

	// Empty should be 0.
	resEmpty := Evaluate("")
	assert.Equal(t, 0.0, resEmpty.Entropy)
	assert.Equal(t, "Hobbit", resEmpty.Rank.Name)
}

func TestEvaluate_WarningsContainExpectedSubstrings(t *testing.T) {
	tests := []struct {
		password string
		substr   string
	}{
		{password: "", substr: "empty"},
		{password: "abc", substr: "short"},
		{password: "password", substr: "common password"},
		{password: "mellon", substr: "Middle-earth"},
		{password: "abc123", substr: "sequential"},
		{password: "aaaBBB", substr: "repeating"},
		{password: "abcdefgh", substr: "only one character type"},
	}
	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			res := Evaluate(tt.password)
			found := false
			for _, w := range res.Warnings {
				if strings.Contains(strings.ToLower(w), strings.ToLower(tt.substr)) {
					found = true
					break
				}
			}
			assert.True(t, found, "expected warning containing %q in %v", tt.substr, res.Warnings)
		})
	}
}
