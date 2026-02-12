import React, { useState, useRef, useEffect, useCallback } from "react";

export type SessionMode = "chat" | "view" | "skill";

type AgentStatus = {
  name: string;
  available: boolean;
  version?: string;
};

type ModeAgentSelectorProps = {
  mode: SessionMode;
  agent: string;
  agents: AgentStatus[];
  onModeChange: (mode: SessionMode) => void;
  onAgentChange: (agent: string) => void;
  compact?: boolean;
  showLabel?: boolean;
};

const modeLabels: Record<SessionMode, string> = {
  chat: "对话",
  view: "生成视图",
  skill: "执行技能",
};

const modeIcons: Record<SessionMode, string> = {
  chat: "💬",
  view: "🎨",
  skill: "⚡",
};

export function ModeAgentSelector({
  mode,
  agent,
  agents,
  onModeChange,
  onAgentChange,
  compact = false,
  showLabel = true,
}: ModeAgentSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [isOpen]);

  const handleModeSelect = useCallback(
    (newMode: SessionMode) => {
      onModeChange(newMode);
    },
    [onModeChange]
  );

  const handleAgentSelect = useCallback(
    (newAgent: string) => {
      onAgentChange(newAgent);
      setIsOpen(false);
    },
    [onAgentChange]
  );

  const currentAgent = agents.find((a) => a.name === agent);

  return (
    <div ref={dropdownRef} style={{ position: "relative" }}>
      {/* 触发按钮 */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        style={{
          display: "flex",
          alignItems: "center",
          gap: compact ? "4px" : "6px",
          padding: compact ? "6px 8px" : "8px 12px",
          borderRadius: compact ? "12px" : "8px",
          border: compact ? "none" : "1px solid var(--border-color)",
          background: compact ? "transparent" : "#fff",
          cursor: "pointer",
          fontSize: compact ? "16px" : "13px",
          fontWeight: 500,
          color: "var(--text-primary)",
          minWidth: compact ? "auto" : "140px",
          transition: "background 0.2s",
          outline: "none",
        }}
        onMouseEnter={(e) => {
          if (compact) e.currentTarget.style.background = "rgba(0,0,0,0.05)";
        }}
        onMouseLeave={(e) => {
          if (compact) e.currentTarget.style.background = "transparent";
        }}
      >
        <span>{modeIcons[mode]}</span>
        {showLabel && !compact && <span>{modeLabels[mode]}</span>}
        {showLabel && !compact && <span style={{ color: "var(--text-secondary)", margin: "0 2px" }}>·</span>}
        {showLabel && (
          <span
            style={{
              color: currentAgent?.available ? "var(--text-primary)" : "var(--text-secondary)",
              fontSize: compact ? "12px" : "inherit",
            }}
          >
            {agent || "未选择"}
          </span>
        )}
        {showLabel && (
          <span
            style={{
              marginLeft: compact ? "2px" : "auto",
              fontSize: "10px",
              color: "var(--text-secondary)",
              transform: isOpen ? "rotate(180deg)" : "rotate(0deg)",
              transition: "transform 0.15s",
            }}
          >
            ▼
          </span>
        )}
      </button>

      {/* 下拉面板 */}
      {isOpen && (
        <div
          style={{
            position: "absolute",
            bottom: "calc(100% + 10px)", // 向上弹出，适合底部输入框
            left: compact ? "-10px" : 0,
            background: "#fff",
            border: "1px solid var(--border-color)",
            borderRadius: "12px",
            boxShadow: "0 8px 32px rgba(0,0,0,0.15)",
            zIndex: 1000,
            display: "flex",
            overflow: "hidden",
            minWidth: "320px",
          }}
        >
          {/* 左侧: 模式列表 */}
          <div
            style={{
              borderRight: "1px solid var(--border-color)",
              padding: "8px 0",
              minWidth: "140px",
            }}
          >
            <div
              style={{
                padding: "6px 12px",
                fontSize: "11px",
                fontWeight: 600,
                color: "var(--text-secondary)",
                textTransform: "uppercase",
              }}
            >
              模式
            </div>
            {(["chat", "view", "skill"] as SessionMode[]).map((m) => (
              <button
                key={m}
                type="button"
                onClick={() => handleModeSelect(m)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "8px",
                  width: "100%",
                  padding: "10px 12px",
                  border: "none",
                  background: m === mode ? "rgba(59, 130, 246, 0.08)" : "transparent",
                  cursor: "pointer",
                  fontSize: "13px",
                  color: m === mode ? "#3b82f6" : "var(--text-primary)",
                  fontWeight: m === mode ? 500 : 400,
                  textAlign: "left",
                }}
              >
                <span>{modeIcons[m]}</span>
                <span>{modeLabels[m]}</span>
              </button>
            ))}
          </div>

          {/* 右侧: Agent 列表 */}
          <div style={{ padding: "8px 0", minWidth: "160px" }}>
            <div
              style={{
                padding: "6px 12px",
                fontSize: "11px",
                fontWeight: 600,
                color: "var(--text-secondary)",
                textTransform: "uppercase",
              }}
            >
              Agent
            </div>
            {agents.map((a) => (
              <button
                key={a.name}
                type="button"
                onClick={() => a.available && handleAgentSelect(a.name)}
                disabled={!a.available}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "8px",
                  width: "100%",
                  padding: "10px 12px",
                  border: "none",
                  background: a.name === agent ? "rgba(59, 130, 246, 0.08)" : "transparent",
                  cursor: a.available ? "pointer" : "not-allowed",
                  fontSize: "13px",
                  color: !a.available
                    ? "var(--text-secondary)"
                    : a.name === agent
                    ? "#3b82f6"
                    : "var(--text-primary)",
                  fontWeight: a.name === agent ? 500 : 400,
                  textAlign: "left",
                  opacity: a.available ? 1 : 0.6,
                }}
              >
                <span
                  style={{
                    width: "8px",
                    height: "8px",
                    borderRadius: "50%",
                    background: a.available ? "#22c55e" : "#ef4444",
                  }}
                />
                <span>{a.name}</span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
