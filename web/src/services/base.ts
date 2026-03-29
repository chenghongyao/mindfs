function relayPrefix(): string {
  if (typeof window === "undefined") {
    return "";
  }
  const match = /^\/n\/[^/]+/.exec(window.location.pathname);
  return match ? match[0] : "";
}

export function isRelayNodePage(): boolean {
  if (typeof window === "undefined") {
    return false;
  }
  return /^\/n\/[^/]+/.test(window.location.pathname);
}

function ensureLeadingSlash(path: string): string {
  return path.startsWith("/") ? path : `/${path}`;
}

export function appPath(path: string): string {
  return `${relayPrefix()}${ensureLeadingSlash(path)}`;
}

export function appURL(path: string, params?: URLSearchParams): string {
  const pathname = appPath(path);
  if (!params || !params.toString()) {
    return pathname;
  }
  return `${pathname}?${params.toString()}`;
}

export function wsURL(path: string, params?: URLSearchParams): string {
  const protocol = typeof window !== "undefined" && window.location.protocol === "https:"
    ? "wss:"
    : "ws:";
  const host = typeof window !== "undefined" ? window.location.host : "";
  return `${protocol}//${host}${appURL(path, params)}`;
}
