<script setup lang="ts">
import { onMounted } from "vue";
import { Database, ScreenShare } from "lucide-vue-next";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import RoomCodeDisplay from "@/room/components/RoomCodeDisplay.vue";
import ConnectionStatus from "@/room/components/ConnectionStatus.vue";
import { useRoomStore } from "@/room/store";
import { useRoomSocket } from "@/room/useRoomSocket";

const store = useRoomStore();
const { error, connect } = useRoomSocket();

onMounted(() => {
  connect();
});
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <Card class="w-full max-w-sm">
      <CardHeader>
        <CardTitle>Room</CardTitle>
      </CardHeader>
      <CardContent class="flex flex-col items-center gap-6">
        <RoomCodeDisplay :code="store.roomCode ?? ''" />

        <div class="flex w-full flex-col gap-2">
          <ConnectionStatus
            label="Server"
            :icon="Database"
            :connected="store.serverConnected"
          />
          <ConnectionStatus
            label="Desktop"
            :icon="ScreenShare"
            :connected="store.desktopConnected"
          />
        </div>

        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
      </CardContent>
    </Card>
  </div>
</template>
