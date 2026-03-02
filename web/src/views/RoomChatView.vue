<script setup lang="ts">
import { useRoute } from "vue-router";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const route = useRoute();
const roomId = route.params.roomId as string;
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <Card class="w-full max-w-sm">
      <CardHeader>
        <CardTitle>Room {{ roomId }}</CardTitle>
      </CardHeader>
      <CardContent class="text-center text-sm text-muted-foreground">
        Chat interface coming soon…
      </CardContent>
    </Card>
  </div>
</template>
