###Senbar

A taskbar for i3.

To get it, do this:

`go get -u github.com/TShadwell/senbar/senbar`

It also has some optional features that are specific to my laptop, and expect `alsa`, as well as read access to `/dev/input/event0`. If you intend to use it you might need to modify the switch case in `senbar/senbar_laptop.go` to match with your buttons (_input-events_ is useful for finding the appropriate key codes). To install the laptop version, use:
`go get -tags 'laptop' -u github.com/TShadwell/senbar/senbar/`

###The Interesting Bits
This project also includes a pretty good, but incomplete i3 library, a simple asynchronous interface to `/dev/input/eventx`, as well as a native golang implimentation of some dzen gadgets. All docs can be found [here](http://go.pkgdoc.org/github.com/TShadwell/senbar).
