import { FormEvent, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getWorkspace } from '../api/workspaces';
import { listTasks, createTask, getRanked, type TaskFilters } from '../api/tasks';
import type { Workspace, Task, TaskWithRank } from '../api/types';
import { TASK_STATUSES } from '../api/types';

export default function WorkspacePage() {
  const { id } = useParams<{ id: string }>();
  const [ws, setWs] = useState<Workspace | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [ranked, setRanked] = useState<TaskWithRank[]>([]);
  const [view, setView] = useState<'list' | 'ranked'>('list');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [title, setTitle] = useState('');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const load = async () => {
    if (!id) return;
    try {
      const [w, t, r] = await Promise.all([
        getWorkspace(id),
        listTasks({ workspaceID: id, sortBy: 'finalScore', sortDesc: true, search: search || undefined, status: statusFilter || undefined } as TaskFilters),
        getRanked(id),
      ]);
      setWs(w);
      setTasks(t.tasks || []);
      setRanked(r.result?.tasks || []);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [id, search, statusFilter]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    if (!id) return;
    try {
      await createTask({ workspaceId: id, title });
      setTitle('');
      setShowForm(false);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };


  if (loading) return <p>Загрузка...</p>;

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>{ws?.name}</h2>
          <span className="badge">{ws?.scoring?.type}</span>
        </div>
        <div className="actions">
          <button className={view === 'list' ? 'btn btn-active' : 'btn'} onClick={() => setView('list')}>Задачи</button>
          <button className={view === 'ranked' ? 'btn btn-active' : 'btn'} onClick={() => setView('ranked')}>Приоритеты</button>
          <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>+ Задача</button>
        </div>
      </div>

      {error && <div className="error banner">{error}</div>}

      {showForm && (
        <form className="card form" onSubmit={handleCreate}>
          <input placeholder="Название задачи" value={title} onChange={e => setTitle(e.target.value)} required autoFocus />
          <button type="submit" className="btn btn-primary">Создать</button>
        </form>
      )}

      {view === 'list' && (
        <>
          <div className="filters">
            <input placeholder="Поиск..." value={search} onChange={e => setSearch(e.target.value)} />
            <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)}>
              <option value="">Все статусы</option>
              {TASK_STATUSES.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>

          {tasks.length === 0 ? (
            <p className="empty">Нет задач</p>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>Задача</th>
                  <th>Статус</th>
                  <th>Score</th>
                  <th>Голоса</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map(t => (
                  <tr key={t.id}>
                    <td>
                      <Link to={`/workspaces/${id}/tasks/${t.id}`} className="task-link">{t.title}</Link>
                      {t.labels?.map(l => <span key={l} className="label">{l}</span>)}
                    </td>
                    <td><span className={`status status-${t.status}`}>{t.status}</span></td>
                    <td className="score">{t.finalScore > 0 ? t.finalScore.toFixed(1) : '—'}</td>
                    <td>{t.votes?.length || 0}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}

      {view === 'ranked' && (
        ranked.length === 0 ? (
          <p className="empty">Нет оценённых задач</p>
        ) : (
          <table className="table">
            <thead>
              <tr><th>#</th><th>Задача</th><th>Score</th><th>Percentile</th></tr>
            </thead>
            <tbody>
              {ranked.map((t, i) => (
                <tr key={t.id} className={i === 0 ? 'top-rank' : ''}>
                  <td className="rank">{t.rank}</td>
                  <td><Link to={`/workspaces/${id}/tasks/${t.id}`} className="task-link">{t.title}</Link></td>
                  <td className="score">{t.finalScore > 0 ? t.finalScore.toFixed(1) : '—'}</td>
                  <td className="muted">{t.percentile?.toFixed(0)}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        )
      )}
    </div>
  );
}
