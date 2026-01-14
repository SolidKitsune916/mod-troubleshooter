import { useState, useCallback, useMemo } from 'react';

import { useFomodAnalysis } from '@hooks/useFomod.ts';
import { ApiError } from '@services/api.ts';

import type {
  Dependency,
  FomodData,
  InstallStep,
  OptionGroup,
  Plugin,
  GroupType,
  PluginType,
} from '@/types/index.ts';

// ============================================
// Condition Flag Types and Helpers
// ============================================

/** Map of flag names to their current values */
type FlagState = Map<string, string>;

/**
 * Evaluates a dependency condition against current flag state.
 * Returns true if the condition is satisfied or undefined.
 */
function evaluateDependency(dep: Dependency | undefined, flags: FlagState): boolean {
  // No dependency means always visible
  if (!dep) {
    return true;
  }

  // Handle flag dependency
  if (dep.flagDependency) {
    const currentValue = flags.get(dep.flagDependency.flag);
    return currentValue === dep.flagDependency.value;
  }

  // Handle file dependency - for now, assume files are not present
  // This would need integration with actual file state tracking
  if (dep.fileDependency) {
    // Default behavior: Missing files are considered not present
    // Active/Inactive would need actual mod manager integration
    if (dep.fileDependency.state === 'Missing') {
      return true; // Files are assumed missing by default
    }
    return false;
  }

  // Handle game/fomm dependencies - assume satisfied for visualization
  if (dep.gameDependency || dep.fommDependency) {
    return true;
  }

  // Handle composite dependencies with children
  if (dep.children && dep.children.length > 0) {
    const operator = dep.operator ?? 'And';

    if (operator === 'And') {
      return dep.children.every(child => evaluateDependency(child, flags));
    } else {
      return dep.children.some(child => evaluateDependency(child, flags));
    }
  }

  // No specific condition, default to visible
  return true;
}

/**
 * Collects all condition flags set by selected plugins across all steps.
 */
function collectFlags(
  steps: InstallStep[],
  selections: Map<string, Set<string>>,
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

// ============================================
// Props Interfaces
// ============================================

interface FomodViewerProps {
  game: string;
  modId: number;
  fileId: number;
}

interface FomodHeaderProps {
  data: FomodData;
  cached: boolean;
}

interface FomodStepNavigatorProps {
  steps: InstallStep[];
  currentStepIndex: number;
  onStepChange: (index: number) => void;
  stepVisibility: boolean[];
}

interface FomodStepViewProps {
  step: InstallStep;
  selections: Map<string, Set<string>>;
  onSelectionChange: (groupName: string, pluginName: string, selected: boolean, groupType: GroupType) => void;
}

interface OptionGroupViewProps {
  group: OptionGroup;
  stepName: string;
  selectedPlugins: Set<string>;
  onSelectionChange: (pluginName: string, selected: boolean) => void;
}

interface PluginCardProps {
  plugin: Plugin;
  selected: boolean;
  onSelect: () => void;
  inputType: 'radio' | 'checkbox';
  groupName: string;
}

interface FomodSummaryProps {
  steps: InstallStep[];
  selections: Map<string, Set<string>>;
}

// ============================================
// Helper Functions
// ============================================

/** Get input type based on group type */
function getInputTypeForGroupType(groupType: GroupType): 'radio' | 'checkbox' {
  return groupType === 'SelectExactlyOne' || groupType === 'SelectAtMostOne'
    ? 'radio'
    : 'checkbox';
}

/** Get display label for group type */
function getGroupTypeLabel(groupType: GroupType): string {
  switch (groupType) {
    case 'SelectExactlyOne':
      return 'Select exactly one';
    case 'SelectAtMostOne':
      return 'Select at most one';
    case 'SelectAtLeastOne':
      return 'Select at least one';
    case 'SelectAny':
      return 'Select any';
    case 'SelectAll':
      return 'All required';
    default:
      return groupType;
  }
}

/** Get plugin type badge color */
function getPluginTypeBadgeClass(pluginType: PluginType | undefined): string {
  switch (pluginType) {
    case 'Required':
      return 'bg-error/20 text-error';
    case 'Recommended':
      return 'bg-accent/20 text-accent';
    case 'Optional':
      return 'bg-text-muted/20 text-text-secondary';
    case 'NotUsable':
      return 'bg-error/30 text-error line-through';
    case 'CouldBeUsable':
      return 'bg-warning/20 text-warning';
    default:
      return 'bg-text-muted/20 text-text-secondary';
  }
}

/** Get plugin type from descriptor */
function getPluginType(plugin: Plugin): PluginType {
  return plugin.typeDescriptor?.type ?? 'Optional';
}

// ============================================
// Loading Skeleton
// ============================================

const FomodSkeleton: React.FC = () => (
  <div className="space-y-6 animate-pulse">
    {/* Header skeleton */}
    <div className="p-6 rounded-sm bg-bg-card border border-border">
      <div className="space-y-3">
        <div className="h-8 w-1/2 bg-bg-secondary rounded-xs" />
        <div className="h-4 w-1/3 bg-bg-secondary rounded-xs" />
        <div className="h-4 w-2/3 bg-bg-secondary rounded-xs" />
      </div>
    </div>
    {/* Step navigator skeleton */}
    <div className="flex gap-2 p-4 rounded-sm bg-bg-card border border-border">
      {[1, 2, 3].map((i) => (
        <div key={i} className="h-10 w-32 bg-bg-secondary rounded-xs" />
      ))}
    </div>
    {/* Step content skeleton */}
    <div className="p-6 rounded-sm bg-bg-card border border-border space-y-4">
      <div className="h-6 w-1/4 bg-bg-secondary rounded-xs" />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-32 bg-bg-secondary rounded-xs" />
        ))}
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
  let message = 'An unexpected error occurred while analyzing the FOMOD.';

  if (error instanceof ApiError) {
    if (error.status === 404) {
      message = 'Mod file not found. Please check the mod ID and file ID.';
    } else if (error.status === 401 || error.status === 403) {
      message = 'API key is missing or invalid. Please configure the backend.';
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
// No FOMOD Message
// ============================================

const NoFomodMessage: React.FC = () => (
  <div className="p-6 rounded-sm bg-bg-card border border-border text-center">
    <p className="text-text-secondary text-lg mb-2">No FOMOD Found</p>
    <p className="text-text-muted">
      This mod file does not contain a FOMOD installer configuration.
    </p>
  </div>
);

// ============================================
// Header Component
// ============================================

const FomodHeader: React.FC<FomodHeaderProps> = ({ data, cached }) => {
  const info = data.info;
  const moduleName = data.config.moduleName;
  const displayName = info?.name ?? moduleName;

  return (
    <header className="p-6 rounded-sm bg-bg-card border border-border">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-2">
          <h2 className="text-2xl font-bold text-text-primary">{displayName}</h2>
          {info?.author && (
            <p className="text-text-secondary">
              by <span className="text-text-primary">{info.author}</span>
              {info.version && (
                <span className="text-text-muted ml-2">v{info.version}</span>
              )}
            </p>
          )}
          {info?.description && (
            <p className="text-text-muted text-sm max-w-2xl">{info.description}</p>
          )}
        </div>
        <div className="flex flex-col items-end gap-2">
          <span
            className={`px-3 py-1 rounded-full text-xs font-medium ${
              cached
                ? 'bg-accent/20 text-accent'
                : 'bg-text-muted/20 text-text-secondary'
            }`}
          >
            {cached ? 'Cached' : 'Fresh'}
          </span>
          {info?.website && (
            <a
              href={info.website}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-accent hover:text-accent/80
                focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
                transition-colors motion-reduce:transition-none"
            >
              View on Nexus
            </a>
          )}
        </div>
      </div>
    </header>
  );
};

// ============================================
// Step Navigator Component
// ============================================

const FomodStepNavigator: React.FC<FomodStepNavigatorProps> = ({
  steps,
  currentStepIndex,
  onStepChange,
  stepVisibility,
}) => (
  <nav
    aria-label="Installation steps"
    className="p-4 rounded-sm bg-bg-card border border-border"
  >
    <ol className="flex flex-wrap gap-2" role="list">
      {steps.map((step, index) => {
        const isActive = index === currentStepIndex;
        const isPast = index < currentStepIndex;
        const isVisible = stepVisibility[index];

        return (
          <li key={step.name}>
            <button
              onClick={() => onStepChange(index)}
              aria-current={isActive ? 'step' : undefined}
              aria-hidden={!isVisible}
              className={`min-h-11 px-4 py-2 rounded-sm font-medium transition-colors
                motion-reduce:transition-none
                focus-visible:outline-3 focus-visible:outline-focus focus-visible:outline-offset-2
                ${
                  !isVisible
                    ? 'bg-bg-secondary/50 text-text-muted opacity-50'
                    : isActive
                      ? 'bg-accent text-white'
                      : isPast
                        ? 'bg-accent/20 text-accent hover:bg-accent/30'
                        : 'bg-bg-secondary text-text-secondary hover:bg-bg-secondary/80'
                }`}
            >
              <span className="mr-2 text-sm opacity-60">{index + 1}.</span>
              {step.name}
              {!isVisible && (
                <span className="ml-2 text-xs">(hidden)</span>
              )}
            </button>
          </li>
        );
      })}
    </ol>
  </nav>
);

// ============================================
// Plugin Card Component
// ============================================

const PluginCard: React.FC<PluginCardProps> = ({
  plugin,
  selected,
  onSelect,
  inputType,
  groupName,
}) => {
  const pluginType = getPluginType(plugin);
  const isDisabled = pluginType === 'NotUsable';
  const inputId = `${groupName}-${plugin.name}`.replace(/\s+/g, '-').toLowerCase();

  return (
    <label
      htmlFor={inputId}
      className={`
        relative flex flex-col p-4 rounded-sm border cursor-pointer
        transition-colors motion-reduce:transition-none
        ${
          isDisabled
            ? 'bg-bg-secondary/50 border-border opacity-60 cursor-not-allowed'
            : selected
              ? 'bg-accent/10 border-accent'
              : 'bg-bg-card border-border hover:border-border-hover'
        }
        focus-within:outline-3 focus-within:outline-focus focus-within:outline-offset-2
      `}
    >
      <div className="flex items-start gap-3">
        <input
          id={inputId}
          type={inputType}
          name={groupName}
          checked={selected}
          disabled={isDisabled}
          onChange={onSelect}
          className="mt-1 w-5 h-5 accent-accent
            focus:outline-none"
          aria-describedby={plugin.description ? `${inputId}-desc` : undefined}
        />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-medium text-text-primary">{plugin.name}</span>
            <span
              className={`px-2 py-0.5 rounded-full text-xs font-medium ${getPluginTypeBadgeClass(pluginType)}`}
            >
              {pluginType}
            </span>
          </div>
          {plugin.description && (
            <p
              id={`${inputId}-desc`}
              className="mt-1 text-sm text-text-muted line-clamp-2"
            >
              {plugin.description}
            </p>
          )}
        </div>
      </div>
    </label>
  );
};

// ============================================
// Option Group View Component
// ============================================

const OptionGroupView: React.FC<OptionGroupViewProps> = ({
  group,
  stepName,
  selectedPlugins,
  onSelectionChange,
}) => {
  const inputType = getInputTypeForGroupType(group.type);
  const groupId = `${stepName}-${group.name}`.replace(/\s+/g, '-').toLowerCase();

  return (
    <fieldset className="space-y-4">
      <legend className="flex items-center gap-2 text-lg font-semibold text-text-primary">
        <span>{group.name}</span>
        <span className="text-sm font-normal text-text-muted">
          ({getGroupTypeLabel(group.type)})
        </span>
      </legend>
      <div
        className="grid grid-cols-1 md:grid-cols-2 gap-3"
        role={inputType === 'radio' ? 'radiogroup' : 'group'}
        aria-labelledby={groupId}
      >
        {group.plugins?.map((plugin) => (
          <PluginCard
            key={plugin.name}
            plugin={plugin}
            selected={selectedPlugins.has(plugin.name)}
            onSelect={() =>
              onSelectionChange(plugin.name, !selectedPlugins.has(plugin.name))
            }
            inputType={inputType}
            groupName={`${stepName}-${group.name}`}
          />
        ))}
      </div>
    </fieldset>
  );
};

// ============================================
// Step View Component
// ============================================

const FomodStepView: React.FC<FomodStepViewProps> = ({
  step,
  selections,
  onSelectionChange,
}) => {
  const handleGroupSelectionChange = useCallback(
    (groupName: string, groupType: GroupType) =>
      (pluginName: string, selected: boolean) => {
        onSelectionChange(groupName, pluginName, selected, groupType);
      },
    [onSelectionChange],
  );

  return (
    <section
      aria-label={`Step: ${step.name}`}
      className="p-6 rounded-sm bg-bg-card border border-border space-y-8"
    >
      <h3 className="text-xl font-bold text-text-primary">{step.name}</h3>
      {step.optionGroups?.map((group) => (
        <OptionGroupView
          key={group.name}
          group={group}
          stepName={step.name}
          selectedPlugins={selections.get(`${step.name}-${group.name}`) ?? new Set()}
          onSelectionChange={handleGroupSelectionChange(
            `${step.name}-${group.name}`,
            group.type,
          )}
        />
      ))}
      {(!step.optionGroups || step.optionGroups.length === 0) && (
        <p className="text-text-muted text-center py-4">
          This step has no options to configure.
        </p>
      )}
    </section>
  );
};

// ============================================
// Summary Component
// ============================================

const FomodSummary: React.FC<FomodSummaryProps> = ({ steps, selections }) => {
  const hasSelections = Array.from(selections.values()).some((s) => s.size > 0);

  if (!hasSelections) {
    return (
      <aside
        aria-label="Selection summary"
        className="p-4 rounded-sm bg-bg-card border border-border"
      >
        <h3 className="text-lg font-semibold text-text-primary mb-2">Summary</h3>
        <p className="text-text-muted">No options selected yet.</p>
      </aside>
    );
  }

  return (
    <aside
      aria-label="Selection summary"
      className="p-4 rounded-sm bg-bg-card border border-border"
    >
      <h3 className="text-lg font-semibold text-text-primary mb-4">Summary</h3>
      <ul className="space-y-3" role="list">
        {steps.map((step) => {
          const stepSelections: string[] = [];
          step.optionGroups?.forEach((group) => {
            const groupKey = `${step.name}-${group.name}`;
            const selected = selections.get(groupKey);
            if (selected?.size) {
              stepSelections.push(...Array.from(selected));
            }
          });

          if (stepSelections.length === 0) return null;

          return (
            <li key={step.name}>
              <p className="text-text-secondary font-medium">{step.name}</p>
              <ul className="mt-1 ml-4 space-y-1" role="list">
                {stepSelections.map((name) => (
                  <li key={name} className="text-text-muted text-sm flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-accent" aria-hidden="true" />
                    {name}
                  </li>
                ))}
              </ul>
            </li>
          );
        })}
      </ul>
    </aside>
  );
};

// ============================================
// Main FomodViewer Component
// ============================================

/** Interactive FOMOD installer visualization */
export const FomodViewer: React.FC<FomodViewerProps> = ({ game, modId, fileId }) => {
  const [currentStepIndex, setCurrentStepIndex] = useState(0);
  const [selections, setSelections] = useState<Map<string, Set<string>>>(new Map());

  const { data, isLoading, error, refetch } = useFomodAnalysis(game, modId, fileId);

  const steps = useMemo(() => data?.data?.config.installSteps ?? [], [data]);

  // Calculate condition flags from current selections
  const flags = useMemo(
    () => collectFlags(steps, selections),
    [steps, selections],
  );

  // Calculate visibility for each step based on its dependency conditions
  const stepVisibility = useMemo(
    () => steps.map(step => evaluateDependency(step.visible, flags)),
    [steps, flags],
  );

  // Count visible steps for screen reader announcement
  const visibleStepCount = useMemo(
    () => stepVisibility.filter(Boolean).length,
    [stepVisibility],
  );

  const handleSelectionChange = useCallback(
    (groupKey: string, pluginName: string, selected: boolean, groupType: GroupType) => {
      setSelections((prev) => {
        const newSelections = new Map(prev);
        const currentSet = new Set(prev.get(groupKey) ?? []);

        if (groupType === 'SelectExactlyOne' || groupType === 'SelectAtMostOne') {
          // Radio behavior - clear and set
          currentSet.clear();
          if (selected) {
            currentSet.add(pluginName);
          }
        } else {
          // Checkbox behavior - toggle
          if (selected) {
            currentSet.add(pluginName);
          } else {
            currentSet.delete(pluginName);
          }
        }

        newSelections.set(groupKey, currentSet);
        return newSelections;
      });
    },
    [],
  );

  const handleStepChange = useCallback((index: number) => {
    setCurrentStepIndex(index);
  }, []);

  // Auto-navigate to next visible step if current step becomes hidden
  const currentStepVisible = stepVisibility[currentStepIndex];
  const adjustedStepIndex = useMemo(() => {
    if (currentStepVisible) {
      return currentStepIndex;
    }
    // Find next visible step
    const nextVisible = stepVisibility.findIndex(
      (visible, i) => visible && i > currentStepIndex,
    );
    if (nextVisible !== -1) {
      return nextVisible;
    }
    // Find previous visible step
    for (let i = currentStepIndex - 1; i >= 0; i--) {
      if (stepVisibility[i]) {
        return i;
      }
    }
    // Fall back to first step
    return 0;
  }, [currentStepIndex, currentStepVisible, stepVisibility]);

  // Update step index if it was adjusted
  if (adjustedStepIndex !== currentStepIndex && steps.length > 0) {
    setCurrentStepIndex(adjustedStepIndex);
  }

  // Loading state
  if (isLoading) {
    return <FomodSkeleton />;
  }

  // Error state
  if (error) {
    return <ErrorDisplay error={error} onRetry={() => refetch()} />;
  }

  // No FOMOD found
  if (!data?.hasFomod || !data.data) {
    return <NoFomodMessage />;
  }

  const currentStep = steps[currentStepIndex];

  return (
    <div className="space-y-6">
      <div aria-live="polite" className="sr-only">
        Loaded FOMOD installer with {steps.length} steps ({visibleStepCount} visible based on current selections)
      </div>

      <FomodHeader data={data.data} cached={data.cached} />

      {steps.length > 0 && (
        <>
          <FomodStepNavigator
            steps={steps}
            currentStepIndex={currentStepIndex}
            onStepChange={handleStepChange}
            stepVisibility={stepVisibility}
          />

          {currentStep && stepVisibility[currentStepIndex] && (
            <FomodStepView
              step={currentStep}
              selections={selections}
              onSelectionChange={handleSelectionChange}
            />
          )}

          {currentStep && !stepVisibility[currentStepIndex] && (
            <div className="p-6 rounded-sm bg-bg-card border border-border text-center">
              <p className="text-text-muted">
                This step is hidden based on your current selections.
              </p>
            </div>
          )}

          <FomodSummary steps={steps} selections={selections} />
        </>
      )}

      {steps.length === 0 && (
        <div className="p-6 rounded-sm bg-bg-card border border-border text-center">
          <p className="text-text-secondary">
            This FOMOD does not have any installation steps to configure.
          </p>
          {data.data.config.requiredInstallFiles && (
            <p className="text-text-muted mt-2 text-sm">
              Required files will be installed automatically.
            </p>
          )}
        </div>
      )}
    </div>
  );
};
