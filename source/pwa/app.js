// 注册 Service Worker
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/service-worker.js')
        .then(() => console.log('Service Worker registered'));
}

// 替换成你的 WebSocket 服务地址（本机或局域网 IP）
const ws = new WebSocket("ws://127.0.0.1:9000");

const messagesDiv = document.getElementById("messages");

ws.onopen = () => {
    console.log("WebSocket connected");
};

ws.onmessage = (event) => {
    const msg = document.createElement("div");
    msg.className = "msg";
    msg.textContent = event.data;
    messagesDiv.appendChild(msg);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
};

ws.onclose = () => console.log("WebSocket disconnected");
ws.onerror = (err) => console.error("WebSocket error:", err);