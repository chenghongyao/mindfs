import React, { useMemo } from "react";
import { TextChunk } from "./TextChunk";
import { ThinkingBlock } from "./ThinkingBlock";
import { ToolCallCard } from "./ToolCallCard";

export type StreamChunkData = {
  type: "text" | "thinking" | "tool_call" | "tool_result" | "progress" | "done" | "error";
  content?: string;
  tool?: string;
  callId?: string;
  status?: string;
  result?: string;
  percent?: number;
  error?: string;
};

type StreamMessageProps = {
  chunks: StreamChunkData[];
  isStreaming?: boolean;
};

type GroupedContent = {
  type: "text" | "thinking" | "tool";
  content?: string;
  tool?: string;
  callId?: string;
  status?: string;
  result?: string;
};

export function StreamMessage({ chunks, isStreaming = false }: StreamMessageProps) {
  // Group consecutive chunks of the same type
  const grouped = useMemo(() => {
    const result: GroupedContent[] = [];
    let currentText = "";
    let currentThinking = "";
    const toolCalls = new Map<string, GroupedContent>();

    for (const chunk of chunks) {
      switch (chunk.type) {
        case "text":
          if (currentThinking) {
            result.push({ type: "thinking", content: currentThinking });
            currentThinking = "";
          }
          currentText += chunk.content || "";
          break;

        case "thinking":
          if (currentText) {
            result.push({ type: "text", content: currentText });
            currentText = "";
          }
          currentThinking += chunk.content || "";
          break;

        case "tool_call":
          if (currentText) {
            result.push({ type: "text", content: currentText });
            currentText = "";
          }
          if (currentThinking) {
            result.push({ type: "thinking", content: currentThinking });
            currentThinking = "";
          }
          if (chunk.callId) {
            toolCalls.set(chunk.callId, {
              type: "tool",
              tool: chunk.tool,
              callId: chunk.callId,
              status: "running",
            });
          }
          break;

        case "tool_result":
          if (chunk.callId && toolCalls.has(chunk.callId)) {
            const tc = toolCalls.get(chunk.callId)!;
            tc.status = "complete";
            tc.result = chunk.result;
          }
          break;
      }
    }

    // Flush remaining content
    if (currentText) {
      result.push({ type: "text", content: currentText });
    }
    if (currentThinking) {
      result.push({ type: "thinking", content: currentThinking });
    }

    // Add tool calls
    for (const tc of toolCalls.values()) {
      result.push(tc);
    }

    return result;
  }, [chunks]);

  if (grouped.length === 0 && !isStreaming) {
    return null;
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
      {grouped.map((item, index) => {
        switch (item.type) {
          case "text":
            return <TextChunk key={index} content={item.content || ""} />;
          case "thinking":
            return <ThinkingBlock key={index} content={item.content || ""} />;
          case "tool":
            return (
              <ToolCallCard
                key={item.callId || index}
                tool={item.tool || "unknown"}
                callId={item.callId || ""}
                status={item.status || "running"}
                result={item.result}
              />
            );
          default:
            return null;
        }
      })}

      {isStreaming && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "6px",
            fontSize: "12px",
            color: "var(--text-secondary)",
          }}
        >
          <span
            style={{
              width: "8px",
              height: "8px",
              borderRadius: "50%",
              background: "#3b82f6",
              animation: "pulse 1s infinite",
            }}
          />
          正在生成...
        </div>
      )}

      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
      `}</style>
    </div>
  );
}
