export type CatalogRole = 'main' | 'tooling';
export type CatalogEnvironment = 'production' | 'staging' | 'development';
export type CatalogStatus = 'healthy' | 'degraded';

export type CatalogApplication = {
  id: string;
  name: string;
  environment: CatalogEnvironment;
  status: CatalogStatus;
  apiKeys: number;
};

export type CatalogDeployment = {
  id: string;
  name: string;
  env: CatalogEnvironment;
  region: string;
  status: CatalogStatus;
};

export type CatalogProject = {
  id: string;
  name: string;
  label: string;
  role: CatalogRole;
  description: string;
  environment: CatalogEnvironment;
  grafanaUrl: string;
  posthogUrl: string;
  stripeUrl: string;
  applications: CatalogApplication[];
  deployments: CatalogDeployment[];
};

export type CatalogApplicationRow = CatalogApplication & {
  projectId: string;
  projectName: string;
  projectLabel: string;
};

export type CatalogDeploymentRow = CatalogDeployment & {
  projectId: string;
  projectName: string;
  projectLabel: string;
};

export const CATALOG_PROJECTS: CatalogProject[] = [
  {
    id: 'proj_notifications',
    name: 'notifications',
    label: 'Omnichannel',
    role: 'main',
    description: 'Core notifications control plane, delivery workflows, and built-in template studio.',
    environment: 'production',
    grafanaUrl: 'https://grafana.com/',
    posthogUrl: 'https://posthog.com/',
    stripeUrl: 'https://dashboard.stripe.com/',
    applications: [
      {
        id: 'app_notif_frontend',
        name: 'notifications-frontend',
        environment: 'production',
        status: 'healthy',
        apiKeys: 3,
      },
      {
        id: 'app_notif_api',
        name: 'notifications-api',
        environment: 'production',
        status: 'healthy',
        apiKeys: 5,
      },
      {
        id: 'app_notif_worker',
        name: 'notifications-worker',
        environment: 'staging',
        status: 'degraded',
        apiKeys: 2,
      },
      {
        id: 'app_template_studio',
        name: 'template-studio',
        environment: 'staging',
        status: 'healthy',
        apiKeys: 1,
      },
    ],
    deployments: [
      {
        id: 'dep_notifications_api_prod',
        name: 'notifications-api',
        env: 'production',
        region: 'us-central1',
        status: 'healthy',
      },
      {
        id: 'dep_notifications_worker_prod',
        name: 'notifications-worker',
        env: 'production',
        region: 'us-central1',
        status: 'healthy',
      },
      {
        id: 'dep_notifications_temporal_prod',
        name: 'notifications-temporal',
        env: 'production',
        region: 'us-central1',
        status: 'healthy',
      },
      {
        id: 'dep_template_studio_staging',
        name: 'template-studio',
        env: 'staging',
        region: 'us-central1',
        status: 'healthy',
      },
    ],
  },
];

export const CATALOG_APPLICATIONS: CatalogApplicationRow[] = CATALOG_PROJECTS.flatMap((project) =>
  project.applications.map((application) => ({
    ...application,
    projectId: project.id,
    projectName: project.name,
    projectLabel: project.label,
  })),
);

export const CATALOG_DEPLOYMENTS: CatalogDeploymentRow[] = CATALOG_PROJECTS.flatMap((project) =>
  project.deployments.map((deployment) => ({
    ...deployment,
    projectId: project.id,
    projectName: project.name,
    projectLabel: project.label,
  })),
);

export function catalogProjectCounts(projectId: string): { applications: number; deployments: number } {
  const project = CATALOG_PROJECTS.find((entry) => entry.id === projectId);
  if (!project) {
    return { applications: 0, deployments: 0 };
  }
  return {
    applications: project.applications.length,
    deployments: project.deployments.length,
  };
}
