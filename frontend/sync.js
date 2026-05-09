// 1. Cấu hình cơ bản
// Giả lập ID người dùng và truyện. Thực tế sẽ lấy từ LocalStorage và URL.
const USER_ID = 1;
const MANGA_ID = 101;
let currentChapter = 1;

// Các phần tử UI
const chapterDisplays = [document.getElementById('currentChapterDisplay'), document.getElementById('chapterNumberText')];
const terminal = document.getElementById('terminalContent');
const connStatus = document.getElementById('connStatus');

// 2. Hàm in Log xuống Terminal
function logTerminal(direction, payloadStr) {
    const time = new Date().toLocaleTimeString();
    const logEntry = document.createElement('div');

    let dirColor = direction === 'SENT' ? 'text-blue-400' : 'text-purple-400';
    let icon = direction === 'SENT' ? '►' : '◄';

    // Highlight ký tự \n (bản chất của TCP Delimiter)
    const formattedPayload = payloadStr.replace('\n', '<span class="text-red-500">\\n</span>');

    logEntry.innerHTML = `
        <span class="text-slate-500">[${time}]</span> 
        <span class="${dirColor} font-bold">${icon} ${direction}</span> 
        <span class="text-yellow-300 ml-2">${formattedPayload}</span>
    `;
    terminal.appendChild(logEntry);
    terminal.parentElement.scrollTop = terminal.parentElement.scrollHeight;
}

// 3. Khởi tạo cầu nối TCP qua WebSocket
const ws = new WebSocket('ws://localhost:8080/api/ws-tcp-bridge');

ws.onopen = () => {
    connStatus.innerText = "Đã kết nối TCP Sync Server (:9090)";
    connStatus.classList.replace('text-slate-500', 'text-green-500');

    // Ngay khi kết nối, gửi gói tin AUTH để "xưng danh" với Server TCP
    const authPayload = JSON.stringify({
        type: "AUTH",
        user_id: USER_ID
    }) + "\n"; // QUAN TRỌNG: Bắt buộc có \n ở cuối cho giao thức TCP

    ws.send(authPayload);
    logTerminal('SENT', authPayload);
};

// 4. Lắng nghe gói tin Broadcast từ Server
ws.onmessage = (event) => {
    const rawData = event.data;
    logTerminal('RECV', rawData);

    try {
        // Dữ liệu nhận được từ TCP vẫn còn dấu \n, cần parse cẩn thận
        const payload = JSON.parse(rawData.trim());

        if (payload.type === "UPDATE_PROGRESS" && payload.manga_id === MANGA_ID) {
            currentChapter = payload.chapter;
            updateUI();
        }
    } catch (e) {
        console.error("Lỗi parse gói tin TCP:", e);
    }
};

ws.onclose = () => {
    connStatus.innerText = "Mất kết nối TCP";
    connStatus.classList.replace('text-green-500', 'text-red-500');
};

// 5. Xử lý nút bấm Lật chương
function sendSyncUpdate() {
    const payload = JSON.stringify({
        type: "UPDATE_PROGRESS",
        user_id: USER_ID,
        manga_id: MANGA_ID,
        chapter: currentChapter
    }) + "\n";

    ws.send(payload);
    logTerminal('SENT', payload);
    updateUI();
}

document.getElementById('btnPrev').addEventListener('click', () => {
    if (currentChapter > 1) {
        currentChapter--;
        sendSyncUpdate();
    }
});

document.getElementById('btnNext').addEventListener('click', () => {
    currentChapter++;
    sendSyncUpdate();
});

function updateUI() {
    chapterDisplays.forEach(el => el.innerText = currentChapter);
}