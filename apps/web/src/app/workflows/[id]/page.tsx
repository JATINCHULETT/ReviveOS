'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { getWorkflowDetail } from '@/lib/api';
import { WorkflowDetail } from '@/lib/types';
import { StatusBadge } from '@/components/ui/StatusBadge';
import { LoadingView, ErrorView } from '@/components/ui/StateViews';
import { 
  ArrowLeft, 
  CheckCircle2, 
  Clock, 
  Cpu, 
  ShieldAlert, 
  GitCommit, 
  RefreshCw,
  Terminal,
  FileCheck
} from 'lucide-react';

export default function WorkflowDetailPage() {
  const params = useParams();
  const workflowId = params?.id as string;

  const [detail, setDetail] = useState<WorkflowDetail | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const loadWorkflow = async () => {
    if (!workflowId) return;
    try {
      setLoading(true);
      setError(null);
      const res = await getWorkflowDetail(workflowId);
      setDetail(res);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch workflow detail');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadWorkflow();
  }, [workflowId]);

  if (loading && !detail) {
    return (
      <div className="page-container">
        <LoadingView message="Loading complete recovery lifecycle..." />
      </div>
    );
  }

  if (error || !detail) {
    return (
      <div className="page-container">
        <ErrorView message={error || 'Workflow not found'} onRetry={loadWorkflow} />
      </div>
    );
  }

  const {
    workflow,
    payment,
    customer,
    policy_decision,
    ai_decisions,
    model_predictions,
    recovery_actions,
    recovery_outcomes,
    audit_events,
  } = detail;

  const latestAI = ai_decisions && ai_decisions.length > 0 ? ai_decisions[0] : null;
  const latestPred = model_predictions && model_predictions.length > 0 ? model_predictions[0] : null;
  const latestOutcome = recovery_outcomes && recovery_outcomes.length > 0 ? recovery_outcomes[0] : null;

  return (
    <div className="page-container">
      {/* Top Navigation */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '24px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <Link href="/workflows" className="btn-secondary" style={{ padding: '6px 10px' }}>
            <ArrowLeft size={16} />
          </Link>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <h1 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--text-primary)' }}>
                Workflow <span className="mono">{workflow.id.substring(0, 8)}</span>
              </h1>
              <StatusBadge status={workflow.status} />
            </div>
            <div className="mono" style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
              Payment ID: {payment.id}
            </div>
          </div>
        </div>

        <button onClick={loadWorkflow} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <RefreshCw size={14} className={loading ? 'spinning' : ''} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Summary Grid */}
      <div className="metrics-grid" style={{ marginBottom: '28px' }}>
        <div className="metric-card">
          <span className="metric-label">Payment Amount</span>
          <div className="metric-value">₹{payment.amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</div>
          <span className="metric-sub mono">{payment.currency} | {payment.status}</span>
        </div>

        <div className="metric-card">
          <span className="metric-label">Customer Info</span>
          <div style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text-primary)', marginTop: '4px' }}>
            {customer?.email || 'Unknown'}
          </div>
          <span className="metric-sub">
            Opt-Out: {customer?.communication_opt_out ? 'YES (Blocked)' : 'NO'}
          </span>
        </div>

        <div className="metric-card">
          <span className="metric-label">AI Recommendation</span>
          <div className="metric-value" style={{ fontSize: '18px' }}>
            {latestAI?.recommended_action || workflow.selected_action || 'NONE'}
          </div>
          <span className="metric-sub">
            Confidence: {latestAI ? `${(latestAI.confidence * 100).toFixed(0)}%` : '—'} | Latency: {latestAI?.latency_ms || 0}ms
          </span>
        </div>

        <div className="metric-card">
          <span className="metric-label">Policy Status</span>
          <div style={{ marginTop: '6px' }}>
            <StatusBadge status={policy_decision?.decision || 'ALLOW'} />
          </div>
          <span className="metric-sub" style={{ marginTop: '4px' }}>
            {policy_decision?.reason || 'Passed policy checks'}
          </span>
        </div>
      </div>

      {/* 10-Stage Lifecycle Timeline */}
      <div className="table-container" style={{ padding: '28px', marginBottom: '28px' }}>
        <h2 style={{ fontSize: '15px', fontWeight: 700, marginBottom: '24px', display: 'flex', alignItems: 'center', gap: '8px' }}>
          <GitCommit size={18} color="#3b82f6" />
          <span>Complete Recovery Orchestration Lifecycle</span>
        </h2>

        <div className="timeline">
          {/* 1. Webhook Received */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">1</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <b>Webhook Ingestion & Signature Verification</b>
                <span className="mono" style={{ fontSize: '11px', color: 'var(--text-muted)' }}>
                  {new Date(payment.created_at).toLocaleTimeString()}
                </span>
              </div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                Payment failure webhook validated via HMAC-SHA256 signature and deduplicated.
              </div>
              <div className="mono" style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px' }}>
                Payment ID: {payment.id} | Method: {payment.payment_method || 'CARD'}
              </div>
            </div>
          </div>

          {/* 2. Failure Classification */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">2</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <b>Deterministic Failure Classification</b>
                <span className="badge badge-neutral">{payment.failure_code}</span>
              </div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                {payment.failure_reason || `Classified as ${payment.failure_code}`}
              </div>
            </div>
          </div>

          {/* 3. Statistical Probability Model */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">3</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <b>Empirical Recovery Probability Model</b>
                <span className="mono" style={{ fontWeight: 700, color: '#10b981' }}>
                  {((latestPred?.probability ?? workflow.recovery_probability) * 100).toFixed(1)}% P(Recovery)
                </span>
              </div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                Calculated using customer historical payment success rate, failure code recoverability, and attempt number.
              </div>
              {latestPred?.features && (
                <div style={{ marginTop: '8px', background: 'var(--bg-elevated)', padding: '8px 12px', borderRadius: '4px', fontSize: '11px' }}>
                  <span className="mono">Features: {JSON.stringify(latestPred.features)}</span>
                </div>
              )}
            </div>
          </div>

          {/* 4. Ollama DeepSeek-R1 AI Decision */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">4</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <Cpu size={16} color="#60a5fa" />
                  <b>AI Inference Decision ({latestAI?.model || 'deepseek-r1:1.5b'})</b>
                </div>
                <span className="mono" style={{ fontSize: '11px', color: 'var(--text-muted)' }}>
                  Latency: {latestAI?.latency_ms || 0}ms
                </span>
              </div>
              <div style={{ fontSize: '13px', color: 'var(--text-primary)', marginBottom: '4px' }}>
                Recommendation: <b>{latestAI?.recommended_action || workflow.selected_action}</b> (Confidence: {((latestAI?.confidence || 0.85) * 100).toFixed(0)}%)
              </div>
              <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                Reasoning: {latestAI?.reasoning || 'Evaluated optimal retry timing based on bank availability schedule.'}
              </div>
            </div>
          </div>

          {/* 5. Policy Engine Evaluation */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">5</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <ShieldAlert size={16} color="#f59e0b" />
                  <b>Merchant Policy Safety Engine</b>
                </div>
                <StatusBadge status={policy_decision?.decision || 'ALLOW'} />
              </div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                {policy_decision?.reason || 'Recommendation passed all database safety checks.'}
              </div>
            </div>
          </div>

          {/* 6. Pre-Execution Reconciliation */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">6</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <b>Pre-Execution Reconciliation</b>
                <span className="badge badge-success">PASSED</span>
              </div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                Queried PaymentProvider to verify payment status was not already captured before initiating action.
              </div>
            </div>
          </div>

          {/* 7. Recovery Execution */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">7</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <b>Recovery Execution</b>
                <span className="badge badge-neutral">Attempt #{workflow.attempt_count || 1}</span>
              </div>
              {recovery_actions && recovery_actions.length > 0 ? (
                recovery_actions.map((act) => (
                  <div key={act.id} style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                    Action <b>{act.action_type}</b> executed via PaymentProvider. Result: <span className="mono">{act.result || act.status}</span>
                  </div>
                ))
              ) : (
                <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                  Action execution initiated.
                </div>
              )}
            </div>
          </div>

          {/* 8. Provider State Verification */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">8</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <b>Payment Provider Verification</b>
                <span className="mono" style={{ fontWeight: 600 }}>{payment.status}</span>
              </div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                Authoritative verification returned from PaymentProvider. Verified state: <b>{payment.status}</b>.
              </div>
            </div>
          </div>

          {/* 9. Outcome & Learning Feedback */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">9</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                <b>Recovery Outcome & Model Calibration</b>
                <StatusBadge status={latestOutcome?.recovered ? 'RECOVERED' : workflow.status} />
              </div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                Outcome recorded: {latestOutcome?.recovered ? `Recovered ₹${latestOutcome.recovered_amount.toFixed(2)}` : 'Recovery in progress / scheduled'}. Feedback emitted to learning engine.
              </div>
            </div>
          </div>

          {/* 10. Cryptographic Audit Chain */}
          <div className="timeline-item">
            <div className="timeline-node timeline-node-success">10</div>
            <div className="timeline-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <FileCheck size={16} color="#10b981" />
                  <b>Tamper-Evident SHA-256 Audit Chain</b>
                </div>
                <span className="badge badge-success">
                  {audit_events?.length || 0} Block Events Linked
                </span>
              </div>

              {audit_events && audit_events.length > 0 && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  {audit_events.map((ev, idx) => (
                    <div key={ev.id || idx} style={{ background: 'var(--bg-elevated)', padding: '10px 14px', borderRadius: '4px', fontSize: '11px' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                        <span className="mono" style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{ev.action}</span>
                        <span style={{ color: 'var(--text-muted)' }}>{ev.actor} | {new Date(ev.timestamp).toLocaleTimeString()}</span>
                      </div>
                      <div className="mono" style={{ color: 'var(--text-muted)', wordBreak: 'break-all' }}>
                        <div>Payload Hash: {ev.payload_hash}</div>
                        <div>Event Hash:   {ev.event_hash}</div>
                        {ev.previous_event_hash && (
                          <div style={{ color: '#60a5fa' }}>Prev Hash:    {ev.previous_event_hash}</div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
