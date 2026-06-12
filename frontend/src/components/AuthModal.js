'use client';

import React, { useState } from 'react';
import { useUser } from '../context/UserContext';

export default function AuthModal({ initialRegister = false, onClose }) {
  const { login, register } = useUser();
  const [isReg, setIsReg]   = useState(initialRegister);
  const [error, setError]   = useState('');
  const [username, setUsername] = useState('');
  const [form, setForm] = useState({ username:'', email:'', github_username:'', lightning_address:'' });

  const handleLogin = async (e) => {
    e.preventDefault(); setError('');
    try { await login(username.trim()); onClose(); }
    catch (err) { setError(err.message); }
  };

  const handleRegister = async (e) => {
    e.preventDefault(); setError('');
    const { username: u, email, github_username, lightning_address } = form;
    if (!u || !email || !github_username) { setError('Username, Email, and GitHub Username are required'); return; }
    try { await register({ username: u.trim(), email: email.trim(), github_username: github_username.trim(), lightning_address: lightning_address.trim() }); onClose(); }
    catch (err) { setError(err.message); }
  };

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-box" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">{isReg ? 'Create Profile' : 'Sign In'}</h2>
          <button className="modal-close" onClick={onClose}>&times;</button>
        </div>

        {error && (
          <div style={{ background:'rgba(239,68,68,0.1)', color:'var(--accent-red)', border:'1px solid rgba(239,68,68,0.2)', borderRadius:'6px', padding:'0.75rem', marginBottom:'1rem', fontSize:'0.9rem' }}>
            {error}
          </div>
        )}

        {isReg ? (
          <form onSubmit={handleRegister}>
            <div className="form-group"><label className="form-label">Username</label>
              <input className="input-field" placeholder="alice" value={form.username} onChange={e=>setForm({...form,username:e.target.value})} required /></div>
            <div className="form-group"><label className="form-label">Email</label>
              <input type="email" className="input-field" placeholder="alice@example.com" value={form.email} onChange={e=>setForm({...form,email:e.target.value})} required /></div>
            <div className="form-group"><label className="form-label">GitHub Username</label>
              <input className="input-field" placeholder="alice-gh" value={form.github_username} onChange={e=>setForm({...form,github_username:e.target.value})} required />
              <div className="form-helper">Links your PRs to submissions automatically.</div></div>
            <div className="form-group"><label className="form-label">Lightning Address</label>
              <input className="input-field" placeholder="alice@getalby.com" value={form.lightning_address} onChange={e=>setForm({...form,lightning_address:e.target.value})} />
              <div className="form-helper">Required to receive sats payouts.</div></div>
            <button type="submit" className="btn btn-primary" style={{width:'100%',marginTop:'0.5rem'}}>Register</button>
            <p style={{marginTop:'1rem',textAlign:'center',fontSize:'0.85rem',color:'var(--text-secondary)'}}>
              Already registered? <span onClick={()=>setIsReg(false)} style={{color:'var(--accent-orange)',cursor:'pointer',fontWeight:600}}>Sign In</span>
            </p>
          </form>
        ) : (
          <form onSubmit={handleLogin}>
            <div className="form-group"><label className="form-label">Username</label>
              <input className="input-field" placeholder="Enter your username" value={username} onChange={e=>setUsername(e.target.value)} required /></div>
            <button type="submit" className="btn btn-primary" style={{width:'100%',marginTop:'0.5rem'}}>Sign In</button>
            <p style={{marginTop:'1rem',textAlign:'center',fontSize:'0.85rem',color:'var(--text-secondary)'}}>
              No profile yet? <span onClick={()=>setIsReg(true)} style={{color:'var(--accent-orange)',cursor:'pointer',fontWeight:600}}>Register</span>
            </p>
          </form>
        )}
      </div>
    </div>
  );
}
