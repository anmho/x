import { stdin, stdout } from "node:process";

export type Json = null | boolean | number | string | Json[] | { [key: string]: Json };

type RequestHandler = (params: any) => Promise<any> | any;

type ServerOptions = {
  serverName: string;
  serverVersion: string;
  toolHandlers: Record<string, (args: Record<string, unknown>) => Promise<string> | string>;
  tools: Array<{
    name: string;
    description: string;
    inputSchema: Record<string, unknown>;
  }>;
};

function ok(id: Json, result: Json) {
  stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id, result })}\n`);
}

function err(id: Json, code: number, message: string) {
  stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id, error: { code, message } })}\n`);
}

export function createMcpServer(options: ServerOptions) {
  const handlers: Record<string, RequestHandler> = {
    initialize: async () => ({
      protocolVersion: "2024-11-05",
      capabilities: {
        tools: {},
      },
      serverInfo: {
        name: options.serverName,
        version: options.serverVersion,
      },
    }),
    "notifications/initialized": async () => null,
    ping: async () => ({}),
    "tools/list": async () => ({
      tools: options.tools,
    }),
    "tools/call": async (params: any) => {
      const name = String(params?.name ?? "");
      const handler = options.toolHandlers[name];
      if (!handler) {
        throw new Error(`Unknown tool: ${name}`);
      }
      const text = await handler((params?.arguments ?? {}) as Record<string, unknown>);
      return {
        content: [{ type: "text", text }],
        isError: false,
      };
    },
  };

  let buffer = "";
  stdin.setEncoding("utf8");
  stdin.on("data", (chunk: string) => {
    buffer += chunk;
    while (true) {
      const newlineIndex = buffer.indexOf("\n");
      if (newlineIndex === -1) {
        return;
      }
      const line = buffer.slice(0, newlineIndex).trim();
      buffer = buffer.slice(newlineIndex + 1);
      if (!line) {
        continue;
      }

      let message: any;
      try {
        message = JSON.parse(line);
      } catch (error) {
        err(null, -32700, `Invalid JSON: ${(error as Error).message}`);
        continue;
      }

      const method = String(message?.method ?? "");
      const id = message?.id ?? null;
      const handler = handlers[method];

      if (!handler) {
        if (id !== null) {
          err(id, -32601, `Method not found: ${method}`);
        }
        continue;
      }

      Promise.resolve(handler(message?.params))
        .then((result) => {
          if (id !== null) {
            ok(id, result as Json);
          }
        })
        .catch((error: Error) => {
          if (id !== null) {
            err(id, -32000, error.message);
          }
        });
    }
  });
}
