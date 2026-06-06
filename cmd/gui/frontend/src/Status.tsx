import { useEffect, useState } from 'react';
import { GetStatus } from '../wailsjs/go/main/App';
import { status } from '../wailsjs/go/models';
import { Cpu, HardDrive, MemoryStick, Battery, Activity, Network } from 'lucide-react';

export default function StatusView() {
  const [metrics, setMetrics] = useState<status.MetricsSnapshot | null>(null);

  useEffect(() => {
    let interval = setInterval(async () => {
      try {
        const data = await GetStatus();
        setMetrics(data);
      } catch (e) {
        console.error("Failed to get status:", e);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  if (!metrics) return <div style={{padding: '2rem'}}>Loading system metrics...</div>;

  const { cpu, memory, disks, batteries, hardware } = metrics;

  return (
    <div className="status-grid animate-fade-in">
      {/* Hardware Header */}
      <div className="status-card full-width">
        <h2 style={{margin: '0 0 1rem 0'}}>{hardware.model || 'System Info'}</h2>
        <div style={{display: 'flex', gap: '2rem', color: 'var(--text-muted)'}}>
          <div><strong>CPU:</strong> {hardware.cpu_model}</div>
          <div><strong>RAM:</strong> {hardware.total_ram}</div>
          <div><strong>OS:</strong> {hardware.os_version}</div>
        </div>
      </div>

      {/* CPU */}
      <div className="status-card">
        <h3 className="card-title"><Cpu size={18} /> CPU</h3>
        <div className="stat-row">
          <span>Usage</span>
          <span>{cpu.usage.toFixed(1)}%</span>
        </div>
        <div className="progress-bar-bg">
          <div className="progress-bar-fill" style={{width: `${cpu.usage}%`, background: cpu.usage > 80 ? 'var(--danger)' : 'var(--accent)'}}></div>
        </div>
        <div className="stat-row" style={{marginTop: '1rem'}}>
          <span>Load Avg</span>
          <span>{cpu.load1.toFixed(2)} / {cpu.load5.toFixed(2)} / {cpu.load15.toFixed(2)}</span>
        </div>
      </div>

      {/* Memory */}
      <div className="status-card">
        <h3 className="card-title"><MemoryStick size={18} /> Memory</h3>
        <div className="stat-row">
          <span>Used</span>
          <span>{memory.used_percent.toFixed(1)}%</span>
        </div>
        <div className="progress-bar-bg">
          <div className="progress-bar-fill" style={{width: `${memory.used_percent}%`, background: memory.used_percent > 80 ? 'var(--danger)' : 'var(--accent)'}}></div>
        </div>
        <div className="stat-row" style={{marginTop: '1rem', color: 'var(--text-muted)', fontSize: '0.9rem'}}>
          <span>Pressure: {memory.pressure || 'Normal'}</span>
        </div>
      </div>

      {/* Disks */}
      <div className="status-card">
        <h3 className="card-title"><HardDrive size={18} /> Disks</h3>
        {disks && disks.length > 0 ? disks.map((d: status.DiskStatus, i: number) => (
          <div key={i} style={{marginBottom: '1rem'}}>
            <div className="stat-row">
              <span>{d.device}</span>
              <span>{d.used_percent.toFixed(1)}%</span>
            </div>
            <div className="progress-bar-bg">
              <div className="progress-bar-fill" style={{width: `${d.used_percent}%`, background: d.used_percent > 85 ? 'var(--danger)' : 'var(--accent)'}}></div>
            </div>
          </div>
        )) : <div>No disk info</div>}
      </div>

      {/* Battery */}
      {batteries && batteries.length > 0 && (
        <div className="status-card">
          <h3 className="card-title"><Battery size={18} /> Battery</h3>
          <div className="stat-row">
            <span>Level</span>
            <span>{batteries[0].percent.toFixed(1)}% ({batteries[0].status})</span>
          </div>
          <div className="progress-bar-bg">
            <div className="progress-bar-fill" style={{width: `${batteries[0].percent}%`, background: batteries[0].percent < 20 ? 'var(--danger)' : 'var(--success)'}}></div>
          </div>
          <div className="stat-row" style={{marginTop: '1rem', color: 'var(--text-muted)', fontSize: '0.9rem'}}>
            <span>Cycles: {batteries[0].cycle_count}</span>
            <span>Health: {batteries[0].health || 'Unknown'}</span>
          </div>
        </div>
      )}

    </div>
  );
}
