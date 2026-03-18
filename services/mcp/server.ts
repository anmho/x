/**
 * MCP HTTP gateway server.
 *
 * Exposes all internal MCP tools over HTTP with API key authentication.
 * Supports the MCP Streamable HTTP transport (POST /, application/json).
 *
 * Environment variables:
 *   MCP_PORT        HTTP port (default: 8765)
 *   MCP_API_KEYS    Comma-separated API keys (auto-generated if omitted)
 *   MCP_KEYS_FILE   Path to persisted keys JSON (default: ~/.x-mcp/keys.json)
 *
 * Auth:
 *   Authorization: Bearer <key>
 *   X-Api-Key: <key>
 */

import { loadOrGenerateApiKeys, validateRequest } from "./auth.ts";
import { handleMessage } from "./transport.ts";
import { tools, toolHandlers } from "./tools.ts";

const PORT = Number(process.env.MCP_PORT ?? 8765);
const apiKeys = loadOrGenerateApiKeys();

const serverOptions = {
  serverName: "x-mcp",
  serverVersion: "0.1.0",
  tools,
  toolHandlers,
};

const server = Bun.serve({
  port: PORT,

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    // Health check — no auth required
    if (url.pathname === "/health" && request.method === "GET") {
      return Response.json({ status: "ok", tools: tools.map((t) => t.name) });
    }

    // Auth
    if (!validateRequest(request, apiKeys)) {
      return Response.json(
        { error: "Unauthorized — provide Authorization: Bearer <key> or X-Api-Key: <key>" },
        { status: 401 },
      );
    }

    // MCP endpoint — POST /  or  POST /mcp
    if (request.method === "POST" && (url.pathname === "/" || url.pathname === "/mcp")) {
      let body: unknown;
      try {
        body = await request.json();
      } catch {
        return Response.json({ jsonrpc: "2.0", id: null, error: { code: -32700, message: "Invalid JSON" } }, { status: 400 });
      }

      // Batch support: MCP allows arrays of requests
      if (Array.isArray(body)) {
        const responses = await Promise.all(body.map((msg) => handleMessage(msg, serverOptions)));
        return Response.json(responses);
      }

      const response = await handleMessage(body, serverOptions);
      return Response.json(response);
    }

    return Response.json({ error: "Not found" }, { status: 404 });
  },
});

console.log(`[mcp] Server running at http://localhost:${PORT}`);
console.log(`[mcp] POST http://localhost:${PORT}/mcp  — MCP endpoint`);
console.log(`[mcp] GET  http://localhost:${PORT}/health — health check`);
console.log(`[mcp] ${tools.length} tools available: ${tools.map((t) => t.name).join(", ")}`);
