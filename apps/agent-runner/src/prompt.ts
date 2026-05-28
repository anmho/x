import type { RunEnvelope } from "./types.ts";

export function materializePrompt(run: RunEnvelope): string {
  const sections: string[] = [];
  if (run.resources && run.resources.length > 0) {
    sections.push(renderResources(run.resources));
  }
  if (run.message.trim()) {
    sections.push(run.message.trim());
  }
  return sections.filter(Boolean).join("\n\n").trim();
}

export function renderResources(resources: NonNullable<RunEnvelope["resources"]>): string {
  const lines: string[] = ["Resources:"];
  for (const resource of resources) {
    const label = [resource.title, resource.uri].filter(Boolean).join(" - ");
    lines.push(`- ${label}`);
    if (resource.mimeType) {
      lines.push(`  mime: ${resource.mimeType}`);
    }
    if (resource.text) {
      lines.push(`  text: ${resource.text}`);
    }
  }
  return lines.join("\n");
}
