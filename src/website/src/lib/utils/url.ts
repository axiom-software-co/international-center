// URL parsing utilities for International Center Platform
// Provides consistent URL parsing and manipulation

/**
 * Extract slug from URL pathname
 * @param pathname - The pathname to parse (e.g., '/company/news/article-slug')
 * @param segmentFromEnd - Which segment to extract from the end (0 = last segment)
 * @returns The extracted slug or empty string
 */
export function extractSlugFromPath(pathname: string, segmentFromEnd: number = 0): string {
  if (!pathname || typeof pathname !== 'string') {
    return '';
  }
  
  // Remove trailing slash and split into parts
  const cleanPath = pathname.replace(/\/$/, '');
  const parts = cleanPath.split('/').filter(part => part.length > 0);
  
  if (parts.length === 0) {
    return '';
  }
  
  // Get segment from the end
  const targetIndex = parts.length - 1 - segmentFromEnd;
  
  if (targetIndex < 0 || targetIndex >= parts.length) {
    return '';
  }
  
  return parts[targetIndex];
}

/**
 * Get current page slug from window location (client-side only)
 * @param segmentFromEnd - Which segment to extract from the end (0 = last segment)
 * @returns The extracted slug or empty string
 */
export function getCurrentSlug(segmentFromEnd: number = 0): string {
  if (typeof window === 'undefined') {
    return '';
  }
  
  return extractSlugFromPath(window.location.pathname, segmentFromEnd);
}

/**
 * Extract article slug from news URLs
 * Expected format: /company/news/article-slug
 * Returns empty string if URL is just /company/news/ (no article)
 */
export function getNewsSlugFromUrl(): string {
  if (typeof window === 'undefined') {
    return '';
  }
  
  const path = window.location.pathname;
  
  // Check if this is a news article path
  if (!path.startsWith('/company/news/')) {
    return '';
  }
  
  // Extract the part after /company/news/
  const afterNews = path.replace('/company/news/', '');
  const slug = afterNews.replace(/\/$/, ''); // Remove trailing slash
  
  // If it's empty or just contains slashes, return empty
  if (!slug || slug.match(/^\/+$/)) {
    return '';
  }
  
  // Return the first segment after /company/news/
  const segments = slug.split('/').filter(s => s.length > 0);
  return segments[0] || '';
}

/**
 * Extract service slug from service URLs  
 * Expected format: /services/service-slug
 */
export function getServiceSlugFromUrl(): string {
  return getCurrentSlug(0); // Last segment
}

/**
 * Extract event slug from event URLs
 * Expected format: /community/events/event-slug
 * Returns empty string if URL is just /community/events/ (no event)
 */
export function getEventSlugFromUrl(): string {
  if (typeof window === 'undefined') {
    return '';
  }
  
  const path = window.location.pathname;
  
  // Check if this is an events path
  if (!path.startsWith('/community/events/')) {
    return '';
  }
  
  // Extract the part after /community/events/
  const afterEvents = path.replace('/community/events/', '');
  const slug = afterEvents.replace(/\/$/, ''); // Remove trailing slash
  
  // If it's empty or just contains slashes, return empty
  if (!slug || slug.match(/^\/+$/)) {
    return '';
  }
  
  // Return the first segment after /community/events/
  const segments = slug.split('/').filter(s => s.length > 0);
  return segments[0] || '';
}

/**
 * Extract research slug from research URLs
 * Expected format: /community/research/research-slug
 * Returns empty string if URL is just /community/research/ (no article)
 */
export function getResearchSlugFromUrl(): string {
  if (typeof window === 'undefined') {
    return '';
  }
  
  const path = window.location.pathname;
  
  // Check if this is a research path
  if (!path.startsWith('/community/research/')) {
    return '';
  }
  
  // Extract the part after /community/research/
  const afterResearch = path.replace('/community/research/', '');
  const slug = afterResearch.replace(/\/$/, ''); // Remove trailing slash
  
  // If it's empty or just contains slashes, return empty
  if (!slug || slug.match(/^\/+$/)) {
    return '';
  }
  
  // Return the first segment after /community/research/
  const segments = slug.split('/').filter(s => s.length > 0);
  return segments[0] || '';
}

/**
 * Build a clean URL slug from a title
 */
export function createSlugFromTitle(title: string): string {
  if (!title) return '';
  
  return title
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '') // Remove special characters except spaces and hyphens
    .replace(/\s+/g, '-') // Replace spaces with hyphens
    .replace(/-+/g, '-') // Replace multiple hyphens with single hyphen
    .replace(/^-|-$/g, ''); // Remove leading/trailing hyphens
}

/**
 * Check if current path matches a pattern
 * @param pattern - Pattern to match (e.g., '/company/news/*')
 * @param pathname - Optional pathname (defaults to window.location.pathname)
 */
export function matchesPath(pattern: string, pathname?: string): boolean {
  if (typeof window === 'undefined' && !pathname) {
    return false;
  }
  
  const currentPath = pathname || window.location.pathname;
  
  // Convert pattern to regex
  const regexPattern = pattern
    .replace(/\*/g, '.*') // Replace * with .*
    .replace(/\//g, '\\/'); // Escape forward slashes
    
  const regex = new RegExp(`^${regexPattern}$`);
  return regex.test(currentPath);
}

/**
 * Get breadcrumb segments from URL path
 * @param pathname - The pathname to parse
 * @param baseSegments - Number of base segments to skip (e.g., 0 for all segments)
 * @returns Array of path segments for breadcrumbs
 */
export function getBreadcrumbSegments(pathname: string, baseSegments: number = 0): string[] {
  if (!pathname) return [];
  
  const segments = pathname
    .split('/')
    .filter(segment => segment.length > 0)
    .slice(baseSegments); // Skip base segments
    
  return segments;
}

/**
 * Build a relative URL from segments
 */
export function buildPath(...segments: string[]): string {
  const cleanSegments = segments
    .filter(segment => segment && segment.length > 0)
    .map(segment => segment.replace(/^\/|\/$/g, '')); // Remove leading/trailing slashes
    
  return '/' + cleanSegments.join('/');
}

/**
 * Add query parameters to a URL
 */
export function addQueryParams(url: string, params: Record<string, string | number | boolean>): string {
  const urlObj = new URL(url, 'https://example.com'); // Base URL for relative paths
  
  Object.entries(params).forEach(([key, value]) => {
    if (value !== null && value !== undefined) {
      urlObj.searchParams.set(key, String(value));
    }
  });
  
  return urlObj.pathname + urlObj.search;
}