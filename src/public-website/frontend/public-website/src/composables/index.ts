// Vue 3 Composables - Contract-based exports for all domain operations
// Modern Vue Composition API patterns with contract-generated type safety

// Contract-based content domain composables
export {
  useContractNews,
  useContractResearch,
  useContractServices, 
  useContractEvents,
  useContractHealth,
  useContractInquiries,
  useContractApi
} from './useContractApi';

// Legacy hook compatibility (migrated to use contract clients internally)
export {
  useNews,
  useNewsArticle,
  useFeaturedNews,
  useRecentNews,
  useNewsSearch
} from '../hooks/useNews';

export {
  useResearch
} from '../hooks/useResearch';

// Legacy composable aliases for research
export const useResearchArticles = () => {
  return useResearch({ enabled: true });
};

export const useFeaturedResearch = () => {
  const researchComposable = useContractResearch();
  return {
    articles: researchComposable.research,
    loading: researchComposable.loading,
    error: researchComposable.error,
    refetch: () => researchComposable.fetchFeaturedResearch()
  };
};

// Legacy composable aliases for events
export const useEvents = () => {
  const eventsComposable = useContractEvents();
  return {
    events: eventsComposable.events,
    loading: eventsComposable.loading,
    error: eventsComposable.error,
    refetch: () => eventsComposable.fetchEvents()
  };
};

export const useFeaturedEvents = () => {
  const eventsComposable = useContractEvents();
  return {
    events: eventsComposable.events,
    loading: eventsComposable.loading,
    error: eventsComposable.error,
    refetch: () => eventsComposable.fetchFeaturedEvents()
  };
};

// Legacy composable aliases for services
export const useServices = () => {
  const servicesComposable = useContractServices();
  return {
    services: servicesComposable.services,
    loading: servicesComposable.loading,
    error: servicesComposable.error,
    refetch: () => servicesComposable.fetchServices()
  };
};

export const useFeaturedServices = () => {
  const servicesComposable = useContractServices();
  return {
    services: servicesComposable.services,
    loading: servicesComposable.loading,
    error: servicesComposable.error,
    refetch: () => servicesComposable.fetchFeaturedServices()
  };
};

// Legacy compatibility exports for smooth migration
export const useBusinessInquiry = () => {
  const inquiryComposable = useContractInquiries();
  return {
    ...inquiryComposable,
    submitInquiry: inquiryComposable.submitBusinessInquiry
  };
};

export const useBusinessInquirySubmission = () => {
  return useBusinessInquiry();
};

export const useDonationsInquiry = () => {
  const inquiryComposable = useContractInquiries();
  return {
    ...inquiryComposable,
    submitInquiry: inquiryComposable.submitBusinessInquiry // Donations use business inquiry endpoint
  };
};

export const useDonationsInquirySubmission = () => {
  return useDonationsInquiry();
};

export const useMediaInquiry = () => {
  const inquiryComposable = useContractInquiries();
  return {
    ...inquiryComposable,
    submitInquiry: inquiryComposable.submitMediaInquiry
  };
};

export const useMediaInquirySubmission = () => {
  return useMediaInquiry();
};

export const useVolunteerInquiry = () => {
  const inquiryComposable = useContractInquiries();
  return {
    ...inquiryComposable,
    submitInquiry: inquiryComposable.submitMediaInquiry // Volunteers use media inquiry endpoint for now
  };
};

export const useVolunteerInquirySubmission = () => {
  return useVolunteerInquiry();
};

// Import contract composables
import {
  useContractNews,
  useContractResearch,
  useContractServices, 
  useContractEvents,
  useContractHealth,
  useContractInquiries,
  useContractApi
} from './useContractApi';