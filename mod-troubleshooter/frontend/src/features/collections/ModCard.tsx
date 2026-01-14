import type { ModFileReference } from '@/types/index.ts';

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
    <article
      className="flex gap-4 p-4 rounded-sm bg-bg-card border border-border
        hover:border-border-hover hover:bg-bg-hover
        transition-colors motion-reduce:transition-none"
    >
      {pictureUrl ? (
        <img
          src={pictureUrl}
          alt=""
          className="w-20 h-20 rounded-xs object-cover flex-shrink-0"
          loading="lazy"
        />
      ) : (
        <div
          className="w-20 h-20 rounded-xs bg-bg-secondary flex-shrink-0
            flex items-center justify-center text-text-muted"
          aria-hidden="true"
        >
          No Image
        </div>
      )}

      <div className="flex-1 min-w-0">
        <div className="flex items-start gap-2 mb-1">
          <h3 className="font-semibold text-text-primary truncate">{modName}</h3>
          {isOptional && (
            <span className="flex-shrink-0 px-2 py-0.5 text-xs rounded-xs bg-secondary/20 text-secondary">
              Optional
            </span>
          )}
        </div>

        <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-text-secondary mb-2">
          {modVersion && <span>v{modVersion}</span>}
          <span>by {modAuthor}</span>
          {fileSize > 0 && <span>{formatFileSize(fileSize)}</span>}
          {category && (
            <span className="text-text-muted">{category}</span>
          )}
        </div>

        {modSummary && (
          <p className="text-sm text-text-secondary line-clamp-2">{modSummary}</p>
        )}
      </div>
    </article>
  );
};
