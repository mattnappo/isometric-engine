package main

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/geo/r2"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type vec2d struct {
	x int
	y int
}
type vec2df struct {
	x float64
	y float64
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
	origin    = vec2d{5, 1}
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

// pointToScreenSpace takes coordinates from the world space and maps them to
// coordinates in the virtual screen space.
func pointToScreenSpace(x, y int) pixel.Vec {
	return pixel.V(
		float64((origin.x*tileSize.x+(x-y)*(tileSize.x/2))+tileSize.x/2),
		float64((origin.y*tileSize.y+(x+y)*(tileSize.y/2))+tileSize.y/2),
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

	// Initialize the sprites (going to implement a spritesheet soon!)
	rawBlankTile, err := loadPicture("resources/tiles/blank.png")
	if err != nil {
		panic(err)
	}
	rawSelectedTile, err := loadPicture("resources/tiles/selected.png")

	if err != nil {
		panic(err)
	}
	blankTile := pixel.NewSprite(rawBlankTile, rawBlankTile.Bounds())
	selectedTile := pixel.NewSprite(rawSelectedTile, rawSelectedTile.Bounds())

	// Initialize the world map to blank tiles
	for y, _ := range world {
		for x, _ := range world[y] {
			world[y][x] = blank
		}
	}

	// Initialize text rendering
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	text := text.New(pixel.V(200, float64(worldSize.y-200)), atlas)
	fmt.Fprintln(text, "SOME TEXT!")

	// Main loop
	for !win.Closed() {
		// Clear the screen
		win.Clear(colornames.White)

		mouseVec := win.MousePosition() // Get the position of the mouse
		boardSpaceCell := vec2d{
			int(math.Floor(mouseVec.X / float64(tileSize.x))), // x position
			int(math.Floor(mouseVec.Y / float64(tileSize.y))), // y position
		}
		_ = vec2d{
			int(mouseVec.X) % tileSize.x, // x offset
			int(mouseVec.Y) % tileSize.y, // y offset
		}
		// Map the cell coords in screen space to those in cell space
		cellSpaceCell := vec2d{
			(boardSpaceCell.y - origin.y) + (boardSpaceCell.x - origin.x),
			(boardSpaceCell.y - origin.y) - (boardSpaceCell.x - origin.x),
		}

		// Render all of the tiles, y first to add depth
		for y := 0; y < worldSize.y; y++ {
			for x := 0; x < worldSize.x; x++ {
				// Give 2d coord of where to draw tile onto screen
				screenVec := pointToScreenSpace(x, y) // Transform to screen space
				switch world[x][y] {
				case blank:
					// Draw the blank tile sprite in the middle of the window
					blankTile.Draw(win, pixel.IM.Moved(screenVec))
					break
				}
			}
		}

		// fmt.Printf("selected: %d, %d\n", cellSpaceCell.x, cellSpaceCell.y)

		imd := imdraw.New(nil)           // Initialize the mesh
		imd.Color = pixel.RGB(255, 0, 0) // Red

		/*// Square vertices (the square is "wrong" now, but that's fine)
		xpos := float64(boardSpaceCell.x*tileSize.x) - float64(tileSize.x/2)
		ypos := float64(boardSpaceCell.y*tileSize.y) - float64(tileSize.y/2)
		imd.Push(pixel.V(xpos, ypos))
		imd.Push(pixel.V(xpos+float64(tileSize.x), ypos))
		imd.Push(pixel.V(xpos+float64(tileSize.x), ypos+float64(tileSize.y)))
		imd.Push(pixel.V(xpos, ypos+float64(tileSize.y)))
		imd.Push(pixel.V(xpos, ypos))
		imd.Line(1) // Make the polygon*/

		// Calculate where the point is in relation to the border
		tx := float64(tileSize.x)
		ty := float64(tileSize.y)
		P := r2.Point{mouseVec.X, mouseVec.Y}
		O := r2.Point{float64(boardSpaceCell.x) * tx, float64(boardSpaceCell.y) * ty}
		A := r2.Point{
			O.X + tx/2,
			O.Y,
		}
		B := r2.Point{
			O.X,
			O.Y + ty/2,
		}
		C := r2.Point{
			O.X + tx/2,
			O.Y + ty,
		}
		D := r2.Point{
			O.X + tx,
			O.Y + ty/2,
		}

		dAB := (P.X-A.X)*(B.Y-A.Y) - (P.Y-A.Y)*(B.X-A.X)
		dBC := (P.X-B.X)*(C.Y-B.Y) - (P.Y-B.Y)*(C.X-B.X)
		dCD := (P.X-C.X)*(D.Y-C.Y) - (P.Y-C.Y)*(D.X-C.X)
		dDA := (P.X-D.X)*(A.Y-D.Y) - (P.Y-D.Y)*(A.X-D.X)

		fmt.Printf("dAB: %f\ndBC: %f\ndCD: %f\ndDA: %f\n\n", dAB, dBC, dCD, dDA)
		// Change the cellSpaceCell accordingly
		if dAB < 0 { // Bottom left
			fmt.Println("BOTTOM LEFT")
		} else if dBC > 0 { // Top left
			fmt.Println("TOP LEFT")
		} else if dCD > 0 { // Top right
			fmt.Println("TOP RIGHT")
		} else if dDA < 0 { // Bottom right
			fmt.Println("BOTTOM RIGHT")
		} else { // Center
			fmt.Println(" -- CENTER -- ")
		}

		imd.Push(pixel.V(A.X, A.Y))
		imd.Circle(10, 1)
		imd.Push(pixel.V(B.X, B.Y))
		imd.Circle(10, 1)
		imd.Push(pixel.V(C.X, C.Y))
		imd.Circle(10, 1)
		imd.Push(pixel.V(D.X, D.Y))
		imd.Circle(10, 1)
		imd.Draw(win)

		selectedTile.Draw(win, pixel.IM.Moved(
			pointToScreenSpace(cellSpaceCell.x, cellSpaceCell.y),
		)) // Draw the highlighted sprite on the cell

		text.Draw(win, pixel.IM.Scaled(text.Orig, 10))
		win.Update() // Update the window
	}
}

func main() {
	pixelgl.Run(run) // Set the run function = my run function
}
