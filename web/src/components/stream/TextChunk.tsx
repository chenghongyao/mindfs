import React from "react";

type TextChunkProps = {
  content: string;
};

export function TextChunk({ content }: TextChunkProps) {
  if (!content) return null;

  return (
    <div
      style={{
        fontSize: "13px",
        lineHeight: 1.6,
        color: "var(--text-primary)",
        whiteSpace: "pre-wrap",
        wordBreak: "break-word",
      }}
    >
      {content}
    </div>
  );
}
