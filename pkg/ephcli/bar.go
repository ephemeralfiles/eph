package ephcli

import "github.com/schollz/progressbar/v3"

func (c *ClientEphemeralfiles) InitProgressBar(msg string, totalSize int64) {
	c.bar = progressbar.NewOptions64(totalSize,
		progressbar.OptionClearOnFinish(),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(DefaultBarWidth),
		progressbar.OptionSetVisibility(!c.noProgressBar),
		progressbar.OptionSetDescription(msg))
}

func (c *ClientEphemeralfiles) CloseProgressBar() {
	_ = c.bar.Clear()
	c.bar.Close()
}
