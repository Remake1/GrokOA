<script setup lang="ts">
import { Plus, Cog } from "lucide-vue-next";
import MenuCard from "@/components/MenuCard.vue";
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div class="flex gap-4">
      <MenuCard :icon="Plus" label="Create Room" />
      <MenuCard :icon="Cog" label="Settings" />
    </div>
  </div>
</template>
