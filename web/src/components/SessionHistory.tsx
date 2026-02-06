import React from "react";
import type { SessionInfo } from "./AgentFloatingPanel";
import { StreamMessage, type StreamChunkData } from "./stream";

type Exchange = {
  role: "user" | "agent";
  content: string;
  timestamp?: string;
};

type SessionSummary = {
  title: string;
  description: string;
  keyActions: string[];
  outputs: string[];
  generatedAt: string;
};

type SessionHistoryProps = {
  session: SessionInfo | null;
  summary?: SessionSummary | null;
  exchanges?: Exchange[];
  relatedFiles?: { path: string; name: string }[];
  onRestore?: () => void;
  onFileClick?: (path: string) => void;
  onClose?: () => void;
};

const typeLabels: Record<string, string> = {
  chat: "对话",
  view: "视图生成",
  skill: "技能执行",
};

export function SessionHistory({
  session,
  summary,
  exchanges = [],
  relatedFiles = [],
  onRestore,
  onFileClick,
  onClose,
}: SessionHistoryProps) {
  if (!session) return null;

  return (
    <div
      style={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        background: "#fff",
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: "16px 20px",
          borderBottom: "1px solid var(--border-color)",
          display: "flex",
          alignItems: "center",
          gap: "12px",
          background: "rgba(0,0,0,0.02)",
        }}
      >
        <span style={{ fontSize: "20px" }}>
          {session.type === "chat" ? "💬" : session.type === "view" ? "🎨" : "⚡"}
        </span>
        <div style={{ flex: 1 }}>
          <div style={{ fontSize: "16px", fontWeight: 600 }}>
            {session.name || `Session ${session.key.slice(0, 8)}`}
          </div>
          <div style={{ fontSize: "12px", color: "var(--text-secondary)" }}>
            {typeLabels[session.type]} · {session.agent} · 已关闭
          </div>
        </div>
        <button
          onClick={onRestore}
          style={{
            padding: "8px 16px",
            borderRadius: "8px",
            border: "1px solid #3b82f6",
            background: "#fff",
            color: "#3b82f6",
            fontSize: "13px",
            fontWeight: 500,
            cursor: "pointer",
            display: "flex",
            alignItems: "center",
            gap: "6px",
          }}
        >
          ↻ 恢复
        </button>
        {onClose && (
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
        )}
      </div>

      {/* Content */}
      <div
        style={{
          flex: 1,
          overflow: "auto",
          padding: "20px",
        }}
      >
        {/* Summary */}
        {summary && (
          <div
            style={{
              marginBottom: "24px",
              padding: "16px",
              background: "rgba(59, 130, 246, 0.05)",
              borderRadius: "12px",
              border: "1px solid rgba(59, 130, 246, 0.1)",
            }}
          >
            <div
              style={{
                fontSize: "14px",
                fontWeight: 600,
                marginBottom: "8px",
                color: "#3b82f6",
              }}
            >
              摘要
            </div>
            <div
              style={{
                fontSize: "15px",
                fontWeight: 500,
                marginBottom: "8px",
              }}
            >
              {summary.title}
            </div>
            <div
              style={{
                fontSize: "13px",
                color: "var(--text-secondary)",
                lineHeight: 1.6,
                marginBottom: "12px",
              }}
            >
              {summary.description}
            </div>

            {summary.keyActions.length > 0 && (
              <div style={{ marginBottom: "12px" }}>
                <div
                  style={{
                    fontSize: "12px",
                    fontWeight: 500,
                    color: "var(--text-secondary)",
                    marginBottom: "6px",
                  }}
                >
                  关键操作
                </div>
                <ul
                  style={{
                    margin: 0,
                    paddingLeft: "20px",
                    fontSize: "13px",
                    color: "var(--text-primary)",
                  }}
                >
                  {summary.keyActions.map((action, i) => (
                    <li key={i} style={{ marginBottom: "4px" }}>
                      {action}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {summary.outputs.length > 0 && (
              <div>
                <div
                  style={{
                    fontSize: "12px",
                    fontWeight: 500,
                    color: "var(--text-secondary)",
                    marginBottom: "6px",
                  }}
                >
                  输出文件
                </div>
                <div style={{ display: "flex", flexWrap: "wrap", gap: "6px" }}>
                  {summary.outputs.map((output, i) => (
                    <span
                      key={i}
                      style={{
                        padding: "4px 8px",
                        background: "rgba(0,0,0,0.05)",
                        borderRadius: "4px",
                        fontSize: "12px",
                        fontFamily: "monospace",
                      }}
                    >
                      {output}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {/* 对话历史 */}
        <div style={{ marginBottom: "24px" }}>
          <div
            style={{
              fontSize: "14px",
              fontWeight: 600,
              marginBottom: "16px",
              color: "var(--text-primary)",
            }}
          >
            对话历史
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
            {exchanges.map((ex, i) => (
              <div
                key={i}
                style={{
                  display: "flex",
                  flexDirection: "column",
                  gap: "4px",
                }}
              >
                <div
                  style={{
                    fontSize: "11px",
                    color: "var(--text-secondary)",
                    fontWeight: 500,
                  }}
                >
                  {ex.role === "user" ? "用户" : "Agent"}
                  {ex.timestamp && ` · ${ex.timestamp}`}
                </div>
                <div
                  style={{
                    padding: "12px 16px",
                    borderRadius: ex.role === "user" ? "12px 12px 12px 4px" : "12px 12px 4px 12px",
                    background: ex.role === "user" ? "#3b82f6" : "rgba(0,0,0,0.05)",
                    color: ex.role === "user" ? "#fff" : "var(--text-primary)",
                    fontSize: "13px",
                    lineHeight: 1.6,
                    whiteSpace: "pre-wrap",
                    maxWidth: "85%",
                    alignSelf: ex.role === "user" ? "flex-start" : "flex-start",
                  }}
                >
                  {ex.content}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* 关联文件 */}
        {relatedFiles.length > 0 && (
          <div>
            <div
              style={{
                fontSize: "14px",
                fontWeight: 600,
                marginBottom: "12px",
                color: "var(--text-primary)",
              }}
            >
              关联文件
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
              {relatedFiles.map((file, i) => (
                <button
                  key={i}
                  onClick={() => onFileClick?.(file.path)}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "8px",
                    padding: "10px 12px",
                    background: "rgba(0,0,0,0.02)",
                    border: "1px solid var(--border-color)",
                    borderRadius: "8px",
                    cursor: "pointer",
                    textAlign: "left",
                    transition: "background 0.15s",
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = "rgba(0,0,0,0.05)";
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = "rgba(0,0,0,0.02)";
                  }}
                >
                  <span style={{ fontSize: "14px" }}>📄</span>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div
                      style={{
                        fontSize: "13px",
                        fontWeight: 500,
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {file.name}
                    </div>
                    <div
                      style={{
                        fontSize: "11px",
                        color: "var(--text-secondary)",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {file.path}
                    </div>
                  </div>
                  <span style={{ fontSize: "12px", color: "var(--text-secondary)" }}>→</span>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
