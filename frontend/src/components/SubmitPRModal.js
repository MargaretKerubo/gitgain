'use client';

import React, { useState } from 'react';
import { API_BASE, useUser } from '../context/UserContext';

export default function SubmitPRModal({ challenge, onClose }) {
  const { user } = useUser();
  const [prURL, setPrURL]   = useState('');
  const [prNum, setPrNum]   = useState('');
  const [error, setError]   = useState('');
  const [submitting, setSub] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault(); setError('');
    if (!prURL || !prNum) { setError('Both PR URL and number are required.'); return; }
    setSub(true);
    try {
      const res = await fetch(`${API_BASE}/submissions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ challenge_id: challenge.id, user_id: user.id, pull_request_url: prURL, pull_request_number: parseInt(prNum, 10) }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Submission failed');
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setSub(false);
    }
  };

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-box" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">Submit Pull Request</h2>
          <button className="modal-close" onClick={onClose}>&times;</button>
        </div>
        <p style={{ color:'var(--text-secondary)', marginBottom:'1.25rem', fontSize:'0.9rem' }}>
          Challenge: <strong style={{ color:'var(--text-primary)' }}>{challenge.title}</strong>
          &nbsp;·&nbsp; Reward: <span style={{ color:'var(--accent-orange)', fontWeight:700 }}>⚡ {challenge.reward_sats.toLocaleString()} sats</span>
        </p>
        {error && <div style={{ background:'rgba(239,68,68,0.1)', color:'var(--accent-red)', border:'1px solid rgba(239,68,68,0.2)', borderRadius:'6px', padding:'0.75rem', marginBottom:'1rem', fontSize:'0.9rem' }}>{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-group"><label className="form-label">Pull Request URL</label>
            <input className="input-field" placeholder="https://github.com/owner/repo/pull/42" value={prURL} onChange={e=>setPrURL(e.target.value)} required />
            <div className="form-helper">Paste the full GitHub PR URL.</div></div>
          <div className="form-group"><label className="form-label">PR Number</label>
            <input type="number" className="input-field" placeholder="42" min="1" value={prNum} onChange={e=>setPrNum(e.target.value)} required /></div>
          <button type="submit" className="btn btn-primary" style={{width:'100%',marginTop:'0.5rem'}} disabled={submitting}>
            {submitting ? 'Submitting…' : 'Submit for Review'}
          </button>
        </form>
      </div>
    </div>
  );
}
