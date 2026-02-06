import React from "react";
import type { SessionInfo } from "./AgentFloatingPanel";

type AgentBubbleProps = {
  session: SessionInfo | null;
  isStreaming?: boolean;
  onClick?: () => void;
};

const statusColors: Record<string, string> = {
  active: "#22c55e",
  idle: "#f59e0b",
  closed: "#6b7280",
};

export function AgentBubble({ session, isStreaming, onClick }: AgentBubbleProps) {
  // 无活跃 Session 时不显示
  if (!session || session.status === "closed") return null;

  const statusColor = statusColors[session.status] || statusColors.idle;

  return (
    <button
      onClick={onClick}
      style={{
        position: "absolute",
        bottom: "24px",
        right: "24px",
        display: "flex",
        alignItems: "center",
        gap: "10px",
        padding: "12px 16px",
        background: "#fff",
        border: "1px solid var(--border-color)",
        borderRadius: "24px",
        boxShadow: "0 4px 16px rgba(0,0,0,0.12)",
        cursor: "pointer",
        zIndex: 50,
        transition: "transform 0.2s, box-shadow 0.2s",
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.transform = "scale(1.02)";
        e.currentTarget.style.boxShadow = "0 6px 20px rgba(0,0,0,0.15)";
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.transform = "scale(1)";
        e.currentTarget.style.boxShadow = "0 4px 16px rgba(0,0,0,0.12)";
      }}
    >
      {/* 类型图标 */}
      <span style={{ fontSize: "18px" }}>
        {session.type === "chat" ? "💬" : session.type === "view" ? "🎨" : "⚡"}
      </span>

      {/* Session 信息 */}
      <div style={{ textAlign: "left" }}>
        <div
          style={{
            fontSize: "13px",
            fontWeight: 600,
            color: "var(--text-primary)",
            maxWidth: "160px",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
        >
          {session.name || `Session ${session.key.slice(0, 8)}`}
        </div>
        <div
          style={{
            fontSize: "11px",
            color: "var(--text-secondary)",
            display: "flex",
            alignItems: "center",
            gap: "6px",
          }}
        >
          <span
            style={{
              display: "inline-block",
              width: "6px",
              height: "6px",
              borderRadius: "50%",
              background: statusColor,
            }}
          />
          {session.agent}
          {isStreaming && (
            <span style={{ color: "#3b82f6" }}>
              <StreamingDots />
            </span>
          )}
        </div>
      </div>

      {/* 展开箭头 */}
      <span
        style={{
          fontSize: "14px",
          color: "var(--text-secondary)",
          marginLeft: "4px",
        }}
      >
        ↗
      </span>
    </button>
  );
}

// 流式输出动画点
function StreamingDots() {
  return (
    <span
      style={{
        display: "inline-flex",
        gap: "2px",
        marginLeft: "4px",
      }}
    >
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          style={{
            width: "4px",
            height: "4px",
            borderRadius: "50%",
            background: "#3b82f6",
            animation: `bubbleDot 1.2s ease-in-out ${i * 0.2}s infinite`,
          }}
        />
      ))}
      <style>
        {`
          @keyframes bubbleDot {
            0%, 80%, 100% { opacity: 0.3; transform: scale(0.8); }
            40% { opacity: 1; transform: scale(1); }
          }
        `}
      </style>
    </span>
  );
}
