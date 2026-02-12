import React, { useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

type RelatedFile = {
  path: string;
  name: string;
  created_at?: string;
};

type SessionItem = {
  key?: string;
  session_key?: string;
  type?: string;
  name?: string;
  agent?: string;
  scope?: string;
  purpose?: string;
  summary?: string | { title?: string; description?: string };
  exchanges?: Array<{ role?: string; content?: string; timestamp?: string }>;
  closed_at?: string;
  related_files?: Array<RelatedFile | { path: string; relation?: string; created_by_session?: boolean }>;
};

type SessionViewerProps = {
  session: SessionItem | null;
  onFileClick?: (path: string) => void;
};

export function SessionViewer({ session, onFileClick }: SessionViewerProps) {
  const [showAllFiles, setShowAllFiles] = useState(false);

  if (!session) {
    return (
      <div style={{ padding: "40px", textAlign: "center", color: "var(--text-secondary)" }}>
        选择一个会话查看内容
      </div>
    );
  }

  const relatedFiles = (session.related_files || []).map((f: any) => {
    const path = typeof f?.path === "string" ? f.path : "";
    const name = typeof f?.name === "string" ? f.name : path.split("/").pop() || path;
    return { path, name };
  });
  const displayFiles = showAllFiles ? relatedFiles : relatedFiles.slice(0, 5);
  const hasMoreFiles = relatedFiles.length > 5;
  const exchanges = Array.isArray(session.exchanges) ? session.exchanges : [];
  
  const summaryText =
    typeof session.summary === "string"
      ? session.summary
      : session.summary?.description || "";
  
  const scope = session.type ?? session.scope ?? "chat";
  const displayName = session.name ?? session.purpose ?? session.key ?? session.session_key ?? "Session";

  return (
    <div style={{ padding: "24px 32px", maxWidth: "1000px", width: "100%", margin: "0 auto", display: "flex", flexDirection: "column", gap: "24px", boxSizing: "border-box" }}>
      {/* 打薄后的 Header */}
      <div style={{ borderBottom: "1px solid var(--border-color)", paddingBottom: "12px", display: "flex", alignItems: "baseline", gap: "12px" }}>
        <h1 style={{ fontSize: "18px", fontWeight: 600, margin: 0 }}>{displayName}</h1>
        <div style={{ fontSize: "12px", color: "var(--text-secondary)", display: "flex", gap: "8px" }}>
          <strong>{session.agent ?? "agent"}</strong>
          <span>•</span>
          <span>{scope}</span>
        </div>
      </div>

      {/* Content - 确保铺满 */}
      <div style={{ display: "flex", flexDirection: "column", gap: "24px", width: "100%" }}>
        {exchanges.length > 0 ? (
          exchanges.map((item, idx) => {
            const isUser = (item.role || "").toLowerCase() === "user";
            return (
              <div
                key={idx}
                style={{
                  alignSelf: isUser ? "flex-end" : "flex-start",
                  width: isUser ? "auto" : "100%", // Agent 或 结果 必须铺满
                  maxWidth: isUser ? "80%" : "100%",
                }}
              >
                {isUser ? (
                  <div
                    style={{
                      padding: "10px 16px",
                      borderRadius: "18px 18px 4px 18px",
                      background: "var(--accent-color)",
                      color: "#fff",
                      fontSize: "14px",
                      lineHeight: "1.5",
                      boxShadow: "0 4px 12px rgba(59,130,246,0.15)",
                    }}
                  >
                    {item.content || ""}
                  </div>
                ) : (
                  <div style={{ color: "var(--text-primary)", fontSize: "15px", lineHeight: "1.7", width: "100%" }}>
                    <ReactMarkdown 
                      remarkPlugins={[remarkGfm]}
                      components={{
                        p: ({node, ...props}) => <p style={{ margin: "0 0 1em 0", width: "100%" }} {...props} />,
                        pre: ({node, ...props}) => (
                          <pre style={{ 
                            background: "rgba(0,0,0,0.04)", 
                            padding: "16px", 
                            borderRadius: "8px", 
                            overflow: "auto",
                            fontSize: "13px",
                            margin: "1.5em 0",
                            width: "100%",
                            boxSizing: "border-box"
                          }} {...props} />
                        )
                      }}
                    >
                      {item.content || ""}
                    </ReactMarkdown>
                  </div>
                )}
              </div>
            );
          })
        ) : summaryText ? (
          <div style={{ 
            color: "var(--text-primary)", 
            lineHeight: "1.7", 
            fontSize: "15px",
            width: "100%",
            background: "rgba(0,0,0,0.02)",
            padding: "20px",
            borderRadius: "12px"
          }}>
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{summaryText}</ReactMarkdown>
          </div>
        ) : (
          <div style={{ padding: "40px", textAlign: "center", color: "var(--text-secondary)", fontStyle: "italic" }}>
            暂无对话内容记录
          </div>
        )}
      </div>

      {/* 关联文件列表 */}
      {relatedFiles.length > 0 && (
        <div style={{ marginTop: "12px", padding: "20px", background: "rgba(0,0,0,0.02)", borderRadius: "12px", width: "100%", boxSizing: "border-box" }}>
          <div
            style={{
              fontSize: "14px",
              fontWeight: 600,
              color: "var(--text-primary)",
              marginBottom: "16px",
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
                  color: "var(--accent-color)",
                  fontSize: "12px",
                }}
              >
                {showAllFiles ? "收起" : "查看全部"}
              </button>
            )}
          </div>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))", gap: "12px" }}>
            {displayFiles.map((file, i) => (
              <div
                key={i}
                onClick={() => onFileClick?.(file.path)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "10px",
                  padding: "10px",
                  background: "#fff",
                  border: "1px solid var(--border-color)",
                  borderRadius: "8px",
                  cursor: "pointer",
                  transition: "all 0.2s",
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = "var(--accent-color)";
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = "var(--border-color)";
                }}
              >
                <span style={{ fontSize: "16px" }}>📄</span>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div
                    style={{
                      fontSize: "12px",
                      fontWeight: 500,
                      color: "var(--text-primary)",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {file.name}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
