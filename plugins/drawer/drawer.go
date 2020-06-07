package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"

	"github.com/avoronkov/spil/types"
)

var TypeDrawer = types.Type("drawer")

type Drawer struct {
	filename string
	img      *image.RGBA
}

var _ types.Expr = (*Drawer)(nil)
var _ io.Closer = (*Drawer)(nil)

func NewDrawer(filename string, width, height int) *Drawer {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// fill the canvas with black
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.SetRGBA(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	return &Drawer{
		filename: filename,
		img:      img,
	}
}

func (f *Drawer) String() string {
	return fmt.Sprintf("{drawer: %v}", f.filename)
}

func (f *Drawer) Print(w io.Writer) {
	fmt.Fprintf(w, "%v", f.String())
}

func (f *Drawer) Hash() (string, error) {
	return "", fmt.Errorf("Hashing is not supported for drawer")
}

func (f *Drawer) Type() types.Type {
	return TypeDrawer
}

func (f *Drawer) Close() error {
	file, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, f.img)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

func (f *Drawer) DrawPoint(x, y, r, g, b int) {
	c := color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	f.img.SetRGBA(x, y, c)
}

func (f *Drawer) DrawLine(x0, y0, x1, y1, r, g, b int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	c := color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	if dx >= dy {
		if x0 > x1 {
			x0, y0, x1, y1 = x1, y1, x0, y0
		}
		dx++
		derr := dy + 1
		e := 0
		diry := sign(y1 - y0)
		for x, y := x0, y0; x <= x1; x++ {
			f.img.SetRGBA(x, y, c)
			e += derr
			if e >= dx {
				y += diry
				e -= dx
			}
		}
	} else {
		if y0 > y1 {
			x0, y0, x1, y1 = x1, y1, x0, y0
		}
		dy++
		derr := dx + 1
		dirx := sign(x1 - x0)
		for x, y, e := x0, y0, 0; y <= y1; y++ {
			f.img.SetRGBA(x, y, c)
			e += derr
			if e >= dy {
				x += dirx
				e -= dy
			}
		}
	}
}
