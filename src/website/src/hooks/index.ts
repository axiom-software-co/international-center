// Vue 3 Composables - Clean exports (Migrated from React to Vue Composition API)
// Domain-driven composables for component integration

// Services composables (REST-enabled through Public Gateway)
export {
  useServices,
  useService,
  useFeaturedServices,
  useServiceCategories,
  useSearchServices,
} from '../composables/useServices';
export type {
  UseServicesResult,
  UseServicesOptions,
  UseServiceResult,
  UseFeaturedServicesResult,
  UseServiceCategoriesResult,
  UseSearchServicesResult,
} from '../composables/useServices';

// ====================================================================================
// APIS ON HOLD - These composables will be migrated when development resumes
// ====================================================================================

// News composables (ON HOLD - React hooks preserved but unused)
// export { useNews, useNewsArticle, useNewsSearch, useFeaturedNews, useRecentNews } from './useNews';

// Research composables (ON HOLD - React hooks preserved but unused)
// export { useResearch, useResearchArticle, useFeaturedResearch, useRecentResearch } from './useResearch';

// Contacts composables (REMOVED - React hooks eliminated in favor of Vue inquiry composables)
