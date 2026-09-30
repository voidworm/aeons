export type Envelope = { type: string; payload?: Record<string, unknown> };

export function connect(onMessage: (msg: Envelope) => void): WebSocket {
  const ws = new WebSocket('ws://localhost:8080/ws');

  ws.onmessage = (event: MessageEvent<string>) => {
    onMessage(JSON.parse(event.data) as Envelope);
  };

  return ws;
}

