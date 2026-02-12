import React from "react";
import { MarkdownViewer } from "./MarkdownViewer";
import { CodeViewer } from "./CodeViewer";
import { ImageViewer } from "./ImageViewer";
import { BinaryViewer } from "./BinaryViewer";

type FilePayload = {
  name: string;
  path: string;
  content: string;
  encoding: string;
  truncated: boolean;
  size: number;
  ext?: string;
  mime?: string;
  root?: string;
};

type FileMeta = {
  source_session?: string;
  session_name?: string;
  agent?: string;
  created_at?: string;
  created_by?: string;
};

type FileViewerProps = {
  file?: FilePayload | null;
  meta?: FileMeta | null;
  onSessionClick?: (sessionKey: string) => void;
};

// 路径分割组件
function Breadcrumbs({ path }: { path: string }) {
  const parts = path.split('/').filter(Boolean);
  return (
    <div style={{ 
      display: 'flex', 
      alignItems: 'center', 
      gap: '4px', 
      fontSize: '13px', 
      color: 'var(--text-secondary)',
      overflow: 'hidden',
      whiteSpace: 'nowrap',
      flex: 1,
      justifyContent: 'flex-start'
    }}>
      {parts.map((part, index) => (
        <React.Fragment key={index}>
          <span style={{ 
            fontWeight: index === parts.length - 1 ? 600 : 400,
            color: index === parts.length - 1 ? 'var(--text-primary)' : 'inherit',
            flexShrink: index === parts.length - 1 ? 0 : 1,
            overflow: 'hidden',
            textOverflow: 'ellipsis'
          }}>
            {part}
          </span>
          {index < parts.length - 1 && (
            <span style={{ opacity: 0.4, fontSize: '10px', flexShrink: 0 }}>❯</span>
          )}
        </React.Fragment>
      ))}
    </div>
  );
}

export function FileViewer({ file, meta, onSessionClick }: FileViewerProps) {
  if (!file) {
    return (
      <div style={{ height: "100%", display: "flex", alignItems: "center", justifyContent: "center", color: "var(--text-secondary)", flexDirection: "column", gap: "12px" }}>
        <div style={{ fontSize: "48px", opacity: 0.2 }}>📄</div>
        <p>Select a file to preview</p>
      </div>
    );
  }

  const ext = file.ext || (file.path.includes(".") ? `.${file.path.split(".").pop()}` : "");

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%", background: "white" }}>
      <header
        style={{
          padding: "10px 20px",
          borderBottom: "1px solid rgba(0,0,0,0.05)",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          background: "#fff",
          minHeight: "44px",
          boxSizing: "border-box"
        }}
      >
        <div style={{ display: "flex", alignItems: "center", overflow: "hidden", flex: 1 }}>
          <Breadcrumbs path={file.path} />
        </div>
        
        <div style={{ fontSize: "11px", color: "var(--text-secondary)", marginLeft: "12px", flexShrink: 0, opacity: 0.7 }}>
          {(file.size / 1024).toFixed(1)} KB
        </div>
      </header>

      {/* 来源 Session 显示 */}
      {meta?.source_session && (
        <div
          style={{
            padding: "6px 20px",
            background: "rgba(59, 130, 246, 0.03)",
            borderBottom: "1px solid rgba(0,0,0,0.03)",
            display: "flex",
            alignItems: "center",
            gap: "8px",
            fontSize: "12px",
          }}
        >
          <span style={{ color: "var(--text-secondary)", opacity: 0.8 }}>来源:</span>
          <button
            type="button"
            onClick={() => onSessionClick?.(meta.source_session!)}
            style={{
              background: "none",
              border: "none",
              padding: 0,
              cursor: "pointer",
              color: "#3b82f6",
              fontSize: "12px",
              fontWeight: 500,
              display: "flex",
              alignItems: "center",
              gap: "4px",
            }}
          >
            {meta.session_name || `Session ${meta.source_session.slice(0, 8)}`}
          </button>
          {meta.agent && (
            <span style={{ padding: "1px 6px", borderRadius: "4px", background: "rgba(0,0,0,0.04)", color: "var(--text-secondary)", fontSize: "11px" }}>
              {meta.agent}
            </span>
          )}
        </div>
      )}

      <div style={{ flex: 1, overflow: "auto", position: "relative" }}>
        {file.mime?.startsWith("image/") || [".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"].includes(ext.toLowerCase()) ? (
          <ImageViewer path={file.path} root={file.root} />
        ) : file.encoding === "binary" ? (
          <BinaryViewer />
        ) : ext === ".md" || ext === ".markdown" ? (
          <MarkdownViewer content={file.content} />
        ) : (
          <CodeViewer content={file.content} ext={ext} />
        )}
      </div>
      
      {file.truncated && (
        <div style={{ padding: "8px 20px", background: "#fff", borderTop: "1px solid var(--border-color)", fontSize: "11px", color: "#f59e0b", display: "flex", alignItems: "center", gap: "8px" }}>
          ⚠️ 内容已截断
        </div>
      )}
    </div>
  );
}
