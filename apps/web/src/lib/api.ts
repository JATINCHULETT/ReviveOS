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

export async function loginUser(email: string, password: string):Promise<{ token: string; user: { id: string; email: string; name: string; role: string; merchant_id?: string } }> {
  return fetchJSON('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export async function getMerchants(): Promise<any[]> {
  return fetchJSON<any[]>('/merchants');
}

export async function createMerchant(data: { name: string; max_retries?: number; confidence_threshold?: number; amount_threshold?: number }): Promise<any> {
  return fetchJSON('/merchants', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function getMerchantDashboard(merchantId?: string): Promise<any> {
  const qs = merchantId ? `?merchant_id=${encodeURIComponent(merchantId)}` : '';
  return fetchJSON(`/merchant/dashboard${qs}`);
}

export async function createSubscription(data: {
  merchant_id: string;
  customer_email: string;
  customer_phone?: string;
  amount: number;
  plan_id?: string;
  billing_interval?: string;
}): Promise<any> {
  return fetchJSON('/merchant/subscriptions', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function createSandboxPaymentLink(data: {
  merchant_id: string;
  customer_email: string;
  customer_phone?: string;
  amount: number;
  description?: string;
  trigger_failure_immediately?: boolean;
  failure_code?: string;
}): Promise<any> {
  return fetchJSON('/merchant/sandbox/payment-link', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

