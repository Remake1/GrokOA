<script setup lang="ts">
import { computed, nextTick, ref, shallowRef, watch } from "vue";
import { Database, ScreenShare } from "lucide-vue-next";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { ChatMessage } from "@/room/types";
import { renderAiMessage } from "@/room/lib/aiMessageRenderer";

const props = defineProps<{
  messages: ChatMessage[];
  roomCode: string;
  serverConnected: boolean;
  desktopConnected: boolean;
}>();

const scrollRef = ref<InstanceType<typeof ScrollArea> | null>(null);
const renderedAiMessages = shallowRef<Record<string, string>>({});
const aiMessageVersions = new Map<string, string>();
const messageSignature = computed(() =>
  props.messages
    .map((msg) => `${msg.id}:${msg.content.length}:${msg.streaming ? 1 : 0}`)
    .join("|"),
);

watch(
  messageSignature,
  async () => {
    void updateRenderedAiMessages();
    await nextTick();
    const viewport = scrollRef.value?.$el?.querySelector(
      "[data-reka-scroll-area-viewport]",
    );
    if (viewport) {
      viewport.scrollTop = viewport.scrollHeight;
    }
  },
  { immediate: true },
);

function formatTime(date: Date | string): string {
  return new Date(date).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

function isAiMessage(message: ChatMessage): boolean {
  return message.kind === "ai";
}

function setRenderedAiMessage(messageId: string, html: string): void {
  renderedAiMessages.value = {
    ...renderedAiMessages.value,
    [messageId]: html,
  };
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function renderAiFallback(content: string): string {
  if (!content) return "";
  return `<p>${escapeHtml(content).replace(/\n/g, "<br />")}</p>`;
}

function pruneRenderedAiMessages(activeIds: Set<string>): void {
  const nextEntries = Object.entries(renderedAiMessages.value).filter(([id]) => activeIds.has(id));
  if (nextEntries.length === Object.keys(renderedAiMessages.value).length) return;
  renderedAiMessages.value = Object.fromEntries(nextEntries);
}

async function renderMessageHtml(message: ChatMessage, version: string): Promise<void> {
  try {
    const html = await renderAiMessage(message.content);
    if (aiMessageVersions.get(message.id) !== version) return;
    setRenderedAiMessage(message.id, html);
  } catch (error) {
    if (aiMessageVersions.get(message.id) !== version) return;
    console.error("Failed to render AI message markdown", error);
  }
}

async function updateRenderedAiMessages(): Promise<void> {
  const aiMessages = props.messages.filter(isAiMessage);
  const activeIds = new Set(aiMessages.map((message) => message.id));

  pruneRenderedAiMessages(activeIds);

  for (const [messageId] of aiMessageVersions) {
    if (!activeIds.has(messageId)) {
      aiMessageVersions.delete(messageId);
    }
  }

  for (const message of aiMessages) {
    const version = `${message.streaming ? 1 : 0}:${message.content}`;
    if (aiMessageVersions.get(message.id) === version) continue;

    aiMessageVersions.set(message.id, version);
    setRenderedAiMessage(message.id, renderAiFallback(message.content));

    if (!message.streaming && message.content) {
      void renderMessageHtml(message, version);
    }
  }
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
            class="rounded-lg px-3 py-2"
            :class="
              msg.kind === 'ai'
                ? 'w-full bg-muted text-foreground'
                : msg.role === 'user'
                  ? 'max-w-[75%] bg-primary text-primary-foreground'
                  : 'max-w-[75%] bg-muted text-foreground'
            "
          >
            <div
              v-if="isAiMessage(msg)"
              class="chat-markdown text-sm leading-relaxed"
              v-html="
                msg.streaming
                  ? renderAiFallback(msg.content)
                  : renderedAiMessages[msg.id] ?? renderAiFallback(msg.content)
              "
            />
            <p v-else class="text-sm leading-relaxed whitespace-pre-wrap">{{ msg.content }}</p>
            <span class="mt-1 block text-[10px] opacity-60">
              {{ formatTime(msg.timestamp) }}
              <span v-if="msg.streaming"> · Streaming</span>
            </span>
          </div>
        </div>
      </div>
    </ScrollArea>
  </div>
</template>

<style scoped>
.chat-markdown :deep(p + p),
.chat-markdown :deep(ul + p),
.chat-markdown :deep(ol + p),
.chat-markdown :deep(pre + p),
.chat-markdown :deep(blockquote + p) {
  margin-top: 0.75rem;
}

.chat-markdown :deep(ul),
.chat-markdown :deep(ol) {
  margin: 0.75rem 0;
  padding-left: 1.25rem;
}

.chat-markdown :deep(li + li) {
  margin-top: 0.25rem;
}

.chat-markdown :deep(a) {
  color: var(--color-primary);
  text-decoration: underline;
  text-underline-offset: 0.15em;
}

.chat-markdown :deep(blockquote) {
  margin: 0.75rem 0;
  border-left: 2px solid var(--color-border);
  padding-left: 0.75rem;
  color: var(--color-muted-foreground);
}

.chat-markdown :deep(code):not(:where(pre code)) {
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
  background: color-mix(in oklab, var(--color-background) 88%, var(--color-foreground) 12%);
  padding: 0.1rem 0.35rem;
  font-family: var(--font-mono);
  font-size: 0.875em;
}

.chat-markdown :deep(pre) {
  margin: 0.75rem 0;
  overflow-x: auto;
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
}

.chat-markdown :deep(pre code) {
  display: block;
  padding: 0.875rem 1rem;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  line-height: 1.6;
}

.chat-markdown :deep(pre.shiki) {
  margin: 0;
  min-width: 100%;
}

.chat-markdown :deep(pre.shiki code) {
  background: transparent;
}
</style>
