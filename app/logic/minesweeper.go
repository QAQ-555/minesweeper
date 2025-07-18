package logic

import (
	"context"
	"fmt"
	"math/rand/v2"
	"minesweeper/app/model"
	"minesweeper/app/service"
	"os"
	"strconv"

	"github.com/gogf/gf/errors/gerror"
	"github.com/google/uuid"
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
func CreateRealMap(x, y, n, safeX, safeY uint) ([][]bool, error) {
	if x == 0 || y == 0 {
		return nil, gerror.New("board dimensions must be greater than zero")
	}
	if n == 0 {
		return nil, gerror.New("number of mines must be greater than zero")
	}
	if n >= x*y {
		return nil, gerror.New("number of mines exceeds or equals board size")
	}
	if safeX >= x || safeY >= y {
		return nil, gerror.New("safe coordinate out of bounds")
	}

	board := make([][]bool, y)
	for i := range board {
		board[i] = make([]bool, x)
	}

	// 定义首次点击位置的 8 格范围
	var safeArea []int
	directions := [8]struct{ X, Y int }{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}
	for _, dir := range directions {
		newX := int(safeX) + dir.X
		newY := int(safeY) + dir.Y
		if newX >= 0 && newY >= 0 && newX < int(x) && newY < int(y) {
			safeArea = append(safeArea, newY*int(x)+newX)
		}
	}

	// 生成非 8 格范围的下标
	nonSafeIndices := make([]int, 0, int(x*y))
	for i := 0; i < int(x*y); i++ {
		// 排除安全位置和 8 格范围
		if i == int(safeY*x+safeX) {
			continue
		}
		isSafeArea := false
		for _, safeIdx := range safeArea {
			if i == safeIdx {
				isSafeArea = true
				break
			}
		}
		if !isSafeArea {
			nonSafeIndices = append(nonSafeIndices, i)
		}
	}

	// 打乱非 8 格范围的下标
	rand.Shuffle(len(nonSafeIndices), func(i, j int) {
		nonSafeIndices[i], nonSafeIndices[j] = nonSafeIndices[j], nonSafeIndices[i]
	})

	minesPlaced := 0
	// 先在非 8 格范围布雷
	for i := 0; i < len(nonSafeIndices) && minesPlaced < int(n); i++ {
		idx := nonSafeIndices[i]
		row := idx / int(x)
		col := idx % int(x)
		board[row][col] = true
		minesPlaced++
	}

	// 若非 8 格范围格子不足，在 8 格范围随机布雷
	if minesPlaced < int(n) {
		// 打乱 8 格范围下标
		rand.Shuffle(len(safeArea), func(i, j int) {
			safeArea[i], safeArea[j] = safeArea[j], safeArea[i]
		})
		for i := 0; i < len(safeArea) && minesPlaced < int(n); i++ {
			idx := safeArea[i]
			row := idx / int(x)
			col := idx % int(x)
			board[row][col] = true
			minesPlaced++
		}
	}

	// 确保 safe cell 没有雷
	board[safeY][safeX] = false

	return board, nil
}

// GetUserMap is a placeholder function that should return a client view of the map.
// Parameters:
//
// Realmap - a pointer to the real map generated by CreateRealMap
//
// Returns:
//
// A pointer to a 2D byte array representing the client view of the map.
// Currently returns nil as a placeholder.

func GetUserMap(Realmap *[][]bool) *[][]byte {
	return nil
}

func HandleLeftClick(ctx context.Context, x, y uint, c *model.Client) {
	model.Logger.Infof(ctx, "Left click received at position (%d, %d) for client %s", x, y, c.ID)

	switch {
	case c.MapClient[y][x] == model.Flag:
		return

	case c.MapClient[y][x] != model.Unknown:
		handleKnownCell(ctx, x, y, c)
		return

	case c.MapServer[y][x] == model.MineCell:
		handleMineClick(ctx, x, y, c)
		return

	default:
		handleBlankClick(ctx, x, y, c)
	}
}

func handleKnownCell(ctx context.Context, x, y uint, c *model.Client) {
	if getAroundFlagNum(x, y, c) == getAroundMineNum(x, y, c) {
		model.Logger.Infof(ctx, "No mine around position (%d, %d) for client %s", x, y, c.ID)
		processAroundCells(ctx, x, y, c)
	} else {
		model.Logger.Infof(ctx, "Have mine around position (%d, %d) for client %s", x, y, c.ID)
		revealUnknownAround(ctx, x, y, c)
	}
}

func processAroundCells(ctx context.Context, x, y uint, c *model.Client) {
	for _, dir := range getDirections() {
		newX, newY := int(x)+dir.X, int(y)+dir.Y
		if !inBounds(newX, newY, c) {
			continue
		}

		if c.MapClient[newY][newX] != model.Flag && c.MapServer[newY][newX] == model.MineCell {
			handleMineClick(ctx, uint(newX), uint(newY), c)
			return
		}

		var result []model.ClickResultpayload
		blankCellRecursive(ctx, uint(newX), uint(newY), c, &result)
		sendRevealMessages(ctx, result, c, model.TypeResult)
	}
}

func revealUnknownAround(ctx context.Context, x, y uint, c *model.Client) {
	var result []model.ClickResultpayload
	msgChainId := uuid.NewString()

	for _, dir := range getDirections() {
		newX, newY := int(x)+dir.X, int(y)+dir.Y
		if !inBounds(newX, newY, c) {
			continue
		}
		if c.MapClient[newY][newX] == model.Unknown {
			result = append(result, model.ClickResultpayload{
				X:       uint(newX),
				Y:       uint(newY),
				MineNum: 11,
				MsgId:   msgChainId,
				IsEnd:   false,
			})
		}
	}

	sendRevealMessages(ctx, result, c, model.TypeResult)
}

func handleMineClick(ctx context.Context, x, y uint, c *model.Client) {
	model.Logger.Infof(ctx, "Mine clicked at position (%d, %d). Sending boom messages.", x, y)
	msgChainId := uuid.NewString()
	count := 0

	for yIdx, row := range c.MapServer {
		for xIdx, val := range row {
			if val == model.MineCell {
				count++
				isEnd := count == int(c.Map_mine_num)
				payload := model.ClickResultpayload{
					X:       uint(xIdx),
					Y:       uint(yIdx),
					MsgId:   msgChainId,
					IsEnd:   isEnd,
					MineNum: 10,
				}
				sendMessage(ctx, c, payload, model.TypeBoom)
			}
		}
	}

	model.Logger.Infof(ctx, "All boom messages sent.")
}

func handleBlankClick(ctx context.Context, x, y uint, c *model.Client) {
	model.Logger.Infof(ctx, "No mine at position (%d, %d). Starting recursive reveal.", x, y)
	var result []model.ClickResultpayload
	blankCellRecursive(ctx, x, y, c, &result)
	sendRevealMessages(ctx, result, c, model.TypeResult)
}

func sendRevealMessages(ctx context.Context, result []model.ClickResultpayload, c *model.Client, msgType byte) {
	if len(result) == 0 {
		model.Logger.Infof(ctx, "Reveal completed. No cells to reveal.")
		return
	}

	model.Logger.Infof(ctx, "Reveal completed. Found %d cells to reveal.", len(result))
	msgChainId := uuid.NewString()

	for i := range result {
		result[i].IsEnd = (i == len(result)-1)
		result[i].MsgId = msgChainId
		sendMessage(ctx, c, result[i], msgType)
	}

	model.Logger.Infof(ctx, "All reveal messages sent.")
}

func sendMessage(ctx context.Context, c *model.Client, payload model.ClickResultpayload, msgType byte) {
	data, err := service.PackWebMessageJson(ctx, msgType, payload, payload.MsgId)
	if err != nil {
		model.Logger.Panicf(ctx, "Failed to pack web message for position (%d, %d): %v", payload.X, payload.Y, err)
	}
	err = c.Conn.WriteMessage(1, data)
	if err != nil {
		model.Logger.Errorf(ctx, "Failed to send message for position (%d, %d): %v", payload.X, payload.Y, err)
	} else {
		model.Logger.Infof(ctx, "Message sent for position (%d, %d), type: %d", payload.X, payload.Y, msgType)
	}
}

func getDirections() [8]struct{ X, Y int } {
	return [8]struct{ X, Y int }{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}
}

func inBounds(x, y int, c *model.Client) bool {
	return x >= 0 && y >= 0 && x < int(c.Map_size_x) && y < int(c.Map_size_y)
}

func HandleRightClick(ctx context.Context, x, y uint, c *model.Client) {
	if c.MapClient[y][x] != model.Flag && c.MapClient[y][x] != model.Unknown {
		handleKnownCell(ctx, x, y, c)
		return
	}
	oldValue := c.MapClient[y][x]
	var chose byte = oldValue
	switch oldValue {
	case model.Unknown:
		c.MapClient[y][x] = model.Flag
		model.Logger.Infof(ctx, "MapClient[%d][%d] changed from %d to %d", y, x, oldValue, model.Flag)
		chose = model.Flag
	case model.Flag:
		c.MapClient[y][x] = model.Unknown
		model.Logger.Infof(ctx, "MapClient[%d][%d] changed from %d to %d", y, x, oldValue, model.Unknown)
		chose = model.Unknown
	}
	result := model.ClickResultpayload{
		X:       x,
		Y:       y,
		MineNum: chose,
		IsEnd:   true,
	}
	data, err := service.PackWebMessageJson(ctx, model.TypeResult, result, "")
	if err != nil {
		model.Logger.Panicf(ctx, "unkonw fail!!")
	}
	err = c.Conn.WriteMessage(1, data)
	if err != nil {
		model.Logger.Errorf(ctx, "send massage fail")
	}
}

func blankCellRecursive(ctx context.Context, x, y uint, c *model.Client, result *[]model.ClickResultpayload) {
	directions := [8]struct{ X, Y int }{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}
	// 如果已经揭开过了，直接返回
	if c.MapClient[y][x] != model.Unknown {
		return
	}

	// 计算周围雷数
	count := getAroundMineNum(x, y, c)
	model.Logger.Infof(ctx, "around[%d,%d] mine num: %d", x, y, count)
	oldValue := c.MapClient[y][x]
	c.MapClient[y][x] = byte(count)
	model.Logger.Infof(ctx, "MapClient[%d][%d] changed from %d to %d", y, x, oldValue, byte(count))

	if count == 0 {
		// 如果周围没有雷，递归揭开周围
		for _, dir := range directions {
			newX := int(x) + dir.X
			newY := int(y) + dir.Y
			if newX < 0 || newY < 0 || newX >= int(c.Map_size_x) || newY >= int(c.Map_size_y) {
				continue
			}
			blankCellRecursive(ctx, uint(newX), uint(newY), c, result)
		}
	}

	oneResult := &model.ClickResultpayload{
		X:       x,
		Y:       y,
		MineNum: c.MapClient[y][x],
	}
	*result = append(*result, *oneResult)
}

func getAroundMineNum(x, y uint, c *model.Client) uint {
	directions := [8]struct{ X, Y int }{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}
	count := 0
	for _, dir := range directions {
		newX := int(x) + dir.X
		newY := int(y) + dir.Y
		if newX < 0 || newY < 0 || newX >= int(c.Map_size_x) || newY >= int(c.Map_size_y) {
			continue
		}
		if c.MapServer[newY][newX] == model.MineCell {
			count++
		}
	}
	return uint(count)
}

func getAroundFlagNum(x, y uint, c *model.Client) uint {
	directions := [8]struct{ X, Y int }{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}
	count := 0
	for _, dir := range directions {
		newX := int(x) + dir.X
		newY := int(y) + dir.Y
		if newX < 0 || newY < 0 || newX >= int(c.Map_size_x) || newY >= int(c.Map_size_y) {
			continue
		}
		if c.MapClient[newY][newX] == model.Flag {
			count++
		}
	}
	return uint(count)
}
func SaveBoardWithCoords(ctx context.Context, board [][]bool) error {
	name := ctx.Value("requestID")
	filename := "default.txt"
	if str, ok := name.(string); ok && str != "" {
		filename = fmt.Sprintf("%s.txt", str)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 横坐标头
	_, _ = file.WriteString("   ") // 缩进3空格
	for x := 0; x < len(board[0]); x++ {
		_, _ = file.WriteString(strconv.Itoa(x) + " ")
	}
	_, _ = file.WriteString("\n")

	// 每行写纵坐标和内容
	for y, row := range board {
		_, _ = file.WriteString(strconv.Itoa(y) + " ") // 写纵坐标
		if y < 10 {
			_, _ = file.WriteString(" ") // 保证对齐，两位数以上可调整
		}
		for _, cell := range row {
			if cell {
				_, _ = file.WriteString("■ ")
			} else {
				_, _ = file.WriteString("□ ")
			}
		}
		_, _ = file.WriteString("\n")
	}

	return nil
}

// 删除ctx中requestID命名的文件
func DeleteBoardFile(ctx context.Context) error {
	name := ctx.Value("requestID")
	filename := "default.txt"
	if str, ok := name.(string); ok && str != "" {
		filename = fmt.Sprintf("%s.txt", str)
	}

	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
