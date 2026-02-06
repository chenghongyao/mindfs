import React, { useState, useCallback, useRef, useEffect } from "react";

type FileEntry = {
  name: string;
  path: string;
  is_dir: boolean;
};

type FileMeta = {
  source_session?: string;
  session_name?: string;
};

type ContextMenuState = {
  visible: boolean;
  x: number;
  y: number;
  entry: FileEntry | null;
  rootId: string | null;
};

type FileTreeProps = {
  entries: FileEntry[];
  childrenByPath: Record<string, FileEntry[]>;
  expanded: string[];
  selectedDir?: string | null;
  rootId?: string | null;
  managedRoots?: string[];
  fileMetas?: Record<string, FileMeta>;
  activeSessionKey?: string | null;
  onSelectFile?: (entry: FileEntry, rootId: string) => void;
  onToggleDir?: (entry: FileEntry, rootId: string) => void;
  onOpenSettings?: (rootId: string) => void;
};

// Simple Icons
const FolderIcon = ({ isOpen }: { isOpen: boolean }) => (
  <svg
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill={isOpen ? "#3b82f6" : "#64748b"}
    stroke="none"
    xmlns="http://www.w3.org/2000/svg"
    style={{ minWidth: 16 }}
  >
    <path d="M4 4C2.89543 4 2 4.89543 2 6V18C2 19.1046 2.89543 20 4 20H20C21.1046 20 22 19.1046 22 18V8C22 6.89543 21.1046 6 20 6H12L10 4H4Z" />
  </svg>
);

const FileIcon = () => (
  <svg
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="#64748b"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    style={{ minWidth: 16 }}
  >
    <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
    <polyline points="14 2 14 8 20 8" />
  </svg>
);

const ChevronRight = ({ isOpen }: { isOpen: boolean }) => (
  <svg
    width="12"
    height="12"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    style={{
      transform: isOpen ? "rotate(90deg)" : "rotate(0deg)",
      transition: "transform 0.1s",
      color: "#9ca3af",
      marginRight: 4,
    }}
  >
    <polyline points="9 18 15 12 9 6" />
  </svg>
);

export function FileTree({
  entries,
  childrenByPath,
  expanded,
  selectedDir,
  rootId,
  managedRoots = [],
  fileMetas = {},
  activeSessionKey,
  onSelectFile,
  onToggleDir,
  onOpenSettings,
}: FileTreeProps) {
  const [contextMenu, setContextMenu] = useState<ContextMenuState>({
    visible: false,
    x: 0,
    y: 0,
    entry: null,
    rootId: null,
  });
  const menuRef = useRef<HTMLDivElement>(null);

  // 关闭右键菜单
  const closeContextMenu = useCallback(() => {
    setContextMenu((prev) => ({ ...prev, visible: false }));
  }, []);

  // 点击外部关闭菜单
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        closeContextMenu();
      }
    };
    if (contextMenu.visible) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [contextMenu.visible, closeContextMenu]);

  // 右键菜单处理
  const handleContextMenu = useCallback(
    (e: React.MouseEvent, entry: FileEntry, entryRoot: string) => {
      if (!entry.is_dir) return; // 只对目录显示右键菜单
      e.preventDefault();
      setContextMenu({
        visible: true,
        x: e.clientX,
        y: e.clientY,
        entry,
        rootId: entryRoot,
      });
    },
    []
  );

  const expandedSet = new Set(expanded);
  const managedSet = new Set(managedRoots);
  const childKeyFor = (entry: FileEntry, depth: number, branchRoot?: string | null) => {
    if (depth === 0 || managedSet.has(entry.path)) {
      return `${entry.path}:.`;
    }
    const rootKey = branchRoot ?? rootId ?? entry.path;
    return `${rootKey}:${entry.path}`;
  };

  const renderEntries = (items: FileEntry[], depth: number, branchRoot?: string | null) => (
    <ul style={{ listStyle: "none", padding: 0, margin: 0 }}>
      {items.map((entry) => {
        const isOpen = expandedSet.has(entry.path);
        const entryRoot =
          depth === 0 || managedSet.has(entry.path)
            ? entry.path
            : branchRoot ?? rootId ?? entry.path;
        const children =
          childrenByPath[childKeyFor(entry, depth, entryRoot)] ??
          childrenByPath[entry.path] ??
          [];
        const isSelected = entry.path === selectedDir;

        // 检查文件是否有 Session 关联
        const meta = fileMetas[entry.path];
        const hasSessionLink = !entry.is_dir && meta?.source_session;
        const isFromActiveSession = hasSessionLink && meta.source_session === activeSessionKey;

        return (
          <li key={entry.path}>
            <button
              type="button"
              onClick={() =>
                entry.is_dir
                  ? onToggleDir?.(entry, entryRoot)
                  : onSelectFile?.(entry, entryRoot)
              }
              onContextMenu={(e) => handleContextMenu(e, entry, entryRoot)}
              title={hasSessionLink ? `来源: ${meta.session_name || meta.source_session}` : undefined}
              style={{
                border: "none",
                background: isSelected ? "var(--selection-bg)" : "transparent",
                cursor: "pointer",
                padding: "4px 8px",
                paddingLeft: 8 + depth * 16,
                display: "flex",
                alignItems: "center",
                gap: "6px",
                width: "100%",
                textAlign: "left",
                color: isSelected ? "var(--accent-color)" : "var(--text-primary)",
                fontSize: "13px",
                borderRadius: "4px",
                transition: "background 0.1s",
                fontWeight: isSelected ? 500 : 400,
                outline: "none",
              }}
              onMouseEnter={(e) => {
                if (!isSelected) e.currentTarget.style.background = "rgba(0,0,0,0.03)";
              }}
              onMouseLeave={(e) => {
                if (!isSelected) e.currentTarget.style.background = "transparent";
              }}
            >
              <div style={{ width: 12, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                 {entry.is_dir && <ChevronRight isOpen={isOpen} />}
              </div>
              {entry.is_dir ? <FolderIcon isOpen={isOpen} /> : <FileIcon />}
              <span style={{ whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis", flex: 1 }}>
                {entry.name}
              </span>
              {/* Session 关联标记 */}
              {hasSessionLink && (
                <span
                  style={{
                    fontSize: "10px",
                    color: isFromActiveSession ? "#3b82f6" : "#9ca3af",
                  }}
                  title={isFromActiveSession ? "当前 Session 生成" : "其他 Session 生成"}
                >
                  {isFromActiveSession ? "◆" : "◇"}
                </span>
              )}
            </button>
            {entry.is_dir && isOpen && children.length > 0
              ? renderEntries(children, depth + 1, entryRoot)
              : null}
          </li>
        );
      })}
    </ul>
  );

  return (
    <div style={{ padding: "12px 0" }}>
      <div
        style={{
          padding: "0 16px 8px",
          marginBottom: "4px",
          borderBottom: "1px solid var(--border-color)",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center"
        }}
      >
        <h3
          style={{
            margin: 0,
            fontSize: "11px",
            fontWeight: 600,
            color: "var(--text-secondary)",
            letterSpacing: "0.5px",
            textTransform: "uppercase",
          }}
        >
          Project
        </h3>
      </div>
      <div style={{ padding: "4px 8px" }}>{renderEntries(entries, 0, null)}</div>

      {/* 右键菜单 */}
      {contextMenu.visible && contextMenu.entry && (
        <div
          ref={menuRef}
          style={{
            position: "fixed",
            left: contextMenu.x,
            top: contextMenu.y,
            background: "#fff",
            border: "1px solid var(--border-color)",
            borderRadius: "8px",
            boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
            padding: "4px 0",
            minWidth: "160px",
            zIndex: 1000,
          }}
        >
          <button
            type="button"
            onClick={() => {
              if (contextMenu.rootId) {
                onOpenSettings?.(contextMenu.rootId);
              }
              closeContextMenu();
            }}
            style={{
              display: "flex",
              alignItems: "center",
              gap: "8px",
              width: "100%",
              padding: "8px 12px",
              border: "none",
              background: "transparent",
              cursor: "pointer",
              fontSize: "13px",
              color: "var(--text-primary)",
              textAlign: "left",
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = "rgba(0,0,0,0.05)";
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = "transparent";
            }}
          >
            <span>⚙️</span>
            <span>目录设置</span>
          </button>
        </div>
      )}
    </div>
  );
}
