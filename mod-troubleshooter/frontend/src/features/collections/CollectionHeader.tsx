import type { Collection } from '@/types/index.ts';

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
    <header className="flex gap-6 p-6 rounded-sm bg-bg-card border border-border">
      {collection.tileImage?.url ? (
        <img
          src={collection.tileImage.url}
          alt=""
          className="w-32 h-32 rounded-xs object-cover flex-shrink-0"
        />
      ) : (
        <div
          className="w-32 h-32 rounded-xs bg-bg-secondary flex-shrink-0
            flex items-center justify-center text-text-muted"
          aria-hidden="true"
        >
          No Image
        </div>
      )}

      <div className="flex-1 min-w-0">
        <h2 className="text-2xl font-bold text-text-primary mb-2">
          {collection.name}
        </h2>

        <div className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-text-secondary mb-3">
          <span>
            by{' '}
            <span className="text-text-primary font-medium">
              {collection.user.name}
            </span>
          </span>
          <span>
            for{' '}
            <span className="text-text-primary font-medium">
              {collection.game.name}
            </span>
          </span>
          <span>{modCount} mods</span>
          <span>{formatNumber(collection.totalDownloads)} downloads</span>
          <span>{formatNumber(collection.endorsements)} endorsements</span>
        </div>

        <p className="text-text-secondary line-clamp-3">{collection.summary}</p>
      </div>
    </header>
  );
};
