'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useUser } from '../context/UserContext';
import AuthModal from './AuthModal';

export default function Header() {
  const pathname = usePathname();
  const { user, loading, logout } = useUser();
  const [showModal, setShowModal]     = useState(false);
  const [registerMode, setRegMode]    = useState(false);

  const openLogin    = () => { setRegMode(false); setShowModal(true); };
  const openRegister = () => { setRegMode(true);  setShowModal(true); };

  return (
    <>
      <header className="header">
        <div className="nav-container">
          <Link href="/" className="logo">
            ⚡ Git<span className="logo-orange">Gain</span>
          </Link>

          <nav>
            <ul className="nav-links">
              <li><Link href="/"           className={`nav-link ${pathname === '/'           ? 'active' : ''}`}>Home</Link></li>
              <li><Link href="/challenges" className={`nav-link ${pathname === '/challenges' ? 'active' : ''}`}>Challenges</Link></li>
              <li><Link href="/dashboard"  className={`nav-link ${pathname === '/dashboard'  ? 'active' : ''}`}>Dashboard</Link></li>
              <li style={{ marginLeft: '1rem' }}>
                {loading ? (
                  <span className="animate-pulse" style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>Loading…</span>
                ) : user ? (
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <span className="badge badge-success" style={{ textTransform: 'none' }}>🟢 {user.username}</span>
                    <button onClick={logout} className="btn btn-secondary" style={{ padding: '0.35rem 0.8rem', fontSize: '0.85rem' }}>Logout</button>
                  </div>
                ) : (
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <button onClick={openLogin}    className="btn btn-secondary" style={{ padding: '0.35rem 0.8rem', fontSize: '0.85rem' }}>Sign In</button>
                    <button onClick={openRegister} className="btn btn-primary"   style={{ padding: '0.35rem 0.8rem', fontSize: '0.85rem' }}>Register</button>
                  </div>
                )}
              </li>
            </ul>
          </nav>
        </div>
      </header>

      {showModal && (
        <AuthModal
          initialRegister={registerMode}
          onClose={() => setShowModal(false)}
        />
      )}
    </>
  );
}
