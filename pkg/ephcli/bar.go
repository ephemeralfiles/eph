// Package ephcli provides progress bar functionality for file operations.
package ephcli

import "github.com/schollz/progressbar/v3"

// InitProgressBar initializes a progress bar with the given message and total size.
func (c *ClientEphemeralfiles) InitProgressBar(msg string, totalSize int64) {
	c.bar = progressbar.NewOptions64(totalSize,
		progressbar.OptionClearOnFinish(),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(DefaultBarWidth),
		progressbar.OptionSetVisibility(!c.noProgressBar),
		progressbar.OptionSetDescription(msg))
}

// CloseProgressBar clears and closes the progress bar.
func (c *ClientEphemeralfiles) CloseProgressBar() {
	_ = c.bar.Clear()
	_ = c.bar.Close()
}
