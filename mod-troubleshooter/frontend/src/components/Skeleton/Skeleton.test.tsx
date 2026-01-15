import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import {
  Skeleton,
  SkeletonText,
  SkeletonCard,
  SkeletonList,
  SkeletonGrid,
  SkeletonStats,
} from './Skeleton.tsx';

describe('Skeleton Components', () => {
  describe('Skeleton', () => {
    it('renders with default props', () => {
      render(<Skeleton ariaLabel="Loading" />);
      expect(screen.getByRole('status')).toBeInTheDocument();
      expect(screen.getByText('Loading')).toBeInTheDocument();
    });

    it('renders with custom dimensions', () => {
      render(<Skeleton width="100px" height="50px" ariaLabel="Loading" />);
      const skeleton = screen.getByRole('status');
      expect(skeleton).toHaveStyle({ width: '100px', height: '50px' });
    });

    it('applies circular variant class', () => {
      render(<Skeleton variant="circular" ariaLabel="Loading" />);
      const skeleton = screen.getByRole('status');
      expect(skeleton.className).toContain('rounded-full');
    });

    it('applies custom className', () => {
      render(<Skeleton className="custom-class" ariaLabel="Loading" />);
      const skeleton = screen.getByRole('status');
      expect(skeleton.className).toContain('custom-class');
    });

    it('has aria-busy attribute', () => {
      render(<Skeleton ariaLabel="Loading" />);
      expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
    });
  });

  describe('SkeletonText', () => {
    it('renders default number of lines', () => {
      const { container } = render(<SkeletonText />);
      const textContainer = container.querySelector('.space-y-2');
      // Default is 3 lines - each line is a Skeleton component
      expect(textContainer?.querySelectorAll('[role="status"]').length).toBe(3);
    });

    it('renders custom number of lines', () => {
      const { container } = render(<SkeletonText lines={5} />);
      const textContainer = container.querySelector('.space-y-2');
      expect(textContainer?.querySelectorAll('[role="status"]').length).toBe(5);
    });

    it('has accessible label', () => {
      const { container } = render(<SkeletonText ariaLabel="Loading paragraph" />);
      const textContainer = container.querySelector('.space-y-2');
      expect(textContainer).toHaveAttribute('aria-label', 'Loading paragraph');
    });
  });

  describe('SkeletonCard', () => {
    it('renders with header by default', () => {
      const { container } = render(<SkeletonCard />);
      const cardContainer = container.querySelector('.p-4.rounded-sm.bg-bg-card');
      expect(cardContainer).toBeInTheDocument();
      expect(cardContainer).toHaveAttribute('role', 'status');
    });

    it('renders without header when showHeader is false', () => {
      const { container } = render(<SkeletonCard showHeader={false} />);
      // First child should not be the header (wider skeleton)
      const firstChild = container.querySelector('[role="status"] > div:first-child');
      expect(firstChild).not.toHaveClass('mb-4');
    });

    it('renders action buttons when showActions is true', () => {
      const { container } = render(<SkeletonCard showActions />);
      const actionsContainer = container.querySelector('.mt-4.flex.gap-2');
      expect(actionsContainer).toBeInTheDocument();
    });

    it('renders correct number of content lines', () => {
      const { container } = render(<SkeletonCard contentLines={5} showHeader={false} />);
      const contentContainer = container.querySelector('.space-y-2');
      expect(contentContainer?.children.length).toBe(5);
    });
  });

  describe('SkeletonList', () => {
    it('renders default number of items', () => {
      const { container } = render(<SkeletonList />);
      const listContainer = container.querySelector('.space-y-2');
      // Default is 5 items
      expect(listContainer?.querySelectorAll('.flex.items-center').length).toBe(5);
    });

    it('renders custom number of items', () => {
      const { container } = render(<SkeletonList itemCount={3} />);
      const listContainer = container.querySelector('.space-y-2');
      expect(listContainer?.querySelectorAll('.flex.items-center').length).toBe(3);
    });

    it('shows icons by default', () => {
      const { container } = render(<SkeletonList itemCount={1} />);
      const iconContainer = container.querySelector('.flex.items-center.gap-3');
      expect(iconContainer?.children[0]).toBeInTheDocument();
    });

    it('hides icons when showIcon is false', () => {
      const { container } = render(<SkeletonList itemCount={1} showIcon={false} />);
      const item = container.querySelector('.flex.items-center.gap-3');
      // First child should be the flex-1 content, not an icon
      expect(item?.children[0]?.className).toContain('flex-1');
    });
  });

  describe('SkeletonGrid', () => {
    it('renders default number of items', () => {
      const { container } = render(<SkeletonGrid />);
      const gridContainer = container.querySelector('.grid');
      // Default is 6 items
      expect(gridContainer?.querySelectorAll('.rounded-sm').length).toBe(6);
    });

    it('renders custom number of items', () => {
      const { container } = render(<SkeletonGrid itemCount={4} />);
      const gridContainer = container.querySelector('.grid');
      expect(gridContainer?.querySelectorAll('.rounded-sm').length).toBe(4);
    });

    it('applies correct column classes', () => {
      const { container } = render(<SkeletonGrid columns={2} />);
      const gridContainer = container.querySelector('.grid');
      expect(gridContainer?.className).toContain('sm:grid-cols-2');
    });

    it('applies custom item height', () => {
      const { container } = render(<SkeletonGrid itemCount={1} itemHeight="10rem" />);
      const item = container.querySelector('.grid > div');
      expect(item).toHaveStyle({ height: '10rem' });
    });
  });

  describe('SkeletonStats', () => {
    it('renders default number of stats', () => {
      const { container } = render(<SkeletonStats />);
      const statsContainer = container.querySelector('.grid');
      // Default is 4 stats
      expect(statsContainer?.querySelectorAll('.space-y-2').length).toBe(4);
    });

    it('renders custom number of stats', () => {
      const { container } = render(<SkeletonStats statCount={6} />);
      const statsContainer = container.querySelector('.grid');
      expect(statsContainer?.querySelectorAll('.space-y-2').length).toBe(6);
    });

    it('has accessible label', () => {
      const { container } = render(<SkeletonStats ariaLabel="Loading stats" />);
      const statsContainer = container.querySelector('.grid');
      expect(statsContainer).toHaveAttribute('aria-label', 'Loading stats');
    });
  });
});
