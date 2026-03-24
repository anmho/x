import { describe, expect, test } from "bun:test";

import { materializePrompt } from "./prompt.ts";

describe("materializePrompt", () => {
  test("includes resources before the canonical message", () => {
    const prompt = materializePrompt({
      id: "run-123",
      status: "PENDING",
      message: "Implement the requested workflow.",
      resources: [{ uri: "https://linear.app/anmho/issue/ANM-194", title: "Linear issue", text: "Preserve ConnectRPC." }],
    });

    expect(prompt).toContain("Resources:");
    expect(prompt).toContain("Preserve ConnectRPC.");
    expect(prompt).toContain("Implement the requested workflow.");
  });
});
