import { useMemo, useCallback, useEffect } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  Panel,
  MarkerType,
  useReactFlow,
} from '@xyflow/react';
import type { Node, Edge } from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import type {
  LoadOrderPluginInfo,
  LoadOrderPluginType,
  LoadOrderIssue,
} from '@/types/index.ts';

// ============================================
// Types
// ============================================

interface DependencyGraphViewProps {
  plugins: LoadOrderPluginInfo[];
  dependencyGraph: Record<string, string[]>;
  issues: LoadOrderIssue[];
  onSelectPlugin: (plugin: LoadOrderPluginInfo | null) => void;
  selectedPlugin: LoadOrderPluginInfo | null;
}

interface PluginNodeData extends Record<string, unknown> {
  label: string;
  type: LoadOrderPluginType;
  index: number;
  hasIssues: boolean;
  issueCount: number;
  masterCount: number;
  dependentCount: number;
}

// ============================================
// Constants
// ============================================

const NODE_WIDTH = 180;
const NODE_HEIGHT = 60;
const HORIZONTAL_SPACING = 250;
const VERTICAL_SPACING = 100;

// Plugin type colors matching the existing theme
const TYPE_COLORS: Record<LoadOrderPluginType, { bg: string; border: string; text: string }> = {
  ESM: {
    bg: 'rgba(245, 158, 11, 0.2)',
    border: '#f59e0b',
    text: '#f59e0b',
  },
  ESP: {
    bg: 'rgba(107, 114, 128, 0.2)',
    border: '#6b7280',
    text: '#9ca3af',
  },
  ESL: {
    bg: 'rgba(234, 179, 8, 0.2)',
    border: '#eab308',
    text: '#eab308',
  },
};

// ============================================
// Helper Functions
// ============================================

/** Calculate node positions using a layered layout algorithm */
function calculateLayout(
  plugins: LoadOrderPluginInfo[],
  dependencyGraph: Record<string, string[]>
): Map<string, { x: number; y: number }> {
  const positions = new Map<string, { x: number; y: number }>();
  const pluginMap = new Map(plugins.map(p => [p.filename.toLowerCase(), p]));

  // Build reverse dependency map (what depends on each plugin)
  const dependents = new Map<string, string[]>();
  for (const [plugin, masters] of Object.entries(dependencyGraph)) {
    for (const master of masters) {
      const existing = dependents.get(master.toLowerCase()) ?? [];
      existing.push(plugin);
      dependents.set(master.toLowerCase(), existing);
    }
  }

  // Calculate layer for each plugin (how many steps from root)
  const layers = new Map<string, number>();

  function getLayer(filename: string, visited: Set<string> = new Set()): number {
    const key = filename.toLowerCase();
    if (layers.has(key)) return layers.get(key)!;
    if (visited.has(key)) return 0; // Circular dependency fallback

    visited.add(key);
    const masters = dependencyGraph[filename] ?? [];

    if (masters.length === 0) {
      layers.set(key, 0);
      return 0;
    }

    const maxMasterLayer = Math.max(
      ...masters.map(m => {
        const masterPlugin = pluginMap.get(m.toLowerCase());
        if (masterPlugin) {
          return getLayer(masterPlugin.filename, visited);
        }
        return -1; // Missing master
      })
    );

    const layer = maxMasterLayer + 1;
    layers.set(key, layer);
    return layer;
  }

  // Calculate layers for all plugins
  for (const plugin of plugins) {
    getLayer(plugin.filename);
  }

  // Group plugins by layer
  const layerGroups = new Map<number, LoadOrderPluginInfo[]>();
  for (const plugin of plugins) {
    const layer = layers.get(plugin.filename.toLowerCase()) ?? 0;
    const existing = layerGroups.get(layer) ?? [];
    existing.push(plugin);
    layerGroups.set(layer, existing);
  }

  // Position nodes within each layer
  for (const [layer, layerPlugins] of layerGroups) {
    // Sort by index within layer for consistency
    layerPlugins.sort((a, b) => a.index - b.index);

    const layerWidth = layerPlugins.length * HORIZONTAL_SPACING;
    const startX = -(layerWidth / 2) + HORIZONTAL_SPACING / 2;

    layerPlugins.forEach((plugin, idx) => {
      positions.set(plugin.filename.toLowerCase(), {
        x: startX + idx * HORIZONTAL_SPACING,
        y: layer * VERTICAL_SPACING,
      });
    });
  }

  return positions;
}

/** Build React Flow nodes from plugin data */
function buildNodes(
  plugins: LoadOrderPluginInfo[],
  dependencyGraph: Record<string, string[]>,
  positions: Map<string, { x: number; y: number }>,
  issues: LoadOrderIssue[],
  selectedPlugin: LoadOrderPluginInfo | null
): Node<PluginNodeData>[] {
  const issueMap = new Map<string, number>();
  for (const issue of issues) {
    issueMap.set(issue.plugin.toLowerCase(), (issueMap.get(issue.plugin.toLowerCase()) ?? 0) + 1);
  }

  // Build reverse dependency map
  const dependents = new Map<string, number>();
  for (const masters of Object.values(dependencyGraph)) {
    for (const master of masters) {
      dependents.set(master.toLowerCase(), (dependents.get(master.toLowerCase()) ?? 0) + 1);
    }
  }

  return plugins.map((plugin) => {
    const position = positions.get(plugin.filename.toLowerCase()) ?? { x: 0, y: 0 };
    const colors = TYPE_COLORS[plugin.type];
    const issueCount = issueMap.get(plugin.filename.toLowerCase()) ?? 0;
    const isSelected = selectedPlugin?.filename === plugin.filename;

    return {
      id: plugin.filename,
      type: 'default',
      position,
      data: {
        label: plugin.filename,
        type: plugin.type,
        index: plugin.index,
        hasIssues: issueCount > 0,
        issueCount,
        masterCount: (dependencyGraph[plugin.filename] ?? plugin.masters).length,
        dependentCount: dependents.get(plugin.filename.toLowerCase()) ?? 0,
      },
      style: {
        width: NODE_WIDTH,
        height: NODE_HEIGHT,
        background: isSelected ? colors.border : colors.bg,
        border: `2px solid ${issueCount > 0 ? '#ef4444' : colors.border}`,
        borderRadius: '8px',
        color: isSelected ? '#1f2937' : colors.text,
        fontSize: '12px',
        fontWeight: 500,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '8px',
        boxShadow: isSelected ? '0 0 0 3px rgba(96, 165, 250, 0.5)' : 'none',
      },
    };
  });
}

/** Build React Flow edges from dependency graph */
function buildEdges(
  plugins: LoadOrderPluginInfo[],
  dependencyGraph: Record<string, string[]>,
  issues: LoadOrderIssue[]
): Edge[] {
  const edges: Edge[] = [];
  const pluginSet = new Set(plugins.map(p => p.filename.toLowerCase()));

  // Build issue map for coloring edges
  const wrongOrderIssues = new Set<string>();
  const missingMasterIssues = new Set<string>();

  for (const issue of issues) {
    if (issue.type === 'wrong_order' && issue.relatedPlugin) {
      wrongOrderIssues.add(`${issue.relatedPlugin.toLowerCase()}-${issue.plugin.toLowerCase()}`);
    }
    if (issue.type === 'missing_master' && issue.relatedPlugin) {
      missingMasterIssues.add(`${issue.relatedPlugin.toLowerCase()}-${issue.plugin.toLowerCase()}`);
    }
  }

  for (const plugin of plugins) {
    const masters = dependencyGraph[plugin.filename] ?? plugin.masters;

    for (const master of masters) {
      const masterLower = master.toLowerCase();
      const edgeKey = `${masterLower}-${plugin.filename.toLowerCase()}`;

      // Determine edge color based on issues
      let strokeColor = '#4b5563'; // Default gray
      let animated = false;

      if (wrongOrderIssues.has(edgeKey)) {
        strokeColor = '#ef4444'; // Red for wrong order
        animated = true;
      } else if (!pluginSet.has(masterLower)) {
        strokeColor = '#f59e0b'; // Orange for missing master
        animated = true;
      }

      // Only create edge if master exists in plugins (or show as broken)
      if (pluginSet.has(masterLower)) {
        edges.push({
          id: `${master}-${plugin.filename}`,
          source: plugins.find(p => p.filename.toLowerCase() === masterLower)?.filename ?? master,
          target: plugin.filename,
          animated,
          style: {
            stroke: strokeColor,
            strokeWidth: 2,
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: strokeColor,
          },
        });
      }
    }
  }

  return edges;
}

// ============================================
// Graph Controls Component
// ============================================

interface GraphControlsProps {
  pluginCount: number;
  edgeCount: number;
  onFitView: () => void;
}

const GraphControls: React.FC<GraphControlsProps> = ({
  pluginCount,
  edgeCount,
  onFitView,
}) => (
  <div className="flex items-center gap-4 p-3 rounded-sm bg-bg-card border border-border">
    <div className="text-sm text-text-secondary">
      <span className="font-medium text-text-primary">{pluginCount}</span> plugins
    </div>
    <div className="text-sm text-text-secondary">
      <span className="font-medium text-text-primary">{edgeCount}</span> dependencies
    </div>
    <button
      onClick={onFitView}
      className="min-h-9 px-3 py-1 rounded-sm text-sm
        bg-bg-secondary text-text-primary
        hover:bg-bg-hover
        focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
        transition-colors motion-reduce:transition-none"
    >
      Fit View
    </button>
  </div>
);

// ============================================
// Legend Component
// ============================================

const GraphLegend: React.FC = () => (
  <div className="flex flex-wrap gap-4 p-3 rounded-sm bg-bg-card border border-border text-xs">
    <div className="flex items-center gap-2">
      <div
        className="w-4 h-4 rounded-xs"
        style={{ background: TYPE_COLORS.ESM.bg, border: `1px solid ${TYPE_COLORS.ESM.border}` }}
      />
      <span className="text-text-secondary">Master (ESM)</span>
    </div>
    <div className="flex items-center gap-2">
      <div
        className="w-4 h-4 rounded-xs"
        style={{ background: TYPE_COLORS.ESP.bg, border: `1px solid ${TYPE_COLORS.ESP.border}` }}
      />
      <span className="text-text-secondary">Plugin (ESP)</span>
    </div>
    <div className="flex items-center gap-2">
      <div
        className="w-4 h-4 rounded-xs"
        style={{ background: TYPE_COLORS.ESL.bg, border: `1px solid ${TYPE_COLORS.ESL.border}` }}
      />
      <span className="text-text-secondary">Light (ESL)</span>
    </div>
    <div className="flex items-center gap-2">
      <div className="w-4 h-0.5 bg-error" />
      <span className="text-text-secondary">Issue</span>
    </div>
  </div>
);

// ============================================
// Internal Graph Component (needs ReactFlowProvider context)
// ============================================

interface InternalGraphProps {
  plugins: LoadOrderPluginInfo[];
  dependencyGraph: Record<string, string[]>;
  issues: LoadOrderIssue[];
  onSelectPlugin: (plugin: LoadOrderPluginInfo | null) => void;
  selectedPlugin: LoadOrderPluginInfo | null;
}

const InternalGraph: React.FC<InternalGraphProps> = ({
  plugins,
  dependencyGraph,
  issues,
  onSelectPlugin,
  selectedPlugin,
}) => {
  const { fitView } = useReactFlow();

  // Calculate positions and build graph elements
  const { initialNodes, initialEdges } = useMemo(() => {
    const positions = calculateLayout(plugins, dependencyGraph);
    const nodes = buildNodes(plugins, dependencyGraph, positions, issues, selectedPlugin);
    const edges = buildEdges(plugins, dependencyGraph, issues);
    return { initialNodes: nodes, initialEdges: edges };
  }, [plugins, dependencyGraph, issues, selectedPlugin]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Update nodes when selection changes
  useEffect(() => {
    setNodes(initialNodes);
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  // Handle node click
  const onNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      const plugin = plugins.find(p => p.filename === node.id);
      onSelectPlugin(plugin ?? null);
    },
    [plugins, onSelectPlugin]
  );

  // Handle pane click (deselect)
  const onPaneClick = useCallback(() => {
    onSelectPlugin(null);
  }, [onSelectPlugin]);

  // Fit view on mount
  useEffect(() => {
    const timer = setTimeout(() => fitView({ padding: 0.2 }), 100);
    return () => clearTimeout(timer);
  }, [fitView]);

  const handleFitView = useCallback(() => {
    fitView({ padding: 0.2, duration: 300 });
  }, [fitView]);

  return (
    <div
      className="w-full h-[600px] rounded-sm bg-bg-card border border-border overflow-hidden"
      role="img"
      aria-label={`Dependency graph showing ${plugins.length} plugins and their relationships`}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        fitView
        minZoom={0.1}
        maxZoom={2}
        defaultEdgeOptions={{
          type: 'smoothstep',
        }}
        proOptions={{ hideAttribution: true }}
      >
        <Background color="#374151" gap={20} />
        <Controls
          showInteractive={false}
          className="!bg-bg-card !border-border !shadow-none"
        />
        <MiniMap
          nodeColor={(node) => {
            const data = node.data as PluginNodeData;
            return TYPE_COLORS[data.type]?.border ?? '#6b7280';
          }}
          className="!bg-bg-secondary !border-border"
          maskColor="rgba(0, 0, 0, 0.6)"
        />
        <Panel position="top-left">
          <GraphControls
            pluginCount={plugins.length}
            edgeCount={edges.length}
            onFitView={handleFitView}
          />
        </Panel>
        <Panel position="bottom-left">
          <GraphLegend />
        </Panel>
      </ReactFlow>
    </div>
  );
};

// ============================================
// Main Component with Provider
// ============================================

import { ReactFlowProvider } from '@xyflow/react';

/** Displays plugin dependencies as an interactive graph */
export const DependencyGraphView: React.FC<DependencyGraphViewProps> = (props) => {
  // Show message if no plugins
  if (props.plugins.length === 0) {
    return (
      <div className="p-6 rounded-sm bg-bg-card border border-border text-center">
        <p className="text-text-secondary">No plugins to display in graph view.</p>
      </div>
    );
  }

  return (
    <ReactFlowProvider>
      <InternalGraph {...props} />
    </ReactFlowProvider>
  );
};
