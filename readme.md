# Minesweeper 服务端项目结构说明

## 目录结构

```
.
├── app/
│   ├── controller/      # 控制器层，处理路由和请求分发
│   ├── logic/           # 业务逻辑层，扫雷核心算法等
│   ├── model/           # 数据模型层，定义棋盘、玩家等结构体
│   ├── service/         # 服务层，组织业务流程
│   └── ws/              # WebSocket 相关处理
├── boot/                # 启动初始化（如路由注册、配置加载）
├── config/              # 配置文件目录
├── public/              # 静态资源目录（前端页面、图片等）
├── main.go              # 程序入口
└── go.mod               # Go 依赖管理文件
```

## 各目录说明

- **app/controller/**  
  控制器层，负责接收 HTTP/WebSocket 请求，参数校验，调用 service 层处理业务。

- **app/logic/**  
  业务逻辑层，封装扫雷核心算法、棋盘生成、判定等纯逻辑代码。

- **app/model/**  
  数据模型层，定义棋盘、房间、玩家等结构体，便于数据管理和扩展。

- **app/service/**  
  服务层，组织业务流程，调用 logic 和 model，处理事务。

- **app/ws/**  
  WebSocket 相关处理，管理连接、消息分发等。

- **boot/**  
  启动初始化代码，如路由注册、全局中间件、配置加载等。

- **config/**  
  配置文件目录，存放如 `config.toml` 等配置文件。

- **public/**  
  静态资源目录，存放前端页面、图片等静态文件。

- **main.go**  
  程序入口，负责启动服务。
