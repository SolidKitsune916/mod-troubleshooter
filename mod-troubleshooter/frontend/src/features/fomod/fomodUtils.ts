/**
 * Shared FOMOD utility functions for file collection and flag management.
 */

import type {
  InstallStep,
  ModuleConfig,
  FileList,
  Dependency,
} from '@/types/index.ts';

// ============================================
// Types
// ============================================

/** Map of flag names to their current values */
export type FlagState = Map<string, string>;

/** Selection state: groupKey -> Set of selected plugin names */
export type SelectionsMap = Map<string, Set<string>>;

/** Represents a file to be installed */
export interface InstallFile {
  source: string;
  destination: string;
  priority: number;
  isFolder: boolean;
  category: 'required' | 'selected' | 'conditional';
}

// ============================================
// Helper Functions
// ============================================

/**
 * Evaluates a dependency condition against current flag state.
 */
export function evaluateDependency(dep: Dependency | undefined, flags: FlagState): boolean {
  if (!dep) return true;

  if (dep.flagDependency) {
    const currentValue = flags.get(dep.flagDependency.flag);
    return currentValue === dep.flagDependency.value;
  }

  if (dep.fileDependency) {
    if (dep.fileDependency.state === 'Missing') return true;
    return false;
  }

  if (dep.gameDependency || dep.fommDependency) return true;

  if (dep.children && dep.children.length > 0) {
    const operator = dep.operator ?? 'And';
    if (operator === 'And') {
      return dep.children.every(child => evaluateDependency(child, flags));
    }
    return dep.children.some(child => evaluateDependency(child, flags));
  }

  return true;
}

/**
 * Collects all condition flags from selected plugins.
 */
export function collectFlags(
  steps: InstallStep[],
  selections: SelectionsMap,
): FlagState {
  const flags: FlagState = new Map();

  for (const step of steps) {
    if (!step.optionGroups) continue;

    for (const group of step.optionGroups) {
      const groupKey = `${step.name}-${group.name}`;
      const selectedPlugins = selections.get(groupKey);

      if (!selectedPlugins || !group.plugins) continue;

      for (const plugin of group.plugins) {
        if (selectedPlugins.has(plugin.name) && plugin.conditionFlags) {
          for (const flag of plugin.conditionFlags) {
            flags.set(flag.name, flag.value);
          }
        }
      }
    }
  }

  return flags;
}

/**
 * Extracts files from a FileList structure.
 */
export function extractFilesFromFileList(
  fileList: FileList | undefined,
  category: InstallFile['category'],
): InstallFile[] {
  const result: InstallFile[] = [];

  if (fileList?.files) {
    for (const file of fileList.files) {
      result.push({
        source: file.source,
        destination: file.destination ?? file.source,
        priority: file.priority ?? 0,
        isFolder: false,
        category,
      });
    }
  }

  if (fileList?.folders) {
    for (const folder of fileList.folders) {
      result.push({
        source: folder.source,
        destination: folder.destination ?? folder.source,
        priority: folder.priority ?? 0,
        isFolder: true,
        category,
      });
    }
  }

  return result;
}

/**
 * Collects all files to be installed based on selections.
 */
export function collectInstallFiles(
  config: ModuleConfig,
  steps: InstallStep[],
  selections: SelectionsMap,
  flags: FlagState,
): InstallFile[] {
  const files: InstallFile[] = [];

  // Required install files
  files.push(...extractFilesFromFileList(config.requiredInstallFiles, 'required'));

  // Files from selected plugins
  for (const step of steps) {
    if (!step.optionGroups) continue;

    for (const group of step.optionGroups) {
      const groupKey = `${step.name}-${group.name}`;
      const selectedPlugins = selections.get(groupKey);

      if (!selectedPlugins || !group.plugins) continue;

      for (const plugin of group.plugins) {
        if (selectedPlugins.has(plugin.name) && plugin.files) {
          files.push(...extractFilesFromFileList(plugin.files, 'selected'));
        }
      }
    }
  }

  // Conditional file installs
  if (config.conditionalFileInstalls) {
    for (const item of config.conditionalFileInstalls) {
      if (evaluateDependency(item.dependencies, flags)) {
        files.push(...extractFilesFromFileList(item.files, 'conditional'));
      }
    }
  }

  return files;
}
