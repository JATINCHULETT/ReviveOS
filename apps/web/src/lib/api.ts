import {
  AnalyticsOverview,
  WorkflowsResponse,
  WorkflowDetail,
  SystemHealthResponse,
  SystemQueuesResponse,
} from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function fetchJSON<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
  const url = `${API_BASE}${cleanEndpoint}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    cache: 'no-store',
  });

  if (!res.ok) {
    let errorDetail = '';
    try {
      const clonedRes = res.clone();
      const errJson = await clonedRes.json();
      errorDetail = errJson.message || errJson.error || JSON.stringify(errJson);
    } catch {
      try {
        errorDetail = await res.text();
      } catch {
        errorDetail = res.statusText;
      }
    }
    throw new Error(`API error (${res.status}): ${errorDetail || res.statusText}`);
  }

  return res.json();
}

export async function getAnalyticsOverview(): Promise<AnalyticsOverview> {
  return fetchJSON<AnalyticsOverview>('/analytics/overview');
}

export async function getWorkflows(params?: {
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<WorkflowsResponse> {
  const query = new URLSearchParams();
  if (params?.status && params.status !== 'ALL') query.set('status', params.status);
  if (params?.limit) query.set('limit', params.limit.toString());
  if (params?.offset) query.set('offset', params.offset.toString());

  const qs = query.toString();
  return fetchJSON<WorkflowsResponse>(`/workflows${qs ? `?${qs}` : ''}`);
}

export async function getWorkflowDetail(id: string): Promise<WorkflowDetail> {
  return fetchJSON<WorkflowDetail>(`/workflows/${id}`);
}

export async function getSystemHealth(): Promise<SystemHealthResponse> {
  return fetchJSON<SystemHealthResponse>('/system/health');
}

export async function getSystemQueues(): Promise<SystemQueuesResponse> {
  return fetchJSON<SystemQueuesResponse>('/system/queues');
}
