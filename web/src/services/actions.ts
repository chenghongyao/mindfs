export type ActionRequest = {
  action: string;
  path?: string;
  context?: Record<string, unknown>;
  meta?: Record<string, unknown>;
  version: string;
  root?: string;
};

export type ActionResponse = {
  status: "ok" | "error";
  handled: boolean;
  data?: Record<string, unknown>;
  view?: { type: "patch" | "full"; payload: unknown };
  effects?: unknown[];
  error?: { code: string; message: string };
};

type JSONRPCRequest = {
  jsonrpc: "2.0";
  id: string;
  method: string;
  params: Record<string, unknown>;
};

export class ActionClient {
  private ws: WebSocket | null = null;
  private id = 0;

  connect(url: string) {
    this.ws = new WebSocket(url);
  }

  async dispatch(request: ActionRequest): Promise<ActionResponse> {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return this.dispatchWS(request);
    }
    return this.dispatchHTTP(request);
  }

  private dispatchHTTP(request: ActionRequest): Promise<ActionResponse> {
    return fetch("/api/action", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    }).then(async (res) => {
      const payload = await res.json().catch(() => ({}));
      if (!res.ok) {
        const message =
          (payload as { error?: string }).error ?? `http ${res.status}`;
        return {
          status: "error",
          handled: false,
          error: { code: "http_error", message },
        } as ActionResponse;
      }
      return payload as ActionResponse;
    });
  }

  private dispatchWS(request: ActionRequest): Promise<ActionResponse> {
    return new Promise((resolve, reject) => {
      if (!this.ws) {
        reject(new Error("ws not connected"));
        return;
      }
      const id = String(++this.id);
      const message: JSONRPCRequest = {
        jsonrpc: "2.0",
        id,
        method: "action.dispatch",
        params: request as unknown as Record<string, unknown>,
      };
      const ws = this.ws;
      const handler = (event: MessageEvent) => {
        try {
          const payload = JSON.parse(event.data);
          if (payload.id !== id) return;
          ws.removeEventListener("message", handler);
          if (payload.error) {
            reject(new Error(payload.error.message ?? "action error"));
            return;
          }
          resolve(payload.result as ActionResponse);
        } catch (err) {
          reject(err);
        }
      };
      ws.addEventListener("message", handler);
      ws.send(JSON.stringify(message));
    });
  }
}
