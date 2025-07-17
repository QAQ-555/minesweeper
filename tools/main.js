let ws;
let xSize, ySize;

// 用于缓存多个msgid的数据
const messageBuffers = new Map();

document.addEventListener("contextmenu", e => e.preventDefault());

document.getElementById('start').onclick = () => {
  xSize = parseInt(document.getElementById('x').value);
  ySize = parseInt(document.getElementById('y').value);
  const num = parseInt(document.getElementById('num').value);

  ws = new WebSocket("ws://localhost:8001/ws");

  ws.onopen = () => {
    ws.send(JSON.stringify({
      type: 0,
      id: "sss",
      payload: { x: xSize, y: ySize, num: num }
    }));

    document.getElementById('config').style.display = 'none';
    document.getElementById('game').style.display = 'block';

    createBoardWithCoords(xSize, ySize);
  };

  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if ((msg.type === 16 || msg.type === 17) && msg.payload && msg.payload.msgid !== undefined) {
        handlePartialUpdate(msg.payload);
    }  else {
      console.log("server:", e.data);
    }
  };
};

function createBoardWithCoords(x, y) {
  const board = document.getElementById('board');
  board.style.gridTemplateColumns = `30px repeat(${x}, 30px)`;
  board.innerHTML = '';

  const corner = document.createElement('div');
  corner.className = 'coord';
  board.appendChild(corner);

  for (let col = 0; col < x; col++) {
    const coordX = document.createElement('div');
    coordX.className = 'coord';
    coordX.textContent = col;
    board.appendChild(coordX);
  }

  for (let row = 0; row < y; row++) {
    const coordY = document.createElement('div');
    coordY.className = 'coord';
    coordY.textContent = row;
    board.appendChild(coordY);

    for (let col = 0; col < x; col++) {
      const cell = document.createElement('div');
      cell.className = 'cell';
      cell.dataset.x = col;
      cell.dataset.y = row;
      cell.dataset.flag = "false";
      cell.dataset.revealed = "false";

      cell.onmousedown = (e) => {
        e.preventDefault();
        const isLeftClick = (e.button === 0);
        const isRightClick = (e.button === 2);

        if (isRightClick) {
          if (cell.dataset.revealed === "true") return;

          if (cell.dataset.flag === "false") {
            cell.dataset.flag = "true";
            cell.textContent = "🚩";
            cell.style.backgroundColor = "lightyellow";
          } else {
            cell.dataset.flag = "false";
            cell.textContent = "";
            cell.style.backgroundColor = "lightgray";
          }
        } else if (isLeftClick) {
          if (cell.dataset.flag === "true") {
            cell.dataset.flag = "false";
            cell.textContent = "";
            cell.style.backgroundColor = "lightgray";
          } else {
            if (ws && ws.readyState === WebSocket.OPEN) {
              ws.send(JSON.stringify({
                type: 15,
                id: "sss",
                payload: {
                  x: col,
                  y: row,
                  click: true
                }
              }));
            }
          }
        }
      };

      board.appendChild(cell);
    }
  }
}

function handlePartialUpdate(payload) {
  const { msgid, x, y, num, end } = payload;

  if (!messageBuffers.has(msgid)) {
    messageBuffers.set(msgid, []);
  }
  const buffer = messageBuffers.get(msgid);
  buffer.push({ x, y, num });

  if (end) {
    renderCells(buffer);
    messageBuffers.delete(msgid);
  }
}

function renderCells(cells) {
  for (const cell of cells) {
    const selector = `.cell[data-x="${cell.x}"][data-y="${cell.y}"]`;
    const elem = document.querySelector(selector);
    if (!elem) continue;

    elem.dataset.flag = "false";
    elem.dataset.revealed = "true";

    if (cell.num === 0) {
      elem.textContent = '';
      elem.style.backgroundColor = '#eee';
    } else if (cell.num === 9) {
      elem.textContent = '🚩';
      elem.style.backgroundColor = 'lightyellow';
      elem.dataset.flag = "true";
    } else if (cell.num === 10) {
      elem.textContent = '💣';
      elem.style.backgroundColor = 'red';
    } else {
      elem.textContent = cell.num;
      elem.style.backgroundColor = '#ddd';
    }
  }
}
