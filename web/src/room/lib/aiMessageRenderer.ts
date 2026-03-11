import javascript from "@shikijs/langs/javascript";
import jsx from "@shikijs/langs/jsx";
import typescript from "@shikijs/langs/typescript";
import tsx from "@shikijs/langs/tsx";
import css from "@shikijs/langs/css";
import html from "@shikijs/langs/html";
import vue from "@shikijs/langs/vue";
import c from "@shikijs/langs/c";
import cpp from "@shikijs/langs/cpp";
import go from "@shikijs/langs/go";
import sql from "@shikijs/langs/sql";
import githubDark from "@shikijs/themes/github-dark";
import { createHighlighterCore } from "shiki/core";
import { createJavaScriptRegexEngine } from "shiki/engine/javascript";

const SHIKI_THEME = githubDark;
const SHIKI_LANGUAGES = [
  ...javascript,
  ...jsx,
  ...typescript,
  ...tsx,
  ...css,
  ...html,
  ...vue,
  ...c,
  ...cpp,
  ...go,
  ...sql,
] as const;

const SHIKI_LANGUAGE_ALIASES: Record<string, string> = {
  js: "javascript",
  javascript: "javascript",
  jsx: "jsx",
  ts: "typescript",
  typescript: "typescript",
  tsx: "tsx",
  css: "css",
  html: "html",
  vue: "vue",
  c: "c",
  cpp: "cpp",
  "c++": "cpp",
  go: "go",
  golang: "go",
  sql: "sql",
};

interface AiMessageRenderer {
  render(markdown: string): Promise<string>;
}

let rendererPromise: Promise<AiMessageRenderer> | null = null;

function normalizeCodeLanguage(lang?: string | null): string {
  if (!lang) {
    return "text";
  }

  return SHIKI_LANGUAGE_ALIASES[lang.trim().toLowerCase()] ?? "text";
}

async function createAiMessageRenderer(): Promise<AiMessageRenderer> {
  const [{ default: MarkdownItAsync }, { setupMarkdownWithCodeToHtml }] =
    await Promise.all([
    import("markdown-it-async"),
    import("@shikijs/markdown-it/async"),
  ]);

  const markdown = MarkdownItAsync({
    html: false,
    breaks: true,
    linkify: true,
    typographer: true,
  });

  const defaultLinkOpen =
    markdown.renderer.rules.link_open ??
    ((tokens, index, options, _env, self) =>
      self.renderToken(tokens, index, options));

  markdown.renderer.rules.link_open = (tokens, index, options, env, self) => {
    const token = tokens[index];
    if (token) {
      token.attrSet("target", "_blank");
      token.attrSet("rel", "noreferrer noopener");
    }
    return defaultLinkOpen(tokens, index, options, env, self);
  };

  const highlighterPromise = createHighlighterCore({
    themes: [SHIKI_THEME],
    langs: [...SHIKI_LANGUAGES],
    engine: createJavaScriptRegexEngine(),
  });

  setupMarkdownWithCodeToHtml(
    markdown,
    async (code, options) => {
      const highlighter = await highlighterPromise;
      return highlighter.codeToHtml(code, {
        ...options,
        lang: normalizeCodeLanguage(options.lang),
        theme: SHIKI_THEME,
      });
    },
    {
      theme: SHIKI_THEME,
    },
  );

  return {
    async render(content: string): Promise<string> {
      return markdown.renderAsync(content);
    },
  };
}

export async function renderAiMessage(markdown: string): Promise<string> {
  rendererPromise ??= createAiMessageRenderer();
  const renderer = await rendererPromise;
  return renderer.render(markdown);
}
