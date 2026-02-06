import React, { useCallback } from "react";

export type Shortcut = {
  id: string;
  label: string;
  action: string;
  position?: "left" | "center" | "right";
  type?: "button" | "text";
  icon?: string;
  params?: Record<string, unknown>;
  disabled?: boolean;
};

type ShortcutBarProps = {
  shortcuts: Shortcut[];
  onAction: (action: string, params?: Record<string, unknown>) => void;
};

export function ShortcutBar({ shortcuts, onAction }: ShortcutBarProps): JSX.Element | null {
  const handleClick = useCallback(
    (shortcut: Shortcut) => {
      if (shortcut.disabled) return;
      onAction(shortcut.action, shortcut.params);
    },
    [onAction]
  );

  if (!shortcuts || shortcuts.length === 0) {
    return null;
  }

  // Group by position
  const left = shortcuts.filter((s) => s.position === "left");
  const center = shortcuts.filter((s) => !s.position || s.position === "center");
  const right = shortcuts.filter((s) => s.position === "right");

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "12px 16px",
        borderTop: "1px solid var(--border-color)",
        background: "rgba(255,255,255,0.95)",
        backdropFilter: "blur(8px)",
      }}
    >
      {/* Left */}
      <div style={{ display: "flex", gap: "8px" }}>
        {left.map((shortcut) => (
          <ShortcutButton key={shortcut.id} shortcut={shortcut} onClick={handleClick} />
        ))}
      </div>

      {/* Center */}
      <div style={{ display: "flex", gap: "8px" }}>
        {center.map((shortcut) => (
          <ShortcutButton key={shortcut.id} shortcut={shortcut} onClick={handleClick} />
        ))}
      </div>

      {/* Right */}
      <div style={{ display: "flex", gap: "8px" }}>
        {right.map((shortcut) => (
          <ShortcutButton key={shortcut.id} shortcut={shortcut} onClick={handleClick} />
        ))}
      </div>
    </div>
  );
}

type ShortcutButtonProps = {
  shortcut: Shortcut;
  onClick: (shortcut: Shortcut) => void;
};

function ShortcutButton({ shortcut, onClick }: ShortcutButtonProps): JSX.Element {
  const isText = shortcut.type === "text";

  if (isText) {
    return (
      <span
        style={{
          fontSize: "13px",
          color: "var(--text-secondary)",
          padding: "8px 12px",
        }}
      >
        {shortcut.icon && <span style={{ marginRight: "6px" }}>{shortcut.icon}</span>}
        {shortcut.label}
      </span>
    );
  }

  return (
    <button
      type="button"
      onClick={() => onClick(shortcut)}
      disabled={shortcut.disabled}
      style={{
        display: "flex",
        alignItems: "center",
        gap: "6px",
        padding: "8px 16px",
        borderRadius: "8px",
        border: "1px solid var(--border-color)",
        background: "#fff",
        fontSize: "13px",
        fontWeight: 500,
        color: shortcut.disabled ? "var(--text-secondary)" : "var(--text-primary)",
        cursor: shortcut.disabled ? "not-allowed" : "pointer",
        opacity: shortcut.disabled ? 0.6 : 1,
        transition: "background 0.15s, border-color 0.15s",
      }}
      onMouseEnter={(e) => {
        if (!shortcut.disabled) {
          e.currentTarget.style.background = "rgba(0,0,0,0.02)";
          e.currentTarget.style.borderColor = "var(--accent-color)";
        }
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = "#fff";
        e.currentTarget.style.borderColor = "var(--border-color)";
      }}
    >
      {shortcut.icon && <span>{shortcut.icon}</span>}
      <span>{shortcut.label}</span>
    </button>
  );
}
