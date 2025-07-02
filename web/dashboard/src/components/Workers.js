import React, { useState, useEffect } from 'react';

const Workers = () => {
  const [workers, setWorkers] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setWorkers([
      {
        id: 'worker-1',
        address: 'localhost:9001',
        task_types: ['etl', 'generic'],
        status: 'active',
        last_heartbeat: new Date().toISOString(),
        current_tasks: ['task-123', 'task-456']
      },
      {
        id: 'worker-2', 
        address: 'localhost:9002',
        task_types: ['ml_training'],
        status: 'active',
        last_heartbeat: new Date(Date.now() - 30000).toISOString(),
        current_tasks: []
      },
      {
        id: 'worker-3',
        address: 'localhost:9003', 
        task_types: ['ci', 'generic'],
        status: 'idle',
        last_heartbeat: new Date(Date.now() - 120000).toISOString(),
        current_tasks: []
      }
    ]);
    setLoading(false);
  }, []);

  const getStatusBadgeClass = (status, lastHeartbeat) => {
    const heartbeatAge = new Date() - new Date(lastHeartbeat);
    if (heartbeatAge > 120000) { // 2 minutes
      return 'status-badge status-failed';
    }
    switch (status) {
      case 'active': return 'status-badge status-running';
      case 'idle': return 'status-badge status-pending';
      default: return 'status-badge';
    }
  };

  const getStatusText = (status, lastHeartbeat) => {
    const heartbeatAge = new Date() - new Date(lastHeartbeat);
    if (heartbeatAge > 120000) {
      return 'offline';
    }
    return status;
  };

  const formatLastSeen = (timestamp) => {
    const diff = new Date() - new Date(timestamp);
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) return `${hours}h ago`;
    if (minutes > 0) return `${minutes}m ago`;
    return `${seconds}s ago`;
  };

  if (loading) {
    return <div className="loading">Loading workers...</div>;
  }

  return (
    <div>
      <div className="workflow-header">
        <h1 style={{ color: '#2d3748' }}>Workers</h1>
        <div className="workflow-actions">
          <button className="btn btn-primary">Register Worker</button>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '1.5rem', marginBottom: '2rem' }}>
        <div className="metric-card">
          <div className="metric-value">{workers.filter(w => getStatusText(w.status, w.last_heartbeat) === 'active').length}</div>
          <div className="metric-label">Active Workers</div>
        </div>
        <div className="metric-card">
          <div className="metric-value">{workers.filter(w => getStatusText(w.status, w.last_heartbeat) === 'idle').length}</div>
          <div className="metric-label">Idle Workers</div>
        </div>
        <div className="metric-card">
          <div className="metric-value">{workers.filter(w => getStatusText(w.status, w.last_heartbeat) === 'offline').length}</div>
          <div className="metric-label">Offline Workers</div>
        </div>
        <div className="metric-card">
          <div className="metric-value">{workers.reduce((sum, w) => sum + w.current_tasks.length, 0)}</div>
          <div className="metric-label">Total Running Tasks</div>
        </div>
      </div>

      <div className="card">
        <h2 style={{ marginBottom: '1rem', color: '#4a5568' }}>Worker Status</h2>
        {workers.length === 0 ? (
          <div className="empty-state">
            <h3>No workers registered</h3>
            <p>Workers will appear here once they register with the scheduler.</p>
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Worker ID</th>
                <th>Address</th>
                <th>Task Types</th>
                <th>Status</th>
                <th>Current Tasks</th>
                <th>Last Seen</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {workers.map(worker => (
                <tr key={worker.id}>
                  <td style={{ fontWeight: '600' }}>{worker.id}</td>
                  <td style={{ color: '#718096' }}>{worker.address}</td>
                  <td>
                    <div style={{ display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
                      {worker.task_types.map(type => (
                        <span key={type} style={{
                          backgroundColor: '#edf2f7',
                          color: '#4a5568',
                          padding: '0.125rem 0.375rem',
                          borderRadius: '4px',
                          fontSize: '0.75rem'
                        }}>
                          {type}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td>
                    <span className={getStatusBadgeClass(worker.status, worker.last_heartbeat)}>
                      {getStatusText(worker.status, worker.last_heartbeat)}
                    </span>
                  </td>
                  <td>
                    <div style={{ color: '#718096' }}>
                      {worker.current_tasks.length === 0 ? (
                        'None'
                      ) : (
                        <div>
                          {worker.current_tasks.length} task{worker.current_tasks.length !== 1 ? 's' : ''}
                          <div style={{ fontSize: '0.8rem', marginTop: '0.25rem' }}>
                            {worker.current_tasks.slice(0, 2).map(taskId => (
                              <div key={taskId}>{taskId}</div>
                            ))}
                            {worker.current_tasks.length > 2 && (
                              <div>+{worker.current_tasks.length - 2} more</div>
                            )}
                          </div>
                        </div>
                      )}
                    </div>
                  </td>
                  <td style={{ color: '#718096' }}>{formatLastSeen(worker.last_heartbeat)}</td>
                  <td>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <button 
                        className="btn btn-primary" 
                        style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}
                      >
                        Details
                      </button>
                      {getStatusText(worker.status, worker.last_heartbeat) !== 'offline' && (
                        <button 
                          className="btn btn-danger" 
                          style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}
                        >
                          Drain
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default Workers;
