import { ref, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useRoomStore } from "./store";
import { getToken } from "@/auth/cookies";

interface ServerMessage {
    type: string;
    code?: string;
    message?: string;
}

export function useRoomSocket() {
    const store = useRoomStore();
    const router = useRouter();
    const error = ref<string | null>(null);

    let ws: WebSocket | null = null;

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

            case "desktop_connected":
                store.setDesktopConnected(true);
                router.push(`/room/${store.roomCode}`);
                break;

            case "desktop_disconnected":
                store.setDesktopConnected(false);
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
        };

        socket.onmessage = handleMessage;

        socket.onclose = () => {
            store.setServerConnected(false);
            ws = null;
        };
    }

    function connectFresh() {
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

    function disconnect() {
        if (ws) {
            ws.close();
            ws = null;
        }
    }

    onUnmounted(() => {
        disconnect();
    });

    return { error, connect, disconnect };
}
