import { useEffect, useId, useRef } from 'react';

import type { ShortcutDefinition } from '@hooks/useKeyboardShortcuts.ts';

// ============================================
// Types
// ============================================

interface KeyboardShortcutsHelpProps {
  /** Whether the help overlay is visible */
  isOpen: boolean;
  /** Close the help overlay */
  onClose: () => void;
  /** List of shortcuts to display */
  shortcuts: ShortcutDefinition[];
  /** Currently pending key sequence */
  pendingKey: string | null;
}

// ============================================
// Helper Functions
// ============================================

/** Group shortcuts by category */
function groupByCategory(
  shortcuts: ShortcutDefinition[]
): Record<string, ShortcutDefinition[]> {
  const groups: Record<string, ShortcutDefinition[]> = {
    navigation: [],
    actions: [],
    views: [],
  };

  for (const shortcut of shortcuts) {
    if (groups[shortcut.category]) {
      groups[shortcut.category].push(shortcut);
    }
  }

  return groups;
}

/** Get human-readable category name */
function getCategoryLabel(category: string): string {
  switch (category) {
    case 'navigation':
      return 'Navigation';
    case 'actions':
      return 'Actions';
    case 'views':
      return 'View Switching';
    default:
      return category;
  }
}

/** Format key for display */
function formatKey(key: string): string[] {
  return key.split(' ').map((k) => {
    switch (k) {
      case '/':
        return '/';
      case '?':
        return '?';
      default:
        return k.toUpperCase();
    }
  });
}

// ============================================
// Key Badge Component
// ============================================

const KeyBadge: React.FC<{ keyChar: string }> = ({ keyChar }) => (
  <kbd
    className="inline-flex items-center justify-center min-w-[1.5rem] h-6 px-1.5
      rounded-xs bg-bg-secondary border border-border
      font-mono text-xs text-text-primary"
  >
    {keyChar}
  </kbd>
);

// ============================================
// Main Component
// ============================================

/**
 * Keyboard shortcuts help overlay.
 * Shows all available keyboard shortcuts grouped by category.
 */
export const KeyboardShortcutsHelp: React.FC<KeyboardShortcutsHelpProps> = ({
  isOpen,
  onClose,
  shortcuts,
  pendingKey,
}) => {
  const titleId = useId();
  const closeButtonRef = useRef<HTMLButtonElement>(null);

  // Focus trap and close on Escape
  useEffect(() => {
    if (!isOpen) return;

    // Focus close button when opened
    closeButtonRef.current?.focus();

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  // Prevent body scroll when open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  const groups = groupByCategory(shortcuts);

  // Always show the help shortcut
  const builtInShortcuts: ShortcutDefinition[] = [
    {
      key: '?',
      description: 'Show/hide this help',
      category: 'actions',
      handler: () => {},
    },
    {
      key: 'esc',
      description: 'Close dialogs / Clear',
      category: 'actions',
      handler: () => {},
    },
  ];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Dialog */}
      <div
        className="relative w-full max-w-xl max-h-[80vh] overflow-y-auto
          rounded-sm bg-bg-card border border-border shadow-xl"
      >
        {/* Header */}
        <div className="sticky top-0 flex items-center justify-between p-4 border-b border-border bg-bg-card z-10">
          <h2 id={titleId} className="text-lg font-semibold text-text-primary flex items-center gap-2">
            <span aria-hidden="true">⌨️</span>
            Keyboard Shortcuts
          </h2>
          <button
            ref={closeButtonRef}
            onClick={onClose}
            className="min-h-9 min-w-9 flex items-center justify-center rounded-sm
              text-text-muted hover:text-text-primary hover:bg-bg-secondary
              focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
              transition-colors motion-reduce:transition-none"
            aria-label="Close help"
          >
            <span aria-hidden="true" className="text-xl">×</span>
          </button>
        </div>

        {/* Pending key indicator */}
        {pendingKey && (
          <div className="p-3 bg-accent/10 border-b border-accent/30">
            <p className="text-sm text-accent flex items-center gap-2">
              <span>Waiting for next key:</span>
              {formatKey(pendingKey).map((k, i) => (
                <KeyBadge key={i} keyChar={k} />
              ))}
            </p>
          </div>
        )}

        {/* Shortcuts list */}
        <div className="p-4 space-y-6">
          {/* Built-in shortcuts */}
          <section>
            <h3 className="text-sm font-semibold text-text-muted mb-3 uppercase tracking-wider">
              General
            </h3>
            <div className="space-y-2">
              {builtInShortcuts.map((shortcut) => (
                <ShortcutRow key={shortcut.key} shortcut={shortcut} />
              ))}
            </div>
          </section>

          {/* Grouped shortcuts */}
          {Object.entries(groups).map(([category, categoryShortcuts]) => {
            if (categoryShortcuts.length === 0) return null;

            return (
              <section key={category}>
                <h3 className="text-sm font-semibold text-text-muted mb-3 uppercase tracking-wider">
                  {getCategoryLabel(category)}
                </h3>
                <div className="space-y-2">
                  {categoryShortcuts.map((shortcut) => (
                    <ShortcutRow key={shortcut.key} shortcut={shortcut} />
                  ))}
                </div>
              </section>
            );
          })}
        </div>

        {/* Footer */}
        <div className="sticky bottom-0 p-3 border-t border-border bg-bg-secondary/50 text-center">
          <p className="text-xs text-text-muted">
            Press <KeyBadge keyChar="?" /> anytime to show this help
          </p>
        </div>
      </div>
    </div>
  );
};

// ============================================
// Shortcut Row Component
// ============================================

const ShortcutRow: React.FC<{ shortcut: ShortcutDefinition }> = ({ shortcut }) => {
  const keys = formatKey(shortcut.key);

  return (
    <div className="flex items-center justify-between py-1">
      <span className="text-sm text-text-secondary">{shortcut.description}</span>
      <div className="flex items-center gap-1">
        {keys.map((key, index) => (
          <span key={index} className="flex items-center gap-1">
            {index > 0 && <span className="text-text-muted text-xs">then</span>}
            <KeyBadge keyChar={key} />
          </span>
        ))}
      </div>
    </div>
  );
};
