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
	"github.com/golang/geo/r2"
	"golang.org/x/image/colornames"
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
	stone tileType = iota

	stoneEdgeN tileType = iota
	stoneEdgeE tileType = iota
	stoneEdgeS tileType = iota
	stoneEdgeW tileType = iota
)

const (
	worldSizeX = 10
	worldSizeY = 10
)

var (
	worldSize = vec2d{worldSizeX, worldSizeY}
	tileSize  = vec2d{63, 32}
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
		Title: "@xoreo isometric-engine",
		Bounds: pixel.R(
			0,
			0,
			float64((worldSize.x+2)*tileSize.x),
			float64((worldSize.y)*tileSize.x),
		),
	}

	// Create the window itself
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Initialize the sprites
	spriteSheet, err := loadPicture("resources/spritesheet.png")
	if err != nil {
		panic(err)
	}

	var tileSprites [6]*pixel.Sprite

	tileSprites[grass] = pixel.NewSprite(spriteSheet, pixel.R(257, 67, tileSize.x, tileSize.y))
	tileSprites[stone] = pixel.NewSprite(spriteSheet, pixel.R(1, 34, tileSize.x, tileSize.y))

	// Initialize the world map to blank tiles
	for y, _ := range world {
		for x, _ := range world[y] {
			world[y][x] = grass
		}
	}

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
				case grass:
					// Draw the grass tile sprite
					tileSprites[grass].Draw(win, pixel.IM.Moved(screenVec))
					break
				}
			}
		}

		imd := imdraw.New(nil)           // Initialize the mesh
		imd.Color = pixel.RGB(255, 0, 0) // Red

		// Calculate where the point is in relation to the border of the tile
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

		// Calculate the cross products
		dAB := (P.X-A.X)*(B.Y-A.Y) - (P.Y-A.Y)*(B.X-A.X)
		dBC := (P.X-B.X)*(C.Y-B.Y) - (P.Y-B.Y)*(C.X-B.X)
		dCD := (P.X-C.X)*(D.Y-C.Y) - (P.Y-C.Y)*(D.X-C.X)
		dDA := (P.X-D.X)*(A.Y-D.Y) - (P.Y-D.Y)*(A.X-D.X)
		fmt.Printf("dAB: %f\ndBC: %f\ndCD: %f\ndDA: %f\n\n", dAB, dBC, dCD, dDA)

		// Change the cellSpaceCell accordingly
		if dAB < 0 { // Bottom left
			cellSpaceCell.x -= 1
		} else if dBC < 0 { // Top left
			cellSpaceCell.y += 1
		} else if dCD < 0 { // Top right
			cellSpaceCell.x += 1
		} else if dDA < 0 { // Bottom right
			cellSpaceCell.y -= 1
		}

		// Check that the cell is within the board
		if cellSpaceCell.x >= 0 && cellSpaceCell.x < worldSize.x { // Check x bounds
			if cellSpaceCell.y >= 0 && cellSpaceCell.y < worldSize.y { // Check y bounds
				selectedTile.Draw(win, pixel.IM.Moved(
					pointToScreenSpace(cellSpaceCell.x, cellSpaceCell.y),
				)) // Draw the highlighted sprite on the cell
			}
		}

		win.Update() // Update the window
	}
}

func main() {
	pixelgl.Run(run) // Set the run function = my run function
}
