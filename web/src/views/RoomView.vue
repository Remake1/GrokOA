<script setup lang="ts">
import { Database, ScreenShare } from "lucide-vue-next";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import RoomCodeDisplay from "@/room/components/RoomCodeDisplay.vue";
import ConnectionStatus from "@/room/components/ConnectionStatus.vue";
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <Card class="w-full max-w-sm">
      <CardHeader>
        <CardTitle>Room</CardTitle>
      </CardHeader>
      <CardContent class="flex flex-col items-center gap-6">
        <RoomCodeDisplay code="A1B2" />

        <div class="flex w-full flex-col gap-2">
          <ConnectionStatus label="Server" :icon="Database" :connected="true" />
          <ConnectionStatus label="Desktop" :icon="ScreenShare" :connected="false" />
        </div>
      </CardContent>
    </Card>
  </div>
</template>
