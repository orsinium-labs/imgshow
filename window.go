package imgshow

import (
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/shm"
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
	c   Config
	w   *xwindow.Window
	x   *xgbutil.XUtil
	shm bool
}

func (w *Window) Create() error {
	var err error
	xgb.Logger.SetOutput(ioutil.Discard)
	xgbutil.Logger.SetOutput(ioutil.Discard)

	w.x, err = xgbutil.NewConn()
	if err != nil {
		return fmt.Errorf("connect to X server: %v", err)
	}

	w.shm = true
	err = shm.Init(w.x.Conn())
	if err != nil {
		w.shm = false
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

	return nil
}

func (w *Window) Destroy() error {
	w.w.Destroy()
	return nil
}

func (w *Window) Draw(img image.Image) error {
	ximg, err := w.newImage()
	if err != nil {
		return fmt.Errorf("new image: %v", err)
	}

	// resize image
	img = resize.Resize(0, uint(w.c.Height), img, resize.NearestNeighbor)

	// apply image to X-image
	xrect, err := w.w.Geometry()
	if err != nil {
		return fmt.Errorf("get window geometry: %v", err)
	}
	offset := (xrect.Width() - img.Bounds().Max.X) / 2
	rect := img.Bounds().Add(image.Pt(offset, 0))
	draw.Draw(ximg, rect, img, image.Point{}, draw.Over)

	// draw X-image in window
	err = ximg.CreatePixmap()
	if err != nil {
		return fmt.Errorf("create pixmap: %v", err)
	}

	cbExp := xevent.ExposeFun(func(xu *xgbutil.XUtil, e xevent.ExposeEvent) {
		if e.ExposeEvent.Count == 0 {
			ximg.XDraw()
			ximg.XExpPaint(w.w.Id, 0, 0)
		}
	})
	cbExp.Connect(w.x, w.w.Id)

	w.w.Map()
	xevent.Main(w.x)

	return nil
}

func (w *Window) newImage() (*xgraphics.Image, error) {
	rect, err := w.w.Geometry()
	if err != nil {
		return nil, fmt.Errorf("get window geometry: %v", err)
	}
	img := xgraphics.New(w.x, image.Rect(0, 0, rect.Width(), rect.Height()))
	return img, nil
}
