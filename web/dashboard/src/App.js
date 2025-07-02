import React from 'react';
import { BrowserRouter as Router, Routes, Route, NavLink } from 'react-router-dom';
import Dashboard from './components/Dashboard';
import Workflows from './components/Workflows';
import WorkflowDetail from './components/WorkflowDetail';
import Workers from './components/Workers';
import './App.css';

function App() {
  return (
    <Router>
      <div className="App">
        <nav className="navbar">
          <div className="nav-brand">
            <h1>FlowCtl</h1>
            <span>Workflow Orchestrator</span>
          </div>
          <div className="nav-links">
            <NavLink to="/" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Dashboard
            </NavLink>
            <NavLink to="/workflows" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Workflows
            </NavLink>
            <NavLink to="/workers" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Workers
            </NavLink>
          </div>
        </nav>

        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/workflows" element={<Workflows />} />
            <Route path="/workflows/:id" element={<WorkflowDetail />} />
            <Route path="/workers" element={<Workers />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
