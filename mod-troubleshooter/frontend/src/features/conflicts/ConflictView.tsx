import { useState, useMemo } from 'react';

import { useConflictAnalysis } from '@hooks/useConflicts.ts';
import { ApiError } from '@services/api.ts';
import { ConflictGraphView } from './ConflictGraphView.tsx';

import type {
  Conflict,
  ConflictStats,
  ConflictSeverity,
  FileType,
  ModConflictSummary,
} from '@/types/index.ts';

/** View mode for the conflict display */
type ViewMode = 'list' | 'graph';

// ============================================
// Props Interfaces
// ============================================

interface ConflictViewProps {
  slug: string;
  revision: number;
}

interface ConflictHeaderProps {
  stats: ConflictStats;
  cached: boolean;
  exportToolbar?: React.ReactNode;
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
}

interface ConflictFiltersProps {
  selectedSeverity: ConflictSeverity | null;
  selectedFileType: FileType | null;
  selectedMod: string | null;
  searchQuery: string;
  onSeverityChange: (severity: ConflictSeverity | null) => void;
  onFileTypeChange: (fileType: FileType | null) => void;
  onModChange: (modId: string | null) => void;
  onSearchChange: (query: string) => void;
  modSummaries: ModConflictSummary[];
  stats: ConflictStats;
}

interface ConflictListProps {
  conflicts: Conflict[];
  selectedConflict: Conflict | null;
  onSelectConflict: (conflict: Conflict | null) => void;
}

interface ConflictRowProps {
  conflict: Conflict;
  isSelected: boolean;
  onSelect: () => void;
}

interface ConflictDetailsProps {
  conflict: Conflict;
}

// ============================================
// Helper Functions
// ============================================

/** Get badge color class for severity */
function getSeverityBadgeClass(severity: ConflictSeverity): string {
  switch (severity) {
    case 'critical':
      return 'bg-error/20 text-error';
    case 'high':
      return 'bg-warning/20 text-warning';
    case 'medium':
      return 'bg-accent/20 text-accent';
    case 'low':
      return 'bg-text-muted/20 text-text-secondary';
    case 'info':
      return 'bg-accent/10 text-accent';
    default:
      return 'bg-text-muted/20 text-text-secondary';
  }
}

/** Get human-readable severity label */
function getSeverityLabel(severity: ConflictSeverity): string {
  return severity.charAt(0).toUpperCase() + severity.slice(1);
}

/** Get badge color class for file type */
function getFileTypeBadgeClass(fileType: FileType): string {
  switch (fileType) {
    case 'plugin':
      return 'bg-error/20 text-error';
    case 'bsa':
      return 'bg-warning/20 text-warning';
    case 'script':
      return 'bg-accent/20 text-accent';
    case 'mesh':
    case 'texture':
      return 'bg-text-muted/20 text-text-secondary';
    case 'interface':
      return 'bg-accent/20 text-accent';
    default:
      return 'bg-text-muted/20 text-text-secondary';
  }
}

/** Get human-readable file type label */
function getFileTypeLabel(fileType: FileType): string {
  switch (fileType) {
    case 'plugin':
      return 'Plugin';
    case 'mesh':
      return 'Mesh';
    case 'texture':
      return 'Texture';
    case 'sound':
      return 'Sound';
    case 'script':
      return 'Script';
    case 'interface':
      return 'Interface';
    case 'seq':
      return 'SEQ';
    case 'bsa':
      return 'BSA';
    case 'other':
      return 'Other';
    default:
      return fileType;
  }
}

/** Format file size */
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

// ============================================
// Export Functions
// ============================================

/** Escape a value for CSV (handle commas, quotes, newlines) */
function escapeCsvValue(value: string | number | boolean | null | undefined): string {
  if (value === null || value === undefined) return '';
  const str = String(value);
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

/** Generate CSV export content from conflicts */
function generateConflictsCsv(conflicts: Conflict[], stats: ConflictStats): string {
  const headers = [
    'Path',
    'Severity',
    'Score',
    'File Type',
    'Type',
    'Is Identical',
    'Winner Mod',
    'Winner Mod ID',
    'Loser Mods',
    'Loser Mod IDs',
    'Matched Rules',
    'Message',
  ];

  const rows = conflicts.map((c) => [
    escapeCsvValue(c.path),
    escapeCsvValue(c.severity),
    escapeCsvValue(c.score),
    escapeCsvValue(c.fileType),
    escapeCsvValue(c.type),
    escapeCsvValue(c.isIdentical),
    escapeCsvValue(c.winner?.modName ?? ''),
    escapeCsvValue(c.winner?.modId ?? ''),
    escapeCsvValue(c.losers.map((l) => l.modName).join('; ')),
    escapeCsvValue(c.losers.map((l) => l.modId).join('; ')),
    escapeCsvValue(c.matchedRules?.join('; ') ?? ''),
    escapeCsvValue(c.message),
  ]);

  // Add summary header
  const summary = [
    `# Conflict Report Summary`,
    `# Total Conflicts: ${stats.totalConflicts}`,
    `# Critical: ${stats.criticalCount}, High: ${stats.highCount}, Medium: ${stats.mediumCount}, Low: ${stats.lowCount}`,
    `# Generated: ${new Date().toISOString()}`,
    '',
  ];

  return summary.join('\n') + headers.join(',') + '\n' + rows.map((r) => r.join(',')).join('\n');
}

/** Generate JSON export content from conflict analysis */
function generateConflictsJson(
  conflicts: Conflict[],
  stats: ConflictStats,
  modSummaries: ModConflictSummary[],
  collectionSlug: string
): object {
  return {
    version: 1,
    exportedAt: new Date().toISOString(),
    collection: collectionSlug,
    summary: {
      totalConflicts: stats.totalConflicts,
      criticalCount: stats.criticalCount,
      highCount: stats.highCount,
      mediumCount: stats.mediumCount,
      lowCount: stats.lowCount,
      infoCount: stats.infoCount,
      identicalConflicts: stats.identicalConflicts,
      modsWithConflicts: stats.modsWithConflicts,
    },
    modSummaries: modSummaries.map((m) => ({
      modId: m.modId,
      modName: m.modName,
      totalConflicts: m.totalConflicts,
      winCount: m.winCount,
      loseCount: m.loseCount,
      criticalCount: m.criticalCount,
    })),
    conflicts: conflicts.map((c) => ({
      path: c.path,
      severity: c.severity,
      score: c.score,
      fileType: c.fileType,
      type: c.type,
      isIdentical: c.isIdentical,
      winner: c.winner
        ? { modId: c.winner.modId, modName: c.winner.modName, size: c.winner.size }
        : null,
      losers: c.losers.map((l) => ({ modId: l.modId, modName: l.modName, size: l.size })),
      matchedRules: c.matchedRules ?? [],
      message: c.message,
    })),
  };
}

/** Download text content as a file */
function downloadText(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

// ============================================
// Export Toolbar Component
// ============================================

interface ExportToolbarProps {
  conflicts: Conflict[];
  stats: ConflictStats;
  modSummaries: ModConflictSummary[];
  collectionSlug: string;
}

const ExportToolbar: React.FC<ExportToolbarProps> = ({
  conflicts,
  stats,
  modSummaries,
  collectionSlug,
}) => {
  const timestamp = new Date().toISOString().split('T')[0];
  const baseFilename = `conflicts-${collectionSlug}-${timestamp}`;

  const handleExportCsv = () => {
    const csv = generateConflictsCsv(conflicts, stats);
    downloadText(csv, `${baseFilename}.csv`, 'text/csv;charset=utf-8');
  };

  const handleExportJson = () => {
    const json = generateConflictsJson(conflicts, stats, modSummaries, collectionSlug);
    downloadText(JSON.stringify(json, null, 2), `${baseFilename}.json`, 'application/json');
  };

  return (
    <div className="flex gap-2">
      <button
        onClick={handleExportCsv}
        className="min-h-9 px-3 py-1.5 rounded-sm text-sm
          bg-bg-secondary border border-border
          text-text-secondary
          hover:bg-bg-hover hover:text-text-primary
          focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
          transition-colors motion-reduce:transition-none"
        aria-label="Export conflicts to CSV file"
      >
        Export CSV
      </button>
      <button
        onClick={handleExportJson}
        className="min-h-9 px-3 py-1.5 rounded-sm text-sm
          bg-bg-secondary border border-border
          text-text-secondary
          hover:bg-bg-hover hover:text-text-primary
          focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
          transition-colors motion-reduce:transition-none"
        aria-label="Export conflicts to JSON file"
      >
        Export JSON
      </button>
    </div>
  );
};

// ============================================
// Loading Skeleton
// ============================================

const ConflictSkeleton: React.FC = () => (
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
    {/* Filters skeleton */}
    <div className="p-4 rounded-sm bg-bg-card border border-border">
      <div className="flex gap-4 flex-wrap">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-10 w-32 bg-bg-secondary rounded-xs" />
        ))}
      </div>
    </div>
    {/* List skeleton */}
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div className="lg:col-span-2 p-4 rounded-sm bg-bg-card border border-border space-y-2">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div key={i} className="h-16 bg-bg-secondary rounded-xs" />
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
  let message = 'An unexpected error occurred while analyzing conflicts.';

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
    <p className="text-text-secondary text-lg mb-2">No Conflicts Found</p>
    <p className="text-text-muted">
      This collection does not have any file conflicts between mods.
    </p>
  </div>
);

// ============================================
// No Results State
// ============================================

const NoResultsState: React.FC<{ onClear: () => void }> = ({ onClear }) => (
  <div className="p-6 rounded-sm bg-bg-card border border-border text-center">
    <p className="text-text-secondary text-lg mb-2">No Matching Conflicts</p>
    <p className="text-text-muted mb-4">
      No conflicts match your current filters.
    </p>
    <button
      onClick={onClear}
      className="min-h-11 px-6 py-2 rounded-sm
        bg-accent text-white font-medium
        hover:bg-accent/80
        focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
        transition-colors motion-reduce:transition-none"
    >
      Clear Filters
    </button>
  </div>
);

// ============================================
// View Mode Toggle Component
// ============================================

interface ViewModeToggleProps {
  viewMode: ViewMode;
  onChange: (mode: ViewMode) => void;
}

const ViewModeToggle: React.FC<ViewModeToggleProps> = ({ viewMode, onChange }) => {
  const modes: { value: ViewMode; label: string; icon: string }[] = [
    { value: 'list', label: 'List', icon: '📋' },
    { value: 'graph', label: 'Graph', icon: '🕸️' },
  ];

  return (
    <div
      className="flex rounded-sm overflow-hidden border border-border"
      role="radiogroup"
      aria-label="View mode"
    >
      {modes.map((mode) => (
        <button
          key={mode.value}
          onClick={() => onChange(mode.value)}
          className={`
            min-h-9 px-4 py-2 text-sm font-medium flex items-center gap-1.5
            transition-colors motion-reduce:transition-none
            focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-[-2px]
            ${viewMode === mode.value
              ? 'bg-accent text-white'
              : 'bg-bg-secondary text-text-secondary hover:bg-bg-hover'
            }
          `}
          role="radio"
          aria-checked={viewMode === mode.value}
        >
          <span aria-hidden="true">{mode.icon}</span>
          {mode.label}
        </button>
      ))}
    </div>
  );
};

// ============================================
// Stats Header Component
// ============================================

const ConflictHeader: React.FC<ConflictHeaderProps> = ({ stats, cached, exportToolbar, viewMode, onViewModeChange }) => (
  <header className="p-6 rounded-sm bg-bg-card border border-border">
    <div className="flex items-start justify-between gap-4 mb-4 flex-wrap">
      <h2 className="text-xl font-bold text-text-primary">Conflict Analysis</h2>
      <div className="flex items-center gap-3 flex-wrap">
        <ViewModeToggle viewMode={viewMode} onChange={onViewModeChange} />
        {exportToolbar}
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
      <StatItem label="Total Conflicts" value={stats.totalConflicts} />
      <StatItem
        label="Critical"
        value={stats.criticalCount}
        variant={stats.criticalCount > 0 ? 'error' : 'default'}
      />
      <StatItem
        label="High"
        value={stats.highCount}
        variant={stats.highCount > 0 ? 'warning' : 'default'}
      />
      <StatItem label="Medium" value={stats.mediumCount} variant="accent" />
      <StatItem label="Low" value={stats.lowCount} />
      <StatItem label="Identical" value={stats.identicalConflicts} variant="info" />
    </div>
  </header>
);

interface StatItemProps {
  label: string;
  value: number;
  variant?: 'default' | 'accent' | 'warning' | 'error' | 'info';
}

const StatItem: React.FC<StatItemProps> = ({ label, value, variant = 'default' }) => {
  const valueClass = {
    default: 'text-text-primary',
    accent: 'text-accent',
    warning: 'text-warning',
    error: 'text-error',
    info: 'text-accent/70',
  }[variant];

  return (
    <div className="space-y-1">
      <p className="text-sm text-text-muted">{label}</p>
      <p className={`text-2xl font-bold ${valueClass}`}>{value}</p>
    </div>
  );
};

// ============================================
// Filters Component
// ============================================

const ConflictFilters: React.FC<ConflictFiltersProps> = ({
  selectedSeverity,
  selectedFileType,
  selectedMod,
  searchQuery,
  onSeverityChange,
  onFileTypeChange,
  onModChange,
  onSearchChange,
  modSummaries,
  stats,
}) => {
  const severityOptions: ConflictSeverity[] = ['critical', 'high', 'medium', 'low', 'info'];
  const fileTypeOptions: FileType[] = ['plugin', 'bsa', 'script', 'mesh', 'texture', 'interface', 'sound', 'seq', 'other'];

  // Filter file types to only show those with conflicts
  const availableFileTypes = fileTypeOptions.filter(
    (ft) => stats.byFileType && stats.byFileType[ft] > 0
  );

  return (
    <div className="p-4 rounded-sm bg-bg-card border border-border">
      <div className="flex gap-4 flex-wrap items-end">
        {/* Search */}
        <div className="flex-1 min-w-[200px]">
          <label
            htmlFor="conflict-search"
            className="block text-sm text-text-muted mb-1"
          >
            Search by path
          </label>
          <input
            id="conflict-search"
            type="search"
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="e.g., textures/sky"
            className="w-full min-h-11 px-4 py-2 rounded-sm
              bg-bg-secondary border border-border
              text-text-primary placeholder:text-text-muted
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none"
          />
        </div>

        {/* Severity filter */}
        <div>
          <label
            htmlFor="severity-filter"
            className="block text-sm text-text-muted mb-1"
          >
            Severity
          </label>
          <select
            id="severity-filter"
            value={selectedSeverity ?? ''}
            onChange={(e) =>
              onSeverityChange(e.target.value ? (e.target.value as ConflictSeverity) : null)
            }
            className="min-h-11 px-4 py-2 rounded-sm
              bg-bg-secondary border border-border
              text-text-primary
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none"
          >
            <option value="">All</option>
            {severityOptions.map((s) => (
              <option key={s} value={s}>
                {getSeverityLabel(s)}
              </option>
            ))}
          </select>
        </div>

        {/* File type filter */}
        <div>
          <label
            htmlFor="filetype-filter"
            className="block text-sm text-text-muted mb-1"
          >
            File Type
          </label>
          <select
            id="filetype-filter"
            value={selectedFileType ?? ''}
            onChange={(e) =>
              onFileTypeChange(e.target.value ? (e.target.value as FileType) : null)
            }
            className="min-h-11 px-4 py-2 rounded-sm
              bg-bg-secondary border border-border
              text-text-primary
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none"
          >
            <option value="">All</option>
            {availableFileTypes.map((ft) => (
              <option key={ft} value={ft}>
                {getFileTypeLabel(ft)}
              </option>
            ))}
          </select>
        </div>

        {/* Mod filter */}
        <div>
          <label
            htmlFor="mod-filter"
            className="block text-sm text-text-muted mb-1"
          >
            Mod
          </label>
          <select
            id="mod-filter"
            value={selectedMod ?? ''}
            onChange={(e) => onModChange(e.target.value || null)}
            className="min-h-11 px-4 py-2 rounded-sm
              bg-bg-secondary border border-border
              text-text-primary
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none
              max-w-[200px] truncate"
          >
            <option value="">All Mods</option>
            {modSummaries.map((mod) => (
              <option key={mod.modId} value={mod.modId}>
                {mod.modName} ({mod.totalConflicts})
              </option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
};

// ============================================
// Conflict Row Component
// ============================================

const ConflictRow: React.FC<ConflictRowProps> = ({ conflict, isSelected, onSelect }) => (
  <button
    onClick={onSelect}
    className={`
      w-full min-h-11 px-4 py-3 rounded-sm text-left
      flex flex-col gap-1
      transition-colors motion-reduce:transition-none
      focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
      ${isSelected
        ? 'bg-accent/10 border border-accent'
        : 'bg-bg-secondary border border-transparent hover:bg-bg-secondary/80'
      }
    `}
    aria-pressed={isSelected}
  >
    <div className="flex items-center gap-2 w-full">
      <span
        className={`px-2 py-0.5 rounded-full text-xs font-medium ${getSeverityBadgeClass(conflict.severity)}`}
      >
        {getSeverityLabel(conflict.severity)}
      </span>
      <span
        className={`px-2 py-0.5 rounded-full text-xs font-medium ${getFileTypeBadgeClass(conflict.fileType)}`}
      >
        {getFileTypeLabel(conflict.fileType)}
      </span>
      {conflict.isIdentical && (
        <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-accent/10 text-accent">
          Identical
        </span>
      )}
      <span className="text-xs text-text-muted ml-auto">
        Score: {conflict.score}
      </span>
    </div>
    <span className="font-mono text-sm text-text-primary truncate w-full">
      {conflict.path}
    </span>
    <span className="text-xs text-text-muted">
      {conflict.sources.length} mod{conflict.sources.length !== 1 ? 's' : ''} &middot; Winner:{' '}
      <span className="text-text-secondary">{conflict.winner?.modName ?? 'None'}</span>
    </span>
  </button>
);

// ============================================
// Conflict List Component
// ============================================

const ConflictList: React.FC<ConflictListProps> = ({
  conflicts,
  selectedConflict,
  onSelectConflict,
}) => (
  <section
    aria-label="Conflict list"
    className="p-4 rounded-sm bg-bg-card border border-border"
  >
    <h3 className="text-lg font-semibold text-text-primary mb-4">
      Conflicts ({conflicts.length})
    </h3>
    <div className="space-y-2 max-h-[600px] overflow-y-auto">
      {conflicts.map((conflict) => (
        <ConflictRow
          key={conflict.path}
          conflict={conflict}
          isSelected={selectedConflict?.path === conflict.path}
          onSelect={() =>
            onSelectConflict(
              selectedConflict?.path === conflict.path ? null : conflict
            )
          }
        />
      ))}
    </div>
  </section>
);

// ============================================
// Conflict Details Component
// ============================================

const ConflictDetails: React.FC<ConflictDetailsProps> = ({ conflict }) => (
  <section
    aria-label={`Details for ${conflict.path}`}
    className="p-4 rounded-sm bg-bg-card border border-border"
  >
    <h3 className="text-lg font-semibold text-text-primary mb-4">
      Conflict Details
    </h3>
    <div className="space-y-4">
      {/* Path */}
      <div>
        <p className="text-sm text-text-muted mb-1">Path</p>
        <p className="font-mono text-sm text-text-primary break-all">
          {conflict.path}
        </p>
      </div>

      {/* Type and Severity */}
      <div className="flex gap-4 flex-wrap">
        <div>
          <p className="text-sm text-text-muted mb-1">Severity</p>
          <span
            className={`px-2 py-1 rounded-full text-sm font-medium ${getSeverityBadgeClass(conflict.severity)}`}
          >
            {getSeverityLabel(conflict.severity)}
          </span>
        </div>
        <div>
          <p className="text-sm text-text-muted mb-1">File Type</p>
          <span
            className={`px-2 py-1 rounded-full text-sm font-medium ${getFileTypeBadgeClass(conflict.fileType)}`}
          >
            {getFileTypeLabel(conflict.fileType)}
          </span>
        </div>
        <div>
          <p className="text-sm text-text-muted mb-1">Score</p>
          <span className="text-text-primary font-medium">{conflict.score}</span>
        </div>
        {conflict.isIdentical && (
          <div>
            <p className="text-sm text-text-muted mb-1">Status</p>
            <span className="px-2 py-1 rounded-full text-sm font-medium bg-accent/10 text-accent">
              Identical Files
            </span>
          </div>
        )}
      </div>

      {/* Message */}
      <div>
        <p className="text-sm text-text-muted mb-1">Description</p>
        <p className="text-text-secondary text-sm">{conflict.message}</p>
      </div>

      {/* Winner */}
      {conflict.winner && (
        <div>
          <p className="text-sm text-text-muted mb-1">Winner (loads last)</p>
          <div className="p-3 rounded-xs bg-accent/10 border border-accent/30">
            <p className="font-medium text-text-primary">{conflict.winner.modName}</p>
            <p className="text-xs text-text-muted font-mono mt-1">
              {formatFileSize(conflict.winner.size)}
            </p>
          </div>
        </div>
      )}

      {/* Losers */}
      {conflict.losers.length > 0 && (
        <div>
          <p className="text-sm text-text-muted mb-1">
            Overwritten ({conflict.losers.length})
          </p>
          <ul className="space-y-2">
            {conflict.losers.map((loser) => (
              <li
                key={loser.modId}
                className="p-3 rounded-xs bg-bg-secondary"
              >
                <p className="font-medium text-text-primary">{loser.modName}</p>
                <p className="text-xs text-text-muted font-mono mt-1">
                  {formatFileSize(loser.size)}
                </p>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Matched Rules */}
      {conflict.matchedRules && conflict.matchedRules.length > 0 && (
        <div>
          <p className="text-sm text-text-muted mb-1">Matched Rules</p>
          <div className="flex flex-wrap gap-2">
            {conflict.matchedRules.map((rule) => (
              <span
                key={rule}
                className="px-2 py-1 rounded-full text-xs bg-warning/20 text-warning"
              >
                {rule}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  </section>
);

// ============================================
// Summary Panel Component
// ============================================

interface SummaryPanelProps {
  modSummaries: ModConflictSummary[];
  onSelectMod: (modId: string) => void;
}

const SummaryPanel: React.FC<SummaryPanelProps> = ({ modSummaries, onSelectMod }) => {
  // Sort by total conflicts descending
  const sortedMods = useMemo(
    () => [...modSummaries].sort((a, b) => b.totalConflicts - a.totalConflicts),
    [modSummaries]
  );

  return (
    <section
      aria-label="Mod conflict summary"
      className="p-4 rounded-sm bg-bg-card border border-border"
    >
      <h3 className="text-lg font-semibold text-text-primary mb-4">
        Mods with Conflicts ({sortedMods.length})
      </h3>
      <div className="space-y-2 max-h-[400px] overflow-y-auto">
        {sortedMods.map((mod) => (
          <button
            key={mod.modId}
            onClick={() => onSelectMod(mod.modId)}
            className="w-full p-3 rounded-xs text-left
              bg-bg-secondary
              hover:bg-bg-hover
              transition-colors motion-reduce:transition-none
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2"
          >
            <p className="font-medium text-text-primary truncate">
              {mod.modName}
            </p>
            <div className="flex gap-3 mt-1 text-xs">
              <span className="text-text-muted">
                {mod.totalConflicts} conflict{mod.totalConflicts !== 1 ? 's' : ''}
              </span>
              <span className="text-accent">
                {mod.winCount} win{mod.winCount !== 1 ? 's' : ''}
              </span>
              <span className="text-warning">
                {mod.loseCount} lose{mod.loseCount !== 1 ? 's' : ''}
              </span>
              {mod.criticalCount > 0 && (
                <span className="text-error">
                  {mod.criticalCount} critical
                </span>
              )}
            </div>
          </button>
        ))}
      </div>
    </section>
  );
};

// ============================================
// Main ConflictView Component
// ============================================

/** Container component for conflict analysis view */
export const ConflictView: React.FC<ConflictViewProps> = ({ slug, revision }) => {
  const { data, isLoading, error, refetch } = useConflictAnalysis(slug, revision);
  const [selectedConflict, setSelectedConflict] = useState<Conflict | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>('list');

  // Filter state
  const [selectedSeverity, setSelectedSeverity] = useState<ConflictSeverity | null>(null);
  const [selectedFileType, setSelectedFileType] = useState<FileType | null>(null);
  const [selectedMod, setSelectedMod] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  // Filter conflicts based on current filters
  const conflicts = data?.conflicts;
  const filteredConflicts = useMemo(() => {
    if (!conflicts) return [];

    return conflicts.filter((conflict) => {
      // Severity filter
      if (selectedSeverity && conflict.severity !== selectedSeverity) {
        return false;
      }

      // File type filter
      if (selectedFileType && conflict.fileType !== selectedFileType) {
        return false;
      }

      // Mod filter
      if (selectedMod) {
        const modInvolved = conflict.sources.some((s) => s.modId === selectedMod);
        if (!modInvolved) {
          return false;
        }
      }

      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        if (!conflict.path.toLowerCase().includes(query)) {
          return false;
        }
      }

      return true;
    });
  }, [conflicts, selectedSeverity, selectedFileType, selectedMod, searchQuery]);

  // Clear all filters
  const clearFilters = () => {
    setSelectedSeverity(null);
    setSelectedFileType(null);
    setSelectedMod(null);
    setSearchQuery('');
  };

  // Handle selecting a mod from the summary panel
  const handleSelectMod = (modId: string) => {
    setSelectedMod(modId);
    setSelectedConflict(null);
  };

  // Check if any filters are active
  const hasActiveFilters = selectedSeverity || selectedFileType || selectedMod || searchQuery;

  // Loading state
  if (isLoading) {
    return <ConflictSkeleton />;
  }

  // Error state
  if (error) {
    return <ErrorDisplay error={error} onRetry={() => refetch()} />;
  }

  // No data state
  if (!data) {
    return <EmptyState />;
  }

  // No conflicts state
  if (data.conflicts.length === 0) {
    return (
      <div className="space-y-6">
        <div aria-live="polite" className="sr-only">
          Conflict analysis complete. No conflicts found.
        </div>
        <ConflictHeader
          stats={data.stats}
          cached={data.cached}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
        />
        <EmptyState />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Announce data loaded to screen readers */}
      <div aria-live="polite" className="sr-only">
        Conflict analysis complete. {data.stats.totalConflicts} conflicts found.
      </div>

      {/* Stats header with export */}
      <ConflictHeader
        stats={data.stats}
        cached={data.cached}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        exportToolbar={
          <ExportToolbar
            conflicts={data.conflicts}
            stats={data.stats}
            modSummaries={data.modSummaries}
            collectionSlug={slug}
          />
        }
      />

      {/* View mode: Graph */}
      {viewMode === 'graph' && (
        <ConflictGraphView
          conflicts={data.conflicts}
          modSummaries={data.modSummaries}
          onSelectMod={handleSelectMod}
          selectedModId={selectedMod}
        />
      )}

      {/* View mode: List */}
      {viewMode === 'list' && (
        <>
          {/* Filters */}
          <ConflictFilters
            selectedSeverity={selectedSeverity}
            selectedFileType={selectedFileType}
            selectedMod={selectedMod}
            searchQuery={searchQuery}
            onSeverityChange={setSelectedSeverity}
            onFileTypeChange={setSelectedFileType}
            onModChange={setSelectedMod}
            onSearchChange={setSearchQuery}
            modSummaries={data.modSummaries}
            stats={data.stats}
          />

          {/* No results after filtering */}
          {filteredConflicts.length === 0 && hasActiveFilters && (
            <NoResultsState onClear={clearFilters} />
          )}

          {/* Main content grid */}
          {filteredConflicts.length > 0 && (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Conflict list */}
              <div className="lg:col-span-2">
                <ConflictList
                  conflicts={filteredConflicts}
                  selectedConflict={selectedConflict}
                  onSelectConflict={setSelectedConflict}
                />
              </div>

              {/* Sidebar */}
              <div className="space-y-6">
                {/* Conflict details or summary panel */}
                {selectedConflict ? (
                  <ConflictDetails conflict={selectedConflict} />
                ) : (
                  <SummaryPanel
                    modSummaries={data.modSummaries}
                    onSelectMod={handleSelectMod}
                  />
                )}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};
