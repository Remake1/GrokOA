<script setup lang="ts">
import { Camera, Images, Sparkles, Cog, RefreshCcw } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

defineProps<{
  screenshotDisabled?: boolean;
  serverConnected?: boolean;
}>();

defineEmits<{
  screenshot: [];
  toggleImages: [];
  askAi: [];
  settings: [];
  reconnect: [];
}>();
</script>

<template>
  <aside
    class="flex h-full w-14 shrink-0 flex-col items-center gap-1 border-r border-border bg-card py-3"
  >
    <TooltipProvider :delay-duration="300">
      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="text-card-foreground"
            :disabled="screenshotDisabled"
            @click="$emit('screenshot')"
          >
            <Camera class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">
          <p>Take Screenshot</p>
        </TooltipContent>
      </Tooltip>

      <Separator class="mx-2 w-6" />

      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="text-card-foreground"
            @click="$emit('toggleImages')"
          >
            <Images class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">
          <p>Screenshots</p>
        </TooltipContent>
      </Tooltip>

      <Separator class="mx-2 w-6" />

      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="text-card-foreground"
            @click="$emit('askAi')"
          >
            <Sparkles class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">
          <p>Ask AI</p>
        </TooltipContent>
      </Tooltip>

      <div class="mt-auto flex flex-col items-center gap-1">
        <Tooltip v-if="!serverConnected">
          <TooltipTrigger as-child>
            <Button
              variant="ghost"
              size="icon"
              class="text-destructive"
              @click="$emit('reconnect')"
            >
              <RefreshCcw class="size-5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="right">
            <p>Reconnect</p>
          </TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger as-child>
            <Button
              variant="ghost"
              size="icon"
              class="text-card-foreground"
              @click="$emit('settings')"
            >
              <Cog class="size-5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="right">
            <p>Settings</p>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  </aside>
</template>
