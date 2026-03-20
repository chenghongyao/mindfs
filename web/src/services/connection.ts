import { wsURL } from "./base";

export type ConnectionOptions = {
  baseUrl: string;
  token?: string;
};

export function connectToServer({ baseUrl, token }: ConnectionOptions): WebSocket {
  const url = new URL(baseUrl);
  const params = new URLSearchParams(url.search);
  if (token) {
    params.set("token", token);
  }
  return new WebSocket(wsURL("/ws", params));
}
