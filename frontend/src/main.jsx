import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { Activity, FileText, Play, RefreshCw, Server, Shield, TrendingUp } from 'lucide-react';
import './styles.css';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

function App() {
  const [credentials, setCredentials] = useState(() => ({
    username: localStorage.getItem('dashboardUsername') || 'admin',
    password: ''
  }));
  const [namespace, setNamespace] = useState('');
  const [pods, setPods] = useState([]);
  const [deployments, setDeployments] = useState([]);
  const [logs, setLogs] = useState({ title: 'Select a pod to view logs', body: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const authHeader = useMemo(() => `Basic ${btoa(`${credentials.username}:${credentials.password}`)}`, [credentials]);

  async function request(path, options = {}) {
    const response = await fetch(`${API_URL}${path}`, {
      ...options,
      headers: {
        Authorization: authHeader,
        'Content-Type': 'application/json',
        ...(options.headers || {})
      }
    });

    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || `Request failed with status ${response.status}`);
    }
    return data;
  }

  async function refresh() {
    if (!credentials.password) {
      return;
    }

    setLoading(true);
    setError('');
    try {
      const query = namespace ? `?namespace=${encodeURIComponent(namespace)}` : '';
      const [podData, deploymentData] = await Promise.all([
        request(`/api/pods${query}`),
        request(`/api/deployments${query}`)
      ]);
      setPods(podData);
      setDeployments(deploymentData);
      localStorage.setItem('dashboardUsername', credentials.username);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function viewLogs(pod) {
    setError('');
    try {
      const data = await request(`/api/pods/${pod.namespace}/${pod.name}/logs?tailLines=200`);
      setLogs({ title: `${pod.namespace}/${pod.name}`, body: data.logs || 'No logs returned.' });
    } catch (err) {
      setError(err.message);
    }
  }

  async function restartDeployment(deployment) {
    setError('');
    try {
      await request(`/api/deployments/${deployment.namespace}/${deployment.name}/restart`, { method: 'POST' });
      await refresh();
    } catch (err) {
      setError(err.message);
    }
  }

  async function scaleDeployment(deployment, replicas) {
    setError('');
    try {
      await request(`/api/deployments/${deployment.namespace}/${deployment.name}/scale`, {
        method: 'POST',
        body: JSON.stringify({ replicas })
      });
      await refresh();
    } catch (err) {
      setError(err.message);
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  const runningPods = pods.filter((pod) => pod.status === 'Running').length;
  const readyDeployments = deployments.filter((deployment) => deployment.readyReplicas === deployment.replicas).length;

  return (
    <main className="app-shell">
      <section className="topbar">
        <div>
          <p className="eyebrow">Cluster overview</p>
          <h1>Kubernetes Deployment Dashboard</h1>
        </div>
        <button className="icon-button" onClick={refresh} disabled={loading || !credentials.password} title="Refresh">
          <RefreshCw size={18} className={loading ? 'spin' : ''} />
        </button>
      </section>

      <section className="controls">
        <label>
          <span>Username</span>
          <input value={credentials.username} onChange={(event) => setCredentials({ ...credentials, username: event.target.value })} />
        </label>
        <label>
          <span>Password</span>
          <input type="password" value={credentials.password} onChange={(event) => setCredentials({ ...credentials, password: event.target.value })} />
        </label>
        <label>
          <span>Namespace</span>
          <input value={namespace} placeholder="all namespaces" onChange={(event) => setNamespace(event.target.value)} />
        </label>
        <button className="primary-action" onClick={refresh}>
          <Shield size={18} />
          Connect
        </button>
      </section>

      {error && <div className="error-banner">{error}</div>}

      <section className="metric-grid">
        <Metric icon={<Server />} label="Pods" value={pods.length} detail={`${runningPods} running`} />
        <Metric icon={<Activity />} label="Deployments" value={deployments.length} detail={`${readyDeployments} fully ready`} />
        <Metric icon={<TrendingUp />} label="Replicas" value={deployments.reduce((sum, item) => sum + item.replicas, 0)} detail="desired total" />
      </section>

      <section className="content-grid">
        <Panel title="Deployments">
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Namespace</th>
                  <th>Replicas</th>
                  <th>Ready</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {deployments.map((deployment) => (
                  <DeploymentRow
                    key={`${deployment.namespace}-${deployment.name}`}
                    deployment={deployment}
                    onRestart={restartDeployment}
                    onScale={scaleDeployment}
                  />
                ))}
              </tbody>
            </table>
          </div>
        </Panel>

        <Panel title="Pods">
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Namespace</th>
                  <th>Status</th>
                  <th>Restarts</th>
                  <th>Logs</th>
                </tr>
              </thead>
              <tbody>
                {pods.map((pod) => (
                  <tr key={`${pod.namespace}-${pod.name}`}>
                    <td>{pod.name}</td>
                    <td>{pod.namespace}</td>
                    <td><StatusBadge status={pod.status} /></td>
                    <td>{pod.restarts}</td>
                    <td>
                      <button className="icon-button" onClick={() => viewLogs(pod)} title="View logs">
                        <FileText size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Panel>
      </section>

      <Panel title={`Logs: ${logs.title}`}>
        <pre className="logs">{logs.body}</pre>
      </Panel>
    </main>
  );
}

function Metric({ icon, label, value, detail }) {
  return (
    <article className="metric-card">
      <div className="metric-icon">{icon}</div>
      <div>
        <span>{label}</span>
        <strong>{value}</strong>
        <p>{detail}</p>
      </div>
    </article>
  );
}

function Panel({ title, children }) {
  return (
    <section className="panel">
      <h2>{title}</h2>
      {children}
    </section>
  );
}

function DeploymentRow({ deployment, onRestart, onScale }) {
  const [replicas, setReplicas] = useState(deployment.replicas);

  useEffect(() => {
    setReplicas(deployment.replicas);
  }, [deployment.replicas]);

  return (
    <tr>
      <td>{deployment.name}</td>
      <td>{deployment.namespace}</td>
      <td>
        <input
          className="replica-input"
          type="number"
          min="0"
          value={replicas}
          onChange={(event) => setReplicas(Number(event.target.value))}
        />
      </td>
      <td>{deployment.readyReplicas}/{deployment.replicas}</td>
      <td className="action-cell">
        <button className="icon-button" onClick={() => onScale(deployment, replicas)} title="Scale deployment">
          <TrendingUp size={16} />
        </button>
        <button className="icon-button" onClick={() => onRestart(deployment)} title="Restart deployment">
          <Play size={16} />
        </button>
      </td>
    </tr>
  );
}

function StatusBadge({ status }) {
  return <span className={`status status-${status.toLowerCase()}`}>{status}</span>;
}

createRoot(document.getElementById('root')).render(<App />);
