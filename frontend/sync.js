(function() {
    // =================================================================
    // sync.js - MangaHub TCP Sync Module (Scoped Version)
    // =================================================================

    console.log("[Sync] Script loading...");

    // 1. Lấy Token và USER_ID (Dùng biến cục bộ trong hàm để tránh trùng tên)
    const _localToken = localStorage.getItem('token');
    let _userId = 1;

    if (_localToken) {
        try {
            const base64Url = _localToken.split('.')[1];
            const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
            const jsonPayload = decodeURIComponent(
                window.atob(base64).split('').map(c =>
                    '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
                ).join('')
            );
            _userId = JSON.parse(jsonPayload).id || 1;
        } catch (e) {
            console.error("[Sync] Lỗi giải mã token:", e);
        }
    }

    // 2. Lấy MANGA_ID từ URL
    function getQueryParam(name) {
        const results = new RegExp('[\?&]' + name + '=([^&#]*)').exec(window.location.href);
        return results ? parseInt(results[1]) : 0;
    }

    const MANGA_ID = getQueryParam('id');
    let currentChapter = 1;

    console.log("[Sync] Khởi tạo: USER_ID=" + _userId + " | MANGA_ID=" + MANGA_ID);

    // 3. Phần tử UI
    const chapterDisplays = [
        document.getElementById('currentChapterDisplay'),
        document.getElementById('chapterNumberText')
    ];
    const terminal = document.getElementById('terminalContent');
    const connStatus = document.getElementById('connStatus');

    // 4. Hàm cập nhật UI
    function updateUI() {
        chapterDisplays.forEach(el => { if (el) el.innerText = currentChapter; });
    }

    // 5. Hàm in Log xuống Terminal
    function logTerminal(direction, payloadStr) {
        if (!terminal) return;
        const time = new Date().toLocaleTimeString();
        const logEntry = document.createElement('div');
        const dirColor = direction === 'SENT' ? 'text-blue-400' : 'text-purple-400';
        const icon = direction === 'SENT' ? '►' : '◄';
        const formattedPayload = String(payloadStr).replace('\n', '<span class="text-red-500">\\n</span>');
        logEntry.innerHTML = `
            <span class="text-slate-500">[${time}]</span>
            <span class="${dirColor} font-bold">${icon} ${direction}</span>
            <span class="text-yellow-300 ml-2">${formattedPayload}</span>
        `;
        terminal.appendChild(logEntry);
        if (terminal.parentElement) {
            terminal.parentElement.scrollTop = terminal.parentElement.scrollHeight;
        }
    }

    // 6. Khởi tạo WebSocket
    function connectTCP() {
        const ws = new WebSocket('ws://localhost:8080/api/ws-tcp-bridge');

        ws.onopen = () => {
            console.log("[Sync] Kết nối thành công!");
            if (connStatus) {
                connStatus.innerText = "Đã kết nối TCP Sync Server (:9090)";
                connStatus.className = "text-green-500";
            }

            const authPayload = JSON.stringify({
                type: "AUTH",
                user_id: _userId,
                manga_id: MANGA_ID
            }) + "\n";

            ws.send(authPayload);
            logTerminal('SENT', authPayload);

            // Gắn sự kiện nút bấm
            const btnNext = document.getElementById('btnNext');
            const btnPrev = document.getElementById('btnPrev');

            if (btnNext) {
                btnNext.onclick = () => {
                    currentChapter++;
                    sendUpdate(ws);
                };
            }
            if (btnPrev) {
                btnPrev.onclick = () => {
                    if (currentChapter > 1) {
                        currentChapter--;
                        sendUpdate(ws);
                    }
                };
            }
        };

        ws.onmessage = (event) => {
            const rawData = event.data;
            logTerminal('RECV', rawData);
            try {
                const payload = JSON.parse(rawData.trim());
                if (payload.type === "UPDATE_PROGRESS" && payload.manga_id === MANGA_ID) {
                    currentChapter = payload.chapter;
                    updateUI();
                } else if (payload.type === "SYNC_RESUME" && payload.manga_id === MANGA_ID) {
                    const resumeModal = document.getElementById('resumeModal');
                    const resumeChapterNum = document.getElementById('resumeChapterNum');
                    if (resumeModal && resumeChapterNum) {
                        resumeChapterNum.innerText = `Chương ${payload.chapter}`;
                        resumeModal.classList.remove('hidden');
                        document.getElementById('btnAcceptResume').onclick = () => {
                            currentChapter = payload.chapter;
                            updateUI();
                            resumeModal.classList.add('hidden');
                        };
                        document.getElementById('btnIgnoreResume').onclick = () => {
                            resumeModal.classList.add('hidden');
                        };
                    }
                }
            } catch (e) { console.error("[Sync] Parse error:", e); }
        };

        ws.onclose = () => {
            if (connStatus) {
                connStatus.innerText = "Mất kết nối TCP";
                connStatus.className = "text-red-500";
            }
        };
    }

    function sendUpdate(ws) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        const payload = JSON.stringify({
            type: "UPDATE_PROGRESS",
            user_id: _userId,
            manga_id: MANGA_ID,
            chapter: currentChapter
        }) + "\n";
        ws.send(payload);
        logTerminal('SENT', payload);
        updateUI();
    }

    connectTCP();
})();