import { FormEvent, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { listWorkspaces, createWorkspace } from '../api/workspaces';
import { useAuth } from '../store/AuthContext';
import { SCORING_PRESETS, type Workspace } from '../api/types';

export default function WorkspacesPage() {
  const { user } = useAuth();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState('');
  const [scoringType, setScoringType] = useState('RICE');

  const load = async () => {
    if (!user) return;
    try {
      const res = await listWorkspaces(user.id);
      setWorkspaces(res.workspaces || []);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [user]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await createWorkspace({
        name,
        scoring: { type: scoringType, criteria: SCORING_PRESETS[scoringType] },
      });
      setName('');
      setShowForm(false);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  return (
    <div>
      <div className="page-header">
        <h2>Workspaces</h2>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>+ Создать</button>
      </div>

      {error && <div className="error banner">{error}</div>}
      {showForm && (
        <form className="card form" onSubmit={handleCreate}>
          <input placeholder="Название workspace" value={name} onChange={e => setName(e.target.value)} required />
          <select value={scoringType} onChange={e => setScoringType(e.target.value)}>
            <option value="RICE">RICE (Reach, Impact, Confidence, Effort)</option>
            <option value="ICE">ICE (Impact, Confidence, Ease)</option>
            <option value="WSJF">WSJF (Weighted Shortest Job First)</option>
          </select>
          <button type="submit" className="btn btn-primary">Создать</button>
        </form>
      )}

      {loading ? (
        <p>Загрузка...</p>
      ) : workspaces.length === 0 ? (
        <p className="empty">Нет workspaces. Создайте первый.</p>
      ) : (
        <div className="card-grid">
          {workspaces.map(ws => (
            <Link to={`/workspaces/${ws.id}`} key={ws.id} className="card ws-card">
              <h3>{ws.name}</h3>
              <span className="badge">{ws.scoring?.type || '—'}</span>
              <p className="muted">{new Date(ws.createdAt).toLocaleDateString()}</p>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
