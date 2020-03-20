package main

import (
	"fmt"
	"image"
	_ "image/png"
	"math/rand"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/golang/geo/r2"
	"golang.org/x/image/colornames"
)

type tileType int

const (
	tileTypeCount = 11

	grass1     tileType = iota
	grass2     tileType = iota
	stone      tileType = iota
	selected   tileType = iota
	stoneEdgeN tileType = iota
	stoneEdgeE tileType = iota
	stoneEdgeS tileType = iota
	stoneEdgeW tileType = iota

	log    tileType = iota
	grass3 tileType = iota
	tree   tileType = iota
)

const (
	worldSizeX = 10
	worldSizeY = 10
)

var (
	worldSize = pixel.V(worldSizeX, worldSizeY)
	tileSize  = pixel.V(63, 32)
	origin    = pixel.V(5, 1)
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
func pointToScreenSpace(x, y float64) pixel.Vec {
	return pixel.V(
		(origin.X*tileSize.X+(x-y)*(tileSize.X/2))+tileSize.X/2,
		(origin.Y*tileSize.Y+(x+y)*(tileSize.Y/2))+tileSize.Y/2,
	)
}

// getSprite slices the sprite sheet and returns the proper sprite.
func getSprite(spriteSheet pixel.Picture, row, col float64) *pixel.Sprite {
	minX := col * (tileSize.X + 1)
	minY := row * (tileSize.Y + 1)
	maxX := minX + tileSize.X + 2
	maxY := minY + tileSize.Y + 2
	// fmt.Printf("%f\n%f\n%f\n%f\n\n", minX, minY, maxX, maxY)
	return pixel.NewSprite(spriteSheet, pixel.R(
		minX, minY, maxX, maxY,
	))
}

// getSpriteC slices the sprite with custom coordinates.
func getSpriteC(spriteSheet pixel.Picture, x, y, w, h float64) *pixel.Sprite {
	fmt.Printf("%f\n%f\n%f\n%f\n\n", x, y, x+w, y+h)
	return pixel.NewSprite(spriteSheet, pixel.R(
		x, y, x+w, y+h,
	))
}

// run is the main game function.
func run() {
	// Create the window config
	cfg := pixelgl.WindowConfig{
		Title: "@xoreo isometric-engine",
		Bounds: pixel.R(
			0,
			0,
			(worldSizeX+2)*tileSize.X,
			(worldSizeY)*tileSize.X,
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

	var tileSprites [tileTypeCount + 1]*pixel.Sprite // Init slice of all tile sprites
	tileSprites[grass1] = getSprite(spriteSheet, 2, 4)
	tileSprites[grass2] = getSprite(spriteSheet, 3, 4)
	tileSprites[stone] = getSprite(spriteSheet, 2, 0)
	tileSprites[selected] = getSprite(spriteSheet, 0, 0)
	tileSprites[stoneEdgeN] = getSprite(spriteSheet, 0, 1)
	tileSprites[stoneEdgeS] = getSprite(spriteSheet, 1, 1)
	tileSprites[stoneEdgeE] = getSprite(spriteSheet, 2, 1)
	tileSprites[stoneEdgeW] = getSprite(spriteSheet, 3, 1)

	tileSprites[log] = getSpriteC(spriteSheet, 1, 6, 1, 1)
	tileSprites[grass3] = getSpriteC(spriteSheet, 1, 6, 1, 1)
	tileSprites[tree] = getSpriteC(spriteSheet, 448, 0, 70, 127)

	// Initialize the world map to blank tiles
	for y, _ := range world {
		for x, _ := range world[y] {
			// world[y][x] = tileType(rand.Intn(3))
			r := rand.Intn(3)
			switch r {
			case 0:
				world[y][x] = grass1
				break
			case 1:
				world[y][x] = grass2
				break
			case 2:
				world[y][x] = grass2
				break
			}
		}
	}

	// world[2][2] = tree

	// Main loop
	for !win.Closed() {
		// Clear the screen
		win.Clear(colornames.Lightskyblue)

		mouseVec := win.MousePosition() // Get the position of the mouse
		boardSpaceCell := pixel.V(
			float64(int(mouseVec.X)/int(tileSize.X)), // x position
			float64(int(mouseVec.Y)/int(tileSize.Y)), // y position
		)

		// Map the cell coords in screen space to those in cell space
		cellSpaceCell := pixel.V(
			(boardSpaceCell.Y-origin.Y)+(boardSpaceCell.X-origin.X),
			(boardSpaceCell.Y-origin.Y)-(boardSpaceCell.X-origin.X),
		)

		// Render all of the tiles, y first to add depth
		for x := 0; x < worldSizeX; x++ {
			for y := 0; y < worldSizeY; y++ {
				// Map to screen space
				screenVec := pointToScreenSpace(float64(x), float64(y))

				// Draw the appropriate tile
				tileSprites[world[y][x]].Draw(
					win,
					pixel.IM.Scaled(pixel.ZV, 1.0).Moved(screenVec),
				)
			}
		}

		imd := imdraw.New(nil)           // Initialize the mesh
		imd.Color = pixel.RGB(255, 0, 0) // Red

		// Calculate where the point is in relation to the border of the tile
		tx := tileSize.X
		ty := tileSize.Y
		P := r2.Point{mouseVec.X, mouseVec.Y}
		O := r2.Point{boardSpaceCell.X * tx, boardSpaceCell.Y * ty}
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
		// fmt.Printf("dAB: %f\ndBC: %f\ndCD: %f\ndDA: %f\n\n", dAB, dBC, dCD, dDA)

		// Change the cellSpaceCell accordingly
		if dAB < 0 { // Bottom left
			cellSpaceCell.X -= 1
		} else if dBC < 0 { // Top left
			cellSpaceCell.Y += 1
		} else if dCD < 0 { // Top right
			cellSpaceCell.X += 1
		} else if dDA < 0 { // Bottom right
			cellSpaceCell.Y -= 1
		}

		// Check that the cell is within the board
		if cellSpaceCell.X >= 0 && cellSpaceCell.X < worldSizeX { // Check x bounds
			if cellSpaceCell.Y >= 0 && cellSpaceCell.Y < worldSizeY { // Check y bounds
				tileSprites[selected].Draw(win, pixel.IM.Scaled(pixel.ZV, 1.0).Moved(
					pointToScreenSpace(cellSpaceCell.X, cellSpaceCell.Y),
				)) // Draw the highlighted sprite on the cell
			}
		}

		tileSprites[tree].Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update() // Update the window
	}
}

func main() {
	fmt.Printf("@xoreo's isometric engine\n")
	rand.Seed(time.Now().UTC().UnixNano())
	pixelgl.Run(run) // Set the run function = my run function
}
