import { readFileSync } from "node:fs";
import { resolve } from "node:path";

export type PlatformProject = {
  name: string;
  gcpProjectId?: string;
  gcpRegion?: string;
  gcpServices: string[];
  vercelProjectId?: string;
  vercelDomains: string[];
};

type ControlPlane = {
  projects?: Array<{
    name?: string;
    gcp?: {
      project_id?: string;
      region?: string;
    };
    deployments?: Array<{
      service?: string;
      provider?: string;
    }>;
    domains?: Array<{
      provider?: string;
      project?: string;
      name?: string;
    }>;
  }>;
};

export function loadPlatformProjects(): PlatformProject[] {
  const root = resolve(import.meta.dir, "..", "..");
  const raw = readFileSync(resolve(root, "platform.controlplane.json"), "utf8");
  const parsed = JSON.parse(raw) as ControlPlane;
  return (parsed.projects ?? [])
    .filter((project) => project.name)
    .map((project) => ({
      name: String(project.name),
      gcpProjectId: project.gcp?.project_id,
      gcpRegion: project.gcp?.region,
      gcpServices: (project.deployments ?? [])
        .filter((deployment) => deployment.provider === "gcp-cloud-run" && deployment.service)
        .map((deployment) => String(deployment.service)),
      vercelProjectId: (project.domains ?? []).some((domain) => domain.provider === "vercel")
        ? String(project.name)
        : undefined,
      vercelDomains: (project.domains ?? [])
        .filter((domain) => domain.provider === "vercel" && domain.name)
        .map((domain) => String(domain.name)),
    }));
}

export function getProject(name: string): PlatformProject {
  const project = loadPlatformProjects().find((item) => item.name === name);
  if (!project) {
    throw new Error(`Unknown project '${name}'.`);
  }
  return project;
}
