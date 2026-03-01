<script setup lang="ts">
import type { Component } from "vue";
import { Card, CardContent } from "@/components/ui/card";

defineProps<{
  icon: Component;
  label: string;
}>();

defineEmits<{
  click: [];
}>();
</script>

<template>
  <Card
    class="group w-40 cursor-pointer select-none transition-colors hover:border-primary/40"
    @click="$emit('click')"
  >
    <CardContent class="flex flex-col items-center gap-3 pt-6">
      <component
        :is="icon"
        class="size-10 text-card-foreground transition-colors group-hover:text-primary"
        :stroke-width="1.5"
      />
      <span class="text-sm font-medium text-card-foreground">
        {{ label }}
      </span>
    </CardContent>
  </Card>
</template>
