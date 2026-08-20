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
	tests := map[string]struct {
		input string
		want  bool
	}{
		"elvish mellon lowercase": {
			input: "mellon",
			want:  true,
		},
		"english friend lowercase": {
			input: "friend",
			want:  true,
		},
		"uppercase mellon - exact match false": {
			input: "MELLON",
			want:  false,
		},
		"uppercase friend - exact match false": {
			input: "FRIEND",
			want:  false,
		},
		"empty string": {
			input: "",
			want:  false,
		},
		"random word": {
			input: "gandalf",
			want:  false,
		},
		"mellon with space": {
			input: " mellon",
			want:  false,
		},
		"friend with trailing space": {
			input: "friend ",
			want:  false,
		},
		"amigo old portuguese should be false now": {
			input: "amigo",
			want:  false,
		},
		"partial mellon": {
			input: "mell",
			want:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsMellon(tt.input))
		})
	}
}

func TestConstants(t *testing.T) {
	tests := map[string]struct {
		value string
		want  string
	}{
		"Elvish": {
			value: Elvish,
			want:  "mellon",
		},
		"Translation": {
			value: Translation,
			want:  "friend",
		},
		"SpeakFriend": {
			value: SpeakFriend,
			want:  "Speak, friend, and enter.",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value)
			assert.NotEmpty(t, tt.value)
		})
	}

	assert.NotEmpty(t, PortalInscription)
	assert.NotEmpty(t, PortalTranslation)
	assert.Contains(t, PortalInscription, "mellon")
	assert.Contains(t, PortalTranslation, "Speak, friend, and enter")
}

func TestReveal_Output(t *testing.T) {
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Reveal()

	require.NoError(t, w.Close())
	os.Stdout = origStdout

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	tests := map[string]struct {
		substr string
	}{
		"contains Mellon header": {
			substr: "Mellon",
		},
		"contains PortalInscription": {
			substr: PortalInscription,
		},
		"contains PortalTranslation": {
			substr: PortalTranslation,
		},
		"contains SpeakFriend": {
			substr: SpeakFriend,
		},
		"contains welcome message": {
			substr: "Doors of Durin open",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, output, tt.substr)
		})
	}
}
