// Contract-compliant Vue composables using generated TypeScript clients
import { ref, computed } from 'vue'
import { apiClient } from '@/lib/api-client'
import { ContractErrorHandler } from '@/lib/error-handling'
import type { 
  NewsArticle, 
  ResearchPublication, 
  Service, 
  Event,
  NewsCategory,
  ResearchCategory,
  ServiceCategory,
  EventCategory
} from '@international-center/public-api-client'

// Generic composable for any contract API call
export function useContractApi<T>(context: string = 'api') {
  const data = ref<T | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  const errorHandler = ContractErrorHandler.createErrorComposable(context)

  const execute = async (apiCall: () => Promise<T>) => {
    loading.value = true
    error.value = null

    try {
      const result = await apiCall()
      data.value = result
      return result
    } catch (err) {
      const errorMessage = errorHandler.handleApiError(err)
      error.value = errorMessage
      throw err
    } finally {
      loading.value = false
    }
  }

  const retry = async (apiCall: () => Promise<T>) => {
    if (error.value && errorHandler.isRetryableError(error.value)) {
      await execute(apiCall)
    }
  }

  const clear = () => {
    data.value = null
    error.value = null
  }

  return {
    data,
    loading,
    error,
    isLoading: computed(() => loading.value),
    hasError: computed(() => error.value !== null),
    hasData: computed(() => data.value !== null),
    execute,
    retry,
    clear
  }
}

// News-specific composable
export function useContractNews() {
  const newsApi = useContractApi<NewsArticle[]>('news')
  const singleNewsApi = useContractApi<NewsArticle>('news-article')
  const categoriesApi = useContractApi<NewsCategory[]>('news-categories')

  const fetchNews = async (params?: { page?: number; limit?: number; search?: string; categoryId?: string }) => {
    return await newsApi.execute(() => apiClient.getNews(params).then(r => r.data || []))
  }

  const fetchNewsById = async (id: string) => {
    return await singleNewsApi.execute(() => apiClient.getNewsById(id).then(r => r.data!))
  }

  const fetchFeaturedNews = async () => {
    return await newsApi.execute(() => apiClient.getFeaturedNews().then(r => r.data || []))
  }

  const fetchNewsCategories = async () => {
    return await categoriesApi.execute(() => apiClient.getNewsCategories().then(r => r.data || []))
  }

  return {
    news: newsApi.data,
    singleNews: singleNewsApi.data,
    categories: categoriesApi.data,
    loading: computed(() => newsApi.loading.value || singleNewsApi.loading.value || categoriesApi.loading.value),
    error: computed(() => newsApi.error.value || singleNewsApi.error.value || categoriesApi.error.value),
    fetchNews,
    fetchNewsById,
    fetchFeaturedNews,
    fetchNewsCategories,
    clearNewsError: newsApi.clear,
    clearSingleNewsError: singleNewsApi.clear,
    clearCategoriesError: categoriesApi.clear
  }
}

// Research-specific composable
export function useContractResearch() {
  const researchApi = useContractApi<ResearchPublication[]>('research')
  const singleResearchApi = useContractApi<ResearchPublication>('research-publication')
  const categoriesApi = useContractApi<ResearchCategory[]>('research-categories')

  const fetchResearch = async (params?: { page?: number; limit?: number; search?: string; categoryId?: string }) => {
    return await researchApi.execute(() => apiClient.getResearch(params).then(r => r.data || []))
  }

  const fetchResearchById = async (id: string) => {
    return await singleResearchApi.execute(() => apiClient.getResearchById(id).then(r => r.data!))
  }

  const fetchFeaturedResearch = async () => {
    return await researchApi.execute(() => apiClient.getFeaturedResearch().then(r => r.data || []))
  }

  const fetchResearchCategories = async () => {
    return await categoriesApi.execute(() => apiClient.getResearchCategories().then(r => r.data || []))
  }

  return {
    research: researchApi.data,
    singleResearch: singleResearchApi.data,
    categories: categoriesApi.data,
    loading: computed(() => researchApi.loading.value || singleResearchApi.loading.value || categoriesApi.loading.value),
    error: computed(() => researchApi.error.value || singleResearchApi.error.value || categoriesApi.error.value),
    fetchResearch,
    fetchResearchById,
    fetchFeaturedResearch,
    fetchResearchCategories,
    clearResearchError: researchApi.clear,
    clearSingleResearchError: singleResearchApi.clear,
    clearCategoriesError: categoriesApi.clear
  }
}

// Services-specific composable
export function useContractServices() {
  const servicesApi = useContractApi<Service[]>('services')
  const singleServiceApi = useContractApi<Service>('service')
  const categoriesApi = useContractApi<ServiceCategory[]>('service-categories')

  const fetchServices = async (params?: { page?: number; limit?: number; search?: string; categoryId?: string }) => {
    return await servicesApi.execute(() => apiClient.getServices(params).then(r => r.data || []))
  }

  const fetchServiceById = async (id: string) => {
    return await singleServiceApi.execute(() => apiClient.getServiceById(id).then(r => r.data!))
  }

  const fetchFeaturedServices = async () => {
    return await servicesApi.execute(() => apiClient.getFeaturedServices().then(r => r.data || []))
  }

  const fetchServiceCategories = async () => {
    return await categoriesApi.execute(() => apiClient.getServiceCategories().then(r => r.data || []))
  }

  return {
    services: servicesApi.data,
    singleService: singleServiceApi.data,
    categories: categoriesApi.data,
    loading: computed(() => servicesApi.loading.value || singleServiceApi.loading.value || categoriesApi.loading.value),
    error: computed(() => servicesApi.error.value || singleServiceApi.error.value || categoriesApi.error.value),
    fetchServices,
    fetchServiceById,
    fetchFeaturedServices,
    fetchServiceCategories,
    clearServicesError: servicesApi.clear,
    clearSingleServiceError: singleServiceApi.clear,
    clearCategoriesError: categoriesApi.clear
  }
}

// Events-specific composable
export function useContractEvents() {
  const eventsApi = useContractApi<Event[]>('events')
  const singleEventApi = useContractApi<Event>('event')
  const categoriesApi = useContractApi<EventCategory[]>('event-categories')

  const fetchEvents = async (params?: { page?: number; limit?: number; search?: string; categoryId?: string }) => {
    return await eventsApi.execute(() => apiClient.getEvents(params).then(r => r.data || []))
  }

  const fetchEventById = async (id: string) => {
    return await singleEventApi.execute(() => apiClient.getEventById(id).then(r => r.data!))
  }

  const fetchFeaturedEvents = async () => {
    return await eventsApi.execute(() => apiClient.getFeaturedEvents().then(r => r.data || []))
  }

  const fetchEventCategories = async () => {
    return await categoriesApi.execute(() => apiClient.getEventCategories().then(r => r.data || []))
  }

  return {
    events: eventsApi.data,
    singleEvent: singleEventApi.data,
    categories: categoriesApi.data,
    loading: computed(() => eventsApi.loading.value || singleEventApi.loading.value || categoriesApi.loading.value),
    error: computed(() => eventsApi.error.value || singleEventApi.error.value || categoriesApi.error.value),
    fetchEvents,
    fetchEventById,
    fetchFeaturedEvents,
    fetchEventCategories,
    clearEventsError: eventsApi.clear,
    clearSingleEventError: singleEventApi.clear,
    clearCategoriesError: categoriesApi.clear
  }
}

// Health check composable
export function useContractHealth() {
  const healthApi = useContractApi<any>('health')

  const checkHealth = async () => {
    return await healthApi.execute(() => apiClient.getHealth())
  }

  return {
    health: healthApi.data,
    loading: healthApi.loading,
    error: healthApi.error,
    isHealthy: computed(() => healthApi.data.value?.status === 'healthy'),
    checkHealth,
    clearError: healthApi.clear
  }
}

// Inquiry submission composable
export function useContractInquiries() {
  const inquiryApi = useContractApi<any>('inquiry')
  const success = ref(false)

  const submitMediaInquiry = async (inquiry: any) => {
    try {
      const result = await inquiryApi.execute(() => apiClient.submitMediaInquiry(inquiry))
      success.value = true
      return result
    } catch (err) {
      success.value = false
      throw err
    }
  }

  const submitBusinessInquiry = async (inquiry: any) => {
    try {
      const result = await inquiryApi.execute(() => apiClient.submitBusinessInquiry(inquiry))
      success.value = true
      return result
    } catch (err) {
      success.value = false
      throw err
    }
  }

  const submitInquiry = async (inquiry: any) => {
    // Default to volunteer inquiry (media inquiry for now)
    return await submitMediaInquiry(inquiry)
  }

  const reset = () => {
    success.value = false
    inquiryApi.clear()
  }

  return {
    inquiryResponse: inquiryApi.data,
    loading: inquiryApi.loading,
    error: inquiryApi.error,
    isSubmitting: computed(() => inquiryApi.loading.value),
    isSuccess: computed(() => success.value),
    isError: computed(() => inquiryApi.error.value !== null),
    response: inquiryApi.data,
    submitMediaInquiry,
    submitBusinessInquiry,
    submitInquiry,
    clearError: inquiryApi.clear,
    reset
  }
}