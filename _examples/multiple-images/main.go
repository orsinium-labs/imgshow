package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/orsinium-labs/imgshow"
)

func sendImages(win imgshow.Window, imgs ...image.Image) {
	for {
		for _, img := range imgs {
			err := win.Draw(img)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Second)
		}
	}
}

func show() error {
	stream, err := os.Open("../lenna.png")
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	img1, err := png.Decode(stream)
	if err != nil {
		return fmt.Errorf("decode image: %v", err)
	}

	stream, err = os.Open("../lenna-bw.png")
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	img2, err := png.Decode(stream)
	if err != nil {
		return fmt.Errorf("decode image: %v", err)
	}

	win := imgshow.NewConfig().Window()
	err = win.Create()
	if err != nil {
		return fmt.Errorf("create window: %v", err)
	}
	defer win.Destroy()
	go sendImages(win, img1, img2)
	win.Render()
	if err != nil {
		return fmt.Errorf("render image: %v", err)
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
