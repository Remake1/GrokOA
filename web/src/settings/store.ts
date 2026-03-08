import { ref, computed } from "vue";
import { defineStore } from "pinia";

export const LANGUAGE_OPTIONS = [
    "C++ 20",
    "C++ 14",
    "C++ 11",
    "C",
    "Python",
    "Go",
    "JavaScript",
    "TypeScript",
    "React TypeScript",
    "React JavaScript",
    "Vue 3 TypeScript",
    "Vue 3 JavaScript",
    "Vue 2",
] as const;

export const AI_MODEL_OPTIONS = [
    "gpt-5.3-codex",
    "gpt-5.4-2026-03-05",
    "gpt-5.1-codex-mini",
    "gpt-5-mini-2025-08-07",
    "gemini-3-flash-preview",
    "gemini-3.1-pro-preview",
    "gemini-3.1-flash-lite-preview",
] as const;

export type CodingLanguage = (typeof LANGUAGE_OPTIONS)[number];
export type AiModel = (typeof AI_MODEL_OPTIONS)[number];

export const DEFAULT_LANGUAGE: CodingLanguage = "C++ 20";
export const DEFAULT_AI_MODEL: AiModel = "gemini-3.1-flash-lite-preview";

export const DEFAULT_CODING_PROMPT =
    "Analyze these images (which may be parts of a LeetCode-style problem) and provide a working solution. Explain your approach and provide the complete code solution.";

export const DEFAULT_MCQ_PROMPT =
    "Analyze these images and provide a solution to the problem you see.";

export const useSettingsStore = defineStore(
    "settings",
    () => {
        const codingLanguage = ref<CodingLanguage>(DEFAULT_LANGUAGE);
        const aiModel = ref<AiModel>(DEFAULT_AI_MODEL);
        const codingPrompt = ref(DEFAULT_CODING_PROMPT);
        const mcqPrompt = ref(DEFAULT_MCQ_PROMPT);

        const finalCodingPrompt = computed(
            () =>
                `${codingPrompt.value} Provide solution in ${codingLanguage.value}.`,
        );

        function resetToDefaults() {
            codingLanguage.value = DEFAULT_LANGUAGE;
            aiModel.value = DEFAULT_AI_MODEL;
            codingPrompt.value = DEFAULT_CODING_PROMPT;
            mcqPrompt.value = DEFAULT_MCQ_PROMPT;
        }

        return {
            codingLanguage,
            aiModel,
            codingPrompt,
            mcqPrompt,
            finalCodingPrompt,
            resetToDefaults,
        };
    },
    {
        persist: true,
    },
);
