<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { ArrowLeft, Save } from "lucide-vue-next";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import LanguageSelector from "@/settings/components/LanguageSelector.vue";
import PromptEditor from "@/settings/components/PromptEditor.vue";
import {
    useSettingsStore,
    LANGUAGE_OPTIONS,
    DEFAULT_CODING_PROMPT,
    DEFAULT_MCQ_PROMPT,
} from "@/settings/store";

const router = useRouter();
const settings = useSettingsStore();

// Local draft state — only committed on save
const draftLanguage = ref(settings.codingLanguage);
const draftCodingPrompt = ref(settings.codingPrompt);
const draftMcqPrompt = ref(settings.mcqPrompt);

const hasChanges = computed(
    () =>
        draftLanguage.value !== settings.codingLanguage ||
        draftCodingPrompt.value !== settings.codingPrompt ||
        draftMcqPrompt.value !== settings.mcqPrompt,
);

function save() {
    settings.codingLanguage = draftLanguage.value;
    settings.codingPrompt = draftCodingPrompt.value;
    settings.mcqPrompt = draftMcqPrompt.value;
}
</script>

<template>
    <div class="flex min-h-screen items-center justify-center p-4">
        <Card class="w-full max-w-lg">
            <CardHeader class="flex flex-row items-center gap-3">
                <Button
                    id="settings-back"
                    variant="ghost"
                    size="icon"
                    class="size-8 shrink-0"
                    @click="router.push('/')"
                >
                    <ArrowLeft class="size-4" />
                </Button>
                <CardTitle>Settings</CardTitle>
            </CardHeader>

            <CardContent class="flex flex-col gap-6">
                <!-- Language selector -->
                <LanguageSelector
                    v-model="draftLanguage"
                    :options="LANGUAGE_OPTIONS"
                />

                <Separator />

                <!-- Coding prompt -->
                <PromptEditor
                    v-model="draftCodingPrompt"
                    label="Coding Prompt"
                    :default-value="DEFAULT_CODING_PROMPT"
                />

                <Separator />

                <!-- MCQ prompt -->
                <PromptEditor
                    v-model="draftMcqPrompt"
                    label="MCQ Prompt"
                    :default-value="DEFAULT_MCQ_PROMPT"
                />

                <Separator />

                <!-- Save -->
                <Button
                    id="settings-save"
                    class="w-full gap-2"
                    :disabled="!hasChanges"
                    @click="save"
                >
                    <Save class="size-4" />
                    Save
                </Button>
            </CardContent>
        </Card>
    </div>
</template>
