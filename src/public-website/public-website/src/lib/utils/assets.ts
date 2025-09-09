// Asset URL utilities for International Center Platform
// Handles asset resolution between external URLs and local CDN

import { config } from '../../environments';

/**
 * Resolves asset URLs to use local CDN when available in development
 */
export function resolveAssetUrl(imageUrl: string | null | undefined): string | null {
  if (!imageUrl) {
    return null;
  }

  // If already a local CDN URL, return as-is
  if (imageUrl.startsWith(config.domains.assets.baseUrl)) {
    return imageUrl;
  }

  // If it's a placehold.co URL, try to map to local equivalent
  if (imageUrl.includes('placehold.co')) {
    const localAssetUrl = mapPlaceholderToLocal(imageUrl);
    if (localAssetUrl) {
      return localAssetUrl;
    }
  }

  // For external URLs in development, prefer local CDN if available
  if (config.environment === 'local') {
    const localAssetUrl = tryMapToLocalAsset(imageUrl);
    if (localAssetUrl) {
      return localAssetUrl;
    }
  }

  // Return original URL as fallback
  return imageUrl;
}

/**
 * Maps placehold.co URLs to local assets if available
 */
function mapPlaceholderToLocal(placeholderUrl: string): string | null {
  // Extract text from placehold.co URL
  const textMatch = placeholderUrl.match(/text=([^&]+)/);
  if (!textMatch) return null;

  const text = decodeURIComponent(textMatch[1]);
  const slug = text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');

  // Map to local asset based on URL patterns
  if (placeholderUrl.includes('1200x675') || placeholderUrl.includes('800x450')) {
    // Hero images
    return (
      tryLocalAssetPath(`services/${slug}-hero.png`) ||
      tryLocalAssetPath(`news/${slug}-hero.png`) ||
      tryLocalAssetPath(`research/${slug}-hero.png`)
    );
  } else if (placeholderUrl.includes('400x300') || placeholderUrl.includes('300x200')) {
    // Thumbnail images
    return (
      tryLocalAssetPath(`services/${slug}-thumb.png`) ||
      tryLocalAssetPath(`news/${slug}-thumb.png`) ||
      tryLocalAssetPath(`research/${slug}-thumb.png`)
    );
  }

  return null;
}

/**
 * Attempts to map external URLs to local assets based on content
 */
function tryMapToLocalAsset(externalUrl: string): string | null {
  // This is a simple heuristic - in production you might want more sophisticated mapping
  const urlLower = externalUrl.toLowerCase();

  // Check for service-related keywords
  if (urlLower.includes('prp') || urlLower.includes('therapy')) {
    return tryLocalAssetPath('services/prp-therapy-hero.png');
  }

  if (urlLower.includes('exosome')) {
    return tryLocalAssetPath('services/exosome-therapy-hero.png');
  }

  if (urlLower.includes('stem-cell') || urlLower.includes('stemcell')) {
    return tryLocalAssetPath('services/stem-cell-therapies-hero.png');
  }

  if (urlLower.includes('peptide')) {
    return tryLocalAssetPath('services/peptide-therapy-hero.png');
  }

  if (urlLower.includes('news') || urlLower.includes('article')) {
    return tryLocalAssetPath('news/regenerative-medicine-expansion-hero.png');
  }

  if (urlLower.includes('research') || urlLower.includes('study')) {
    return tryLocalAssetPath('research/prp-sports-medicine-study-hero.png');
  }

  return null;
}

/**
 * Tries to construct a local asset path and returns it if it would be valid
 */
function tryLocalAssetPath(relativePath: string): string | null {
  // In development, we assume the local CDN is available
  // In production, you might want to check if the asset exists
  if (config.environment === 'local') {
    return `${config.domains.assets.baseUrl}/assets/images/${relativePath}`;
  }
  return null;
}

/**
 * Gets a fallback image URL for a given content type
 */
export function getFallbackImageUrl(
  contentType: 'service' | 'news' | 'research' | 'general' = 'general'
): string {
  const fallbacks = {
    service: `${config.domains.assets.baseUrl}/assets/images/services/prp-therapy-hero.png`,
    news: `${config.domains.assets.baseUrl}/assets/images/news/regenerative-medicine-expansion-hero.png`,
    research: `${config.domains.assets.baseUrl}/assets/images/research/prp-sports-medicine-study-hero.png`,
    general: `${config.domains.assets.baseUrl}/assets/images/seed-data/general/content-placeholder-1.png`,
  };

  return fallbacks[contentType];
}

/**
 * Optimizes image URL for different use cases
 */
export function optimizeImageUrl(
  imageUrl: string | null | undefined,
  options: {
    width?: number;
    height?: number;
    quality?: number;
    format?: 'webp' | 'png' | 'jpg';
  } = {}
): string | null {
  const resolvedUrl = resolveAssetUrl(imageUrl);
  if (!resolvedUrl) return null;

  // For local development, return as-is since we're using static images
  if (config.environment === 'local') {
    return resolvedUrl;
  }

  // In production, you might want to add image optimization parameters
  // This is where you'd integrate with services like Cloudinary, ImageKit, etc.
  return resolvedUrl;
}

/**
 * Preloads critical images to improve performance
 */
export function preloadImage(url: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve();
    img.onerror = reject;
    img.src = url;
  });
}

/**
 * Lazy loads images with intersection observer
 */
export function setupLazyLoading(selector: string = 'img[data-lazy]'): void {
  if (typeof window === 'undefined' || !('IntersectionObserver' in window)) {
    return;
  }

  const observer = new IntersectionObserver(entries => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target as HTMLImageElement;
        const src = img.dataset.lazy;
        if (src) {
          img.src = resolveAssetUrl(src) || src;
          img.removeAttribute('data-lazy');
          observer.unobserve(img);
        }
      }
    });
  });

  document.querySelectorAll(selector).forEach(img => {
    observer.observe(img);
  });
}
