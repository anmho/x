import axios from 'axios';
import { apiClient } from '@/lib/api';

export type WorkflowListItem = {
  workflowId: string;
  runId: string;
  workflowType: string;
  status: string;
  startTime?: string;
  closeTime?: string;
  taskQueue: string;
  namespace: string;
  projectId: string;
  environment: string;
  temporalUiUrl: string;
};

export type ListWorkflowsParams = {
  projectId: string;
  environment: string;
  statuses?: string[];
  query?: string;
  pageSize?: number;
  pageToken?: string;
  startTime?: string;
  endTime?: string;
};

function getControlPlaneBaseUrl() {
  const configured = process.env.NEXT_PUBLIC_CONTROL_PLANE_URL;
  if (configured && configured.trim()) {
    return configured.replace(/\/$/, '');
  }
  const apiBase = apiClient.defaults.baseURL || '';
  return apiBase.replace(/\/api\/v1\/?$/, '');
}

function getApiKey() {
  return String(apiClient.defaults.headers.common['X-API-Key'] || process.env.NEXT_PUBLIC_API_KEY || '');
}

function normalizeItem(item: Record<string, unknown>): WorkflowListItem {
  return {
    workflowId: String(item.workflowId ?? item.workflow_id ?? ''),
    runId: String(item.runId ?? item.run_id ?? ''),
    workflowType: String(item.workflowType ?? item.workflow_type ?? ''),
    status: String(item.status ?? ''),
    startTime: (item.startTime ?? item.start_time) as string | undefined,
    closeTime: (item.closeTime ?? item.close_time) as string | undefined,
    taskQueue: String(item.taskQueue ?? item.task_queue ?? ''),
    namespace: String(item.namespace ?? ''),
    projectId: String(item.projectId ?? item.project_id ?? ''),
    environment: String(item.environment ?? ''),
    temporalUiUrl: String(item.temporalUiUrl ?? item.temporal_ui_url ?? ''),
  };
}

export async function listTemporalWorkflows(params: ListWorkflowsParams): Promise<{ items: WorkflowListItem[]; nextPageToken?: string }> {
  const baseURL = getControlPlaneBaseUrl();
  const url = `${baseURL}/temporal.v1.TemporalService/ListWorkflows`;

  const response = await axios.post(
    url,
    {
      projectId: params.projectId,
      environment: params.environment,
      statuses: params.statuses || [],
      query: params.query || '',
      pageSize: params.pageSize || 50,
      pageToken: params.pageToken || '',
      startTime: params.startTime,
      endTime: params.endTime,
    },
    {
      headers: {
        'Content-Type': 'application/json',
        'Connect-Protocol-Version': '1',
        'X-API-Key': getApiKey(),
      },
    },
  );

  const payload = response.data as {
    items?: Record<string, unknown>[];
    nextPageToken?: string;
    next_page_token?: string;
  };

  return {
    items: (payload.items || []).map(normalizeItem),
    nextPageToken: payload.nextPageToken || payload.next_page_token,
  };
}
