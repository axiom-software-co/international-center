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

// Business Inquiry domain composables
export {
  useBusinessInquiry,
  useBusinessInquirySubmission,
} from '../lib/clients/composables/useBusinessInquiry';

export type {
  UseBusinessInquiryResult,
  UseBusinessInquirySubmissionResult,
} from '../lib/clients/composables/useBusinessInquiry';

// Donations Inquiry domain composables
export {
  useDonationsInquiry,
  useDonationsInquirySubmission,
} from '../lib/clients/composables/useDonationsInquiry';

export type {
  UseDonationsInquiryResult,
  UseDonationsInquirySubmissionResult,
} from '../lib/clients/composables/useDonationsInquiry';

// Media Inquiry domain composables
export {
  useMediaInquiry,
  useMediaInquirySubmission,
} from '../lib/clients/composables/useMediaInquiry';

export type {
  UseMediaInquiryResult,
  UseMediaInquirySubmissionResult,
} from '../lib/clients/composables/useMediaInquiry';