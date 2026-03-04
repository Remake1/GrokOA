<script setup lang="ts">
import { ChevronRight, ChevronLeft, X } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { ImagePreview } from "@/room/types";

defineProps<{
  images: ImagePreview[];
  collapsed: boolean;
}>();

defineEmits<{
  toggle: [];
  select: [id: string];
  remove: [id: string];
}>();
</script>

<template>
  <aside class="flex h-full shrink-0 flex-col border-l border-border bg-card transition-[width] duration-200"
    :class="collapsed ? 'w-10' : 'w-44'"
  >
    <!-- Header -->
    <div
      class="flex h-12 items-center border-b border-border"
      :class="collapsed ? 'justify-center px-1' : 'justify-between px-4'"
    >
      <h3 v-if="!collapsed" class="text-sm font-semibold text-card-foreground">Screenshots</h3>
      <Button
        variant="ghost"
        size="icon-sm"
        class="text-muted-foreground"
        @click="$emit('toggle')"
      >
        <ChevronRight v-if="!collapsed" class="size-4" />
        <ChevronLeft v-else class="size-4" />
      </Button>
    </div>

    <!-- Thumbnails (hidden when collapsed) -->
    <ScrollArea v-if="!collapsed" class="flex-1">
      <div class="flex flex-col gap-3 p-3">
        <div
          v-for="img in images"
          :key="img.id"
          class="group relative overflow-hidden rounded-md border border-border transition-colors hover:border-primary/40"
        >
          <button
            class="w-full cursor-pointer"
            @click="$emit('select', img.id)"
          >
            <img
              :src="img.src"
              :alt="img.alt"
              class="aspect-video w-full object-cover"
            />
          </button>

          <!-- Delete button -->
          <button
            class="absolute top-1 right-1 flex size-5 cursor-pointer items-center justify-center rounded-full bg-black/60 text-white transition-opacity hover:bg-red-600"
            title="Remove screenshot"
            @click.stop="$emit('remove', img.id)"
          >
            <X class="size-3" />
          </button>

          <span
            class="absolute bottom-0 left-0 w-full bg-black/50 px-2 py-1 text-[11px] font-mono text-white"
          >
            {{ img.id }}
          </span>
        </div>

        <p
          v-if="images.length === 0"
          class="py-8 text-center text-xs text-muted-foreground"
        >
          No screenshots yet
        </p>
      </div>
    </ScrollArea>
  </aside>
</template>
