<script setup lang="ts">
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { RotateCcw } from "lucide-vue-next";

const props = defineProps<{
    label: string;
    defaultValue: string;
}>();

const model = defineModel<string>({ required: true });

function reset() {
    model.value = props.defaultValue;
}
</script>

<template>
    <div class="flex flex-col gap-2">
        <div class="flex items-center justify-between">
            <Label>{{ label }}</Label>
            <Button
                variant="ghost"
                size="sm"
                class="h-7 gap-1.5 text-xs text-muted-foreground"
                @click="reset"
            >
                <RotateCcw class="size-3" />
                Reset
            </Button>
        </div>
        <Textarea
            v-model="model"
            :placeholder="defaultValue"
            class="min-h-[100px] resize-y"
        />
    </div>
</template>
