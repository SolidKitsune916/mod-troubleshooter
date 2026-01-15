import type { ModFileReference } from '@/types/index.ts';
import styles from './ModCard.module.css';

interface ModCardProps {
  modFile: ModFileReference;
}

/** Formats file size to human-readable string */
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const index = Math.floor(Math.log(bytes) / Math.log(1024));
  const size = bytes / Math.pow(1024, index);
  return `${size.toFixed(index > 0 ? 1 : 0)} ${units[index]}`;
}

/** Displays a single mod file from a collection */
export const ModCard: React.FC<ModCardProps> = ({ modFile }) => {
  const file = modFile.file;
  const mod = file?.mod;

  const modName = mod?.name ?? file?.name ?? 'Unknown Mod';
  const modVersion = file?.version ?? mod?.version ?? '';
  const modAuthor = mod?.author ?? 'Unknown Author';
  const modSummary = mod?.summary ?? '';
  const fileSize = file?.size ?? 0;
  const pictureUrl = mod?.pictureUrl ?? '';
  const category = mod?.modCategory?.name ?? '';
  const isOptional = modFile.optional;

  return (
    <article className={styles.card}>
      {pictureUrl ? (
        <img
          src={pictureUrl}
          alt=""
          className={styles.image}
          loading="lazy"
        />
      ) : (
        <div className={styles.imagePlaceholder} aria-hidden="true">
          No Image
        </div>
      )}

      <div className={styles.content}>
        <div className={styles.titleRow}>
          <h3 className={styles.title}>{modName}</h3>
          {isOptional && (
            <span className={styles.optionalBadge}>Optional</span>
          )}
        </div>

        <div className={styles.meta}>
          {modVersion && <span>v{modVersion}</span>}
          <span>by {modAuthor}</span>
          {fileSize > 0 && <span>{formatFileSize(fileSize)}</span>}
          {category && <span className={styles.category}>{category}</span>}
        </div>

        {modSummary && <p className={styles.summary}>{modSummary}</p>}
      </div>
    </article>
  );
};
