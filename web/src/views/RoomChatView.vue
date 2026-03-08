<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useRoute } from "vue-router";
import RoomLayout from "@/room/components/RoomLayout.vue";
import SettingsModal from "@/room/components/SettingsModal.vue";
import { useRoomStore } from "@/room/store";
import { useRoomSocket } from "@/room/useRoomSocket";
import { useSettingsStore } from "@/settings/store";

const route = useRoute();
const store = useRoomStore();
const settings = useSettingsStore();
const roomCode = computed(() => store.roomCode ?? (route.params.roomId as string));
const { error, connect, disconnect, requestScreenshot, requestAiResponse } = useRoomSocket();

const MAX_SCREENSHOTS = 5;
const showImagePanel = ref(true);
const showSettings = ref(false);
const aiStreaming = computed(() => store.aiRequestStatus === "streaming");
const screenshotDisabled = computed(
  () => store.screenshots.length >= MAX_SCREENSHOTS || aiStreaming.value,
);
const aiSubmitDisabled = computed(
  () =>
    !store.serverConnected ||
    aiStreaming.value ||
    store.screenshots.length === 0,
);

function toggleImagePanel() {
  showImagePanel.value = !showImagePanel.value;
}

function submitAiQuestion(prompt: string, label: string) {
  const submitted = requestAiResponse({
    model: settings.aiModel,
    prompt,
    screenshotIds: store.screenshots.map((image) => image.id),
    label,
  });

  if (submitted) {
    showImagePanel.value = false;
  }
}

function submitMcqQuestion() {
  submitAiQuestion(settings.mcqPrompt, "Submit MCQ question");
}

function submitCodeQuestion() {
  submitAiQuestion(settings.finalCodingPrompt, "Submit code question");
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
    :ai-submit-disabled="aiSubmitDisabled"
    @screenshot="requestScreenshot"
    @toggle-images="toggleImagePanel"
    @submit-mcq="submitMcqQuestion"
    @submit-code="submitCodeQuestion"
    @settings="showSettings = true"
    @reconnect="connect"
    @select-image="() => {}"
    @remove-screenshot="store.removeScreenshot"
  />

  <SettingsModal v-model:open="showSettings" :disconnect="disconnect" />

  <p v-if="error" class="fixed bottom-4 left-1/2 -translate-x-1/2 rounded-lg bg-destructive px-4 py-2 text-sm text-destructive-foreground shadow-lg">
    {{ error }}
  </p>
</template>
