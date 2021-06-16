package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/orsinium-labs/imgshow"
)

func show() error {
	stream, err := os.Open("lenna.png")
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	img, err := png.Decode(stream)
	if err != nil {
		return fmt.Errorf("decode image: %v", err)
	}

	config := imgshow.NewConfig()
	win := config.Window()
	err = win.Create()
	if err != nil {
		return fmt.Errorf("create window: %v", err)
	}
	err = win.Draw(img)
	if err != nil {
		return fmt.Errorf("draw image: %v", err)
	}
	return nil
}

func main() {
	err := show()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("finished")
}
