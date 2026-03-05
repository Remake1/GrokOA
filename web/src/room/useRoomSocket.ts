import { ref, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useRoomStore } from "./store";
import { getToken } from "@/auth/cookies";

interface ServerMessage {
    type: string;
    code?: string;
    message?: string;
    id?: string;
    data?: string;
}

export function useRoomSocket() {
    const store = useRoomStore();
    const router = useRouter();
    const error = ref<string | null>(null);

    let ws: WebSocket | null = null;
    let manualDisconnect = false;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let reconnectDelay = 1000;
    const MAX_RECONNECT_DELAY = 10000;

    function getPersistedRoomCode(): string | null {
        if (store.roomCode) return store.roomCode;

        try {
            const raw = localStorage.getItem("room");
            if (raw) {
                const parsed = JSON.parse(raw);
                if (parsed.roomCode) {
                    store.setRoomCode(parsed.roomCode);
                    return parsed.roomCode;
                }
            }
        } catch {
            // ignore parse errors
        }

        return null;
    }

    function buildWsUrl(code?: string | null): string {
        const token = getToken();
        if (!token) {
            throw new Error("No auth token available");
        }

        const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
        const base = `${proto}//${window.location.host}/api/ws/client`;

        const params = new URLSearchParams({ token });
        if (code) {
            params.set("room", code);
        }

        return `${base}?${params.toString()}`;
    }

    function handleMessage(event: MessageEvent) {
        const data: ServerMessage = JSON.parse(event.data);

        switch (data.type) {
            case "room_created":
                if (data.code) {
                    store.setRoomCode(data.code);
                }
                break;

            case "room_rejoined":
                if (data.code) {
                    store.setRoomCode(data.code);
                }
                break;

            case "desktop_connected":
                store.setDesktopConnected(true);
                router.push(`/room/${store.roomCode}`);
                break;

            case "desktop_disconnected":
                store.setDesktopConnected(false);
                break;

            case "screenshot":
                if (data.id && data.data) {
                    store.addScreenshot(data.id, data.data);
                    store.addChatMessage(
                        "assistant",
                        `Screenshot captured — saved as ${data.id}.`,
                        data.id,
                    );
                }
                break;

            case "error":
                if (data.message?.includes("not found") || data.message?.includes("expired")) {
                    disconnect();
                    connectFresh();
                } else {
                    error.value = data.message ?? "Unknown error";
                }
                break;
        }
    }

    function attachHandlers(socket: WebSocket) {
        socket.onopen = () => {
            store.setServerConnected(true);
            reconnectDelay = 1000;
        };

        socket.onmessage = handleMessage;

        socket.onclose = () => {
            store.setServerConnected(false);
            ws = null;

            if (!manualDisconnect) {
                scheduleReconnect();
            }
        };
    }

    function scheduleReconnect() {
        if (reconnectTimer) return;

        reconnectTimer = setTimeout(() => {
            reconnectTimer = null;

            const code = getPersistedRoomCode();
            if (!code) {
                // No room to reconnect to — start fresh.
                connectFresh();
                return;
            }

            try {
                const url = buildWsUrl(code);
                ws = new WebSocket(url);
                attachHandlers(ws);

                ws.onerror = () => {
                    ws = null;
                    // Increase delay for next attempt.
                    reconnectDelay = Math.min(reconnectDelay * 2, MAX_RECONNECT_DELAY);
                    scheduleReconnect();
                };
            } catch {
                reconnectDelay = Math.min(reconnectDelay * 2, MAX_RECONNECT_DELAY);
                scheduleReconnect();
            }
        }, reconnectDelay);
    }

    function connectFresh() {
        manualDisconnect = false;
        store.reset();
        const url = buildWsUrl();
        ws = new WebSocket(url);

        attachHandlers(ws);

        ws.onerror = () => {
            error.value = "WebSocket connection failed";
        };
    }

    function connect() {
        if (ws) {
            disconnect();
        }

        manualDisconnect = false;
        error.value = null;

        try {
            const existingCode = getPersistedRoomCode();

            if (existingCode) {
                // Try to rejoin existing room first
                const url = buildWsUrl(existingCode);
                ws = new WebSocket(url);

                attachHandlers(ws);

                ws.onerror = () => {
                    // Rejoin failed — fall back to fresh room
                    ws = null;
                    connectFresh();
                };
            } else {
                connectFresh();
            }
        } catch (err) {
            error.value =
                err instanceof Error ? err.message : "Failed to connect";
        }
    }

    function requestScreenshot() {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            error.value = "Not connected to server";
            return;
        }

        ws.send(JSON.stringify({ type: "request_screenshot" }));
        store.addChatMessage("user", "Take screenshot");
    }

    function disconnect() {
        manualDisconnect = true;

        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }

        if (ws) {
            ws.close();
            ws = null;
        }
    }

    onUnmounted(() => {
        disconnect();
    });

    return { error, connect, disconnect, requestScreenshot };
}
