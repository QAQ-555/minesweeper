let ws;
let xSize, ySize, mineCount;
let timer;
let startTime;
let gameOver = false;
// 用于存储未处理完的消息链
const messageBuffers = new Map();

document.addEventListener('contextmenu', (e) => e.preventDefault());

document.getElementById('start').addEventListener('click', startGame);

// 游戏内重新开始按钮事件监听
document.getElementById('in-game-restart').addEventListener('click', restartGame);

function startGame() {
    xSize = parseInt(document.getElementById('x').value);
    ySize = parseInt(document.getElementById('y').value);
    mineCount = parseInt(document.getElementById('num').value);

    document.getElementById('mine-count').textContent = `剩余地雷: ${mineCount}`;

    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close();
    }
    ws = new WebSocket('ws://localhost:8002/ws');

    ws.onopen = () => {
        startTime = Date.now();
        timer = setInterval(updateTimer, 1000);
        sendRestartMessage();

        document.getElementById('config').style.display = 'none';
        document.getElementById('game').style.display = 'block';

        createBoard();
    };

    ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);
        console.log('接收到消息:', msg); // 打印接收到的消息
        switch (msg.type) {
            case 0: // 处理重新开始游戏的消息
                console.log('开始处理重新开始游戏的消息');
                const payload = msg.payload;
                // 更新游戏参数
                xSize = payload.Map_size_x;
                ySize = payload.Map_size_y;
                mineCount = payload.Map_mine_num;
                document.getElementById('mine-count').textContent = `剩余地雷: ${mineCount}`;
                // 重新创建游戏面板
                createBoard();
                // 重置游戏状态
                gameOver = false;
                document.getElementById('game-status').textContent = '游戏进行中';
                startTime = Date.now();
                clearInterval(timer);
                timer = setInterval(updateTimer, 1000);
                document.getElementById('timer').textContent = '用时: 0 秒';
                break;
            case 16: // TypeResult
                console.log('开始处理 TypeResult 消息'); // 打印开始处理消息的日志
                handleResult(msg.payload);
                break;
            case 17: // TypeBoom
                console.log('开始处理 TypeBoom 消息'); // 打印开始处理消息的日志
                handleBoom(msg.payload);
                break;
            default:
                // console.log('未知消息类型:', msg.type);
        }
    };

    ws.onclose = () => {
        clearInterval(timer);
        if (!gameOver) {
            document.getElementById('game-status').textContent = '连接已断开';
        }
    };
}

function restartGame() {
    xSize = parseInt(document.getElementById('in-game-x').value);
    ySize = parseInt(document.getElementById('in-game-y').value);
    mineCount = parseInt(document.getElementById('in-game-num').value);

    document.getElementById('mine-count').textContent = `剩余地雷: ${mineCount}`;
    gameOver = false;
    document.getElementById('game-status').textContent = '游戏进行中';
    document.getElementById('timer').textContent = '用时: 0 秒';

    if (ws && ws.readyState === WebSocket.OPEN) {
        sendRestartMessage();
    } else {
        // 如果连接已关闭，重新建立连接
        startGame();
    }
}

function sendRestartMessage() {
    const message = {
        type: 0,
        id: 'sss',
        payload: {
            x: xSize,
            y: ySize,
            num: mineCount
        }
    };
    ws.send(JSON.stringify(message));
}

// 更新计时器
function updateTimer() {
    const elapsedTime = Math.floor((Date.now() - startTime) / 1000);
    document.getElementById('timer').textContent = `用时: ${elapsedTime} 秒`;
}

function createBoard() {
    const board = document.getElementById('board');
    board.style.gridTemplateColumns = `repeat(${xSize}, 30px)`;
    board.innerHTML = '';
    const coordsElement = document.getElementById('coords');

    for (let y = 0; y < ySize; y++) {
        for (let x = 0; x < xSize; x++) {
            const cell = document.createElement('div');
            cell.className = 'cell';
            cell.dataset.x = x;
            cell.dataset.y = y;
            cell.dataset.flagged = 'false';
            cell.dataset.revealed = 'false';

            cell.addEventListener('mousedown', (e) => {
                if (gameOver) return;

                e.preventDefault();
                const isLeftClick = e.button === 0;
                sendClick(x, y, isLeftClick);
            });

            cell.addEventListener('mouseover', () => {
                coordsElement.textContent = `坐标: (${x}, ${y})`;
            });

            cell.addEventListener('mouseout', () => {
                coordsElement.textContent = '坐标: (0, 0)';
            });

            board.appendChild(cell);
        }
    }
}

function sendClick(x, y, isLeftClick) {
    const message = {
        type: 15,
        id: 'sss',
        payload: {
            x: x,
            y: y,
            click: isLeftClick
        }
    };
    console.log('发送消息:', message); // 打印发送的消息
    ws.send(JSON.stringify(message));
}


function handleResult(payload) {
    console.log('处理 TypeResult 消息的 payload:', payload); // 打印处理的 payload
    const { msgid, x, y, num, end } = payload;
    // 如果消息链不存在，创建一个新的数组来存储消息
    if (!messageBuffers.has(msgid)) {
        messageBuffers.set(msgid, []);
    }
    const buffer = messageBuffers.get(msgid);
    buffer.push({ x, y, num });

    // 当 isEnd 为 true 时，说明消息链完结，渲染所有存储的单元格
    if (end) {
        console.log('TypeResult 消息链完结，开始渲染单元格'); // 打印消息链完结的日志
        renderCells(buffer);
        messageBuffers.delete(msgid);
        checkWin();
    }
}

function handleBoom(payload) {
    console.log('处理 TypeBoom 消息的 payload:', payload); // 打印处理的 payload
    gameOver = true;
    clearInterval(timer);
    document.getElementById('game-status').textContent = '游戏结束，你踩到地雷了！';
    renderCells([payload]);
}

// 移除以下动态添加 CSS 样式的代码
// const style = document.createElement('style');
// style.textContent = `
// @keyframes blink {
//     50% { opacity: 0.5; }
// }
// 
// .blinking {
//     animation: blink 1s infinite;
// }
// `;
// document.head.appendChild(style);

function renderCells(cells) {
    cells.forEach((cellData) => {
        const cell = document.querySelector(`.cell[data-x="${cellData.x}"][data-y="${cellData.y}"]`);
        if (!cell) return;

        if (cellData.num === 15) { // model.Flag
            cell.dataset.flagged = 'true';
            cell.textContent = '🚩';
            cell.classList.add('flagged');
        } else if (cellData.num === 16) { // model.Unknown
            cell.dataset.flagged = 'false';
            cell.dataset.revealed = 'false';
            cell.textContent = '';
            cell.classList.remove('flagged');
            cell.classList.remove('revealed');
        } else if (cellData.num === 11) {
            // 立即移除闪烁类
            cell.classList.remove('blinking-once');
            // 强制浏览器重绘，确保类已被移除
            void cell.offsetWidth; 
            // 使用 setTimeout 在下一个事件循环添加闪烁类
            setTimeout(() => {
                cell.classList.add('blinking-once');
            }, 0);
        } else {
            cell.dataset.flagged = 'false';
            cell.dataset.revealed = 'true';
            cell.classList.remove('flagged');
            cell.classList.add('revealed');

            switch (cellData.num) {
                case 0:
                    cell.textContent = '';
                    break;
                case 10:
                    cell.textContent = '💣';
                    cell.classList.add('mine');
                    break;
                default:
                    cell.textContent = cellData.num;
            }
        }
    });
}

function checkWin() {
    // 计算总格子数量
    const totalCells = xSize * ySize;
    // 获取所有已揭开的单元格
    const revealedCells = document.querySelectorAll('.cell[data-revealed="true"]');
    // 计算未揭开的格子数量
    const unrevealedCellsCount = totalCells - revealedCells.length;

    // 检查未揭开的格子数量是否等于地雷数量
    if (unrevealedCellsCount === mineCount) {
        gameOver = true;
        clearInterval(timer);
        document.getElementById('game-status').textContent = '恭喜你，游戏胜利！';
        
        // 自动标记剩余未揭开的格子为地雷
        const unrevealedCells = document.querySelectorAll('.cell:not([data-revealed="true"])');
        unrevealedCells.forEach(cell => {
            cell.dataset.flagged = 'true';
            cell.textContent = '🚩';
            cell.classList.add('flagged');
        });
    }
}

function updateTimer() {
    const elapsed = Math.floor((Date.now() - startTime) / 1000);
    document.getElementById('timer').textContent = `用时: ${elapsed} 秒`;
}

function initBoard(rows, cols) {
    const board = document.getElementById('board');
    // 设置 CSS 自定义属性
    board.style.setProperty('--rows', rows);
    board.style.setProperty('--cols', cols);

    board.style.gridTemplateColumns = `repeat(${xSize}, 30px)`;
    board.innerHTML = '';
    const coordsElement = document.getElementById('coords');

    for (let y = 0; y < ySize; y++) {
        for (let x = 0; x < xSize; x++) {
            const cell = document.createElement('div');
            cell.className = 'cell';
            cell.dataset.x = x;
            cell.dataset.y = y;
            cell.dataset.flagged = 'false';
            cell.dataset.revealed = 'false';

            cell.addEventListener('mousedown', (e) => {
                if (gameOver) return;

                e.preventDefault();
                const isLeftClick = e.button === 0;
                sendClick(x, y, isLeftClick);
            });

            cell.addEventListener('mouseover', () => {
                coordsElement.textContent = `坐标: (${x}, ${y})`;
            });

            cell.addEventListener('mouseout', () => {
                coordsElement.textContent = '坐标: (0, 0)';
            });

            board.appendChild(cell);
        }
    }
}
