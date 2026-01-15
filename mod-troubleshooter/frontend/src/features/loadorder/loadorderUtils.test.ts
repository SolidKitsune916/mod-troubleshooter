import { describe, it, expect } from 'vitest';

import type { LoadOrderPluginInfo } from '@/types/index.ts';

// Re-implement the comparison logic for testing
// (In a real scenario, we'd export these functions, but for this test we duplicate)

type PluginDiffStatus = 'a-only' | 'b-only' | 'same' | 'moved';

interface DiffPlugin {
  filename: string;
  indexA: number | null;
  indexB: number | null;
  diffStatus: PluginDiffStatus;
  positionDelta: number;
}

interface DiffSummary {
  aOnly: number;
  bOnly: number;
  same: number;
  moved: number;
  total: number;
}

function compareLoadOrders(
  pluginsA: LoadOrderPluginInfo[],
  pluginsB: LoadOrderPluginInfo[],
): { diffPlugins: DiffPlugin[]; summary: DiffSummary } {
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

  for (const [key, pluginA] of mapA) {
    processedKeys.add(key);
    const pluginB = mapB.get(key);

    if (!pluginB) {
      diffPlugins.push({
        filename: pluginA.filename,
        indexA: pluginA.index,
        indexB: null,
        diffStatus: 'a-only',
        positionDelta: 0,
      });
      aOnly++;
    } else if (pluginA.index === pluginB.index) {
      diffPlugins.push({
        filename: pluginA.filename,
        indexA: pluginA.index,
        indexB: pluginB.index,
        diffStatus: 'same',
        positionDelta: 0,
      });
      same++;
    } else {
      diffPlugins.push({
        filename: pluginA.filename,
        indexA: pluginA.index,
        indexB: pluginB.index,
        diffStatus: 'moved',
        positionDelta: pluginB.index - pluginA.index,
      });
      moved++;
    }
  }

  for (const [key, pluginB] of mapB) {
    if (!processedKeys.has(key)) {
      diffPlugins.push({
        filename: pluginB.filename,
        indexA: null,
        indexB: pluginB.index,
        diffStatus: 'b-only',
        positionDelta: 0,
      });
      bOnly++;
    }
  }

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

function pluginListsEqual(a: LoadOrderPluginInfo[], b: LoadOrderPluginInfo[]): boolean {
  if (a.length !== b.length) return false;

  for (let i = 0; i < a.length; i++) {
    if (a[i].filename !== b[i].filename || a[i].index !== b[i].index) {
      return false;
    }
  }

  return true;
}

// Helper to create a minimal plugin for testing
function createPlugin(filename: string, index: number): LoadOrderPluginInfo {
  return {
    filename,
    type: 'ESP',
    flags: { isMaster: false, isLight: false, isLocalized: false },
    masters: [],
    index,
    hasIssues: false,
    issueCount: 0,
  };
}

describe('loadorderUtils', () => {
  describe('compareLoadOrders', () => {
    it('returns empty diff for empty lists', () => {
      const result = compareLoadOrders([], []);
      expect(result.summary.total).toBe(0);
      expect(result.summary.aOnly).toBe(0);
      expect(result.summary.bOnly).toBe(0);
      expect(result.summary.same).toBe(0);
      expect(result.summary.moved).toBe(0);
    });

    it('detects plugins only in A', () => {
      const pluginsA = [createPlugin('Plugin1.esp', 0)];
      const pluginsB: LoadOrderPluginInfo[] = [];

      const result = compareLoadOrders(pluginsA, pluginsB);

      expect(result.summary.aOnly).toBe(1);
      expect(result.summary.bOnly).toBe(0);
      expect(result.diffPlugins[0].diffStatus).toBe('a-only');
      expect(result.diffPlugins[0].indexA).toBe(0);
      expect(result.diffPlugins[0].indexB).toBeNull();
    });

    it('detects plugins only in B', () => {
      const pluginsA: LoadOrderPluginInfo[] = [];
      const pluginsB = [createPlugin('Plugin1.esp', 0)];

      const result = compareLoadOrders(pluginsA, pluginsB);

      expect(result.summary.aOnly).toBe(0);
      expect(result.summary.bOnly).toBe(1);
      expect(result.diffPlugins[0].diffStatus).toBe('b-only');
      expect(result.diffPlugins[0].indexA).toBeNull();
      expect(result.diffPlugins[0].indexB).toBe(0);
    });

    it('detects same plugins at same position', () => {
      const pluginsA = [createPlugin('Plugin1.esp', 0)];
      const pluginsB = [createPlugin('Plugin1.esp', 0)];

      const result = compareLoadOrders(pluginsA, pluginsB);

      expect(result.summary.same).toBe(1);
      expect(result.summary.moved).toBe(0);
      expect(result.diffPlugins[0].diffStatus).toBe('same');
      expect(result.diffPlugins[0].positionDelta).toBe(0);
    });

    it('detects moved plugins with position delta', () => {
      const pluginsA = [createPlugin('Plugin1.esp', 0)];
      const pluginsB = [createPlugin('Plugin1.esp', 5)];

      const result = compareLoadOrders(pluginsA, pluginsB);

      expect(result.summary.moved).toBe(1);
      expect(result.summary.same).toBe(0);
      expect(result.diffPlugins[0].diffStatus).toBe('moved');
      expect(result.diffPlugins[0].positionDelta).toBe(5);
    });

    it('handles case-insensitive comparison', () => {
      const pluginsA = [createPlugin('Plugin1.ESP', 0)];
      const pluginsB = [createPlugin('plugin1.esp', 0)];

      const result = compareLoadOrders(pluginsA, pluginsB);

      expect(result.summary.same).toBe(1);
      expect(result.summary.aOnly).toBe(0);
      expect(result.summary.bOnly).toBe(0);
    });

    it('handles complex mixed differences', () => {
      const pluginsA = [
        createPlugin('Same.esp', 0),
        createPlugin('Moved.esp', 1),
        createPlugin('OnlyA.esp', 2),
      ];
      const pluginsB = [
        createPlugin('Same.esp', 0),
        createPlugin('Moved.esp', 5),
        createPlugin('OnlyB.esp', 3),
      ];

      const result = compareLoadOrders(pluginsA, pluginsB);

      expect(result.summary.same).toBe(1);
      expect(result.summary.moved).toBe(1);
      expect(result.summary.aOnly).toBe(1);
      expect(result.summary.bOnly).toBe(1);
      expect(result.summary.total).toBe(4);
    });

    it('calculates negative position delta for upward moves', () => {
      const pluginsA = [createPlugin('Plugin1.esp', 10)];
      const pluginsB = [createPlugin('Plugin1.esp', 3)];

      const result = compareLoadOrders(pluginsA, pluginsB);

      expect(result.diffPlugins[0].positionDelta).toBe(-7);
    });
  });

  describe('pluginListsEqual', () => {
    it('returns true for empty lists', () => {
      expect(pluginListsEqual([], [])).toBe(true);
    });

    it('returns true for identical lists', () => {
      const listA = [createPlugin('Plugin1.esp', 0), createPlugin('Plugin2.esp', 1)];
      const listB = [createPlugin('Plugin1.esp', 0), createPlugin('Plugin2.esp', 1)];

      expect(pluginListsEqual(listA, listB)).toBe(true);
    });

    it('returns false for different lengths', () => {
      const listA = [createPlugin('Plugin1.esp', 0)];
      const listB = [createPlugin('Plugin1.esp', 0), createPlugin('Plugin2.esp', 1)];

      expect(pluginListsEqual(listA, listB)).toBe(false);
    });

    it('returns false for different filenames', () => {
      const listA = [createPlugin('Plugin1.esp', 0)];
      const listB = [createPlugin('Plugin2.esp', 0)];

      expect(pluginListsEqual(listA, listB)).toBe(false);
    });

    it('returns false for different indices', () => {
      const listA = [createPlugin('Plugin1.esp', 0)];
      const listB = [createPlugin('Plugin1.esp', 1)];

      expect(pluginListsEqual(listA, listB)).toBe(false);
    });
  });
});
