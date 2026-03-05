<script setup lang="ts">
import { ref, nextTick, watch } from "vue";
import { Database, ScreenShare } from "lucide-vue-next";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { ChatMessage } from "@/room/types";

const props = defineProps<{
  messages: ChatMessage[];
  roomCode: string;
  serverConnected: boolean;
  desktopConnected: boolean;
}>();

const scrollRef = ref<InstanceType<typeof ScrollArea> | null>(null);

watch(
  () => props.messages.length,
  async () => {
    await nextTick();
    const viewport = scrollRef.value?.$el?.querySelector(
      "[data-reka-scroll-area-viewport]",
    );
    if (viewport) {
      viewport.scrollTop = viewport.scrollHeight;
    }
  },
);

function formatTime(date: Date | string): string {
  return new Date(date).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}
</script>

<template>
  <div class="flex min-h-0 min-w-0 flex-1 flex-col">
    <!-- Header badge -->
    <div class="shrink-0 p-3">
      <div
        class="inline-flex items-center gap-3 rounded-lg border border-border bg-card px-3 py-1.5"
      >
        <span class="text-xs font-semibold font-mono text-card-foreground">
          {{ roomCode }}
        </span>

        <div class="flex items-center gap-1.5">
          <Database class="size-3.5 text-muted-foreground" />
          <span
            class="size-2 rounded-full"
            :class="serverConnected ? 'bg-green-500' : 'bg-red-500'"
          />
        </div>

        <div class="flex items-center gap-1.5">
          <ScreenShare class="size-3.5 text-muted-foreground" />
          <span
            class="size-2 rounded-full"
            :class="desktopConnected ? 'bg-green-500' : 'bg-red-500'"
          />
        </div>
      </div>
    </div>

    <!-- Messages -->
    <ScrollArea ref="scrollRef" class="flex-1 overflow-hidden">
      <div class="flex flex-col gap-3 p-4">
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="flex"
          :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
        >
          <div
            class="max-w-[75%] rounded-lg px-3 py-2"
            :class="
              msg.role === 'user'
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-foreground'
            "
          >
            <p class="text-sm leading-relaxed whitespace-pre-wrap">{{ msg.content }}</p>
            <span
              class="mt-1 block text-[10px] opacity-60"
            >
              {{ formatTime(msg.timestamp) }}
            </span>
          </div>
        </div>
      </div>
    </ScrollArea>
  </div>
</template>

