import { useMemo, useCallback, useId } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
  type NodeMouseHandler,
  MarkerType,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import type {
  Conflict,
  ModConflictSummary,
  ConflictSeverity,
} from '@/types/index.ts';

// ============================================
// Types
// ============================================

interface ConflictGraphViewProps {
  conflicts: Conflict[];
  modSummaries: ModConflictSummary[];
  onSelectMod: (modId: string) => void;
  selectedModId: string | null;
}

interface ModNode {
  modId: string;
  modName: string;
  totalConflicts: number;
  winCount: number;
  loseCount: number;
  criticalCount: number;
}

interface ModEdge {
  source: string;
  target: string;
  conflictCount: number;
  maxSeverity: ConflictSeverity;
  winner: string; // modId of the winner
}

// ============================================
// Helper Functions
// ============================================

const SEVERITY_ORDER: ConflictSeverity[] = ['critical', 'high', 'medium', 'low', 'info'];

function getMaxSeverity(a: ConflictSeverity, b: ConflictSeverity): ConflictSeverity {
  const indexA = SEVERITY_ORDER.indexOf(a);
  const indexB = SEVERITY_ORDER.indexOf(b);
  return indexA <= indexB ? a : b;
}

/**
 * Get color for node based on win/lose ratio and critical count.
 */
function getNodeColor(mod: ModNode): { bg: string; border: string; text: string } {
  if (mod.criticalCount > 0) {
    return { bg: 'rgba(239, 68, 68, 0.2)', border: '#ef4444', text: '#f87171' };
  }

  const ratio = mod.winCount / Math.max(mod.totalConflicts, 1);

  if (ratio >= 0.7) {
    // Mostly winning - green
    return { bg: 'rgba(34, 197, 94, 0.2)', border: '#22c55e', text: '#4ade80' };
  } else if (ratio >= 0.3) {
    // Mixed - yellow
    return { bg: 'rgba(234, 179, 8, 0.2)', border: '#eab308', text: '#facc15' };
  } else {
    // Mostly losing - orange
    return { bg: 'rgba(249, 115, 22, 0.2)', border: '#f97316', text: '#fb923c' };
  }
}

/**
 * Get edge color based on severity.
 */
function getEdgeColor(severity: ConflictSeverity): string {
  switch (severity) {
    case 'critical':
      return '#ef4444';
    case 'high':
      return '#f97316';
    case 'medium':
      return '#eab308';
    case 'low':
      return '#6b7280';
    case 'info':
      return '#3b82f6';
    default:
      return '#6b7280';
  }
}

/**
 * Build graph data from conflicts and mod summaries.
 */
function buildGraphData(
  conflicts: Conflict[],
  modSummaries: ModConflictSummary[],
): { modNodes: ModNode[]; modEdges: ModEdge[] } {
  // Build nodes from mod summaries
  const modNodes: ModNode[] = modSummaries.map((mod) => ({
    modId: mod.modId,
    modName: mod.modName,
    totalConflicts: mod.totalConflicts,
    winCount: mod.winCount,
    loseCount: mod.loseCount,
    criticalCount: mod.criticalCount,
  }));

  // Build edges from conflicts
  // Track edges between pairs of mods
  const edgeMap = new Map<string, ModEdge>();

  for (const conflict of conflicts) {
    if (!conflict.winner || conflict.losers.length === 0) continue;

    const winnerId = conflict.winner.modId;

    for (const loser of conflict.losers) {
      const loserId = loser.modId;

      // Create a unique key for this pair (sorted to avoid duplicates)
      const key = [winnerId, loserId].sort().join('-');

      const existing = edgeMap.get(key);
      if (existing) {
        existing.conflictCount++;
        existing.maxSeverity = getMaxSeverity(existing.maxSeverity, conflict.severity);
        // Winner is the one who wins more often
        if (winnerId === existing.source) {
          // Still same winner
        } else {
          // Check if we should flip
          // This is already handled by first edge creation
        }
      } else {
        edgeMap.set(key, {
          source: winnerId,
          target: loserId,
          conflictCount: 1,
          maxSeverity: conflict.severity,
          winner: winnerId,
        });
      }
    }
  }

  return {
    modNodes,
    modEdges: Array.from(edgeMap.values()),
  };
}

/**
 * Calculate node positions in a circular layout.
 */
function calculateNodePositions(
  modNodes: ModNode[],
): Node[] {
  const nodeCount = modNodes.length;
  const centerX = 400;
  const centerY = 300;
  const radius = Math.max(200, nodeCount * 30);

  return modNodes.map((mod, index) => {
    const angle = (2 * Math.PI * index) / nodeCount - Math.PI / 2;
    const x = centerX + radius * Math.cos(angle);
    const y = centerY + radius * Math.sin(angle);

    const colors = getNodeColor(mod);

    return {
      id: mod.modId,
      type: 'default',
      position: { x, y },
      data: {
        label: (
          <div className="text-center p-1">
            <div className="font-medium text-xs truncate max-w-[120px]" title={mod.modName}>
              {mod.modName}
            </div>
            <div className="text-[10px] opacity-70 mt-0.5">
              {mod.winCount}W / {mod.loseCount}L
            </div>
          </div>
        ),
      },
      style: {
        background: colors.bg,
        border: `2px solid ${colors.border}`,
        borderRadius: '8px',
        color: colors.text,
        width: 140,
        fontSize: '11px',
        padding: '4px',
      },
    };
  });
}

/**
 * Create edges from mod edges.
 */
function createEdges(modEdges: ModEdge[]): Edge[] {
  return modEdges.map((edge, index) => {
    const color = getEdgeColor(edge.maxSeverity);
    const strokeWidth = Math.min(1 + edge.conflictCount * 0.5, 6);

    return {
      id: `edge-${index}`,
      source: edge.source,
      target: edge.target,
      type: 'default',
      animated: edge.maxSeverity === 'critical' || edge.maxSeverity === 'high',
      label: `${edge.conflictCount}`,
      labelStyle: {
        fill: color,
        fontSize: '10px',
        fontWeight: 'bold',
      },
      labelBgStyle: {
        fill: '#1a1a2e',
        fillOpacity: 0.8,
      },
      style: {
        stroke: color,
        strokeWidth,
      },
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color,
        width: 15,
        height: 15,
      },
    };
  });
}

// ============================================
// Legend Component
// ============================================

const GraphLegend: React.FC = () => {
  const legendId = useId();

  return (
    <div
      id={legendId}
      className="absolute bottom-4 left-4 z-10 p-3 rounded-sm bg-bg-card/90 border border-border backdrop-blur-sm"
      aria-label="Graph legend"
    >
      <h4 className="text-xs font-semibold text-text-primary mb-2">Legend</h4>
      <div className="space-y-1.5 text-[10px]">
        <div className="font-medium text-text-muted mb-1">Node Colors (Win Rate)</div>
        <div className="flex items-center gap-2">
          <span className="w-3 h-3 rounded" style={{ backgroundColor: 'rgba(34, 197, 94, 0.5)' }} />
          <span className="text-text-secondary">High win rate (&gt;70%)</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-3 h-3 rounded" style={{ backgroundColor: 'rgba(234, 179, 8, 0.5)' }} />
          <span className="text-text-secondary">Mixed (30-70%)</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-3 h-3 rounded" style={{ backgroundColor: 'rgba(249, 115, 22, 0.5)' }} />
          <span className="text-text-secondary">Low win rate (&lt;30%)</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-3 h-3 rounded" style={{ backgroundColor: 'rgba(239, 68, 68, 0.5)' }} />
          <span className="text-text-secondary">Has critical conflicts</span>
        </div>

        <div className="font-medium text-text-muted mt-2 mb-1">Edge Colors (Severity)</div>
        <div className="flex items-center gap-2">
          <span className="w-6 h-0.5 rounded" style={{ backgroundColor: '#ef4444' }} />
          <span className="text-text-secondary">Critical</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-6 h-0.5 rounded" style={{ backgroundColor: '#f97316' }} />
          <span className="text-text-secondary">High</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-6 h-0.5 rounded" style={{ backgroundColor: '#eab308' }} />
          <span className="text-text-secondary">Medium</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-6 h-0.5 rounded" style={{ backgroundColor: '#6b7280' }} />
          <span className="text-text-secondary">Low</span>
        </div>
      </div>
    </div>
  );
};

// ============================================
// Stats Overlay Component
// ============================================

interface StatsOverlayProps {
  nodeCount: number;
  edgeCount: number;
}

const StatsOverlay: React.FC<StatsOverlayProps> = ({ nodeCount, edgeCount }) => (
  <div className="absolute top-4 left-4 z-10 p-2 rounded-sm bg-bg-card/90 border border-border backdrop-blur-sm">
    <div className="text-xs text-text-muted">
      <span className="font-medium text-text-primary">{nodeCount}</span> mods &middot;{' '}
      <span className="font-medium text-text-primary">{edgeCount}</span> conflict relationships
    </div>
  </div>
);

// ============================================
// Empty State Component
// ============================================

const EmptyGraphState: React.FC = () => (
  <div className="flex items-center justify-center h-full">
    <div className="text-center p-8">
      <div className="text-4xl mb-4" aria-hidden="true">🕸️</div>
      <h3 className="text-lg font-semibold text-text-primary mb-2">
        No Conflict Relationships
      </h3>
      <p className="text-text-muted max-w-md">
        There are no conflicts with clear winner/loser relationships to visualize.
        The list view shows all conflicts regardless.
      </p>
    </div>
  </div>
);

// ============================================
// Main ConflictGraphView Component
// ============================================

/**
 * Interactive graph visualization of mod conflict relationships.
 * Shows mods as nodes and conflicts as edges between them.
 */
export const ConflictGraphView: React.FC<ConflictGraphViewProps> = ({
  conflicts,
  modSummaries,
  onSelectMod,
  selectedModId,
}) => {
  // Build graph data
  const graphData = useMemo(() => {
    return buildGraphData(conflicts, modSummaries);
  }, [conflicts, modSummaries]);

  // Create React Flow nodes and edges
  const initialNodes = useMemo(() => {
    return calculateNodePositions(graphData.modNodes);
  }, [graphData.modNodes]);

  const initialEdges = useMemo(() => {
    return createEdges(graphData.modEdges);
  }, [graphData.modEdges]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, , onEdgesChange] = useEdgesState(initialEdges);

  // Update node styles when selection changes
  useMemo(() => {
    if (!selectedModId) return;

    setNodes((nds) =>
      nds.map((node) => {
        const isSelected = node.id === selectedModId;
        const isConnected = edges.some(
          (edge) =>
            (edge.source === selectedModId && edge.target === node.id) ||
            (edge.target === selectedModId && edge.source === node.id)
        );

        return {
          ...node,
          style: {
            ...node.style,
            opacity: isSelected || isConnected || !selectedModId ? 1 : 0.4,
            boxShadow: isSelected ? '0 0 10px rgba(59, 130, 246, 0.5)' : undefined,
          },
        };
      })
    );
  }, [selectedModId, edges, setNodes]);

  // Handle node click
  const handleNodeClick: NodeMouseHandler = useCallback(
    (_event, node) => {
      onSelectMod(node.id);
    },
    [onSelectMod]
  );

  // Empty state check
  if (graphData.modNodes.length === 0 || graphData.modEdges.length === 0) {
    return (
      <div className="h-[600px] rounded-sm border border-border bg-bg-card overflow-hidden">
        <EmptyGraphState />
      </div>
    );
  }

  return (
    <div className="h-[600px] rounded-sm border border-border bg-bg-card overflow-hidden relative">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={handleNodeClick}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.1}
        maxZoom={2}
        defaultEdgeOptions={{
          type: 'default',
        }}
      >
        <Background color="#374151" gap={20} size={1} />
        <Controls
          className="!bg-bg-card !border-border"
          showInteractive={false}
        />
        <MiniMap
          nodeColor={(node) => {
            const style = node.style as { border?: string } | undefined;
            return style?.border ?? '#6b7280';
          }}
          maskColor="rgba(0, 0, 0, 0.8)"
          className="!bg-bg-secondary !border-border"
        />
      </ReactFlow>

      <StatsOverlay nodeCount={nodes.length} edgeCount={edges.length} />
      <GraphLegend />
    </div>
  );
};
