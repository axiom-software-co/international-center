// Domain Navigation Configuration
// Centralizes cross-domain navigation patterns and breadcrumb data

export interface DomainConfig {
  domain: string;
  parentPath: string;
  parentLabel: string;
  baseUrl: string;
  singularName: string;
  pluralName: string;
}

export const DOMAIN_CONFIGS: Record<string, DomainConfig> = {
  events: {
    domain: 'events',
    parentPath: '/community/events',
    parentLabel: 'Events',
    baseUrl: '/community/events',
    singularName: 'event',
    pluralName: 'events',
  },
  news: {
    domain: 'news',
    parentPath: '/company/news',
    parentLabel: 'News & Insights',
    baseUrl: '/company/news',
    singularName: 'article',
    pluralName: 'news',
  },
  research: {
    domain: 'research',
    parentPath: '/community/research',
    parentLabel: 'Research & Innovation',
    baseUrl: '/community/research',
    singularName: 'article',
    pluralName: 'research',
  },
  services: {
    domain: 'services',
    parentPath: '/services',
    parentLabel: 'Services',
    baseUrl: '/services',
    singularName: 'service',
    pluralName: 'services',
  },
};

/**
 * Get domain configuration by domain name
 */
export function getDomainConfig(domain: string): DomainConfig {
  const config = DOMAIN_CONFIGS[domain];
  if (!config) {
    throw new Error(`Unknown domain: ${domain}`);
  }
  return config;
}

/**
 * Generate breadcrumb props for a domain item
 */
export function generateBreadcrumbProps(
  domain: string,
  itemName: string,
  title: string,
  category?: string
) {
  const config = getDomainConfig(domain);
  
  return {
    parentPath: config.parentPath,
    parentLabel: config.parentLabel,
    itemName,
    title,
    category,
  };
}

/**
 * Generate item URL for a domain
 */
export function generateItemUrl(domain: string, slug: string): string {
  const config = getDomainConfig(domain);
  return `${config.baseUrl}/${slug}`;
}

/**
 * Generate category URL for a domain
 */
export function generateCategoryUrl(domain: string, categorySlug: string): string {
  const config = getDomainConfig(domain);
  return `${config.baseUrl}/category/${categorySlug}`;
}

/**
 * Generate search URL for a domain
 */
export function generateSearchUrl(domain: string, query?: string): string {
  const config = getDomainConfig(domain);
  const searchPath = `${config.baseUrl}/search`;
  return query ? `${searchPath}?q=${encodeURIComponent(query)}` : searchPath;
}