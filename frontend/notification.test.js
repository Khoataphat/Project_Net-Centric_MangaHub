/**
 * Unit Test for notification.js
 * Phiên bản tối giản không phụ thuộc thư viện ngoài (JSDOM)
 */

// 1. Mock môi trường Browser tối giản
global.window = {
    addEventListener: () => {},
    setTimeout: setTimeout,
    clearTimeout: clearTimeout,
    requestAnimationFrame: (cb) => setTimeout(cb, 0)
};
global.requestAnimationFrame = global.window.requestAnimationFrame;
global.setTimeout = setTimeout;
global.clearTimeout = clearTimeout;

global.document = {
    getElementById: (id) => {
        if (id === 'notification-container') return global.mockContainer;
        return null;
    },
    createElement: (tag) => {
        const el = {
            tagName: tag,
            className: '',
            innerHTML: '',
            children: [],
            classList: {
                add: (...classes) => {
                    classes.forEach(c => {
                        el.className += ' ' + c;
                        // Giả lập transitionend khi thêm class animation ẩn
                        if (c === 'opacity-0' && el.onTransitionEnd) {
                            setTimeout(() => el.onTransitionEnd(), 10);
                        }
                    });
                },
                remove: (c) => el.className = el.className.replace(c, '').trim()
            },
            querySelector: (sel) => {
                if (sel === 'button') return el.button;
                return null;
            },
            addEventListener: (evt, cb) => {
                if (evt === 'transitionend') el.onTransitionEnd = cb;
            },
            remove: () => {
                global.mockContainer.children = global.mockContainer.children.filter(c => c !== el);
            },
            button: {
                click: () => el.button.onclick && el.button.onclick()
            }
        };
        return el;
    },
    addEventListener: () => {}
};

global.mockContainer = {
    children: [],
    appendChild: (el) => global.mockContainer.children.push(el),
    get lastElementChild() { return this.children[this.children.length - 1]; }
};

// Mock WebSocket
global.WebSocket = class {
    constructor(url) {
        this.url = url;
        setTimeout(() => this.onopen && this.onopen(), 0);
    }
    send(data) {}
};

const { showNotification, setupWebSocket } = require('./notification');

async function runTests() {
    console.log('--- STARTING NOTIFICATION UNIT TESTS (MINIMAL MOCK) ---');
    let passed = 0;
    let failed = 0;

    const assert = (condition, message) => {
        if (condition) {
            console.log(`[PASS] ${message}`);
            passed++;
        } else {
            console.error(`[FAIL] ${message}`);
            failed++;
        }
    };

    // [AC1] Kiểm tra khởi tạo WebSocket
    try {
        setupWebSocket();
        assert(true, 'AC1: WebSocket setup function called without error');
    } catch (e) {
        assert(false, 'AC1: WebSocket setup failed: ' + e.message);
    }

    // [AC2, AC3] Kiểm tra hiển thị Toast khi nhận data
    const testData = {
        type: 'NEW_CHAPTER',
        title: 'One Piece',
        chapter: 1111
    };

    showNotification(testData);

    const toast = global.mockContainer.lastElementChild;

    assert(toast !== undefined, 'AC2: Toast element created in container');
    assert(toast.innerHTML.includes('One Piece'), 'AC3: Toast displays Manga Title');
    assert(toast.innerHTML.includes('1111'), 'AC3: Toast displays Chapter Number');

    // [AC4] Kiểm tra nút đóng
    toast.button.click();
    
    // Giả lập kết thúc animation
    if (toast.onTransitionEnd) toast.onTransitionEnd();
    
    assert(global.mockContainer.children.length === 0, 'AC4: Toast removed after clicking close button');

    // [AC4] Kiểm tra tự động đóng sau 5s
    showNotification(testData);
    assert(global.mockContainer.children.length === 1, 'Preparing AC4: Auto-hide test');
    
    console.log('Waiting 5.5s for auto-hide test...');
    await new Promise(resolve => setTimeout(resolve, 5500));
    assert(global.mockContainer.children.length === 0, 'AC4: Toast removed automatically after 5 seconds');

    console.log('\n--- TEST SUMMARY ---');
    console.log(`Total: ${passed + failed}`);
    console.log(`Passed: ${passed}`);
    console.log(`Failed: ${failed}`);

    if (failed > 0) process.exit(1);
}

runTests().catch(err => {
    console.error('Test Execution Error:', err);
    process.exit(1);
});
