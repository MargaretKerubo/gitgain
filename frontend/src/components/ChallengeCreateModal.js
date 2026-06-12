'use client';

import React, { useState } from 'react';
import { API_BASE, useUser } from '../context/UserContext';

export default function ChallengeCreateModal({ onClose }) {
  const { user } = useUser();
  const [form, setForm] = useState({ title:'', description:'', repo_owner:'', repo_name:'', reward_sats:'' });
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault(); setError('');
    const { title, description, repo_owner, repo_name, reward_sats } = form;
    if (!title || !description || !repo_owner || !repo_name || !reward_sats) {
      setError('All fields are required.'); return;
    }
    setSubmitting(true);
    try {
      const res = await fetch(`${API_BASE}/challenges`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, description, repo_owner, repo_name, reward_sats: parseInt(reward_sats, 10), creator_id: user.id }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to create challenge');
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-box" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">Create Challenge</h2>
          <button className="modal-close" onClick={onClose}>&times;</button>
        </div>
        {error && <div style={{ background:'rgba(239,68,68,0.1)', color:'var(--accent-red)', border:'1px solid rgba(239,68,68,0.2)', borderRadius:'6px', padding:'0.75rem', marginBottom:'1rem', fontSize:'0.9rem' }}>{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-group"><label className="form-label">Title</label>
            <input className="input-field" placeholder="Fix bug in auth module" value={form.title} onChange={e=>setForm({...form,title:e.target.value})} required /></div>
          <div className="form-group"><label className="form-label">Description</label>
            <textarea className="input-field" rows={3} placeholder="Describe the task…" value={form.description} onChange={e=>setForm({...form,description:e.target.value})} style={{resize:'vertical'}} required /></div>
          <div className="form-group"><label className="form-label">Repo Owner</label>
            <input className="input-field" placeholder="e.g. torvalds" value={form.repo_owner} onChange={e=>setForm({...form,repo_owner:e.target.value})} required /></div>
          <div className="form-group"><label className="form-label">Repo Name</label>
            <input className="input-field" placeholder="e.g. linux" value={form.repo_name} onChange={e=>setForm({...form,repo_name:e.target.value})} required /></div>
          <div className="form-group"><label className="form-label">Reward (Sats)</label>
            <input type="number" className="input-field" placeholder="e.g. 5000" min="1" value={form.reward_sats} onChange={e=>setForm({...form,reward_sats:e.target.value})} required /></div>
          <button type="submit" className="btn btn-primary" style={{width:'100%',marginTop:'0.5rem'}} disabled={submitting}>
            {submitting ? 'Creating…' : 'Create Challenge'}
          </button>
        </form>
      </div>
    </div>
  );
}
