import { describe, it, expect } from 'vitest';
import {
  collectFlags,
  collectInstallFiles,
  evaluateDependency,
} from './fomodUtils.ts';
import type { InstallStep, ModuleConfig, Dependency } from '@/types/index.ts';

describe('fomodUtils', () => {
  describe('evaluateDependency', () => {
    it('returns true for undefined dependency', () => {
      expect(evaluateDependency(undefined, new Map())).toBe(true);
    });

    it('evaluates flag dependency correctly when flag matches', () => {
      const dep: Dependency = {
        flagDependency: { flag: 'TestFlag', value: 'true' },
      };
      const flags = new Map([['TestFlag', 'true']]);
      expect(evaluateDependency(dep, flags)).toBe(true);
    });

    it('evaluates flag dependency correctly when flag does not match', () => {
      const dep: Dependency = {
        flagDependency: { flag: 'TestFlag', value: 'true' },
      };
      const flags = new Map([['TestFlag', 'false']]);
      expect(evaluateDependency(dep, flags)).toBe(false);
    });

    it('evaluates And operator correctly', () => {
      const dep: Dependency = {
        operator: 'And',
        children: [
          { flagDependency: { flag: 'Flag1', value: 'true' } },
          { flagDependency: { flag: 'Flag2', value: 'true' } },
        ],
      };
      const flagsBothTrue = new Map([
        ['Flag1', 'true'],
        ['Flag2', 'true'],
      ]);
      const flagsOneFalse = new Map([
        ['Flag1', 'true'],
        ['Flag2', 'false'],
      ]);
      expect(evaluateDependency(dep, flagsBothTrue)).toBe(true);
      expect(evaluateDependency(dep, flagsOneFalse)).toBe(false);
    });

    it('evaluates Or operator correctly', () => {
      const dep: Dependency = {
        operator: 'Or',
        children: [
          { flagDependency: { flag: 'Flag1', value: 'true' } },
          { flagDependency: { flag: 'Flag2', value: 'true' } },
        ],
      };
      const flagsOneTrue = new Map([
        ['Flag1', 'true'],
        ['Flag2', 'false'],
      ]);
      const flagsBothFalse = new Map([
        ['Flag1', 'false'],
        ['Flag2', 'false'],
      ]);
      expect(evaluateDependency(dep, flagsOneTrue)).toBe(true);
      expect(evaluateDependency(dep, flagsBothFalse)).toBe(false);
    });
  });

  describe('collectFlags', () => {
    it('returns empty map when no steps', () => {
      const result = collectFlags([], new Map());
      expect(result.size).toBe(0);
    });

    it('collects flags from selected plugins', () => {
      const steps: InstallStep[] = [
        {
          name: 'Step1',
          optionGroups: [
            {
              name: 'Group1',
              type: 'SelectAny',
              plugins: [
                {
                  name: 'Plugin1',
                  conditionFlags: [
                    { name: 'Flag1', value: 'enabled' },
                  ],
                },
                {
                  name: 'Plugin2',
                  conditionFlags: [
                    { name: 'Flag2', value: 'active' },
                  ],
                },
              ],
            },
          ],
        },
      ];
      const selections = new Map([
        ['Step1-Group1', new Set(['Plugin1'])],
      ]);

      const result = collectFlags(steps, selections);

      expect(result.get('Flag1')).toBe('enabled');
      expect(result.has('Flag2')).toBe(false);
    });

    it('collects flags from multiple selected plugins', () => {
      const steps: InstallStep[] = [
        {
          name: 'Step1',
          optionGroups: [
            {
              name: 'Group1',
              type: 'SelectAny',
              plugins: [
                {
                  name: 'Plugin1',
                  conditionFlags: [{ name: 'Flag1', value: 'v1' }],
                },
                {
                  name: 'Plugin2',
                  conditionFlags: [{ name: 'Flag2', value: 'v2' }],
                },
              ],
            },
          ],
        },
      ];
      const selections = new Map([
        ['Step1-Group1', new Set(['Plugin1', 'Plugin2'])],
      ]);

      const result = collectFlags(steps, selections);

      expect(result.get('Flag1')).toBe('v1');
      expect(result.get('Flag2')).toBe('v2');
    });
  });

  describe('collectInstallFiles', () => {
    it('collects required files', () => {
      const config: ModuleConfig = {
        moduleName: 'Test',
        requiredInstallFiles: {
          files: [
            { source: 'required/file.esp' },
          ],
        },
      };

      const result = collectInstallFiles(config, [], new Map(), new Map());

      expect(result).toHaveLength(1);
      expect(result[0].source).toBe('required/file.esp');
      expect(result[0].category).toBe('required');
    });

    it('collects files from selected plugins', () => {
      const config: ModuleConfig = {
        moduleName: 'Test',
      };
      const steps: InstallStep[] = [
        {
          name: 'Step1',
          optionGroups: [
            {
              name: 'Group1',
              type: 'SelectAny',
              plugins: [
                {
                  name: 'Plugin1',
                  files: {
                    files: [{ source: 'plugin1/file.esp' }],
                  },
                },
                {
                  name: 'Plugin2',
                  files: {
                    files: [{ source: 'plugin2/file.esp' }],
                  },
                },
              ],
            },
          ],
        },
      ];
      const selections = new Map([
        ['Step1-Group1', new Set(['Plugin1'])],
      ]);

      const result = collectInstallFiles(config, steps, selections, new Map());

      expect(result).toHaveLength(1);
      expect(result[0].source).toBe('plugin1/file.esp');
      expect(result[0].category).toBe('selected');
    });

    it('collects conditional files when dependency is met', () => {
      const config: ModuleConfig = {
        moduleName: 'Test',
        conditionalFileInstalls: [
          {
            dependencies: {
              flagDependency: { flag: 'EnableConditional', value: 'true' },
            },
            files: {
              files: [{ source: 'conditional/file.esp' }],
            },
          },
        ],
      };
      const flags = new Map([['EnableConditional', 'true']]);

      const result = collectInstallFiles(config, [], new Map(), flags);

      expect(result).toHaveLength(1);
      expect(result[0].source).toBe('conditional/file.esp');
      expect(result[0].category).toBe('conditional');
    });

    it('does not collect conditional files when dependency is not met', () => {
      const config: ModuleConfig = {
        moduleName: 'Test',
        conditionalFileInstalls: [
          {
            dependencies: {
              flagDependency: { flag: 'EnableConditional', value: 'true' },
            },
            files: {
              files: [{ source: 'conditional/file.esp' }],
            },
          },
        ],
      };
      const flags = new Map([['EnableConditional', 'false']]);

      const result = collectInstallFiles(config, [], new Map(), flags);

      expect(result).toHaveLength(0);
    });

    it('collects folders correctly', () => {
      const config: ModuleConfig = {
        moduleName: 'Test',
        requiredInstallFiles: {
          folders: [
            { source: 'textures', destination: 'Data/Textures' },
          ],
        },
      };

      const result = collectInstallFiles(config, [], new Map(), new Map());

      expect(result).toHaveLength(1);
      expect(result[0].source).toBe('textures');
      expect(result[0].destination).toBe('Data/Textures');
      expect(result[0].isFolder).toBe(true);
    });

    it('uses source as destination when destination is not specified', () => {
      const config: ModuleConfig = {
        moduleName: 'Test',
        requiredInstallFiles: {
          files: [
            { source: 'required/file.esp' },
          ],
        },
      };

      const result = collectInstallFiles(config, [], new Map(), new Map());

      expect(result[0].destination).toBe('required/file.esp');
    });
  });
});
