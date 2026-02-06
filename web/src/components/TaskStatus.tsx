import React, { useEffect, useState, useCallback } from "react";

export type TaskStatusType = "pending" | "running" | "completed" | "failed" | "cancelled";

export type Task = {
  id: string;
  session_key: string;
  type: string;
  status: TaskStatusType;
  progress: number;
  message?: string;
  error?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  metadata?: Record<string, unknown>;
};

export type TaskUpdate = {
  task_id: string;
  status: TaskStatusType;
  progress: number;
  message?: string;
  error?: string;
};

type TaskStatusProps = {
  tasks: Task[];
  onTaskUpdate?: (update: TaskUpdate) => void;
};

export function TaskStatus({ tasks, onTaskUpdate }: TaskStatusProps): JSX.Element | null {
  const activeTasks = tasks.filter(
    (t) => t.status === "pending" || t.status === "running"
  );

  if (activeTasks.length === 0) {
    return null;
  }

  return (
    <div
      style={{
        position: "fixed",
        bottom: "80px",
        right: "20px",
        display: "flex",
        flexDirection: "column",
        gap: "8px",
        zIndex: 1000,
      }}
    >
      {activeTasks.map((task) => (
        <TaskCard key={task.id} task={task} />
      ))}
    </div>
  );
}

type TaskCardProps = {
  task: Task;
};

function TaskCard({ task }: TaskCardProps): JSX.Element {
  const statusColors: Record<TaskStatusType, string> = {
    pending: "#f59e0b",
    running: "#3b82f6",
    completed: "#10b981",
    failed: "#ef4444",
    cancelled: "#6b7280",
  };

  const statusLabels: Record<TaskStatusType, string> = {
    pending: "Pending",
    running: "Running",
    completed: "Completed",
    failed: "Failed",
    cancelled: "Cancelled",
  };

  const typeIcons: Record<string, string> = {
    chat: "💬",
    view: "🎨",
    skill: "⚡",
  };

  return (
    <div
      style={{
        background: "#fff",
        borderRadius: "12px",
        boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
        padding: "12px 16px",
        minWidth: "240px",
        maxWidth: "320px",
        border: `2px solid ${statusColors[task.status]}`,
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: "8px",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <span style={{ fontSize: "16px" }}>{typeIcons[task.type] || "📋"}</span>
          <span
            style={{
              fontSize: "13px",
              fontWeight: 600,
              color: "var(--text-primary)",
            }}
          >
            {task.type.charAt(0).toUpperCase() + task.type.slice(1)} Task
          </span>
        </div>
        <span
          style={{
            fontSize: "11px",
            fontWeight: 500,
            color: statusColors[task.status],
            textTransform: "uppercase",
          }}
        >
          {statusLabels[task.status]}
        </span>
      </div>

      {task.message && (
        <div
          style={{
            fontSize: "12px",
            color: "var(--text-secondary)",
            marginBottom: "8px",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
        >
          {task.message}
        </div>
      )}

      {task.status === "running" && (
        <div
          style={{
            height: "4px",
            background: "#e5e7eb",
            borderRadius: "2px",
            overflow: "hidden",
          }}
        >
          <div
            style={{
              height: "100%",
              width: `${task.progress}%`,
              background: statusColors.running,
              borderRadius: "2px",
              transition: "width 0.3s ease",
            }}
          />
        </div>
      )}

      {task.error && (
        <div
          style={{
            fontSize: "11px",
            color: statusColors.failed,
            marginTop: "8px",
            padding: "6px 8px",
            background: "#fef2f2",
            borderRadius: "6px",
          }}
        >
          {task.error}
        </div>
      )}
    </div>
  );
}

// Hook for managing task state with WebSocket updates
export function useTaskStatus(
  ws: WebSocket | null,
  sessionKey?: string
): {
  tasks: Task[];
  isLoading: boolean;
  refresh: () => void;
} {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const refresh = useCallback(() => {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    setIsLoading(true);
    const payload: Record<string, unknown> = {};
    if (sessionKey) {
      payload.session_key = sessionKey;
    }

    ws.send(
      JSON.stringify({
        id: `task-list-${Date.now()}`,
        type: "task.list",
        payload,
      })
    );
  }, [ws, sessionKey]);

  useEffect(() => {
    if (!ws) return;

    const handleMessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);

        if (data.type === "task.list" && data.payload?.tasks) {
          setTasks(data.payload.tasks);
          setIsLoading(false);
        }

        if (data.type === "task.update" && data.payload) {
          const update = data.payload as TaskUpdate;
          setTasks((prev) =>
            prev.map((t) =>
              t.id === update.task_id
                ? {
                    ...t,
                    status: update.status,
                    progress: update.progress,
                    message: update.message || t.message,
                    error: update.error || t.error,
                  }
                : t
            )
          );
        }
      } catch {
        // Ignore parse errors
      }
    };

    ws.addEventListener("message", handleMessage);
    return () => ws.removeEventListener("message", handleMessage);
  }, [ws]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { tasks, isLoading, refresh };
}

// Toast notification for task completion
export function TaskCompletionToast({
  task,
  onDismiss,
}: {
  task: Task;
  onDismiss: () => void;
}): JSX.Element {
  useEffect(() => {
    const timer = setTimeout(onDismiss, 5000);
    return () => clearTimeout(timer);
  }, [onDismiss]);

  const isSuccess = task.status === "completed";

  return (
    <div
      style={{
        position: "fixed",
        bottom: "20px",
        right: "20px",
        background: isSuccess ? "#ecfdf5" : "#fef2f2",
        border: `1px solid ${isSuccess ? "#10b981" : "#ef4444"}`,
        borderRadius: "12px",
        padding: "12px 16px",
        boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
        display: "flex",
        alignItems: "center",
        gap: "12px",
        zIndex: 1001,
        animation: "slideIn 0.3s ease",
      }}
    >
      <span style={{ fontSize: "20px" }}>{isSuccess ? "✅" : "❌"}</span>
      <div>
        <div
          style={{
            fontSize: "13px",
            fontWeight: 600,
            color: isSuccess ? "#065f46" : "#991b1b",
          }}
        >
          Task {isSuccess ? "Completed" : "Failed"}
        </div>
        {task.message && (
          <div
            style={{
              fontSize: "12px",
              color: isSuccess ? "#047857" : "#b91c1c",
              marginTop: "2px",
            }}
          >
            {task.message}
          </div>
        )}
      </div>
      <button
        type="button"
        onClick={onDismiss}
        style={{
          background: "none",
          border: "none",
          cursor: "pointer",
          padding: "4px",
          color: "#6b7280",
          fontSize: "16px",
        }}
      >
        ×
      </button>
    </div>
  );
}
