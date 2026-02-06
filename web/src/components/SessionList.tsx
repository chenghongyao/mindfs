import React from "react";

export type SessionStatus = "active" | "idle" | "closed";
export type SessionType = "chat" | "view" | "skill";

export type SessionItem = {
  key: string;
  type: SessionType;
  agent: string;
  name: string;
  status: SessionStatus;
  created_at: string;
  updated_at: string;
  closed_at?: string;
  summary?: {
    title: string;
    description: string;
  };
  related_files?: Array<{ path: string }>;
};

type SessionListProps = {
  sessions: SessionItem[];
  selectedKey?: string;
  onSelect?: (session: SessionItem) => void;
  onRestore?: (session: SessionItem) => void;
};

const typeIcons: Record<SessionType, string> = {
  chat: "💬",
  view: "🎨",
  skill: "⚡",
};

const statusColors: Record<SessionStatus, string> = {
  active: "#22c55e",
  idle: "#f59e0b",
  closed: "#9ca3af",
};

export function SessionList({
  sessions,
  selectedKey = "",
  onSelect,
  onRestore,
}: SessionListProps) {
  if (!sessions.length) {
    return (
      <div style={{ fontSize: "12px", color: "var(--text-secondary)", padding: "8px 0" }}>
        暂无会话记录
      </div>
    );
  }

  // Group sessions by status
  const activeSessions = sessions.filter((s) => s.status === "active" || s.status === "idle");
  const closedSessions = sessions.filter((s) => s.status === "closed");

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
      {/* Active Sessions */}
      {activeSessions.length > 0 && (
        <div>
          <div
            style={{
              fontSize: "11px",
              fontWeight: 600,
              color: "var(--text-secondary)",
              marginBottom: "6px",
              textTransform: "uppercase",
            }}
          >
            进行中
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            {activeSessions.map((session) => (
              <SessionCard
                key={session.key}
                session={session}
                selected={session.key === selectedKey}
                onSelect={onSelect}
              />
            ))}
          </div>
        </div>
      )}

      {/* Closed Sessions */}
      {closedSessions.length > 0 && (
        <div>
          <div
            style={{
              fontSize: "11px",
              fontWeight: 600,
              color: "var(--text-secondary)",
              marginBottom: "6px",
              textTransform: "uppercase",
            }}
          >
            已结束
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            {closedSessions.map((session) => (
              <SessionCard
                key={session.key}
                session={session}
                selected={session.key === selectedKey}
                onSelect={onSelect}
                onRestore={onRestore}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

type SessionCardProps = {
  session: SessionItem;
  selected: boolean;
  onSelect?: (session: SessionItem) => void;
  onRestore?: (session: SessionItem) => void;
};

function SessionCard({ session, selected, onSelect, onRestore }: SessionCardProps) {
  const isClosed = session.status === "closed";
  const displayName = session.summary?.title || session.name || `Session ${session.key.slice(0, 8)}`;

  return (
    <button
      type="button"
      onClick={() => onSelect?.(session)}
      style={{
        textAlign: "left",
        padding: "10px 12px",
        borderRadius: "10px",
        border: selected
          ? "1px solid rgba(59,130,246,0.6)"
          : "1px solid var(--border-color)",
        background: selected
          ? "rgba(59,130,246,0.08)"
          : isClosed
          ? "rgba(0,0,0,0.02)"
          : "rgba(255,255,255,0.8)",
        cursor: "pointer",
        width: "100%",
        opacity: isClosed ? 0.8 : 1,
      }}
    >
      {/* Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "6px",
          marginBottom: "4px",
        }}
      >
        <span style={{ fontSize: "12px" }}>{typeIcons[session.type]}</span>
        <span
          style={{
            fontSize: "11px",
            padding: "1px 6px",
            borderRadius: "4px",
            background: "rgba(0,0,0,0.05)",
            color: "var(--text-secondary)",
          }}
        >
          {session.agent}
        </span>
        <span
          style={{
            width: "6px",
            height: "6px",
            borderRadius: "50%",
            background: statusColors[session.status],
            marginLeft: "auto",
          }}
        />
      </div>

      {/* Title */}
      <div
        style={{
          fontSize: "13px",
          fontWeight: 500,
          color: "var(--text-primary)",
          marginBottom: "4px",
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
        }}
      >
        {displayName}
      </div>

      {/* Description or Files */}
      {session.summary?.description && (
        <div
          style={{
            fontSize: "11px",
            color: "var(--text-secondary)",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
        >
          {session.summary.description}
        </div>
      )}

      {/* Related Files */}
      {session.related_files && session.related_files.length > 0 && (
        <div
          style={{
            fontSize: "11px",
            color: "var(--text-secondary)",
            marginTop: "4px",
          }}
        >
          📎 {session.related_files.length} 个文件
        </div>
      )}

      {/* Restore Button for Closed Sessions */}
      {isClosed && onRestore && (
        <div style={{ marginTop: "6px" }}>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onRestore(session);
            }}
            style={{
              fontSize: "11px",
              padding: "3px 8px",
              borderRadius: "4px",
              border: "1px solid var(--border-color)",
              background: "#fff",
              cursor: "pointer",
              color: "#3b82f6",
            }}
          >
            ↻ 恢复
          </button>
        </div>
      )}

      {/* Time */}
      <div
        style={{
          fontSize: "10px",
          color: "var(--text-secondary)",
          marginTop: "4px",
        }}
      >
        {isClosed && session.closed_at
          ? `结束于 ${formatTime(session.closed_at)}`
          : `更新于 ${formatTime(session.updated_at)}`}
      </div>
    </button>
  );
}

function formatTime(isoString: string): string {
  const date = new Date(isoString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();

  if (diff < 60000) {
    return "刚刚";
  }
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)} 分钟前`;
  }
  if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)} 小时前`;
  }
  return date.toLocaleDateString();
}
