// frontend/scratch/send_udp_test.js
const dgram = require('dgram');
const client = dgram.createSocket('udp4');

const message = JSON.stringify({
    title: "One Piece",
    chapter: 1115,
    manga_id: "op01"
});

client.send(message, 8888, 'localhost', (err) => {
    console.log("Đã gửi gói tin UDP tới Bridge!");
    client.close();
});
