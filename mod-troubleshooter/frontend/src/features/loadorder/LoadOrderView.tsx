import { useState, useMemo } from 'react';

import { useLoadOrderAnalysis } from '@hooks/useLoadOrder.ts';
import { ApiError } from '@services/api.ts';
import { DependencyGraphView } from './DependencyGraphView.tsx';

import type {
  LoadOrderPluginInfo,
  LoadOrderIssue,
  LoadOrderStats,
  LoadOrderPluginType,
  IssueSeverity,
} from '@/types/index.ts';

/** View mode for the load order display */
type ViewMode = 'list' | 'graph';

// ============================================
// Props Interfaces
// ============================================

interface LoadOrderViewProps {
  slug: string;
  revision: number;
}

interface LoadOrderHeaderProps {
  stats: LoadOrderStats;
  cached: boolean;
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
}

interface LoadOrderListProps {
  plugins: LoadOrderPluginInfo[];
  selectedPlugin: LoadOrderPluginInfo | null;
  onSelectPlugin: (plugin: LoadOrderPluginInfo | null) => void;
  pluginIssues: Map<string, LoadOrderIssue[]>;
}

interface PluginRowProps {
  plugin: LoadOrderPluginInfo;
  isSelected: boolean;
  onSelect: () => void;
  issues: LoadOrderIssue[];
}

interface LoadOrderDetailsProps {
  plugin: LoadOrderPluginInfo;
  issues: LoadOrderIssue[];
  dependencyGraph: Record<string, string[]>;
}

interface WarningPanelProps {
  issues: LoadOrderIssue[];
  onSelectIssue: (plugin: string) => void;
}

// ============================================
// Helper Functions
// ============================================

/** Get badge color class for plugin type */
function getPluginTypeBadgeClass(type: LoadOrderPluginType): string {
  switch (type) {
    case 'ESM':
      return 'bg-accent/20 text-accent';
    case 'ESL':
      return 'bg-warning/20 text-warning';
    case 'ESP':
      return 'bg-text-muted/20 text-text-secondary';
    default:
      return 'bg-text-muted/20 text-text-secondary';
  }
}

/** Get badge color class for issue severity */
function getSeverityBadgeClass(severity: IssueSeverity): string {
  switch (severity) {
    case 'error':
      return 'bg-error/20 text-error';
    case 'warning':
      return 'bg-warning/20 text-warning';
    default:
      return 'bg-text-muted/20 text-text-secondary';
  }
}

/** Get human-readable issue type label */
function getIssueTypeLabel(type: string): string {
  switch (type) {
    case 'missing_master':
      return 'Missing Master';
    case 'wrong_order':
      return 'Wrong Order';
    case 'duplicate_plugin':
      return 'Duplicate Plugin';
    default:
      return type;
  }
}

/** Build a map of plugin filename to their issues */
function buildPluginIssuesMap(issues: LoadOrderIssue[]): Map<string, LoadOrderIssue[]> {
  const map = new Map<string, LoadOrderIssue[]>();
  for (const issue of issues) {
    const existing = map.get(issue.plugin) ?? [];
    existing.push(issue);
    map.set(issue.plugin, existing);
  }
  return map;
}

// ============================================
// Loading Skeleton
// ============================================

const LoadOrderSkeleton: React.FC = () => (
  <div className="space-y-6 animate-pulse">
    {/* Stats header skeleton */}
    <div className="p-6 rounded-sm bg-bg-card border border-border">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="space-y-2">
            <div className="h-4 w-20 bg-bg-secondary rounded-xs" />
            <div className="h-8 w-16 bg-bg-secondary rounded-xs" />
          </div>
        ))}
      </div>
    </div>
    {/* List skeleton */}
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div className="lg:col-span-2 p-4 rounded-sm bg-bg-card border border-border space-y-2">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div key={i} className="h-12 bg-bg-secondary rounded-xs" />
        ))}
      </div>
      <div className="p-4 rounded-sm bg-bg-card border border-border">
        <div className="h-6 w-24 bg-bg-secondary rounded-xs mb-4" />
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-16 bg-bg-secondary rounded-xs" />
          ))}
        </div>
      </div>
    </div>
  </div>
);

// ============================================
// Error Display
// ============================================

interface ErrorDisplayProps {
  error: Error;
  onRetry: () => void;
}

const ErrorDisplay: React.FC<ErrorDisplayProps> = ({ error, onRetry }) => {
  let message = 'An unexpected error occurred while analyzing the load order.';

  if (error instanceof ApiError) {
    if (error.status === 404) {
      message = 'Collection not found. Please check the collection slug.';
    } else if (error.status === 401 || error.status === 403) {
      message = 'API key is missing or invalid. Please configure it in Settings.';
    } else if (error.status === 402) {
      message = 'This feature requires a Nexus Mods Premium account.';
    } else if (error.status >= 500) {
      message = 'Server error. Please try again later.';
    } else {
      message = error.message;
    }
  }

  return (
    <div
      role="alert"
      className="p-6 rounded-sm bg-error/10 border border-error text-center"
    >
      <p className="text-error font-medium mb-4">{message}</p>
      <button
        onClick={onRetry}
        className="min-h-11 px-6 py-2 rounded-sm
          bg-error text-white font-medium
          hover:bg-error/80
          focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
          transition-colors motion-reduce:transition-none"
      >
        Try Again
      </button>
    </div>
  );
};

// ============================================
// Empty State
// ============================================

const EmptyState: React.FC = () => (
  <div className="p-6 rounded-sm bg-bg-card border border-border text-center">
    <p className="text-text-secondary text-lg mb-2">No Plugins Found</p>
    <p className="text-text-muted">
      This collection does not contain any plugin files to analyze.
    </p>
  </div>
);

// ============================================
// View Mode Toggle Component
// ============================================

interface ViewModeToggleProps {
  viewMode: ViewMode;
  onChange: (mode: ViewMode) => void;
}

const ViewModeToggle: React.FC<ViewModeToggleProps> = ({ viewMode, onChange }) => (
  <div
    className="flex rounded-sm overflow-hidden border border-border"
    role="radiogroup"
    aria-label="View mode"
  >
    <button
      onClick={() => onChange('list')}
      className={`
        min-h-9 px-4 py-2 text-sm font-medium
        transition-colors motion-reduce:transition-none
        focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-[-2px]
        ${viewMode === 'list'
          ? 'bg-accent text-white'
          : 'bg-bg-secondary text-text-secondary hover:bg-bg-hover'
        }
      `}
      role="radio"
      aria-checked={viewMode === 'list'}
    >
      List
    </button>
    <button
      onClick={() => onChange('graph')}
      className={`
        min-h-9 px-4 py-2 text-sm font-medium
        transition-colors motion-reduce:transition-none
        focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-[-2px]
        ${viewMode === 'graph'
          ? 'bg-accent text-white'
          : 'bg-bg-secondary text-text-secondary hover:bg-bg-hover'
        }
      `}
      role="radio"
      aria-checked={viewMode === 'graph'}
    >
      Graph
    </button>
  </div>
);

// ============================================
// Constants
// ============================================

/** Maximum number of regular plugin slots (ESM + ESP) in Skyrim SE */
const PLUGIN_SLOT_LIMIT = 254;

/** Warning threshold (percentage of limit used) */
const SLOT_WARNING_THRESHOLD = 0.9; // 90%

/** Critical threshold (percentage of limit used) */
const SLOT_CRITICAL_THRESHOLD = 0.98; // 98%

// ============================================
// Slot Limit Warning Component
// ============================================

interface SlotLimitWarningProps {
  esmCount: number;
  espCount: number;
}

const SlotLimitWarning: React.FC<SlotLimitWarningProps> = ({ esmCount, espCount }) => {
  const usedSlots = esmCount + espCount;
  const remainingSlots = PLUGIN_SLOT_LIMIT - usedSlots;
  const percentUsed = usedSlots / PLUGIN_SLOT_LIMIT;

  // Determine severity
  const isOverLimit = usedSlots > PLUGIN_SLOT_LIMIT;
  const isCritical = usedSlots >= PLUGIN_SLOT_LIMIT * SLOT_CRITICAL_THRESHOLD;
  const isWarning = usedSlots >= PLUGIN_SLOT_LIMIT * SLOT_WARNING_THRESHOLD;

  // Don't show anything if under warning threshold
  if (!isWarning && !isCritical && !isOverLimit) {
    return null;
  }

  const severity = isOverLimit ? 'error' : isCritical ? 'error' : 'warning';
  const bgClass = severity === 'error' ? 'bg-error/10 border-error' : 'bg-warning/10 border-warning';
  const textClass = severity === 'error' ? 'text-error' : 'text-warning';
  const barBgClass = severity === 'error' ? 'bg-error' : 'bg-warning';

  return (
    <div
      role="alert"
      className={`p-4 rounded-sm border ${bgClass}`}
    >
      <div className="flex items-start gap-3">
        <span className="text-xl" aria-hidden="true">
          {isOverLimit ? '⛔' : '⚠️'}
        </span>
        <div className="flex-1 min-w-0">
          <p className={`font-semibold ${textClass}`}>
            {isOverLimit
              ? 'Plugin Slot Limit Exceeded!'
              : isCritical
                ? 'Plugin Slot Limit Critical!'
                : 'Approaching Plugin Slot Limit'}
          </p>
          <p className="text-sm text-text-secondary mt-1">
            {isOverLimit ? (
              <>
                You are using <strong className="text-error">{usedSlots}</strong> regular plugin slots,
                which exceeds the Skyrim SE limit of <strong>{PLUGIN_SLOT_LIMIT}</strong>.
                <br />
                <span className="text-error font-medium">
                  Remove {usedSlots - PLUGIN_SLOT_LIMIT} plugin(s) or convert to ESL format.
                </span>
              </>
            ) : (
              <>
                Using <strong className={textClass}>{usedSlots}</strong> of{' '}
                <strong>{PLUGIN_SLOT_LIMIT}</strong> regular plugin slots ({remainingSlots} remaining).
                <br />
                <span className="text-text-muted">
                  Consider converting some plugins to ESL format to free up slots.
                </span>
              </>
            )}
          </p>

          {/* Progress bar */}
          <div className="mt-3" aria-hidden="true">
            <div className="h-2 bg-bg-secondary rounded-full overflow-hidden">
              <div
                className={`h-full ${barBgClass} transition-all motion-reduce:transition-none`}
                style={{ width: `${Math.min(percentUsed * 100, 100)}%` }}
              />
            </div>
            <div className="flex justify-between mt-1 text-xs text-text-muted">
              <span>{usedSlots} ESM/ESP slots used</span>
              <span>{PLUGIN_SLOT_LIMIT} max</span>
            </div>
          </div>

          <p className="text-xs text-text-muted mt-2">
            Note: ESL (Light) plugins don't count toward this limit.
          </p>
        </div>
      </div>
    </div>
  );
};

// ============================================
// Stats Header Component
// ============================================

const LoadOrderHeader: React.FC<LoadOrderHeaderProps> = ({
  stats,
  cached,
  viewMode,
  onViewModeChange,
}) => (
  <header className="p-6 rounded-sm bg-bg-card border border-border">
    <div className="flex items-start justify-between gap-4 mb-4 flex-wrap">
      <h2 className="text-xl font-bold text-text-primary">Load Order Analysis</h2>
      <div className="flex items-center gap-3">
        <ViewModeToggle viewMode={viewMode} onChange={onViewModeChange} />
        <span
          className={`px-3 py-1 rounded-full text-xs font-medium ${
            cached
              ? 'bg-accent/20 text-accent'
              : 'bg-text-muted/20 text-text-secondary'
          }`}
        >
          {cached ? 'Cached' : 'Fresh'}
        </span>
      </div>
    </div>
    <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
      <StatItem label="Total Plugins" value={stats.totalPlugins} />
      <StatItem label="Masters (ESM)" value={stats.esmCount} variant="accent" />
      <StatItem label="Plugins (ESP)" value={stats.espCount} />
      <StatItem label="Light (ESL)" value={stats.eslCount} variant="warning" />
      <StatItem
        label="Errors"
        value={stats.errorCount}
        variant={stats.errorCount > 0 ? 'error' : 'default'}
      />
      <StatItem
        label="Warnings"
        value={stats.warningCount}
        variant={stats.warningCount > 0 ? 'warning' : 'default'}
      />
    </div>
  </header>
);

interface StatItemProps {
  label: string;
  value: number;
  variant?: 'default' | 'accent' | 'warning' | 'error';
}

const StatItem: React.FC<StatItemProps> = ({ label, value, variant = 'default' }) => {
  const valueClass = {
    default: 'text-text-primary',
    accent: 'text-accent',
    warning: 'text-warning',
    error: 'text-error',
  }[variant];

  return (
    <div className="space-y-1">
      <p className="text-sm text-text-muted">{label}</p>
      <p className={`text-2xl font-bold ${valueClass}`}>{value}</p>
    </div>
  );
};

// ============================================
// Plugin Row Component
// ============================================

const PluginRow: React.FC<PluginRowProps> = ({ plugin, isSelected, onSelect, issues }) => (
  <button
    onClick={onSelect}
    className={`
      w-full min-h-11 px-4 py-3 rounded-sm text-left
      flex items-center gap-3
      transition-colors motion-reduce:transition-none
      focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
      ${isSelected
        ? 'bg-accent/10 border border-accent'
        : 'bg-bg-secondary border border-transparent hover:bg-bg-secondary/80'
      }
    `}
    aria-pressed={isSelected}
  >
    <span className="text-text-muted text-sm font-mono w-8 text-right">
      {String(plugin.index).padStart(3, '0')}
    </span>
    <span
      className={`px-2 py-0.5 rounded-full text-xs font-medium ${getPluginTypeBadgeClass(plugin.type)}`}
    >
      {plugin.type}
    </span>
    <span className="flex-1 font-medium text-text-primary truncate">
      {plugin.filename}
    </span>
    {plugin.masters.length > 0 && (
      <span className="text-xs text-text-muted">
        {plugin.masters.length} master{plugin.masters.length !== 1 ? 's' : ''}
      </span>
    )}
    {issues.length > 0 && (
      <span
        className={`px-2 py-0.5 rounded-full text-xs font-medium ${
          issues.some(i => i.severity === 'error')
            ? 'bg-error/20 text-error'
            : 'bg-warning/20 text-warning'
        }`}
      >
        {issues.length} issue{issues.length !== 1 ? 's' : ''}
      </span>
    )}
  </button>
);

// ============================================
// Load Order List Component
// ============================================

const LoadOrderList: React.FC<LoadOrderListProps> = ({
  plugins,
  selectedPlugin,
  onSelectPlugin,
  pluginIssues,
}) => (
  <section
    aria-label="Plugin list"
    className="p-4 rounded-sm bg-bg-card border border-border"
  >
    <h3 className="text-lg font-semibold text-text-primary mb-4">
      Plugins ({plugins.length})
    </h3>
    <div className="space-y-1 max-h-[600px] overflow-y-auto">
      {plugins.map((plugin) => (
        <PluginRow
          key={`${plugin.index}-${plugin.filename}`}
          plugin={plugin}
          isSelected={selectedPlugin?.filename === plugin.filename}
          onSelect={() =>
            onSelectPlugin(
              selectedPlugin?.filename === plugin.filename ? null : plugin
            )
          }
          issues={pluginIssues.get(plugin.filename) ?? []}
        />
      ))}
    </div>
  </section>
);

// ============================================
// Load Order Details Component
// ============================================

const LoadOrderDetails: React.FC<LoadOrderDetailsProps> = ({
  plugin,
  issues,
  dependencyGraph,
}) => {
  const masters = dependencyGraph[plugin.filename] ?? plugin.masters;

  return (
    <section
      aria-label={`Details for ${plugin.filename}`}
      className="p-4 rounded-sm bg-bg-card border border-border"
    >
      <h3 className="text-lg font-semibold text-text-primary mb-4">
        Plugin Details
      </h3>
      <div className="space-y-4">
        {/* Filename */}
        <div>
          <p className="text-sm text-text-muted mb-1">Filename</p>
          <p className="font-mono text-text-primary">{plugin.filename}</p>
        </div>

        {/* Type and Flags */}
        <div className="flex gap-4">
          <div>
            <p className="text-sm text-text-muted mb-1">Type</p>
            <span
              className={`px-2 py-1 rounded-full text-sm font-medium ${getPluginTypeBadgeClass(plugin.type)}`}
            >
              {plugin.type}
            </span>
          </div>
          <div>
            <p className="text-sm text-text-muted mb-1">Flags</p>
            <div className="flex gap-2">
              {plugin.flags.isMaster && (
                <span className="px-2 py-1 rounded-full text-xs bg-accent/20 text-accent">
                  Master
                </span>
              )}
              {plugin.flags.isLight && (
                <span className="px-2 py-1 rounded-full text-xs bg-warning/20 text-warning">
                  Light
                </span>
              )}
              {plugin.flags.isLocalized && (
                <span className="px-2 py-1 rounded-full text-xs bg-text-muted/20 text-text-secondary">
                  Localized
                </span>
              )}
              {!plugin.flags.isMaster && !plugin.flags.isLight && !plugin.flags.isLocalized && (
                <span className="text-text-muted text-sm">None</span>
              )}
            </div>
          </div>
        </div>

        {/* Author */}
        {plugin.author && (
          <div>
            <p className="text-sm text-text-muted mb-1">Author</p>
            <p className="text-text-primary">{plugin.author}</p>
          </div>
        )}

        {/* Description */}
        {plugin.description && (
          <div>
            <p className="text-sm text-text-muted mb-1">Description</p>
            <p className="text-text-secondary text-sm">{plugin.description}</p>
          </div>
        )}

        {/* Masters */}
        <div>
          <p className="text-sm text-text-muted mb-1">
            Masters ({masters.length})
          </p>
          {masters.length > 0 ? (
            <ul className="space-y-1">
              {masters.map((master) => (
                <li
                  key={master}
                  className="font-mono text-sm text-text-secondary"
                >
                  {master}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-text-muted text-sm">No master dependencies</p>
          )}
        </div>

        {/* Issues for this plugin */}
        {issues.length > 0 && (
          <div>
            <p className="text-sm text-text-muted mb-2">
              Issues ({issues.length})
            </p>
            <ul className="space-y-2">
              {issues.map((issue, idx) => (
                <li
                  key={idx}
                  className={`p-3 rounded-xs ${getSeverityBadgeClass(issue.severity)} bg-opacity-10`}
                >
                  <div className="flex items-center gap-2 mb-1">
                    <span
                      className={`px-2 py-0.5 rounded-full text-xs font-medium ${getSeverityBadgeClass(issue.severity)}`}
                    >
                      {issue.severity.toUpperCase()}
                    </span>
                    <span className="text-sm font-medium">
                      {getIssueTypeLabel(issue.type)}
                    </span>
                  </div>
                  <p className="text-sm">{issue.message}</p>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </section>
  );
};

// ============================================
// Warning Panel Component
// ============================================

const WarningPanel: React.FC<WarningPanelProps> = ({ issues, onSelectIssue }) => {
  const errorCount = issues.filter(i => i.severity === 'error').length;
  const warningCount = issues.filter(i => i.severity === 'warning').length;

  if (issues.length === 0) {
    return (
      <section
        aria-label="No issues"
        className="p-4 rounded-sm bg-accent/10 border border-accent"
      >
        <div className="flex items-center gap-3">
          <span className="text-2xl">&#10003;</span>
          <div>
            <p className="font-semibold text-accent">No Issues Found</p>
            <p className="text-sm text-text-secondary">
              Your load order looks healthy!
            </p>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section
      aria-label="Load order issues"
      className="p-4 rounded-sm bg-bg-card border border-border"
    >
      <h3 className="text-lg font-semibold text-text-primary mb-2">
        Issues ({issues.length})
      </h3>
      <div className="flex gap-4 mb-4 text-sm">
        {errorCount > 0 && (
          <span className="text-error">{errorCount} error{errorCount !== 1 ? 's' : ''}</span>
        )}
        {warningCount > 0 && (
          <span className="text-warning">{warningCount} warning{warningCount !== 1 ? 's' : ''}</span>
        )}
      </div>
      <ul className="space-y-2 max-h-[400px] overflow-y-auto">
        {issues.map((issue, idx) => (
          <li key={idx}>
            <button
              onClick={() => onSelectIssue(issue.plugin)}
              className={`
                w-full p-3 rounded-xs text-left
                transition-colors motion-reduce:transition-none
                focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
                ${issue.severity === 'error'
                  ? 'bg-error/10 hover:bg-error/20'
                  : 'bg-warning/10 hover:bg-warning/20'
                }
              `}
            >
              <div className="flex items-center gap-2 mb-1">
                <span
                  className={`px-2 py-0.5 rounded-full text-xs font-medium ${getSeverityBadgeClass(issue.severity)}`}
                >
                  {issue.severity.toUpperCase()}
                </span>
                <span className="text-sm font-medium text-text-primary">
                  {getIssueTypeLabel(issue.type)}
                </span>
              </div>
              <p className="text-sm text-text-secondary">{issue.message}</p>
              <p className="text-xs text-text-muted mt-1 font-mono">
                {issue.plugin}
              </p>
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
};

// ============================================
// Main LoadOrderView Component
// ============================================

/** Container component for load order analysis view */
export const LoadOrderView: React.FC<LoadOrderViewProps> = ({ slug, revision }) => {
  const { data, isLoading, error, refetch } = useLoadOrderAnalysis(slug, revision);
  const [selectedPlugin, setSelectedPlugin] = useState<LoadOrderPluginInfo | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>('list');

  // Build map of plugin issues
  const issues = data?.issues;
  const pluginIssues = useMemo(() => {
    if (!issues) return new Map<string, LoadOrderIssue[]>();
    return buildPluginIssuesMap(issues);
  }, [issues]);

  // Handle selecting a plugin from issue panel
  const handleSelectIssue = (pluginFilename: string) => {
    const plugin = data?.plugins.find(p => p.filename === pluginFilename);
    if (plugin) {
      setSelectedPlugin(plugin);
    }
  };

  // Loading state
  if (isLoading) {
    return <LoadOrderSkeleton />;
  }

  // Error state
  if (error) {
    return <ErrorDisplay error={error} onRetry={() => refetch()} />;
  }

  // No data state
  if (!data) {
    return <EmptyState />;
  }

  // No plugins state
  if (data.plugins.length === 0) {
    return <EmptyState />;
  }

  return (
    <div className="space-y-6">
      {/* Announce data loaded to screen readers */}
      <div aria-live="polite" className="sr-only">
        Load order analysis complete. {data.stats.totalPlugins} plugins found
        with {data.stats.totalIssues} issues.
      </div>

      {/* Stats header with view mode toggle */}
      <LoadOrderHeader
        stats={data.stats}
        cached={data.cached}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
      />

      {/* Slot limit warning */}
      <SlotLimitWarning
        esmCount={data.stats.esmCount}
        espCount={data.stats.espCount}
      />

      {/* View mode: Graph */}
      {viewMode === 'graph' && (
        <DependencyGraphView
          plugins={data.plugins}
          dependencyGraph={data.dependencyGraph}
          issues={data.issues}
          selectedPlugin={selectedPlugin}
          onSelectPlugin={setSelectedPlugin}
        />
      )}

      {/* View mode: List */}
      {viewMode === 'list' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Plugin list */}
          <div className="lg:col-span-2">
            <LoadOrderList
              plugins={data.plugins}
              selectedPlugin={selectedPlugin}
              onSelectPlugin={setSelectedPlugin}
              pluginIssues={pluginIssues}
            />
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Plugin details or warning panel */}
            {selectedPlugin ? (
              <LoadOrderDetails
                plugin={selectedPlugin}
                issues={pluginIssues.get(selectedPlugin.filename) ?? []}
                dependencyGraph={data.dependencyGraph}
              />
            ) : (
              <WarningPanel
                issues={data.issues}
                onSelectIssue={handleSelectIssue}
              />
            )}
          </div>
        </div>
      )}

      {/* Show details panel below graph when a plugin is selected */}
      {viewMode === 'graph' && selectedPlugin && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <LoadOrderDetails
            plugin={selectedPlugin}
            issues={pluginIssues.get(selectedPlugin.filename) ?? []}
            dependencyGraph={data.dependencyGraph}
          />
          <WarningPanel
            issues={data.issues}
            onSelectIssue={handleSelectIssue}
          />
        </div>
      )}
    </div>
  );
};
