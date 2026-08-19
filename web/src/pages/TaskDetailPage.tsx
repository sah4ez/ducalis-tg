import { FormEvent, useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { getTask, updateTask, deleteTask, setScores, vote, removeVote, estimate } from '../api/tasks';
import { getWorkspace } from '../api/workspaces';
import { useAuth } from '../store/AuthContext';
import type { Task, Workspace } from '../api/types';
import { TASK_STATUSES, TASK_PRIORITIES, SCORING_PRESETS } from '../api/types';

export default function TaskDetailPage() {
  const { wsId, taskId } = useParams<{ wsId: string; taskId: string }>();
  const { user } = useAuth();
  const nav = useNavigate();
  const [task, setTask] = useState<Task | null>(null);
  const [ws, setWs] = useState<Workspace | null>(null);
  const [error, setError] = useState('');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [status, setStatus] = useState('backlog');
  const [priority, setPriority] = useState('');
  const [scores, setScoresState] = useState<Record<string, number>>({});
  const [estValue, setEstValue] = useState('');
  const [estUnit, setEstUnit] = useState('points');

  const load = async () => {
    if (!taskId || !wsId) return;
    try {
      const [t, w] = await Promise.all([getTask(taskId), getWorkspace(wsId)]);
      setTask(t);
      setWs(w);
      setTitle(t.title);
      setDescription(t.description || '');
      setStatus(t.status);
      setPriority(t.priority || '');
      setScoresState(t.scores || {});
    } catch (e: any) {
      setError(e.message);
    }
  };

  useEffect(() => { load(); }, [taskId]);

  if (!task) return error ? <div className="error banner">{error}</div> : <p>Загрузка...</p>;

  const criteria = ws?.scoring?.criteria || SCORING_PRESETS[ws?.scoring?.type || ''] || [];
  const hasVoted = task.votes?.some(v => v.userId === user?.id);
  const myEstimate = task.estimations?.find(e => e.userId === user?.id);

  const handleSave = async (e: FormEvent) => {
    e.preventDefault();
    try {
      const updated = await updateTask(task.id, { title, description, status, priority: priority || undefined });
      setTask(updated);
      setError('');
    } catch (err: any) { setError(err.message); }
  };

  const handleScores = async () => {
    try {
      const updated = await setScores(task.id, scores);
      setTask(updated);
    } catch (err: any) { setError(err.message); }
  };

  const handleVote = async () => {
    try {
      const updated = hasVoted ? await removeVote(task.id) : await vote(task.id, 1.0);
      setTask(updated);
    } catch (err: any) { setError(err.message); }
  };

  const handleEstimate = async (e: FormEvent) => {
    e.preventDefault();
    const val = parseFloat(estValue);
    if (isNaN(val)) return;
    try {
      const updated = await estimate(task.id, val, estUnit);
      setTask(updated);
      setEstValue('');
    } catch (err: any) { setError(err.message); }
  };

  const handleDelete = async () => {
    if (!confirm('Удалить задачу?')) return;
    await deleteTask(task.id);
    nav(`/workspaces/${wsId}`);
  };

  return (
    <div>
      <div className="breadcrumbs">
        <Link to="/workspaces">Workspaces</Link> / <Link to={`/workspaces/${wsId}`}>{ws?.name}</Link> / {task.title}
      </div>

      {error && <div className="error banner">{error}</div>}

      <div className="detail-grid">
        <div className="card">
          <h3>Задача</h3>
          <form onSubmit={handleSave}>
            <input value={title} onChange={e => setTitle(e.target.value)} required />
            <textarea placeholder="Описание" value={description} onChange={e => setDescription(e.target.value)} rows={4} />
            <div className="form-row">
              <select value={status} onChange={e => setStatus(e.target.value)}>
                {TASK_STATUSES.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
              <select value={priority} onChange={e => setPriority(e.target.value)}>
                <option value="">Приоритет: —</option>
                {TASK_PRIORITIES.map(p => <option key={p} value={p}>{p}</option>)}
              </select>
            </div>
            <div className="form-row">
              <button type="submit" className="btn btn-primary">Сохранить</button>
              <button type="button" className="btn btn-danger" onClick={handleDelete}>Удалить</button>
            </div>
          </form>
        </div>

        <div className="card">
          <h3>Оценка ({ws?.scoring?.type})</h3>
          {criteria.length === 0 ? (
            <p className="muted">Workspace без критериев скоринга</p>
          ) : (
            criteria.map(c => (
              <div key={c.id} className="score-row">
                <label>{c.name}</label>
                <input
                  type="number" step="0.5" min="0"
                  value={scores[c.id] ?? ''}
                  onChange={e => setScoresState({ ...scores, [c.id]: parseFloat(e.target.value) || 0 })}
                />
              </div>
            ))
          )}
          <button className="btn btn-primary" onClick={handleScores}>Сохранить оценки</button>
          <div className="final-score">Final Score: <strong>{task.finalScore > 0 ? task.finalScore.toFixed(1) : '—'}</strong></div>
        </div>

        <div className="card">
          <h3>Голосование</h3>
          <button className={`btn ${hasVoted ? 'btn-secondary' : 'btn-primary'}`} onClick={handleVote}>
            {hasVoted ? 'Снять голос' : 'Голосовать'}
          </button>
          <p className="muted">{task.votes?.length || 0} голосов</p>
        </div>

        <div className="card">
          <h3>Оценка трудозатрат</h3>
          <form onSubmit={handleEstimate} className="form-row">
            <input type="number" step="0.5" min="0" placeholder={myEstimate ? `${myEstimate.value} ${myEstimate.unit}` : "Значение"} value={estValue} onChange={e => setEstValue(e.target.value)} />
            <select value={estUnit} onChange={e => setEstUnit(e.target.value)}>
              <option value="points">SP</option>
              <option value="hours">Часы</option>
            </select>
            <button type="submit" className="btn btn-primary">OK</button>
          </form>
          {task.estimations?.map((e, i) => (
            <div key={i} className="estimate-row">
              <span>{e.userId === user?.id ? 'Вы' : e.userId.slice(0, 8)}</span>
              <span>{e.value} {e.unit}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
