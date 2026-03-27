import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

function listPublicAssets(publicDir: string): string[] {
  if (!fs.existsSync(publicDir)) {
    return [];
  }

  const urls: string[] = [];

  const walk = (currentDir: string) => {
    for (const entry of fs.readdirSync(currentDir, { withFileTypes: true })) {
      const absolutePath = path.join(currentDir, entry.name);
      if (entry.isDirectory()) {
        walk(absolutePath);
        continue;
      }

      const relativePath = path.relative(publicDir, absolutePath).split(path.sep).join("/");
      if (!relativePath || relativePath === "service-worker.js") {
        continue;
      }
      urls.push(`./${relativePath}`);
    }
  };

  walk(publicDir);
  return urls.sort();
}

function buildServiceWorker(precacheUrls: string[], version: string): string {
  return `const SHELL_CACHE = "mindfs-shell-${version}";
const RUNTIME_CACHE = "mindfs-runtime-${version}";
const OFFLINE_URL = new URL("./offline.html", self.location.href).toString();
const INDEX_URL = new URL("./index.html", self.location.href).toString();
const PRECACHE_URLS = ${JSON.stringify(precacheUrls, null, 2)};

self.addEventListener("install", (event) => {
  event.waitUntil((async () => {
    const cache = await caches.open(SHELL_CACHE);
    await cache.addAll(PRECACHE_URLS.map((url) => new URL(url, self.location.href).toString()));
    await self.skipWaiting();
  })());
});

self.addEventListener("activate", (event) => {
  event.waitUntil((async () => {
    const cacheKeys = await caches.keys();
    await Promise.all(cacheKeys
      .filter((key) => key !== SHELL_CACHE && key !== RUNTIME_CACHE)
      .map((key) => caches.delete(key)));
    await self.clients.claim();
  })());
});

self.addEventListener("fetch", (event) => {
  const { request } = event;
  if (request.method !== "GET") {
    return;
  }

  const url = new URL(request.url);
  if (url.origin !== self.location.origin) {
    return;
  }
  if (url.pathname.startsWith("/api/") || url.pathname === "/api" || url.pathname === "/ws" || url.pathname === "/health") {
    return;
  }

  if (request.mode === "navigate") {
    event.respondWith(handleNavigationRequest(request));
    return;
  }

  event.respondWith(handleStaticRequest(request));
});

async function handleNavigationRequest(request) {
  const cache = await caches.open(SHELL_CACHE);
  try {
    const response = await fetch(request);
    cache.put(INDEX_URL, response.clone()).catch(() => {});
    return response;
  } catch {
    const cachedIndex = await cache.match(INDEX_URL);
    if (cachedIndex) {
      return cachedIndex;
    }
    const offlineResponse = await cache.match(OFFLINE_URL);
    if (offlineResponse) {
      return offlineResponse;
    }
    throw new Error("offline");
  }
}

async function handleStaticRequest(request) {
  const shellCache = await caches.open(SHELL_CACHE);
  const cachedShellResponse = await shellCache.match(request);
  if (cachedShellResponse) {
    fetchAndStore(request, shellCache);
    return cachedShellResponse;
  }

  const runtimeCache = await caches.open(RUNTIME_CACHE);
  const cachedRuntimeResponse = await runtimeCache.match(request);
  if (cachedRuntimeResponse) {
    fetchAndStore(request, runtimeCache);
    return cachedRuntimeResponse;
  }

  try {
    const response = await fetch(request);
    if (response.ok) {
      runtimeCache.put(request, response.clone()).catch(() => {});
    }
    return response;
  } catch {
    if (request.destination === "image") {
      const iconResponse = await shellCache.match(new URL("./pwa-192.png", self.location.href).toString());
      if (iconResponse) {
        return iconResponse;
      }
    }
    throw new Error("asset unavailable");
  }
}

async function fetchAndStore(request, cache) {
  try {
    const response = await fetch(request);
    if (response.ok) {
      await cache.put(request, response.clone());
    }
  } catch {
  }
}
`;
}

function autoPrecachePlugin() {
  return {
    name: "mindfs-auto-precache",
    apply: "build" as const,
    generateBundle(this: { emitFile: (file: { type: "asset"; fileName: string; source: string }) => void }, _options: unknown, bundle: Record<string, { fileName: string; type: string }>) {
      const publicDir = path.resolve(__dirname, "public");
      const publicAssets = listPublicAssets(publicDir);
      const bundleAssets = Object.values(bundle)
        .map((item) => item.fileName)
        .filter((fileName) => fileName !== "service-worker.js")
        .filter((fileName) => !fileName.endsWith(".map"))
        .map((fileName) => `./${fileName}`)
        .sort();
      const precacheUrls = Array.from(new Set(["./", "./index.html", ...publicAssets, ...bundleAssets]));
      const version = crypto
        .createHash("sha256")
        .update(JSON.stringify(precacheUrls))
        .digest("hex")
        .slice(0, 12);

      this.emitFile({
        type: "asset",
        fileName: "service-worker.js",
        source: buildServiceWorker(precacheUrls, version),
      });
    },
  };
}

export default defineConfig({
  base: "./",
  plugins: [tailwindcss(), react(), autoPrecachePlugin()],
  server: {
    host: "0.0.0.0",
    proxy: {
      "/api": "http://localhost:7331",
      "/ws": {
        target: "ws://localhost:7331",
        ws: true,
      },
    },
  },
});
