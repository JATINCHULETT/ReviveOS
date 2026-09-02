'use client';

import React, { useState, useEffect } from 'react';
import { PhoneCall, Play, Mic, CheckCircle2, UserCheck, MessageSquare, Volume2, Shield, Search } from 'lucide-react';
import { BadgePulse } from '@/components/ui/AnimatedComponents';
import { getWorkflows } from '@/lib/api';
import { getSyntheticWorkflows } from '@/lib/syntheticDataset';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const DEMO_VOICE_LOGS = [
  {
    call_sid: 'call_twilio_9021831',
    provider: 'twilio',
    duration_seconds: 45,
    customer_spoken: 'Haan bhai kal subah 11 baje tak payment kar dunga.',
    intent: 'PROMISE_TO_PAY',
    ptp_date: 'tomorrow',
  },
  {
    call_sid: 'call_twilio_4810924',
    provider: 'twilio',
    duration_seconds: 32,
    customer_spoken: 'Aap mujhe WhatsApp pe UPI link bhej dijiye abhi kar deti hoon.',
    intent: 'REQUEST_LINK',
    ptp_date: null,
  },
  {
    call_sid: 'call_twilio_3310892',
    provider: 'twilio',
    duration_seconds: 51,
    customer_spoken: 'Salary aane me 2 din lagenge, tab karta hoon.',
    intent: 'PROMISE_TO_PAY',
    ptp_date: 'in 2 days',
  },
];

export default function VoiceRecoveryPage() {
  const [calls, setCalls] = useState<any[]>(DEMO_VOICE_LOGS);
  const [preview, setPreview] = useState<any>({
    hinglish_script: 'Namaste Rajesh ji! Main ReviveOS Payments Desk se baat kar raha hoon. Aapka ₹14,999 ka payment due tha...',
  });
  const [calling, setCalling] = useState(false);
  const [callStatus, setCallStatus] = useState<string | null>(null);

  // Custom Dial Modal State
  const [showDialModal, setShowDialModal] = useState(false);
  const [availableWorkflows, setAvailableWorkflows] = useState<any[]>(() => getSyntheticWorkflows().slice(0, 50));
  const [workflowSearch, setWorkflowSearch] = useState('');
  const [customWorkflowId, setCustomWorkflowId] = useState('');
  const [customName, setCustomName] = useState('Rajesh Sharma');
  const [customEmail, setCustomEmail] = useState('rajesh.sharma@example.com');
  const [customPhone, setCustomPhone] = useState('+919876543210');
  const [customAmount, setCustomAmount] = useState('14999');

  const loadWorkflowsList = async () => {
    try {
      const res = await getWorkflows({ limit: 100 });
      if (res?.data && res.data.length > 0) {
        setAvailableWorkflows(res.data);
      }
    } catch {
      setAvailableWorkflows(getSyntheticWorkflows().slice(0, 50));
    }
  };

  const handleSelectWorkflow = (wf: any) => {
    setCustomWorkflowId(wf.id || wf.payment_id);
    if (wf.customer_name) setCustomName(wf.customer_name);
    else if (wf.customer_email) {
      const prefix = wf.customer_email.split('@')[0].replace(/[._0-9]/g, ' ');
      setCustomName(prefix.charAt(0).toUpperCase() + prefix.slice(1));
    }
    if (wf.customer_email) setCustomEmail(wf.customer_email);
    if (wf.customer_phone) setCustomPhone(wf.customer_phone);
    if (wf.amount) setCustomAmount(String(wf.amount));
  };

  const fetchVoiceData = async () => {
    try {
      const [callsRes, previewRes] = await Promise.all([
        fetch(`${API_BASE}/v1/voice`),
        fetch(`${API_BASE}/v1/voice/scripts/preview`),
      ]);
      if (callsRes.ok) {
        const d = await callsRes.json();
        const realCalls = d.calls || [];
        // Merge real calls on top of demo calls, deduplicating by call_sid
        const realSids = new Set(realCalls.map((c: any) => c.call_sid));
        const merged = [...realCalls, ...DEMO_VOICE_LOGS.filter((demo) => !realSids.has(demo.call_sid))];
        setCalls(merged);
      }
      if (previewRes.ok) {
        const p = await previewRes.json();
        setPreview(p);
      }
    } catch (err) {
      console.warn('Backend port 8080 not reachable, using offline demo voice calls:', err);
      setCalls(DEMO_VOICE_LOGS);
    }
  };

  useEffect(() => {
    fetchVoiceData();
    loadWorkflowsList();
  }, []);

  const triggerCallSimulation = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setCalling(true);
    setShowDialModal(false);
    setCallStatus(`Initiating live Hinglish call via Twilio to ${customPhone} (${customName})...`);
    try {
      const res = await fetch(`${API_BASE}/v1/voice`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          workflow_id: customWorkflowId.trim() || undefined,
          customer_name: customName || 'Valued Customer',
          customer_email: customEmail,
          phone: customPhone,
          amount: parseFloat(customAmount) || 14999,
          currency: 'INR',
        }),
      });
      if (res.ok) {
        const data = await res.json();
        setCallStatus(`Twilio Call (${data.call_result?.call_sid}) dispatched to ${customPhone} with intent: ${data.call_result?.intent || 'PROMISE_TO_PAY'}`);
        await fetchVoiceData();
      } else {
        const errData = await res.json().catch(() => ({}));
        setCallStatus(`Twilio Call failed: ${errData.detail || errData.error || 'Server error'}`);
      }
    } catch (err: any) {
      setCallStatus(`Call error: ${err.message}`);
    } finally {
      setCalling(false);
    }
  };

  return (
    <div className="main-container" style={{ padding: '24px 32px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '28px', flexWrap: 'wrap', gap: '16px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '6px' }}>
            <h1 style={{ fontSize: '26px', fontWeight: 800, letterSpacing: '-0.5px' }}>
              Hinglish AI Voice Recovery Engine
            </h1>
            <BadgePulse text="Twilio Active" variant="success" />
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '14px' }}>
            Culturally attuned vernacular Hinglish voice calling with real-time intent classification powered by Twilio Telephony.
          </p>
        </div>

        <button
          onClick={() => setShowDialModal(true)}
          disabled={calling}
          style={{
            padding: '10px 20px',
            background: 'var(--primary)',
            color: '#fff',
            border: 'none',
            borderRadius: '8px',
            fontWeight: 700,
            fontSize: '13.5px',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          <PhoneCall size={16} />
          {calling ? 'Calling Target...' : 'Dial Custom Number (Hinglish)'}
        </button>
      </div>

      {/* Interactive Dial Modal */}
      {showDialModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(0, 0, 0, 0.75)',
          backdropFilter: 'blur(6px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 9999,
          padding: '20px',
        }}>
          <div style={{
            background: 'var(--bg-secondary)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '16px',
            width: '100%',
            maxWidth: '460px',
            padding: '28px',
            boxShadow: '0 20px 40px rgba(0,0,0,0.5)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                <PhoneCall size={20} color="#3b82f6" />
                <h3 style={{ fontSize: '18px', fontWeight: 800, margin: 0 }}>Trigger Live Hinglish Call</h3>
              </div>
              <button
                onClick={() => setShowDialModal(false)}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', fontSize: '20px', cursor: 'pointer' }}
              >
                &times;
              </button>
            </div>

            <form onSubmit={triggerCallSimulation} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              {/* Workflow Search & Dropdown Selector */}
              <div>
                <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Select Workflow (Autofills Customer Email & Amount)
                </label>
                <div style={{ position: 'relative', marginBottom: '8px' }}>
                  <Search size={15} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
                  <input
                    type="text"
                    value={workflowSearch}
                    onChange={(e) => setWorkflowSearch(e.target.value)}
                    placeholder="Search workflows by email, ID, or amount..."
                    style={{
                      width: '100%',
                      padding: '8px 12px 8px 34px',
                      borderRadius: '8px',
                      background: 'var(--bg-tertiary)',
                      border: '1px solid var(--border-subtle)',
                      color: 'var(--text-primary)',
                      fontSize: '12.5px',
                    }}
                  />
                </div>

                <select
                  value={customWorkflowId}
                  onChange={(e) => {
                    const selected = availableWorkflows.find((w: any) => (w.id || w.payment_id) === e.target.value);
                    if (selected) {
                      handleSelectWorkflow(selected);
                    } else {
                      setCustomWorkflowId(e.target.value);
                    }
                  }}
                  style={{
                    width: '100%',
                    padding: '10px 14px',
                    borderRadius: '8px',
                    background: 'var(--bg-tertiary)',
                    border: '1px solid var(--border-subtle)',
                    color: 'var(--text-primary)',
                    fontSize: '13px',
                    cursor: 'pointer',
                  }}
                >
                  <option value="">-- Choose from Active Workflows --</option>
                  {availableWorkflows
                    .filter((wf: any) => {
                      if (!workflowSearch) return true;
                      const q = workflowSearch.toLowerCase();
                      const email = (wf.customer_email || '').toLowerCase();
                      const id = (wf.id || wf.payment_id || '').toLowerCase();
                      const amt = String(wf.amount || '');
                      return email.includes(q) || id.includes(q) || amt.includes(q);
                    })
                    .slice(0, 30)
                    .map((wf: any) => (
                      <option key={wf.id || wf.payment_id} value={wf.id || wf.payment_id}>
                        {wf.customer_email || 'User'} — ₹{wf.amount?.toLocaleString('en-IN') || 0} ({wf.payment_id?.substring(0, 15)}...)
                      </option>
                    ))}
                </select>

                {/* Workflow Summary Badge */}
                {customWorkflowId && (
                  <div style={{
                    marginTop: '8px',
                    padding: '8px 12px',
                    borderRadius: '6px',
                    background: 'rgba(59, 130, 246, 0.08)',
                    border: '1px solid rgba(59, 130, 246, 0.2)',
                    fontSize: '12px',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                  }}>
                    <span style={{ color: '#60a5fa', fontWeight: 600 }}>
                      Linked: {customWorkflowId.substring(0, 18)}...
                    </span>
                    <span style={{ color: 'var(--color-emerald)', fontWeight: 700 }}>
                      ₹{parseFloat(customAmount || '0').toLocaleString('en-IN')} Due
                    </span>
                  </div>
                )}
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Target Customer Name
                </label>
                <input
                  type="text"
                  required
                  value={customName}
                  onChange={(e) => setCustomName(e.target.value)}
                  placeholder="e.g. Rajesh Sharma"
                  style={{
                    width: '100%',
                    padding: '10px 14px',
                    borderRadius: '8px',
                    background: 'var(--bg-tertiary)',
                    border: '1px solid var(--border-subtle)',
                    color: 'var(--text-primary)',
                    fontSize: '13.5px',
                  }}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                  Customer Mobile Number (E.164 format)
                </label>
                <input
                  type="tel"
                  required
                  value={customPhone}
                  onChange={(e) => setCustomPhone(e.target.value)}
                  placeholder="+919876543210"
                  style={{
                    width: '100%',
                    padding: '10px 14px',
                    borderRadius: '8px',
                    background: 'var(--bg-tertiary)',
                    border: '1px solid var(--border-subtle)',
                    color: 'var(--text-primary)',
                    fontSize: '13.5px',
                  }}
                />
                <span style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px', display: 'block' }}>
                  Twilio will initiate an outbound call and speak vernacular Hinglish script upon answer.
                </span>
              </div>

              <div style={{ display: 'none' }}>
                {/* Hidden auto-fetched email and amount stored in payload */}
                <input type="hidden" value={customEmail} />
                <input type="hidden" value={customAmount} />
              </div>

              <div style={{ display: 'flex', gap: '10px', marginTop: '10px' }}>
                <button
                  type="button"
                  onClick={() => setShowDialModal(false)}
                  style={{
                    flex: 1,
                    padding: '10px',
                    borderRadius: '8px',
                    background: 'var(--bg-tertiary)',
                    border: '1px solid var(--border-subtle)',
                    color: 'var(--text-secondary)',
                    fontWeight: 600,
                    fontSize: '13px',
                    cursor: 'pointer',
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={calling}
                  style={{
                    flex: 2,
                    padding: '10px',
                    borderRadius: '8px',
                    background: 'var(--primary)',
                    border: 'none',
                    color: '#fff',
                    fontWeight: 700,
                    fontSize: '13px',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '6px',
                  }}
                >
                  <PhoneCall size={14} />
                  <span>{calling ? 'Dialing...' : 'Place Hinglish Call'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {callStatus && (
        <div style={{
          padding: '14px 18px',
          background: 'rgba(59, 130, 246, 0.12)',
          border: '1px solid rgba(59, 130, 246, 0.3)',
          borderRadius: '10px',
          marginBottom: '24px',
          color: '#60a5fa',
          fontSize: '14px',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
        }}>
          <CheckCircle2 size={18} />
          <span>{callStatus}</span>
        </div>
      )}

      {/* Telephony Script & Intent Preview Box */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(360px, 1fr))', gap: '20px', marginBottom: '32px' }}>
        <div style={{ background: 'var(--bg-secondary)', borderRadius: '14px', border: '1px solid var(--border-subtle)', padding: '22px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '14px' }}>
            <Volume2 size={18} color="#3b82f6" />
            <h3 style={{ fontSize: '16px', fontWeight: 700 }}>Dynamic Hinglish Prompt Preview</h3>
          </div>
          <p style={{
            background: 'var(--bg-tertiary)',
            padding: '16px',
            borderRadius: '10px',
            border: '1px solid var(--border-subtle)',
            fontSize: '13.5px',
            lineHeight: '1.6',
            color: 'var(--text-primary)',
            fontStyle: 'italic',
          }}>
            "{preview?.hinglish_script || 'Namaste Rajesh ji! Main ReviveOS Payments Desk se baat kar raha hoon...'}"
          </p>
          <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '12px' }}>
            Tone: Empathetic & Respectful | Latency: ~450ms | Language: Indian Hinglish
          </div>
        </div>

        <div style={{ background: 'var(--bg-secondary)', borderRadius: '14px', border: '1px solid var(--border-subtle)', padding: '22px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '14px' }}>
            <Mic size={18} color="#10b981" />
            <h3 style={{ fontSize: '16px', fontWeight: 700 }}>Recognized Customer Responses</h3>
          </div>
          <ul style={{ listStyle: 'none', padding: 0, margin: 0, display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <li style={{ fontSize: '12.5px', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ color: '#10b981', fontWeight: 700 }}>• PROMISE_TO_PAY:</span> "Kal tak pakka ho jayega" / "Salary somwar ko aayegi"
            </li>
            <li style={{ fontSize: '12.5px', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ color: '#3b82f6', fontWeight: 700 }}>• REQUEST_LINK:</span> "WhatsApp pe UPI payment link bhejo"
            </li>
            <li style={{ fontSize: '12.5px', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ color: '#f59e0b', fontWeight: 700 }}>• CALL_LATER:</span> "Main abhi driving kar raha hoon baad me phone karo"
            </li>
            <li style={{ fontSize: '12.5px', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ color: '#ef4444', fontWeight: 700 }}>• DISPUTE:</span> "Amount galat hai, hamara issue resolve nahi hua"
            </li>
          </ul>
        </div>
      </div>

      {/* Call History */}
      <div style={{ background: 'var(--bg-secondary)', borderRadius: '14px', border: '1px solid var(--border-subtle)', overflow: 'hidden' }}>
        <div style={{ padding: '18px 24px', borderBottom: '1px solid var(--border-subtle)' }}>
          <h2 style={{ fontSize: '17px', fontWeight: 700 }}>Recent Voice Recovery Telephony Logs</h2>
        </div>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13.5px' }}>
            <thead>
              <tr style={{ background: 'var(--bg-tertiary)', borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-secondary)' }}>
                <th style={{ padding: '14px 20px' }}>Call ID</th>
                <th style={{ padding: '14px 20px' }}>Provider</th>
                <th style={{ padding: '14px 20px' }}>Duration</th>
                <th style={{ padding: '14px 20px' }}>Customer Speech Transcript</th>
                <th style={{ padding: '14px 20px' }}>Extracted Intent</th>
                <th style={{ padding: '14px 20px' }}>Next Action</th>
              </tr>
            </thead>
            <tbody>
              {calls.map((call, idx) => (
                <tr key={idx} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  <td style={{ padding: '16px 20px', fontWeight: 600 }}>{call.call_sid}</td>
                  <td style={{ padding: '16px 20px', textTransform: 'uppercase', fontSize: '11px', fontWeight: 700, color: '#3b82f6' }}>
                    {call.provider}
                  </td>
                  <td style={{ padding: '16px 20px' }}>{call.duration_seconds}s</td>
                  <td style={{ padding: '16px 20px', fontStyle: 'italic', maxWidth: '300px' }}>
                    "{call.customer_spoken}"
                  </td>
                  <td style={{ padding: '16px 20px' }}>
                    <span style={{
                      padding: '4px 10px',
                      borderRadius: '6px',
                      fontSize: '11px',
                      fontWeight: 700,
                      background: 'rgba(16, 185, 129, 0.15)',
                      color: '#10b981',
                    }}>
                      {call.intent}
                    </span>
                  </td>
                  <td style={{ padding: '16px 20px', fontSize: '12px', color: 'var(--text-muted)' }}>
                    {call.ptp_date ? 'PTP Tracked' : 'Link Dispatched'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
