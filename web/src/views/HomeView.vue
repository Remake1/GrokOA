<script setup lang="ts">
import { useRouter } from "vue-router";
import { Plus, Cog } from "lucide-vue-next";
import MenuCard from "@/components/MenuCard.vue";

const router = useRouter();

function onCreateRoom() {
  router.push("/room/setup");
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div class="flex gap-4">
      <MenuCard :icon="Plus" label="Create Room" @click="onCreateRoom" />
      <MenuCard :icon="Cog" label="Settings" />
    </div>
  </div>
</template>
