package imgshow

import "sync"

// Configuration for windows.
// Don't create directly, use `imgshow.NewConfig` instead.
type Config struct {
	Width  int
	Height int
	Title  string
}

func NewConfig() Config {
	return Config{
		Width:  800,
		Height: 600,
		Title:  "imgshow",
	}
}

// Get a window for this configuration.
func (c Config) Window() Window {
	return Window{c: c, mu: &sync.Mutex{}}
}
