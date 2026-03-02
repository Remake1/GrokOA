import { ref } from "vue";
import { defineStore } from "pinia";

export const useRoomStore = defineStore(
    "room",
    () => {
        const roomCode = ref<string | null>(null);
        const serverConnected = ref(false);
        const desktopConnected = ref(false);

        function setRoomCode(code: string) {
            roomCode.value = code;
        }

        function setServerConnected(value: boolean) {
            serverConnected.value = value;
        }

        function setDesktopConnected(value: boolean) {
            desktopConnected.value = value;
        }

        function reset() {
            roomCode.value = null;
            serverConnected.value = false;
            desktopConnected.value = false;
        }

        return {
            roomCode,
            serverConnected,
            desktopConnected,
            setRoomCode,
            setServerConnected,
            setDesktopConnected,
            reset,
        };
    },
    {
        persist: {
            pick: ["roomCode"],
        },
    },
);
