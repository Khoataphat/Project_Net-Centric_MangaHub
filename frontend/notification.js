/**
 * Notification Service for MangaHub
 * Handles real-time notifications via WebSocket (UDP -> Bridge -> WS)
 */

const NOTIFICATION_CONTAINER_ID = 'notification-container';
const AUTO_HIDE_MS = 5000;
const WS_URL = 'ws://localhost:8080/api/ws/chat';

/**
 * [AC1] Khởi tạo kết nối WebSocket khi tải trang Dashboard
 */
function setupWebSocket() {
    console.log('[Notification] Connecting to WebSocket...');
    const ws = new WebSocket(WS_URL);

    ws.onopen = () => {
        console.log('[Notification] Connected to WebSocket Bridge');
    };

    ws.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            
            // [AC2] Kiểm tra event loại NEW_CHAPTER
            if (data.type === 'NEW_CHAPTER') {
                showNotification(data);
            }
        } catch (err) {
            console.error('[Notification] Error parsing WS message:', err);
        }
    };

    ws.onclose = () => {
        console.warn('[Notification] WS connection closed. Reconnecting in 5s...');
        setTimeout(setupWebSocket, 5000);
    };

    ws.onerror = (err) => {
        console.error('[Notification] WS error:', err);
    };
}

/**
 * [AC2, AC3, AC4, AC5] Hiển thị Toast UI
 * @param {Object} data - Dữ liệu từ Bridge (title, chapter)
 */
function showNotification(data) {
    const container = document.getElementById(NOTIFICATION_CONTAINER_ID);
    if (!container) return;

    // [AC3] Tạo element Toast với Tailwind CSS
    const toast = document.createElement('div');
    toast.className = `
        pointer-events-auto
        flex items-center w-full max-w-xs p-4 
        text-gray-900 bg-white rounded-lg shadow-xl border-l-4 border-blue-600
        transition-all duration-500 ease-out transform translate-x-full opacity-0
    `;

    // Template nội dung Toast [AC3]
    toast.innerHTML = `
        <div class="inline-flex items-center justify-center flex-shrink-0 w-8 h-8 text-blue-500 bg-blue-100 rounded-lg">
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M10 2a6 6 0 00-6 6v3.586l-.707.707A1 1 0 004 14h12a1 1 0 00.707-1.707L16 11.586V8a6 6 0 00-6-6zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z"></path></svg>
        </div>
        <div class="ml-3 text-sm font-normal">
            <span class="mb-1 text-sm font-semibold text-gray-900">Chương mới!</span>
            <div class="text-sm font-normal">
                <span class="font-bold text-blue-600">${data.title}</span> 
                vừa có chương <span class="font-bold">${data.chapter}</span>
            </div>
        </div>
        <button type="button" class="ml-auto -mx-1.5 -my-1.5 bg-white text-gray-400 hover:text-gray-900 rounded-lg focus:ring-2 focus:ring-gray-300 p-1.5 hover:bg-gray-100 inline-flex items-center justify-center h-8 w-8" aria-label="Close">
            <span class="sr-only">Close</span>
            <svg class="w-3 h-3" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 14">
                <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"/>
            </svg>
        </button>
    `;

    container.appendChild(toast);

    // [AC5] Animation: Slide-in/Fade-in
    requestAnimationFrame(() => {
        toast.classList.remove('translate-x-full', 'opacity-0');
        toast.classList.add('translate-x-0', 'opacity-100');
    });

    // [AC4] Xử lý nút đóng thủ công
    const closeBtn = toast.querySelector('button');
    closeBtn.onclick = () => removeToast(toast);

    // [AC4] Tự động xóa sau 5 giây
    const autoHideTimeout = setTimeout(() => {
        removeToast(toast);
    }, AUTO_HIDE_MS);

    // Hàm xóa Toast kèm animation
    function removeToast(el) {
        clearTimeout(autoHideTimeout);
        el.addEventListener('transitionend', () => {
            el.remove();
        }, { once: true });
        el.classList.add('translate-x-full', 'opacity-0');
    }
}

// Khởi chạy khi DOM sẵn sàng [AC1]
if (typeof window !== 'undefined') {
    document.addEventListener('DOMContentLoaded', setupWebSocket);
}

// Export cho Unit Test (Node.js)
if (typeof module !== 'undefined') {
    module.exports = { showNotification, setupWebSocket };
}
