package imgshow

import (
	"fmt"
	"image"
)

// Create a new window, draw the image, and wait until the window is closed.
func Show(img image.Image) error {
	config := NewConfig()
	config.Width = img.Bounds().Dx()
	config.Height = img.Bounds().Dy()

	win := config.Window()
	err := win.Create()
	if err != nil {
		return fmt.Errorf("create window: %v", err)
	}
	defer win.Destroy()

	err = win.Draw(img)
	if err != nil {
		return fmt.Errorf("draw image: %v", err)
	}
	win.Render()
	if err != nil {
		return fmt.Errorf("render image: %v", err)
	}
	return nil
}
