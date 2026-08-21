package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wellingtonlope/gandalf/internal/generator"
)

func TestRun_HelpAndVersion(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStdout string
		wantCode   int
	}{
		{name: "help flag", args: []string{"--help"}, wantStdout: "gandalf - Magical", wantCode: 0},
		{name: "help short", args: []string{"-h"}, wantStdout: "gandalf - Magical", wantCode: 0},
		{name: "help command", args: []string{"help"}, wantStdout: "gandalf - Magical", wantCode: 0},
		{name: "Help case insensitive", args: []string{"HELP"}, wantStdout: "gandalf - Magical", wantCode: 0},
		{name: "version flag", args: []string{"--version"}, wantStdout: "gandalf", wantCode: 0},
		{name: "version short", args: []string{"-v"}, wantStdout: "gandalf", wantCode: 0},
		{name: "version command", args: []string{"version"}, wantStdout: "gandalf", wantCode: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr, nil)
			assert.Equal(t, tt.wantCode, code)
			assert.Contains(t, stdout.String(), tt.wantStdout)
		})
	}
}

func TestRun_Generate(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantLen  int
		wantErr  string
	}{
		{name: "default generates 16", args: []string{}, wantCode: 0, wantLen: 16},
		{name: "custom length 12", args: []string{"-l", "12"}, wantCode: 0, wantLen: 12},
		{name: "long length 32", args: []string{"--length", "32"}, wantCode: 0, wantLen: 32},
		{name: "count 3", args: []string{"-c", "3"}, wantCode: 0},
		{name: "invalid length zero", args: []string{"-l", "0"}, wantCode: 1, wantErr: "length must be greater than zero"},
		{name: "invalid length too large", args: []string{"-l", "300"}, wantCode: 1, wantErr: "maximum length is 256"},
		{name: "no charset", args: []string{"--no-upper", "--no-lower", "--no-digits", "--no-symbols"}, wantCode: 1, wantErr: "at least one character set"},
		{name: "length less than requiredCount - 2 chars with 4 sets", args: []string{"-l", "2"}, wantCode: 1, wantErr: "length must be at least 4"},
		{name: "length 3 with 4 sets should fail", args: []string{"-l", "3"}, wantCode: 1, wantErr: "length must be at least 4"},
		{name: "length 2 with 2 sets should pass", args: []string{"-l", "2", "--no-digits", "--no-symbols"}, wantCode: 0, wantLen: 2},
		{name: "invalid count zero via CLI", args: []string{"-c", "0"}, wantCode: 1, wantErr: "count must be greater than zero"},
		{name: "invalid count negative via CLI", args: []string{"-c", "-5"}, wantCode: 1, wantErr: "count must be greater than zero"},
		{name: "flag parse error - invalid length value", args: []string{"-l", "abc"}, wantCode: 1, wantErr: "invalid value"},
		{name: "flag parse error - unknown flag", args: []string{"--unknown"}, wantCode: 1, wantErr: "flag provided but not defined"},
		{name: "flag parse error - missing value", args: []string{"-l"}, wantCode: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr, nil)
			assert.Equal(t, tt.wantCode, code)
			if tt.wantCode == 0 {
				out := strings.TrimSpace(stdout.String())
				if tt.wantLen > 0 {
					lines := strings.Split(out, "\n")
					if strings.Contains(tt.name, "count 3") {
						assert.Len(t, lines, 3)
					} else {
						require.Len(t, lines, 1)
						assert.Len(t, lines[0], tt.wantLen)
					}
				} else if strings.Contains(tt.name, "count 3") {
					lines := strings.Split(out, "\n")
					assert.Len(t, lines, 3)
				}
				assert.Empty(t, stderr.String())
			} else {
				if tt.wantErr != "" {
					assert.Contains(t, stderr.String(), tt.wantErr)
				}
				// All CLI validation errors should show hint.
				if strings.Contains(tt.wantErr, "length") || strings.Contains(tt.wantErr, "character set") || strings.Contains(tt.wantErr, "count") {
					assert.Contains(t, stderr.String(), "Hint: try 'gandalf --help'")
				}
			}
		})
	}
}

func TestRun_CharsetFilteringViaCLI(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		shouldContain    string
		shouldNotContain string
	}{
		{name: "no symbols via CLI", args: []string{"--no-symbols", "-l", "20"}, shouldContain: generator.LowerLetters + generator.UpperLetters + generator.Digits, shouldNotContain: generator.Symbols},
		{name: "no digits no symbols", args: []string{"--no-digits", "--no-symbols", "-l", "20"}, shouldContain: generator.LowerLetters + generator.UpperLetters, shouldNotContain: generator.Digits + generator.Symbols},
		{name: "only lower via no flags", args: []string{"--no-upper", "--no-digits", "--no-symbols", "-l", "20"}, shouldContain: generator.LowerLetters, shouldNotContain: generator.UpperLetters + generator.Digits + generator.Symbols},
		{name: "only digits", args: []string{"--no-lower", "--no-upper", "--no-symbols", "-l", "20"}, shouldContain: generator.Digits, shouldNotContain: generator.LowerLetters + generator.UpperLetters + generator.Symbols},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for iter := 0; iter < 20; iter++ {
				var stdout, stderr bytes.Buffer
				code := run(tt.args, &stdout, &stderr, nil)
				require.Equal(t, 0, code, "iteration %d stderr: %s", iter, stderr.String())
				pwd := strings.TrimSpace(stdout.String())
				lines := strings.Split(pwd, "\n")
				require.Len(t, lines, 1)
				for _, ch := range lines[0] {
					assert.False(t, strings.ContainsRune(tt.shouldNotContain, ch), "iter %d char %q should be excluded", iter, ch)
					assert.True(t, strings.ContainsRune(tt.shouldContain, ch), "iter %d char %q not in allowed set", iter, ch)
				}
			}
		})
	}
}

func TestRun_CopyFlag(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		mockCopy        func(string) error
		wantCode        int
		wantCopy        bool
		wantStderr      string
		wantStdoutLines int
	}{
		{name: "copy flag success", args: []string{"--copy", "-l", "10"}, mockCopy: func(s string) error { return nil }, wantCode: 0, wantCopy: true, wantStderr: "Copied to clipboard", wantStdoutLines: 1},
		{name: "copy shorthand -x", args: []string{"-x", "-l", "10"}, mockCopy: func(s string) error { return nil }, wantCode: 0, wantCopy: true, wantStderr: "Copied to clipboard", wantStdoutLines: 1},
		{name: "copy with error should still succeed and warn", args: []string{"--copy", "-l", "12"}, mockCopy: func(s string) error { return errors.New("clipboard unavailable") }, wantCode: 0, wantCopy: true, wantStderr: "Warning: could not copy to clipboard", wantStdoutLines: 1},
		{name: "copy with multiple passwords copies all", args: []string{"--copy", "-c", "3", "-l", "10"}, mockCopy: func(s string) error { return nil }, wantCode: 0, wantCopy: true, wantStderr: "Copied to clipboard", wantStdoutLines: 3},
		{name: "without copy flag should not call clipboard", args: []string{"-l", "10"}, mockCopy: func(s string) error { t.Errorf("clipboard should not be called"); return nil }, wantCode: 0, wantCopy: false, wantStdoutLines: 1},
		{name: "copy with invalid opts should not call clipboard", args: []string{"--copy", "-l", "0"}, mockCopy: func(s string) error { t.Errorf("clipboard should not be called on error"); return nil }, wantCode: 1, wantCopy: false},
		{name: "copy with nil func should not panic", args: []string{"--copy", "-l", "10"}, mockCopy: nil, wantCode: 0, wantCopy: false, wantStdoutLines: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			called := false
			var copiedText string
			var copyFn func(string) error
			if tt.mockCopy != nil {
				copyFn = func(s string) error {
					called = true
					copiedText = s
					return tt.mockCopy(s)
				}
			} else {
				copyFn = nil
			}

			code := run(tt.args, &stdout, &stderr, copyFn)
			assert.Equal(t, tt.wantCode, code)

			if tt.wantCopy {
				assert.True(t, called, "clipboard should be called")
				if tt.wantCode == 0 {
					assert.NotEmpty(t, copiedText)
					expected := strings.TrimSpace(stdout.String())
					assert.Equal(t, expected, copiedText)
				}
				if tt.wantStderr != "" {
					assert.Contains(t, stderr.String(), tt.wantStderr)
				}
			} else {
				assert.False(t, called, "clipboard should not be called")
			}

			if tt.wantCode == 0 {
				lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
				assert.Len(t, lines, tt.wantStdoutLines)
			}
		})
	}
}

func TestRun_CopyFlag_Integration(t *testing.T) {
	var stdout bytes.Buffer
	code := run([]string{"--help"}, &stdout, io.Discard, nil)
	require.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "--copy")
	assert.Contains(t, stdout.String(), "-x, --copy")
	assert.Contains(t, stdout.String(), "clipboard")
}

func TestRun_Mellon(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		shouldBeMellon bool
	}{
		{name: "mellon", args: []string{"mellon"}, shouldBeMellon: true},
		{name: "friend", args: []string{"friend"}, shouldBeMellon: true},
		{name: "MELLON uppercase", args: []string{"MELLON"}, shouldBeMellon: true},
		{name: "FRIEND uppercase", args: []string{"FRIEND"}, shouldBeMellon: true},
		{name: "mellon with leading/trailing spaces", args: []string{"  mellon  "}, shouldBeMellon: true},
		{name: "friend with spaces and mixed case", args: []string{"  FrIeNd  "}, shouldBeMellon: true},
		{name: "amigo should NOT be mellon", args: []string{"amigo"}, shouldBeMellon: false},
		{name: "partial should not be mellon", args: []string{"mell"}, shouldBeMellon: false},
		{name: "gandalf word not mellon", args: []string{"gandalf"}, shouldBeMellon: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureRunWithMellon(t, tt.args)

			if tt.shouldBeMellon {
				assert.Contains(t, output.capturedStdout, "Mellon")
				assert.Empty(t, output.stdoutBuf)
			} else {
				// Non-mellon should not print Doors, should generate password via stdoutBuf.
				assert.NotContains(t, output.capturedStdout, "Mellon")
				// For amigo/gandalf etc, it falls through to password generation (default length 16).
				// We verify it did NOT trigger mellon path.
				if tt.args[0] == "amigo" || tt.args[0] == "gandalf" || tt.args[0] == "mell" {
					assert.NotEmpty(t, output.stdoutBuf, "non-mellon arg should generate password")
				}
			}
			assert.Equal(t, 0, output.code)
		})
	}
}

func TestRun_HintOnError(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "zero length", args: []string{"-l", "0"}},
		{name: "too large", args: []string{"-l", "500"}},
		{name: "no charset", args: []string{"--no-lower", "--no-upper", "--no-digits", "--no-symbols"}},
		{name: "count zero", args: []string{"-c", "0"}},
		{name: "length less than sets", args: []string{"-l", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr, nil)
			assert.Equal(t, 1, code)
			assert.Contains(t, stderr.String(), "Error:")
			assert.Contains(t, stderr.String(), "Hint: try 'gandalf --help'")
			_ = stdout
		})
	}
}

func TestPrintUsage_ContainsCopy(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	assert.Contains(t, buf.String(), "--copy")
	assert.Contains(t, buf.String(), "-x, --copy")
	assert.Contains(t, buf.String(), "clipboard")
	assert.Contains(t, buf.String(), "gandalf - Magical")
}

// captureRunWithMellon captures os.Stdout during run and returns both
// the os.Stdout capture and the stdout buffer passed to run.
type runCapture struct {
	capturedStdout string
	stdoutBuf      string
	stderrBuf      string
	code           int
}

func captureRunWithMellon(t *testing.T, args []string) runCapture {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr, nil)

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	return runCapture{
		capturedStdout: buf.String(),
		stdoutBuf:      stdout.String(),
		stderrBuf:      stderr.String(),
		code:           code,
	}
}
