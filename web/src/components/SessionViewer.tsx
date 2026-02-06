import React, { useState } from "react";

type RelatedFile = {
  path: string;
  name: string;
  created_at?: string;
};

type SessionItem = {
  session_key: string;
  agent?: string;
  scope?: string;
  purpose?: string;
  summary?: string;
  closed_at?: string;
  related_files?: RelatedFile[];
};

type SessionViewerProps = {
  session: SessionItem | null;
  onFileClick?: (path: string) => void;
};

export function SessionViewer({ session, onFileClick }: SessionViewerProps) {
  const [showAllFiles, setShowAllFiles] = useState(false);

  if (!session) {
    return (
      <div style={{ padding: "24px", color: "var(--text-secondary)" }}>
        选择一个会话查看内容
      </div>
    );
  }

  const relatedFiles = session.related_files || [];
  const displayFiles = showAllFiles ? relatedFiles : relatedFiles.slice(0, 5);
  const hasMoreFiles = relatedFiles.length > 5;

  return (
    <div style={{ padding: "24px", display: "flex", flexDirection: "column", gap: "16px" }}>
      <div style={{ fontSize: "14px", color: "var(--text-secondary)" }}>
        {session.agent ?? "agent"} · {session.scope ?? "scope"} · {session.purpose ?? "purpose"}
      </div>
      <div style={{ fontSize: "20px", fontWeight: 700, color: "var(--text-primary)" }}>
        会话总结
      </div>
      <div
        style={{
          background: "rgba(255,255,255,0.9)",
          borderRadius: "12px",
          padding: "16px",
          border: "1px solid var(--border-color)",
          lineHeight: 1.6,
        }}
      >
        {session.summary ?? "(no summary)"}
      </div>

      {/* 关联文件列表 */}
      {relatedFiles.length > 0 && (
        <div>
          <div
            style={{
              fontSize: "14px",
              fontWeight: 600,
              color: "var(--text-primary)",
              marginBottom: "12px",
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
            }}
          >
            <span>关联文件 ({relatedFiles.length})</span>
            {hasMoreFiles && (
              <button
                type="button"
                onClick={() => setShowAllFiles(!showAllFiles)}
                style={{
                  background: "none",
                  border: "none",
                  padding: 0,
                  cursor: "pointer",
                  color: "#3b82f6",
                  fontSize: "12px",
                }}
              >
                {showAllFiles ? "收起" : "查看全部"}
              </button>
            )}
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
            {displayFiles.map((file, i) => (
              <button
                key={i}
                type="button"
                onClick={() => onFileClick?.(file.path)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "10px",
                  padding: "10px 12px",
                  background: "rgba(255,255,255,0.9)",
                  border: "1px solid var(--border-color)",
                  borderRadius: "8px",
                  cursor: "pointer",
                  textAlign: "left",
                  transition: "background 0.15s",
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = "rgba(0,0,0,0.02)";
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = "rgba(255,255,255,0.9)";
                }}
              >
                <span style={{ fontSize: "16px" }}>📄</span>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div
                    style={{
                      fontSize: "13px",
                      fontWeight: 500,
                      color: "var(--text-primary)",
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

      {session.closed_at ? (
        <div style={{ fontSize: "12px", color: "var(--text-secondary)" }}>
          结束时间：{new Date(session.closed_at).toLocaleString()}
        </div>
      ) : null}
    </div>
  );
}
