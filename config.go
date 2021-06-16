package imgshow

import "sync"

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

func (c Config) Window() Window {
	return Window{c: c, mu: &sync.Mutex{}}
}
