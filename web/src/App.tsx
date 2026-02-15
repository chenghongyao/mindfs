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

type ViewRoutePayload = {
  view_data?: UITree | null;
};

type TreeResponse = {
  entries?: FileEntry[];
  view_routes?: ViewRoutePayload[];
};

type FileResponse = {
  file?: FilePayload;
  view_routes?: ViewRoutePayload[];
};

export function App() {
  const [rootEntries, setRootEntries] = useState<FileEntry[]>([]);
  const [entriesByPath, setEntriesByPath] = useState<Record<string, FileEntry[]>>({});
  const [expanded, setExpanded] = useState<string[]>([]);
  const [selectedDir, setSelectedDir] = useState<string | null>(null);
  const [mainEntries, setMainEntries] = useState<FileEntry[]>([]);
  const [status, setStatus] = useState("Connecting");
  const [file, setFile] = useState<FilePayload | null>(null);
  const [viewTree, setViewTree] = useState<UITree | null>(null);
  const [currentRootId, setCurrentRootId] = useState<string | null>(null);
  const [managedRootIds, setManagedRootIds] = useState<string[]>([]);
  
  // 关键：创建 Ref 镜像以解决 Action 闭包状态滞后问题
  const currentRootIdRef = useRef<string | null>(null);
  const managedRootIdsRef = useRef<Set<string>>(new Set());
  const expandedRef = useRef<string[]>([]);
  const selectedDirRef = useRef<string | null>(null);
  const fileRef = useRef<FilePayload | null>(null);
  
  const [sessions, setSessions] = useState<any[]>([]);
  const [sessionsByRoot, setSessionsByRoot] = useState<Record<string, any[]>>({});
  const [sessionsReady, setSessionsReady] = useState(false);
  const [selectedSession, setSelectedSession] = useState<SessionSummary | null>(null);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);
  const [currentSessionExchanges, setCurrentSessionExchanges] = useState<any[]>([]);
  const [currentSessionByRoot, setCurrentSessionByRoot] = useState<Record<string, Session>>({});
  const [sessionExchangesByRootSession, setSessionExchangesByRootSession] = useState<Record<string, any[]>>({});
  const [rightCollapsed, setRightCollapsed] = useState(false);
  const [isFloatingOpen, setIsFloatingOpen] = useState(false);
  const rootSessionKey = useCallback((rootId: string, sessionKey: string) => `${rootId}::${sessionKey}`, []);

  const buildSessionNameFromMessage = useCallback((message: string): string => {
    const oneLine = message.replace(/\s+/g, " ").trim();
    if (!oneLine) return "";
    const max = 60;
    return oneLine.length > max ? `${oneLine.slice(0, max)}...` : oneLine;
  }, []);

  // 同步状态到 Ref
  useEffect(() => { currentRootIdRef.current = currentRootId; }, [currentRootId]);
  useEffect(() => { expandedRef.current = expanded; }, [expanded]);
  useEffect(() => { selectedDirRef.current = selectedDir; }, [selectedDir]);
  useEffect(() => { fileRef.current = file; }, [file]);

  const pickViewTree = useCallback((routes: ViewRoutePayload[] | undefined): UITree | null => {
    const first = (Array.isArray(routes) ? routes : []).find((item) => item?.view_data);
    return (first?.view_data as UITree | null) || null;
  }, []);

  const normalizeTreeResponse = useCallback((payload: unknown): { entries: FileEntry[]; viewRoutes: ViewRoutePayload[] } => {
    if (Array.isArray(payload)) {
      return { entries: payload as FileEntry[], viewRoutes: [] };
    }
    const obj = (payload && typeof payload === "object") ? (payload as TreeResponse) : {};
    return {
      entries: Array.isArray(obj.entries) ? obj.entries : [],
      viewRoutes: Array.isArray(obj.view_routes) ? obj.view_routes : [],
    };
  }, []);

  const normalizeFileResponse = useCallback((payload: unknown): { file: FilePayload | null; viewRoutes: ViewRoutePayload[] } => {
    const obj = (payload && typeof payload === "object") ? (payload as FileResponse) : {};
    if (obj.file && typeof obj.file === "object") {
      return {
        file: obj.file,
        viewRoutes: Array.isArray(obj.view_routes) ? obj.view_routes : [],
      };
    }
    // Backward compatibility: old API returned file payload directly.
    const raw = payload as Record<string, unknown> | null;
    if (raw && typeof raw.path === "string") {
      return {
        file: raw as unknown as FilePayload,
        viewRoutes: [],
      };
    }
    return { file: null, viewRoutes: [] };
  }, []);
  useEffect(() => {
    if (isFloatingOpen) return;
    if (!currentRootId) return;
    const session = currentSessionByRoot[currentRootId];
    if (!session) {
      setCurrentSession(null);
      setCurrentSessionExchanges([]);
      return;
    }
    setCurrentSession(session);
    setCurrentSessionExchanges(sessionExchangesByRootSession[rootSessionKey(currentRootId, session.key)] || []);
  }, [currentRootId, currentSessionByRoot, sessionExchangesByRootSession, rootSessionKey, isFloatingOpen]);

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
        if (managedEntries.length === 0) return;

        const first = managedEntries[0];
        setCurrentRootId(first.path);
        const treeRes = await fetch(`/api/tree?root=${encodeURIComponent(first.path)}&dir=.`);
        const treePayload = await treeRes.json();
        if (cancelled) return;
        const { entries, viewRoutes } = normalizeTreeResponse(treePayload);
        
        const cacheKey = `${first.path}:.`;
        setEntriesByPath((prev) => ({ ...prev, [cacheKey]: entries }));
        setExpanded([first.path]);
        setSelectedDir(first.path);
        setMainEntries(entries);
        setViewTree(pickViewTree(viewRoutes));
        setStatus("Connected");
      } catch (err) { console.error("Init failed:", err); }
    };
    load();
    return () => { cancelled = true; };
  }, [normalizeTreeResponse, pickViewTree]);

  const handleSelectSession = useCallback(
    async (session: any) => {
      const key = session?.key || session?.session_key;
      const targetRoot = (session?.root_id as string | undefined) || currentRootIdRef.current;
      
      console.log("[handleSelectSession] Navigating:", { key, targetRoot });
      
      if (!targetRoot || !key) {
        console.error("[handleSelectSession] Failed: context missing.");
        return;
      }
      
      try {
        const full = await sessionService.getSession(targetRoot, key);
        if (!full) return;
        
        setSelectedSession(full as any);
        setFile(null); 
        
        if (full.status !== "closed") {
          setCurrentSessionByRoot((prev) => ({ ...prev, [targetRoot]: full as Session }));
          setSessionExchangesByRootSession((prev) => ({ ...prev, [rootSessionKey(targetRoot, full.key)]: full.exchanges || [] }));
          setCurrentSession(full as any);
          setCurrentSessionExchanges(full.exchanges || []);
          setIsFloatingOpen(true);
        } else {
          setIsFloatingOpen(false);
        }
      } catch (err) { console.error(err); }
    },
    [rootSessionKey]
  );

  const handleSendMessage = useCallback(
    async (message: string, mode: "chat" | "view" | "skill", agent: string) => {
      const activeRoot = currentRootIdRef.current;
      if (!activeRoot) return;
      
      let session = currentSessionByRoot[activeRoot];
      if (!session || session.status === "closed" || session.type !== mode || session.agent !== agent) {
        const sessionName = buildSessionNameFromMessage(message);
        session = await sessionService.createSession(activeRoot, mode, agent, sessionName);
        if (!session) return;
        setCurrentSessionByRoot((prev) => ({ ...prev, [activeRoot]: session as Session }));
        setSessionExchangesByRootSession((prev) => ({ ...prev, [rootSessionKey(activeRoot, session!.key)]: [] }));
        setCurrentSession(session);
        setCurrentSessionExchanges([]); 
      } else {
        setCurrentSession(session);
        setCurrentSessionExchanges(sessionExchangesByRootSession[rootSessionKey(activeRoot, session.key)] || []);
      }
      
      setIsFloatingOpen(true);
      const nowISO = new Date().toISOString();
      const newUserExchange = { role: "user", content: message, timestamp: nowISO };
      const exchangeKey = rootSessionKey(activeRoot, session.key);
      setSessionExchangesByRootSession((prev) => ({
        ...prev,
        [exchangeKey]: [...(prev[exchangeKey] || []), newUserExchange],
      }));
      setCurrentSessionExchanges((prev) => [...prev, newUserExchange]);
      const context = buildClientContext({
        currentRoot: activeRoot,
        currentPath: file?.path ?? selectedDir ?? undefined,
      });
      await sessionService.sendMessage(activeRoot, session.key, message, context);
    },
    [currentSessionByRoot, sessionExchangesByRootSession, buildSessionNameFromMessage, file?.path, selectedDir, rootSessionKey]
  );

  const handleAgentResponseAppend = useCallback((content: string) => {
    const activeRoot = currentRootIdRef.current;
    if (activeRoot && currentSession) {
      const key = rootSessionKey(activeRoot, currentSession.key);
      setSessionExchangesByRootSession((prev) => ({
        ...prev,
        [key]: [...(prev[key] || []), { role: "agent", content, timestamp: new Date().toISOString() }],
      }));
    }
    const newAgentExchange = { role: "agent", content, timestamp: new Date().toISOString() };
    setCurrentSessionExchanges((prev) => [...prev, newAgentExchange]);
  }, [currentSession, rootSessionKey]);

  const handleOpenBubbleSession = useCallback((session: any) => {
    if (!session?.key) return;
    const rootID = session.root_id as string | undefined;
    if (!rootID) return;
    const exchangeKey = rootSessionKey(rootID, session.key);
    const exchanges = sessionExchangesByRootSession[exchangeKey] || [];
    const full = { ...session, exchanges };
    // Floating panel is an overlay shortcut and should not affect main view selection.
    setCurrentSession(full);
    setCurrentSessionExchanges(exchanges);
  }, [sessionExchangesByRootSession, rootSessionKey]);

  const activeSessions = useMemo(() => {
    const list: any[] = [];
    Object.entries(sessionsByRoot).forEach(([rootID, rootSessions]) => {
      (rootSessions || []).forEach((s: any) => {
        if (!s || s.status === "closed" || !s.key) return;
        const key = rootSessionKey(rootID, s.key);
        const exchanges = sessionExchangesByRootSession[key] || [];
        if (exchanges.length === 0) return;
        list.push({ ...s, root_id: rootID, exchanges });
      });
    });
    return list;
  }, [sessionsByRoot, sessionExchangesByRootSession, rootSessionKey]);

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
        handleOpenBubbleSession,
        sessionsReady ? activeSessions : [],
        currentSession ? { ...currentSession, exchanges: currentSessionExchanges } : null,
        handleSendMessage,
        () => setIsFloatingOpen((prev) => !prev),
        rightCollapsed,
        () => setRightCollapsed((prev) => !prev),
        isFloatingOpen,
        setIsFloatingOpen,
        handleAgentResponseAppend
      ),
    [rootEntries, entriesByPath, expanded, selectedDir, currentRootId, managedRootIds, mainEntries, status, file, sessions, selectedSession, activeSessions, currentSession, currentSessionExchanges, handleSendMessage, rightCollapsed, handleSelectSession, handleOpenBubbleSession, isFloatingOpen, handleAgentResponseAppend]
  );

  const tree = useMemo(() => {
    const isSelectedSessionActive = currentSession && (selectedSession?.key === currentSession.key || selectedSession?.session_key === currentSession.key);
    const showSessionInMain = selectedSession && !isSelectedSessionActive;
    return showSessionInMain || file ? shellTree : mergeViewIntoShell(shellTree, viewTree);
  }, [shellTree, viewTree, selectedSession, currentSession, file]);

  const actionHandlers = useMemo(
    () => {
      const getParentKeys = (path: string, root: string) => {
        const parts = path.split('/').filter(Boolean);
        const parentKeys = [root];
        for (let i = 1; i < parts.length; i++) {
          const parentPath = parts.slice(0, i).join('/');
          parentKeys.push(`${root}:${parentPath}`);
        }
        return parentKeys;
      };

      return {
        select_session: async (params: Record<string, unknown>) => {
          console.log("[Action:select_session] Invoked with:", params);
          if (params.key) {
            handleSelectSession({ key: params.key, root_id: params.root });
          }
        },
        open: async (params: Record<string, unknown>) => {
          const path = params.path as string | undefined;
          const rootParam = params.root as string | undefined;
          if (!path) return;
          const root = rootParam || currentRootIdRef.current || managedRootIds[0] || "";
          if (!root) return;

          const parents = getParentKeys(path, root);
          setExpanded((prev) => Array.from(new Set([...prev, ...parents])));
          try {
            if (root !== currentRootIdRef.current) setCurrentRootId(root);
            const query = new URLSearchParams({ path, root });
            const res = await fetch(`/api/file?${query.toString()}`);
            const payload = await res.json().catch(() => ({}));
            if (res.ok) {
              const next = normalizeFileResponse(payload);
              if (next.file) {
                setFile(next.file);
              }
              setViewTree(pickViewTree(next.viewRoutes));
              setSelectedSession(null);
            }
          } catch (err) {}
        },
        open_dir: async (params: Record<string, unknown>) => {
          const path = params.path as string | undefined;
          const rootParam = params.root as string | undefined;
          const isToggle = !!params.toggle;
          if (!path) return;
          const isActuallyRoot = managedRootIdsRef.current.has(path);
          const root = isActuallyRoot ? path : (rootParam || currentRootIdRef.current || managedRootIds[0]);
          const expandedKey = isActuallyRoot ? path : `${root}:${path}`;
          const apiDir = isActuallyRoot ? "." : path;
          if (isToggle && expandedRef.current.includes(expandedKey)) {
            setExpanded((prev) => prev.filter(k => k !== expandedKey));
            return;
          }
          if (isActuallyRoot) {
            setCurrentRootId(path);
            setExpanded((prev) => Array.from(new Set([...prev, path])));
          } else {
            const parents = getParentKeys(path, root);
            setExpanded((prev) => Array.from(new Set([...prev, ...parents, expandedKey])));
          }
          try {
            const res = await fetch(`/api/tree?root=${encodeURIComponent(root)}&dir=${encodeURIComponent(apiDir)}`);
            const payload = await res.json();
            const parsed = normalizeTreeResponse(payload);
            const cacheKey = `${root}:${apiDir}`;
            setEntriesByPath((prev) => ({ ...prev, [cacheKey]: parsed.entries }));
            setSelectedDir(path);
            setMainEntries(parsed.entries);
            setViewTree(pickViewTree(parsed.viewRoutes));
            setFile(null);
            setSelectedSession(null);
          } catch (err) {}
        },
      };
    },
    [handleSelectSession, normalizeFileResponse, normalizeTreeResponse, pickViewTree]
  );

  useEffect(() => {
    if (!currentRootId) return;
    sessionService.connect(currentRootId);
    let cancelled = false;
    const loadSessions = async (rootID: string) => {
      try {
        const res = await fetch(`/api/sessions?root=${encodeURIComponent(rootID)}`);
        const payload = await res.json();
        if (!cancelled) {
          const next = Array.isArray(payload) ? payload : [];
          if (rootID === currentRootIdRef.current) {
            setSessions(next);
          }
          setSessionsByRoot((prev) => ({ ...prev, [rootID]: next }));
        }
      } catch {}
    };

    const refreshCurrentFile = async (rootID: string, path: string) => {
      try {
        const query = new URLSearchParams({ path, root: rootID });
        const res = await fetch(`/api/file?${query.toString()}`);
        const payload = await res.json().catch(() => ({}));
        if (!res.ok || cancelled) return;
        const next = normalizeFileResponse(payload);
        if (!next.file) return;
        setFile(next.file);
        setViewTree(pickViewTree(next.viewRoutes));
      } catch {}
    };

    const refreshCurrentDir = async (rootID: string, dir: string) => {
      const apiDir = managedRootIdsRef.current.has(dir) ? "." : dir;
      try {
        const res = await fetch(`/api/tree?root=${encodeURIComponent(rootID)}&dir=${encodeURIComponent(apiDir)}`);
        const payload = await res.json().catch(() => ({}));
        if (!res.ok || cancelled) return;
        const parsed = normalizeTreeResponse(payload);
        const cacheKey = `${rootID}:${apiDir}`;
        setEntriesByPath((prev) => ({ ...prev, [cacheKey]: parsed.entries }));
        setMainEntries(parsed.entries);
        setViewTree(pickViewTree(parsed.viewRoutes));
      } catch {}
    };

    const isPathInDir = (path: string, dir: string, rootID: string): boolean => {
      if (!path) return false;
      if (managedRootIdsRef.current.has(dir) || dir === rootID || dir === "." || dir === "") {
        return true;
      }
      return path === dir || path.startsWith(`${dir}/`);
    };

    const unsubscribeEvents = sessionService.subscribeEvents((event) => {
      if (["session.done", "session.created", "session.closed", "session.resumed"].includes(event.type)) {
        loadSessions(currentRootId);
        return;
      }
      if (event.type !== "file.changed") {
        return;
      }
      const payload = event.payload || {};
      const eventRoot = typeof payload.root_id === "string" ? payload.root_id : "";
      const eventPath = typeof payload.path === "string" ? payload.path : "";
      if (!eventRoot || eventRoot !== currentRootIdRef.current) {
        return;
      }
      const activeFile = fileRef.current;
      if (activeFile?.path && activeFile.path === eventPath) {
        refreshCurrentFile(eventRoot, activeFile.path);
        return;
      }
      const activeDir = selectedDirRef.current;
      if (!activeDir) {
        return;
      }
      if (isPathInDir(eventPath, activeDir, eventRoot)) {
        refreshCurrentDir(eventRoot, activeDir);
      }
    });
    loadSessions(currentRootId);
    return () => { cancelled = true; unsubscribeEvents(); sessionService.disconnect(); };
  }, [currentRootId, normalizeFileResponse, normalizeTreeResponse, pickViewTree]);

  useEffect(() => {
    if (managedRootIds.length === 0) return;
    setSessionsReady(false);
    let cancelled = false;
    const loadAllRoots = async () => {
      await Promise.all(
        managedRootIds.map(async (rootID) => {
          try {
            const res = await fetch(`/api/sessions?root=${encodeURIComponent(rootID)}`);
            const payload = await res.json();
            if (cancelled) return;
            const next = Array.isArray(payload) ? payload : [];
            setSessionsByRoot((prev) => ({ ...prev, [rootID]: next }));
            if (rootID === currentRootIdRef.current) {
              setSessions(next);
            }
          } catch {}
        })
      );
      if (!cancelled) {
        setSessionsReady(true);
      }
    };
    loadAllRoots();
    return () => {
      cancelled = true;
    };
  }, [managedRootIds]);

  return (
    <JSONUIProvider registry={registry} initialData={{}} actionHandlers={actionHandlers}>
      <Renderer tree={tree} registry={registry} />
    </JSONUIProvider>
  );
}
