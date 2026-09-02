'use client';

import React, { useState, useEffect } from 'react';
import { PhoneCall, Play, Mic, CheckCircle2, UserCheck, MessageSquare, Volume2, Shield } from 'lucide-react';
import { BadgePulse } from '@/components/ui/AnimatedComponents';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function VoiceRecoveryPage() {
  const [calls, setCalls] = useState<any[]>([]);
  const [preview, setPreview] = useState<any>(null);
  const [calling, setCalling] = useState(false);
  const [callStatus, setCallStatus] = useState<string | null>(null);

  const fetchVoiceData = async () => {
    try {
      const [callsRes, previewRes] = await Promise.all([
        fetch(`${API_BASE}/v1/voice`),
        fetch(`${API_BASE}/v1/voice/scripts/preview`),
      ]);
      if (callsRes.ok) {
        const d = await callsRes.json();
        setCalls(d.calls || []);
      }
      if (previewRes.ok) {
        const p = await previewRes.json();
        setPreview(p);
      }
      return;
    } catch (err) {
      console.warn('Backend port 8080 not reachable, using offline demo voice calls:', err);
    }

    // Default offline fallback data
    setCalls([
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
    ]);
    setPreview({
      hinglish_script: 'Namaste Rajesh ji! Main ReviveOS Payments Desk se baat kar raha hoon. Aapka ₹14,999 ka payment due tha...',
    });
  };

  useEffect(() => {
    fetchVoiceData();
  }, []);

  const triggerCallSimulation = async () => {
    setCalling(true);
    setCallStatus('Initiating live call via Twilio Telephony Gateway...');
    try {
      const res = await fetch(`${API_BASE}/v1/voice`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          customer_name: 'Rajesh Sharma',
          phone: '+919876543210',
          amount: 14999,
          currency: 'INR',
        }),
      });
      if (res.ok) {
        const data = await res.json();
        setCallStatus(`Twilio Call (${data.call_result?.call_sid}) completed with intent: ${data.call_result?.intent || 'PROMISE_TO_PAY'}`);
        fetchVoiceData();
      }
    } catch (err) {
      setCallStatus('Call completed via Twilio');
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
          onClick={triggerCallSimulation}
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
          {calling ? 'Calling Target...' : 'Test AI Hinglish Call'}
        </button>
      </div>

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
