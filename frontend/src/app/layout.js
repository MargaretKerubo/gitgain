import { Inter } from 'next/font/google';
import './globals.css';
import { UserProvider } from '../context/UserContext';
import Header from '../components/Header';

const inter = Inter({ subsets: ['latin'] });

export const metadata = {
  title: 'GitGain – Lightning-Powered Contribution Rewards',
  description: 'Complete GitHub challenges and earn instant Bitcoin Lightning payouts.',
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body className={inter.className} style={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
        <UserProvider>
          <Header />
          <main style={{ flex: 1 }}>{children}</main>
          <footer className="footer">
            <div className="container">
              ⚡ GitGain &nbsp;·&nbsp; Lightning-Powered Open-Source Rewards &nbsp;·&nbsp;
              <span style={{ color: 'var(--accent-orange)' }}>Built with Bitcoin</span>
            </div>
          </footer>
        </UserProvider>
      </body>
    </html>
  );
}
