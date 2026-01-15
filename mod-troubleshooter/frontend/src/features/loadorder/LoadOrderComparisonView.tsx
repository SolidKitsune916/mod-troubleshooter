import { useState, useMemo, useId } from 'react';

import type {
  LoadOrderPluginInfo,
  LoadOrderPluginType,
} from '@/types/index.ts';

// ============================================
// Types
// ============================================

/** Load order snapshot for comparison */
export interface LoadOrderSnapshot {
  id: string;
  label: string;
  plugins: LoadOrderPluginInfo[];
  savedAt: Date;
}

/** Props for LoadOrderComparisonView */
interface LoadOrderComparisonViewProps {
  currentPlugins: LoadOrderPluginInfo[];
  onSaveSnapshotA: () => void;
  onSaveSnapshotB: () => void;
  snapshotA: LoadOrderSnapshot | null;
  snapshotB: LoadOrderSnapshot | null;
}

/** Plugin diff status */
type PluginDiffStatus = 'a-only' | 'b-only' | 'same' | 'moved';

/** Plugin with diff information */
interface DiffPlugin {
  filename: string;
  type: LoadOrderPluginType;
  indexA: number | null;
  indexB: number | null;
  diffStatus: PluginDiffStatus;
  positionDelta: number; // Positive = moved down, negative = moved up
}

interface DiffSummary {
  aOnly: number;
  bOnly: number;
  same: number;
  moved: number;
  total: number;
}

// ============================================
// Helper Functions
// ============================================

/**
 * Compare two plugin lists and generate diff information.
 */
function compareLoadOrders(
  pluginsA: LoadOrderPluginInfo[],
  pluginsB: LoadOrderPluginInfo[],
): { diffPlugins: DiffPlugin[]; summary: DiffSummary } {
  // Create maps by filename
  const mapA = new Map<string, LoadOrderPluginInfo>();
  const mapB = new Map<string, LoadOrderPluginInfo>();

  for (const plugin of pluginsA) {
    mapA.set(plugin.filename.toLowerCase(), plugin);
  }

  for (const plugin of pluginsB) {
    mapB.set(plugin.filename.toLowerCase(), plugin);
  }

  const diffPlugins: DiffPlugin[] = [];
  const processedKeys = new Set<string>();

  let aOnly = 0;
  let bOnly = 0;
  let same = 0;
  let moved = 0;

  // Process plugins from A
  for (const [key, pluginA] of mapA) {
    processedKeys.add(key);
    const pluginB = mapB.get(key);

    if (!pluginB) {
      diffPlugins.push({
        filename: pluginA.filename,
        type: pluginA.type,
        indexA: pluginA.index,
        indexB: null,
        diffStatus: 'a-only',
        positionDelta: 0,
      });
      aOnly++;
    } else if (pluginA.index === pluginB.index) {
      diffPlugins.push({
        filename: pluginA.filename,
        type: pluginA.type,
        indexA: pluginA.index,
        indexB: pluginB.index,
        diffStatus: 'same',
        positionDelta: 0,
      });
      same++;
    } else {
      diffPlugins.push({
        filename: pluginA.filename,
        type: pluginA.type,
        indexA: pluginA.index,
        indexB: pluginB.index,
        diffStatus: 'moved',
        positionDelta: pluginB.index - pluginA.index,
      });
      moved++;
    }
  }

  // Process plugins only in B
  for (const [key, pluginB] of mapB) {
    if (!processedKeys.has(key)) {
      diffPlugins.push({
        filename: pluginB.filename,
        type: pluginB.type,
        indexA: null,
        indexB: pluginB.index,
        diffStatus: 'b-only',
        positionDelta: 0,
      });
      bOnly++;
    }
  }

  // Sort by: a-only first (by A index), then same/moved (by A index), then b-only (by B index)
  diffPlugins.sort((a, b) => {
    // Group by status first
    const statusOrder = { 'a-only': 0, 'same': 1, 'moved': 1, 'b-only': 2 };
    const orderDiff = statusOrder[a.diffStatus] - statusOrder[b.diffStatus];
    if (orderDiff !== 0) return orderDiff;

    // Then by index
    const indexA = a.indexA ?? a.indexB ?? 0;
    const indexB = b.indexA ?? b.indexB ?? 0;
    return indexA - indexB;
  });

  return {
    diffPlugins,
    summary: {
      aOnly,
      bOnly,
      same,
      moved,
      total: diffPlugins.length,
    },
  };
}

/**
 * Get display class for diff status.
 */
function getDiffStatusClass(status: PluginDiffStatus): string {
  switch (status) {
    case 'a-only':
      return 'bg-accent/10 border-accent/30';
    case 'b-only':
      return 'bg-warning/10 border-warning/30';
    case 'moved':
      return 'bg-error/10 border-error/30';
    case 'same':
      return 'bg-text-muted/5 border-border';
  }
}

/**
 * Get badge for diff status.
 */
function getDiffStatusBadge(status: PluginDiffStatus): { label: string; class: string } {
  switch (status) {
    case 'a-only':
      return { label: 'A only', class: 'bg-accent/20 text-accent' };
    case 'b-only':
      return { label: 'B only', class: 'bg-warning/20 text-warning' };
    case 'moved':
      return { label: 'Moved', class: 'bg-error/20 text-error' };
    case 'same':
      return { label: 'Same', class: 'bg-text-muted/20 text-text-muted' };
  }
}

/**
 * Get badge color class for plugin type.
 */
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

/**
 * Check if two plugin lists have the same content.
 */
function pluginListsEqual(a: LoadOrderPluginInfo[], b: LoadOrderPluginInfo[]): boolean {
  if (a.length !== b.length) return false;

  for (let i = 0; i < a.length; i++) {
    if (a[i].filename !== b[i].filename || a[i].index !== b[i].index) {
      return false;
    }
  }

  return true;
}

// ============================================
// Component Props
// ============================================

interface SnapshotCardProps {
  snapshot: LoadOrderSnapshot | null;
  label: string;
  isCurrent: boolean;
  onSave: () => void;
}

interface DiffSummaryPanelProps {
  summary: DiffSummary;
  filter: PluginDiffStatus | 'all';
  onFilterChange: (filter: PluginDiffStatus | 'all') => void;
}

interface PluginDiffListProps {
  plugins: DiffPlugin[];
  showFilter: PluginDiffStatus | 'all';
}

// ============================================
// Snapshot Card Component
// ============================================

const SnapshotCard: React.FC<SnapshotCardProps> = ({
  snapshot,
  label,
  isCurrent,
  onSave,
}) => {
  const labelId = useId();
  const pluginCount = snapshot?.plugins.length ?? 0;

  return (
    <div
      className={`
        p-4 rounded-sm border
        ${snapshot ? 'bg-bg-card border-border' : 'bg-bg-secondary/30 border-dashed border-text-muted/30'}
      `}
    >
      <div className="flex items-center justify-between mb-3">
        <h4
          id={labelId}
          className="font-semibold text-text-primary flex items-center gap-2"
        >
          <span
            aria-hidden="true"
            className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold
              ${label === 'Load Order A' ? 'bg-accent/20 text-accent' : 'bg-warning/20 text-warning'}`}
          >
            {label === 'Load Order A' ? 'A' : 'B'}
          </span>
          {label}
        </h4>
        {isCurrent && (
          <span className="px-2 py-0.5 rounded-full text-xs bg-accent/20 text-accent">
            Current
          </span>
        )}
      </div>

      {snapshot ? (
        <div className="space-y-2 mb-3">
          <p className="text-sm text-text-secondary">
            <span className="text-text-muted">Saved:</span>{' '}
            {snapshot.savedAt.toLocaleTimeString()}
          </p>
          <p className="text-sm text-text-secondary">
            <span className="text-text-muted">Plugins:</span> {pluginCount}
          </p>
        </div>
      ) : (
        <p className="text-sm text-text-muted mb-3">
          No snapshot saved yet. Save your current load order to compare.
        </p>
      )}

      <button
        onClick={onSave}
        aria-describedby={labelId}
        className="min-h-9 px-4 py-1.5 rounded-sm text-sm font-medium w-full
          bg-bg-secondary text-text-secondary
          hover:bg-bg-secondary/80 hover:text-text-primary
          focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
          transition-colors motion-reduce:transition-none"
      >
        {snapshot ? 'Update Snapshot' : 'Save Current'}
      </button>
    </div>
  );
};

// ============================================
// Diff Summary Panel Component
// ============================================

const DiffSummaryPanel: React.FC<DiffSummaryPanelProps> = ({
  summary,
  filter,
  onFilterChange,
}) => {
  const filters: { value: PluginDiffStatus | 'all'; label: string; count: number }[] = [
    { value: 'all', label: 'All', count: summary.total },
    { value: 'a-only', label: 'A Only', count: summary.aOnly },
    { value: 'b-only', label: 'B Only', count: summary.bOnly },
    { value: 'moved', label: 'Moved', count: summary.moved },
    { value: 'same', label: 'Same', count: summary.same },
  ];

  return (
    <div
      role="group"
      aria-label="Filter plugin differences"
      className="flex flex-wrap gap-2 p-3 rounded-sm bg-bg-secondary/50 border border-border"
    >
      {filters.map((f) => (
        <button
          key={f.value}
          onClick={() => onFilterChange(f.value)}
          aria-pressed={filter === f.value}
          className={`
            min-h-9 px-4 py-1.5 rounded-sm text-sm font-medium
            flex items-center gap-2
            focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
            transition-colors motion-reduce:transition-none
            ${
              filter === f.value
                ? 'bg-accent text-white'
                : 'bg-bg-card text-text-secondary hover:bg-bg-hover'
            }
          `}
        >
          {f.label}
          <span
            className={`
              px-1.5 py-0.5 rounded-full text-xs
              ${filter === f.value ? 'bg-white/20' : 'bg-bg-secondary'}
            `}
          >
            {f.count}
          </span>
        </button>
      ))}
    </div>
  );
};

// ============================================
// Plugin Diff List Component
// ============================================

const PluginDiffList: React.FC<PluginDiffListProps> = ({ plugins, showFilter }) => {
  const listId = useId();

  const filteredPlugins = useMemo(() => {
    if (showFilter === 'all') return plugins;
    return plugins.filter((p) => p.diffStatus === showFilter);
  }, [plugins, showFilter]);

  if (filteredPlugins.length === 0) {
    return (
      <div className="p-4 text-center text-text-muted">
        No plugins matching the current filter.
      </div>
    );
  }

  return (
    <div className="max-h-96 overflow-y-auto">
      <ul
        id={listId}
        role="list"
        aria-label="Plugin differences"
        className="divide-y divide-border"
      >
        {filteredPlugins.map((plugin) => {
          const badge = getDiffStatusBadge(plugin.diffStatus);
          return (
            <li
              key={plugin.filename}
              className={`
                p-3 flex items-center gap-3 border-l-2
                ${getDiffStatusClass(plugin.diffStatus)}
              `}
            >
              {/* Index columns */}
              <div className="flex gap-2 shrink-0">
                <span
                  className="w-12 text-center text-sm font-mono"
                  title="Position in A"
                >
                  {plugin.indexA !== null ? (
                    <span className="text-accent">
                      {String(plugin.indexA).padStart(3, '0')}
                    </span>
                  ) : (
                    <span className="text-text-muted">---</span>
                  )}
                </span>
                <span className="text-text-muted">→</span>
                <span
                  className="w-12 text-center text-sm font-mono"
                  title="Position in B"
                >
                  {plugin.indexB !== null ? (
                    <span className="text-warning">
                      {String(plugin.indexB).padStart(3, '0')}
                    </span>
                  ) : (
                    <span className="text-text-muted">---</span>
                  )}
                </span>
              </div>

              {/* Plugin type badge */}
              <span
                className={`px-2 py-0.5 rounded-full text-xs font-medium shrink-0 ${getPluginTypeBadgeClass(plugin.type)}`}
              >
                {plugin.type}
              </span>

              {/* Plugin filename */}
              <span className="flex-1 font-mono text-sm text-text-primary truncate" title={plugin.filename}>
                {plugin.filename}
              </span>

              {/* Position delta for moved plugins */}
              {plugin.diffStatus === 'moved' && plugin.positionDelta !== 0 && (
                <span
                  className={`text-xs font-medium shrink-0 ${
                    plugin.positionDelta > 0 ? 'text-error' : 'text-accent'
                  }`}
                  title={plugin.positionDelta > 0 ? 'Moved down' : 'Moved up'}
                >
                  {plugin.positionDelta > 0 ? '↓' : '↑'}
                  {Math.abs(plugin.positionDelta)}
                </span>
              )}

              {/* Status badge */}
              <span className={`shrink-0 px-2 py-0.5 rounded-full text-xs font-medium ${badge.class}`}>
                {badge.label}
              </span>
            </li>
          );
        })}
      </ul>
    </div>
  );
};

// ============================================
// No Comparison State Component
// ============================================

const NoComparisonState: React.FC = () => (
  <div className="p-8 text-center rounded-sm bg-bg-card border border-border">
    <div className="text-4xl mb-4" aria-hidden="true">📊</div>
    <h3 className="text-lg font-semibold text-text-primary mb-2">
      No Load Orders to Compare
    </h3>
    <p className="text-text-muted max-w-md mx-auto">
      Save at least two load order snapshots using the buttons above to compare
      how different mod setups affect plugin positions and presence.
    </p>
  </div>
);

// ============================================
// Main LoadOrderComparisonView Component
// ============================================

/**
 * Side-by-side comparison view for load orders.
 * Allows users to save two different load order snapshots and
 * see the differences in plugin presence and positions.
 */
export const LoadOrderComparisonView: React.FC<LoadOrderComparisonViewProps> = ({
  currentPlugins,
  onSaveSnapshotA,
  onSaveSnapshotB,
  snapshotA,
  snapshotB,
}) => {
  const [filter, setFilter] = useState<PluginDiffStatus | 'all'>('all');
  const headingId = useId();

  // Check if current plugins match either snapshot
  const isCurrentA = useMemo(() => {
    if (!snapshotA) return false;
    return pluginListsEqual(currentPlugins, snapshotA.plugins);
  }, [currentPlugins, snapshotA]);

  const isCurrentB = useMemo(() => {
    if (!snapshotB) return false;
    return pluginListsEqual(currentPlugins, snapshotB.plugins);
  }, [currentPlugins, snapshotB]);

  // Calculate diff when both snapshots exist
  const diff = useMemo(() => {
    if (!snapshotA || !snapshotB) return null;
    return compareLoadOrders(snapshotA.plugins, snapshotB.plugins);
  }, [snapshotA, snapshotB]);

  return (
    <section
      aria-labelledby={headingId}
      className="space-y-6"
    >
      <div className="p-4 rounded-sm bg-bg-card border border-border">
        <h3
          id={headingId}
          className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2"
        >
          <span aria-hidden="true">📊</span>
          Load Order Comparison
        </h3>

        <p className="text-sm text-text-muted mb-4">
          Save your current load order to Snapshot A or B, then modify your collection
          and save again to see how plugin positions and presence have changed.
        </p>

        {/* Snapshot cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <SnapshotCard
            snapshot={snapshotA}
            label="Load Order A"
            isCurrent={isCurrentA}
            onSave={onSaveSnapshotA}
          />
          <SnapshotCard
            snapshot={snapshotB}
            label="Load Order B"
            isCurrent={isCurrentB}
            onSave={onSaveSnapshotB}
          />
        </div>
      </div>

      {/* Diff view */}
      {diff ? (
        <div className="space-y-4">
          {/* Summary and filters */}
          <DiffSummaryPanel
            summary={diff.summary}
            filter={filter}
            onFilterChange={setFilter}
          />

          {/* Screen reader summary */}
          <div aria-live="polite" className="sr-only">
            Comparing load orders: {diff.summary.aOnly} plugins only in A,
            {diff.summary.bOnly} plugins only in B,
            {diff.summary.moved} plugins with different positions,
            {diff.summary.same} plugins unchanged.
          </div>

          {/* Plugin diff list */}
          <div className="rounded-sm border border-border overflow-hidden bg-bg-card">
            <div className="px-4 py-3 bg-bg-secondary border-b border-border">
              <h4 className="font-medium text-text-primary flex items-center justify-between">
                <span>Plugin Comparison</span>
                <span className="text-sm font-normal text-text-muted">
                  {diff.summary.total} plugins total
                </span>
              </h4>
            </div>
            <PluginDiffList
              plugins={diff.diffPlugins}
              showFilter={filter}
            />
          </div>

          {/* Legend */}
          <div className="flex flex-wrap gap-4 text-xs text-text-muted p-3 rounded-sm bg-bg-secondary/30">
            <span className="flex items-center gap-1">
              <span className="w-3 h-3 rounded bg-accent/30" aria-hidden="true" />
              A Only: Plugins present only in Load Order A
            </span>
            <span className="flex items-center gap-1">
              <span className="w-3 h-3 rounded bg-warning/30" aria-hidden="true" />
              B Only: Plugins present only in Load Order B
            </span>
            <span className="flex items-center gap-1">
              <span className="w-3 h-3 rounded bg-error/30" aria-hidden="true" />
              Moved: Same plugin at different position
            </span>
            <span className="flex items-center gap-1">
              <span className="w-3 h-3 rounded bg-text-muted/20" aria-hidden="true" />
              Same: Unchanged plugin at same position
            </span>
          </div>
        </div>
      ) : (
        <NoComparisonState />
      )}
    </section>
  );
};
