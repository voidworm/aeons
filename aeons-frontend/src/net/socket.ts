export type Envelope = { type: string; payload?: unknown };

let socket: WebSocket | null = null;

export function connect(onMessage: (msg: Envelope) => void): Promise<WebSocket> {
  const promise = new Promise<WebSocket>((resolve, reject) => {
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onopen = () => resolve(ws);
    ws.onmessage = (event: MessageEvent<string>) => {
      onMessage(JSON.parse(event.data) as Envelope);
    };
    ws.onerror = () => reject(new Error('WebSocket connection failed'));

    socket = ws;
  });

  return promise;
}

export function send(type: string, payload?: unknown): void {
  socket?.send(JSON.stringify({ type, payload }));
}

export function disconnect(): void {
  socket?.close();
  socket = null;
}
