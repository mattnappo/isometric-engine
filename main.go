package main

import (
	"image"
	_ "image/png"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type vec2d struct {
	x int
	y int
}

type tileType int

const (
	blank tileType = iota
	grass tileType = iota
)

const (
	worldSizeX = 10
	worldSizeY = 10
)

var (
	worldSize = vec2d{worldSizeX, worldSizeY}
	tileSize  = vec2d{80, 40}
	origin    = vec2d{6, 1}
	world     [worldSizeX][worldSizeY]tileType
)

// loadPicture loads a picture from memory and returns a pixel picture.
func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

// toScreenSpace takes coordinates from the world space and maps them to
// coordinates in the virtual screen space.
func toScreenSpace(x, y int) pixel.Vec {
	return pixel.V(
		float64(origin.x*tileSize.x+(x-y)*(tileSize.x/2)),
		float64(origin.y*tileSize.y+(x+y)*(tileSize.y/2)),
	)
}

// run is the main game function.
func run() {
	// Create the window config
	cfg := pixelgl.WindowConfig{
		Title:  "@xoreo isometric-engine",
		Bounds: pixel.R(0, 0, float64((worldSize.x+2)*tileSize.x), 700),
	}

	// Create the window itself
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Load the blank tile
	rawBlankTile, err := loadPicture("resources/tiles/blank.png")
	if err != nil {
		panic(err)
	}
	blankTile := pixel.NewSprite(rawBlankTile, rawBlankTile.Bounds())

	// Initialize the world map to blank tiles
	for y, _ := range world {
		for x, _ := range world[y] {
			world[y][x] = blank
		}
	}

	// Clear the screen
	win.Clear(colornames.White)

	// Render all of the tiles, y first to add depth
	for y := 0; y < worldSize.y; y++ {
		for x := 0; x < worldSize.x; x++ {
			// Give 2d coord of where to draw tile onto screen
			screenVec := toScreenSpace(x, y) // Transform the coord to screen space
			switch world[x][y] {
			case blank:
				// Draw the blank tile sprite in the middle of the window
				blankTile.Draw(win, pixel.IM.Moved(screenVec))
				break
			}
		}
	}

	// Update the window while it is still open
	for !win.Closed() {
		win.Update()
	}
}

func main() {
	pixelgl.Run(run) // Set the run function = my run function
}
