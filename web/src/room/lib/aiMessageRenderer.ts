const SHIKI_LANGUAGES = [
  "text",
  "plaintext",
  "bash",
  "shell",
  "json",
  "javascript",
  "js",
  "jsx",
  "typescript",
  "ts",
  "tsx",
  "vue",
  "html",
  "css",
  "scss",
  "markdown",
  "md",
  "yaml",
  "yml",
  "xml",
  "diff",
  "cpp",
  "c",
  "go",
  "sql",
  "rust",
  "python",
  "py",
  "java",
] as const;

const SHIKI_THEMES = {
  light: "github-light",
  dark: "github-dark",
} as const;

interface AiMessageRenderer {
  render(markdown: string): Promise<string>;
}

let rendererPromise: Promise<AiMessageRenderer> | null = null;

async function createAiMessageRenderer(): Promise<AiMessageRenderer> {
  const [
    { default: MarkdownItAsync },
    { setupMarkdownWithCodeToHtml },
    { createHighlighter },
  ] = await Promise.all([
    import("markdown-it-async"),
    import("@shikijs/markdown-it/async"),
    import("shiki"),
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

  const highlighterPromise = createHighlighter({
    themes: Object.values(SHIKI_THEMES),
    langs: [...SHIKI_LANGUAGES],
  });

  setupMarkdownWithCodeToHtml(
    markdown,
    async (code, options) => {
      const highlighter = await highlighterPromise;
      return highlighter.codeToHtml(code, options);
    },
    {
      themes: SHIKI_THEMES,
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
