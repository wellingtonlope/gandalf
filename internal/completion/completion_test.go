package completion

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZshScript(t *testing.T) {
	script := ZshScript()
	require.NotEmpty(t, script)

	tests := []struct {
		name   string
		substr string
	}{
		{name: "compdef header", substr: "#compdef gandalf"},
		{name: "function definition", substr: "_gandalf()"},
		{name: "compdef call", substr: "compdef _gandalf gandalf"},
		{name: "check subcommand", substr: "'check:"},
		{name: "completion subcommand", substr: "'completion:"},
		{name: "mellon subcommand", substr: "'mellon:"},
		{name: "friend subcommand", substr: "'friend:"},
		{name: "help subcommand", substr: "'help:"},
		{name: "version subcommand", substr: "'version:"},
		{name: "zsh shell in completion arg", substr: "'zsh:"},
		{name: "describe", substr: "_describe"},
		{name: "arguments", substr: "_arguments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, script, tt.substr)
		})
	}

	// Ensure only subcommands are completed, not flags (as requested).
	assert.NotContains(t, script, "--length")
	assert.NotContains(t, script, "--no-upper")
	assert.NotContains(t, script, "-l")
}

func TestZshScript_Deterministic(t *testing.T) {
	// Script should be deterministic.
	assert.Equal(t, ZshScript(), ZshScript())
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		want  bool
	}{
		{name: "zsh supported", shell: "zsh", want: true},
		{name: "bash not supported per request", shell: "bash", want: false},
		{name: "fish not supported", shell: "fish", want: false},
		{name: "empty not supported", shell: "", want: false},
		{name: "ZSH case sensitive false", shell: "ZSH", want: false},
		{name: "zsh lowercase exact true", shell: "zsh", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSupported(tt.shell))
		})
	}
}

func TestSupportedShells(t *testing.T) {
	require.NotEmpty(t, SupportedShells)
	assert.Contains(t, SupportedShells, "zsh")
	// Only zsh per user request "A, apenas subcomandos" for zsh.
	assert.Len(t, SupportedShells, 1)
	for _, s := range SupportedShells {
		assert.Equal(t, strings.ToLower(s), s)
	}
}
