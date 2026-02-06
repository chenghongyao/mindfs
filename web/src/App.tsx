import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { JSONUIProvider } from "@json-render/react";
import { Renderer } from "./renderer/Renderer";
import {
  buildDefaultTree,
  type FileEntry,
  type FilePayload,
  type SessionSummary,
  type UITree,
} from "./renderer/defaultTree";
import { registry } from "./renderer/registry";
import { ActionClient } from "./services/actions";
import { applyViewUpdate } from "./services/viewUpdates";
import { mergeViewIntoShell } from "./renderer/merge";

type ManagedDir = {
  id: string;
  root_path: string;
  display_name?: string;
  created_at: string;
  updated_at: string;
};

export function App() {
  const [rootEntries, setRootEntries] = useState<FileEntry[]>([]);
  const [entriesByPath, setEntriesByPath] = useState<
    Record<string, FileEntry[]>
  >({});
  const [expanded, setExpanded] = useState<string[]>([]);
  const [selectedDir, setSelectedDir] = useState<string | null>(null);
  const [mainEntries, setMainEntries] = useState<FileEntry[]>([]);
  const [status, setStatus] = useState("Connecting");
  const [file, setFile] = useState<FilePayload | null>(null);
  const [viewTree, setViewTree] = useState<UITree | null>(null);
  const [pendingView, setPendingView] = useState(false);
  const viewIdRef = useRef<string | null>(null);
  const previousViewRef = useRef<UITree | null>(null);
  const actionClientRef = useRef<ActionClient | null>(null);
  const [currentRootId, setCurrentRootId] = useState<string | null>(null);
  const [managedRootIds, setManagedRootIds] = useState<string[]>([]);
  const managedRootIdsRef = useRef<Set<string>>(new Set());
  const expandedRef = useRef<string[]>([]);
  const [sessions, setSessions] = useState<SessionSummary[]>([]);
  const [selectedSession, setSelectedSession] = useState<SessionSummary | null>(null);
  const [rightCollapsed, setRightCollapsed] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);

  useEffect(() => {
    actionClientRef.current = new ActionClient();
    let cancelled = false;
    const load = async () => {
      try {
        const dirsRes = await fetch("/api/dirs");
        const dirsPayload = await dirsRes.json();
        if (cancelled) return;
        const dirs = (dirsPayload.dirs as ManagedDir[]) ?? [];
        const ids = dirs.map((dir) => dir.id);
        managedRootIdsRef.current = new Set(ids);
        setManagedRootIds(ids);
        const managedEntries: FileEntry[] = dirs.map((dir) => ({
          name: dir.display_name ?? dir.id,
          path: dir.id,
          is_dir: true,
        }));
        setRootEntries(managedEntries);
        if (managedEntries.length === 0) {
          setStatus("No managed dirs");
          return;
        }
        const first = managedEntries[0];
        setCurrentRootId(first.path);
        const treeRes = await fetch(`/api/tree?root=${encodeURIComponent(first.path)}&dir=.`);
        const treePayload = await treeRes.json();
        const list = Array.isArray(treePayload.tree) ? treePayload.tree : [];
        if (cancelled) return;
        setEntriesByPath((prev) => ({ ...prev, [`${first.path}:.`]: list, ".": list }));
        setExpanded([first.path]);
        setSelectedDir(first.path);
        setMainEntries(list);
        setViewTree(null);
        setPendingView(false);
        viewIdRef.current = null;
        previousViewRef.current = null;
        setSelectedSession(null);
        setSettingsOpen(false);
        setStatus("Connected");
      } catch {
        if (cancelled) return;
        setStatus("Failed to load");
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    expandedRef.current = expanded;
  }, [expanded]);

  const handleAcceptView = useCallback(async () => {
    if (!currentRootId) return;
    try {
      await fetch(`/api/view/accept?root=${encodeURIComponent(currentRootId)}`, {
        method: "POST",
      });
      setPendingView(false);
      previousViewRef.current = null;
    } catch {
      // ignore accept errors
    }
  }, [currentRootId]);

  const handleRevertView = useCallback(async () => {
    if (!currentRootId) return;
    try {
      const res = await fetch(`/api/view/revert?root=${encodeURIComponent(currentRootId)}`, {
        method: "POST",
      });
      const payload = await res.json().catch(() => ({}));
      if (payload.view) {
        const next = applyViewUpdate(null, { type: "full", payload: payload.view });
        setViewTree(next as UITree);
      } else if (previousViewRef.current) {
        setViewTree(previousViewRef.current);
      }
      setPendingView(false);
      previousViewRef.current = null;
    } catch {
      setPendingView(false);
    }
  }, [currentRootId]);

  const handleSelectSession = useCallback((session: SessionSummary) => {
    setSelectedSession(session);
    setFile(null);
  }, []);

  const handleToggleRight = useCallback(() => {
    setRightCollapsed((prev) => !prev);
  }, []);

  const handleOpenSettings = useCallback(() => {
    setSettingsOpen((prev) => !prev);
    setRightCollapsed(false);
  }, []);

  const shellTree = useMemo(
    () =>
      buildDefaultTree(
        rootEntries,
        entriesByPath,
        expanded,
        selectedDir,
        currentRootId,
        managedRootIds,
        mainEntries,
        status,
        file,
        pendingView,
        handleAcceptView,
        handleRevertView,
        sessions,
        selectedSession,
        handleSelectSession,
        rightCollapsed,
        handleToggleRight,
        handleOpenSettings,
        settingsOpen
      ),
    [
      rootEntries,
      entriesByPath,
      expanded,
      selectedDir,
      currentRootId,
      managedRootIds,
      mainEntries,
      status,
      file,
      pendingView,
      handleAcceptView,
      handleRevertView,
      sessions,
      selectedSession,
      rightCollapsed,
      handleSelectSession,
      handleToggleRight,
      handleOpenSettings,
      settingsOpen,
    ]
  );
  const tree = useMemo(
    () =>
      selectedSession || file
        ? shellTree
        : mergeViewIntoShell(shellTree, viewTree),
    [shellTree, viewTree, selectedSession, file]
  );

  const actionHandlers = useMemo(
    () => ({
      open: async (params: Record<string, unknown>) => {
        const path = params.path as string | undefined;
        const rootParam = params.root as string | undefined;
        if (!path || !actionClientRef.current) return;
        setStatus(`Opening ${path}`);
        try {
          const resp = await actionClientRef.current.dispatch({
            action: "open",
            path,
            version: "v1",
            root: rootParam ?? currentRootId ?? undefined,
          });
          if (resp.status === "ok" && resp.data?.file) {
            setFile(resp.data.file as FilePayload);
            setSelectedSession(null);
            setStatus("Connected");
            return;
          }
          setStatus(resp.error?.message ?? "Open failed");
          console.error("open failed", resp);
        } catch (err) {
          setStatus("Open failed");
          console.error(err);
        }
      },
      open_dir: async (params: Record<string, unknown>) => {
        const path = params.path as string | undefined;
        const rootParam = params.root as string | undefined;
        if (!path) return;
        if (managedRootIdsRef.current.has(path)) {
          if (expandedRef.current.includes(path)) {
            setExpanded((prev) => prev.filter((p) => p !== path));
            return;
          }
          setCurrentRootId(path);
          const res = await fetch(`/api/tree?root=${encodeURIComponent(path)}&dir=.`);
          const payload = await res.json();
          const list = Array.isArray(payload.tree) ? payload.tree : [];
        setEntriesByPath((prev) => ({ ...prev, [`${path}:.`]: list, ".": list }));
        setExpanded((prev) => (prev.includes(path) ? prev : [...prev, path]));
        setSelectedDir(path);
        setMainEntries(list);
        setFile(null);
        setViewTree(null);
        setPendingView(false);
        viewIdRef.current = null;
        previousViewRef.current = null;
        setSelectedSession(null);
        setSettingsOpen(false);
        return;
      }
        const rootId = rootParam ?? currentRootId ?? "";
        const res = await fetch(
          `/api/tree?root=${encodeURIComponent(rootId)}&dir=${encodeURIComponent(path)}`
        );
        const payload = await res.json();
        const list = Array.isArray(payload.tree) ? payload.tree : [];
        const key = `${rootId}:${path}`;
        setEntriesByPath((prev) => ({ ...prev, [key]: list, [path]: list }));
        setExpanded((prev) =>
          prev.includes(path) ? prev.filter((p) => p !== path) : [...prev, path]
        );
        setSelectedDir(path);
        setMainEntries(list);
        setFile(null);
        setSelectedSession(null);
        setSettingsOpen(false);
      },
    }),
    [currentRootId, expanded]
  );

  useEffect(() => {
    if (!currentRootId) return;
    let cancelled = false;
    const loadSessions = async () => {
      try {
        const res = await fetch(`/api/sessions?root=${encodeURIComponent(currentRootId)}`);
        if (!res.ok) return;
        const payload = await res.json();
        if (cancelled) return;
        const list = Array.isArray(payload.sessions) ? payload.sessions : [];
        setSessions(list);
      } catch {
        // ignore
      }
    };
    const poll = async () => {
      try {
        const res = await fetch(`/api/view?root=${encodeURIComponent(currentRootId)}`);
        if (!res.ok) return;
        const payload = await res.json();
        if (cancelled) return;
        if (payload.view) {
          const viewId = (payload.view_id as string | undefined) ?? null;
          if (viewId && viewIdRef.current === viewId) {
            setPendingView(Boolean(payload.pending));
            return;
          }
          const next = applyViewUpdate(null, { type: "full", payload: payload.view });
          if (payload.pending) {
            previousViewRef.current = viewTree;
            setViewTree(next as UITree);
            setPendingView(true);
          } else {
            setViewTree(next as UITree);
            setPendingView(false);
            previousViewRef.current = null;
          }
          viewIdRef.current = viewId;
        }
      } catch {
        // ignore polling errors
      }
    };
    poll();
    loadSessions();
    const timer = window.setInterval(poll, 5000);
    const sessionTimer = window.setInterval(loadSessions, 8000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
      window.clearInterval(sessionTimer);
    };
  }, [currentRootId, viewTree]);

  return (
    <JSONUIProvider
      registry={registry}
      initialData={{}}
      actionHandlers={actionHandlers}
    >
      <Renderer tree={tree} registry={registry} />
    </JSONUIProvider>
  );
}
