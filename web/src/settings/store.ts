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

export type CodingLanguage = (typeof LANGUAGE_OPTIONS)[number];

export const DEFAULT_LANGUAGE: CodingLanguage = "C++ 20";

export const DEFAULT_CODING_PROMPT =
    "Analyze these images (which may be parts of a LeetCode-style problem) and provide a working solution. Explain your approach and provide the complete code solution.";

export const DEFAULT_MCQ_PROMPT =
    "Analyze these images and provide a solution to the problem you see.";

export const useSettingsStore = defineStore(
    "settings",
    () => {
        const codingLanguage = ref<CodingLanguage>(DEFAULT_LANGUAGE);
        const codingPrompt = ref(DEFAULT_CODING_PROMPT);
        const mcqPrompt = ref(DEFAULT_MCQ_PROMPT);

        const finalCodingPrompt = computed(
            () =>
                `${codingPrompt.value} Provide solution in ${codingLanguage.value}.`,
        );

        function resetToDefaults() {
            codingLanguage.value = DEFAULT_LANGUAGE;
            codingPrompt.value = DEFAULT_CODING_PROMPT;
            mcqPrompt.value = DEFAULT_MCQ_PROMPT;
        }

        return {
            codingLanguage,
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
