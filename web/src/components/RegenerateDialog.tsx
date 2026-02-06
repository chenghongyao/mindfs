import React, { useState, useCallback } from "react";

type RegenerateDialogProps = {
  isOpen: boolean;
  routeName: string;
  lastPrompt?: string;
  versions?: string[];
  onClose: () => void;
  onGenerate: (prompt: string, baseVersion?: string) => void;
};

export function RegenerateDialog({
  isOpen,
  routeName,
  lastPrompt = "",
  versions = [],
  onClose,
  onGenerate,
}: RegenerateDialogProps): JSX.Element | null {
  const [prompt, setPrompt] = useState(lastPrompt);
  const [baseVersion, setBaseVersion] = useState<string | null>(null);
  const [generating, setGenerating] = useState(false);

  const handleGenerate = useCallback(async () => {
    if (!prompt.trim()) return;

    setGenerating(true);
    try {
      await onGenerate(prompt.trim(), baseVersion || undefined);
      onClose();
    } catch (err) {
      console.error("Failed to generate:", err);
    } finally {
      setGenerating(false);
    }
  }, [prompt, baseVersion, onGenerate, onClose]);

  if (!isOpen) return null;

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        style={{
          background: "#fff",
          borderRadius: "16px",
          width: "100%",
          maxWidth: "480px",
          boxShadow: "0 8px 32px rgba(0,0,0,0.2)",
          overflow: "hidden",
        }}
      >
        {/* Header */}
        <div
          style={{
            padding: "16px 20px",
            borderBottom: "1px solid var(--border-color)",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <div>
            <div style={{ fontSize: "16px", fontWeight: 600 }}>重新生成视图</div>
            <div style={{ fontSize: "12px", color: "var(--text-secondary)" }}>{routeName}</div>
          </div>
          <button
            type="button"
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

        {/* Content */}
        <div style={{ padding: "20px" }}>
          {/* Prompt input */}
          <div style={{ marginBottom: "16px" }}>
            <label
              style={{
                display: "block",
                fontSize: "13px",
                fontWeight: 500,
                marginBottom: "8px",
                color: "var(--text-primary)",
              }}
            >
              描述你想要的视图
            </label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="例如：创建一个小说阅读器视图，支持章节导航..."
              style={{
                width: "100%",
                minHeight: "100px",
                padding: "12px",
                borderRadius: "8px",
                border: "1px solid var(--border-color)",
                fontSize: "13px",
                resize: "vertical",
                fontFamily: "inherit",
              }}
            />
          </div>

          {/* Base version selection */}
          {versions.length > 0 && (
            <div style={{ marginBottom: "16px" }}>
              <label
                style={{
                  display: "block",
                  fontSize: "13px",
                  fontWeight: 500,
                  marginBottom: "8px",
                  color: "var(--text-primary)",
                }}
              >
                基于版本
              </label>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "8px" }}>
                <button
                  type="button"
                  onClick={() => setBaseVersion(null)}
                  style={{
                    padding: "8px 12px",
                    borderRadius: "6px",
                    border: "1px solid var(--border-color)",
                    background: baseVersion === null ? "#3b82f6" : "#fff",
                    color: baseVersion === null ? "#fff" : "var(--text-primary)",
                    fontSize: "12px",
                    cursor: "pointer",
                  }}
                >
                  全新生成
                </button>
                {versions.map((v) => (
                  <button
                    key={v}
                    type="button"
                    onClick={() => setBaseVersion(v)}
                    style={{
                      padding: "8px 12px",
                      borderRadius: "6px",
                      border: "1px solid var(--border-color)",
                      background: baseVersion === v ? "#3b82f6" : "#fff",
                      color: baseVersion === v ? "#fff" : "var(--text-primary)",
                      fontSize: "12px",
                      cursor: "pointer",
                    }}
                  >
                    基于 {v}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Last prompt hint */}
          {lastPrompt && lastPrompt !== prompt && (
            <div
              style={{
                padding: "12px",
                background: "rgba(0,0,0,0.02)",
                borderRadius: "8px",
                marginBottom: "16px",
              }}
            >
              <div
                style={{
                  fontSize: "11px",
                  fontWeight: 500,
                  color: "var(--text-secondary)",
                  marginBottom: "4px",
                }}
              >
                上次提示词
              </div>
              <div
                style={{
                  fontSize: "12px",
                  color: "var(--text-primary)",
                  lineHeight: 1.5,
                }}
              >
                {lastPrompt}
              </div>
              <button
                type="button"
                onClick={() => setPrompt(lastPrompt)}
                style={{
                  marginTop: "8px",
                  padding: "4px 8px",
                  borderRadius: "4px",
                  border: "1px solid var(--border-color)",
                  background: "#fff",
                  fontSize: "11px",
                  cursor: "pointer",
                  color: "#3b82f6",
                }}
              >
                使用此提示词
              </button>
            </div>
          )}
        </div>

        {/* Footer */}
        <div
          style={{
            padding: "16px 20px",
            borderTop: "1px solid var(--border-color)",
            display: "flex",
            justifyContent: "flex-end",
            gap: "12px",
          }}
        >
          <button
            type="button"
            onClick={onClose}
            style={{
              padding: "10px 20px",
              borderRadius: "8px",
              border: "1px solid var(--border-color)",
              background: "#fff",
              fontSize: "13px",
              cursor: "pointer",
            }}
          >
            取消
          </button>
          <button
            type="button"
            onClick={handleGenerate}
            disabled={!prompt.trim() || generating}
            style={{
              padding: "10px 20px",
              borderRadius: "8px",
              border: "none",
              background: "#3b82f6",
              color: "#fff",
              fontSize: "13px",
              fontWeight: 500,
              cursor: !prompt.trim() || generating ? "not-allowed" : "pointer",
              opacity: !prompt.trim() || generating ? 0.6 : 1,
            }}
          >
            {generating ? "生成中..." : "开始生成"}
          </button>
        </div>
      </div>
    </div>
  );
}
