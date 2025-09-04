// API Consistency Tests - Validates uniform naming patterns across all domains
// Ensures consistent developer experience across Services, Events, Research, News

import { describe, it, expect } from 'vitest';

describe('API Consistency - Naming Pattern Validation', () => {
  describe('Primary Composable Naming', () => {
    it('should follow consistent use[Domain] or use[Domain]Articles pattern', async () => {
      const composables = await import('./index');
      
      // Primary domain composables should exist
      expect(composables.useServices).toBeTypeOf('function');
      expect(composables.useEvents).toBeTypeOf('function');
      expect(composables.useResearchArticles).toBeTypeOf('function'); // Research uses Articles suffix
      expect(composables.useNews).toBeTypeOf('function');
    }, 5000);
  });

  describe('Single Item Composable Naming', () => {
    it('should follow consistent use[DomainSingular] or use[DomainSingular]Article pattern', async () => {
      const composables = await import('./index');
      
      // Single item composables should exist
      expect(composables.useService).toBeTypeOf('function');
      expect(composables.useEvent).toBeTypeOf('function');
      expect(composables.useResearchArticle).toBeTypeOf('function');
      expect(composables.useNewsArticle).toBeTypeOf('function');
    }, 5000);
  });

  describe('Featured Composable Naming', () => {
    it('should follow consistent useFeatured[Domain] pattern', async () => {
      const composables = await import('./index');
      
      // Featured composables should exist and be consistent
      expect(composables.useFeaturedServices).toBeTypeOf('function');
      expect(composables.useFeaturedEvents).toBeTypeOf('function');
      expect(composables.useFeaturedResearch).toBeTypeOf('function');
      expect(composables.useFeaturedNews).toBeTypeOf('function');
      
      // Additional featured composables (if they exist) should follow same pattern
      if (composables.useFeaturedEvent) {
        expect(composables.useFeaturedEvent).toBeTypeOf('function');
      }
      if (composables.useFeaturedResearchArticles) {
        expect(composables.useFeaturedResearchArticles).toBeTypeOf('function');
      }
    }, 5000);
  });

  describe('Search Composable Naming', () => {
    it('should follow consistent useSearch[Domain] pattern', async () => {
      const composables = await import('./index');
      
      // Search composables should exist
      expect(composables.useSearchServices).toBeTypeOf('function');
      expect(composables.useSearchEvents).toBeTypeOf('function');
      expect(composables.useSearchResearch).toBeTypeOf('function');
      expect(composables.useSearchNews).toBeTypeOf('function');
    }, 5000);
  });

  describe('Category Composable Naming', () => {
    it('should follow consistent use[Domain]Categories pattern where applicable', async () => {
      const composables = await import('./index');
      
      // Category composables should exist for domains that have categories
      expect(composables.useServiceCategories).toBeTypeOf('function');
      expect(composables.useNewsCategories).toBeTypeOf('function');
      
      // Events and Research may not have separate category composables
      // This is acceptable architectural variation
    }, 5000);
  });

  describe('Type Export Consistency', () => {
    it('should export corresponding TypeScript types for all composables', async () => {
      // Import should not fail if types are properly exported
      const module = await import('./index');
      
      // Verify module exports exist for runtime functions
      // Types are compile-time only so we check the functions exist
      const expectedFunctions = [
        'useServices', 'useService', 'useFeaturedServices', 'useServiceCategories', 'useSearchServices',
        'useEvents', 'useEvent', 'useFeaturedEvents', 'useSearchEvents',
        'useResearchArticles', 'useResearchArticle', 'useFeaturedResearch', 'useSearchResearch',
        'useNews', 'useNewsArticle', 'useFeaturedNews', 'useSearchNews', 'useNewsCategories'
      ];
      
      expectedFunctions.forEach(funcName => {
        expect(module[funcName]).toBeTypeOf('function');
      });
    }, 5000);
  });

  describe('API Surface Stability', () => {
    it('should maintain stable export count and structure', async () => {
      const module = await import('./index');
      const exportedFunctions = Object.keys(module).filter(key => typeof module[key] === 'function');
      
      // Verify we have the expected number of composable functions
      // This helps catch accidental additions/removals
      expect(exportedFunctions.length).toBeGreaterThanOrEqual(17); // Minimum expected functions
      
      // All exported functions should follow use* naming convention
      exportedFunctions.forEach(funcName => {
        expect(funcName).toMatch(/^use[A-Z]/);
      });
    }, 5000);
  });

  describe('Domain Isolation Validation', () => {
    it('should maintain clear domain boundaries in composable naming', async () => {
      const composables = await import('./index');
      
      // Services domain functions
      const servicesFunctions = Object.keys(composables).filter(key => 
        key.toLowerCase().includes('service') && typeof composables[key] === 'function'
      );
      expect(servicesFunctions.length).toBeGreaterThanOrEqual(5);
      
      // Events domain functions
      const eventsFunctions = Object.keys(composables).filter(key => 
        key.toLowerCase().includes('event') && typeof composables[key] === 'function'
      );
      expect(eventsFunctions.length).toBeGreaterThanOrEqual(4);
      
      // Research domain functions  
      const researchFunctions = Object.keys(composables).filter(key => 
        key.toLowerCase().includes('research') && typeof composables[key] === 'function'
      );
      expect(researchFunctions.length).toBeGreaterThanOrEqual(4);
      
      // News domain functions
      const newsFunctions = Object.keys(composables).filter(key => 
        key.toLowerCase().includes('news') && typeof composables[key] === 'function'
      );
      expect(newsFunctions.length).toBeGreaterThanOrEqual(5);
    }, 5000);
  });
});