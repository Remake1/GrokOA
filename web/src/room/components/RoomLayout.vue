<script setup lang="ts">
import ActionSidebar from "./ActionSidebar.vue";
import ChatPanel from "./ChatPanel.vue";
import ImagePanel from "./ImagePanel.vue";
import type { ChatMessage, ImagePreview } from "@/room/types";

defineProps<{
  roomCode: string;
  messages: ChatMessage[];
  images: ImagePreview[];
  showImagePanel: boolean;
  serverConnected: boolean;
  desktopConnected: boolean;
  screenshotDisabled?: boolean;
  aiSubmitDisabled?: boolean;
}>();

defineEmits<{
  screenshot: [];
  toggleImages: [];
  submitMcq: [];
  submitCode: [];
  settings: [];
  reconnect: [];
  selectImage: [id: string];
  removeScreenshot: [id: string];
}>();
</script>

<template>
  <div class="flex h-dvh w-full overflow-hidden bg-background">
    <ActionSidebar
      :screenshot-disabled="screenshotDisabled"
      :ai-submit-disabled="aiSubmitDisabled"
      :server-connected="serverConnected"
      @screenshot="$emit('screenshot')"
      @submit-mcq="$emit('submitMcq')"
      @submit-code="$emit('submitCode')"
      @settings="$emit('settings')"
      @reconnect="$emit('reconnect')"
    />

    <ChatPanel
      :messages="messages"
      :room-code="roomCode"
      :server-connected="serverConnected"
      :desktop-connected="desktopConnected"
    />

    <ImagePanel
      :images="images"
      :collapsed="!showImagePanel"
      @toggle="$emit('toggleImages')"
      @select="$emit('selectImage', $event)"
      @remove="$emit('removeScreenshot', $event)"
    />
  </div>
</template>
