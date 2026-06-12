'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { API_BASE } from '../../context/UserContext';
import { useUser } from '../../context/UserContext';
import ChallengeCreateModal from '../../components/ChallengeCreateModal';
import SubmitPRModal from '../../components/SubmitPRModal';

export default function ChallengesPage() {
  const { user } = useUser();
  const [challenges, setChallenges] = useState([]);
  const [loading, setLoading]       = useState(true);
  const [expanded, setExpanded]     = useState(null);
  const [showCreate, setShowCreate] = useState(false);
  const [submitFor, setSubmitFor]   = useState(null);

  const load = useCallback(() => {
    setLoading(true);
    fetch(`${API_BASE}/challenges`)
      .then(r => r.json())
      .then(data => { setChallenges(Array.isArray(data) ? data : []); setLoading(false); })
      .catch(() => setLoading(false));
  }, []);

  useEffect(() => { load(); }, [load]);

  const toggle = (id) => setExpanded(e => (e === id ? null : id));

  return (
    <div className="container fade-in" style={{ padding: '3rem 1.5rem' }}>
      <div className="section-header">
        <h1 style={{ fontSize: '2rem', fontWeight: 700 }}>⚡ Open Challenges</h1>
        {user && (
          <button onClick={() => setShowCreate(true)} className="btn btn-primary">
            + Create Challenge
          </button>
        )}
      </div>

      {loading ? (
        <div className="empty-box"><span className="animate-pulse">Loading challenges…</span></div>
      ) : challenges.length === 0 ? (
        <div className="empty-box">
          <p style={{ fontSize: '1.1rem', marginBottom: '0.5rem' }}>No challenges yet.</p>
          {user && <button onClick={() => setShowCreate(true)} className="btn btn-primary" style={{ marginTop: '1rem' }}>Create the first one</button>}
        </div>
      ) : (
        <div className="card">
          {challenges.map(ch => (
            <div key={ch.id} className="challenge-item">
              <div className="challenge-top" onClick={() => toggle(ch.id)}>
                <div>
                  <h3 style={{ fontWeight: 600 }}>{ch.title}</h3>
                  <div className="challenge-meta">
                    <span>📂 {ch.repo_owner}/{ch.repo_name}</span>
                    <span className={`badge badge-${ch.status === 'active' ? 'success' : 'pending'}`}>{ch.status}</span>
                  </div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div className="challenge-reward">⚡ {ch.reward_sats.toLocaleString()} sats</div>
                  <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginTop: '0.3rem' }}>{expanded === ch.id ? '▲ Hide' : '▼ Details'}</div>
                </div>
              </div>

              {expanded === ch.id && (
                <div className="challenge-body">
                  <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem', lineHeight: 1.6 }}>{ch.description}</p>
                  <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
                    <a href={`https://github.com/${ch.repo_owner}/${ch.repo_name}`} target="_blank" rel="noreferrer" className="btn btn-secondary" style={{ fontSize: '0.85rem', padding: '0.4rem 0.9rem' }}>View Repo →</a>
                    {user && ch.status === 'active' && (
                      <button onClick={() => setSubmitFor(ch)} className="btn btn-primary" style={{ fontSize: '0.85rem', padding: '0.4rem 0.9rem' }}>Submit PR</button>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {showCreate && <ChallengeCreateModal onClose={() => { setShowCreate(false); load(); }} />}
      {submitFor  && <SubmitPRModal challenge={submitFor} onClose={() => { setSubmitFor(null); load(); }} />}
    </div>
  );
}
