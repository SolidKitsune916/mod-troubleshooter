import { useState, useCallback, useMemo, useId } from 'react';

import type {
  InstallStep,
  ModuleConfig,
} from '@/types/index.ts';
import type { FlagState, SelectionsMap, InstallFile } from './fomodUtils.ts';

// ============================================
// Types
// ============================================

/** Configuration snapshot for comparison */
export interface ConfigSnapshot {
  id: string;
  label: string;
  selections: SelectionsMap;
  files: InstallFile[];
  savedAt: Date;
}

/** Props for FomodComparisonView */
interface FomodComparisonViewProps {
  config: ModuleConfig;
  steps: InstallStep[];
  currentSelections: SelectionsMap;
  currentFlags: FlagState;
  onSaveConfigA: () => void;
  onSaveConfigB: () => void;
  onLoadConfig: (selections: SelectionsMap) => void;
  configA: ConfigSnapshot | null;
  configB: ConfigSnapshot | null;
}

/** File diff status */
type DiffStatus = 'a-only' | 'b-only' | 'both' | 'different';

/** File with diff information */
interface DiffFile extends InstallFile {
  diffStatus: DiffStatus;
  otherSource?: string; // Different source if 'different' status
}

interface DiffSummary {
  aOnly: number;
  bOnly: number;
  both: number;
  different: number;
}

// ============================================
// Helper Functions
// ============================================

/**
 * Compare two file lists and generate diff information.
 */
function compareFileLists(
  filesA: InstallFile[],
  filesB: InstallFile[],
): { diffA: DiffFile[]; diffB: DiffFile[]; summary: DiffSummary } {
  // Create maps by destination path
  const mapA = new Map<string, InstallFile>();
  const mapB = new Map<string, InstallFile>();

  for (const file of filesA) {
    const key = file.destination.toLowerCase().replace(/\\/g, '/');
    mapA.set(key, file);
  }

  for (const file of filesB) {
    const key = file.destination.toLowerCase().replace(/\\/g, '/');
    mapB.set(key, file);
  }

  const diffA: DiffFile[] = [];
  const diffB: DiffFile[] = [];

  let aOnly = 0;
  let bOnly = 0;
  let both = 0;
  let different = 0;

  // Process files from A
  for (const [key, fileA] of mapA) {
    const fileB = mapB.get(key);

    if (!fileB) {
      diffA.push({ ...fileA, diffStatus: 'a-only' });
      aOnly++;
    } else if (fileA.source.toLowerCase() !== fileB.source.toLowerCase()) {
      diffA.push({ ...fileA, diffStatus: 'different', otherSource: fileB.source });
      different++;
    } else {
      diffA.push({ ...fileA, diffStatus: 'both' });
      both++;
    }
  }

  // Process files from B
  for (const [key, fileB] of mapB) {
    const fileA = mapA.get(key);

    if (!fileA) {
      diffB.push({ ...fileB, diffStatus: 'b-only' });
      bOnly++;
    } else if (fileB.source.toLowerCase() !== fileA.source.toLowerCase()) {
      diffB.push({ ...fileB, diffStatus: 'different', otherSource: fileA.source });
    } else {
      diffB.push({ ...fileB, diffStatus: 'both' });
    }
  }

  // Sort by destination
  const sortFn = (a: DiffFile, b: DiffFile) =>
    a.destination.localeCompare(b.destination);

  diffA.sort(sortFn);
  diffB.sort(sortFn);

  return {
    diffA,
    diffB,
    summary: { aOnly, bOnly, both, different },
  };
}

/**
 * Get display class for diff status.
 */
function getDiffStatusClass(status: DiffStatus): string {
  switch (status) {
    case 'a-only':
      return 'bg-accent/10 border-accent/30 text-accent';
    case 'b-only':
      return 'bg-warning/10 border-warning/30 text-warning';
    case 'different':
      return 'bg-error/10 border-error/30 text-error';
    case 'both':
      return 'bg-text-muted/5 border-border text-text-secondary';
  }
}

/**
 * Get badge for diff status.
 */
function getDiffStatusBadge(status: DiffStatus): { label: string; class: string } {
  switch (status) {
    case 'a-only':
      return { label: 'A only', class: 'bg-accent/20 text-accent' };
    case 'b-only':
      return { label: 'B only', class: 'bg-warning/20 text-warning' };
    case 'different':
      return { label: 'Different', class: 'bg-error/20 text-error' };
    case 'both':
      return { label: 'Same', class: 'bg-text-muted/20 text-text-muted' };
  }
}

/**
 * Check if two SelectionsMaps are equal.
 */
function selectionsEqual(a: SelectionsMap, b: SelectionsMap): boolean {
  if (a.size !== b.size) return false;

  for (const [key, setA] of a) {
    const setB = b.get(key);
    if (!setB || setA.size !== setB.size) return false;
    for (const item of setA) {
      if (!setB.has(item)) return false;
    }
  }

  return true;
}

// ============================================
// Component Props
// ============================================

interface ConfigCardProps {
  config: ConfigSnapshot | null;
  label: string;
  isCurrent: boolean;
  onSave: () => void;
  onLoad: () => void;
}

interface FileDiffListProps {
  files: DiffFile[];
  label: string;
  showFilter: DiffStatus | 'all';
}

interface DiffSummaryPanelProps {
  summary: DiffSummary;
  filter: DiffStatus | 'all';
  onFilterChange: (filter: DiffStatus | 'all') => void;
}

// ============================================
// Config Card Component
// ============================================

const ConfigCard: React.FC<ConfigCardProps> = ({
  config,
  label,
  isCurrent,
  onSave,
  onLoad,
}) => {
  const labelId = useId();
  const fileCount = config?.files.length ?? 0;
  const selectionCount = config
    ? Array.from(config.selections.values()).reduce((sum, set) => sum + set.size, 0)
    : 0;

  return (
    <div
      className={`
        p-4 rounded-sm border
        ${config ? 'bg-bg-card border-border' : 'bg-bg-secondary/30 border-dashed border-text-muted/30'}
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
              ${label === 'Configuration A' ? 'bg-accent/20 text-accent' : 'bg-warning/20 text-warning'}`}
          >
            {label === 'Configuration A' ? 'A' : 'B'}
          </span>
          {label}
        </h4>
        {isCurrent && (
          <span className="px-2 py-0.5 rounded-full text-xs bg-accent/20 text-accent">
            Current
          </span>
        )}
      </div>

      {config ? (
        <div className="space-y-2 mb-3">
          <p className="text-sm text-text-secondary">
            <span className="text-text-muted">Saved:</span>{' '}
            {config.savedAt.toLocaleTimeString()}
          </p>
          <p className="text-sm text-text-secondary">
            <span className="text-text-muted">Selections:</span> {selectionCount}
          </p>
          <p className="text-sm text-text-secondary">
            <span className="text-text-muted">Files:</span> {fileCount}
          </p>
        </div>
      ) : (
        <p className="text-sm text-text-muted mb-3">
          No configuration saved yet. Save your current selections to compare.
        </p>
      )}

      <div className="flex gap-2">
        <button
          onClick={onSave}
          aria-describedby={labelId}
          className="min-h-9 px-4 py-1.5 rounded-sm text-sm font-medium flex-1
            bg-bg-secondary text-text-secondary
            hover:bg-bg-secondary/80 hover:text-text-primary
            focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
            transition-colors motion-reduce:transition-none"
        >
          {config ? 'Update' : 'Save Current'}
        </button>
        {config && (
          <button
            onClick={onLoad}
            aria-describedby={labelId}
            className="min-h-9 px-4 py-1.5 rounded-sm text-sm font-medium
              bg-bg-secondary text-text-secondary
              hover:bg-bg-secondary/80 hover:text-text-primary
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none"
          >
            Load
          </button>
        )}
      </div>
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
  const filters: { value: DiffStatus | 'all'; label: string; count: number }[] = [
    { value: 'all', label: 'All', count: summary.aOnly + summary.bOnly + summary.both + summary.different },
    { value: 'a-only', label: 'A Only', count: summary.aOnly },
    { value: 'b-only', label: 'B Only', count: summary.bOnly },
    { value: 'different', label: 'Different Source', count: summary.different },
    { value: 'both', label: 'Same', count: summary.both },
  ];

  return (
    <div
      role="group"
      aria-label="Filter file differences"
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
// File Diff List Component
// ============================================

const FileDiffList: React.FC<FileDiffListProps> = ({ files, label, showFilter }) => {
  const listId = useId();

  const filteredFiles = useMemo(() => {
    if (showFilter === 'all') return files;
    return files.filter((f) => f.diffStatus === showFilter);
  }, [files, showFilter]);

  if (filteredFiles.length === 0) {
    return (
      <div className="p-4 text-center text-text-muted">
        No files matching the current filter.
      </div>
    );
  }

  return (
    <div className="max-h-96 overflow-y-auto">
      <ul
        id={listId}
        role="list"
        aria-label={label}
        className="divide-y divide-border"
      >
        {filteredFiles.map((file, index) => {
          const badge = getDiffStatusBadge(file.diffStatus);
          return (
            <li
              key={`${file.destination}-${index}`}
              className={`
                p-3 flex flex-col gap-1
                ${getDiffStatusClass(file.diffStatus)}
              `}
            >
              <div className="flex items-start gap-2">
                <span aria-hidden="true" className="text-text-muted shrink-0">
                  {file.isFolder ? '📁' : '📄'}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="font-mono text-sm truncate" title={file.destination}>
                    {file.destination}
                  </p>
                  {file.diffStatus === 'different' && (
                    <p className="text-xs text-text-muted mt-1">
                      Source: <span className="text-text-secondary">{file.source}</span>
                      {file.otherSource && (
                        <span className="block">
                          Other: <span className="text-text-secondary">{file.otherSource}</span>
                        </span>
                      )}
                    </p>
                  )}
                </div>
                <span className={`shrink-0 px-2 py-0.5 rounded-full text-xs font-medium ${badge.class}`}>
                  {badge.label}
                </span>
              </div>
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
    <div className="text-4xl mb-4" aria-hidden="true">⚖️</div>
    <h3 className="text-lg font-semibold text-text-primary mb-2">
      No Configurations to Compare
    </h3>
    <p className="text-text-muted max-w-md mx-auto">
      Save at least two configurations using the buttons above to compare
      how different option selections affect the files that will be installed.
    </p>
  </div>
);

// ============================================
// Main FomodComparisonView Component
// ============================================

/**
 * Side-by-side comparison view for FOMOD configurations.
 * Allows users to save two different selection configurations and
 * see the differences in files that will be installed.
 */
export const FomodComparisonView: React.FC<FomodComparisonViewProps> = ({
  // config, steps, and currentFlags are used by parent for saving configs
  config: _config,
  steps: _steps,
  currentSelections,
  currentFlags: _currentFlags,
  onSaveConfigA,
  onSaveConfigB,
  onLoadConfig,
  configA,
  configB,
}) => {
  // These are intentionally unused in this component - they're used by parent
  void _config;
  void _steps;
  void _currentFlags;

  const [filter, setFilter] = useState<DiffStatus | 'all'>('all');
  const headingId = useId();

  // Check if current selections match either config
  const isCurrentA = useMemo(() => {
    if (!configA) return false;
    return selectionsEqual(currentSelections, configA.selections);
  }, [currentSelections, configA]);

  const isCurrentB = useMemo(() => {
    if (!configB) return false;
    return selectionsEqual(currentSelections, configB.selections);
  }, [currentSelections, configB]);

  // Calculate diff when both configs exist
  const diff = useMemo(() => {
    if (!configA || !configB) return null;
    return compareFileLists(configA.files, configB.files);
  }, [configA, configB]);

  const handleLoadA = useCallback(() => {
    if (configA) {
      onLoadConfig(configA.selections);
    }
  }, [configA, onLoadConfig]);

  const handleLoadB = useCallback(() => {
    if (configB) {
      onLoadConfig(configB.selections);
    }
  }, [configB, onLoadConfig]);

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
          <span aria-hidden="true">⚖️</span>
          Configuration Comparison
        </h3>

        <p className="text-sm text-text-muted mb-4">
          Save your current selections to Configuration A or B, then switch between
          them to compare which files each configuration will install.
        </p>

        {/* Configuration cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <ConfigCard
            config={configA}
            label="Configuration A"
            isCurrent={isCurrentA}
            onSave={onSaveConfigA}
            onLoad={handleLoadA}
          />
          <ConfigCard
            config={configB}
            label="Configuration B"
            isCurrent={isCurrentB}
            onSave={onSaveConfigB}
            onLoad={handleLoadB}
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
            Comparing configurations: {diff.summary.aOnly} files only in A,
            {diff.summary.bOnly} files only in B,
            {diff.summary.different} files with different sources,
            {diff.summary.both} files the same in both.
          </div>

          {/* Side-by-side file lists */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <div className="rounded-sm border border-accent/30 overflow-hidden">
              <div className="px-4 py-2 bg-accent/10 border-b border-accent/30">
                <h4 className="font-medium text-accent flex items-center gap-2">
                  <span
                    aria-hidden="true"
                    className="w-5 h-5 rounded-full bg-accent/20 flex items-center justify-center text-xs font-bold"
                  >
                    A
                  </span>
                  Configuration A Files
                  <span className="text-sm font-normal text-text-muted">
                    ({diff.diffA.length} total)
                  </span>
                </h4>
              </div>
              <FileDiffList
                files={diff.diffA}
                label="Configuration A files"
                showFilter={filter}
              />
            </div>

            <div className="rounded-sm border border-warning/30 overflow-hidden">
              <div className="px-4 py-2 bg-warning/10 border-b border-warning/30">
                <h4 className="font-medium text-warning flex items-center gap-2">
                  <span
                    aria-hidden="true"
                    className="w-5 h-5 rounded-full bg-warning/20 flex items-center justify-center text-xs font-bold"
                  >
                    B
                  </span>
                  Configuration B Files
                  <span className="text-sm font-normal text-text-muted">
                    ({diff.diffB.length} total)
                  </span>
                </h4>
              </div>
              <FileDiffList
                files={diff.diffB}
                label="Configuration B files"
                showFilter={filter}
              />
            </div>
          </div>

          {/* Legend */}
          <div className="flex flex-wrap gap-4 text-xs text-text-muted p-3 rounded-sm bg-bg-secondary/30">
            <span className="flex items-center gap-1">
              <span className="w-3 h-3 rounded bg-accent/30" aria-hidden="true" />
              A Only: Files unique to Configuration A
            </span>
            <span className="flex items-center gap-1">
              <span className="w-3 h-3 rounded bg-warning/30" aria-hidden="true" />
              B Only: Files unique to Configuration B
            </span>
            <span className="flex items-center gap-1">
              <span className="w-3 h-3 rounded bg-error/30" aria-hidden="true" />
              Different: Same destination, different source
            </span>
            <span className="flex items-center gap-1">
              <span className="w-3 h-3 rounded bg-text-muted/20" aria-hidden="true" />
              Same: Identical in both configurations
            </span>
          </div>
        </div>
      ) : (
        <NoComparisonState />
      )}
    </section>
  );
};
