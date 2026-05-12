const CONFIG = {
    // Tự động lấy IP của máy đang chạy server (máy tính của bạn)
    // Nếu bạn truy cập qua localhost, nó sẽ là localhost
    // Nếu bạn truy cập qua IP 10.x.x.x, nó sẽ là 10.x.x.x
    API_BASE_URL: `${window.location.protocol}//${window.location.hostname}:8080`,
    WS_BASE_URL: `ws://${window.location.hostname}:8080`,
    TCP_PORT: 9090,
    UDP_PORT: 9999,
    // Hàm tự động sửa lỗi URL hình ảnh
    fixUrl: (url) => {
        if (!url) return "https://placehold.co/150x200?text=No+Image";
        
    // 1. Xử lý ảnh MangaDex - Dùng link trực tiếp (sẽ kết hợp với referrerpolicy ở HTML)
        if (typeof url === 'string' && url.includes("mangadex.org")) {
            return url;
        }

        // 2. Xử lý đường dẫn local (bắt đầu bằng /data hoặc data/)
        if (typeof url === 'string' && (url.startsWith("/data") || url.startsWith("data/"))) {
            const path = url.startsWith("/") ? url : "/" + url;
            return CONFIG.API_BASE_URL + path;
        } 

        // 3. Xử lý localhost cứng
        if (typeof url === 'string' && url.startsWith("http://localhost:8080")) {
            return url.replace("http://localhost:8080", CONFIG.API_BASE_URL);
        }

        return url;
    }
};

console.log("[Config] Loaded:", CONFIG);
