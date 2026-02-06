import React, { useEffect, useState } from "react";

export type PermissionRequest = {
  requestId: string;
  permission: string;
  resource?: string;
  action?: string;
};

type PermissionDialogProps = {
  request: PermissionRequest | null;
  onRespond: (requestId: string, granted: boolean, always?: boolean) => void;
  timeout?: number; // seconds
};

const permissionIcons: Record<string, string> = {
  file_write: "📝",
  file_read: "📖",
  command_exec: "⌨️",
  network: "🌐",
  default: "🔐",
};

const permissionLabels: Record<string, string> = {
  file_write: "文件写入",
  file_read: "文件读取",
  command_exec: "命令执行",
  network: "网络访问",
};

export function PermissionDialog({
  request,
  onRespond,
  timeout = 30,
}: PermissionDialogProps) {
  const [remaining, setRemaining] = useState(timeout);

  useEffect(() => {
    if (!request) return;

    setRemaining(timeout);
    const timer = setInterval(() => {
      setRemaining((prev) => {
        if (prev <= 1) {
          // Auto-deny on timeout
          onRespond(request.requestId, false);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [request, timeout, onRespond]);

  if (!request) return null;

  const icon = permissionIcons[request.permission] || permissionIcons.default;
  const label = permissionLabels[request.permission] || request.permission;

  return (
    <div
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: "rgba(0,0,0,0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
      }}
    >
      <div
        style={{
          background: "#fff",
          borderRadius: "12px",
          padding: "20px",
          maxWidth: "400px",
          width: "90%",
          boxShadow: "0 4px 20px rgba(0,0,0,0.15)",
        }}
      >
        {/* Header */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "10px",
            marginBottom: "16px",
          }}
        >
          <span style={{ fontSize: "24px" }}>{icon}</span>
          <div>
            <div style={{ fontSize: "16px", fontWeight: 600 }}>权限请求</div>
            <div style={{ fontSize: "12px", color: "var(--text-secondary)" }}>
              {label}
            </div>
          </div>
        </div>

        {/* Description */}
        <div
          style={{
            padding: "12px",
            borderRadius: "8px",
            background: "rgba(0,0,0,0.03)",
            marginBottom: "16px",
          }}
        >
          {request.action && (
            <div style={{ fontSize: "13px", marginBottom: "8px" }}>
              <strong>操作:</strong> {request.action}
            </div>
          )}
          {request.resource && (
            <div
              style={{
                fontSize: "12px",
                fontFamily: "monospace",
                color: "var(--text-secondary)",
                wordBreak: "break-all",
              }}
            >
              {request.resource}
            </div>
          )}
        </div>

        {/* Timeout */}
        <div
          style={{
            fontSize: "11px",
            color: remaining <= 10 ? "#ef4444" : "var(--text-secondary)",
            marginBottom: "16px",
            textAlign: "center",
          }}
        >
          {remaining} 秒后自动拒绝
        </div>

        {/* Actions */}
        <div
          style={{
            display: "flex",
            gap: "8px",
            justifyContent: "flex-end",
          }}
        >
          <button
            onClick={() => onRespond(request.requestId, false)}
            style={{
              padding: "8px 16px",
              borderRadius: "6px",
              border: "1px solid var(--border-color)",
              background: "#fff",
              fontSize: "13px",
              cursor: "pointer",
            }}
          >
            拒绝
          </button>
          <button
            onClick={() => onRespond(request.requestId, true)}
            style={{
              padding: "8px 16px",
              borderRadius: "6px",
              border: "none",
              background: "#3b82f6",
              color: "#fff",
              fontSize: "13px",
              cursor: "pointer",
            }}
          >
            允许一次
          </button>
          <button
            onClick={() => onRespond(request.requestId, true, true)}
            style={{
              padding: "8px 16px",
              borderRadius: "6px",
              border: "none",
              background: "#22c55e",
              color: "#fff",
              fontSize: "13px",
              cursor: "pointer",
            }}
          >
            始终允许
          </button>
        </div>
      </div>
    </div>
  );
}
