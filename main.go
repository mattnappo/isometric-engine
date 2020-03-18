package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func run() {
	// Create the window config
	cfg := pixelgl.WindowConfig{
		Title:  "@xoreo isometric-engine",
		Bounds: pixel.R(0, 0, 700, 700),
	}

	// Create the window itself
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Clear(colornames.Skyblue)

	// Update the window while it is still open
	for !win.Closed() {
		win.Update()
	}
}

func main() {
	pixelgl.Run(run) // Set the run function = my run function
}
