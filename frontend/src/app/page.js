'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { API_BASE } from '../context/UserContext';

export default function HomePage() {
  const [stats, setStats] = useState(null);

  useEffect(() => {
    fetch(`${API_BASE}/stats`)
      .then(r => r.json())
      .then(setStats)
      .catch(() => {});
  }, []);

  return (
    <div className="fade-in">
      {/* Hero */}
      <section style={{ padding: '5rem 0 3rem', textAlign: 'center' }}>
        <div className="container">
          <div className="hero">
            <h1 className="hero-title">
              Earn <span className="text-gradient">Bitcoin</span> for<br />Open-Source Work
            </h1>
            <p className="hero-subtitle">
              Complete GitHub challenges, open a pull request, and receive instant ⚡ Lightning payouts —
              automatically triggered when your CI passes.
            </p>
            <div className="hero-buttons">
              <Link href="/challenges" className="btn btn-primary" style={{ fontSize: '1rem', padding: '0.85rem 2rem' }}>
                Browse Challenges →
              </Link>
              <Link href="/dashboard" className="btn btn-secondary" style={{ fontSize: '1rem', padding: '0.85rem 2rem' }}>
                My Dashboard
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Stats */}
      <section style={{ padding: '2rem 0 4rem' }}>
        <div className="container">
          <div className="grid-3">
            <div className="stat-card slide-up">
              <div className="stat-value orange">{stats ? stats.total_sats_paid.toLocaleString() : '—'}</div>
              <div className="stat-label">Sats Paid Out</div>
            </div>
            <div className="stat-card slide-up">
              <div className="stat-value">{stats ? stats.active_challenges : '—'}</div>
              <div className="stat-label">Active Challenges</div>
            </div>
            <div className="stat-card slide-up">
              <div className="stat-value">{stats ? stats.completed_submissions : '—'}</div>
              <div className="stat-label">Completed Bounties</div>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
