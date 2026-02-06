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

export function FileViewer({ file, meta, onSessionClick }: FileViewerProps) {
  if (!file) {
    return (
      <div
        style={{
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: "var(--text-secondary)",
          flexDirection: "column",
          gap: "12px",
        }}
      >
        <div style={{ fontSize: "48px", opacity: 0.2 }}>📄</div>
        <p>Select a file to preview</p>
      </div>
    );
  }

  const ext = file.ext || (file.path.includes(".") ? `.${file.path.split(".").pop()}` : "");

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        background: "white",
      }}
    >
      <header
        style={{
          padding: "12px 24px",
          borderBottom: "1px solid var(--border-color)",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          background: "#fff",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
          <div
            style={{
              width: "32px",
              height: "32px",
              background: "#f1f5f9",
              borderRadius: "6px",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              fontSize: "16px",
            }}
          >
            📝
          </div>
          <div>
            <h2
              style={{
                margin: 0,
                fontSize: "14px",
                fontWeight: 600,
                color: "var(--text-primary)",
              }}
            >
              {file.name}
            </h2>
            <span style={{ fontSize: "12px", color: "var(--text-secondary)" }}>
              {file.path}
            </span>
          </div>
        </div>
        <div style={{ fontSize: "12px", color: "var(--text-secondary)" }}>
          {file.size} bytes
        </div>
      </header>

      {/* 来源 Session 显示 */}
      {meta?.source_session && (
        <div
          style={{
            padding: "8px 24px",
            background: "rgba(59, 130, 246, 0.05)",
            borderBottom: "1px solid var(--border-color)",
            display: "flex",
            alignItems: "center",
            gap: "8px",
            fontSize: "12px",
          }}
        >
          <span style={{ color: "var(--text-secondary)" }}>来源:</span>
          <button
            type="button"
            onClick={() => onSessionClick?.(meta.source_session!)}
            style={{
              background: "none",
              border: "none",
              padding: 0,
              cursor: "pointer",
              color: "#3b82f6",
              fontWeight: 500,
              display: "flex",
              alignItems: "center",
              gap: "4px",
            }}
          >
            <span>💬</span>
            <span>{meta.session_name || `Session ${meta.source_session.slice(0, 8)}`}</span>
          </button>
          {meta.agent && (
            <span
              style={{
                padding: "2px 6px",
                borderRadius: "4px",
                background: "rgba(0,0,0,0.05)",
                color: "var(--text-secondary)",
              }}
            >
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
        <div
          style={{
            padding: "8px 24px",
            background: "#ffffff",
            borderTop: "1px solid var(--border-color)",
            fontSize: "12px",
            color: "#f59e0b",
            display: "flex",
            alignItems: "center",
            gap: "8px",
          }}
        >
          ⚠️ Preview truncated
        </div>
      )}
    </div>
  );
}
