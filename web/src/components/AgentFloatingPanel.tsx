import React, { useCallback, useRef, useState } from "react";
import { StreamMessage, type StreamChunkData } from "./stream";
import { useSessionStream } from "../hooks/useSessionStream";
import { sessionService } from "../services/session";
import { PermissionDialog } from "./dialog/PermissionDialog";

export type SessionInfo = {
  key: string;
  type: "chat" | "view" | "skill";
  agent: string;
  name: string;
  status: "active" | "idle" | "closed";
};

type Exchange = {
  role: "user" | "agent";
  content: string;
};

type AgentFloatingPanelProps = {
  session: SessionInfo | null;
  rootId: string | null;
  exchanges?: Exchange[];
  onClose?: () => void;
  onFileClick?: (path: string) => void;
};

const typeLabels: Record<string, string> = {
  chat: "对话",
  view: "视图生成",
  skill: "技能执行",
};

const statusLabels: Record<string, string> = {
  active: "活跃",
  idle: "空闲",
  closed: "已关闭",
};

export function AgentFloatingPanel({
  session,
  rootId,
  exchanges = [],
  onClose,
  onFileClick,
}: AgentFloatingPanelProps) {
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { chunks, isStreaming, permissionRequest, respondToPermission, clearChunks } =
    useSessionStream(session?.key ?? null);

  const handleSend = useCallback(async () => {
    if (!input.trim() || !session || !rootId || sending) return;

    const message = input.trim();
    setInput("");
    setSending(true);
    clearChunks();

    try {
      await sessionService.sendMessage(rootId, session.key, message);
    } catch (err) {
      console.error("Failed to send message:", err);
    } finally {
      setSending(false);
    }
  }, [input, session, rootId, sending, clearChunks]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend]
  );

  if (!session) return null;

  // Convert chunks to StreamChunkData format
  const streamChunks: StreamChunkData[] = chunks.map((c) => ({
    type: c.type as StreamChunkData["type"],
    content: c.content,
    tool: c.tool,
    error: c.error,
    percent: c.percent,
  }));

  return (
    <>
      <div
        style={{
          position: "absolute",
          top: "10%",
          left: "10%",
          right: "10%",
          bottom: "10%",
          background: "#fff",
          borderRadius: "16px",
          boxShadow: "0 8px 32px rgba(0,0,0,0.15)",
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
          zIndex: 100,
        }}
      >
        {/* Header */}
        <div
          style={{
            padding: "12px 16px",
            borderBottom: "1px solid var(--border-color)",
            display: "flex",
            alignItems: "center",
            gap: "12px",
            background: "rgba(0,0,0,0.02)",
          }}
        >
          <span style={{ fontSize: "16px" }}>
            {session.type === "chat" ? "💬" : session.type === "view" ? "🎨" : "⚡"}
          </span>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: "14px", fontWeight: 600 }}>
              {session.name || `Session ${session.key.slice(0, 8)}`}
            </div>
            <div style={{ fontSize: "11px", color: "var(--text-secondary)" }}>
              {typeLabels[session.type]} · {session.agent} · {statusLabels[session.status]}
            </div>
          </div>
          <button
            onClick={onClose}
            style={{
              background: "none",
              border: "none",
              cursor: "pointer",
              fontSize: "18px",
              color: "var(--text-secondary)",
              padding: "4px 8px",
            }}
          >
            ✕
          </button>
        </div>

        {/* Messages */}
        <div
          style={{
            flex: 1,
            overflow: "auto",
            padding: "16px",
            display: "flex",
            flexDirection: "column",
            gap: "12px",
          }}
        >
          {/* Historical exchanges */}
          {exchanges.map((ex, i) => (
            <div
              key={i}
              style={{
                alignSelf: ex.role === "user" ? "flex-end" : "flex-start",
                maxWidth: "80%",
              }}
            >
              <div
                style={{
                  padding: "10px 14px",
                  borderRadius: ex.role === "user" ? "16px 16px 4px 16px" : "16px 16px 16px 4px",
                  background: ex.role === "user" ? "#3b82f6" : "rgba(0,0,0,0.05)",
                  color: ex.role === "user" ? "#fff" : "var(--text-primary)",
                  fontSize: "13px",
                  lineHeight: 1.5,
                  whiteSpace: "pre-wrap",
                }}
              >
                {ex.content}
              </div>
            </div>
          ))}

          {/* Streaming content */}
          {(streamChunks.length > 0 || isStreaming) && (
            <div style={{ alignSelf: "flex-start", maxWidth: "90%" }}>
              <StreamMessage chunks={streamChunks} isStreaming={isStreaming} />
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        {session.status !== "closed" && (
          <div
            style={{
              padding: "12px 16px",
              borderTop: "1px solid var(--border-color)",
              display: "flex",
              gap: "8px",
            }}
          >
            <textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="输入消息..."
              disabled={sending || isStreaming}
              style={{
                flex: 1,
                padding: "10px 12px",
                borderRadius: "8px",
                border: "1px solid var(--border-color)",
                fontSize: "13px",
                resize: "none",
                minHeight: "40px",
                maxHeight: "120px",
                fontFamily: "inherit",
              }}
            />
            <button
              onClick={handleSend}
              disabled={!input.trim() || sending || isStreaming}
              style={{
                padding: "10px 20px",
                borderRadius: "8px",
                border: "none",
                background: "#3b82f6",
                color: "#fff",
                fontSize: "13px",
                cursor: !input.trim() || sending || isStreaming ? "not-allowed" : "pointer",
                opacity: !input.trim() || sending || isStreaming ? 0.5 : 1,
              }}
            >
              {sending || isStreaming ? "..." : "发送"}
            </button>
          </div>
        )}
      </div>

      {/* Permission Dialog */}
      <PermissionDialog
        request={permissionRequest}
        onRespond={respondToPermission}
      />
    </>
  );
}
