'use client';

import React, { useState } from 'react';
import { API_BASE, useUser } from '../../context/UserContext';

export default function ProfilePanel() {
  const { user, refreshUser } = useUser();
  const [form, setForm] = useState({
    username: user?.username || '',
    email: user?.email || '',
    github_username: user?.github_username || '',
    lightning_address: user?.lightning_address || '',
  });
  const [saving, setSaving] = useState(false);
  const [saved, setSaved]   = useState(false);
  const [error, setError]   = useState('');

  const handleSave = async (e) => {
    e.preventDefault(); setError(''); setSaved(false);
    setSaving(true);
    try {
      const res = await fetch(`${API_BASE}/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Update failed');
      await refreshUser();
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="card slide-up">
      <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '1.5rem' }}>👤 Profile Settings</h2>
      {error && <div style={{ background:'rgba(239,68,68,0.1)', color:'var(--accent-red)', border:'1px solid rgba(239,68,68,0.2)', borderRadius:'6px', padding:'0.75rem', marginBottom:'1rem', fontSize:'0.9rem' }}>{error}</div>}
      {saved && <div style={{ background:'rgba(16,185,129,0.1)', color:'var(--accent-green)', border:'1px solid rgba(16,185,129,0.2)', borderRadius:'6px', padding:'0.75rem', marginBottom:'1rem', fontSize:'0.9rem' }}>✓ Profile updated successfully!</div>}
      <form onSubmit={handleSave}>
        <div className="form-group"><label className="form-label">Username</label>
          <input className="input-field" value={form.username} onChange={e=>setForm({...form,username:e.target.value})} required /></div>
        <div className="form-group"><label className="form-label">Email</label>
          <input type="email" className="input-field" value={form.email} onChange={e=>setForm({...form,email:e.target.value})} required /></div>
        <div className="form-group"><label className="form-label">GitHub Username</label>
          <input className="input-field" value={form.github_username} onChange={e=>setForm({...form,github_username:e.target.value})} required /></div>
        <div className="form-group"><label className="form-label">Lightning Address</label>
          <input className="input-field" placeholder="you@getalby.com" value={form.lightning_address} onChange={e=>setForm({...form,lightning_address:e.target.value})} />
          <div className="form-helper">Required to receive automatic sats payouts.</div></div>
        <button type="submit" className="btn btn-primary" disabled={saving}>{saving ? 'Saving…' : 'Save Profile'}</button>
      </form>
    </div>
  );
}
