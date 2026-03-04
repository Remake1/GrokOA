<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useRoute } from "vue-router";
import RoomLayout from "@/room/components/RoomLayout.vue";
import { useRoomStore } from "@/room/store";
import { useRoomSocket } from "@/room/useRoomSocket";

const route = useRoute();
const store = useRoomStore();
const roomCode = computed(() => store.roomCode ?? (route.params.roomId as string));
const { error, connect, requestScreenshot } = useRoomSocket();

const MAX_SCREENSHOTS = 5;
const showImagePanel = ref(true);
const screenshotDisabled = computed(() => store.screenshots.length >= MAX_SCREENSHOTS);

function toggleImagePanel() {
  showImagePanel.value = !showImagePanel.value;
}

// Auto-expand panel when a new screenshot arrives
watch(
  () => store.screenshots.length,
  (newLen, oldLen) => {
    if (newLen > oldLen) {
      showImagePanel.value = true;
    }
  },
);

onMounted(() => {
  connect();
});
</script>

<template>
  <RoomLayout
    :room-code="roomCode"
    :messages="store.chatMessages"
    :images="store.screenshots"
    :show-image-panel="showImagePanel"
    :server-connected="store.serverConnected"
    :desktop-connected="store.desktopConnected"
    :screenshot-disabled="screenshotDisabled"
    @screenshot="requestScreenshot"
    @toggle-images="toggleImagePanel"
    @ask-ai="() => {}"
    @settings="() => {}"
    @reconnect="connect"
    @select-image="() => {}"
    @remove-screenshot="store.removeScreenshot"
  />

  <p v-if="error" class="fixed bottom-4 left-1/2 -translate-x-1/2 rounded-lg bg-destructive px-4 py-2 text-sm text-destructive-foreground shadow-lg">
    {{ error }}
  </p>
</template>
