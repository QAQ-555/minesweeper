package model

const (
	MineCell  = true
	EmptyCell = false

	blank byte = 0
	one   byte = 1
	two   byte = 2
	three byte = 3
	four  byte = 4
	five  byte = 5
	six   byte = 6
	seven byte = 7
	eight byte = 8
	// flag is used to mark a cell as flagged
	flag byte = 9
) //8向方位代码

type GameOptionPayload struct {
	X       uint `json:"x"`
	Y       uint `json:"y"`
	MineNUM uint `json:num`
}
