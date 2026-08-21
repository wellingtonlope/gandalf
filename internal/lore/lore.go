package lore

import (
	"fmt"
	"math"
	"strings"
)

// Rank represents a Middle-earth themed password strength level.
type Rank struct {
	Name    string
	Title   string
	MinBits int
	MaxBits int // exclusive, except last rank uses math.MaxInt
	Color   string
	Quote   string
	Advice  string
}

// Predefined ranks ordered by increasing strength.
// Thresholds align with NIST guidance: <28 instant crack, 28-35 weak, 36-59 fair, 60-79 strong, 80-99 very strong, 100-127 wizard-level, 128+ legendary.
var (
	Hobbit = Rank{
		Name:    "Hobbit",
		Title:   "Hobbit of the Shire",
		MinBits: 0,
		MaxBits: 28,
		Color:   "red",
		Quote:   "Like butter scraped over too much bread. Even a Hobbit could guess it.",
		Advice:  "Forge a longer password: use 12+ chars with upper, lower, digits and symbols like Dwarven runes.",
	}
	Orc = Rank{
		Name:    "Orc",
		Title:   "Orc of Mordor",
		MinBits: 28,
		MaxBits: 36,
		Color:   "red",
		Quote:   "Weak as an Orc in sunlight. It will not survive the road.",
		Advice:  "Add more character types and length. Avoid common words.",
	}
	Ranger = Rank{
		Name:    "Ranger",
		Title:   "Ranger of the North",
		MinBits: 36,
		MaxBits: 60,
		Color:   "yellow",
		Quote:   "A Ranger wouldn't trust the road with this. Better, but still watchful.",
		Advice:  "Increase length to 12+ and mix all character sets.",
	}
	Elf = Rank{
		Name:    "Elf",
		Title:   "Elf of Rivendell",
		MinBits: 60,
		MaxBits: 80,
		Color:   "green",
		Quote:   "Keen as Elven sight. Strong enough for many journeys.",
		Advice:  "Great! For critical vaults, aim for 16+ chars.",
	}
	Dwarf = Rank{
		Name:    "Dwarf",
		Title:   "Dwarf Lord of Erebor",
		MinBits: 80,
		MaxBits: 100,
		Color:   "green",
		Quote:   "Hard as the Doors of Durin. Crafted with Dwarven steel.",
		Advice:  "Excellent - hard as Khazad-Dûm doors. Keep it safe.",
	}
	Wizard = Rank{
		Name:    "Wizard",
		Title:   "Gandalf the White",
		MinBits: 100,
		MaxBits: 128,
		Color:   "blue",
		Quote:   "You shall not pass! Attackers shall not pass this password.",
		Advice:  "Legendary strength. Guard it as the White Council guards secrets.",
	}
	Sauron = Rank{
		Name:    "Sauron",
		Title:   "The One Ring",
		MinBits: 128,
		MaxBits: math.MaxInt,
		Color:   "yellow",
		Quote:   "One password to rule them all. Forged in Mount Doom.",
		Advice:  "Mythical. Even the Eye cannot pierce it easily. Don't lose it!",
	}
)

// Ranks is the ordered list of all ranks from weakest to strongest.
var Ranks = []Rank{Hobbit, Orc, Ranger, Elf, Dwarf, Wizard, Sauron}

// RankFor returns the Rank for the given entropy bits.
// Negative bits are treated as 0 (Hobbit).
func RankFor(bits float64) Rank {
	if bits < 0 {
		bits = 0
	}
	b := int(math.Floor(bits))
	for _, r := range Ranks {
		if b >= r.MinBits && b < r.MaxBits {
			return r
		}
	}
	return Sauron
}

// Result holds the evaluation of a password.
type Result struct {
	Password    string
	Entropy     float64
	CharsetSize int
	CharsetDesc string
	Rank        Rank
	Warnings    []string
	CrackTime   string
}

// Evaluate analyzes the password and returns its strength result.
// It estimates entropy as len * log2(charsetSize) where charsetSize is inferred
// from character types present, then applies thematic penalties for patterns.
func Evaluate(password string) Result {
	size, desc := charsetInfo(password)
	bits := entropyBits(password, size)

	// Apply penalties for weak patterns; they reduce effective entropy.
	penalty := 0.0
	warnings := detectWarnings(password)

	// Each warning type contributes a penalty. Keep it deterministic for tests.
	for _, w := range warnings {
		switch {
		case strings.Contains(w, "common password"):
			penalty += 20
		case strings.Contains(w, "Middle-earth"):
			penalty += 15
		case strings.Contains(w, "sequential"):
			penalty += 10
		case strings.Contains(w, "repeating"):
			penalty += 10
		case strings.Contains(w, "only one character type"):
			penalty += 5
		case strings.Contains(w, "short"):
			penalty += 5
		}
	}

	effective := bits - penalty
	if effective < 0 {
		effective = 0
	}

	rank := RankFor(effective)

	return Result{
		Password:    password,
		Entropy:     effective,
		CharsetSize: size,
		CharsetDesc: desc,
		Rank:        rank,
		Warnings:    warnings,
		CrackTime:   crackTime(effective),
	}
}

// charsetInfo infers charset size and description from the password content.
func charsetInfo(password string) (int, string) {
	if password == "" {
		return 0, "none"
	}
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(password, "0123456789")
	hasSymbol := false
	// Symbols as defined in generator.Symbols: "!@#$%^&*()-_=+[]{}|;:,.<>?/~`"
	const symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?/~`"
	if strings.ContainsAny(password, symbols) {
		hasSymbol = true
	} else {
		// Also consider any non-alphanumeric as symbol for check purposes.
		for _, r := range password {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				hasSymbol = true
				break
			}
		}
	}

	size := 0
	parts := []string{}
	if hasLower {
		size += 26
		parts = append(parts, "lower")
	}
	if hasUpper {
		size += 26
		parts = append(parts, "upper")
	}
	if hasDigit {
		size += 10
		parts = append(parts, "digits")
	}
	if hasSymbol {
		size += 29 // matches len(generator.Symbols)
		parts = append(parts, "symbols")
	}
	if size == 0 {
		return 0, "none"
	}
	return size, strings.Join(parts, "+")
}

// entropyBits calculates raw entropy without penalties.
func entropyBits(password string, charsetSize int) float64 {
	if password == "" || charsetSize <= 0 {
		return 0
	}
	return float64(len(password)) * math.Log2(float64(charsetSize))
}

// crackTime estimates time to crack at 10B guesses/sec.
func crackTime(bits float64) string {
	if bits <= 0 {
		return "instant"
	}
	guesses := math.Pow(2, bits)
	const perSecond = 1e10
	seconds := guesses / perSecond

	switch {
	case seconds < 1:
		return "instant"
	case seconds < 60:
		return formatDuration(seconds, "second")
	case seconds < 3600:
		return formatDuration(seconds/60, "minute")
	case seconds < 86400:
		return formatDuration(seconds/3600, "hour")
	case seconds < 31536000:
		return formatDuration(seconds/86400, "day")
	case seconds < 31536000*100:
		return formatDuration(seconds/31536000, "year")
	case seconds < 31536000*10000:
		years := seconds / 31536000
		if years >= 1e6 {
			return formatLargeYears(years)
		}
		return formatDuration(years, "year")
	default:
		years := seconds / 31536000
		return formatLargeYears(years)
	}
}

func formatDuration(v float64, unit string) string {
	var s string
	if v < 10 {
		s = fmt.Sprintf("%.1f", v)
		s = strings.TrimSuffix(s, ".0")
	} else {
		s = fmt.Sprintf("%.0f", v)
	}
	if v >= 0.95 && v <= 1.05 {
		// Treat ~1 as singular for readability (1.0 -> 1 second).
		// Use exact check for 1 to keep tests deterministic.
		if s == "1" || s == "1.0" {
			return s + " " + unit
		}
	}
	return s + " " + unit + "s"
}

func formatLargeYears(years float64) string {
	if years >= 1e12 {
		return fmt.Sprintf("%.1f trillion years", years/1e12)
	}
	if years >= 1e9 {
		return fmt.Sprintf("%.1f billion years", years/1e9)
	}
	if years >= 1e6 {
		return fmt.Sprintf("%.1f million years", years/1e6)
	}
	return fmt.Sprintf("%.0f years", years)
}

// detectWarnings returns thematic warnings for weak patterns.
func detectWarnings(password string) []string {
	var warns []string
	if password == "" {
		warns = append(warns, "empty password - even a Hobbit wouldn't use this")
		return warns
	}
	lower := strings.ToLower(password)

	// Length warning.
	if len(password) < 8 {
		warns = append(warns, "short: less than 8 characters - too short for the road to Mordor")
	}

	// Only one charset type.
	size, _ := charsetInfo(password)
	if size == 26 || size == 10 || size == 29 {
		// 26 could be lower OR upper, 10 digits, 30 symbols-only.
		hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
		hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		hasDigit := strings.ContainsAny(password, "0123456789")
		countTypes := 0
		if hasLower {
			countTypes++
		}
		if hasUpper {
			countTypes++
		}
		if hasDigit {
			countTypes++
		}
		// Check symbol separately
		hasSymbol := false
		for _, r := range password {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				hasSymbol = true
				break
			}
		}
		if hasSymbol {
			countTypes++
		}
		if countTypes == 1 {
			warns = append(warns, "only one character type - forge with more runes (mix upper, lower, digits, symbols)")
		}
	}

	// Common passwords (top list, case-insensitive exact or substring if short).
	commonPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123", "password1",
		"12345678", "111111", "123123", "admin", "letmein", "welcome",
		"monkey", "dragon", "master", "login", "princess", "solo",
		"starwars", "654321", "qwerty123", "1234567", "1234567890",
	}
	for _, c := range commonPasswords {
		if lower == c || (len(password) <= 12 && strings.Contains(lower, c)) {
			warns = append(warns, "common password pattern detected: '"+c+"' - The Eye sees this!")
			break
		}
	}

	// Middle-earth lore words penalty (thematic).
	loreWords := []string{
		"mellon", "friend", "gandalf", "frodo", "shire", "hobbit",
		"gondor", "aragorn", "legolas", "gimli", "sauron", "mordor",
		"precious", "ring", "middle", "earth", "baggins", "samwise",
		"bilbo", "gollum", "smeagol", "orcs", "orc", "elf", "dwarf",
		"shire", "rohan", "isengard",
	}
	for _, w := range loreWords {
		if strings.Contains(lower, w) {
			warns = append(warns, "contains Middle-earth word '"+w+"' - The Eye knows the old tongues!")
			break
		}
	}

	// Sequential patterns: abc, bcd, 123, 321, qwerty fragments.
	if hasSequential(password) {
		warns = append(warns, "sequential characters detected (abc, 123, qwerty) - too predictable for a Ranger")
	}

	// Repeating characters:aaa, 111, etc.
	if hasRepeating(password) {
		warns = append(warns, "repeating characters detected - like Orc drums, easy to guess")
	}

	return warns
}

func hasSequential(s string) bool {
	lower := strings.ToLower(s)
	// Check for 3-char sequences in alphabet, digits, qwerty.
	sequences := []string{
		"abcdefghijklmnopqrstuvwxyz",
		"0123456789",
		"qwertyuiop",
		"asdfghjkl",
		"zxcvbnm",
	}
	for _, seq := range sequences {
		for i := 0; i <= len(seq)-3; i++ {
			sub := seq[i : i+3]
			rev := reverseString(sub)
			if strings.Contains(lower, sub) || strings.Contains(lower, rev) {
				return true
			}
		}
	}
	return false
}

func hasRepeating(s string) bool {
	if len(s) < 3 {
		return false
	}
	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= 3 {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

func reverseString(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
