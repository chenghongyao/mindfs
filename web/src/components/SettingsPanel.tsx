import React, { useCallback, useEffect, useState } from "react";

type DirConfig = {
  defaultAgent: string;
  userDescription: string;
};

type AgentStatus = {
  name: string;
  available: boolean;
  version?: string;
  error?: string;
};

type SettingsPanelProps = {
  open?: boolean;
  rootId?: string | null;
  onClose?: () => void;
};

export function SettingsPanel({ open = false, rootId, onClose }: SettingsPanelProps) {
  const [config, setConfig] = useState<DirConfig>({ defaultAgent: "", userDescription: "" });
  const [agents, setAgents] = useState<AgentStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load config and agents
  useEffect(() => {
    if (!open || !rootId) return;
    let cancelled = false;

    const load = async () => {
      setLoading(true);
      setError(null);
      try {
        const [configRes, agentsRes] = await Promise.all([
          fetch(`/api/dirs/${encodeURIComponent(rootId)}/config`),
          fetch("/api/agents"),
        ]);
        if (cancelled) return;

        if (configRes.ok) {
          const cfg = await configRes.json();
          setConfig({
            defaultAgent: cfg.defaultAgent || "",
            userDescription: cfg.userDescription || "",
          });
        }

        if (agentsRes.ok) {
          const data = await agentsRes.json();
          setAgents(Array.isArray(data) ? data : []);
        }
      } catch (err) {
        if (!cancelled) {
          setError("加载配置失败");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    load();
    return () => {
      cancelled = true;
    };
  }, [open, rootId]);

  const handleSave = useCallback(async () => {
    if (!rootId) return;
    setSaving(true);
    setError(null);
    try {
      const res = await fetch(`/api/dirs/${encodeURIComponent(rootId)}/config`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(config),
      });
      if (!res.ok) {
        throw new Error("保存失败");
      }
      onClose?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存失败");
    } finally {
      setSaving(false);
    }
  }, [rootId, config, onClose]);

  const handleProbeAgent = useCallback(async (name: string) => {
    try {
      const res = await fetch(`/api/agents/${encodeURIComponent(name)}/probe`, {
        method: "POST",
      });
      if (res.ok) {
        const status = await res.json();
        setAgents((prev) =>
          prev.map((a) => (a.name === name ? status : a))
        );
      }
    } catch {
      // ignore
    }
  }, []);

  if (!open) return null;

  return (
    <div
      style={{
        marginTop: "16px",
        padding: "16px",
        borderRadius: "12px",
        border: "1px solid var(--border-color)",
        background: "rgba(255,255,255,0.95)",
        display: "flex",
        flexDirection: "column",
        gap: "16px",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <div style={{ fontSize: "14px", fontWeight: 600 }}>目录配置</div>
        <button
          onClick={onClose}
          style={{
            background: "none",
            border: "none",
            cursor: "pointer",
            fontSize: "16px",
            color: "var(--text-secondary)",
          }}
        >
          ✕
        </button>
      </div>

      {loading ? (
        <div style={{ fontSize: "12px", color: "var(--text-secondary)" }}>
          加载中...
        </div>
      ) : (
        <>
          {/* User Description */}
          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            <label
              style={{
                fontSize: "12px",
                fontWeight: 500,
                color: "var(--text-secondary)",
              }}
            >
              目录描述
            </label>
            <textarea
              value={config.userDescription}
              onChange={(e) =>
                setConfig((prev) => ({ ...prev, userDescription: e.target.value }))
              }
              placeholder="描述这个目录的用途，帮助 Agent 更好地理解上下文..."
              style={{
                padding: "8px 10px",
                borderRadius: "6px",
                border: "1px solid var(--border-color)",
                fontSize: "13px",
                minHeight: "60px",
                resize: "vertical",
                fontFamily: "inherit",
              }}
            />
          </div>

          {/* Default Agent */}
          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            <label
              style={{
                fontSize: "12px",
                fontWeight: 500,
                color: "var(--text-secondary)",
              }}
            >
              默认 Agent
            </label>
            <select
              value={config.defaultAgent}
              onChange={(e) =>
                setConfig((prev) => ({ ...prev, defaultAgent: e.target.value }))
              }
              style={{
                padding: "8px 10px",
                borderRadius: "6px",
                border: "1px solid var(--border-color)",
                fontSize: "13px",
                background: "#fff",
              }}
            >
              <option value="">自动选择</option>
              {agents.map((agent) => (
                <option
                  key={agent.name}
                  value={agent.name}
                  disabled={!agent.available}
                >
                  {agent.name}
                  {agent.available
                    ? agent.version
                      ? ` (${agent.version})`
                      : " ✓"
                    : " ✗"}
                </option>
              ))}
            </select>
          </div>

          {/* Agent Status */}
          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            <label
              style={{
                fontSize: "12px",
                fontWeight: 500,
                color: "var(--text-secondary)",
              }}
            >
              Agent 状态
            </label>
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "4px",
                fontSize: "12px",
              }}
            >
              {agents.length === 0 ? (
                <div style={{ color: "var(--text-secondary)" }}>
                  无可用 Agent
                </div>
              ) : (
                agents.map((agent) => (
                  <div
                    key={agent.name}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "8px",
                      padding: "4px 0",
                    }}
                  >
                    <span
                      style={{
                        color: agent.available ? "#22c55e" : "#ef4444",
                      }}
                    >
                      {agent.available ? "●" : "○"}
                    </span>
                    <span style={{ flex: 1 }}>{agent.name}</span>
                    {agent.version && (
                      <span style={{ color: "var(--text-secondary)" }}>
                        {agent.version}
                      </span>
                    )}
                    {agent.error && (
                      <span
                        style={{ color: "#ef4444", fontSize: "11px" }}
                        title={agent.error}
                      >
                        {agent.error.slice(0, 20)}
                        {agent.error.length > 20 ? "..." : ""}
                      </span>
                    )}
                    <button
                      onClick={() => handleProbeAgent(agent.name)}
                      style={{
                        background: "none",
                        border: "1px solid var(--border-color)",
                        borderRadius: "4px",
                        padding: "2px 6px",
                        fontSize: "11px",
                        cursor: "pointer",
                      }}
                    >
                      刷新
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Error */}
          {error && (
            <div style={{ fontSize: "12px", color: "#ef4444" }}>{error}</div>
          )}

          {/* Actions */}
          <div style={{ display: "flex", gap: "8px", justifyContent: "flex-end" }}>
            <button
              onClick={onClose}
              style={{
                padding: "6px 12px",
                borderRadius: "6px",
                border: "1px solid var(--border-color)",
                background: "#fff",
                fontSize: "13px",
                cursor: "pointer",
              }}
            >
              取消
            </button>
            <button
              onClick={handleSave}
              disabled={saving}
              style={{
                padding: "6px 12px",
                borderRadius: "6px",
                border: "none",
                background: "#3b82f6",
                color: "#fff",
                fontSize: "13px",
                cursor: saving ? "not-allowed" : "pointer",
                opacity: saving ? 0.7 : 1,
              }}
            >
              {saving ? "保存中..." : "保存"}
            </button>
          </div>
        </>
      )}
    </div>
  );
}
