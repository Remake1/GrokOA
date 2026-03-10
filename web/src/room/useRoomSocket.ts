import { ref } from "vue";
import { useRoomStore } from "./store";
import { getToken } from "@/auth/cookies";
import router from "@/router";

interface ServerMessage {
    type: string;
    code?: string;
    message?: string;
    id?: string;
    data?: string;
    delta?: string;
}

const error = ref<string | null>(null);
let ws: WebSocket | null = null;
let manualDisconnect = false;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let reconnectDelay = 1000;
const MAX_RECONNECT_DELAY = 10000;

export function useRoomSocket() {
    const store = useRoomStore();

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
                        {
                            imageId: data.id,
                            kind: "screenshot",
                        },
                    );
                }
                break;

            case "ai_chat_chunk":
                if (typeof data.delta === "string") {
                    store.appendAiResponseChunk(data.delta);
                }
                break;

            case "ai_chat_done":
                store.finishAiResponseStream();
                break;

            case "error":
                if (store.aiRequestStatus === "streaming") {
                    store.failAiResponseStream();
                }

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
        manualDisconnect = false;
        error.value = null;

        if (ws?.readyState === WebSocket.OPEN || ws?.readyState === WebSocket.CONNECTING) {
            return;
        }

        if (ws) {
            ws.close();
            ws = null;
        }

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

        if (store.aiRequestStatus === "streaming") {
            error.value = "Wait for the current AI response to finish";
            return;
        }

        ws.send(JSON.stringify({ type: "request_screenshot" }));
        store.addChatMessage("user", "Take screenshot");
    }

    function requestAiResponse({
        model,
        prompt,
        screenshotIds,
        label,
    }: {
        model: string;
        prompt: string;
        screenshotIds: string[];
        label: string;
    }) {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            error.value = "Not connected to server";
            return false;
        }

        if (store.aiRequestStatus === "streaming") {
            error.value = "Wait for the current AI response to finish";
            return false;
        }

        error.value = null;

        ws.send(
            JSON.stringify({
                type: "ai_chat",
                model,
                prompt,
                screenshot_ids: screenshotIds,
            }),
        );

        store.addChatMessage("user", label);
        store.startAiResponseStream();
        store.clearScreenshots();

        return true;
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

    return { error, connect, disconnect, requestScreenshot, requestAiResponse };
}
