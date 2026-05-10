(function() {
    // =================================================================
    // sync.js - MangaHub TCP Sync Module (Scoped Version)
    // =================================================================

    console.log("[Sync] Script loading...");

    // 1. Lấy Token và USER_ID từ JWT
    const _localToken = localStorage.getItem('token');
    let _userId = 0;
    let _username = 'User';

    if (_localToken) {
        try {
            const base64Url = _localToken.split('.')[1];
            const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
            const jsonPayload = decodeURIComponent(
                window.atob(base64).split('').map(c =>
                    '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
                ).join('')
            );
            const parsed = JSON.parse(jsonPayload);
            _userId = parsed.id || 0;
            _username = parsed.username || 'User';
        } catch (e) {
            console.error("[Sync] Lỗi giải mã token:", e);
        }
    }

    // 2. Lấy MANGA_ID từ URL (số nguyên, phải > 0)
    function getQueryParam(name) {
        const results = new RegExp('[?&]' + name + '=([^&#]*)').exec(window.location.href);
        return results ? parseInt(results[1], 10) : 0;
    }

    const MANGA_ID = getQueryParam('id');
    let currentChapter = 1;

    console.log("[Sync] Khởi tạo: USER_ID=" + _userId + " | MANGA_ID=" + MANGA_ID + " | USER=" + _username);

    // Không kết nối nếu thiếu thông tin cần thiết
    if (_userId <= 0 || MANGA_ID <= 0) {
        console.error("[Sync] Thiếu USER_ID hoặc MANGA_ID, hủy kết nối TCP Sync.");
        const connStatus = document.getElementById('connStatus');
        if (connStatus) {
            connStatus.innerText = "Sync bị vô hiệu hóa (thiếu user/manga ID)";
            connStatus.className = "text-red-500";
        }
        return;
    }

    // 3. Phần tử UI
    const chapterDisplays = [
        document.getElementById('currentChapterDisplay'),
        document.getElementById('chapterNumberText')
    ];
    const terminal = document.getElementById('terminalContent');
    const connStatus = document.getElementById('connStatus');

    // 4. Hàm cập nhật UI chương
    function updateUI() {
        chapterDisplays.forEach(el => { if (el) el.innerText = currentChapter; });
    }

    // 4b. Hàm cập nhật Live Presence Badge
    function updatePresenceBadge(count) {
        const badge = document.getElementById('liveReadersBadge');
        const text  = document.getElementById('liveReadersText');
        if (!badge || !text) return;

        if (count > 1) {
            text.textContent = `👁 ${count} người đang đọc`;
            // Dùng flex để hiện badge (thay thế hidden)
            badge.classList.remove('hidden');
            badge.classList.add('flex');
        } else {
            badge.classList.add('hidden');
            badge.classList.remove('flex');
        }
    }

    // 5. Hàm in Log xuống Terminal
    function logTerminal(direction, payloadStr) {
        if (!terminal) return;
        const time = new Date().toLocaleTimeString();
        const logEntry = document.createElement('div');
        const dirColor = direction === 'SENT' ? 'text-blue-400' : 'text-purple-400';
        const icon = direction === 'SENT' ? '►' : '◄';
        const safePayload = String(payloadStr).replace(/\n/g, '<span class="text-red-500">\\n</span>');
        logEntry.innerHTML = `
            <span class="text-slate-500">[${time}]</span>
            <span class="${dirColor} font-bold">${icon} ${direction}</span>
            <span class="text-yellow-300 ml-2">${safePayload}</span>
        `;
        terminal.appendChild(logEntry);
        if (terminal.parentElement) {
            terminal.parentElement.scrollTop = terminal.parentElement.scrollHeight;
        }
    }

    // 6. Khởi tạo WebSocket → TCP Bridge
    function connectTCP() {
        const ws = new WebSocket('ws://localhost:8080/api/ws-tcp-bridge');

        ws.onopen = () => {
            console.log("[Sync] Kết nối thành công!");
            if (connStatus) {
                connStatus.innerText = `[User:${_username}] Đã kết nối TCP Sync Server (:9090)`;
                connStatus.className = "text-green-500";
            }

            // Gửi AUTH với đầy đủ user_id và manga_id
            const authPayload = JSON.stringify({
                type: "AUTH",
                user_id: _userId,
                manga_id: MANGA_ID
            }) + "\n";

            ws.send(authPayload);
            logTerminal('SENT', authPayload.trim());

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
            // Defensive: tách từng dòng JSON riêng (phòng trường hợp nhiều message ghép chung 1 frame)
            const lines = event.data.split('\n').map(l => l.trim()).filter(l => l.length > 0);
            for (const rawData of lines) {
                logTerminal('RECV', rawData);
                try {
                    const payload = JSON.parse(rawData);
                    handlePayload(payload);
                } catch (e) {
                    console.error("[Sync] Parse error:", e, "| Raw:", rawData);
                }
            }
        };

        function handlePayload(payload) {
            if (payload.type === "UPDATE_PROGRESS" && payload.manga_id === MANGA_ID) {
                // Nhận cập nhật chương từ thiết bị khác của cùng user
                currentChapter = payload.chapter;
                updateUI();
                logTerminal('SYS', `[Sync] Thiết bị khác đang đọc Chương ${currentChapter}`);

            } else if (payload.type === "SYNC_RESUME" && payload.manga_id === MANGA_ID) {
                // Server tìm thấy tiến độ cũ → hỏi user có muốn tiếp tục không
                const resumeModal = document.getElementById('resumeModal');
                const resumeChapterNum = document.getElementById('resumeChapterNum');
                if (resumeModal && resumeChapterNum) {
                    resumeChapterNum.innerText = `Chương ${payload.chapter}`;
                    resumeModal.classList.remove('hidden');

                    document.getElementById('btnAcceptResume').onclick = () => {
                        currentChapter = payload.chapter;
                        updateUI();
                        resumeModal.classList.add('hidden');
                        logTerminal('SYS', `[Resume] Đã nhảy đến Chương ${currentChapter}`);
                    };
                    document.getElementById('btnIgnoreResume').onclick = () => {
                        resumeModal.classList.add('hidden');
                        logTerminal('SYS', `[Resume] Bỏ qua tiến độ cũ`);
                    };
                }

            } else if (payload.type === "PRESENCE_UPDATE" && payload.manga_id === MANGA_ID) {
                // Cập nhật badge Live Presence
                updatePresenceBadge(payload.count);
                logTerminal('SYS', `[Presence] ${payload.count} người đang đọc manga này`);
            }
        }

        ws.onclose = () => {
            if (connStatus) {
                connStatus.innerText = "Mất kết nối TCP";
                connStatus.className = "text-red-500";
            }
            // Tự kết nối lại sau 3 giây
            setTimeout(connectTCP, 3000);
        };

        ws.onerror = (err) => {
            console.error("[Sync] WebSocket error:", err);
        };
    }

    function sendUpdate(ws) {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            console.warn("[Sync] WebSocket chưa sẵn sàng để gửi.");
            return;
        }
        const payload = JSON.stringify({
            type: "UPDATE_PROGRESS",
            user_id: _userId,
            manga_id: MANGA_ID,
            chapter: currentChapter
        }) + "\n";
        ws.send(payload);
        logTerminal('SENT', payload.trim());
        updateUI();
    }

    connectTCP();
})();