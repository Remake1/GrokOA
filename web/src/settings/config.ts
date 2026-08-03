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
  { value: "gpt-5.4", label: "gpt-5.4" },
  { value: "gpt-5.4-mini", label: "gpt-5.4-mini" },
  { value: "gpt-5.6-sol", label: "gpt-5.6-sol (medium)" },
  { value: "gpt-5.6-terra", label: "gpt-5.6-terra (medium)" },
  { value: "gemini-3-flash-preview", label: "gemini-3-flash-preview" },
  { value: "gemini-3.1-pro-preview", label: "gemini-3.1-pro-preview" },
  {
    value: "gemini-3.1-flash-lite-preview",
    label: "gemini-3.1-flash-lite-preview",
  },
] as const;

export type CodingLanguage = (typeof LANGUAGE_OPTIONS)[number];
export type AiModel = (typeof AI_MODEL_OPTIONS)[number]["value"];

export const DEFAULT_LANGUAGE: CodingLanguage = "C++ 20";
export const DEFAULT_AI_MODEL: AiModel = "gemini-3.1-flash-lite-preview";

export const DEFAULT_CODING_PROMPT =
  "Analyze these images (which may be parts of a LeetCode-style problem) and provide a working solution. Explain your approach and provide the complete code solution.";

export const DEFAULT_MCQ_PROMPT =
  "Analyze these images and provide a solution to the problem you see.";
