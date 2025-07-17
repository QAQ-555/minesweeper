package main

import (
	"minesweeper/app/ws"

	"github.com/gogf/gf/v2/frame/g"
)

func main() {
	s := g.Server()

	// Bind WebSocket handler to /ws endpoint
	s.BindHandler("/ws", ws.Handler)

	// Configure static file serving
	s.SetServerRoot("tools") // 启用静态文件服务，设置服务器根目录为 tools

	// Set server port
	s.SetPort(8002)
	// Start the server
	s.Run()
}
