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
        <Label>Default Language</Label>
        <Select v-model="model">
            <SelectTrigger id="language-selector">
                <SelectValue placeholder="Select language" />
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
