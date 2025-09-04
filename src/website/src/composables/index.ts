// Vue 3 Composables - Clean exports for all domain composables
// Modern Vue Composition API patterns for reactive data management

// Services domain composables
export {
  useServices,
  useService,
  useFeaturedServices,
  useServiceCategories,
  useSearchServices,
} from './useServices';

export type {
  UseServicesResult,
  UseServicesOptions,
  UseServiceResult,
  UseFeaturedServicesResult,
  UseServiceCategoriesResult,
  UseSearchServicesResult,
} from './useServices';

// Events domain composables
export {
  useEvents,
  useEvent,
  useFeaturedEvent,
  useFeaturedEvents,
  useSearchEvents,
} from './useEvents';

export type {
  UseEventsResult,
  UseEventsOptions,
  UseEventResult,
  UseFeaturedEventResult,
  UseSearchEventsResult,
} from './useEvents';

// Research domain composables
export {
  useResearchArticles,
  useResearchArticle,
  useFeaturedResearch,
  useFeaturedResearchArticles,
  useSearchResearch,
} from './useResearch';

export type {
  UseResearchArticlesResult,
  UseResearchArticlesOptions,
  UseResearchArticleResult,
  UseFeaturedResearchResult,
  UseSearchResearchResult,
} from './useResearch';

// News domain composables
export {
  useNews,
  useNewsArticle,
  useFeaturedNews,
  useSearchNews,
  useNewsCategories,
} from './useNews';

export type {
  UseNewsResult,
  UseNewsOptions,
  UseNewsArticleResult,
  UseFeaturedNewsResult,
  UseSearchNewsResult,
  UseNewsCategoriesResult,
} from './useNews';