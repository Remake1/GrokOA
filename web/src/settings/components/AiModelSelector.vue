<script setup lang="ts">
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

defineProps<{
    options: readonly string[];
}>();

const model = defineModel<string>({ required: true });
</script>

<template>
    <div class="flex flex-col gap-2">
        <Label>AI Model</Label>
        <Select v-model="model">
            <SelectTrigger id="ai-model-selector">
                <SelectValue placeholder="Select model" />
            </SelectTrigger>
            <SelectContent>
                <SelectItem
                    v-for="option in options"
                    :key="option"
                    :value="option"
                >
                    {{ option }}
                </SelectItem>
            </SelectContent>
        </Select>
    </div>
</template>
