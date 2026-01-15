import { useId } from 'react';

// ============================================
// Types
// ============================================

interface SkeletonProps {
  /** Width of the skeleton (CSS value) */
  width?: string | number;
  /** Height of the skeleton (CSS value) */
  height?: string | number;
  /** Border radius variant */
  variant?: 'rectangular' | 'circular' | 'text' | 'rounded';
  /** Additional CSS classes */
  className?: string;
  /** Accessible label for screen readers */
  ariaLabel?: string;
}

interface SkeletonTextProps {
  /** Number of text lines to show */
  lines?: number;
  /** Width pattern for lines (in percentage) */
  lineWidths?: number[];
  /** Line height (CSS value) */
  lineHeight?: string;
  /** Gap between lines */
  gap?: string;
  /** Accessible label for screen readers */
  ariaLabel?: string;
}

interface SkeletonCardProps {
  /** Whether to show a header line */
  showHeader?: boolean;
  /** Number of content lines */
  contentLines?: number;
  /** Whether to show an action area */
  showActions?: boolean;
  /** Accessible label for screen readers */
  ariaLabel?: string;
}

// ============================================
// Utility Functions
// ============================================

function formatDimension(value: string | number | undefined): string | undefined {
  if (value === undefined) return undefined;
  if (typeof value === 'number') return `${value}px`;
  return value;
}

function getVariantClass(variant: SkeletonProps['variant']): string {
  switch (variant) {
    case 'circular':
      return 'rounded-full';
    case 'text':
      return 'rounded-xs';
    case 'rounded':
      return 'rounded-sm';
    case 'rectangular':
    default:
      return 'rounded-none';
  }
}

// ============================================
// Base Skeleton Component
// ============================================

/**
 * Base skeleton loading placeholder component.
 * Use this for custom skeleton layouts.
 */
export const Skeleton: React.FC<SkeletonProps> = ({
  width,
  height = '1rem',
  variant = 'text',
  className = '',
  ariaLabel,
}) => {
  const labelId = useId();

  return (
    <div
      className={`
        animate-pulse bg-bg-secondary
        ${getVariantClass(variant)}
        ${className}
      `}
      style={{
        width: formatDimension(width),
        height: formatDimension(height),
      }}
      role="status"
      aria-busy="true"
      aria-label={ariaLabel}
      aria-labelledby={ariaLabel ? undefined : labelId}
    >
      <span id={labelId} className="sr-only">
        {ariaLabel ?? 'Loading...'}
      </span>
    </div>
  );
};

// ============================================
// SkeletonText Component
// ============================================

/**
 * Multi-line text skeleton for paragraphs and descriptions.
 */
export const SkeletonText: React.FC<SkeletonTextProps> = ({
  lines = 3,
  lineWidths = [100, 100, 66],
  lineHeight = '1rem',
  gap = '0.5rem',
  ariaLabel = 'Loading text content',
}) => {
  return (
    <div
      className="space-y-2"
      style={{ gap }}
      role="status"
      aria-busy="true"
      aria-label={ariaLabel}
    >
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          height={lineHeight}
          width={`${lineWidths[i % lineWidths.length]}%`}
          variant="text"
        />
      ))}
      <span className="sr-only">{ariaLabel}</span>
    </div>
  );
};

// ============================================
// SkeletonCard Component
// ============================================

/**
 * Card skeleton with header, content, and optional actions.
 */
export const SkeletonCard: React.FC<SkeletonCardProps> = ({
  showHeader = true,
  contentLines = 3,
  showActions = false,
  ariaLabel = 'Loading card content',
}) => {
  return (
    <div
      className="p-4 rounded-sm bg-bg-card border border-border animate-pulse"
      role="status"
      aria-busy="true"
      aria-label={ariaLabel}
    >
      {showHeader && (
        <div className="mb-4">
          <Skeleton height="1.5rem" width="50%" variant="text" />
        </div>
      )}

      <div className="space-y-2">
        {Array.from({ length: contentLines }).map((_, i) => (
          <Skeleton
            key={i}
            height="1rem"
            width={i === contentLines - 1 ? '66%' : '100%'}
            variant="text"
          />
        ))}
      </div>

      {showActions && (
        <div className="mt-4 flex gap-2">
          <Skeleton height="2.5rem" width="5rem" variant="rounded" />
          <Skeleton height="2.5rem" width="5rem" variant="rounded" />
        </div>
      )}

      <span className="sr-only">{ariaLabel}</span>
    </div>
  );
};

// ============================================
// SkeletonList Component
// ============================================

interface SkeletonListProps {
  /** Number of list items */
  itemCount?: number;
  /** Height of each item */
  itemHeight?: string;
  /** Whether to show icons/avatars */
  showIcon?: boolean;
  /** Gap between items */
  gap?: string;
  /** Accessible label for screen readers */
  ariaLabel?: string;
}

/**
 * List skeleton for lists and tables.
 */
export const SkeletonList: React.FC<SkeletonListProps> = ({
  itemCount = 5,
  itemHeight = '3rem',
  showIcon = true,
  gap = '0.5rem',
  ariaLabel = 'Loading list content',
}) => {
  return (
    <div
      className="space-y-2"
      style={{ gap }}
      role="status"
      aria-busy="true"
      aria-label={ariaLabel}
    >
      {Array.from({ length: itemCount }).map((_, i) => (
        <div
          key={i}
          className="flex items-center gap-3 p-3 rounded-xs bg-bg-secondary animate-pulse"
          style={{ height: itemHeight }}
        >
          {showIcon && (
            <Skeleton
              height="2rem"
              width="2rem"
              variant="rounded"
            />
          )}
          <div className="flex-1 space-y-2">
            <Skeleton height="0.75rem" width="40%" variant="text" />
            <Skeleton height="0.5rem" width="60%" variant="text" />
          </div>
        </div>
      ))}
      <span className="sr-only">{ariaLabel}</span>
    </div>
  );
};

// ============================================
// SkeletonGrid Component
// ============================================

interface SkeletonGridProps {
  /** Number of grid items */
  itemCount?: number;
  /** Number of columns (responsive) */
  columns?: 1 | 2 | 3 | 4;
  /** Height of each grid item */
  itemHeight?: string;
  /** Accessible label for screen readers */
  ariaLabel?: string;
}

/**
 * Grid skeleton for card grids and galleries.
 */
export const SkeletonGrid: React.FC<SkeletonGridProps> = ({
  itemCount = 6,
  columns = 3,
  itemHeight = '8rem',
  ariaLabel = 'Loading grid content',
}) => {
  const columnClasses = {
    1: 'grid-cols-1',
    2: 'grid-cols-1 sm:grid-cols-2',
    3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
  };

  return (
    <div
      className={`grid gap-4 ${columnClasses[columns]}`}
      role="status"
      aria-busy="true"
      aria-label={ariaLabel}
    >
      {Array.from({ length: itemCount }).map((_, i) => (
        <div
          key={i}
          className="rounded-sm bg-bg-secondary animate-pulse"
          style={{ height: itemHeight }}
        />
      ))}
      <span className="sr-only">{ariaLabel}</span>
    </div>
  );
};

// ============================================
// SkeletonStats Component
// ============================================

interface SkeletonStatsProps {
  /** Number of stat items */
  statCount?: number;
  /** Accessible label for screen readers */
  ariaLabel?: string;
}

/**
 * Stats skeleton for header statistics.
 */
export const SkeletonStats: React.FC<SkeletonStatsProps> = ({
  statCount = 4,
  ariaLabel = 'Loading statistics',
}) => {
  return (
    <div
      className="grid grid-cols-2 md:grid-cols-4 gap-4 p-6 rounded-sm bg-bg-card border border-border animate-pulse"
      role="status"
      aria-busy="true"
      aria-label={ariaLabel}
    >
      {Array.from({ length: statCount }).map((_, i) => (
        <div key={i} className="space-y-2">
          <Skeleton height="0.875rem" width="5rem" variant="text" />
          <Skeleton height="2rem" width="4rem" variant="text" />
        </div>
      ))}
      <span className="sr-only">{ariaLabel}</span>
    </div>
  );
};
