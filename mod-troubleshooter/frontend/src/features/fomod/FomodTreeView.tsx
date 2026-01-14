import { useState, useCallback, useMemo, useId } from 'react';

import type {
  FomodData,
  InstallStep,
  OptionGroup,
  Plugin,
  GroupType,
  PluginType,
  Dependency,
  FileList,
  ConditionalInstallItem,
} from '@/types/index.ts';

// ============================================
// Tree Node Types
// ============================================

type TreeNodeType =
  | 'root'
  | 'info'
  | 'dependencies'
  | 'required-files'
  | 'steps'
  | 'step'
  | 'group'
  | 'plugin'
  | 'files'
  | 'flags'
  | 'conditional-installs';

interface BaseTreeNode {
  id: string;
  type: TreeNodeType;
  label: string;
  children?: TreeNode[];
}

interface RootNode extends BaseTreeNode {
  type: 'root';
}

interface InfoNode extends BaseTreeNode {
  type: 'info';
  details: { name: string; author?: string; version?: string; description?: string };
}

interface DependenciesNode extends BaseTreeNode {
  type: 'dependencies';
  dependency: Dependency;
}

interface RequiredFilesNode extends BaseTreeNode {
  type: 'required-files';
  fileList: FileList;
}

interface StepsNode extends BaseTreeNode {
  type: 'steps';
  stepCount: number;
}

interface StepNode extends BaseTreeNode {
  type: 'step';
  step: InstallStep;
  visible?: Dependency;
}

interface GroupNode extends BaseTreeNode {
  type: 'group';
  group: OptionGroup;
  groupType: GroupType;
}

interface PluginNode extends BaseTreeNode {
  type: 'plugin';
  plugin: Plugin;
  pluginType: PluginType;
}

interface FilesNode extends BaseTreeNode {
  type: 'files';
  fileList: FileList;
}

interface FlagsNode extends BaseTreeNode {
  type: 'flags';
  flags: { name: string; value: string }[];
}

interface ConditionalInstallsNode extends BaseTreeNode {
  type: 'conditional-installs';
  items: ConditionalInstallItem[];
}

type TreeNode =
  | RootNode
  | InfoNode
  | DependenciesNode
  | RequiredFilesNode
  | StepsNode
  | StepNode
  | GroupNode
  | PluginNode
  | FilesNode
  | FlagsNode
  | ConditionalInstallsNode;

// ============================================
// Props Interfaces
// ============================================

interface FomodTreeViewProps {
  data: FomodData;
}

interface TreeNodeViewProps {
  node: TreeNode;
  depth: number;
  expandedNodes: Set<string>;
  onToggle: (nodeId: string) => void;
}

// ============================================
// Helper Functions
// ============================================

/** Get plugin type from descriptor */
function getPluginType(plugin: Plugin): PluginType {
  return plugin.typeDescriptor?.type ?? 'Optional';
}

/** Get display label for group type */
function getGroupTypeLabel(groupType: GroupType): string {
  switch (groupType) {
    case 'SelectExactlyOne':
      return 'Select exactly one';
    case 'SelectAtMostOne':
      return 'Select at most one';
    case 'SelectAtLeastOne':
      return 'Select at least one';
    case 'SelectAny':
      return 'Select any';
    case 'SelectAll':
      return 'All required';
    default:
      return groupType;
  }
}

/** Get badge color for plugin type */
function getPluginTypeBadgeClass(pluginType: PluginType): string {
  switch (pluginType) {
    case 'Required':
      return 'bg-error/20 text-error';
    case 'Recommended':
      return 'bg-accent/20 text-accent';
    case 'Optional':
      return 'bg-text-muted/20 text-text-secondary';
    case 'NotUsable':
      return 'bg-error/30 text-error line-through';
    case 'CouldBeUsable':
      return 'bg-warning/20 text-warning';
    default:
      return 'bg-text-muted/20 text-text-secondary';
  }
}

/** Get badge color for group type */
function getGroupTypeBadgeClass(groupType: GroupType): string {
  switch (groupType) {
    case 'SelectExactlyOne':
      return 'bg-accent/20 text-accent';
    case 'SelectAtMostOne':
      return 'bg-warning/20 text-warning';
    case 'SelectAtLeastOne':
      return 'bg-error/20 text-error';
    case 'SelectAny':
      return 'bg-text-muted/20 text-text-secondary';
    case 'SelectAll':
      return 'bg-error/20 text-error';
    default:
      return 'bg-text-muted/20 text-text-secondary';
  }
}

/** Get icon for node type */
function getNodeIcon(type: TreeNodeType): string {
  switch (type) {
    case 'root':
      return '📦';
    case 'info':
      return 'ℹ️';
    case 'dependencies':
      return '🔗';
    case 'required-files':
      return '📋';
    case 'steps':
      return '📝';
    case 'step':
      return '👣';
    case 'group':
      return '📂';
    case 'plugin':
      return '🔌';
    case 'files':
      return '📄';
    case 'flags':
      return '🚩';
    case 'conditional-installs':
      return '❓';
    default:
      return '📄';
  }
}

/** Build dependency description string */
function describeDependency(dep: Dependency, indent = 0): string {
  const prefix = '  '.repeat(indent);

  if (dep.flagDependency) {
    return `${prefix}Flag "${dep.flagDependency.flag}" = "${dep.flagDependency.value}"`;
  }
  if (dep.fileDependency) {
    return `${prefix}File "${dep.fileDependency.file}" is ${dep.fileDependency.state}`;
  }
  if (dep.gameDependency) {
    return `${prefix}Game version >= ${dep.gameDependency.version}`;
  }
  if (dep.fommDependency) {
    return `${prefix}FOMM version >= ${dep.fommDependency.version}`;
  }
  if (dep.children && dep.children.length > 0) {
    const op = dep.operator ?? 'And';
    const childDescriptions = dep.children.map((c) => describeDependency(c, indent + 1));
    return `${prefix}${op}:\n${childDescriptions.join('\n')}`;
  }
  return `${prefix}(No conditions)`;
}

/** Count files in a FileList */
function countFiles(fileList: FileList | undefined): number {
  if (!fileList) return 0;
  return (fileList.files?.length ?? 0) + (fileList.folders?.length ?? 0);
}

// ============================================
// Build Tree Structure
// ============================================

/** Build a tree structure from FOMOD data */
function buildTreeFromFomod(data: FomodData): TreeNode {
  const children: TreeNode[] = [];
  let nodeId = 0;

  const nextId = () => `node-${nodeId++}`;

  // Info section
  if (data.info) {
    children.push({
      id: nextId(),
      type: 'info',
      label: 'Mod Information',
      details: {
        name: data.info.name ?? 'Unknown',
        author: data.info.author,
        version: data.info.version,
        description: data.info.description,
      },
    });
  }

  // Module dependencies
  if (data.config.moduleDependencies) {
    children.push({
      id: nextId(),
      type: 'dependencies',
      label: 'Module Dependencies',
      dependency: data.config.moduleDependencies,
    });
  }

  // Required install files
  if (data.config.requiredInstallFiles && countFiles(data.config.requiredInstallFiles) > 0) {
    children.push({
      id: nextId(),
      type: 'required-files',
      label: `Required Files (${countFiles(data.config.requiredInstallFiles)})`,
      fileList: data.config.requiredInstallFiles,
    });
  }

  // Install steps
  if (data.config.installSteps && data.config.installSteps.length > 0) {
    const stepsChildren: TreeNode[] = data.config.installSteps.map((step) => {
      const groupChildren: TreeNode[] = (step.optionGroups ?? []).map((group) => {
        const pluginChildren: TreeNode[] = (group.plugins ?? []).map((plugin) => {
          const pluginNodeChildren: TreeNode[] = [];

          // Plugin files
          if (plugin.files && countFiles(plugin.files) > 0) {
            pluginNodeChildren.push({
              id: nextId(),
              type: 'files',
              label: `Files (${countFiles(plugin.files)})`,
              fileList: plugin.files,
            });
          }

          // Plugin condition flags
          if (plugin.conditionFlags && plugin.conditionFlags.length > 0) {
            pluginNodeChildren.push({
              id: nextId(),
              type: 'flags',
              label: `Sets ${plugin.conditionFlags.length} flag(s)`,
              flags: plugin.conditionFlags,
            });
          }

          return {
            id: nextId(),
            type: 'plugin' as const,
            label: plugin.name,
            plugin,
            pluginType: getPluginType(plugin),
            children: pluginNodeChildren.length > 0 ? pluginNodeChildren : undefined,
          };
        });

        return {
          id: nextId(),
          type: 'group' as const,
          label: group.name,
          group,
          groupType: group.type,
          children: pluginChildren.length > 0 ? pluginChildren : undefined,
        };
      });

      return {
        id: nextId(),
        type: 'step' as const,
        label: step.name,
        step,
        visible: step.visible,
        children: groupChildren.length > 0 ? groupChildren : undefined,
      };
    });

    children.push({
      id: nextId(),
      type: 'steps',
      label: `Install Steps (${data.config.installSteps.length})`,
      stepCount: data.config.installSteps.length,
      children: stepsChildren,
    });
  }

  // Conditional file installs
  if (data.config.conditionalFileInstalls && data.config.conditionalFileInstalls.length > 0) {
    children.push({
      id: nextId(),
      type: 'conditional-installs',
      label: `Conditional Installs (${data.config.conditionalFileInstalls.length})`,
      items: data.config.conditionalFileInstalls,
    });
  }

  return {
    id: 'root',
    type: 'root',
    label: data.config.moduleName,
    children,
  };
}

// ============================================
// Tree Node Content Components
// ============================================

interface InfoContentProps {
  details: { name: string; author?: string; version?: string; description?: string };
}

const InfoContent: React.FC<InfoContentProps> = ({ details }) => (
  <div className="pl-8 py-2 text-sm space-y-1">
    <p className="text-text-primary">
      <span className="text-text-muted">Name:</span> {details.name}
    </p>
    {details.author && (
      <p className="text-text-secondary">
        <span className="text-text-muted">Author:</span> {details.author}
      </p>
    )}
    {details.version && (
      <p className="text-text-secondary">
        <span className="text-text-muted">Version:</span> {details.version}
      </p>
    )}
    {details.description && (
      <p className="text-text-muted">
        <span className="text-text-muted">Description:</span> {details.description}
      </p>
    )}
  </div>
);

interface DependencyContentProps {
  dependency: Dependency;
}

const DependencyContent: React.FC<DependencyContentProps> = ({ dependency }) => (
  <div className="pl-8 py-2">
    <pre className="text-xs text-text-muted font-mono whitespace-pre-wrap bg-bg-secondary/50 p-2 rounded-xs">
      {describeDependency(dependency)}
    </pre>
  </div>
);

interface FileListContentProps {
  fileList: FileList;
}

const FileListContent: React.FC<FileListContentProps> = ({ fileList }) => (
  <div className="pl-8 py-2 space-y-2">
    {fileList.files && fileList.files.length > 0 && (
      <div>
        <p className="text-xs text-text-muted mb-1">Files:</p>
        <ul className="text-sm text-text-secondary space-y-0.5">
          {fileList.files.map((f, i) => (
            <li key={i} className="flex items-center gap-2">
              <span aria-hidden="true" className="text-text-muted">
                📄
              </span>
              <span className="truncate" title={f.source}>
                {f.source}
              </span>
              {f.destination && f.destination !== f.source && (
                <span className="text-text-muted text-xs">→ {f.destination}</span>
              )}
            </li>
          ))}
        </ul>
      </div>
    )}
    {fileList.folders && fileList.folders.length > 0 && (
      <div>
        <p className="text-xs text-text-muted mb-1">Folders:</p>
        <ul className="text-sm text-text-secondary space-y-0.5">
          {fileList.folders.map((f, i) => (
            <li key={i} className="flex items-center gap-2">
              <span aria-hidden="true" className="text-accent">
                📁
              </span>
              <span className="truncate" title={f.source}>
                {f.source}
              </span>
              {f.destination && f.destination !== f.source && (
                <span className="text-text-muted text-xs">→ {f.destination}</span>
              )}
            </li>
          ))}
        </ul>
      </div>
    )}
  </div>
);

interface FlagsContentProps {
  flags: { name: string; value: string }[];
}

const FlagsContent: React.FC<FlagsContentProps> = ({ flags }) => (
  <div className="pl-8 py-2">
    <ul className="text-sm space-y-1">
      {flags.map((flag, i) => (
        <li key={i} className="flex items-center gap-2">
          <span aria-hidden="true" className="text-warning">
            🚩
          </span>
          <code className="text-text-primary bg-bg-secondary/50 px-1 rounded-xs">
            {flag.name}
          </code>
          <span className="text-text-muted">=</span>
          <code className="text-accent bg-bg-secondary/50 px-1 rounded-xs">{flag.value}</code>
        </li>
      ))}
    </ul>
  </div>
);

interface ConditionalInstallsContentProps {
  items: ConditionalInstallItem[];
}

const ConditionalInstallsContent: React.FC<ConditionalInstallsContentProps> = ({ items }) => (
  <div className="pl-8 py-2 space-y-3">
    {items.map((item, i) => (
      <div key={i} className="border-l-2 border-warning/30 pl-3">
        <p className="text-xs text-text-muted mb-1">Condition {i + 1}:</p>
        {item.dependencies && (
          <pre className="text-xs text-text-muted font-mono whitespace-pre-wrap bg-bg-secondary/50 p-2 rounded-xs mb-2">
            {describeDependency(item.dependencies)}
          </pre>
        )}
        {item.files && countFiles(item.files) > 0 && (
          <p className="text-sm text-text-secondary">
            Installs {countFiles(item.files)} file(s)/folder(s)
          </p>
        )}
      </div>
    ))}
  </div>
);

// ============================================
// Tree Node View Component
// ============================================

const TreeNodeView: React.FC<TreeNodeViewProps> = ({ node, depth, expandedNodes, onToggle }) => {
  const isExpanded = expandedNodes.has(node.id);
  const hasChildren = node.children && node.children.length > 0;
  const hasContent = node.type === 'info' || node.type === 'dependencies' || node.type === 'required-files'
    || node.type === 'files' || node.type === 'flags' || node.type === 'conditional-installs';
  const canExpand = hasChildren || hasContent;

  const handleToggle = useCallback(() => {
    onToggle(node.id);
  }, [node.id, onToggle]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        handleToggle();
      }
    },
    [handleToggle],
  );

  // Render badges based on node type
  const renderBadge = () => {
    if (node.type === 'plugin') {
      const pluginNode = node as PluginNode;
      return (
        <span
          className={`ml-2 px-2 py-0.5 rounded-full text-xs font-medium ${getPluginTypeBadgeClass(pluginNode.pluginType)}`}
        >
          {pluginNode.pluginType}
        </span>
      );
    }
    if (node.type === 'group') {
      const groupNode = node as GroupNode;
      return (
        <span
          className={`ml-2 px-2 py-0.5 rounded-full text-xs font-medium ${getGroupTypeBadgeClass(groupNode.groupType)}`}
        >
          {getGroupTypeLabel(groupNode.groupType)}
        </span>
      );
    }
    if (node.type === 'step') {
      const stepNode = node as StepNode;
      if (stepNode.visible) {
        return (
          <span
            className="ml-2 px-2 py-0.5 rounded-full text-xs font-medium bg-warning/20 text-warning"
            title="Has visibility conditions"
          >
            Conditional
          </span>
        );
      }
    }
    return null;
  };

  // Render expanded content for leaf-ish nodes
  const renderContent = () => {
    if (!isExpanded) return null;

    switch (node.type) {
      case 'info':
        return <InfoContent details={(node as InfoNode).details} />;
      case 'dependencies':
        return <DependencyContent dependency={(node as DependenciesNode).dependency} />;
      case 'required-files':
        return <FileListContent fileList={(node as RequiredFilesNode).fileList} />;
      case 'files':
        return <FileListContent fileList={(node as FilesNode).fileList} />;
      case 'flags':
        return <FlagsContent flags={(node as FlagsNode).flags} />;
      case 'conditional-installs':
        return <ConditionalInstallsContent items={(node as ConditionalInstallsNode).items} />;
      default:
        return null;
    }
  };

  return (
    <li className="select-none">
      <div
        role="treeitem"
        aria-expanded={canExpand ? isExpanded : undefined}
        aria-selected={false}
        tabIndex={0}
        onClick={handleToggle}
        onKeyDown={handleKeyDown}
        className={`
          flex items-center gap-2 py-1.5 px-2 rounded-xs cursor-pointer
          hover:bg-bg-secondary/50
          focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-1
          transition-colors motion-reduce:transition-none
        `}
        style={{ paddingLeft: `${depth * 20 + 8}px` }}
      >
        {canExpand ? (
          <span
            aria-hidden="true"
            className={`text-text-muted text-xs transition-transform motion-reduce:transition-none ${
              isExpanded ? 'rotate-90' : ''
            }`}
          >
            ▶
          </span>
        ) : (
          <span className="w-3" aria-hidden="true" />
        )}
        <span aria-hidden="true">{getNodeIcon(node.type)}</span>
        <span className="text-text-primary font-medium">{node.label}</span>
        {renderBadge()}
      </div>

      {/* Expanded content */}
      {renderContent()}

      {/* Children */}
      {isExpanded && hasChildren && (
        <ul role="group" className="list-none">
          {node.children?.map((child) => (
            <TreeNodeView
              key={child.id}
              node={child}
              depth={depth + 1}
              expandedNodes={expandedNodes}
              onToggle={onToggle}
            />
          ))}
        </ul>
      )}
    </li>
  );
};

// ============================================
// Main FomodTreeView Component
// ============================================

/** Tree view of the full FOMOD structure */
export const FomodTreeView: React.FC<FomodTreeViewProps> = ({ data }) => {
  const treeId = useId();
  const treeData = useMemo(() => buildTreeFromFomod(data), [data]);

  // Start with root and first-level nodes expanded
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(() => {
    const initial = new Set<string>(['root']);
    if (treeData.children) {
      for (const child of treeData.children) {
        initial.add(child.id);
      }
    }
    return initial;
  });

  const handleToggle = useCallback((nodeId: string) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  }, []);

  const handleExpandAll = useCallback(() => {
    const allNodeIds = new Set<string>();

    const collectIds = (node: TreeNode) => {
      allNodeIds.add(node.id);
      if (node.children) {
        for (const child of node.children) {
          collectIds(child);
        }
      }
    };

    collectIds(treeData);
    setExpandedNodes(allNodeIds);
  }, [treeData]);

  const handleCollapseAll = useCallback(() => {
    setExpandedNodes(new Set(['root']));
  }, []);

  return (
    <section
      aria-label="FOMOD structure tree"
      className="p-4 rounded-sm bg-bg-card border border-border"
    >
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-text-primary">FOMOD Structure</h3>
        <div className="flex gap-2">
          <button
            onClick={handleExpandAll}
            className="min-h-9 px-3 py-1.5 rounded-sm text-sm font-medium
              bg-bg-secondary text-text-secondary
              hover:bg-bg-secondary/80
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none"
          >
            Expand All
          </button>
          <button
            onClick={handleCollapseAll}
            className="min-h-9 px-3 py-1.5 rounded-sm text-sm font-medium
              bg-bg-secondary text-text-secondary
              hover:bg-bg-secondary/80
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none"
          >
            Collapse All
          </button>
        </div>
      </div>

      <div className="max-h-[600px] overflow-y-auto rounded-xs border border-border bg-bg-secondary/30">
        <ul
          id={treeId}
          role="tree"
          aria-label={`FOMOD structure for ${data.config.moduleName}`}
          className="py-2 list-none"
        >
          <TreeNodeView
            node={treeData}
            depth={0}
            expandedNodes={expandedNodes}
            onToggle={handleToggle}
          />
        </ul>
      </div>

      <div className="mt-3 flex flex-wrap gap-4 text-xs text-text-muted">
        <span className="flex items-center gap-1">
          <span aria-hidden="true">👣</span> Step
        </span>
        <span className="flex items-center gap-1">
          <span aria-hidden="true">📂</span> Option Group
        </span>
        <span className="flex items-center gap-1">
          <span aria-hidden="true">🔌</span> Plugin
        </span>
        <span className="flex items-center gap-1">
          <span aria-hidden="true">📄</span> Files
        </span>
        <span className="flex items-center gap-1">
          <span aria-hidden="true">🚩</span> Condition Flags
        </span>
      </div>
    </section>
  );
};
