import { AnalyticsOverview, WorkflowSummary, WorkflowDetail } from './types';

// Deterministic pseudo-random number generator for reproducible high-fidelity dataset
function createPRNG(seed: number) {
  let s = seed % 2147483647;
  if (s <= 0) s += 2147483646;
  return () => {
    s = (s * 16807) % 2147483647;
    return (s - 1) / 2147483646;
  };
}

const random = createPRNG(42);

const FIRST_NAMES = [
  'Aarav', 'Aditi', 'Rohan', 'Priya', 'Vikram', 'Sneha', 'Ananya', 'Rahul', 'Kavita', 'Siddharth',
  'Meera', 'Arjun', 'Neha', 'Rajesh', 'Pooja', 'Varun', 'Tanvi', 'Nikhil', 'Ishita', 'Gaurav',
  'Divya', 'Karan', 'Shreya', 'Amit', 'Ritu', 'Manish', 'Simran', 'Akash', 'Swati', 'Harsh'
];

const LAST_NAMES = [
  'Sharma', 'Verma', 'Patel', 'Mehta', 'Chopra', 'Gupta', 'Nair', 'Iyer', 'Rao', 'Joshi',
  'Singhania', 'Reddy', 'Deshmukh', 'Kapoor', 'Bhatia', 'Saxena', 'Mishra', 'Banerjee', 'Kulkarni', 'Aggarwal'
];

const DOMAINS = ['revive-os.me', 'revive-os.me', 'gmail.com', 'outlook.com', 'enterprise.in', 'techcorp.io', 'startup.co', 'fintech.ai', 'acme.com'];

const PLAN_AMOUNTS = [499, 999, 1499, 2499, 4999, 9999, 14999, 24999, 49999, 85000];

// Generate full 2,500 synthetic payment workflows
export function generateSyntheticWorkflows(count: number = 2500): WorkflowSummary[] {
  const prng = createPRNG(42);
  const items: WorkflowSummary[] = [];
  const now = new Date('2026-09-01T12:00:00Z');

  for (let i = 0; i < count; i++) {
    const fn = FIRST_NAMES[Math.floor(prng() * FIRST_NAMES.length)];
    const ln = LAST_NAMES[Math.floor(prng() * LAST_NAMES.length)];
    const domain = DOMAINS[Math.floor(prng() * DOMAINS.length)];
    const custId = `cust_${fn.toLowerCase()}_${i + 100}`;
    const email = `${fn.toLowerCase()}.${ln.toLowerCase()}${Math.floor(prng() * 900) + 100}@${domain}`;
    const phone = `+91${Math.floor(9800000000 + prng() * 199999999)}`;

    // Timestamp spanning last 30 days
    const hoursAgo = prng() * 720; // 30 days
    const createdAt = new Date(now.getTime() - hoursAgo * 3600 * 1000);
    const updatedAt = new Date(createdAt.getTime() + 15 * 60 * 1000);

    // Method selection (45% card, 35% upi, 15% netbanking, 5% wallet)
    const mRoll = prng();
    const method = mRoll < 0.45 ? 'card' : mRoll < 0.80 ? 'upi' : mRoll < 0.95 ? 'netbanking' : 'wallet';

    // Amount selection
    const amount = prng() < 0.85
      ? PLAN_AMOUNTS[Math.floor(prng() * 6)] // Standard ₹499 - ₹9,999
      : PLAN_AMOUNTS[6 + Math.floor(prng() * 4)]; // Enterprise ₹14,999 - ₹85,000

    // Failure case distribution:
    // 35% Insufficient funds, 25% Bank downtime, 18% Auth failed, 10% Expired card, 6% Limit, 6% Fraud
    const caseRoll = prng();
    let failureCode = 'INSUFFICIENT_FUNDS';
    let failureReason = 'Insufficient account balance at billing cycle time';
    let recoveryProb = 0.78;
    let selectedAction = 'DELAYED_RETRY';
    let fraudProb = 0.03;
    let fraudLevel = 'LOW';
    let returnProb = 0.02;
    let returnLevel = 'LOW';
    let overallRisk = 'LOW';
    let expectedLoss = 0;
    let riskAction = 'ALLOW';
    let isRecovered = false;

    if (caseRoll < 0.35) {
      // 1. INSUFFICIENT_FUNDS
      failureCode = 'INSUFFICIENT_FUNDS';
      failureReason = 'Customer balance low at billing cycle time; scheduled on payday window';
      recoveryProb = 0.74 + (prng() * 0.12);
      selectedAction = 'DELAYED_RETRY';
      fraudProb = 0.02 + prng() * 0.04;
      returnProb = 0.01 + prng() * 0.03;
      isRecovered = prng() < 0.78;
    } else if (caseRoll < 0.60) {
      // 2. BANK_DOWNTIME
      failureCode = ['BANK_UNAVAILABLE', 'BANK_DOWNTIME', 'GATEWAY_TIMEOUT'][Math.floor(prng() * 3)];
      failureReason = 'Temporary issuing bank gateway downtime; exponential jitter retry executed';
      recoveryProb = 0.91 + (prng() * 0.07);
      selectedAction = 'IMMEDIATE_RETRY';
      fraudProb = 0.01 + prng() * 0.03;
      returnProb = 0.01 + prng() * 0.02;
      isRecovered = prng() < 0.92;
    } else if (caseRoll < 0.78) {
      // 3. AUTH_FAILED
      failureCode = ['AUTHENTICATION_FAILED', '3DS_TIMEOUT', 'OTP_EXPIRED'][Math.floor(prng() * 3)];
      failureReason = 'Customer dropped 3DS authentication; generated interactive smart recovery link';
      recoveryProb = 0.66 + (prng() * 0.12);
      selectedAction = 'PAYMENT_LINK';
      fraudProb = 0.06 + prng() * 0.08;
      returnProb = 0.03 + prng() * 0.05;
      isRecovered = prng() < 0.72;
    } else if (caseRoll < 0.88) {
      // 4. EXPIRED_CARD
      failureCode = ['EXPIRED_CARD', 'MANDATE_LAPSED'][Math.floor(prng() * 2)];
      failureReason = 'Card expired or mandate lapsed; generated secure payment update link';
      recoveryProb = 0.58 + (prng() * 0.12);
      selectedAction = 'UPDATE_PAYMENT_METHOD';
      fraudProb = 0.02 + prng() * 0.04;
      returnProb = 0.02 + prng() * 0.04;
      isRecovered = prng() < 0.64;
    } else if (caseRoll < 0.94) {
      // 5. LIMIT_EXCEEDED
      failureCode = ['LIMIT_EXCEEDED', 'VELOCITY_LIMIT_EXCEEDED'][Math.floor(prng() * 2)];
      failureReason = 'Banking velocity cap reached; dispatched multi-rail payment fallback';
      recoveryProb = 0.72 + (prng() * 0.12);
      selectedAction = 'PAYMENT_LINK';
      fraudProb = 0.05 + prng() * 0.08;
      returnProb = 0.02 + prng() * 0.04;
      isRecovered = prng() < 0.70;
    } else {
      // 6. FRAUD & RETURN RISK
      failureCode = ['SUSPECTED_FRAUD', 'STOLEN_CARD', 'RISK_ANOMALY_BLOCKED'][Math.floor(prng() * 3)];
      failureReason = 'High risk anomaly flagged by Random Forest model; retries permanently halted';
      recoveryProb = 0.08 + (prng() * 0.07);
      selectedAction = 'HALT';
      fraudProb = 0.78 + (prng() * 0.20);
      fraudLevel = 'HIGH';
      returnProb = 0.65 + (prng() * 0.28);
      returnLevel = 'HIGH';
      overallRisk = 'HIGH';
      expectedLoss = amount * fraudProb;
      riskAction = 'BLOCK';
      isRecovered = false;
    }

    // Status assignment
    let status = 'FAILED';
    if (caseRoll >= 0.94) {
      status = 'HALTED';
    } else if (hoursAgo < 2.5 && prng() < 0.35) {
      // In-flight active recovery
      status = ['ANALYZING', 'SCHEDULED', 'EXECUTING', 'VERIFYING'][Math.floor(prng() * 4)];
      isRecovered = false;
    } else if (isRecovered) {
      status = 'RECOVERED';
    } else {
      status = 'FAILED';
    }

    const pmtId = `pay_synth_${(100000 + i).toString()}_${method.substring(0, 2)}`;
    const wfId = `wf_synth_${(100000 + i).toString()}`;

    items.push({
      id: wfId,
      payment_id: pmtId,
      amount: amount,
      currency: 'INR',
      customer_id: custId,
      customer_email: email,
      customer_phone: phone,
      customer_success_count: Math.floor(prng() * 8) + 2,
      customer_failed_count: Math.floor(prng() * 3),
      fraud_probability: parseFloat(fraudProb.toFixed(3)),
      return_probability: parseFloat(returnProb.toFixed(3)),
      overall_risk: overallRisk,
      expected_loss: parseFloat(expectedLoss.toFixed(2)),
      risk_action: riskAction,
      failure_code: failureCode,
      failure_reason: failureReason,
      status: status,
      recovery_probability: parseFloat(recoveryProb.toFixed(3)),
      selected_action: selectedAction,
      attempt_count: isRecovered ? 1 : 2,
      recovered: isRecovered,
      created_at: createdAt.toISOString(),
      updated_at: updatedAt.toISOString(),
    });
  }

  // Prepend escalated human review test cases with @revive-os.me emails
  const testInterventions = getSyntheticInterventions();
  testInterventions.forEach((item, idx) => {
    items.unshift({
      id: item.id,
      payment_id: item.payment_id,
      amount: item.amount,
      currency: item.currency,
      customer_id: item.customer_id,
      customer_email: item.customer_email,
      customer_phone: item.customer_phone,
      customer_success_count: item.customer_success_count,
      customer_failed_count: item.customer_failed_count,
      fraud_probability: item.fraud_probability,
      return_probability: item.return_probability,
      overall_risk: item.overall_risk,
      expected_loss: item.expected_loss,
      risk_action: item.risk_action,
      failure_code: item.failure_code,
      failure_reason: item.escalation_reason,
      status: 'ESCALATED',
      recovery_probability: item.recovery_probability,
      selected_action: item.selected_action,
      attempt_count: 1,
      recovered: false,
      created_at: item.created_at,
      updated_at: item.created_at,
    });
  });

  return items;
}

// Dedicated Synthetic Human Interventions with fake @revive-os.me emails
export function getSyntheticInterventions(): any[] {
  const now = new Date('2026-09-01T14:30:00Z');
  return [
    {
      id: 'wf_esc_revive_101',
      payment_id: 'pay_revive_85000_ent_01',
      amount: 85000,
      currency: 'INR',
      customer_id: 'cust_alex_revive',
      customer_email: 'alex.morgan@revive-os.me',
      customer_phone: '+919876543210',
      customer_success_count: 8,
      customer_failed_count: 1,
      fraud_probability: 0.12,
      return_probability: 0.04,
      overall_risk: 'LOW',
      expected_loss: 10200.0,
      risk_action: 'ALLOW',
      failure_code: 'LIMIT_EXCEEDED',
      escalation_reason: 'High-value Enterprise Tier (₹85,000) threshold exceeded — requires operator authorization',
      selected_action: 'PAYMENT_LINK',
      latest_confidence: 0.92,
      recovery_probability: 0.92,
      communication_opt_out: false,
      created_at: new Date(now.getTime() - 18 * 60 * 1000).toISOString(),
    },
    {
      id: 'wf_esc_revive_102',
      payment_id: 'pay_revive_24999_pro_02',
      amount: 24999,
      currency: 'INR',
      customer_id: 'cust_sarah_revive',
      customer_email: 'sarah.chen@revive-os.me',
      customer_phone: '+919812345678',
      customer_success_count: 4,
      customer_failed_count: 2,
      fraud_probability: 0.78,
      return_probability: 0.55,
      overall_risk: 'HIGH',
      expected_loss: 19499.22,
      risk_action: 'REVIEW',
      failure_code: 'AUTHENTICATION_FAILED',
      escalation_reason: 'ML Risk Guard flagged anomalous velocity spike (78% fraud risk score) from novel subnet',
      selected_action: 'PAYMENT_LINK',
      latest_confidence: 0.65,
      recovery_probability: 0.65,
      communication_opt_out: false,
      created_at: new Date(now.getTime() - 42 * 60 * 1000).toISOString(),
    },
    {
      id: 'wf_esc_revive_103',
      payment_id: 'pay_revive_14999_ann_03',
      amount: 14999,
      currency: 'INR',
      customer_id: 'cust_vikram_revive',
      customer_email: 'vikram.mehta@revive-os.me',
      customer_phone: '+919823456789',
      customer_success_count: 6,
      customer_failed_count: 0,
      fraud_probability: 0.05,
      return_probability: 0.02,
      overall_risk: 'LOW',
      expected_loss: 749.95,
      risk_action: 'ALLOW',
      failure_code: 'MANDATE_LAPSED',
      escalation_reason: 'Customer communication opt-out conflict — automated SMS/WhatsApp halted, manual dispatch required',
      selected_action: 'UPDATE_PAYMENT_METHOD',
      latest_confidence: 0.88,
      recovery_probability: 0.88,
      communication_opt_out: true,
      created_at: new Date(now.getTime() - 75 * 60 * 1000).toISOString(),
    },
    {
      id: 'wf_esc_revive_104',
      payment_id: 'pay_revive_49999_cld_04',
      amount: 49999,
      currency: 'INR',
      customer_id: 'cust_billing_revive',
      customer_email: 'billing.ops@revive-os.me',
      customer_phone: '+919834567890',
      customer_success_count: 12,
      customer_failed_count: 3,
      fraud_probability: 0.08,
      return_probability: 0.03,
      overall_risk: 'LOW',
      expected_loss: 3999.92,
      risk_action: 'ALLOW',
      failure_code: 'INSUFFICIENT_FUNDS',
      escalation_reason: '3 recurring billing cycle failures on mandate — requires manual payment method re-auth',
      selected_action: 'DELAYED_RETRY',
      latest_confidence: 0.84,
      recovery_probability: 0.84,
      communication_opt_out: false,
      created_at: new Date(now.getTime() - 110 * 60 * 1000).toISOString(),
    },
    {
      id: 'wf_esc_revive_105',
      payment_id: 'pay_revive_85000_sec_05',
      amount: 85000,
      currency: 'INR',
      customer_id: 'cust_security_revive',
      customer_email: 'security.team@revive-os.me',
      customer_phone: '+919845678901',
      customer_success_count: 15,
      customer_failed_count: 1,
      fraud_probability: 0.42,
      return_probability: 0.18,
      overall_risk: 'MEDIUM',
      expected_loss: 35700.0,
      risk_action: 'REVIEW',
      failure_code: 'SUSPECTED_FRAUD',
      escalation_reason: 'Card network high-risk decline code (05_DO_NOT_HONOUR) on corporate credit rail',
      selected_action: 'PAYMENT_LINK',
      latest_confidence: 0.71,
      recovery_probability: 0.71,
      communication_opt_out: false,
      created_at: new Date(now.getTime() - 140 * 60 * 1000).toISOString(),
    },
  ];
}

// Cached singleton synthetic dataset
let cachedWorkflows: WorkflowSummary[] | null = null;

export function getSyntheticWorkflows(): WorkflowSummary[] {
  if (!cachedWorkflows) {
    cachedWorkflows = generateSyntheticWorkflows(2500);
  }
  return cachedWorkflows;
}

// Generate Aggregate Analytics Overview matching the 2,500 dataset
export function getSyntheticAnalyticsOverview(): AnalyticsOverview {
  const workflows = getSyntheticWorkflows();
  const totalPayments = workflows.length;
  const failedPayments = workflows.length;
  const recoveredWorkflows = workflows.filter((w) => w.recovered).length;
  const recoveryRate = parseFloat((recoveredWorkflows / totalPayments).toFixed(3));
  
  let recoveredRevenue = 0;
  let totalAtRiskRevenue = 0;
  let activeRecoveries = 0;
  let pendingActions = 0;
  const categoryBreakdown: Record<string, number> = {};

  workflows.forEach((w) => {
    totalAtRiskRevenue += w.amount;
    if (w.recovered) {
      recoveredRevenue += w.amount;
    }
    if (['ANALYZING', 'SCHEDULED', 'EXECUTING', 'VERIFYING'].includes(w.status)) {
      activeRecoveries++;
      pendingActions++;
    }
    const cat = w.failure_code || 'UNKNOWN';
    categoryBreakdown[cat] = (categoryBreakdown[cat] || 0) + 1;
  });

  const recentOutcomes = workflows.slice(0, 25).map((w, idx) => ({
    payment_id: w.payment_id,
    amount: w.amount,
    currency: w.currency,
    failure_code: w.failure_code,
    action_type: w.selected_action,
    status: w.recovered ? 'CAPTURED' : w.status,
    recovered: w.recovered,
    recovered_amount: w.recovered ? w.amount : 0,
    created_at: w.created_at,
  }));

  return {
    total_payments: totalPayments,
    failed_payments: failedPayments,
    total_workflows: totalPayments,
    recovered_workflows: recoveredWorkflows,
    recovery_rate: recoveryRate,
    recovered_revenue: recoveredRevenue,
    total_at_risk_revenue: totalAtRiskRevenue,
    active_recoveries: activeRecoveries,
    pending_actions: pendingActions,
    category_breakdown: categoryBreakdown,
    recent_outcomes: recentOutcomes,
  };
}

// Generate Detailed Workflow Detail with full AI, Risk, and Audit Chains
export function getSyntheticWorkflowDetail(id: string): WorkflowDetail {
  const workflows = getSyntheticWorkflows();
  const wf = workflows.find((w) => w.id === id || w.payment_id === id) || workflows[0];

  const isHalted = wf.status === 'HALTED';
  const isRecovered = wf.recovered;

  return {
    workflow: {
      id: wf.id,
      payment_id: wf.payment_id,
      status: wf.status,
      current_state: wf.status,
      recovery_probability: wf.recovery_probability,
      selected_action: wf.selected_action,
      attempt_count: wf.attempt_count,
      attempts_count: wf.attempt_count,
      created_at: wf.created_at,
      updated_at: wf.updated_at,
      amount: wf.amount,
      currency: wf.currency,
      payment_status: isRecovered ? 'CAPTURED' : 'FAILED',
      payment_method: wf.payment_id.includes('up') ? 'upi' : wf.payment_id.includes('ne') ? 'netbanking' : 'card',
      failure_code: wf.failure_code,
      failure_reason: wf.failure_reason,
      merchant_id: '00000000-0000-0000-0000-000000000001',
      razorpay_payment_id: wf.payment_id,
      customer_id: wf.customer_id,
      customer_email: wf.customer_email,
      customer_phone: wf.customer_phone,
      customer_success_count: wf.customer_success_count || 4,
      customer_failed_count: wf.customer_failed_count || 1,
      fraud_probability: wf.fraud_probability,
      return_probability: wf.return_probability,
      overall_risk: wf.overall_risk,
      expected_loss: wf.expected_loss,
      risk_action: wf.risk_action,
      communication_opt_out: false,
      scheduled_at: wf.selected_action === 'DELAYED_RETRY' ? new Date(new Date(wf.created_at).getTime() + 24 * 3600 * 1000).toISOString() : undefined,
    },
    payment: {
      id: wf.payment_id,
      merchant_id: '00000000-0000-0000-0000-000000000001',
      customer_id: wf.customer_id,
      amount: wf.amount,
      currency: 'INR',
      status: isRecovered ? 'CAPTURED' : 'FAILED',
      payment_method: wf.payment_id.includes('up') ? 'upi' : 'card',
      failure_code: wf.failure_code,
      failure_reason: wf.failure_reason,
      created_at: wf.created_at,
      razorpay_payment_id: wf.payment_id,
    },
    ai_decisions: [
      {
        id: `ai_${wf.id}`,
        provider: 'ollama',
        model: 'deepseek-r1:1.5b',
        recommended_action: wf.selected_action,
        recommended_delay_hours: wf.selected_action === 'DELAYED_RETRY' ? 24 : 0,
        confidence: 0.94,
        reasoning: isHalted
          ? 'High-risk anomaly detected by Random Forest model. Retries permanently halted to prevent merchant chargeback.'
          : `Customer has positive payment history. Empirical probability curve indicates optimal capture via ${wf.selected_action}.`,
        latency_ms: 142,
        created_at: wf.created_at,
      },
    ],
    model_predictions: [
      {
        id: `pred_${wf.id}`,
        model_version: 'logistic-v1',
        probability: wf.recovery_probability,
        features: {
          amount: wf.amount,
          customer_success_count: wf.customer_success_count,
          failure_code: wf.failure_code,
          fraud_score: wf.fraud_probability,
        },
        created_at: wf.created_at,
      },
    ],
    risk_assessments: [
      {
        id: `risk_${wf.id}`,
        payment_id: wf.payment_id,
        workflow_id: wf.id,
        event_type: 'payment.failed',
        fraud_probability: wf.fraud_probability || 0.03,
        fraud_risk_level: (wf.fraud_probability || 0) > 0.5 ? 'HIGH' : 'LOW',
        return_probability: wf.return_probability || 0.02,
        return_risk_level: (wf.return_probability || 0) > 0.5 ? 'HIGH' : 'LOW',
        overall_risk_level: wf.overall_risk || 'LOW',
        expected_loss: wf.expected_loss || 0,
        recommended_action: wf.risk_action || 'ALLOW',
        reason: isHalted ? 'Flagged by Random Forest anomaly detector' : 'Low anomaly score; normal consumer profile',
        model_version: 'fraud-rf-v1.0',
        created_at: wf.created_at,
      },
    ],
    recovery_actions: [
      {
        id: `act_${wf.id}_1`,
        action_type: wf.selected_action,
        status: isHalted ? 'FAILED' : 'EXECUTED',
        attempt: 1,
        executed_at: new Date(new Date(wf.created_at).getTime() + 120 * 1000).toISOString(),
        result: isRecovered ? 'Payment recovered successfully' : isHalted ? 'Halted by ML Fraud Guard' : 'Retry failed',
        created_at: wf.created_at,
      },
    ],
    recovery_outcomes: isRecovered
      ? [
          {
            id: `out_${wf.id}`,
            action_id: `act_${wf.id}_1`,
            payment_id: wf.payment_id,
            recovered: true,
            recovered_amount: wf.amount,
            time_to_recovery: '15m 0s',
            created_at: wf.updated_at,
          },
        ]
      : [],
    audit_events: [
      {
        event_id: `audit_${wf.id}_1`,
        workflow_id: wf.id,
        timestamp: wf.created_at,
        actor: 'GATEWAY_WEBHOOK_INGESTION',
        action: 'PAYMENT_FAILURE_INGESTED',
        payload_hash: '3f8b91a2c4e5d6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1',
        previous_event_hash: '0000000000000000000000000000000000000000000000000000000000000000',
        event_hash: 'a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2',
      },
      {
        event_id: `audit_${wf.id}_2`,
        workflow_id: wf.id,
        timestamp: new Date(new Date(wf.created_at).getTime() + 60 * 1000).toISOString(),
        actor: 'ML_RISK_GUARD',
        action: isHalted ? 'RECOVERY_HALTED_FRAUD_RISK' : 'RISK_EVALUATION_PASSED',
        payload_hash: 'b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3',
        previous_event_hash: 'a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2',
        event_hash: 'c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4',
      },
      {
        event_id: `audit_${wf.id}_3`,
        workflow_id: wf.id,
        timestamp: wf.updated_at,
        actor: 'AUTONOMOUS_EXECUTION_ENGINE',
        action: isRecovered ? 'PAYMENT_RECOVERY_VERIFIED' : isHalted ? 'WORKFLOW_TERMINATED_GUARDRAILS' : 'RECOVERY_EXHAUSTED',
        payload_hash: 'd4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5',
        previous_event_hash: 'c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4',
        event_hash: 'e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6',
      },
    ],
  };
}
