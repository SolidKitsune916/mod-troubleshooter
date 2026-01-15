import type { ModFileReference } from '@/types/index.ts';
import { ModCard } from './ModCard.tsx';
import styles from './ModList.module.css';

interface ModListProps {
  modFiles: ModFileReference[];
}

/** Displays a list of mods from a collection revision */
export const ModList: React.FC<ModListProps> = ({ modFiles }) => {
  const requiredMods = modFiles.filter((m) => !m.optional);
  const optionalMods = modFiles.filter((m) => m.optional);

  if (modFiles.length === 0) {
    return (
      <div className={styles.emptyState}>
        <p>No mods found in this collection.</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      {requiredMods.length > 0 && (
        <section aria-labelledby="required-mods-heading" className={styles.section}>
          <h3 id="required-mods-heading" className={styles.sectionTitle}>
            Required Mods ({requiredMods.length})
          </h3>
          <div className={styles.modList}>
            {requiredMods.map((modFile) => (
              <ModCard key={modFile.fileId} modFile={modFile} />
            ))}
          </div>
        </section>
      )}

      {optionalMods.length > 0 && (
        <section aria-labelledby="optional-mods-heading" className={styles.section}>
          <h3 id="optional-mods-heading" className={styles.sectionTitle}>
            Optional Mods ({optionalMods.length})
          </h3>
          <div className={styles.modList}>
            {optionalMods.map((modFile) => (
              <ModCard key={modFile.fileId} modFile={modFile} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
};
