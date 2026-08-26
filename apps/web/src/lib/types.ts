export interface AnalyticsOverview {
  total_payments: number;
  failed_payments: number;
  total_workflows: number;
  recovered_workflows: number;
  recovery_rate: number;
  recovered_revenue: number;
  total_at_risk_revenue: number;
  active_recoveries: number;
  pending_actions: number;
  category_breakdown?: Record<string, number>;
  recent_outcomes?: Array<{
    payment_id: string;
    amount: number;
    currency: string;
    failure_code: string;
    action_type: string;
    status: string;
    recovered: boolean;
    recovered_amount: number;
    created_at: string;
  }>;
}

export interface WorkflowSummary {
  id: string;
  payment_id: string;
  amount: number;
  currency: string;
  customer_id?: string;
  customer_email?: string;
  customer_phone?: string;
  failure_code: string;
  failure_reason?: string;
  status: string;
  recovery_probability: number;
  selected_action: string;
  attempt_count: number;
  recovered: boolean;
  created_at: string;
  updated_at: string;
}

export interface WorkflowsResponse {
  data?: WorkflowSummary[];
  workflows?: WorkflowSummary[];
  total: number;
  limit: number;
  offset: number;
}

export interface WorkflowDetail {
  workflow: {
    id: string;
    payment_id: string;
    status: string;
    current_state: string;
    recovery_probability: number;
    selected_action: string;
    attempt_count: number;
    created_at: string;
    updated_at: string;
  };
  payment: {
    id: string;
    merchant_id: string;
    customer_id: string;
    amount: number;
    currency: string;
    status: string;
    payment_method: string;
    failure_code: string;
    failure_reason: string;
    razorpay_payment_id?: string;
    created_at: string;
  };
  customer?: {
    id: string;
    email: string;
    phone?: string;
    communication_opt_out: boolean;
  };
  policy_decision?: {
    decision: 'ALLOW' | 'BLOCK' | 'ESCALATE' | 'MODIFY';
    reason: string;
    evaluated_at: string;
    thresholds?: {
      max_retries: number;
      confidence_threshold: number;
      amount_threshold: number;
    };
  };
  ai_decisions: Array<{
    id: string;
    provider: string;
    model: string;
    recommended_action: string;
    confidence: number;
    reasoning?: string;
    latency_ms: number;
    created_at: string;
    raw_response?: string;
  }>;
  model_predictions: Array<{
    id: string;
    model_version: string;
    probability: number;
    features: Record<string, any>;
    created_at: string;
  }>;
  recovery_actions: Array<{
    id: string;
    action_type: string;
    status: string;
    attempt: number;
    result?: string;
    created_at: string;
    executed_at?: string;
  }>;
  recovery_outcomes: Array<{
    id: string;
    action_id?: string;
    payment_id: string;
    recovered: boolean;
    recovered_amount: number;
    time_to_recovery_seconds?: number;
    created_at: string;
  }>;
  audit_events: Array<{
    id: string;
    actor: string;
    action: string;
    payload_hash: string;
    previous_event_hash?: string;
    event_hash: string;
    metadata?: Record<string, any>;
    timestamp: string;
  }>;
}

export interface ComponentStatus {
  name: string;
  status: 'HEALTHY' | 'DEGRADED' | 'UNHEALTHY' | 'AVAILABLE' | 'UNAVAILABLE';
  latency_ms: number;
  message?: string;
  details?: Record<string, any>;
}

export interface SystemHealthResponse {
  overall_status: 'HEALTHY' | 'DEGRADED' | 'UNHEALTHY';
  components: ComponentStatus[];
  timestamp: string;
}

export interface QueueStats {
  queue: string;
  size: number;
  memory_usage_bytes: number;
  active: number;
  pending: number;
  scheduled: number;
  retry: number;
  archived: number;
  completed: number;
  paused: boolean;
}

export interface ServerInfo {
  id: string;
  host: string;
  pid: number;
  concurrency: number;
  queues: string[];
  started: string;
  status: string;
}

export interface SystemQueuesResponse {
  redis_status: string;
  queues: QueueStats[];
  servers: ServerInfo[];
  timestamp: string;
}
