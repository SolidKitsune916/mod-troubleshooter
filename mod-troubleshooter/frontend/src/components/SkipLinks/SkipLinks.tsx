import './SkipLinks.css';

export function SkipLinks() {
  return (
    <nav className="skip-links" aria-label="Skip navigation">
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>
      <a href="#sidebar-nav" className="skip-link">
        Skip to navigation
      </a>
    </nav>
  );
}
