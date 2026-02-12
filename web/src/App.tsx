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
import { mergeViewIntoShell } from "./renderer/merge";
import { sessionService, type Session } from "./services/session";
import { buildClientContext } from "./services/context";

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
  const [currentRootId, setCurrentRootId] = useState<string | null>(null);
  const [managedRootIds, setManagedRootIds] = useState<string[]>([]);
  const managedRootIdsRef = useRef<Set<string>>(new Set());
  const expandedRef = useRef<string[]>([]);
  const [sessions, setSessions] = useState<any[]>([]);
  const [selectedSession, setSelectedSession] = useState<SessionSummary | null>(null);
  const selectedSessionRef = useRef<SessionSummary | null>(null);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);
  const [currentSessionExchanges, setCurrentSessionExchanges] = useState<any[]>([]);
  const [rightCollapsed, setRightCollapsed] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [isFloatingOpen, setIsFloatingOpen] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        const dirsRes = await fetch("/api/dirs");
        const dirsPayload = await dirsRes.json();
        if (cancelled) return;
        const dirs = (Array.isArray(dirsPayload) ? dirsPayload : []) as ManagedDir[];
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
        const list = Array.isArray(treePayload) ? treePayload : [];
        if (cancelled) return;
        setEntriesByPath((prev) => ({ ...prev, [`${first.path}:.`]: list, ".": list }));
        setExpanded([first.path]);
        setSelectedDir(first.path);
        setMainEntries(list);
        setViewTree(null);
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

  useEffect(() => {
    selectedSessionRef.current = selectedSession;
  }, [selectedSession]);

  const handleSelectSession = useCallback(
    async (session: any) => {
      const key = session?.key || session?.session_key;
      if (!currentRootId || !key) {
        setSelectedSession(null);
        return;
      }
      
      // If selecting the active session, open the floating panel instead of replacing main view
      if (currentSession && key === currentSession.key && currentSession.status !== "closed") {
        setIsFloatingOpen(true);
        const full = await sessionService.getSession(currentRootId, key);
        if (full) {
          setSelectedSession(full as any);
          setCurrentSessionExchanges(full.exchanges || []);
        }
        return;
      }

      const full = await sessionService.getSession(currentRootId, key);
      if (!full) return;
      
      if (full.status === "closed") {
        setSelectedSession(full as any);
        setFile(null);
        setIsFloatingOpen(false);
      } else {
        // Active or idle session - show in floating panel
        setSelectedSession(full as any);
        setCurrentSession(full as any);
        setCurrentSessionExchanges(full.exchanges || []);
        setIsFloatingOpen(true);
      }
    },
    [currentRootId, currentSession]
  );

  const handleSendMessage = useCallback(
    async (message: string, mode: "chat" | "view" | "skill", agent: string) => {
      if (!currentRootId) return;
      let session = currentSession;
      if (
        !session ||
        session.status === "closed" ||
        session.type !== mode ||
        session.agent !== agent
      ) {
        session = await sessionService.createSession(currentRootId, mode, agent);
        if (!session) return;
        setCurrentSession(session);
      }
      
      // Open floating panel
      setIsFloatingOpen(true);

      // Ensure we have the latest session info
      const snapshot = await sessionService.getSession(currentRootId, session.key);
      setSelectedSession((snapshot as any) ?? (session as any));
      
      // Do NOT setFile(null) if we are in chat/skill mode to keep context visible
      if (mode === "view") {
        setFile(null); // But for view generation, we might want to clear it? 
        // Actually, maybe we keep it until the view is generated.
      }

      const nowISO = new Date().toISOString();
      const newUserExchange = {
        role: "user",
        content: message,
        timestamp: nowISO,
      };

      setCurrentSessionExchanges((prev) => [...prev, newUserExchange]);

      setSelectedSession((prev: any) => {
        if (!prev) return prev;
        const prevKey = prev.key || prev.session_key;
        if (prevKey !== session!.key) return prev;
        const prevExchanges = Array.isArray(prev.exchanges) ? prev.exchanges : [];
        return {
          ...prev,
          exchanges: [
            ...prevExchanges,
            newUserExchange,
          ],
        };
      });

      const context = buildClientContext({
        currentRoot: currentRootId,
        currentPath: file?.path ?? selectedDir ?? undefined,
      });
      await sessionService.sendMessage(currentRootId, session.key, message, context);
    },
    [currentRootId, currentSession, file?.path, selectedDir]
  );

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
        sessions,
        selectedSession,
        handleSelectSession,
        currentSession
          ? {
              key: currentSession.key,
              name: currentSession.name,
              type: currentSession.type,
              status: currentSession.status,
              agent: currentSession.agent,
              exchanges: currentSessionExchanges,
            }
          : null,
        handleSendMessage,
        () => {
          setIsFloatingOpen((prev) => !prev);
        },
        rightCollapsed,
        handleToggleRight,
        handleOpenSettings,
        settingsOpen,
        isFloatingOpen,
        setIsFloatingOpen
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
      sessions,
      selectedSession,
      currentSession,
      handleSendMessage,
      rightCollapsed,
      handleSelectSession,
      handleToggleRight,
      handleOpenSettings,
      settingsOpen,
      isFloatingOpen,
    ]
  );
  const tree = useMemo(() => {
    const isSelectedSessionActive =
      currentSession &&
      (selectedSession?.key === currentSession.key ||
        selectedSession?.session_key === currentSession.key);
    const showSessionInMain = selectedSession && !isSelectedSessionActive;

    return showSessionInMain || file
      ? shellTree
      : mergeViewIntoShell(shellTree, viewTree);
  }, [shellTree, viewTree, selectedSession, currentSession, file]);

  const actionHandlers = useMemo(
    () => ({
      select_session: async (params: Record<string, unknown>) => {
        const key = params.key as string | undefined;
        if (key) {
          handleSelectSession({ key });
        }
      },
      open: async (params: Record<string, unknown>) => {
        const path = params.path as string | undefined;
        const rootParam = params.root as string | undefined;
        if (!path) return;
        setStatus(`Opening ${path}`);
        try {
          const root = rootParam ?? currentRootId ?? undefined;
          
          // 如果点击的文件属于不同的 root，同步切换当前活跃 root
          if (rootParam && rootParam !== currentRootId) {
            setCurrentRootId(rootParam);
          }

          const query = new URLSearchParams({ path });
          if (root) query.set("root", root);
          const res = await fetch(`/api/file?${query.toString()}`);
          const payload = await res.json().catch(() => ({}));
          if (res.ok && payload) {
            setFile(payload as FilePayload);
            setSelectedSession(null); // Clear selected history session to show file
            setStatus("Connected");
            return;
          }
          const msg = (payload as { error?: string })?.error ?? `http ${res.status}`;
          setStatus(msg || "Open failed");
          console.error("open failed", payload);
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
          const list = Array.isArray(payload) ? payload : [];
        setEntriesByPath((prev) => ({ ...prev, [`${path}:.`]: list, ".": list }));
        setExpanded((prev) => (prev.includes(path) ? prev : [...prev, path]));
        setSelectedDir(path);
        setMainEntries(list);
        setFile(null);
        setViewTree(null);
        setSelectedSession(null);
        setSettingsOpen(false);
        return;
      }
        const rootId = rootParam ?? currentRootId ?? "";
        const res = await fetch(
          `/api/tree?root=${encodeURIComponent(rootId)}&dir=${encodeURIComponent(path)}`
        );
        const payload = await res.json();
        const list = Array.isArray(payload) ? payload : [];
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
    sessionService.connect(currentRootId);
    let cancelled = false;
    let viewTimer: number | null = null;
    let sessionTimer: number | null = null;

    const sessionDelay = () => (document.visibilityState === "visible" ? 30000 : 120000);
    const viewDelay = () => (document.visibilityState === "visible" ? 30000 : 120000);

    const loadSessions = async () => {
      try {
        const res = await fetch(`/api/sessions?root=${encodeURIComponent(currentRootId)}`);
        if (!res.ok) return;
        const payload = await res.json();
        if (cancelled) return;
        const list = Array.isArray(payload) ? payload : [];
        setSessions(list);
        setCurrentSession((prev) => {
          if (!prev) return prev;
          const refreshed = list.find((s: any) => s.key === prev.key);
          return refreshed ? (refreshed as Session) : prev;
        });
      } catch {
        // ignore
      }
    };

    const pollView = async () => {
      try {
        const res = await fetch(`/api/view/routes?root=${encodeURIComponent(currentRootId)}`);
        if (!res.ok) return;
        const payload = await res.json();
        if (cancelled) return;
        const routes = Array.isArray(payload) ? payload : [];
        const first = routes.find((r: any) => r.view_data) ?? null;
        if (first?.view_data) {
          setViewTree(first.view_data as UITree);
        }
      } catch {
        // ignore polling errors
      }
    };

    const scheduleSessions = () => {
      if (cancelled) return;
      sessionTimer = window.setTimeout(async () => {
        await loadSessions();
        scheduleSessions();
      }, sessionDelay());
    };

    const scheduleView = () => {
      if (cancelled) return;
      viewTimer = window.setTimeout(async () => {
        await pollView();
        scheduleView();
      }, viewDelay());
    };

    const unsubscribeEvents = sessionService.subscribeEvents((event) => {
      if (
        event.type === "session.done" ||
        event.type === "session.created" ||
        event.type === "session.closed" ||
        event.type === "session.resumed"
      ) {
        void loadSessions();
        
        if (event.sessionKey && currentRootId) {
          void sessionService.getSession(currentRootId, event.sessionKey).then((full) => {
            if (full) {
              // Update floating panel data if this is the active session
              if (currentSession?.key === event.sessionKey) {
                setCurrentSessionExchanges(full.exchanges || []);
              }
              
              // Update main view if this is the selected session
              if (
                (selectedSessionRef.current?.key as string | undefined) === event.sessionKey ||
                (selectedSessionRef.current?.session_key as string | undefined) === event.sessionKey ||
                !selectedSessionRef.current
              ) {
                setSelectedSession(full as any);
              }
            }
          });
        }
      }
    });

    const onVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        void loadSessions();
        void pollView();
      }
    };

    document.addEventListener("visibilitychange", onVisibilityChange);
    void loadSessions();
    void pollView();
    scheduleSessions();
    scheduleView();

    return () => {
      cancelled = true;
      if (viewTimer) window.clearTimeout(viewTimer);
      if (sessionTimer) window.clearTimeout(sessionTimer);
      unsubscribeEvents();
      document.removeEventListener("visibilitychange", onVisibilityChange);
      sessionService.disconnect();
    };
  }, [currentRootId]);

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
