import type { Collection } from '@/types/index.ts';
import styles from './CollectionHeader.module.css';

interface CollectionHeaderProps {
  collection: Collection;
}

/** Formats a number with locale-specific separators */
function formatNumber(num: number): string {
  return num.toLocaleString();
}

/** Displays collection metadata header */
export const CollectionHeader: React.FC<CollectionHeaderProps> = ({
  collection,
}) => {
  const modCount = collection.latestPublishedRevision?.modFiles.length ?? 0;

  return (
    <header className={styles.header}>
      {collection.tileImage?.url ? (
        <img
          src={collection.tileImage.url}
          alt=""
          className={styles.image}
        />
      ) : (
        <div className={styles.imagePlaceholder} aria-hidden="true">
          No Image
        </div>
      )}

      <div className={styles.content}>
        <h2 className={styles.title}>{collection.name}</h2>

        <div className={styles.meta}>
          <span>
            by{' '}
            <span className={styles.metaHighlight}>{collection.user.name}</span>
          </span>
          <span>
            for{' '}
            <span className={styles.metaHighlight}>{collection.game.name}</span>
          </span>
          <span>{modCount} mods</span>
          <span>{formatNumber(collection.totalDownloads)} downloads</span>
          <span>{formatNumber(collection.endorsements)} endorsements</span>
        </div>

        <p className={styles.summary}>{collection.summary}</p>
      </div>
    </header>
  );
};
