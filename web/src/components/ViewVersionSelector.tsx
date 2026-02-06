import React, { useState, useRef, useEffect, useCallback } from "react";

export type ViewInfo = {
  routeId: string;
  routeName: string;
  priority: number;
  isDefault: boolean;
  versions: string[];
  active?: string;
};

type ViewVersionSelectorProps = {
  views: ViewInfo[];
  selectedView: string | null;
  selectedVersion: string | null;
  onViewChange: (routeId: string) => void;
  onVersionChange: (version: string) => void;
  disabled?: boolean;
};

export function ViewVersionSelector({
  views,
  selectedView,
  selectedVersion,
  onViewChange,
  onVersionChange,
  disabled,
}: ViewVersionSelectorProps): JSX.Element {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // 点击外部关闭
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

  const handleViewSelect = useCallback(
    (routeId: string) => {
      onViewChange(routeId);
    },
    [onViewChange]
  );

  const handleVersionSelect = useCallback(
    (version: string) => {
      onVersionChange(version);
      setIsOpen(false);
    },
    [onVersionChange]
  );

  const currentView = views.find((v) => v.routeId === selectedView);
  const currentVersions = currentView?.versions || [];

  // 如果没有视图，显示默认状态
  if (views.length === 0) {
    return (
      <div
        style={{
          padding: "8px 12px",
          borderRadius: "8px",
          border: "1px solid var(--border-color)",
          background: "#fff",
          fontSize: "13px",
          color: "var(--text-secondary)",
        }}
      >
        无可用视图
      </div>
    );
  }

  // 如果只有默认视图且无版本，简化显示
  if (views.length === 1 && views[0].isDefault && currentVersions.length === 0) {
    return (
      <div
        style={{
          padding: "8px 12px",
          borderRadius: "8px",
          border: "1px solid var(--border-color)",
          background: "#fff",
          fontSize: "13px",
          color: "var(--text-primary)",
        }}
      >
        {views[0].routeName}
      </div>
    );
  }

  return (
    <div ref={dropdownRef} style={{ position: "relative" }}>
      {/* 触发按钮 */}
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        style={{
          display: "flex",
          alignItems: "center",
          gap: "6px",
          padding: "8px 12px",
          borderRadius: "8px",
          border: "1px solid var(--border-color)",
          background: "#fff",
          cursor: disabled ? "not-allowed" : "pointer",
          fontSize: "13px",
          fontWeight: 500,
          color: "var(--text-primary)",
          minWidth: "140px",
          opacity: disabled ? 0.6 : 1,
        }}
      >
        <span>🎨</span>
        <span>{currentView?.routeName || "选择视图"}</span>
        {selectedVersion && (
          <>
            <span style={{ color: "var(--text-secondary)", margin: "0 2px" }}>·</span>
            <span style={{ color: "var(--text-secondary)" }}>{selectedVersion}</span>
          </>
        )}
        <span
          style={{
            marginLeft: "auto",
            fontSize: "10px",
            color: "var(--text-secondary)",
            transform: isOpen ? "rotate(180deg)" : "rotate(0deg)",
            transition: "transform 0.15s",
          }}
        >
          ▼
        </span>
      </button>

      {/* 下拉面板 */}
      {isOpen && (
        <div
          style={{
            position: "absolute",
            top: "calc(100% + 4px)",
            left: 0,
            background: "#fff",
            border: "1px solid var(--border-color)",
            borderRadius: "12px",
            boxShadow: "0 8px 24px rgba(0,0,0,0.12)",
            zIndex: 100,
            display: "flex",
            overflow: "hidden",
            minWidth: "320px",
          }}
        >
          {/* 左侧: 视图列表 */}
          <div
            style={{
              borderRight: "1px solid var(--border-color)",
              padding: "8px 0",
              minWidth: "160px",
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
              视图
            </div>
            {views.map((view) => (
              <button
                key={view.routeId}
                type="button"
                onClick={() => handleViewSelect(view.routeId)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "8px",
                  width: "100%",
                  padding: "10px 12px",
                  border: "none",
                  background:
                    view.routeId === selectedView ? "rgba(59, 130, 246, 0.08)" : "transparent",
                  cursor: "pointer",
                  fontSize: "13px",
                  color: view.routeId === selectedView ? "#3b82f6" : "var(--text-primary)",
                  fontWeight: view.routeId === selectedView ? 500 : 400,
                  textAlign: "left",
                }}
                onMouseEnter={(e) => {
                  if (view.routeId !== selectedView)
                    e.currentTarget.style.background = "rgba(0,0,0,0.03)";
                }}
                onMouseLeave={(e) => {
                  if (view.routeId !== selectedView)
                    e.currentTarget.style.background = "transparent";
                }}
              >
                <span>{view.routeName}</span>
                {view.isDefault && (
                  <span
                    style={{
                      fontSize: "10px",
                      padding: "2px 4px",
                      borderRadius: "4px",
                      background: "rgba(0,0,0,0.05)",
                      color: "var(--text-secondary)",
                    }}
                  >
                    默认
                  </span>
                )}
                {view.routeId === selectedView && (
                  <span style={{ marginLeft: "auto", fontSize: "12px" }}>✓</span>
                )}
              </button>
            ))}
          </div>

          {/* 右侧: 版本列表 */}
          <div style={{ padding: "8px 0", minWidth: "140px" }}>
            <div
              style={{
                padding: "6px 12px",
                fontSize: "11px",
                fontWeight: 600,
                color: "var(--text-secondary)",
                textTransform: "uppercase",
              }}
            >
              版本
            </div>
            {currentVersions.length === 0 ? (
              <div
                style={{
                  padding: "10px 12px",
                  fontSize: "12px",
                  color: "var(--text-secondary)",
                }}
              >
                无历史版本
              </div>
            ) : (
              currentVersions.map((version) => (
                <button
                  key={version}
                  type="button"
                  onClick={() => handleVersionSelect(version)}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "8px",
                    width: "100%",
                    padding: "10px 12px",
                    border: "none",
                    background:
                      version === selectedVersion ? "rgba(59, 130, 246, 0.08)" : "transparent",
                    cursor: "pointer",
                    fontSize: "13px",
                    color: version === selectedVersion ? "#3b82f6" : "var(--text-primary)",
                    fontWeight: version === selectedVersion ? 500 : 400,
                    textAlign: "left",
                  }}
                  onMouseEnter={(e) => {
                    if (version !== selectedVersion)
                      e.currentTarget.style.background = "rgba(0,0,0,0.03)";
                  }}
                  onMouseLeave={(e) => {
                    if (version !== selectedVersion)
                      e.currentTarget.style.background = "transparent";
                  }}
                >
                  <span>{version}</span>
                  {version === currentView?.active && (
                    <span
                      style={{
                        fontSize: "10px",
                        padding: "2px 4px",
                        borderRadius: "4px",
                        background: "rgba(34, 197, 94, 0.1)",
                        color: "#15803d",
                      }}
                    >
                      当前
                    </span>
                  )}
                  {version === selectedVersion && (
                    <span style={{ marginLeft: "auto", fontSize: "12px" }}>✓</span>
                  )}
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
