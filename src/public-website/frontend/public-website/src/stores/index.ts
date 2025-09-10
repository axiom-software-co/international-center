// Store index file - centralizes all contract-based Pinia stores
export { useNewsStore } from './news';
export { useResearchStore } from './research';
export { useServicesStore } from './services';
export { useEventsStore } from './events';

// Re-export store types for convenience
export type * from './interfaces';