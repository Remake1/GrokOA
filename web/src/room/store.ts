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
            imageId?: string,
        ) {
            chatMessages.value.push({
                id: String(nextMsgId++),
                role,
                content,
                imageId,
                timestamp: new Date(),
            });
        }

        function reset() {
            roomCode.value = null;
            serverConnected.value = false;
            desktopConnected.value = false;
            screenshots.value = [];
            chatMessages.value = [];
        }

        return {
            roomCode,
            serverConnected,
            desktopConnected,
            screenshots,
            chatMessages,
            setRoomCode,
            setServerConnected,
            setDesktopConnected,
            addScreenshot,
            removeScreenshot,
            addChatMessage,
            reset,
        };
    },
    {
        persist: {
            pick: ["roomCode", "chatMessages"],
        },
    },
);
