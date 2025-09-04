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