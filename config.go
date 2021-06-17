package imgshow

import "sync"

// Configuration for windows.
type Config struct {
	Width     int
	Height    int
	Title     string
	generated bool
}

func NewConfig() Config {
	return Config{
		Width:     800,
		Height:    600,
		Title:     "imgshow",
		generated: true,
	}
}

// Get a window for this configuration.
func (c Config) Window() *Window {
	return &Window{c: c, mu: &sync.Mutex{}}
}
