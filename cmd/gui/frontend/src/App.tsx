import { useState, useEffect, useRef } from 'react';
import { RunClean, ListApps, RunUninstall, RunOptimize, RunPurge, CancelOperation, PreviewClean } from '../wailsjs/go/main/App';
import { platform, clean } from '../wailsjs/go/models';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { Shield, Trash2, Zap, Archive, CheckCircle, Loader2, X, AlertTriangle, Eye } from 'lucide-react';
import './style.css';

function App() {
  const [activeTab, setActiveTab] = useState('clean');
  const [logs, setLogs] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  
  const [apps, setApps] = useState<platform.AppInfo[]>([]);
  const [selectedApp, setSelectedApp] = useState('');

  const [previewData, setPreviewData] = useState<clean.PreviewEntry[] | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);

  const [confirmModal, setConfirmModal] = useState<{ isOpen: boolean, title: string, message: string, onConfirm: () => void }>({
    isOpen: false, title: '', message: '', onConfirm: () => {}
  });

  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Scroll to bottom of logs
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs]);

  useEffect(() => {
    const unsub = EventsOn('log', (msg: string) => {
      setLogs((prev) => {
        const newLogs = [...prev, msg];
        if (newLogs.length > 500) return newLogs.slice(newLogs.length - 500);
        return newLogs;
      });
    });
    return () => {
      unsub();
    };
  }, []);

  const addLog = (msg: string) => {
    setLogs((prev) => {
      const newLogs = [...prev, msg];
      if (newLogs.length > 500) return newLogs.slice(newLogs.length - 500);
      return newLogs;
    });
  };

  const handleAction = async (actionFn: (dryRun: boolean) => Promise<string>, name: string) => {
    setConfirmModal({ isOpen: false, title: '', message: '', onConfirm: () => {} });
    setLoading(true);
    setLogs([]);
    addLog(`=== Starting ${name} ===`);
    try {
      const result = await actionFn(false);
      addLog(`[System]: ${result}`);
    } catch (err) {
      addLog(`[Error]: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  const requestAction = (actionFn: (dryRun: boolean) => Promise<string>, name: string) => {
    setConfirmModal({
      isOpen: true,
      title: `Confirm ${name}`,
      message: `Are you sure you want to execute ${name}? This action is destructive and cannot be undone.`,
      onConfirm: () => handleAction(actionFn, name)
    });
  };

  const requestUninstall = () => {
    if (!selectedApp) return;
    setConfirmModal({
      isOpen: true,
      title: `Confirm Uninstall`,
      message: `Are you sure you want to uninstall ${selectedApp} and wipe all leftover app data?`,
      onConfirm: async () => {
        setConfirmModal({ isOpen: false, title: '', message: '', onConfirm: () => {} });
        setLoading(true);
        setLogs([]);
        addLog(`=== Starting uninstallation for ${selectedApp} ===`);
        try {
          const result = await RunUninstall(selectedApp, false);
          addLog(`[System]: ${result}`);
          fetchApps();
        } catch (err) {
          addLog(`[Error]: ${err}`);
        } finally {
          setLoading(false);
        }
      }
    });
  };

  const handlePreviewClean = async () => {
    setPreviewLoading(true);
    setPreviewData(null);
    try {
      const data = await PreviewClean();
      setPreviewData(data || []);
    } catch (err) {
      addLog(`[Error]: Preview failed: ${err}`);
    } finally {
      setPreviewLoading(false);
    }
  };

  const fetchApps = async () => {
    setLoading(true);
    try {
      const appList = await ListApps();
      setApps(appList);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async () => {
    await CancelOperation();
  };

  useEffect(() => {
    if (activeTab === 'uninstall') {
      fetchApps();
    }
    setPreviewData(null);
  }, [activeTab]);

  return (
    <div className="app-container animate-fade-in">
      {confirmModal.isOpen && (
        <div className="modal-overlay">
          <div className="modal-content animate-fade-in">
            <h2 className="flex items-center gap-2" style={{color: 'var(--danger)', margin: '0 0 1rem 0'}}>
              <AlertTriangle size={24} /> {confirmModal.title}
            </h2>
            <p>{confirmModal.message}</p>
            <div style={{ display: 'flex', gap: '1rem', marginTop: '2rem', justifyContent: 'flex-end' }}>
              <button className="btn" style={{ background: 'var(--bg-panel-hover)' }} onClick={() => setConfirmModal({ ...confirmModal, isOpen: false })}>Cancel</button>
              <button className="btn btn-danger" onClick={confirmModal.onConfirm}>Confirm Action</button>
            </div>
          </div>
        </div>
      )}

      <div className="sidebar">
        <div className="brand">
          <Shield className="text-accent" /> Mole
        </div>
        
        <div className={`nav-item ${activeTab === 'clean' ? 'active' : ''}`} onClick={() => setActiveTab('clean')}>
          <Trash2 size={18} /> Clean System
        </div>
        <div className={`nav-item ${activeTab === 'uninstall' ? 'active' : ''}`} onClick={() => setActiveTab('uninstall')}>
          <Archive size={18} /> Uninstaller
        </div>
        <div className={`nav-item ${activeTab === 'optimize' ? 'active' : ''}`} onClick={() => setActiveTab('optimize')}>
          <Zap size={18} /> Optimize
        </div>
        <div className={`nav-item ${activeTab === 'purge' ? 'active' : ''}`} onClick={() => setActiveTab('purge')}>
          <Archive size={18} /> Purge Projects
        </div>
      </div>

      <div className="main-content">
        <div className="header">
          <h1>
            {activeTab === 'clean' && 'System Cleanup'}
            {activeTab === 'uninstall' && 'App Uninstaller'}
            {activeTab === 'optimize' && 'System Optimization'}
            {activeTab === 'purge' && 'Project Purge'}
          </h1>
          <p>
            {activeTab === 'clean' && 'Deep clean browser caches, temps, and developer leftovers.'}
            {activeTab === 'uninstall' && 'Smart application uninstaller that cleans leftover data.'}
            {activeTab === 'optimize' && 'Optimize Windows system components and networks.'}
            {activeTab === 'purge' && 'Remove heavy project artifacts like node_modules.'}
          </p>
        </div>

        <div className="card" style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
            {activeTab === 'clean' && (
              <>
                {!loading ? (
                  <>
                    <button className="btn btn-danger" onClick={() => requestAction(RunClean, 'System Cleanup')}>
                      <Trash2 size={18} /> Start Cleanup
                    </button>
                    <button className="btn" style={{ background: 'var(--bg-panel-hover)' }} onClick={handlePreviewClean} disabled={previewLoading}>
                      {previewLoading ? <Loader2 className="animate-spin" size={18} /> : <Eye size={18} />} Preview Items
                    </button>
                  </>
                ) : (
                  <button className="btn" style={{ background: 'var(--bg-panel-hover)', color: 'var(--danger)' }} onClick={handleCancel}>
                    <X size={18} /> Stop Operation
                  </button>
                )}
              </>
            )}

            {activeTab === 'uninstall' && (
              <>
                {!loading ? (
                  <button className="btn btn-danger" onClick={requestUninstall} disabled={!selectedApp}>
                    <Archive size={18} /> Uninstall Selected
                  </button>
                ) : (
                  <button className="btn" style={{ background: 'var(--bg-panel-hover)', color: 'var(--danger)' }} onClick={handleCancel}>
                    <X size={18} /> Stop Operation
                  </button>
                )}
              </>
            )}

            {activeTab === 'optimize' && (
              <>
                {!loading ? (
                  <button className="btn" onClick={() => requestAction(RunOptimize, 'Optimization')}>
                    <Zap size={18} /> Start Optimization
                  </button>
                ) : (
                  <button className="btn" style={{ background: 'var(--bg-panel-hover)', color: 'var(--danger)' }} onClick={handleCancel}>
                    <X size={18} /> Stop Operation
                  </button>
                )}
              </>
            )}

            {activeTab === 'purge' && (
              <>
                {!loading ? (
                  <button className="btn btn-danger" onClick={() => requestAction(RunPurge, 'Project Purge')}>
                    <CheckCircle size={18} /> Start Purge
                  </button>
                ) : (
                  <button className="btn" style={{ background: 'var(--bg-panel-hover)', color: 'var(--danger)' }} onClick={handleCancel}>
                    <X size={18} /> Stop Operation
                  </button>
                )}
              </>
            )}
          </div>

          {activeTab === 'uninstall' && apps.length > 0 && !loading && (
            <div className="app-list" style={{ maxHeight: '300px', overflowY: 'auto' }}>
              {apps.map((app, idx) => (
                <div key={idx} className={`app-item ${selectedApp === app.Name ? 'selected' : ''}`} onClick={() => setSelectedApp(app.Name)}>
                  <div>
                    <strong>{app.Name}</strong>
                    <div style={{fontSize: '0.8rem', color: 'var(--text-muted)'}}>{app.Path}</div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {activeTab === 'clean' && previewData !== null && !loading && (
            <div style={{ background: 'rgba(0,0,0,0.3)', padding: '1rem', borderRadius: '0.5rem' }}>
              <h3 style={{ margin: '0 0 1rem 0' }}>Preview Results</h3>
              {previewData.length === 0 ? <p>No items found.</p> : (
                <div style={{ maxHeight: '300px', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {previewData.map((item, idx) => (
                    <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', borderBottom: '1px solid var(--border-color)' }}>
                      <span style={{ fontSize: '0.9rem' }}>{item.Path}</span>
                      <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', whiteSpace: 'nowrap' }}>{(item.Size / 1024 / 1024).toFixed(2)} MB</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {(logs.length > 0 || loading) && (
            <div className="output-console animate-fade-in">
              {logs.map((log, idx) => <div key={idx}>{log}</div>)}
              {loading && <div className="flex items-center gap-2 mt-2" style={{color: 'var(--text-muted)'}}><Loader2 className="animate-spin" size={14} /> Processing...</div>}
              <div ref={logsEndRef} />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
