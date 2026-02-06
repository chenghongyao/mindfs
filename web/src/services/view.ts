// View service for MindFS

import type { ViewInfo } from "../components/ViewVersionSelector";

const API_BASE = "/api";

export type ViewData = {
  view: Record<string, unknown> | null;
  updated_at?: string;
  pending?: boolean;
  view_id?: string;
};

export type ViewRouteInfo = {
  route_id: string;
  route_name: string;
  priority: number;
  is_default: boolean;
  versions: string[];
  active?: string;
};

// Fetch current view for a root
export async function fetchView(rootId: string): Promise<ViewData> {
  const response = await fetch(`${API_BASE}/view?root=${encodeURIComponent(rootId)}`);

  if (!response.ok) {
    throw new Error(`Failed to fetch view: ${response.statusText}`);
  }

  return await response.json();
}

// Accept pending view
export async function acceptView(rootId: string): Promise<ViewData> {
  const response = await fetch(`${API_BASE}/view/accept?root=${encodeURIComponent(rootId)}`, {
    method: "POST",
  });

  if (!response.ok) {
    throw new Error(`Failed to accept view: ${response.statusText}`);
  }

  return await response.json();
}

// Revert to previous view
export async function revertView(rootId: string): Promise<ViewData> {
  const response = await fetch(`${API_BASE}/view/revert?root=${encodeURIComponent(rootId)}`, {
    method: "POST",
  });

  if (!response.ok) {
    throw new Error(`Failed to revert view: ${response.statusText}`);
  }

  return await response.json();
}

// Fetch available views for a path
export async function fetchViewRoutes(rootId: string, path?: string): Promise<ViewInfo[]> {
  let url = `${API_BASE}/view/routes?root=${encodeURIComponent(rootId)}`;
  if (path) {
    url += `&path=${encodeURIComponent(path)}`;
  }

  const response = await fetch(url);

  if (!response.ok) {
    // Return empty array if endpoint not available
    return [];
  }

  const data = await response.json();
  return (data.routes || []).map((r: ViewRouteInfo) => ({
    routeId: r.route_id,
    routeName: r.route_name,
    priority: r.priority,
    isDefault: r.is_default,
    versions: r.versions || [],
    active: r.active,
  }));
}

// Switch to a different view route
export async function switchViewRoute(rootId: string, routeId: string): Promise<void> {
  const response = await fetch(`${API_BASE}/view/switch`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      root_id: rootId,
      route_id: routeId,
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to switch view: ${response.statusText}`);
  }
}

// Switch to a different version
export async function switchViewVersion(
  rootId: string,
  routeId: string,
  version: string
): Promise<void> {
  const response = await fetch(`${API_BASE}/view/switch`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      root_id: rootId,
      route_id: routeId,
      version: version,
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to switch version: ${response.statusText}`);
  }
}

// Save user preference for a path
export async function saveViewPreference(
  rootId: string,
  path: string,
  routeId: string,
  version?: string
): Promise<void> {
  const response = await fetch(`${API_BASE}/view/preference`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      root_id: rootId,
      path: path,
      route_id: routeId,
      version: version,
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to save preference: ${response.statusText}`);
  }
}

// Get view versions for a route
export async function fetchViewVersions(rootId: string, routeId: string): Promise<string[]> {
  const response = await fetch(
    `${API_BASE}/view/versions/${encodeURIComponent(routeId)}?root=${encodeURIComponent(rootId)}`
  );

  if (!response.ok) {
    return [];
  }

  const data = await response.json();
  return data.versions || [];
}

// Generate a new view version
export async function generateView(
  rootId: string,
  routeId: string,
  prompt: string,
  baseVersion?: string
): Promise<{ sessionKey: string }> {
  const response = await fetch(`${API_BASE}/view/generate`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      root_id: rootId,
      route_id: routeId,
      prompt: prompt,
      base_version: baseVersion,
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to generate view: ${response.statusText}`);
  }

  return await response.json();
}
