'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useUser, API_BASE } from '../../context/UserContext';
import ProfilePanel from '../../components/ProfilePanel';


export default function DashboardPage() {
  const { user, loading } = useUser();
  const [activeTab, setActiveTab]   = useState('submissions'); // submissions, profile
  const [submissions, setSubmissions] = useState([]);
  const [fetching, setFetching]       = useState(false);

  const fetchSubs = useCallback(async () => {
    if (!user) return;
    setFetching(true);
    try {
      const res = await fetch(`${API_BASE}/submissions`);
      if (res.ok) {
        const data = await res.json();
        // Filter submissions to show only current user's submissions
        const filtered = (Array.isArray(data) ? data : []).filter(
          s => s.user_id === user.id || s.user?.id === user.id
        );
        setSubmissions(filtered);
      }
    } catch (_) {}
    setFetching(false);
  }, [user]);

  useEffect(() => { fetchSubs(); }, [fetchSubs]);

  const handlePayout = async (subId) => {
    if (!confirm('Are you sure you want to trigger manual payout?')) return;
    try {
      const res = await fetch(`${API_BASE}/admin/payout/${subId}`, { method: 'POST' });
      const data = await res.json();
      if (!res.ok) alert(data.error || 'Payout failed');
      else { alert('Payout processed successfully!'); fetchSubs(); }
    } catch (err) { alert(err.message); }
  };

  if (loading) return <div className="empty-box"><span className="animate-pulse">Loading dashboard...</span></div>;
  if (!user) return <div className="empty-box"><p className="empty-box-title">Please Sign In or Register to access your Dashboard.</p></div>;

  return (
    <div className="container fade-in" style={{ padding: '3rem 1.5rem' }}>
      <h1 style={{ fontSize: '2rem', fontWeight: 700, marginBottom: '1.5rem' }}>⚡ Dashboard</h1>
      <div className="tabs">
        <button className={`tab ${activeTab === 'submissions' ? 'active' : ''}`} onClick={() => setActiveTab('submissions')}>Submissions & Payouts</button>
        <button className={`tab ${activeTab === 'profile' ? 'active' : ''}`} onClick={() => setActiveTab('profile')}>Profile Settings</button>
      </div>

      {activeTab === 'profile' && <ProfilePanel />}

      {activeTab === 'submissions' && (
        <div className="card slide-up">
          <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '1.5rem' }}>📋 Your Submissions</h2>
          {fetching ? (
            <div className="animate-pulse" style={{ color: 'var(--text-secondary)' }}>Loading submissions...</div>
          ) : submissions.length === 0 ? (
            <div style={{ color: 'var(--text-muted)' }}>No submissions found for your profile.</div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              {submissions.map(sub => (
                <div key={sub.id} style={{ borderBottom: '1px solid var(--border-color)', paddingBottom: '1.5rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <div>
                      <h3 style={{ fontWeight: 600 }}>{sub.challenge?.title || `Challenge #${sub.challenge_id}`}</h3>
                      <div className="challenge-meta" style={{ marginTop: '0.4rem' }}>
                        <a href={sub.pull_request_url} target="_blank" rel="noreferrer" className="nav-link" style={{ color: 'var(--accent-orange)' }}>PR #{sub.pull_request_number} ↗</a>
                        <span className={`badge badge-${sub.status === 'completed' ? 'success' : sub.status === 'failed' ? 'failed' : 'pending'}`}>{sub.status}</span>
                      </div>
                    </div>
                    {sub.status !== 'completed' && (
                      <button className="btn btn-primary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem' }} onClick={() => handlePayout(sub.id)}>Trigger Payout</button>
                    )}
                  </div>
                  {sub.payment_hash && (
                    <div style={{ marginTop: '0.5rem', fontSize: '0.85rem' }}>
                      <span style={{ color: 'var(--text-muted)' }}>Payment Preimage:</span> <code className="preimage-code">{sub.payment_hash}</code>
                    </div>
                  )}
                  {sub.error_message && (
                    <div style={{ marginTop: '0.5rem', fontSize: '0.85rem', color: 'var(--accent-red)' }}>
                      <strong>Error:</strong> {sub.error_message}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
