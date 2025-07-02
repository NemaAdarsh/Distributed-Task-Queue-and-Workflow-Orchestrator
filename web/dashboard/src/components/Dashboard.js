import React, { useState, useEffect } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import axios from 'axios';

const Dashboard = () => {
  const [metrics, setMetrics] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 10000);
    return () => clearInterval(interval);
  }, []);

  const fetchMetrics = async () => {
    try {
      const response = await axios.get('/api/v1/metrics');
      setMetrics(response.data);
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="loading">Loading dashboard...</div>;
  }

  const workflowData = [
    { name: 'Running', value: metrics?.workflows?.running || 0, color: '#38b2ac' },
    { name: 'Completed', value: metrics?.workflows?.completed || 0, color: '#38a169' },
    { name: 'Failed', value: metrics?.workflows?.failed || 0, color: '#e53e3e' },
    { name: 'Pending', value: metrics?.workflows?.total - (metrics?.workflows?.running + metrics?.workflows?.completed + metrics?.workflows?.failed) || 0, color: '#d69e2e' }
  ];

  const taskData = [
    { name: 'Pending', value: metrics?.tasks?.pending || 0 },
    { name: 'Running', value: metrics?.tasks?.running || 0 },
    { name: 'Completed', value: metrics?.tasks?.completed || 0 },
    { name: 'Failed', value: metrics?.tasks?.failed || 0 }
  ];

  return (
    <div>
      <h1 style={{ marginBottom: '2rem', color: '#2d3748' }}>Dashboard</h1>
      
      <div className="metrics-grid">
        <div className="metric-card">
          <div className="metric-value">{metrics?.workflows?.total || 0}</div>
          <div className="metric-label">Total Workflows</div>
        </div>
        <div className="metric-card">
          <div className="metric-value">{metrics?.workflows?.running || 0}</div>
          <div className="metric-label">Running Workflows</div>
        </div>
        <div className="metric-card">
          <div className="metric-value">{metrics?.tasks?.total || 0}</div>
          <div className="metric-label">Total Tasks</div>
        </div>
        <div className="metric-card">
          <div className="metric-value">{metrics?.workers?.active || 0}</div>
          <div className="metric-label">Active Workers</div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem', marginTop: '2rem' }}>
        <div className="card">
          <h2 style={{ marginBottom: '1rem', color: '#4a5568' }}>Workflow Status Distribution</h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={workflowData}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={100}
                paddingAngle={5}
                dataKey="value"
              >
                {workflowData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
          <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem', marginTop: '1rem' }}>
            {workflowData.map((entry, index) => (
              <div key={index} style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <div style={{ width: '12px', height: '12px', backgroundColor: entry.color, borderRadius: '50%' }}></div>
                <span style={{ fontSize: '0.9rem', color: '#4a5568' }}>{entry.name}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="card">
          <h2 style={{ marginBottom: '1rem', color: '#4a5568' }}>Task Statistics</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={taskData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Bar dataKey="value" fill="#667eea" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="card" style={{ marginTop: '2rem' }}>
        <h2 style={{ marginBottom: '1rem', color: '#4a5568' }}>System Health</h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
          <div>
            <div style={{ fontSize: '1.2rem', fontWeight: '600', color: '#38a169' }}>Healthy</div>
            <div style={{ color: '#718096' }}>Scheduler Status</div>
          </div>
          <div>
            <div style={{ fontSize: '1.2rem', fontWeight: '600', color: '#38a169' }}>Connected</div>
            <div style={{ color: '#718096' }}>Database Status</div>
          </div>
          <div>
            <div style={{ fontSize: '1.2rem', fontWeight: '600', color: '#38a169' }}>Online</div>
            <div style={{ color: '#718096' }}>Queue Status</div>
          </div>
          <div>
            <div style={{ fontSize: '1.2rem', fontWeight: '600', color: '#d69e2e' }}>
              {metrics?.workers?.active || 0} / {(metrics?.workers?.active || 0) + (metrics?.workers?.idle || 0)}
            </div>
            <div style={{ color: '#718096' }}>Worker Utilization</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
