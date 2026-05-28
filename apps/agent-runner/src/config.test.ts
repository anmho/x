import { describe, expect, test } from "bun:test";

import { defaultMcpConfigPath, emptyToUndefined, loadConfig } from "./config.ts";

describe("config", () => {
  test("emptyToUndefined trims empty strings", () => {
    expect(emptyToUndefined("")).toBeUndefined();
    expect(emptyToUndefined("   ")).toBeUndefined();
    expect(emptyToUndefined("value")).toBe("value");
  });

  test("loadConfig applies defaults", () => {
    const config = loadConfig({});
    expect(config.controlPlaneBaseUrl).toBe("http://localhost:8090");
    expect(config.mcpConfigPath).toBe(defaultMcpConfigPath());
    expect(config.provider).toBe("claude");
  });
});
