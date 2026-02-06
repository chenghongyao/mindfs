export type ConnectionOptions = {
  baseUrl: string;
  token?: string;
};

export function connectToServer({ baseUrl, token }: ConnectionOptions): WebSocket {
  const url = new URL(baseUrl);
  if (url.protocol === "http:") {
    url.protocol = "ws:";
  }
  if (url.protocol === "https:") {
    url.protocol = "wss:";
  }
  url.pathname = "/ws";
  if (token) {
    url.searchParams.set("token", token);
  }
  return new WebSocket(url.toString());
}
