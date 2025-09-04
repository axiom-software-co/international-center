// Date formatting utilities for International Center Platform
// Provides consistent date formatting across the application

/**
 * Format date for display in articles and news
 */
export function formatArticleDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

/**
 * Format date with time for detailed views
 */
export function formatArticleDateTime(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  });
}

/**
 * Calculate time ago (e.g., "2 hours ago", "3 days ago")
 */
export function getTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);
  const diffMonths = Math.floor(diffDays / 30);
  const diffYears = Math.floor(diffDays / 365);

  if (diffSecs < 60) return 'Just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  if (diffDays < 30) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
  if (diffMonths < 12) return `${diffMonths} month${diffMonths !== 1 ? 's' : ''} ago`;
  return `${diffYears} year${diffYears !== 1 ? 's' : ''} ago`;
}

/**
 * Check if a date is recent (within last 7 days)
 */
export function isRecentDate(dateString: string): boolean {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  return diffDays <= 7;
}

/**
 * Get the most recent date between published_at and created_on
 */
export function getDisplayDate(published_at?: string, created_on?: string): string {
  const publishedDate = published_at ? new Date(published_at) : null;
  const createdDate = created_on ? new Date(created_on) : null;
  
  // Use published date if available and valid
  if (publishedDate && !isNaN(publishedDate.getTime())) {
    return published_at!;
  }
  
  // Fall back to created date
  if (createdDate && !isNaN(createdDate.getTime())) {
    return created_on!;
  }
  
  // Last resort - current date
  return new Date().toISOString();
}