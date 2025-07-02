import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';

const Workflows = () => {
  const [workflows, setWorkflows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');

  useEffect(() => {
    fetchWorkflows();
  }, []);

  const fetchWorkflows = async () => {
    try {
      const response = await axios.get('/api/v1/workflows');
      setWorkflows(response.data.workflows || []);
    } catch (error) {
      console.error('Failed to fetch workflows:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredWorkflows = workflows.filter(workflow => {
    const matchesSearch = workflow.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         workflow.description.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === 'all' || workflow.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'pending': return 'status-badge status-pending';
      case 'running': return 'status-badge status-running';
      case 'completed': return 'status-badge status-completed';
      case 'failed': return 'status-badge status-failed';
      case 'cancelled': return 'status-badge status-cancelled';
      default: return 'status-badge';
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString();
  };

  if (loading) {
    return <div className="loading">Loading workflows...</div>;
  }

  return (
    <div>
      <div className="workflow-header">
        <h1 style={{ color: '#2d3748' }}>Workflows</h1>
        <div className="workflow-actions">
          <button className="btn btn-primary">Create Workflow</button>
        </div>
      </div>

      <div className="card">
        <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
          <input
            type="text"
            placeholder="Search workflows..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
            style={{ flex: 1 }}
          />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            style={{ padding: '0.75rem', border: '1px solid #e2e8f0', borderRadius: '6px' }}
          >
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="running">Running</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>

        {filteredWorkflows.length === 0 ? (
          <div className="empty-state">
            <h3>No workflows found</h3>
            <p>Create your first workflow to get started.</p>
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Status</th>
                <th>Tasks</th>
                <th>Created</th>
                <th>Duration</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredWorkflows.map(workflow => (
                <tr key={workflow.id}>
                  <td>
                    <Link 
                      to={`/workflows/${workflow.id}`}
                      style={{ textDecoration: 'none', color: '#667eea', fontWeight: '600' }}
                    >
                      {workflow.name}
                    </Link>
                  </td>
                  <td style={{ color: '#718096' }}>{workflow.description}</td>
                  <td>
                    <span className={getStatusBadgeClass(workflow.status)}>
                      {workflow.status}
                    </span>
                  </td>
                  <td>{workflow.tasks?.length || 0}</td>
                  <td style={{ color: '#718096' }}>{formatDate(workflow.created_at)}</td>
                  <td style={{ color: '#718096' }}>
                    {workflow.completed_at && workflow.started_at
                      ? `${Math.round((new Date(workflow.completed_at) - new Date(workflow.started_at)) / 1000)}s`
                      : workflow.started_at
                      ? `${Math.round((new Date() - new Date(workflow.started_at)) / 1000)}s`
                      : '-'
                    }
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <Link to={`/workflows/${workflow.id}`} className="btn btn-primary" style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}>
                        View
                      </Link>
                      {workflow.status === 'running' && (
                        <button className="btn btn-danger" style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}>
                          Cancel
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

export default Workflows;
