// Business Services - Main exports
// Clean exports for all business logic services

// Content Service
export { ContentService, contentService } from './ContentService';
export type { ContentItem, FeaturedContent, SearchResults } from './ContentService';

// Navigation Service
export { NavigationService, navigationService } from './NavigationService';
export type { NavigationItem, NavigationMenu } from './NavigationService';

// Search Service
export { SearchService, searchService } from './SearchService';
export type { SearchResult, SearchResponse, SearchOptions } from './SearchService';

// Metadata Service
export { MetadataService, metadataService } from './MetadataService';
export type { PageMetadata, SiteMetadata } from './MetadataService';
