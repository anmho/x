/**
 * MCP JSON-RPC 2.0 over HTTP transport.
 * Handles a single MCP message and returns the response.
 */

type Json = null | boolean | number | string | Json[] | { [key: string]: Json };

export type Tool = {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
};

export type ServerOptions = {
  serverName: string;
  serverVersion: string;
  tools: Tool[];
  toolHandlers: Record<string, (args: Record<string, unknown>) => Promise<string> | string>;
};

function ok(id: Json, result: Json): Json {
  return { jsonrpc: "2.0", id, result };
}

function err(id: Json, code: number, message: string): Json {
  return { jsonrpc: "2.0", id, error: { code, message } };
}

export async function handleMessage(message: unknown, options: ServerOptions): Promise<Json> {
  const id = (message as any)?.id ?? null;
  const method = String((message as any)?.method ?? "");
  const params = (message as any)?.params;

  try {
    switch (method) {
      case "initialize":
        return ok(id, {
          protocolVersion: "2024-11-05",
          capabilities: { tools: {} },
          serverInfo: { name: options.serverName, version: options.serverVersion },
        });

      case "ping":
        return ok(id, {});

      case "notifications/initialized":
        return ok(id, null);

      case "tools/list":
        return ok(id, { tools: options.tools });

      case "tools/call": {
        const name = String(params?.name ?? "");
        const handler = options.toolHandlers[name];
        if (!handler) {
          return err(id, -32601, `Unknown tool: ${name}`);
        }
        const text = await handler((params?.arguments ?? {}) as Record<string, unknown>);
        return ok(id, { content: [{ type: "text", text }], isError: false });
      }

      default:
        if (id !== null) {
          return err(id, -32601, `Method not found: ${method}`);
        }
        return ok(null, null);
    }
  } catch (error: unknown) {
    return err(id, -32000, (error as Error).message);
  }
}
