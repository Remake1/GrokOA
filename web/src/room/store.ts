import { ref } from "vue";
import { defineStore } from "pinia";
import type { ChatMessage, ImagePreview } from "./types";

let nextMsgId = 1;

export const useRoomStore = defineStore(
    "room",
    () => {
        const roomCode = ref<string | null>(null);
        const serverConnected = ref(false);
        const desktopConnected = ref(false);
        const screenshots = ref<ImagePreview[]>([]);
        const chatMessages = ref<ChatMessage[]>([]);
        const aiRequestStatus = ref<"idle" | "streaming">("idle");
        const activeAiMessageId = ref<string | null>(null);

        function setRoomCode(code: string) {
            roomCode.value = code;
        }

        function setServerConnected(value: boolean) {
            serverConnected.value = value;
        }

        function setDesktopConnected(value: boolean) {
            desktopConnected.value = value;
        }

        function addScreenshot(id: string, base64Data: string) {
            screenshots.value.push({
                id,
                src: `data:image/png;base64,${base64Data}`,
                alt: `Screenshot ${id}`,
                timestamp: new Date(),
            });
        }

        function removeScreenshot(id: string) {
            screenshots.value = screenshots.value.filter((s) => s.id !== id);
        }

        function addChatMessage(
            role: "user" | "assistant",
            content: string,
            options?: {
                imageId?: string;
                kind?: ChatMessage["kind"];
                streaming?: boolean;
                fullWidth?: boolean;
            },
        ) {
            const message: ChatMessage = {
                id: String(nextMsgId++),
                role,
                content,
                kind: options?.kind ?? "default",
                imageId: options?.imageId,
                streaming: options?.streaming ?? false,
                fullWidth: options?.fullWidth ?? false,
                timestamp: new Date(),
            };

            chatMessages.value.push(message);
            return message.id;
        }

        function startAiResponseStream() {
            aiRequestStatus.value = "streaming";

            const messageId = addChatMessage("assistant", "", {
                kind: "ai",
                streaming: true,
                fullWidth: true,
            });

            activeAiMessageId.value = messageId;
            return messageId;
        }

        function appendAiResponseChunk(delta: string) {
            if (!activeAiMessageId.value) return;

            const message = chatMessages.value.find(
                (entry) => entry.id === activeAiMessageId.value,
            );

            if (!message) return;

            message.content += delta;
        }

        function finishAiResponseStream() {
            if (activeAiMessageId.value) {
                const message = chatMessages.value.find(
                    (entry) => entry.id === activeAiMessageId.value,
                );

                if (message) {
                    message.streaming = false;
                }
            }

            activeAiMessageId.value = null;
            aiRequestStatus.value = "idle";
        }

        function failAiResponseStream() {
            finishAiResponseStream();
        }

        function clearScreenshots() {
            screenshots.value = [];
        }

        function reset() {
            roomCode.value = null;
            serverConnected.value = false;
            desktopConnected.value = false;
            screenshots.value = [];
            chatMessages.value = [];
            aiRequestStatus.value = "idle";
            activeAiMessageId.value = null;
        }

        return {
            roomCode,
            serverConnected,
            desktopConnected,
            screenshots,
            chatMessages,
            aiRequestStatus,
            setRoomCode,
            setServerConnected,
            setDesktopConnected,
            addScreenshot,
            removeScreenshot,
            addChatMessage,
            startAiResponseStream,
            appendAiResponseChunk,
            finishAiResponseStream,
            failAiResponseStream,
            clearScreenshots,
            reset,
        };
    },
    {
        persist: {
            pick: ["roomCode", "chatMessages"],
        },
    },
);
