package main

import (
	"embed"
	"fmt"
	"image/color"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

//go:embed pieces/*.png
var pieceFiles embed.FS

const (
	squareSize = 90
	boardSize  = 8 * squareSize
	borderSize = 24
	canvasSize = boardSize + 2*borderSize
)

type Game struct {
	selectedRow  int
	selectedCol  int
	hasSelection bool
	board        [8][8]rune
	pieceImages  map[rune]*ebiten.Image
}

func newGame() (*Game, error) {
	game := &Game{pieceImages: make(map[rune]*ebiten.Image)}
	// Uppercase pieces are white, lowercase pieces are black, and zero is empty.
	for _, piece := range "KQBNRPkqbnrp" {
		// The PNG filenames use the hexadecimal value of each piece letter.
		path := fmt.Sprintf("pieces/%x.png", piece)
		file, err := pieceFiles.Open(path)
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", path, err)
		}
		game.pieceImages[piece] = ebiten.NewImageFromImage(img)
	}

	for col, piece := range "rnbqkbnr" {
		game.board[0][col] = piece
		game.board[1][col] = 'p'
		game.board[6][col] = 'P'
		game.board[7][col] = piece - 'a' + 'A'
	}
	return game, nil
}

func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		// Convert window coordinates to coordinates relative to the board.
		x -= borderSize
		y -= borderSize
		if x >= 0 && x < boardSize && y >= 0 && y < boardSize {
			row := y / squareSize
			col := x / squareSize

			if g.hasSelection {
				piece := g.board[g.selectedRow][g.selectedCol]
				g.board[g.selectedRow][g.selectedCol] = 0
				g.board[row][col] = piece
				g.hasSelection = false
			} else if g.board[row][col] != 0 {
				g.selectedCol = col
				g.selectedRow = row
				g.hasSelection = true
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	light := color.RGBA{R: 220, G: 190, B: 155, A: 255}
	dark := color.RGBA{R: 100, G: 75, B: 40, A: 255}
	highlightColor := color.RGBA{R: 175, G: 145, B: 100, A: 255}

	// Draw the frame in layers, then paint the board over its center.
	screen.Fill(color.RGBA{R: 38, G: 26, B: 19, A: 255})
	vector.FillRect(screen, 4, 4, canvasSize-8, canvasSize-8,
		color.RGBA{R: 78, G: 52, B: 36, A: 255}, false)
	vector.FillRect(screen, borderSize-2, borderSize-2, boardSize+4, boardSize+4,
		color.RGBA{R: 161, G: 119, B: 76, A: 255}, false)

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			squareColor := light
			if (row+col)%2 == 1 {
				squareColor = dark
			}
			x := borderSize + col*squareSize
			y := borderSize + row*squareSize
			vector.FillRect(screen, float32(x), float32(y), squareSize, squareSize, squareColor, false)
			if img := g.pieceImages[g.board[row][col]]; img != nil {
				options := &ebiten.DrawImageOptions{}
				offsetX := (squareSize - img.Bounds().Dx()) / 2
				offsetY := (squareSize - img.Bounds().Dy()) / 2
				options.GeoM.Translate(float64(x+offsetX), float64(y+offsetY))
				screen.DrawImage(img, options)
			}
		}
	}

	if g.hasSelection {
		x := borderSize + g.selectedCol*squareSize
		y := borderSize + g.selectedRow*squareSize
		vector.StrokeRect(screen, float32(x)+2, float32(y)+2, squareSize-4, squareSize-4, 4, highlightColor, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return canvasSize, canvasSize
}

func main() {
	ebiten.SetWindowSize(canvasSize, canvasSize)
	ebiten.SetWindowTitle("Chess")

	game, err := newGame()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
