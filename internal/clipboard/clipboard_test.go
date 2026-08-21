package clipboard

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	original := WriteAll
	defer func() { WriteAll = original }()

	tests := map[string]struct {
		text         string
		mockWriteAll func(string) error
		wantErr      string
		wantCalled   bool
	}{
		"success with simple password": {
			text: "s3cureP@ss!",
			mockWriteAll: func(s string) error {
				return nil
			},
		},
		"success with long password": {
			text: "aVeryLongPassword123!@#WithSymbols",
			mockWriteAll: func(s string) error {
				assert.Equal(t, "aVeryLongPassword123!@#WithSymbols", s)
				return nil
			},
		},
		"clipboard unavailable - xclip not found": {
			text: "myPassword123",
			mockWriteAll: func(s string) error {
				return errors.New("clipboard unavailable: xclip not found")
			},
			wantErr: "clipboard unavailable: xclip not found",
		},
		"clipboard error with empty mock": {
			text: "test123",
			mockWriteAll: func(s string) error {
				return errors.New("No clipboard utility available")
			},
			wantErr: "No clipboard utility available",
		},
		"empty text should fail without calling WriteAll": {
			text: "",
			mockWriteAll: func(s string) error {
				t.Errorf("WriteAll should not be called for empty text")
				return nil
			},
			wantErr: "cannot copy empty text to clipboard",
		},
		"password with spaces and symbols": {
			text: "p@ss w0rd! $pecial",
			mockWriteAll: func(s string) error {
				assert.Equal(t, "p@ss w0rd! $pecial", s)
				return nil
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			called := false
			origMock := tt.mockWriteAll
			WriteAll = func(s string) error {
				called = true
				if origMock != nil {
					return origMock(s)
				}
				return nil
			}

			err := Copy(tt.text)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
				if tt.text == "" {
					assert.False(t, called, "WriteAll should not be called for empty text")
				}
				return
			}

			require.NoError(t, err)
			assert.True(t, called, "WriteAll should be called")
		})
	}
}

func TestCopy_EmptyText(t *testing.T) {
	original := WriteAll
	defer func() { WriteAll = original }()

	WriteAll = func(s string) error {
		t.Fatal("WriteAll should not be called for empty text")
		return nil
	}

	tests := map[string]struct {
		text    string
		wantErr string
	}{
		"empty string": {
			text:    "",
			wantErr: "cannot copy empty text to clipboard",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := Copy(tt.text)
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestWriteAll_Mockable(t *testing.T) {
	original := WriteAll
	defer func() { WriteAll = original }()

	calledWith := ""
	WriteAll = func(s string) error {
		calledWith = s
		return nil
	}

	err := Copy("hello-mellon")
	require.NoError(t, err)
	assert.Equal(t, "hello-mellon", calledWith)

	// Verify that WriteAll is indeed the mocked function
	assert.NotNil(t, WriteAll)
}
