<script setup lang="ts">
import type { Component } from "vue";

defineProps<{
  icon: Component;
  label: string;
  connected: boolean;
}>();
</script>

<template>
  <div class="flex items-center gap-2">
    <component
      :is="icon"
      class="size-4 text-muted-foreground"
      :stroke-width="1.5"
    />
    <span class="text-sm text-muted-foreground">{{ label }}</span>
    <span
      class="ml-auto text-xs"
      :class="connected ? 'text-green-500' : 'text-red-500'"
    >
      {{ connected ? "Connected" : "Disconnected" }}
    </span>
    <span
      class="size-2 rounded-full"
      :class="connected ? 'bg-green-500' : 'bg-red-500'"
    />
  </div>
</template>
