package clipboard

import (
	"errors"

	"github.com/atotto/clipboard"
)

// WriteAll is exposed for testing (allows mocking).
// In production it points to clipboard.WriteAll.
var WriteAll = clipboard.WriteAll

// Copy copies text to the system clipboard.
// It returns an error if the clipboard is unavailable.
func Copy(text string) error {
	if text == "" {
		return errors.New("cannot copy empty text to clipboard")
	}
	return WriteAll(text)
}
