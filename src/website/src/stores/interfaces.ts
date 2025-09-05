// Store Interfaces - Unified contracts for domain stores with composable integration
// Ensures backward compatibility with existing composable interfaces

import type { Ref } from 'vue';
import type { 
  NewsArticle, 
  NewsCategory, 
  GetNewsParams, 
  SearchNewsParams 
} from '../lib/clients/news/types';
import type { 
  Service, 
  ServiceCategory, 
  GetServicesParams, 
  SearchServicesParams 
} from '../lib/clients/services/types';
import type { 
  Event, 
  EventCategory, 
  GetEventsParams, 
  SearchEventsParams 
} from '../lib/clients/events/types';
import type { 
  ResearchArticle, 
  ResearchCategory, 
  GetResearchParams, 
  SearchResearchParams 
} from '../lib/clients/research/types';

// Base Store State Interface - Common structure for all domain stores
export interface BaseStoreState {
  loading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
  searchTotal: number;
  cacheKey: string | null;
  lastCacheTime: number;
}

// Base Store Actions Interface - Common actions for all domain stores
export interface BaseStoreActions {
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
  invalidateCache: () => void;
}

// Cache Options for API calls
export interface CacheOptions {
  useCache?: boolean;
  cacheTimeout?: number; // milliseconds, default 5 minutes
}

// NEWS STORE INTERFACES
export interface NewsStoreState extends BaseStoreState {
  news: NewsArticle[];
  article: NewsArticle | null; // Individual news article state
  categories: NewsCategory[];
  featuredNews: NewsArticle[];
  searchResults: NewsArticle[];
}

export interface NewsStoreActions extends BaseStoreActions {
  setNews: (news: NewsArticle[], total: number, page: number, pageSize: number) => void;
  setCategories: (categories: NewsCategory[]) => void;
  setFeaturedNews: (news: NewsArticle[]) => void;
  setSearchResults: (results: NewsArticle[], total: number) => void;
  
  // API Actions - Mirror existing composable functionality
  fetchNews: (params?: GetNewsParams, options?: CacheOptions) => Promise<void>;
  fetchNewsArticle: (slug: string) => Promise<NewsArticle | null>;
  fetchFeaturedNews: (limit?: number) => Promise<void>;
  searchNews: (params: SearchNewsParams) => Promise<void>;
  fetchNewsCategories: () => Promise<void>;
}

export interface NewsStoreGetters {
  totalPages: number;
  hasNews: boolean;
  newsByCategory: Record<string, NewsArticle[]>;
  recentNews: NewsArticle[];
}

export interface NewsStore extends NewsStoreState, NewsStoreActions, NewsStoreGetters {}

// SERVICES STORE INTERFACES
export interface ServicesStoreState extends BaseStoreState {
  services: Service[];
  service: Service | null; // Individual service state
  categories: ServiceCategory[];
  featuredServices: Service[];
  searchResults: Service[];
}

export interface ServicesStoreActions extends BaseStoreActions {
  setServices: (services: Service[], total: number, page: number, pageSize: number) => void;
  setCategories: (categories: ServiceCategory[]) => void;
  setFeaturedServices: (services: Service[]) => void;
  setSearchResults: (results: Service[], total: number) => void;
  
  // API Actions - Mirror existing composable functionality
  fetchServices: (params?: GetServicesParams, options?: CacheOptions) => Promise<void>;
  fetchService: (slug: string) => Promise<Service | null>;
  fetchFeaturedServices: (limit?: number) => Promise<void>;
  searchServices: (params: SearchServicesParams) => Promise<void>;
  fetchServiceCategories: () => Promise<void>;
}

export interface ServicesStoreGetters {
  totalPages(): number;
  hasServices(): boolean;
  servicesByCategory(this: any): Record<string, Service[]>;
  servicesByDeliveryMode(this: any): Record<string, Service[]>;
}

export interface ServicesStore extends ServicesStoreState, ServicesStoreActions, ServicesStoreGetters {}

// EVENTS STORE INTERFACES
export interface EventsStoreState extends BaseStoreState {
  events: Event[];
  event: Event | null; // Individual event state
  categories: EventCategory[];
  featuredEvents: Event[];
  searchResults: Event[];
}

export interface EventsStoreActions extends BaseStoreActions {
  setEvents: (events: Event[], total: number, page: number, pageSize: number) => void;
  setCategories: (categories: EventCategory[]) => void;
  setFeaturedEvents: (events: Event[]) => void;
  setSearchResults: (results: Event[], total: number) => void;
  
  // API Actions - Mirror existing composable functionality
  fetchEvents: (params?: GetEventsParams, options?: CacheOptions) => Promise<void>;
  fetchEvent: (slug: string) => Promise<Event | null>;
  fetchFeaturedEvents: (limit?: number) => Promise<void>;
  searchEvents: (params: SearchEventsParams) => Promise<void>;
  fetchEventCategories: () => Promise<void>;
}

export interface EventsStoreGetters {
  totalPages: number;
  hasEvents: boolean;
  eventsByCategory: Record<string, Event[]>;
  eventsByType: Record<string, Event[]>;
  upcomingEvents: Event[];
  pastEvents: Event[];
}

export interface EventsStore extends EventsStoreState, EventsStoreActions, EventsStoreGetters {}

// RESEARCH STORE INTERFACES
export interface ResearchStoreState extends BaseStoreState {
  research: ResearchArticle[];
  article: ResearchArticle | null; // Individual research article state
  categories: ResearchCategory[];
  featuredResearch: ResearchArticle[];
  searchResults: ResearchArticle[];
}

export interface ResearchStoreActions extends BaseStoreActions {
  setResearch: (research: ResearchArticle[], total: number, page: number, pageSize: number) => void;
  setCategories: (categories: ResearchCategory[]) => void;
  setFeaturedResearch: (research: ResearchArticle[]) => void;
  setSearchResults: (results: ResearchArticle[], total: number) => void;
  
  // API Actions - Mirror existing composable functionality
  fetchResearch: (params?: GetResearchParams, options?: CacheOptions) => Promise<void>;
  fetchResearchArticle: (slug: string) => Promise<ResearchArticle | null>;
  fetchFeaturedResearch: (limit?: number) => Promise<void>;
  searchResearch: (params: SearchResearchParams) => Promise<void>;
  fetchResearchCategories: () => Promise<void>;
}

export interface ResearchStoreGetters {
  totalPages: number;
  hasResearch: boolean;
  researchByCategory: Record<string, ResearchArticle[]>;
  researchByType: Record<string, ResearchArticle[]>;
  researchByPrimaryAuthor: Record<string, ResearchArticle[]>;
  recentResearch: ResearchArticle[];
}

export interface ResearchStore extends ResearchStoreState, ResearchStoreActions, ResearchStoreGetters {}

// COMPOSABLE-STORE INTEGRATION CONTRACTS
// These interfaces ensure stores can seamlessly integrate with existing composables

export interface ComposableIntegrationContract<TItem, TParams, TSearchParams> {
  // Must expose refs that match existing composable return types
  items: Ref<TItem[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  
  // Must provide methods that match existing composable behavior
  fetch: (params?: TParams, options?: CacheOptions) => Promise<void>;
  search?: (params: TSearchParams) => Promise<void>;
  refetch: () => Promise<void>;
}

// News Composable Integration Contract
export interface NewsComposableIntegration extends ComposableIntegrationContract<NewsArticle, GetNewsParams, SearchNewsParams> {
  // Additional properties specific to news composables
  results?: Ref<NewsArticle[]>; // For search composables
  categories?: Ref<NewsCategory[]>;
  featuredNews?: Ref<NewsArticle[]>;
}

// Services Composable Integration Contract  
export interface ServicesComposableIntegration extends ComposableIntegrationContract<Service, GetServicesParams, SearchServicesParams> {
  // Additional properties specific to services composables
  results?: Ref<Service[]>; // For search composables
  categories?: Ref<ServiceCategory[]>;
  featuredServices?: Ref<Service[]>;
}

// Events Composable Integration Contract
export interface EventsComposableIntegration extends ComposableIntegrationContract<Event, GetEventsParams, SearchEventsParams> {
  // Additional properties specific to events composables
  results?: Ref<Event[]>; // For search composables
  categories?: Ref<EventCategory[]>;
  featuredEvents?: Ref<Event[]>;
  upcomingEvents?: Ref<Event[]>;
  pastEvents?: Ref<Event[]>;
}

// Research Composable Integration Contract
export interface ResearchComposableIntegration extends ComposableIntegrationContract<ResearchArticle, GetResearchParams, SearchResearchParams> {
  // Additional properties specific to research composables
  results?: Ref<ResearchArticle[]>; // For search composables
  articles?: Ref<ResearchArticle[]>; // Research uses 'articles' instead of generic 'items'
  categories?: Ref<ResearchCategory[]>;
  featuredResearch?: Ref<ResearchArticle[]>;
}

// Store Factory Contract - Ensures consistent store creation patterns
export interface StoreFactory<TStore> {
  create(): TStore;
  reset(): void;
}

// Cache Management Contract - Consistent caching behavior across stores
export interface CacheManager {
  get<T>(key: string): T | null;
  set<T>(key: string, data: T, ttl?: number): void;
  delete(key: string): void;
  clear(): void;
  isExpired(key: string): boolean;
}

// Store Plugin Contract - For Pinia plugins and middleware
export interface StorePlugin<TStore> {
  install(store: TStore): void;
  uninstall?(store: TStore): void;
}

// Validation Contracts - Ensure data integrity
export interface ValidationContract<TItem> {
  validate(item: TItem): boolean;
  sanitize?(item: TItem): TItem;
  transform?(raw: any): TItem;
}

export interface NewsValidation extends ValidationContract<NewsArticle> {}
export interface ServicesValidation extends ValidationContract<Service> {}
export interface EventsValidation extends ValidationContract<Event> {}
export interface ResearchValidation extends ValidationContract<ResearchArticle> {}

// Error Handling Contract - Consistent error management across stores
export interface ErrorHandler {
  handle(error: Error, context: string): string;
  log(error: Error, context: string): void;
  notify?(error: Error, context: string): void;
}

// Store Initialization Options
export interface StoreInitOptions {
  enableCache?: boolean;
  cacheTimeout?: number;
  enableLogging?: boolean;
  errorHandler?: ErrorHandler;
  validator?: ValidationContract<any>;
}

// Integration Test Contract - Ensures stores work with existing components
export interface IntegrationTestContract {
  testBackwardCompatibility(): boolean;
  testComposableIntegration(): boolean;
  testComponentRendering(): boolean;
  testStateManagement(): boolean;
}