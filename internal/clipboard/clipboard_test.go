package clipboard

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	original := WriteAll
	t.Cleanup(func() { WriteAll = original })

	tests := []struct {
		name         string
		text         string
		mockWriteAll func(string) error
		wantErr      string
		wantCalled   bool
	}{
		{
			name: "success with simple password",
			text: "s3cureP@ss!",
			mockWriteAll: func(s string) error {
				return nil
			},
			wantCalled: true,
		},
		{
			name: "success with long password",
			text: "aVeryLongPassword123!@#WithSymbols",
			mockWriteAll: func(s string) error {
				assert.Equal(t, "aVeryLongPassword123!@#WithSymbols", s)
				return nil
			},
			wantCalled: true,
		},
		{
			name: "clipboard unavailable - xclip not found",
			text: "myPassword123",
			mockWriteAll: func(s string) error {
				return errors.New("clipboard unavailable: xclip not found")
			},
			wantErr: "clipboard unavailable: xclip not found",
		},
		{
			name: "clipboard error with empty mock",
			text: "test123",
			mockWriteAll: func(s string) error {
				return errors.New("No clipboard utility available")
			},
			wantErr: "No clipboard utility available",
		},
		{
			name: "empty text should fail without calling WriteAll",
			text: "",
			mockWriteAll: func(s string) error {
				t.Errorf("WriteAll should not be called for empty text")
				return nil
			},
			wantErr: "cannot copy empty text to clipboard",
		},
		{
			name: "password with spaces and symbols",
			text: "p@ss w0rd! $pecial",
			mockWriteAll: func(s string) error {
				assert.Equal(t, "p@ss w0rd! $pecial", s)
				return nil
			},
			wantCalled: true,
		},
		{
			name: "single word mellon",
			text: "hello-mellon",
			mockWriteAll: func(s string) error {
				assert.Equal(t, "hello-mellon", s)
				return nil
			},
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			assert.NotNil(t, WriteAll)
		})
	}
}
