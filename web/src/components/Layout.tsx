import { ReactNode } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../store/AuthContext';

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();
  const nav = useNavigate();

  const handleLogout = () => {
    logout();
    nav('/login');
  };

  return (
    <div className="layout">
      <header className="header">
        <Link to="/workspaces" className="logo">Ducalis</Link>
        <nav className="nav">
          <Link to="/workspaces">Workspaces</Link>
        </nav>
        <div className="user-info">
          <span>{user?.name}</span>
          <button onClick={handleLogout} className="btn btn-ghost">Выйти</button>
        </div>
      </header>
      <main className="content">{children}</main>
    </div>
  );
}
