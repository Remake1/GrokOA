import { ref, computed } from "vue";
import { defineStore } from "pinia";
import {
  DEFAULT_AI_MODEL,
  DEFAULT_CODING_PROMPT,
  DEFAULT_LANGUAGE,
  DEFAULT_MCQ_PROMPT,
  type AiModel,
  type CodingLanguage,
} from "@/settings/config";

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
