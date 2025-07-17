let ws;
let xSize, ySize, mineCount;
let timer;
let startTime;
let gameOver = false;
// 用于存储未处理完的消息链
const messageBuffers = new Map();

document.addEventListener('contextmenu', (e) => e.preventDefault());

document.getElementById('start').addEventListener('click', () => {
    xSize = parseInt(document.getElementById('x').value);
    ySize = parseInt(document.getElementById('y').value);
    mineCount = parseInt(document.getElementById('num').value);

    document.getElementById('mine-count').textContent = `剩余地雷: ${mineCount}`;

    ws = new WebSocket('ws://localhost:8001/ws');

    ws.onopen = () => {
        startTime = Date.now();
        timer = setInterval(updateTimer, 1000);
        ws.send(JSON.stringify({
            type: 0,
            id: 'sss',
            payload: { x: xSize, y: ySize, num: mineCount }
        }));

        document.getElementById('config').style.display = 'none';
        document.getElementById('game').style.display = 'block';

        createBoard();
    };
    // 保留这一个 ws.onmessage 定义
    ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);
        console.log('接收到消息:', msg); // 打印接收到的消息
        switch (msg.type) {
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
});

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
    const unrevealedCells = document.querySelectorAll('.cell:not([data-revealed="true"]):not([data-flagged="true"])');
    if (unrevealedCells.length === 0) {
        gameOver = true;
        clearInterval(timer);
        document.getElementById('game-status').textContent = '恭喜你，游戏胜利！';
    }
}

function updateTimer() {
    const elapsed = Math.floor((Date.now() - startTime) / 1000);
    document.getElementById('timer').textContent = `用时: ${elapsed} 秒`;
}
