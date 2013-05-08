package main

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xwindow"
	i3lib "github.com/TShadwell/senbar/i3"
)

const (
	Bar_Height       = 15
	Background_Color = 0x0c0201
	Name             = "senbar"
)

func main() {
	//Connnect to i3
	i3, err := i3lib.Attach()
	if err != nil {
		panic(err)
	}
	//Connect to X
	X, err := xgbutil.NewConn()

	lock := make(chan bool)
	if err != nil {
		panic(err)
		lock <- true
	}
	go func() {
		panic(i3.Listen())
	}()
	op, err := i3.Outputs()
	if err != nil {
		panic(err)
	}

	bars := make([]*xwindow.Window, len(op))

	for i, v := range op {
		if v.Active {
			bars[i] = xwindow.Must(xwindow.Generate(X))
			defer bars[i].Destroy()

			bars[i].Create(
				X.RootWin(),
				int(v.Rect.X),
				int(v.Rect.Y),
				int(v.Rect.Width),
				Bar_Height,
				xproto.CwBackPixel,
				Background_Color,
			)
			bars[i].Map()
			err = ewmh.WmWindowTypeSet(
				X,
				bars[i].Id,
				[]string{"_NET_WM_WINDOW_TYPE_DOCK"},
			)
			if err != nil {
				panic(err)
			}
			err = ewmh.WmStateSet(
				X,
				bars[i].Id,
				[]string{
					"_NET_WM_STATE_ABOVE",
					"_NET_WM_STATE_STICKY",
				},
			)
			if err != nil {
				panic(err)
			}

			err = ewmh.WmNameSet(
				X,
				bars[i].Id,
				Name,
			)

			if err != nil {
				panic(err)
			}

			//Make struts
			if err != nil {
				panic(err)
			}
			err = ewmh.WmStrutPartialSet(
				X,
				bars[i].Id,
				&ewmh.WmStrutPartial{
					Top:       Bar_Height,
					TopStartX: v.Rect.X,
					TopEndX:   v.Rect.X + v.Rect.Width - 1,
				},
			)

			//Look, I have no idea, but Dzen2 does it, okay.
			err = ewmh.WmDesktopSet(
				X,
				bars[i].Id,
				4294967295,
			)
			if err != nil {
				panic(err)
			}

		}
	}
	fmt.Println("Locking...")
	<-lock

}
