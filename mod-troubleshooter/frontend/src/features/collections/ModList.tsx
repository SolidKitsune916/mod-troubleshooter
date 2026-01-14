import type { ModFileReference } from '@/types/index.ts';

import { ModCard } from './ModCard.tsx';

interface ModListProps {
  modFiles: ModFileReference[];
}

/** Displays a list of mods from a collection revision */
export const ModList: React.FC<ModListProps> = ({ modFiles }) => {
  const requiredMods = modFiles.filter((m) => !m.optional);
  const optionalMods = modFiles.filter((m) => m.optional);

  if (modFiles.length === 0) {
    return (
      <div className="text-center py-8 text-text-secondary">
        <p>No mods found in this collection.</p>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {requiredMods.length > 0 && (
        <section aria-labelledby="required-mods-heading">
          <h3
            id="required-mods-heading"
            className="text-lg font-semibold text-text-primary mb-4"
          >
            Required Mods ({requiredMods.length})
          </h3>
          <div className="space-y-3">
            {requiredMods.map((modFile) => (
              <ModCard key={modFile.fileId} modFile={modFile} />
            ))}
          </div>
        </section>
      )}

      {optionalMods.length > 0 && (
        <section aria-labelledby="optional-mods-heading">
          <h3
            id="optional-mods-heading"
            className="text-lg font-semibold text-text-primary mb-4"
          >
            Optional Mods ({optionalMods.length})
          </h3>
          <div className="space-y-3">
            {optionalMods.map((modFile) => (
              <ModCard key={modFile.fileId} modFile={modFile} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
};
