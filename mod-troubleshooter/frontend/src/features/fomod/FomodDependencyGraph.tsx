import { useMemo, useCallback, useEffect, useId } from 'react';
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
  ReactFlowProvider,
} from '@xyflow/react';
import type { Node, Edge } from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import type {
  FomodData,
  Dependency,
} from '@/types/index.ts';

// ============================================
// Types
// ============================================

interface FomodDependencyGraphProps {
  data: FomodData;
  onNavigateToStep?: (stepIndex: number) => void;
}

type FomodNodeType = 'step' | 'group' | 'plugin' | 'flag' | 'conditional';

interface FomodNodeData extends Record<string, unknown> {
  label: string;
  nodeType: FomodNodeType;
  stepIndex?: number;
  pluginType?: string;
  flagValue?: string;
}

interface FlagInfo {
  name: string;
  value: string;
  source: { stepName: string; groupName: string; pluginName: string };
}

// ============================================
// Constants
// ============================================

const NODE_WIDTH = 180;
const NODE_HEIGHT = 50;
const HORIZONTAL_SPACING = 220;
const VERTICAL_SPACING = 80;

// Node type colors
const NODE_COLORS: Record<FomodNodeType, { bg: string; border: string; text: string }> = {
  step: {
    bg: 'rgba(59, 130, 246, 0.2)',
    border: '#3b82f6',
    text: '#60a5fa',
  },
  group: {
    bg: 'rgba(139, 92, 246, 0.2)',
    border: '#8b5cf6',
    text: '#a78bfa',
  },
  plugin: {
    bg: 'rgba(34, 197, 94, 0.2)',
    border: '#22c55e',
    text: '#4ade80',
  },
  flag: {
    bg: 'rgba(234, 179, 8, 0.2)',
    border: '#eab308',
    text: '#facc15',
  },
  conditional: {
    bg: 'rgba(6, 182, 212, 0.2)',
    border: '#06b6d4',
    text: '#22d3ee',
  },
};

// ============================================
// Helper Functions
// ============================================

/**
 * Extract all flags that can be set by selecting plugins
 */
function extractFlags(data: FomodData): FlagInfo[] {
  const flags: FlagInfo[] = [];

  const steps = data.config.installSteps ?? [];
  for (const step of steps) {
    const groups = step.optionGroups ?? [];
    for (const group of groups) {
      const plugins = group.plugins ?? [];
      for (const plugin of plugins) {
        const conditionFlags = plugin.conditionFlags ?? [];
        for (const flag of conditionFlags) {
          flags.push({
            name: flag.name,
            value: flag.value,
            source: {
              stepName: step.name,
              groupName: group.name,
              pluginName: plugin.name,
            },
          });
        }
      }
    }
  }

  return flags;
}

/**
 * Find all flag dependencies recursively from a dependency tree
 */
function extractFlagDependencies(dep: Dependency | undefined): string[] {
  if (!dep) return [];

  const flagNames: string[] = [];

  if (dep.flagDependency) {
    flagNames.push(dep.flagDependency.flag);
  }

  if (dep.children) {
    for (const child of dep.children) {
      flagNames.push(...extractFlagDependencies(child));
    }
  }

  return flagNames;
}

/**
 * Build graph nodes and edges from FOMOD data
 */
function buildGraph(
  data: FomodData,
): { nodes: Node<FomodNodeData>[]; edges: Edge[] } {
  const nodes: Node<FomodNodeData>[] = [];
  const edges: Edge[] = [];

  const steps = data.config.installSteps ?? [];
  const conditionalInstalls = data.config.conditionalFileInstalls ?? [];

  // Track unique flags
  const flagNodes = new Map<string, { id: string; x: number; y: number }>();
  const allFlags = extractFlags(data);

  // Position tracking
  let currentY = 0;

  // Add step nodes
  steps.forEach((step, stepIndex) => {
    const stepId = `step-${stepIndex}`;
    const stepX = 0;
    const stepY = currentY;

    nodes.push({
      id: stepId,
      type: 'default',
      position: { x: stepX, y: stepY },
      data: {
        label: step.name,
        nodeType: 'step',
        stepIndex,
      },
      style: buildNodeStyle('step', false),
    });

    // Check if step has visibility dependency
    const visibilityFlags = extractFlagDependencies(step.visible);
    for (const flagName of visibilityFlags) {
      const flagId = `flag-${flagName}`;
      if (!flagNodes.has(flagName)) {
        flagNodes.set(flagName, {
          id: flagId,
          x: -HORIZONTAL_SPACING * 1.5,
          y: currentY,
        });
      }
      edges.push({
        id: `${flagId}->${stepId}-visibility`,
        source: flagId,
        target: stepId,
        animated: true,
        style: { stroke: '#eab308', strokeWidth: 2, strokeDasharray: '5 5' },
        label: 'shows/hides',
        labelStyle: { fontSize: 10, fill: '#9ca3af' },
        markerEnd: { type: MarkerType.ArrowClosed, color: '#eab308' },
      });
    }

    // Add group nodes
    const groups = step.optionGroups ?? [];
    groups.forEach((group, groupIndex) => {
      const groupId = `step-${stepIndex}-group-${groupIndex}`;
      const groupX = HORIZONTAL_SPACING;
      const groupY = currentY + (groupIndex + 1) * VERTICAL_SPACING * 0.7;

      nodes.push({
        id: groupId,
        type: 'default',
        position: { x: groupX, y: groupY },
        data: {
          label: `${group.name} (${group.type})`,
          nodeType: 'group',
        },
        style: buildNodeStyle('group', false),
      });

      // Connect step to group
      edges.push({
        id: `${stepId}->${groupId}`,
        source: stepId,
        target: groupId,
        style: { stroke: '#4b5563', strokeWidth: 1 },
      });

      // Add plugin nodes
      const plugins = group.plugins ?? [];
      plugins.forEach((plugin, pluginIndex) => {
        const pluginId = `step-${stepIndex}-group-${groupIndex}-plugin-${pluginIndex}`;
        const pluginX = HORIZONTAL_SPACING * 2;
        const pluginY = groupY + pluginIndex * VERTICAL_SPACING * 0.5;

        const pluginType = plugin.typeDescriptor?.type ?? 'Optional';

        nodes.push({
          id: pluginId,
          type: 'default',
          position: { x: pluginX, y: pluginY },
          data: {
            label: plugin.name,
            nodeType: 'plugin',
            pluginType,
          },
          style: buildNodeStyle('plugin', false, pluginType),
        });

        // Connect group to plugin
        edges.push({
          id: `${groupId}->${pluginId}`,
          source: groupId,
          target: pluginId,
          style: { stroke: '#4b5563', strokeWidth: 1 },
        });

        // If plugin sets flags, create flag nodes and edges
        const conditionFlags = plugin.conditionFlags ?? [];
        for (const flag of conditionFlags) {
          const flagId = `flag-${flag.name}`;

          if (!flagNodes.has(flag.name)) {
            flagNodes.set(flag.name, {
              id: flagId,
              x: HORIZONTAL_SPACING * 3.5,
              y: pluginY,
            });
          }

          edges.push({
            id: `${pluginId}->${flagId}`,
            source: pluginId,
            target: flagId,
            style: { stroke: '#22c55e', strokeWidth: 2 },
            label: `sets "${flag.value}"`,
            labelStyle: { fontSize: 10, fill: '#9ca3af' },
            markerEnd: { type: MarkerType.ArrowClosed, color: '#22c55e' },
          });
        }
      });
    });

    // Update Y for next step
    const groupCount = groups.length;
    const maxPlugins = Math.max(...groups.map(g => (g.plugins?.length ?? 0)), 1);
    currentY += (groupCount * maxPlugins * VERTICAL_SPACING * 0.5) + VERTICAL_SPACING * 1.5;
  });

  // Add flag nodes
  for (const [flagName, flagInfo] of flagNodes) {
    const flagData = allFlags.find(f => f.name === flagName);
    nodes.push({
      id: flagInfo.id,
      type: 'default',
      position: { x: flagInfo.x, y: flagInfo.y },
      data: {
        label: flagName,
        nodeType: 'flag',
        flagValue: flagData?.value,
      },
      style: buildNodeStyle('flag', false),
    });
  }

  // Add conditional install nodes
  conditionalInstalls.forEach((item, index) => {
    const condId = `conditional-${index}`;
    const condY = currentY + index * VERTICAL_SPACING;

    nodes.push({
      id: condId,
      type: 'default',
      position: { x: HORIZONTAL_SPACING * 3.5, y: condY },
      data: {
        label: `Conditional Install ${index + 1}`,
        nodeType: 'conditional',
      },
      style: buildNodeStyle('conditional', false),
    });

    // Connect flags to conditional installs
    const depFlags = extractFlagDependencies(item.dependencies);
    for (const flagName of depFlags) {
      const flagId = `flag-${flagName}`;
      if (flagNodes.has(flagName)) {
        edges.push({
          id: `${flagId}->${condId}`,
          source: flagId,
          target: condId,
          style: { stroke: '#06b6d4', strokeWidth: 2 },
          label: 'triggers',
          labelStyle: { fontSize: 10, fill: '#9ca3af' },
          markerEnd: { type: MarkerType.ArrowClosed, color: '#06b6d4' },
        });
      }
    }
  });

  return { nodes, edges };
}

/**
 * Build node style based on type
 */
function buildNodeStyle(
  nodeType: FomodNodeType,
  isSelected: boolean,
  pluginType?: string,
): React.CSSProperties {
  const colors = NODE_COLORS[nodeType];

  // Adjust plugin colors based on type
  let bg = colors.bg;
  let border = colors.border;

  if (nodeType === 'plugin' && pluginType) {
    switch (pluginType) {
      case 'Required':
        bg = 'rgba(239, 68, 68, 0.2)';
        border = '#ef4444';
        break;
      case 'Recommended':
        bg = 'rgba(96, 165, 250, 0.2)';
        border = '#60a5fa';
        break;
      case 'NotUsable':
        bg = 'rgba(107, 114, 128, 0.3)';
        border = '#6b7280';
        break;
    }
  }

  return {
    width: NODE_WIDTH,
    height: NODE_HEIGHT,
    background: isSelected ? border : bg,
    border: `2px solid ${border}`,
    borderRadius: '8px',
    color: isSelected ? '#1f2937' : colors.text,
    fontSize: '11px',
    fontWeight: 500,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    padding: '6px',
    textAlign: 'center',
    boxShadow: isSelected ? '0 0 0 3px rgba(96, 165, 250, 0.5)' : 'none',
  };
}

// ============================================
// Legend Component
// ============================================

const GraphLegend: React.FC = () => (
  <div className="flex flex-wrap gap-3 p-3 rounded-sm bg-bg-card border border-border text-xs">
    <div className="flex items-center gap-2">
      <div
        className="w-4 h-4 rounded-xs"
        style={{ background: NODE_COLORS.step.bg, border: `1px solid ${NODE_COLORS.step.border}` }}
      />
      <span className="text-text-secondary">Step</span>
    </div>
    <div className="flex items-center gap-2">
      <div
        className="w-4 h-4 rounded-xs"
        style={{ background: NODE_COLORS.group.bg, border: `1px solid ${NODE_COLORS.group.border}` }}
      />
      <span className="text-text-secondary">Group</span>
    </div>
    <div className="flex items-center gap-2">
      <div
        className="w-4 h-4 rounded-xs"
        style={{ background: NODE_COLORS.plugin.bg, border: `1px solid ${NODE_COLORS.plugin.border}` }}
      />
      <span className="text-text-secondary">Plugin</span>
    </div>
    <div className="flex items-center gap-2">
      <div
        className="w-4 h-4 rounded-xs"
        style={{ background: NODE_COLORS.flag.bg, border: `1px solid ${NODE_COLORS.flag.border}` }}
      />
      <span className="text-text-secondary">Flag</span>
    </div>
    <div className="flex items-center gap-2">
      <div
        className="w-4 h-4 rounded-xs"
        style={{ background: NODE_COLORS.conditional.bg, border: `1px solid ${NODE_COLORS.conditional.border}` }}
      />
      <span className="text-text-secondary">Conditional</span>
    </div>
    <div className="flex items-center gap-2">
      <div className="w-4 h-0.5" style={{ background: '#eab308', borderStyle: 'dashed' }} />
      <span className="text-text-secondary">Visibility</span>
    </div>
  </div>
);

// ============================================
// Controls Component
// ============================================

interface GraphControlsProps {
  nodeCount: number;
  edgeCount: number;
  onFitView: () => void;
}

const GraphControls: React.FC<GraphControlsProps> = ({
  nodeCount,
  edgeCount,
  onFitView,
}) => (
  <div className="flex items-center gap-4 p-3 rounded-sm bg-bg-card border border-border">
    <div className="text-sm text-text-secondary">
      <span className="font-medium text-text-primary">{nodeCount}</span> nodes
    </div>
    <div className="text-sm text-text-secondary">
      <span className="font-medium text-text-primary">{edgeCount}</span> connections
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
// Internal Graph Component
// ============================================

interface InternalGraphProps {
  data: FomodData;
  onNavigateToStep?: (stepIndex: number) => void;
}

const InternalGraph: React.FC<InternalGraphProps> = ({
  data,
  onNavigateToStep,
}) => {
  const { fitView } = useReactFlow();
  const graphId = useId();

  // Build graph elements
  const { initialNodes, initialEdges } = useMemo(() => {
    const { nodes, edges } = buildGraph(data);
    return { initialNodes: nodes, initialEdges: edges };
  }, [data]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Update when data changes
  useEffect(() => {
    setNodes(initialNodes);
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  // Handle node click
  const onNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      const nodeData = node.data as FomodNodeData;
      if (nodeData.nodeType === 'step' && nodeData.stepIndex !== undefined && onNavigateToStep) {
        onNavigateToStep(nodeData.stepIndex);
      }
    },
    [onNavigateToStep]
  );

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
      id={graphId}
      className="w-full h-[600px] rounded-sm bg-bg-card border border-border overflow-hidden"
      role="img"
      aria-label={`FOMOD dependency graph showing ${nodes.length} elements and their relationships`}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
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
            const data = node.data as FomodNodeData;
            return NODE_COLORS[data.nodeType]?.border ?? '#6b7280';
          }}
          className="!bg-bg-secondary !border-border"
          maskColor="rgba(0, 0, 0, 0.6)"
        />
        <Panel position="top-left">
          <GraphControls
            nodeCount={nodes.length}
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
// Main Component
// ============================================

/**
 * Displays FOMOD structure as an interactive dependency graph.
 * Shows relationships between steps, groups, plugins, flags, and conditional installs.
 */
export const FomodDependencyGraph: React.FC<FomodDependencyGraphProps> = (props) => {
  const steps = props.data.config.installSteps ?? [];

  // Show message if no steps
  if (steps.length === 0) {
    return (
      <div className="p-6 rounded-sm bg-bg-card border border-border text-center">
        <p className="text-text-secondary">
          This FOMOD has no installation steps to visualize.
        </p>
      </div>
    );
  }

  return (
    <ReactFlowProvider>
      <InternalGraph {...props} />
    </ReactFlowProvider>
  );
};
