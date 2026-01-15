import './Header.css';
import type { GameId } from '@/types';
import { GAMES } from '@/types';

interface HeaderProps {
  collectionCount?: number;
  totalMods?: number;
  lastUpdated?: string;
  onMenuToggle?: () => void;
  currentGame: GameId;
  onGameChange: (game: GameId) => void;
}

export function Header({ 
  collectionCount, 
  totalMods, 
  lastUpdated, 
  onMenuToggle, 
  currentGame, 
  onGameChange 
}: HeaderProps) {
  const formattedDate = lastUpdated ? new Date(lastUpdated).toLocaleDateString() : 'Unknown';

  return (
    <header className="header" role="banner">
      <div className="header-left">
        {onMenuToggle && (
          <button
            className="mobile-menu-btn"
            onClick={onMenuToggle}
            aria-label="Open menu"
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z" />
            </svg>
          </button>
        )}
        <div className="logo-container">
          <h1>Mod Troubleshooter</h1>
        </div>
      </div>
      <div className="game-selector">
        <select
          value={currentGame}
          onChange={(e) => onGameChange(e.target.value as GameId)}
          aria-label="Select game"
        >
          {GAMES.map((game) => (
            <option key={game.id} value={game.id}>
              {game.label}
            </option>
          ))}
        </select>
      </div>
      {(collectionCount !== undefined || totalMods !== undefined) && (
        <div className="header-info">
          {collectionCount !== undefined && <span>{collectionCount} collections</span>}
          {collectionCount !== undefined && totalMods !== undefined && (
            <span className="header-separator">|</span>
          )}
          {totalMods !== undefined && <span>{totalMods.toLocaleString()} mods</span>}
          {lastUpdated && (
            <>
              <span className="header-separator">|</span>
              <span>Last updated: {formattedDate}</span>
            </>
          )}
        </div>
      )}
    </header>
  );
}
