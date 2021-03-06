package imgshow

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"
	"log"
	"sync"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/nfnt/resize"
)

type Window struct {
	c  Config
	mu *sync.Mutex

	w    *xwindow.Window
	x    *xgbutil.XUtil
	img  image.Image
	ximg *xgraphics.Image
}

// Create a new empty window.
func (w *Window) Create() error {
	if w.mu == nil {
		return errors.New("Window must be created using Config.Window")
	}
	if !w.c.generated {
		return errors.New("Config must be created using NewConfig")
	}

	var err error
	xgb.Logger.SetOutput(ioutil.Discard)
	xgbutil.Logger.SetOutput(ioutil.Discard)

	w.x, err = xgbutil.NewConn()
	if err != nil {
		return fmt.Errorf("connect to X server: %v", err)
	}

	keybind.Initialize(w.x)
	mousebind.Initialize(w.x)

	w.w, err = xwindow.Generate(w.x)
	if err != nil {
		return fmt.Errorf("generate X window: %v", err)
	}

	w.w.Create(w.x.RootWin(), 0, 0, w.c.Width, w.c.Height, xproto.CwBackPixel, 0x000000)
	w.w.Change(xproto.CwBackingStore, xproto.BackingStoreAlways)

	w.w.WMGracefulClose(func(xwin *xwindow.Window) {
		xevent.Detach(xwin.X, xwin.Id)
		keybind.Detach(xwin.X, xwin.Id)
		mousebind.Detach(xwin.X, xwin.Id)
		xwin.Destroy()
		xevent.Quit(xwin.X)
	})

	err = ewmh.WmWindowTypeSet(w.x, w.w.Id, []string{"_NET_WM_WINDOW_TYPE_DIALOG"})
	if err != nil {
		return fmt.Errorf("set window type: %v", err)
	}

	err = w.w.Listen(
		xproto.EventMaskKeyPress,
		xproto.EventMaskButtonRelease,
		xproto.EventMaskStructureNotify,
		xproto.EventMaskExposure,
	)
	if err != nil {
		return fmt.Errorf("listen for events: %v", err)
	}

	err = ewmh.WmNameSet(w.x, w.w.Id, w.c.Title)
	if err != nil {
		return fmt.Errorf("set window name: %v", err)
	}

	w.watchInit()
	w.watchConfigure()

	return nil
}

// Destroy the window
func (w *Window) Destroy() {
	w.w.Destroy()
}

// Render the image and wait for the window to be closed.
func (w *Window) Render() {
	w.w.Map()
	xevent.Main(w.x)
}

func (w *Window) newImage() error {
	rect, err := w.w.Geometry()
	if err != nil {
		return fmt.Errorf("get window geometry: %v", err)
	}
	w.ximg = xgraphics.New(w.x, image.Rect(0, 0, rect.Width(), rect.Height()))
	return nil
}

func (w *Window) watchInit() {
	cbExp := xevent.ExposeFun(func(xu *xgbutil.XUtil, e xevent.ExposeEvent) {
		if w.ximg == nil {
			return
		}
		if e.ExposeEvent.Count == 0 {
			w.ximg.XDraw()
			w.ximg.XExpPaint(w.w.Id, 0, 0)
		}
	})
	cbExp.Connect(w.x, w.w.Id)
}

func (w *Window) watchConfigure() {
	cbCfg := xevent.ConfigureNotifyFun(func(xu *xgbutil.XUtil, e xevent.ConfigureNotifyEvent) {
		if w.ximg == nil || w.img == nil {
			return
		}
		xrect := w.ximg.Bounds()
		if xrect.Dx() == int(e.Width) && xrect.Dy() == int(e.Height) {
			return
		}
		err := w.Draw(w.img)
		if err != nil {
			err = fmt.Errorf("draw image: %v", err)
			log.Println(err)
		}
	})
	cbCfg.Connect(w.x, w.w.Id)
}

// Draw the image on window.
// `Render` must be called after to actually render the image.
func (w *Window) Draw(img image.Image) error {
	if w.w == nil {
		return errors.New("Window.Create must be called before Window.Draw")
	}

	var err error
	w.mu.Lock()
	defer w.mu.Unlock()

	w.img = img

	// prepare canvas
	err = w.newImage()
	if err != nil {
		return fmt.Errorf("new image: %v", err)
	}

	// resize image
	dxCanvas := w.ximg.Bounds().Dx()
	dyCanvas := w.ximg.Bounds().Dy()
	ratioCanvas := float64(dxCanvas) / float64(dyCanvas)
	dxImage := img.Bounds().Dx()
	dyImage := img.Bounds().Dy()
	if dxImage > dxCanvas || dyImage > dyCanvas {
		ratioImage := float64(dxImage) / float64(dyImage)
		dx := 0
		dy := 0
		if ratioImage > ratioCanvas {
			dx = dxCanvas
		} else {
			dy = dyCanvas
		}
		img = resize.Resize(uint(dx), uint(dy), img, resize.NearestNeighbor)
	}

	// apply image to canvas
	offsetX := (dxCanvas - img.Bounds().Max.X) / 2
	offsetY := (dyCanvas - img.Bounds().Max.Y) / 2
	rect := img.Bounds().Add(image.Pt(offsetX, offsetY))
	draw.Draw(w.ximg, rect, img, image.Point{}, draw.Over)

	// draw canvas in window
	err = w.ximg.CreatePixmap()
	if err != nil {
		return fmt.Errorf("create pixmap: %v", err)
	}
	w.ximg.XDraw()
	w.ximg.XExpPaint(w.w.Id, 0, 0)
	return nil
}
