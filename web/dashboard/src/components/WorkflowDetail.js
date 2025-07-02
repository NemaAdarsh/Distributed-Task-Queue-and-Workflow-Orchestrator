import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import axios from 'axios';

const WorkflowDetail = () => {
  const { id } = useParams();
  const [workflow, setWorkflow] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchWorkflow();
  }, [id]);

  const fetchWorkflow = async () => {
    try {
      const response = await axios.get(`/api/v1/workflows/${id}`);
      setWorkflow(response.data);
    } catch (error) {
      console.error('Failed to fetch workflow:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'pending': return 'status-badge status-pending';
      case 'running': return 'status-badge status-running';
      case 'completed': return 'status-badge status-completed';
      case 'failed': return 'status-badge status-failed';
      case 'cancelled': return 'status-badge status-cancelled';
      case 'retrying': return 'status-badge status-retrying';
      default: return 'status-badge';
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleString();
  };

  const formatDuration = (start, end) => {
    if (!start) return '-';
    const startTime = new Date(start);
    const endTime = end ? new Date(end) : new Date();
    const duration = Math.round((endTime - startTime) / 1000);
    return `${duration}s`;
  };

  if (loading) {
    return <div className="loading">Loading workflow...</div>;
  }

  if (!workflow) {
    return (
      <div className="empty-state">
        <h3>Workflow not found</h3>
        <Link to="/workflows" className="btn btn-primary">Back to Workflows</Link>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '2rem' }}>
        <Link to="/workflows" style={{ textDecoration: 'none', color: '#667eea' }}>← Back</Link>
        <h1 style={{ color: '#2d3748' }}>{workflow.name}</h1>
        <span className={getStatusBadgeClass(workflow.status)}>
          {workflow.status}
        </span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '2rem' }}>
        <div>
          <div className="card" style={{ marginBottom: '2rem' }}>
            <h2 style={{ marginBottom: '1rem', color: '#4a5568' }}>Workflow Information</h2>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div>
                <strong>Description:</strong>
                <div style={{ color: '#718096', marginTop: '0.25rem' }}>{workflow.description || 'No description'}</div>
              </div>
              <div>
                <strong>Created:</strong>
                <div style={{ color: '#718096', marginTop: '0.25rem' }}>{formatDate(workflow.created_at)}</div>
              </div>
              <div>
                <strong>Started:</strong>
                <div style={{ color: '#718096', marginTop: '0.25rem' }}>{formatDate(workflow.started_at)}</div>
              </div>
              <div>
                <strong>Duration:</strong>
                <div style={{ color: '#718096', marginTop: '0.25rem' }}>
                  {formatDuration(workflow.started_at, workflow.completed_at)}
                </div>
              </div>
            </div>
          </div>

          <div className="card">
            <h2 style={{ marginBottom: '1rem', color: '#4a5568' }}>Tasks ({workflow.tasks?.length || 0})</h2>
            {!workflow.tasks || workflow.tasks.length === 0 ? (
              <div className="empty-state">
                <p>No tasks in this workflow</p>
              </div>
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Status</th>
                    <th>Priority</th>
                    <th>Retries</th>
                    <th>Duration</th>
                    <th>Dependencies</th>
                  </tr>
                </thead>
                <tbody>
                  {workflow.tasks.map(task => (
                    <tr key={task.id}>
                      <td style={{ fontWeight: '600' }}>{task.name}</td>
                      <td>
                        <span style={{ 
                          backgroundColor: '#edf2f7', 
                          color: '#4a5568',
                          padding: '0.25rem 0.5rem',
                          borderRadius: '4px',
                          fontSize: '0.8rem'
                        }}>
                          {task.type}
                        </span>
                      </td>
                      <td>
                        <span className={getStatusBadgeClass(task.status)}>
                          {task.status}
                        </span>
                      </td>
                      <td>{task.priority}</td>
                      <td>{task.retry_count}/{task.max_retries}</td>
                      <td style={{ color: '#718096' }}>
                        {formatDuration(task.started_at, task.completed_at)}
                      </td>
                      <td style={{ color: '#718096' }}>
                        {task.dependencies?.length > 0 ? task.dependencies.join(', ') : 'None'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>

        <div>
          <div className="card" style={{ marginBottom: '2rem' }}>
            <h3 style={{ marginBottom: '1rem', color: '#4a5568' }}>Configuration</h3>
            <div style={{ fontSize: '0.9rem' }}>
              <div style={{ marginBottom: '0.5rem' }}>
                <strong>Max Concurrency:</strong> {workflow.config?.max_concurrency || 10}
              </div>
              <div style={{ marginBottom: '0.5rem' }}>
                <strong>Timeout:</strong> {workflow.config?.timeout || '1h'}
              </div>
              <div style={{ marginBottom: '0.5rem' }}>
                <strong>Max Retries:</strong> {workflow.config?.retry_policy?.max_attempts || 3}
              </div>
              <div>
                <strong>Backoff Factor:</strong> {workflow.config?.retry_policy?.backoff_factor || 2.0}
              </div>
            </div>
          </div>

          <div className="card">
            <h3 style={{ marginBottom: '1rem', color: '#4a5568' }}>Progress</h3>
            {workflow.tasks && workflow.tasks.length > 0 ? (
              <div>
                {['pending', 'running', 'completed', 'failed', 'retrying'].map(status => {
                  const count = workflow.tasks.filter(task => task.status === status).length;
                  const percentage = (count / workflow.tasks.length) * 100;
                  
                  return (
                    <div key={status} style={{ marginBottom: '0.75rem' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                        <span style={{ textTransform: 'capitalize', fontSize: '0.9rem' }}>{status}</span>
                        <span style={{ fontSize: '0.9rem' }}>{count}</span>
                      </div>
                      <div style={{ 
                        width: '100%', 
                        height: '8px', 
                        backgroundColor: '#edf2f7', 
                        borderRadius: '4px',
                        overflow: 'hidden'
                      }}>
                        <div style={{ 
                          width: `${percentage}%`, 
                          height: '100%', 
                          backgroundColor: status === 'completed' ? '#38a169' : 
                                         status === 'failed' ? '#e53e3e' :
                                         status === 'running' ? '#38b2ac' :
                                         status === 'retrying' ? '#d69e2e' : '#cbd5e0',
                          transition: 'width 0.3s ease'
                        }}></div>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <p style={{ color: '#718096' }}>No tasks to show progress</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkflowDetail;
