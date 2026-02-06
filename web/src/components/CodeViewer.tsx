import React, { useMemo } from "react";
import Prism from "prismjs";
import "prismjs/themes/prism.css";

// Import common languages
import "prismjs/components/prism-javascript";
import "prismjs/components/prism-typescript";
import "prismjs/components/prism-jsx";
import "prismjs/components/prism-tsx";
import "prismjs/components/prism-go";
import "prismjs/components/prism-python";
import "prismjs/components/prism-java";
import "prismjs/components/prism-c";
import "prismjs/components/prism-cpp";
import "prismjs/components/prism-csharp";
import "prismjs/components/prism-rust";
import "prismjs/components/prism-json";
import "prismjs/components/prism-yaml";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-sql";
import "prismjs/components/prism-css";

const languageByExt: Record<string, string> = {
  ".js": "javascript",
  ".jsx": "jsx",
  ".ts": "typescript",
  ".tsx": "tsx",
  ".go": "go",
  ".py": "python",
  ".java": "java",
  ".c": "c",
  ".cpp": "cpp",
  ".cs": "csharp",
  ".rs": "rust",
  ".json": "json",
  ".yaml": "yaml",
  ".yml": "yaml",
  ".html": "markup",
  ".css": "css",
  ".md": "markdown",
};

export function CodeViewer({ content, ext }: { content: string; ext?: string }) {
  const language = languageByExt[ext ?? ""] ?? "markup";
  const html = useMemo(() => {
    const grammar = Prism.languages[language] ?? Prism.languages.markup;
    return Prism.highlight(content, grammar, language);
  }, [content, language]);

  return (
    <pre
      style={{
        margin: 0,
        padding: "24px",
        fontFamily: 'Menlo, Monaco, "Courier New", monospace',
        fontSize: "13px",
        lineHeight: "1.6",
        color: "#0f172a",
        background: "#f8fafc",
      }}
    >
      <code dangerouslySetInnerHTML={{ __html: html }} />
    </pre>
  );
}
