import React from "react";

type RightSidebarProps = {
  collapsed?: boolean;
  onToggle?: () => void;
  onOpenSettings?: () => void;
  children?: React.ReactNode;
};

export function RightSidebar({
  collapsed = false,
  onToggle,
  onOpenSettings,
  children,
}: RightSidebarProps) {
  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "12px 16px",
          borderBottom: "1px solid var(--border-color)",
          background: "rgba(255,255,255,0.7)",
          position: "sticky",
          top: 0,
          zIndex: 2,
          backdropFilter: "blur(8px)",
        }}
      >
        {collapsed ? (
          <button
            type="button"
            onClick={onToggle}
            style={{
              padding: "4px 6px",
              borderRadius: "6px",
              border: "1px solid var(--border-color)",
              background: "#fff",
              fontSize: "12px",
              cursor: "pointer",
            }}
          >
            展开
          </button>
        ) : (
          <>
            <div style={{ fontSize: "12px", fontWeight: 600, color: "var(--text-secondary)" }}>
              会话
            </div>
            <div style={{ display: "flex", gap: "8px" }}>
              <button
                type="button"
                onClick={onOpenSettings}
                title="目录设置"
                style={{
                  padding: "4px 8px",
                  borderRadius: "6px",
                  border: "1px solid var(--border-color)",
                  background: "#fff",
                  fontSize: "14px",
                  cursor: "pointer",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                ⚙️
              </button>
              <button
                type="button"
                onClick={onToggle}
                style={{
                  padding: "4px 8px",
                  borderRadius: "6px",
                  border: "1px solid var(--border-color)",
                  background: "#fff",
                  fontSize: "12px",
                  cursor: "pointer",
                }}
              >
                折叠
              </button>
            </div>
          </>
        )}
      </div>
      {collapsed ? null : (
        <div style={{ flex: 1, overflow: "auto", padding: "12px 12px 16px" }}>{children}</div>
      )}
    </div>
  );
}
