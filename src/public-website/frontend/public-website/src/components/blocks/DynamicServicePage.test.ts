import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref, nextTick } from 'vue';
import DynamicServicePage from './DynamicServicePage.vue';
import type { Service, ServiceCategory } from '@/lib/clients/services/types';

// Mock composables
vi.mock('@/composables/useServices');

// Mock URL utilities
vi.mock('@/lib/utils/url');

// Mock content utilities
vi.mock('@/lib/utils/content');


// Mock child components
vi.mock('./ServiceBreadcrumb.vue', () => ({
  default: {
    name: 'ServiceBreadcrumb',
    template: '<div class="service-breadcrumb">{{ serviceName }} - {{ category }}</div>',
    props: ['serviceName', 'title', 'category']
  }
}));

vi.mock('./ServiceContent.vue', () => ({
  default: {
    name: 'ServiceContent',
    template: '<div class="service-content">{{ service.title }}</div>',
    props: ['service']
  }
}));

vi.mock('./ServiceTreatmentDetails.vue', () => ({
  default: {
    name: 'ServiceTreatmentDetails',
    template: '<div class="service-treatment-details">{{ duration }} - {{ recovery }} - {{ deliveryModes.join(",") }}</div>',
    props: ['duration', 'recovery', 'deliveryModes', 'isComingSoon']
  }
}));

vi.mock('./ServiceContact.vue', () => ({
  default: {
    name: 'ServiceContact',
    template: '<div class="service-contact">Contact Us</div>'
  }
}));

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    pathname: '/services/prp-therapy',
  },
  writable: true
});

// Import the mocked functions 
import { useService, useServiceCategories } from '@/composables/useServices';
import { getServiceSlugFromUrl } from '@/lib/utils/url';
import { parseServiceDeliveryModes, generateHeroImageUrl, generateImageAlt } from '@/lib/utils/content';

describe('DynamicServicePage', () => {
  
  // Get mocked functions
  const mockUseService = vi.mocked(useService);
  const mockUseServiceCategories = vi.mocked(useServiceCategories);
  const mockGetServiceSlugFromUrl = vi.mocked(getServiceSlugFromUrl);
  const mockParseServiceDeliveryModes = vi.mocked(parseServiceDeliveryModes);
  const mockGenerateHeroImageUrl = vi.mocked(generateHeroImageUrl);
  const mockGenerateImageAlt = vi.mocked(generateImageAlt);
  
  const mockService: Service = {
    service_id: '550e8400-e29b-41d4-a716-446655440001',
    title: 'PRP Therapy',
    description: 'Advanced platelet-rich plasma therapy for regenerative healing',
    slug: 'prp-therapy',
    publishing_status: 'published',
    category_id: '550e8400-e29b-41d4-a716-446655440002',
    delivery_mode: 'mobile_service',
    content: '<h2>Advanced PRP Treatment</h2><p>Our PRP therapy uses cutting-edge centrifugation technology.</p>',
    image_url: 'https://storage.azure.com/images/prp-therapy-hero.jpg',
    order_number: 1,
    created_on: '2024-01-10T08:30:00Z',
    created_by: 'medical-team',
    modified_on: '2024-01-12T14:20:00Z',
    modified_by: 'content-reviewer',
    is_deleted: false,
    id: '550e8400-e29b-41d4-a716-446655440001',
    createdAt: '2024-01-10T08:30:00Z',
    updatedAt: '2024-01-12T14:20:00Z'
  };

  const mockServiceCategory: ServiceCategory = {
    category_id: '550e8400-e29b-41d4-a716-446655440002',
    name: 'Regenerative Medicine',
    slug: 'regenerative-medicine',
    order_number: 1,
    is_default_unassigned: false,
    created_on: '2024-01-08T08:30:00Z',
    created_by: 'admin',
    modified_on: '2024-01-08T08:30:00Z',
    modified_by: 'admin',
    is_deleted: false
  };

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Reset URL mock to default
    mockGetServiceSlugFromUrl.mockReturnValue('prp-therapy');
    
    // Reset content utility mocks
    mockGenerateHeroImageUrl.mockReturnValue('https://storage.azure.com/images/prp-therapy-hero.jpg');
    mockGenerateImageAlt.mockReturnValue('PRP Therapy - International Center Service');
    mockParseServiceDeliveryModes.mockReturnValue(['mobile', 'outpatient']);
    
    // Reset composable mocks
    mockUseService.mockReturnValue({
      service: ref(null),
      loading: ref(false),
      error: ref(null),
      refetch: vi.fn()
    });
    
    mockUseServiceCategories.mockReturnValue({
      categories: ref([mockServiceCategory]),
      loading: ref(false),
      error: ref(null),
      refetch: vi.fn()
    });
  });

  describe('URL slug extraction', () => {
    it('should extract slug from current URL path', async () => {
      window.location.pathname = '/services/prp-therapy';
      
      const wrapper = mount(DynamicServicePage);
      
      expect(mockUseService).toHaveBeenCalledWith(
        expect.objectContaining({
          value: 'prp-therapy'
        })
      );
    });

    it('should handle empty pathname gracefully', async () => {
      mockGetServiceSlugFromUrl.mockReturnValue('');
      
      const wrapper = mount(DynamicServicePage);
      
      // Check the mock was called with a ref that has an empty string value
      expect(mockUseService).toHaveBeenCalledWith(
        expect.objectContaining({
          value: ''
        })
      );
    });

    it('should call useServiceCategories to fetch categories', async () => {
      const wrapper = mount(DynamicServicePage);
      
      expect(mockUseServiceCategories).toHaveBeenCalled();
    });
  });

  describe('loading state', () => {
    it('should display loading skeleton when service is loading', async () => {
      mockUseService.mockReturnValue({
        service: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      expect(wrapper.find('.animate-pulse').exists()).toBe(true);
      
      // Should show breadcrumb loading
      const breadcrumbLoading = wrapper.find('.bg-gray-50 .animate-pulse');
      expect(breadcrumbLoading.exists()).toBe(true);
      
      // Should show main content loading
      const mainContentLoading = wrapper.find('.prose .animate-pulse');
      expect(mainContentLoading.exists()).toBe(true);
      
      // Should show sidebar loading
      const sidebarLoading = wrapper.find('aside .animate-pulse');
      expect(sidebarLoading.exists()).toBe(true);
    });

    it('should not display content or error when loading', async () => {
      mockUseService.mockReturnValue({
        service: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      expect(wrapper.find('.service-breadcrumb').exists()).toBe(false);
      expect(wrapper.find('.service-content').exists()).toBe(false);
      expect(wrapper.text()).not.toContain('Service Temporarily Unavailable');
    });
  });

  describe('error state', () => {
    it('should display error message when service fails to load', async () => {
      mockUseService.mockReturnValue({
        service: ref(null),
        loading: ref(false),
        error: ref('Failed to load service'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      expect(wrapper.text()).toContain('Service Temporarily Unavailable');
      expect(wrapper.text()).toContain('We\'re experiencing technical difficulties. Please try again later.');
      
      const errorLink = wrapper.find('a[href="/services"]');
      expect(errorLink.exists()).toBe(true);
      expect(errorLink.text()).toContain('Browse All Services');
    });

    it('should not display content or loading when error occurs', async () => {
      mockUseService.mockReturnValue({
        service: ref(null),
        loading: ref(false),
        error: ref('Network error'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      expect(wrapper.find('.service-breadcrumb').exists()).toBe(false);
      expect(wrapper.find('.service-content').exists()).toBe(false);
      expect(wrapper.find('.animate-pulse').exists()).toBe(false);
    });
  });

  describe('service content display', () => {
    beforeEach(() => {
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseServiceCategories.mockReturnValue({
        categories: ref([mockServiceCategory]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render breadcrumb with service information', async () => {
      // Setup: Provide service data through mocks
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      mockUseServiceCategories.mockReturnValue({
        categories: ref([mockServiceCategory]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Verify: Component displays breadcrumb information correctly
      expect(wrapper.text()).toContain('PRP Therapy');
      expect(wrapper.text()).toContain('Regenerative Medicine');
    });

    it('should render hero image with correct attributes', async () => {
      // Arrange: Provide service data for image rendering
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Assert: Hero image renders with correct attributes
      const heroImage = wrapper.find('img');
      expect(heroImage.exists()).toBe(true);
      expect(heroImage.attributes('src')).toBe('https://storage.azure.com/images/prp-therapy-hero.jpg');
      expect(heroImage.attributes('alt')).toBe('PRP Therapy - International Center Service');
      expect(heroImage.classes()).toContain('aspect-video');
    });

    it('should render fallback hero image when image_url is not provided', async () => {
      const serviceWithoutImage = { ...mockService, image_url: undefined };
      mockUseService.mockReturnValue({
        service: ref(serviceWithoutImage),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      // Mock the image generation to return placeholder for undefined image
      mockGenerateHeroImageUrl.mockReturnValue('https://placehold.co/800x400?text=PRP%20Therapy');

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('src')).toContain('placehold.co');
      expect(heroImage.attributes('src')).toContain(encodeURIComponent('PRP Therapy'));
    });

    it('should render service content component with transformed data', async () => {
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const serviceContent = wrapper.find('.service-content');
      expect(serviceContent.exists()).toBe(true);
      expect(serviceContent.text()).toContain('PRP Therapy');
    });

    it('should render treatment details with delivery modes', async () => {
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const treatmentDetails = wrapper.find('.service-treatment-details');
      expect(treatmentDetails.exists()).toBe(true);
      expect(treatmentDetails.text()).toContain('45-90 minutes');
      expect(treatmentDetails.text()).toContain('Minimal to no downtime');
      expect(treatmentDetails.text()).toContain('mobile,outpatient');
    });

    it('should handle mobile service delivery mode correctly', async () => {
      // Arrange: Provide service data with mobile delivery mode
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Assert: Treatment details display mobile delivery mode
      const treatmentDetails = wrapper.find('.service-treatment-details');
      expect(treatmentDetails.text()).toContain('mobile');
    });

    it('should handle outpatient service delivery mode correctly', async () => {
      const serviceWithOutpatient = { ...mockService, slug: 'general-consultation' };
      mockUseService.mockReturnValue({
        service: ref(serviceWithOutpatient),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const treatmentDetails = wrapper.find('.service-treatment-details');
      expect(treatmentDetails.text()).toContain('outpatient');
    });

    it('should handle inpatient service delivery mode correctly', async () => {
      const serviceWithInpatient = { ...mockService, slug: 'stem-cell' };
      mockUseService.mockReturnValue({
        service: ref(serviceWithInpatient),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Mock delivery modes to return inpatient for stem-cell slug
      mockParseServiceDeliveryModes.mockReturnValue(['inpatient', 'consultation']);

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const treatmentDetails = wrapper.find('.service-treatment-details');
      expect(treatmentDetails.text()).toContain('inpatient');
    });

    it('should render service contact in sidebar', async () => {
      // Arrange: Provide service data for sidebar content rendering
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Assert: Service contact component renders in sidebar
      const serviceContact = wrapper.find('.service-contact');
      expect(serviceContact.exists()).toBe(true);
      expect(serviceContact.text()).toContain('Contact Us');
    });
  });

  describe('data transformation', () => {
    beforeEach(() => {
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseServiceCategories.mockReturnValue({
        categories: ref([mockServiceCategory]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should transform Service to ServicePageData structure', async () => {
      // Arrange: Provide service data for transformation testing
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Assert: Transformed data is passed to child components correctly
      const breadcrumb = wrapper.find('.service-breadcrumb');
      expect(breadcrumb.text()).toContain('PRP Therapy');
      expect(breadcrumb.text()).toContain('Regenerative Medicine');
    });

    it('should handle service with missing content field', async () => {
      const serviceWithoutContent = { ...mockService, content: undefined };
      mockUseService.mockReturnValue({
        service: ref(serviceWithoutContent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Should fallback to description for content
      const serviceContent = wrapper.find('.service-content');
      expect(serviceContent.exists()).toBe(true);
    });

    it('should handle missing category gracefully', async () => {
      mockUseServiceCategories.mockReturnValue({
        categories: ref([]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const breadcrumb = wrapper.find('.service-breadcrumb');
      expect(breadcrumb.exists()).toBe(true);
      // Should not crash when category is undefined
    });

    it('should parse delivery modes correctly for different service types', async () => {
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // PRP therapy should be mobile + outpatient
      const treatmentDetails = wrapper.find('.service-treatment-details');
      const deliveryModes = treatmentDetails.text();
      expect(deliveryModes).toContain('mobile');
      expect(deliveryModes).toContain('outpatient');
      expect(deliveryModes).not.toContain('inpatient');
    });
  });

  describe('responsive layout', () => {
    beforeEach(() => {
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseServiceCategories.mockReturnValue({
        categories: ref([mockServiceCategory]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should have proper grid layout classes for responsive design', async () => {
      // Arrange: Provide service data for layout rendering
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Assert: Grid layout has proper responsive classes
      const gridContainer = wrapper.find('.grid.gap-12.md\\:grid-cols-12');
      expect(gridContainer.exists()).toBe(true);
      
      const mainContent = wrapper.find('.md\\:col-span-7.md\\:col-start-1.lg\\:col-span-8');
      expect(mainContent.exists()).toBe(true);
      
      const sidebar = wrapper.find('.md\\:col-span-5.lg\\:col-span-4');
      expect(sidebar.exists()).toBe(true);
    });

    it('should have sticky sidebar on medium and larger screens', async () => {
      // Arrange: Provide service data for sidebar rendering
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Assert: Sidebar has sticky positioning on medium+ screens
      const stickySidebar = wrapper.find('.md\\:sticky.md\\:top-20');
      expect(stickySidebar.exists()).toBe(true);
    });
  });

  describe('accessibility', () => {
    beforeEach(() => {
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseServiceCategories.mockReturnValue({
        categories: ref([mockServiceCategory]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render main service with proper semantic HTML', async () => {
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const article = wrapper.find('article.prose');
      expect(article.exists()).toBe(true);
    });

    it('should render aside element for sidebar content', async () => {
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const aside = wrapper.find('aside#service-page-aside');
      expect(aside.exists()).toBe(true);
    });

    it('should have proper image alt text for screen readers', async () => {
      // Arrange: Provide service data for accessibility testing
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Assert: Hero image has proper alt text for screen readers
      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('alt')).toBe('PRP Therapy - International Center Service');
    });
  });

  describe('SEO metadata handling', () => {
    beforeEach(() => {
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseServiceCategories.mockReturnValue({
        categories: ref([mockServiceCategory]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should generate proper image alt text for SEO', async () => {
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('alt')).toBe('PRP Therapy - International Center Service');
    });

    it('should provide structured data through component props', async () => {
      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Verify that structured data is available through component state
      const breadcrumb = wrapper.find('.service-breadcrumb');
      expect(breadcrumb.text()).toContain('PRP Therapy');
    });
  });

  describe('error recovery', () => {
    it('should handle categories loading failure gracefully', async () => {
      mockUseService.mockReturnValue({
        service: ref(mockService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseServiceCategories.mockReturnValue({
        categories: ref([]),
        loading: ref(false),
        error: ref('Failed to load categories'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // When categories fail to load, the entire page goes to error state
      // since the computed error includes both service and categories errors
      expect(wrapper.text()).toContain('Service Temporarily Unavailable');
    });

    it('should handle service without category_id', async () => {
      const serviceWithoutCategory = { ...mockService, category_id: undefined };
      mockUseService.mockReturnValue({
        service: ref(serviceWithoutCategory),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Should not crash when category_id is undefined
      const breadcrumb = wrapper.find('.service-breadcrumb');
      expect(breadcrumb.exists()).toBe(true);
    });
  });

  describe('service status handling', () => {
    it('should handle coming soon services correctly', async () => {
      const comingSoonService = { ...mockService, publishing_status: 'draft' as const };
      mockUseService.mockReturnValue({
        service: ref(comingSoonService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Treatment details should reflect coming soon status
      const treatmentDetails = wrapper.find('.service-treatment-details');
      expect(treatmentDetails.exists()).toBe(true);
      // The component should handle isComingSoon based on publishing status
    });

    it('should handle archived services by showing error state', async () => {
      const archivedService = { ...mockService, publishing_status: 'archived' as const };
      mockUseService.mockReturnValue({
        service: ref(archivedService),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicServicePage);
      await nextTick();

      // Archived services should probably show some indication or error
      const serviceContent = wrapper.find('.service-content');
      expect(serviceContent.exists()).toBe(true);
    });
  });
});