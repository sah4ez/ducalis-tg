import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './store/AuthContext';
import LoginPage from './pages/LoginPage';
import WorkspacesPage from './pages/WorkspacesPage';
import WorkspacePage from './pages/WorkspacePage';
import TaskDetailPage from './pages/TaskDetailPage';
import Layout from './components/Layout';

function App() {
  const { token } = useAuth();

  if (!token) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    );
  }

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/workspaces" replace />} />
        <Route path="/workspaces" element={<WorkspacesPage />} />
        <Route path="/workspaces/:id" element={<WorkspacePage />} />
        <Route path="/workspaces/:wsId/tasks/:taskId" element={<TaskDetailPage />} />
        <Route path="*" element={<Navigate to="/workspaces" replace />} />
      </Routes>
    </Layout>
  );
}

export default App;
