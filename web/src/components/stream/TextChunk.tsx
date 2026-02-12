import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

type TextChunkProps = {
  content: string;
};

export function TextChunk({ content }: TextChunkProps) {
  if (!content) return null;

  return (
    <div
      style={{
        fontSize: "15px",
        lineHeight: "1.7",
        color: "var(--text-primary)",
      }}
    >
      <ReactMarkdown 
        remarkPlugins={[remarkGfm]}
        components={{
          p: ({node, ...props}) => <p style={{ margin: "0 0 1em 0" }} {...props} />,
          pre: ({node, ...props}) => (
            <pre style={{ 
              background: "rgba(0,0,0,0.04)", 
              padding: "16px", 
              borderRadius: "8px", 
              overflow: "auto",
              fontSize: "13px",
              margin: "1em 0"
            }} {...props} />
          ),
          code: ({node, ...props}) => (
            <code style={{ 
              background: "rgba(0,0,0,0.04)", 
              padding: "2px 4px", 
              borderRadius: "4px" 
            }} {...props} />
          ),
          ul: ({node, ...props}) => <ul style={{ paddingLeft: "1.5em", marginBottom: "1em" }} {...props} />,
          ol: ({node, ...props}) => <ol style={{ paddingLeft: "1.5em", marginBottom: "1em" }} {...props} />,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
