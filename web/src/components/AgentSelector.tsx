import React, { useState, useRef, useEffect, useCallback } from "react";
import { AgentIcon } from "./AgentIcon";

type AgentStatus = {
  name: string;
  available: boolean;
  version?: string;
};

type AgentSelectorProps = {
  agent: string;
  agents: AgentStatus[];
  onAgentChange: (agent: string) => void;
  compact?: boolean;
};

export function AgentSelector({
  agent,
  agents,
  onAgentChange,
  compact = false,
}: AgentSelectorProps) {
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

  const handleAgentSelect = useCallback(
    (newAgent: string) => {
      onAgentChange(newAgent);
      setIsOpen(false);
    },
    [onAgentChange]
  );

  return (
    <div ref={dropdownRef} style={{ position: "relative" }}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        style={{
          display: "flex",
          alignItems: "center",
          gap: "4px",
          padding: "6px 8px",
          borderRadius: "12px",
          border: "none",
          background: "transparent",
          cursor: "pointer",
          fontSize: "16px",
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
        <AgentIcon agentName={agent} style={{ width: "16px", height: "16px" }} />
      </button>

      {isOpen && (
        <div
          style={{
            position: "absolute",
            bottom: "calc(100% + 10px)",
            right: 0,
            background: "#fff",
            border: "1px solid var(--border-color)",
            borderRadius: "12px",
            boxShadow: "0 8px 32px rgba(0,0,0,0.15)",
            zIndex: 1000,
            width: "200px",
            padding: "8px 0",
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
              <AgentIcon agentName={a.name} style={{ width: "16px", height: "16px", marginRight: "4px" }} />
              <span>{a.name}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
