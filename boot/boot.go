package boot

import (
	"math/rand"

	"github.com/gogf/gf/errors/gerror"
)

// CreateMap generates a minesweeper board with the given dimensions and number of mines.
// Parameters:
//
//	x - number of rows
//	y - number of columns
//	n - number of mines
//
// Returns:
//
//	A pointer to a 2D boolean array, where true indicates a mine and false indicates an empty cell.
//	An error if the number of mines exceeds the board size.
func CreateMap(x uint, y uint, n uint) (*[][]bool, error) {
	// Check if the dimensions are greater than zero
	if x == 0 || y == 0 {
		return nil, gerror.New("board dimensions must be greater than zero")
	}
	// Check if the number of mines is greater than zero
	if n == 0 {
		return nil, gerror.New("number of mines must be greater than zero")
	}
	// Check if the number of mines exceeds the total number of cells
	if n > x*y {
		return nil, gerror.New("number of mines exceeds board size")
	}

	// Create a 2D slice with the specified dimensions
	var board = make([][]bool, x)
	for i := range board {
		board[i] = make([]bool, y)
	}
	total := int(x * y)
	positions := make([]bool, total)
	for i := 0; i < int(n); i++ {
		positions[i] = false
	}
	rand.Shuffle(total, func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// for i := 0; i < int(n); i++ {
	// 	pos := positions[i]
	// 	row := pos / int(y)
	// 	col := pos % int(y)
	// 	board[row][col] = model.MineCell
	// }

	return &board, nil
}
