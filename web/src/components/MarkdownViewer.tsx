import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeSanitize from "rehype-sanitize";
import Prism from "prismjs";
import "prismjs/themes/prism.css";
// Reuse the language imports from global Prism context (since they are imported in CodeViewer, they might be available if loaded, 
// but strictly speaking we should import them here or centralize. For simplicity, we rely on the side-effects of CodeViewer imports 
// if both are used, or we re-import essential ones here to be safe)
import "prismjs/components/prism-javascript";
import "prismjs/components/prism-typescript";
import "prismjs/components/prism-jsx";
import "prismjs/components/prism-tsx";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-json";

const monoFontFamily = [
  '"SFMono-Regular"',
  '"Cascadia Mono"',
  '"Sarasa Mono SC"',
  '"Noto Sans Mono CJK SC"',
  '"Source Han Mono SC"',
  'Menlo',
  'Monaco',
  '"Courier New"',
  'monospace',
].join(", ");

export function MarkdownViewer({ content }: { content: string }) {
  return (
    <div
      style={{
        padding: "0", // 移除内层 padding，由 FileViewer 统一控制
        color: "var(--text-primary)",
        lineHeight: 1.75,
        fontSize: "15px",
      }}
    >
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeSanitize]}
        components={{
          h1: (props) => (
            <h1 style={{ fontSize: "24px", marginTop: 0 }} {...props} />
          ),
          h2: (props) => (
            <h2 style={{ fontSize: "20px" }} {...props} />
          ),
          h3: (props) => (
            <h3 style={{ fontSize: "17px", marginTop: "1.25em" }} {...props} />
          ),
          p: (props) => (
            <p style={{ margin: "0 0 1em", whiteSpace: "normal" }} {...props} />
          ),
          ul: (props) => (
            <ul style={{ margin: "0 0 1em", paddingLeft: "1.4em" }} {...props} />
          ),
          ol: (props) => (
            <ol style={{ margin: "0 0 1em", paddingLeft: "1.4em" }} {...props} />
          ),
          li: (props) => (
            <li style={{ margin: "0.2em 0" }} {...props} />
          ),
          table: (props) => (
            <div
              style={{
                width: "100%",
                overflowX: "auto",
                margin: "1.25em 0",
                border: "1px solid var(--border-color)",
                borderRadius: "10px",
                background: "rgba(0,0,0,0.02)",
              }}
            >
              <table
                style={{
                  width: "100%",
                  borderCollapse: "collapse",
                  fontSize: "14px",
                  lineHeight: 1.6,
                  minWidth: "520px",
                }}
                {...props}
              />
            </div>
          ),
          thead: (props) => (
            <thead
              style={{
                background: "rgba(0,0,0,0.04)",
              }}
              {...props}
            />
          ),
          tr: (props) => (
            <tr
              style={{
                borderBottom: "1px solid var(--border-color)",
              }}
              {...props}
            />
          ),
          th: (props) => (
            <th
              style={{
                padding: "10px 12px",
                textAlign: "left",
                fontWeight: 600,
                whiteSpace: "nowrap",
                verticalAlign: "top",
              }}
              {...props}
            />
          ),
          td: (props) => (
            <td
              style={{
                padding: "10px 12px",
                verticalAlign: "top",
                borderTop: "1px solid rgba(0,0,0,0.03)",
              }}
              {...props}
            />
          ),
          blockquote: (props) => (
            <blockquote style={{ 
              borderLeft: "3px solid var(--accent-color)", 
              margin: "1.5em 0", 
              paddingLeft: "16px", 
              color: "var(--text-secondary)",
              fontStyle: "italic",
              background: "rgba(0,0,0,0.02)",
              padding: "12px 16px",
              borderRadius: "0 8px 8px 0"
            }} {...props} />
          ),
          code({ className, children, ...props }: any) {
            return (
              <code
                className={className}
                style={{
                  background: "rgba(0,0,0,0.05)",
                  padding: "2px 4px",
                  borderRadius: "4px",
                  color: "inherit",
                  fontFamily: monoFontFamily,
                  fontSize: "0.9em",
                }}
                {...props}
              >
                {children}
              </code>
            );
          },
          pre: ({ children }: any) => {
            const codeElement = React.Children.only(children) as React.ReactElement<any>;
            const className = codeElement?.props?.className || "";
            const rawContent = String(codeElement?.props?.children ?? "").replace(/\n$/, "");
            const match = /language-(\w+)/.exec(className);
            const language = match ? match[1] : "";
            let html = "";
            if (language) {
              const grammar = Prism.languages[language] ?? Prism.languages.markup;
              try {
                html = Prism.highlight(rawContent, grammar, language);
              } catch {
                html = "";
              }
            }
            return (
              <pre
                style={{
                  width: "100%",
                  boxSizing: "border-box",
                  background: "rgba(0,0,0,0.04)",
                  padding: "16px",
                  borderRadius: "10px",
                  overflow: "auto",
                  border: "1px solid var(--border-color)",
                  fontFamily: monoFontFamily,
                  fontSize: "13px",
                  margin: "1.5em 0",
                  lineHeight: "1.6",
                  whiteSpace: "pre",
                  tabSize: 2 as any,
                  fontVariantLigatures: "none",
                  boxShadow: "none",
                }}
              >
                {html ? (
                  <code
                    className={className}
                    dangerouslySetInnerHTML={{ __html: html }}
                    style={{ display: "block", textShadow: "none", fontFamily: monoFontFamily }}
                  />
                ) : (
                  <code
                    className={className}
                    style={{
                      display: "block",
                      textShadow: "none",
                      fontFamily: monoFontFamily,
                      tabSize: 2 as any,
                      fontVariantLigatures: "none",
                      whiteSpace: "pre",
                    }}
                  >
                    {rawContent}
                  </code>
                )}
              </pre>
            );
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
