import { useState, useMemo, useRef, useEffect } from 'react';
import { AnalyzeDisk } from '../wailsjs/go/main/App';
import { analyze } from '../wailsjs/go/models';
import { HardDrive, Loader2, Play } from 'lucide-react';
import * as d3 from 'd3-hierarchy';

export default function AnalyzeView() {
  const [target, setTarget] = useState('C:\\');
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<analyze.Node | null>(null);
  
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height: 400 });

  useEffect(() => {
    if (containerRef.current) {
      setDimensions({
        width: containerRef.current.clientWidth,
        height: Math.max(400, containerRef.current.clientHeight - 60)
      });
    }
    const handleResize = () => {
      if (containerRef.current) {
        setDimensions({
          width: containerRef.current.clientWidth,
          height: Math.max(400, containerRef.current.clientHeight - 60)
        });
      }
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const handleScan = async () => {
    setLoading(true);
    setData(null);
    try {
      const result = await AnalyzeDisk(target);
      setData(result);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
  };

  // Generate Treemap Nodes
  const treeNodes = useMemo(() => {
    if (!data) return [];
    
    // We only want to visualize children of the root directory for a cleaner map, 
    // or just the root itself if it has no children.
    const rootHierarchy = d3.hierarchy<analyze.Node>(data, d => d.children || [])
      .sum(d => d.size || 0)
      .sort((a, b) => (b.value || 0) - (a.value || 0));

    const treemap = d3.treemap<analyze.Node>()
      .size([dimensions.width, dimensions.height])
      .padding(2)
      .round(true);

    const root = treemap(rootHierarchy);
    
    // Return all descendants except the root node itself to show its contents
    return root.leaves();
  }, [data, dimensions]);

  return (
    <div className="animate-fade-in" style={{display: 'flex', flexDirection: 'column', gap: '1rem', height: '100%'}}>
      <div className="card" style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
        <input 
          type="text" 
          value={target} 
          onChange={(e) => setTarget(e.target.value)}
          className="input"
          style={{ flex: 1 }}
          placeholder="Path to scan (e.g. C:\ or /Users/...)"
        />
        <button className="btn" onClick={handleScan} disabled={loading}>
          {loading ? <Loader2 className="animate-spin" size={18} /> : <Play size={18} />} Scan
        </button>
      </div>

      <div className="card" style={{ flex: 1, display: 'flex', flexDirection: 'column' }} ref={containerRef}>
        <h2 style={{margin: '0 0 1rem 0', display: 'flex', alignItems: 'center', gap: '0.5rem'}}>
          <HardDrive size={20} /> Disk Usage Map
        </h2>
        {loading && <div style={{color: 'var(--text-muted)'}}>Scanning directory structure... this may take a moment.</div>}
        {!loading && !data && <div style={{color: 'var(--text-muted)'}}>Enter a path and click scan to analyze disk usage.</div>}
        {!loading && data && treeNodes.length > 0 && (
          <div style={{ position: 'relative', width: dimensions.width, height: dimensions.height, background: 'rgba(0,0,0,0.2)', borderRadius: '0.5rem', overflow: 'hidden' }}>
            {treeNodes.map((node, i) => {
              const width = node.x1 - node.x0;
              const height = node.y1 - node.y0;
              // Don't render tiny boxes
              if (width < 3 || height < 3) return null;
              
              // Calculate a color based on size (larger = more red, smaller = blue/accent)
              const hue = Math.max(0, 220 - ((node.value || 0) / (data.size || 1)) * 220);
              
              return (
                <div
                  key={node.data.path || i}
                  title={`${node.data.name}\n${formatSize(node.value || 0)}`}
                  style={{
                    position: 'absolute',
                    left: node.x0,
                    top: node.y0,
                    width,
                    height,
                    background: `hsl(${hue}, 70%, 50%)`,
                    border: '1px solid rgba(255,255,255,0.1)',
                    boxSizing: 'border-box',
                    padding: '4px',
                    overflow: 'hidden',
                    color: 'white',
                    fontSize: '0.75rem',
                    transition: 'all 0.2s ease',
                    cursor: 'crosshair',
                    opacity: 0.85
                  }}
                  onMouseOver={(e) => {
                    (e.currentTarget as HTMLDivElement).style.opacity = '1';
                    (e.currentTarget as HTMLDivElement).style.zIndex = '10';
                    (e.currentTarget as HTMLDivElement).style.border = '2px solid white';
                  }}
                  onMouseOut={(e) => {
                    (e.currentTarget as HTMLDivElement).style.opacity = '0.85';
                    (e.currentTarget as HTMLDivElement).style.zIndex = '1';
                    (e.currentTarget as HTMLDivElement).style.border = '1px solid rgba(255,255,255,0.1)';
                  }}
                >
                  {width > 50 && height > 20 && (
                    <div style={{whiteSpace: 'nowrap', textOverflow: 'ellipsis', overflow: 'hidden', fontWeight: 'bold'}}>
                      {node.data.name}
                    </div>
                  )}
                  {width > 60 && height > 35 && (
                    <div style={{opacity: 0.8}}>{formatSize(node.value || 0)}</div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
