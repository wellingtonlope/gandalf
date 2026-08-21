package mellon

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMellon(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "elvish mellon lowercase", input: "mellon", want: true},
		{name: "english friend lowercase", input: "friend", want: true},
		{name: "uppercase mellon - exact match false", input: "MELLON", want: false},
		{name: "uppercase friend - exact match false", input: "FRIEND", want: false},
		{name: "empty string", input: "", want: false},
		{name: "random word", input: "gandalf", want: false},
		{name: "mellon with space", input: " mellon", want: false},
		{name: "friend with trailing space", input: "friend ", want: false},
		{name: "amigo old portuguese should be false now", input: "amigo", want: false},
		{name: "partial mellon", input: "mell", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsMellon(tt.input))
		})
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "Elvish", value: Elvish, want: "mellon"},
		{name: "Translation", value: Translation, want: "friend"},
		{name: "SpeakFriend", value: SpeakFriend, want: "Speak, friend, and enter."},
		{name: "PortalInscription", value: PortalInscription, want: PortalInscription},
		{name: "PortalTranslation", value: PortalTranslation, want: PortalTranslation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value)
			assert.NotEmpty(t, tt.value)
		})
	}

	// Additional cross-checks (outside table but deterministic).
	assert.Contains(t, PortalInscription, "mellon")
	assert.Contains(t, PortalTranslation, "Speak, friend, and enter")
}

func TestReveal_Output(t *testing.T) {
	output := captureReveal(t)

	tests := []struct {
		name   string
		substr string
	}{
		{name: "contains Mellon header", substr: "Mellon"},
		{name: "contains PortalInscription", substr: PortalInscription},
		{name: "contains PortalTranslation", substr: PortalTranslation},
		{name: "contains SpeakFriend", substr: SpeakFriend},
		{name: "contains welcome message", substr: "Doors of Durin open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, output, tt.substr)
		})
	}
}

// captureReveal runs Reveal() while capturing os.Stdout and returns output.
// It properly restores os.Stdout via t.Cleanup even if test panics.
func captureReveal(t *testing.T) string {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = origStdout
	})

	Reveal()

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return buf.String()
}
