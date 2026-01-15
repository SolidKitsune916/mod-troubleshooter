import { useMemo } from 'react';
import type { GameId } from '@/types';
import './HeroBanner.css';

interface HeroBannerProps {
  gameId: GameId;
}

// Steam capsule image URLs for each game
const GAME_HERO_IMAGES: Record<GameId, string> = {
  skyrim: 'https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/489830/capsule_616x353.jpg?t=1753715778',
  stardew: 'https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/413150/capsule_616x353.jpg?t=1754692865',
  cyberpunk: 'https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/1495710/capsule_616x353.jpg?t=1696436134',
};

// Get Steam capsule image URL for the game
function getHeroImageUrl(gameId: GameId): string {
  return GAME_HERO_IMAGES[gameId];
}

export function HeroBanner({ gameId }: HeroBannerProps) {
  const heroImageUrl = useMemo(() => getHeroImageUrl(gameId), [gameId]);
  
  // Game-specific alt text
  const altText = useMemo(() => {
    const gameNames: Record<GameId, string> = {
      skyrim: 'The Elder Scrolls V: Skyrim Special Edition',
      stardew: 'Stardew Valley',
      cyberpunk: 'Cyberpunk 2077',
    };
    return `${gameNames[gameId]} hero banner`;
  }, [gameId]);

  return (
    <div className="hero-banner" role="img" aria-label={altText}>
      <img 
        src={heroImageUrl} 
        alt={altText}
        className="hero-image"
        loading="eager"
        onError={(e) => {
          // Fallback to a gradient if image fails to load
          const target = e.target as HTMLImageElement;
          target.style.display = 'none';
          target.parentElement?.classList.add('hero-banner--fallback');
        }}
      />
      <div className="hero-overlay"></div>
    </div>
  );
}
