/**
 * API key management for the MCP HTTP server.
 *
 * Key resolution order:
 *   1. MCP_API_KEYS env var (comma-separated list)
 *   2. Keys file at MCP_KEYS_FILE or ~/.x-mcp/keys.json
 *   3. Auto-generate a key, persist it, and print it once to stdout
 */

import { randomBytes } from "node:crypto";
import { readFileSync, writeFileSync, existsSync, mkdirSync } from "node:fs";
import { join, dirname } from "node:path";

const keysFile = process.env.MCP_KEYS_FILE ?? join(process.env.HOME ?? "/tmp", ".x-mcp", "keys.json");

export function loadOrGenerateApiKeys(): string[] {
  // 1. env var
  const fromEnv = (process.env.MCP_API_KEYS ?? "")
    .split(",")
    .map((k) => k.trim())
    .filter(Boolean);
  if (fromEnv.length > 0) {
    return fromEnv;
  }

  // 2. persisted file
  if (existsSync(keysFile)) {
    try {
      const data = JSON.parse(readFileSync(keysFile, "utf8")) as { keys?: string[] };
      if (Array.isArray(data.keys) && data.keys.length > 0) {
        return data.keys;
      }
    } catch {
      // corrupt file — fall through to regenerate
    }
  }

  // 3. generate + persist
  const key = `mcp_${randomBytes(24).toString("hex")}`;
  const dir = dirname(keysFile);
  mkdirSync(dir, { recursive: true });
  writeFileSync(keysFile, JSON.stringify({ keys: [key] }, null, 2), "utf8");
  console.log(`\n[mcp] No API key configured. Generated one for you:`);
  console.log(`[mcp]   ${key}`);
  console.log(`[mcp] Saved to ${keysFile}`);
  console.log(`[mcp] Set MCP_API_KEYS=${key} to use it persistently.\n`);
  return [key];
}

export function validateRequest(request: Request, apiKeys: string[]): boolean {
  if (apiKeys.length === 0) return true;

  const auth = request.headers.get("authorization");
  if (auth?.startsWith("Bearer ")) {
    return apiKeys.includes(auth.slice(7));
  }

  const apiKey = request.headers.get("x-api-key");
  if (apiKey) {
    return apiKeys.includes(apiKey);
  }

  return false;
}
