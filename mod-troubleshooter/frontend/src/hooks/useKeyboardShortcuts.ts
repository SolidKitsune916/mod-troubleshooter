import { useEffect, useState, useCallback, useRef } from 'react';

// ============================================
// Types
// ============================================

export interface ShortcutDefinition {
  /** Key or key sequence (e.g., 'g c' for g then c) */
  key: string;
  /** Description for help overlay */
  description: string;
  /** Category for grouping in help */
  category: 'navigation' | 'actions' | 'views';
  /** Callback function */
  handler: () => void;
  /** Whether the shortcut is currently enabled */
  enabled?: boolean;
}

interface UseKeyboardShortcutsOptions {
  /** Whether shortcuts are globally enabled */
  enabled?: boolean;
  /** Shortcuts to register */
  shortcuts: ShortcutDefinition[];
}

interface UseKeyboardShortcutsResult {
  /** Whether the help overlay is visible */
  showHelp: boolean;
  /** Toggle help overlay visibility */
  toggleHelp: () => void;
  /** Close help overlay */
  closeHelp: () => void;
  /** Currently active key sequence (for display) */
  pendingKey: string | null;
  /** All registered shortcuts for help display */
  shortcuts: ShortcutDefinition[];
}

// ============================================
// Constants
// ============================================

/** Elements that should prevent shortcut activation */
const INTERACTIVE_ELEMENTS = ['INPUT', 'TEXTAREA', 'SELECT', 'BUTTON'];

/** Timeout for key sequences (ms) */
const SEQUENCE_TIMEOUT = 1000;

// ============================================
// Hook Implementation
// ============================================

/**
 * Hook for managing global keyboard shortcuts.
 * Supports single keys and key sequences (e.g., 'g c' for navigation).
 */
export function useKeyboardShortcuts(
  options: UseKeyboardShortcutsOptions
): UseKeyboardShortcutsResult {
  const { enabled = true, shortcuts } = options;

  const [showHelp, setShowHelp] = useState(false);
  const [pendingKey, setPendingKey] = useState<string | null>(null);
  const pendingKeyRef = useRef<string | null>(null);
  const timeoutRef = useRef<number | null>(null);

  const toggleHelp = useCallback(() => {
    setShowHelp((prev) => !prev);
  }, []);

  const closeHelp = useCallback(() => {
    setShowHelp(false);
  }, []);

  // Clear pending key sequence
  const clearPendingKey = useCallback(() => {
    pendingKeyRef.current = null;
    setPendingKey(null);
    if (timeoutRef.current) {
      window.clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  }, []);

  // Handle keydown events
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      // Don't handle shortcuts when typing in inputs
      const target = event.target as HTMLElement;
      if (
        INTERACTIVE_ELEMENTS.includes(target.tagName) ||
        target.isContentEditable
      ) {
        // Allow Escape in inputs to blur
        if (event.key === 'Escape') {
          target.blur();
        }
        return;
      }

      // Get the pressed key (lowercase for consistency)
      const key = event.key.toLowerCase();

      // Handle help toggle with '?'
      if (key === '?' || (event.shiftKey && key === '/')) {
        event.preventDefault();
        toggleHelp();
        return;
      }

      // Handle Escape to close help
      if (key === 'escape') {
        if (showHelp) {
          event.preventDefault();
          closeHelp();
          return;
        }
        clearPendingKey();
        return;
      }

      // Build current sequence
      const currentSequence = pendingKeyRef.current
        ? `${pendingKeyRef.current} ${key}`
        : key;

      // Check for matching shortcuts
      const enabledShortcuts = shortcuts.filter((s) => s.enabled !== false);

      // Look for exact match first
      const exactMatch = enabledShortcuts.find(
        (s) => s.key.toLowerCase() === currentSequence
      );

      if (exactMatch) {
        event.preventDefault();
        clearPendingKey();
        exactMatch.handler();
        return;
      }

      // Check if this is a prefix of any sequence
      const isPrefix = enabledShortcuts.some((s) =>
        s.key.toLowerCase().startsWith(currentSequence + ' ')
      );

      if (isPrefix) {
        event.preventDefault();
        pendingKeyRef.current = currentSequence;
        setPendingKey(currentSequence);

        // Clear pending key after timeout
        if (timeoutRef.current) {
          window.clearTimeout(timeoutRef.current);
        }
        timeoutRef.current = window.setTimeout(clearPendingKey, SEQUENCE_TIMEOUT);
        return;
      }

      // No match, clear pending key
      clearPendingKey();
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      if (timeoutRef.current) {
        window.clearTimeout(timeoutRef.current);
      }
    };
  }, [enabled, shortcuts, showHelp, toggleHelp, closeHelp, clearPendingKey]);

  return {
    showHelp,
    toggleHelp,
    closeHelp,
    pendingKey,
    shortcuts,
  };
}

// ============================================
// Default Shortcuts Factory
// ============================================

export interface CreateDefaultShortcutsOptions {
  onGoToCollections?: () => void;
  onGoToSettings?: () => void;
  onFocusSearch?: () => void;
  onToggleHelp?: () => void;
  onSwitchView?: (viewIndex: number) => void;
  viewCount?: number;
}

/**
 * Create default application shortcuts.
 */
export function createDefaultShortcuts(
  options: CreateDefaultShortcutsOptions
): ShortcutDefinition[] {
  const {
    onGoToCollections,
    onGoToSettings,
    onFocusSearch,
    onSwitchView,
    viewCount = 0,
  } = options;

  const shortcuts: ShortcutDefinition[] = [];

  // Navigation shortcuts
  if (onGoToCollections) {
    shortcuts.push({
      key: 'g c',
      description: 'Go to Collections',
      category: 'navigation',
      handler: onGoToCollections,
    });
  }

  if (onGoToSettings) {
    shortcuts.push({
      key: 'g s',
      description: 'Go to Settings',
      category: 'navigation',
      handler: onGoToSettings,
    });
  }

  // Action shortcuts
  if (onFocusSearch) {
    shortcuts.push({
      key: '/',
      description: 'Focus search',
      category: 'actions',
      handler: onFocusSearch,
    });
  }

  // View shortcuts
  if (onSwitchView && viewCount > 0) {
    for (let i = 0; i < Math.min(viewCount, 9); i++) {
      shortcuts.push({
        key: String(i + 1),
        description: `Switch to view ${i + 1}`,
        category: 'views',
        handler: () => onSwitchView(i),
      });
    }
  }

  return shortcuts;
}
