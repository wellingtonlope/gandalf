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
)

func TestRun_HelpAndVersion(t *testing.T) {
	tests := map[string]struct {
		args       []string
		wantStdout string
		wantCode   int
	}{
		"help flag": {
			args:       []string{"--help"},
			wantStdout: "gandalf - Magical",
			wantCode:   0,
		},
		"help short": {
			args:       []string{"-h"},
			wantStdout: "gandalf - Magical",
			wantCode:   0,
		},
		"help command": {
			args:       []string{"help"},
			wantStdout: "gandalf - Magical",
			wantCode:   0,
		},
		"version flag": {
			args:       []string{"--version"},
			wantStdout: "gandalf",
			wantCode:   0,
		},
		"version short": {
			args:       []string{"-v"},
			wantStdout: "gandalf",
			wantCode:   0,
		},
		"version command": {
			args:       []string{"version"},
			wantStdout: "gandalf",
			wantCode:   0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr, nil)
			assert.Equal(t, tt.wantCode, code)
			assert.Contains(t, stdout.String(), tt.wantStdout)
		})
	}
}

func TestRun_Generate(t *testing.T) {
	tests := map[string]struct {
		args     []string
		wantCode int
		wantLen  int
		wantErr  string
	}{
		"default generates 16": {
			args:     []string{},
			wantCode: 0,
			wantLen:  16,
		},
		"custom length 12": {
			args:     []string{"-l", "12"},
			wantCode: 0,
			wantLen:  12,
		},
		"long length 32": {
			args:     []string{"--length", "32"},
			wantCode: 0,
			wantLen:  32,
		},
		"count 3": {
			args:     []string{"-c", "3"},
			wantCode: 0,
		},
		"invalid length zero": {
			args:     []string{"-l", "0"},
			wantCode: 1,
			wantErr:  "length must be greater than zero",
		},
		"invalid length too large": {
			args:     []string{"-l", "300"},
			wantCode: 1,
			wantErr:  "maximum length is 256",
		},
		"no charset": {
			args:     []string{"--no-upper", "--no-lower", "--no-digits", "--no-symbols"},
			wantCode: 1,
			wantErr:  "at least one character set",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr, nil)
			assert.Equal(t, tt.wantCode, code)
			if tt.wantCode == 0 {
				out := strings.TrimSpace(stdout.String())
				if tt.wantLen > 0 {
					// For default case with single password, check length
					lines := strings.Split(out, "\n")
					// When count not specified, should be 1 password
					if strings.Contains(name, "count 3") {
						assert.Len(t, lines, 3)
					} else {
						require.Len(t, lines, 1)
						assert.Len(t, lines[0], tt.wantLen)
					}
				}
				assert.Empty(t, stderr.String())
			} else {
				assert.Contains(t, stderr.String(), tt.wantErr)
			}
		})
	}
}

func TestRun_CopyFlag(t *testing.T) {
	tests := map[string]struct {
		args            []string
		mockCopy        func(string) error
		wantCode        int
		wantCopy        bool
		wantCopied      string
		wantStderr      string
		wantStdoutLines int
	}{
		"copy flag success": {
			args:            []string{"--copy", "-l", "10"},
			mockCopy:        func(s string) error { return nil },
			wantCode:        0,
			wantCopy:        true,
			wantStderr:      "Copied to clipboard",
			wantStdoutLines: 1,
		},
		"copy shorthand -x": {
			args:            []string{"-x", "-l", "10"},
			mockCopy:        func(s string) error { return nil },
			wantCode:        0,
			wantCopy:        true,
			wantStderr:      "Copied to clipboard",
			wantStdoutLines: 1,
		},
		"copy shorthand -x with error": {
			args: []string{"-x", "-l", "12"},
			mockCopy: func(s string) error {
				return errors.New("clipboard unavailable")
			},
			wantCode:        0,
			wantCopy:        true,
			wantStderr:      "Warning: could not copy to clipboard",
			wantStdoutLines: 1,
		},
		"copy flag success with short password": {
			args:            []string{"--copy", "-l", "8"},
			mockCopy:        func(s string) error { return nil },
			wantCode:        0,
			wantCopy:        true,
			wantStdoutLines: 1,
		},
		"copy flag with error should still succeed and warn": {
			args: []string{"--copy", "-l", "12"},
			mockCopy: func(s string) error {
				return errors.New("clipboard unavailable")
			},
			wantCode:        0,
			wantCopy:        true,
			wantStderr:      "Warning: could not copy to clipboard",
			wantStdoutLines: 1,
		},
		"copy with multiple passwords copies all": {
			args:            []string{"--copy", "-c", "3", "-l", "10"},
			mockCopy:        func(s string) error { return nil },
			wantCode:        0,
			wantCopy:        true,
			wantStderr:      "Copied to clipboard",
			wantStdoutLines: 3,
		},
		"without copy flag should not call clipboard": {
			args: []string{"-l", "10"},
			mockCopy: func(s string) error {
				t.Errorf("clipboard should not be called")
				return nil
			},
			wantCode:        0,
			wantCopy:        false,
			wantStdoutLines: 1,
		},
		"copy with invalid opts should not call clipboard": {
			args: []string{"--copy", "-l", "0"},
			mockCopy: func(s string) error {
				t.Errorf("clipboard should not be called on error")
				return nil
			},
			wantCode: 1,
			wantCopy: false,
		},
		"copy with nil func should not panic": {
			args:            []string{"--copy", "-l", "10"},
			mockCopy:        nil,
			wantCode:        0,
			wantCopy:        false,
			wantStdoutLines: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
					// Verify copied text is all passwords joined by newline
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
	// Verify that --copy flag is documented in help
	var stdout bytes.Buffer
	code := run([]string{"--help"}, &stdout, io.Discard, nil)
	require.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "--copy")
	assert.Contains(t, stdout.String(), "-x, --copy")
	assert.Contains(t, stdout.String(), "clipboard")
}

func TestRun_Mellon(t *testing.T) {
	// mellon.Reveal writes directly to os.Stdout, so we need to capture it
	tests := map[string]struct {
		args []string
	}{
		"mellon":           {args: []string{"mellon"}},
		"friend":           {args: []string{"friend"}},
		"MELLON uppercase": {args: []string{"MELLON"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Capture os.Stdout
			origStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr, nil)

			w.Close()
			os.Stdout = origStdout
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			assert.Equal(t, 0, code)
			// Reveal prints to os.Stdout, not to stdout buffer
			assert.Contains(t, output, "Mellon")
			assert.Empty(t, stdout.String())
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
