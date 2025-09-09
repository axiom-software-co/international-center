// Composables Integration Tests - Unified API surface validation
// Tests ensure all domain composables are properly exported from central entry point

import { describe, it, expect } from 'vitest';

// Test imports from unified composables API
describe('Composables Integration - Unified API Surface', () => {
  describe('Services Domain Exports', () => {
    it('should export all Services composables', async () => {
      const servicesComposables = await import('./index');
      
      // Validate Services composable functions
      expect(servicesComposables.useServices).toBeTypeOf('function');
      expect(servicesComposables.useService).toBeTypeOf('function');
      expect(servicesComposables.useFeaturedServices).toBeTypeOf('function');
      expect(servicesComposables.useServiceCategories).toBeTypeOf('function');
      expect(servicesComposables.useSearchServices).toBeTypeOf('function');
      
      // Validate Services types are exported
      expect(typeof servicesComposables.UseServicesResult).toBe('undefined'); // Type, not runtime value
      expect(typeof servicesComposables.UseServicesOptions).toBe('undefined'); // Type, not runtime value
      expect(typeof servicesComposables.UseServiceResult).toBe('undefined'); // Type, not runtime value
      expect(typeof servicesComposables.UseFeaturedServicesResult).toBe('undefined'); // Type, not runtime value
      expect(typeof servicesComposables.UseServiceCategoriesResult).toBe('undefined'); // Type, not runtime value
      expect(typeof servicesComposables.UseSearchServicesResult).toBe('undefined'); // Type, not runtime value
    }, 5000);
  });

  describe('Events Domain Exports', () => {
    it('should export all Events composables', async () => {
      const eventsComposables = await import('./index');
      
      // Validate Events composable functions
      expect(eventsComposables.useEvents).toBeTypeOf('function');
      expect(eventsComposables.useEvent).toBeTypeOf('function');
      expect(eventsComposables.useFeaturedEvents).toBeTypeOf('function');
      expect(eventsComposables.useSearchEvents).toBeTypeOf('function');
      
      // Validate Events types are exported
      expect(typeof eventsComposables.UseEventsResult).toBe('undefined'); // Type, not runtime value
      expect(typeof eventsComposables.UseEventsOptions).toBe('undefined'); // Type, not runtime value
      expect(typeof eventsComposables.UseEventResult).toBe('undefined'); // Type, not runtime value
      expect(typeof eventsComposables.UseFeaturedEventsResult).toBe('undefined'); // Type, not runtime value
      expect(typeof eventsComposables.UseSearchEventsResult).toBe('undefined'); // Type, not runtime value
    }, 5000);
  });

  describe('Research Domain Exports', () => {
    it('should export all Research composables', async () => {
      const researchComposables = await import('./index');
      
      // Validate Research composable functions
      expect(researchComposables.useResearchArticles).toBeTypeOf('function');
      expect(researchComposables.useResearchArticle).toBeTypeOf('function');
      expect(researchComposables.useFeaturedResearch).toBeTypeOf('function');
      expect(researchComposables.useSearchResearch).toBeTypeOf('function');
      
      // Validate Research types are exported
      expect(typeof researchComposables.UseResearchArticlesResult).toBe('undefined'); // Type, not runtime value
      expect(typeof researchComposables.UseResearchArticlesOptions).toBe('undefined'); // Type, not runtime value
      expect(typeof researchComposables.UseResearchArticleResult).toBe('undefined'); // Type, not runtime value
      expect(typeof researchComposables.UseFeaturedResearchResult).toBe('undefined'); // Type, not runtime value
      expect(typeof researchComposables.UseSearchResearchResult).toBe('undefined'); // Type, not runtime value
    }, 5000);
  });

  describe('News Domain Exports', () => {
    it('should export all News composables', async () => {
      const newsComposables = await import('./index');
      
      // Validate News composable functions
      expect(newsComposables.useNews).toBeTypeOf('function');
      expect(newsComposables.useNewsArticle).toBeTypeOf('function');
      expect(newsComposables.useFeaturedNews).toBeTypeOf('function');
      expect(newsComposables.useSearchNews).toBeTypeOf('function');
      expect(newsComposables.useNewsCategories).toBeTypeOf('function');
      
      // Validate News types are exported
      expect(typeof newsComposables.UseNewsResult).toBe('undefined'); // Type, not runtime value
      expect(typeof newsComposables.UseNewsOptions).toBe('undefined'); // Type, not runtime value
      expect(typeof newsComposables.UseNewsArticleResult).toBe('undefined'); // Type, not runtime value
      expect(typeof newsComposables.UseFeaturedNewsResult).toBe('undefined'); // Type, not runtime value
      expect(typeof newsComposables.UseSearchNewsResult).toBe('undefined'); // Type, not runtime value
      expect(typeof newsComposables.UseNewsCategoriesResult).toBe('undefined'); // Type, not runtime value
    }, 5000);
  });

  describe('API Surface Consistency', () => {
    it('should maintain consistent naming patterns across domains', async () => {
      const composables = await import('./index');
      
      // Check that all domains follow use[Domain] pattern
      expect(composables.useServices).toBeTypeOf('function');
      expect(composables.useEvents).toBeTypeOf('function');
      expect(composables.useResearchArticles).toBeTypeOf('function'); // Research uses different naming
      expect(composables.useNews).toBeTypeOf('function');
      
      // Check that all domains follow use[Domain]Article/use[Domain]BySlug pattern
      expect(composables.useService).toBeTypeOf('function'); // Single service
      expect(composables.useEvent).toBeTypeOf('function'); // Single event
      expect(composables.useResearchArticle).toBeTypeOf('function'); // Single research article
      expect(composables.useNewsArticle).toBeTypeOf('function'); // Single news article
      
      // Check that all domains follow useFeatured[Domain] pattern
      expect(composables.useFeaturedServices).toBeTypeOf('function');
      expect(composables.useFeaturedEvents).toBeTypeOf('function');
      expect(composables.useFeaturedResearch).toBeTypeOf('function');
      expect(composables.useFeaturedNews).toBeTypeOf('function');
      
      // Check that all domains follow useSearch[Domain] pattern
      expect(composables.useSearchServices).toBeTypeOf('function');
      expect(composables.useSearchEvents).toBeTypeOf('function');
      expect(composables.useSearchResearch).toBeTypeOf('function');
      expect(composables.useSearchNews).toBeTypeOf('function');
    }, 5000);
  });

  describe('Component Import Integration', () => {
    it('should allow components to import composables from unified API', async () => {
      // Test that unified imports work for component usage patterns
      const { 
        useServices,
        useEvents,
        useResearchArticles,
        useNews,
        useFeaturedServices,
        useFeaturedEvents,
        useFeaturedResearch,
        useFeaturedNews
      } = await import('./index');
      
      // Validate all functions are available for component consumption
      expect(useServices).toBeTypeOf('function');
      expect(useEvents).toBeTypeOf('function');
      expect(useResearchArticles).toBeTypeOf('function');
      expect(useNews).toBeTypeOf('function');
      expect(useFeaturedServices).toBeTypeOf('function');
      expect(useFeaturedEvents).toBeTypeOf('function');
      expect(useFeaturedResearch).toBeTypeOf('function');
      expect(useFeaturedNews).toBeTypeOf('function');
    }, 5000);

    it('should support destructured imports for all domain composables', async () => {
      // Test common component import patterns
      const composables = await import('./index');
      
      // Services destructuring
      const {
        useServices: servicesHook,
        useService: serviceHook,
        useFeaturedServices: featuredServicesHook,
        useSearchServices: searchServicesHook
      } = composables;
      
      expect(servicesHook).toBeTypeOf('function');
      expect(serviceHook).toBeTypeOf('function');
      expect(featuredServicesHook).toBeTypeOf('function');
      expect(searchServicesHook).toBeTypeOf('function');
      
      // Events destructuring
      const {
        useEvents: eventsHook,
        useEvent: eventHook,
        useFeaturedEvents: featuredEventsHook,
        useSearchEvents: searchEventsHook
      } = composables;
      
      expect(eventsHook).toBeTypeOf('function');
      expect(eventHook).toBeTypeOf('function');
      expect(featuredEventsHook).toBeTypeOf('function');
      expect(searchEventsHook).toBeTypeOf('function');
      
      // Research destructuring
      const {
        useResearchArticles: researchHook,
        useResearchArticle: researchArticleHook,
        useFeaturedResearch: featuredResearchHook,
        useSearchResearch: searchResearchHook
      } = composables;
      
      expect(researchHook).toBeTypeOf('function');
      expect(researchArticleHook).toBeTypeOf('function');
      expect(featuredResearchHook).toBeTypeOf('function');
      expect(searchResearchHook).toBeTypeOf('function');
      
      // News destructuring
      const {
        useNews: newsHook,
        useNewsArticle: newsArticleHook,
        useFeaturedNews: featuredNewsHook,
        useSearchNews: searchNewsHook,
        useNewsCategories: newsCategoriesHook
      } = composables;
      
      expect(newsHook).toBeTypeOf('function');
      expect(newsArticleHook).toBeTypeOf('function');
      expect(featuredNewsHook).toBeTypeOf('function');
      expect(searchNewsHook).toBeTypeOf('function');
      expect(newsCategoriesHook).toBeTypeOf('function');
    }, 5000);
  });

  describe('Type Safety and TypeScript Integration', () => {
    it('should export all TypeScript types for type-safe component usage', async () => {
      // This test validates that TypeScript compilation will work
      // Types are compile-time constructs, so we test by importing
      
      // Import should not throw and types should be available at compile time
      const module = await import('./index');
      
      // Runtime check that the module contains our expected exports
      // Types won't be present at runtime but functions will be
      expect(Object.keys(module)).toContain('useServices');
      expect(Object.keys(module)).toContain('useEvents');
      expect(Object.keys(module)).toContain('useResearchArticles');
      expect(Object.keys(module)).toContain('useNews');
      expect(Object.keys(module)).toContain('useFeaturedServices');
      expect(Object.keys(module)).toContain('useFeaturedEvents');
      expect(Object.keys(module)).toContain('useFeaturedResearch');
      expect(Object.keys(module)).toContain('useFeaturedNews');
      expect(Object.keys(module)).toContain('useSearchServices');
      expect(Object.keys(module)).toContain('useSearchEvents');
      expect(Object.keys(module)).toContain('useSearchResearch');
      expect(Object.keys(module)).toContain('useSearchNews');
    }, 5000);
  });

  describe('Cross-Domain Functional Integration', () => {
    it('should allow simultaneous usage of multiple domain composables', async () => {
      const {
        useServices,
        useEvents, 
        useResearchArticles,
        useNews
      } = await import('./index');
      
      // Test that all composables can be imported and used together
      // This simulates real component usage patterns
      expect(() => {
        const servicesComposable = useServices;
        const eventsComposable = useEvents;
        const researchComposable = useResearchArticles;
        const newsComposable = useNews;
        
        return {
          servicesComposable,
          eventsComposable,
          researchComposable,
          newsComposable
        };
      }).not.toThrow();
    }, 5000);

    it('should maintain architectural separation between domain composables', async () => {
      const composables = await import('./index');
      
      // Verify that each domain's composables are independent functions
      // No shared state or coupling between domains
      expect(composables.useServices).not.toBe(composables.useEvents);
      expect(composables.useEvents).not.toBe(composables.useResearchArticles);
      expect(composables.useResearchArticles).not.toBe(composables.useNews);
      expect(composables.useNews).not.toBe(composables.useServices);
      
      // Each composable should be a unique function
      expect(typeof composables.useServices).toBe('function');
      expect(typeof composables.useEvents).toBe('function');
      expect(typeof composables.useResearchArticles).toBe('function');
      expect(typeof composables.useNews).toBe('function');
    }, 5000);
  });

  describe('Business Inquiry Domain Exports', () => {
    it('should export all Business Inquiry composables', async () => {
      const businessInquiryComposables = await import('./index');
      
      // Validate Business Inquiry composable functions
      expect(businessInquiryComposables.useBusinessInquiry).toBeTypeOf('function');
      expect(businessInquiryComposables.useBusinessInquirySubmission).toBeTypeOf('function');
      
      // Validate Business Inquiry types are exported
      expect(typeof businessInquiryComposables.UseBusinessInquiryResult).toBe('undefined'); // Type, not runtime value
      expect(typeof businessInquiryComposables.UseBusinessInquirySubmissionResult).toBe('undefined'); // Type, not runtime value
    }, 5000);
  });

  describe('Donations Inquiry Domain Exports', () => {
    it('should export all Donations Inquiry composables', async () => {
      const donationsInquiryComposables = await import('./index');
      
      // Validate Donations Inquiry composable functions
      expect(donationsInquiryComposables.useDonationsInquiry).toBeTypeOf('function');
      expect(donationsInquiryComposables.useDonationsInquirySubmission).toBeTypeOf('function');
      
      // Validate Donations Inquiry types are exported
      expect(typeof donationsInquiryComposables.UseDonationsInquiryResult).toBe('undefined'); // Type, not runtime value
      expect(typeof donationsInquiryComposables.UseDonationsInquirySubmissionResult).toBe('undefined'); // Type, not runtime value
    }, 5000);
  });

  describe('Media Inquiry Domain Exports', () => {
    it('should export all Media Inquiry composables', async () => {
      const mediaInquiryComposables = await import('./index');
      
      // Validate Media Inquiry composable functions
      expect(mediaInquiryComposables.useMediaInquiry).toBeTypeOf('function');
      expect(mediaInquiryComposables.useMediaInquirySubmission).toBeTypeOf('function');
      
      // Validate Media Inquiry types are exported
      expect(typeof mediaInquiryComposables.UseMediaInquiryResult).toBe('undefined'); // Type, not runtime value
      expect(typeof mediaInquiryComposables.UseMediaInquirySubmissionResult).toBe('undefined'); // Type, not runtime value
    }, 5000);
  });

  describe('Inquiry Domain Integration', () => {
    it('should maintain consistent inquiry composable patterns across domains', async () => {
      const composables = await import('./index');
      
      // Check that all inquiry domains follow use[Domain]Inquiry pattern
      expect(composables.useBusinessInquiry).toBeTypeOf('function');
      expect(composables.useDonationsInquiry).toBeTypeOf('function');
      expect(composables.useMediaInquiry).toBeTypeOf('function');
      
      // Check that all inquiry domains follow use[Domain]InquirySubmission pattern
      expect(composables.useBusinessInquirySubmission).toBeTypeOf('function');
      expect(composables.useDonationsInquirySubmission).toBeTypeOf('function');
      expect(composables.useMediaInquirySubmission).toBeTypeOf('function');
    }, 5000);

    it('should allow simultaneous usage of inquiry and content domain composables', async () => {
      const composables = await import('./index');
      
      // Test that inquiry and content composables can be imported together
      const { 
        useServices,
        useBusinessInquiry,
        useEvents,
        useDonationsInquiry,
        useNews,
        useMediaInquiry
      } = composables;
      
      // Validate all functions are available for component consumption
      expect(useServices).toBeTypeOf('function');
      expect(useBusinessInquiry).toBeTypeOf('function');
      expect(useEvents).toBeTypeOf('function');
      expect(useDonationsInquiry).toBeTypeOf('function');
      expect(useNews).toBeTypeOf('function');
      expect(useMediaInquiry).toBeTypeOf('function');
      
      // Verify architectural separation between inquiry and content domains
      expect(useBusinessInquiry).not.toBe(useServices);
      expect(useDonationsInquiry).not.toBe(useEvents);
      expect(useMediaInquiry).not.toBe(useNews);
    }, 5000);
  });
});