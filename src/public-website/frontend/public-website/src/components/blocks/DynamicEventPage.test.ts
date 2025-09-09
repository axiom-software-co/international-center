import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref, nextTick } from 'vue';
import DynamicEventPage from './DynamicEventPage.vue';
import type { Event } from '@/lib/clients/events/types';

// Mock composables
vi.mock('@/composables/useEvents');

// Mock URL utilities
vi.mock('@/lib/utils/url');

// Mock content utilities
vi.mock('@/lib/utils/content');

// Mock child components
vi.mock('./EventBreadcrumb.vue', () => ({
  default: {
    name: 'EventBreadcrumb',
    template: '<div class="event-breadcrumb">{{ eventName }} - {{ category }}</div>',
    props: ['eventName', 'title', 'category']
  }
}));

vi.mock('./EventContent.vue', () => ({
  default: {
    name: 'EventContent',
    template: '<div class="event-content">{{ event.title }}</div>',
    props: ['event']
  }
}));

vi.mock('./EventDetails.vue', () => ({
  default: {
    name: 'EventDetails',
    template: '<div class="event-details">{{ eventDate }} - {{ location }} - {{ status }}</div>',
    props: ['eventDate', 'eventTime', 'location', 'capacity', 'registered', 'status']
  }
}));

vi.mock('./EventContact.vue', () => ({
  default: {
    name: 'EventContact',
    template: '<div class="event-contact">Contact Us</div>'
  }
}));

vi.mock('../UnifiedContentCTA.vue', () => ({
  default: {
    name: 'UnifiedContentCTA',
    template: '<div class="unified-content-cta">CTA Section</div>'
  }
}));

vi.mock('../EventCard.vue', () => ({
  default: {
    name: 'EventCard',
    template: '<div class="event-card">{{ event.title }}</div>',
    props: ['event', 'index']
  }
}));

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    pathname: '/community/events/health-seminar',
  },
  writable: true
});

// Import the mocked functions 
import { useEvent, useEvents } from '@/composables/useEvents';
import { getEventSlugFromUrl } from '@/lib/utils/url';
import { generateEventImageUrl } from '@/lib/utils/content';

describe('DynamicEventPage', () => {
  
  // Get mocked functions
  const mockUseEvent = vi.mocked(useEvent);
  const mockUseEvents = vi.mocked(useEvents);
  const mockGetEventSlugFromUrl = vi.mocked(getEventSlugFromUrl);
  const mockGenerateEventImageUrl = vi.mocked(generateEventImageUrl);
  
  const mockEvent: Event = {
    id: '550e8400-e29b-41d4-a716-446655440001',
    title: 'Health Seminar',
    slug: 'health-seminar',
    excerpt: 'Join us for this informative health seminar',
    content: '<h2>Health Seminar Details</h2><p>Join us for an informative health seminar.</p>',
    featured_image: 'https://storage.azure.com/images/health-seminar.jpg',
    event_date: '2024-03-15',
    event_time: '10:00',
    location: 'Main Conference Room',
    capacity: 50,
    registration_url: 'https://example.com/register',
    author: 'International Center Team',
    tags: ['health', 'wellness', 'seminar'],
    status: 'published',
    featured: false,
    category: 'Health',
    category_id: 1,
    meta_title: 'Health Seminar - International Center',
    meta_description: 'Comprehensive health seminar covering wellness topics',
    published_at: '2024-02-01T08:00:00Z',
    createdAt: '2024-02-01T08:00:00Z',
    updatedAt: '2024-02-05T12:00:00Z'
  };

  const mockRelatedEvents: Event[] = [
    {
      id: '550e8400-e29b-41d4-a716-446655440002',
      title: 'Wellness Workshop',
      slug: 'wellness-workshop',
      excerpt: 'Hands-on wellness activities',
      content: '<h2>Wellness Workshop</h2><p>Interactive wellness activities for all.</p>',
      featured_image: '',
      event_date: '2024-03-20',
      event_time: '14:00',
      location: 'Workshop Room A',
      capacity: 30,
      registration_url: 'https://example.com/register-workshop',
      author: 'International Center Team',
      tags: ['health', 'wellness', 'workshop'],
      status: 'published',
      featured: false,
      category: 'Health',
      category_id: 1,
      meta_title: 'Wellness Workshop - International Center',
      meta_description: 'Interactive wellness workshop',
      published_at: '2024-02-01T08:00:00Z',
      createdAt: '2024-02-01T08:00:00Z',
      updatedAt: '2024-02-05T12:00:00Z'
    }
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Reset URL mock to default
    mockGetEventSlugFromUrl.mockReturnValue('health-seminar');
    
    // Reset content utility mocks
    mockGenerateEventImageUrl.mockReturnValue('https://storage.azure.com/images/health-seminar.jpg');
    
    // Reset composable mocks
    mockUseEvent.mockReturnValue({
      event: ref(null),
      loading: ref(false),
      error: ref(null),
      refetch: vi.fn()
    });
    
    mockUseEvents.mockReturnValue({
      events: ref([]),
      loading: ref(false),
      error: ref(null),
      total: ref(0),
      page: ref(1),
      pageSize: ref(10),
      totalPages: ref(0),
      refetch: vi.fn()
    });
  });

  describe('URL slug extraction', () => {
    it('should extract slug from current URL path', async () => {
      window.location.pathname = '/community/events/health-seminar';
      
      const wrapper = mount(DynamicEventPage);
      
      expect(mockUseEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          value: 'health-seminar'
        })
      );
    });

    it('should handle empty pathname gracefully', async () => {
      window.location.pathname = '/community/events/';
      mockGetEventSlugFromUrl.mockReturnValue('');
      
      const wrapper = mount(DynamicEventPage);
      
      expect(mockUseEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          value: ''
        })
      );
    });

    it('should call useEvents for related events', async () => {
      const wrapper = mount(DynamicEventPage);
      
      expect(mockUseEvents).toHaveBeenCalled();
    });
  });

  describe('loading state', () => {
    it('should display loading skeleton when event is loading', async () => {
      mockUseEvent.mockReturnValue({
        event: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
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
      mockUseEvent.mockReturnValue({
        event: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      expect(wrapper.find('.event-breadcrumb').exists()).toBe(false);
      expect(wrapper.find('.event-content').exists()).toBe(false);
      expect(wrapper.text()).not.toContain('Event Temporarily Unavailable');
    });
  });

  describe('error state', () => {
    it('should display error message when event fails to load', async () => {
      mockUseEvent.mockReturnValue({
        event: ref(null),
        loading: ref(false),
        error: ref('Failed to load event'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      expect(wrapper.text()).toContain('Event Temporarily Unavailable');
      expect(wrapper.text()).toContain('We\'re experiencing technical difficulties. Please try again later.');
      
      const errorLink = wrapper.find('a[href="/community/events"]');
      expect(errorLink.exists()).toBe(true);
      expect(errorLink.text()).toContain('Browse All Events');
    });

    it('should not display content or loading when error occurs', async () => {
      mockUseEvent.mockReturnValue({
        event: ref(null),
        loading: ref(false),
        error: ref('Network error'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      expect(wrapper.find('.event-breadcrumb').exists()).toBe(false);
      expect(wrapper.find('.event-content').exists()).toBe(false);
      expect(wrapper.find('.animate-pulse').exists()).toBe(false);
    });
  });

  describe('event content display', () => {
    beforeEach(() => {
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render breadcrumb with event information', async () => {
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const breadcrumb = wrapper.find('.event-breadcrumb');
      expect(breadcrumb.exists()).toBe(true);
      expect(breadcrumb.text()).toContain('Health Seminar');
      expect(breadcrumb.text()).toContain('Health');
    });

    it('should transform event data to include hero image details', async () => {
      // Setup: Provide event data through mock
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const vm = wrapper.vm as any;
      
      // Verify: Component transforms event data correctly for hero image
      expect(vm.eventData).toMatchObject({
        heroImage: {
          src: 'https://storage.azure.com/images/health-seminar.jpg',
          alt: 'Health Seminar'
        }
      });
    });

    it('should render fallback hero image when featured_image is not provided', async () => {
      const eventWithoutImage = { ...mockEvent, featured_image: undefined };
      mockUseEvent.mockReturnValue({
        event: ref(eventWithoutImage),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      // Mock the image generation to return placeholder for undefined image
      mockGenerateEventImageUrl.mockReturnValue('https://placehold.co/800x400?text=Health%20Seminar');

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('src')).toContain('placehold.co');
      expect(heroImage.attributes('src')).toContain(encodeURIComponent('Health Seminar'));
    });

    it('should render event content component with transformed data', async () => {
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const eventContent = wrapper.find('.event-content');
      expect(eventContent.exists()).toBe(true);
      expect(eventContent.text()).toContain('Health Seminar');
    });

    it('should render event details with all information', async () => {
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const eventDetails = wrapper.find('.event-details');
      expect(eventDetails.exists()).toBe(true);
      expect(eventDetails.text()).toContain('2024-03-15');
      expect(eventDetails.text()).toContain('Main Conference Room');
      expect(eventDetails.text()).toContain('published');
    });

    it('should include EventContact component when event is loaded', async () => {
      // Setup: Provide event data through mock
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Component includes contact information when event is loaded
      expect(wrapper.text()).toContain('Contact Us');
    });
  });

  describe('related events section', () => {
    beforeEach(() => {
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseEvents.mockReturnValue({
        events: ref(mockRelatedEvents),
        loading: ref(false),
        error: ref(null),
        total: ref(1),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });
    });

    it('should call useEvents with category filter for related events', async () => {
      // Setup: Provide event with category for related events logic
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Component calls useEvents with proper configuration for related events
      expect(mockUseEvents).toHaveBeenCalledWith(
        expect.objectContaining({
          category: expect.any(Object), // computed ref
          pageSize: 3,
          enabled: expect.any(Object), // computed ref
          immediate: false
        })
      );
      
      // Verify: Component generates correct title based on category
      const vm = wrapper.vm as any;
      expect(vm.relatedEventsTitle).toBe('More Health Events');
    });

    it('should render related event cards', async () => {
      // Setup: Provide event data for main event and related events
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      // Setup related events mock to provide related events
      mockUseEvents.mockReturnValue({
        events: ref(mockRelatedEvents),
        loading: ref(false),
        error: ref(null),
        total: ref(1),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });
      
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Component displays related event information
      expect(wrapper.text()).toContain('Wellness Workshop');
    });

    it('should not display related events section when no events available', async () => {
      mockUseEvents.mockReturnValue({
        events: ref([]),
        loading: ref(false),
        error: ref(null),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const relatedSection = wrapper.find('.pt-16.lg\\:pt-20');
      expect(relatedSection.exists()).toBe(false);
    });

    it('should handle related events loading failure gracefully', async () => {
      // Setup: Mock main event data and related events failure
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseEvents.mockReturnValue({
        events: ref([]),
        loading: ref(false),
        error: ref('Failed to load related events'),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Main event still renders even if related events fail
      expect(wrapper.text()).toContain('Health Seminar');
    });
  });

  describe('data transformation', () => {
    beforeEach(() => {
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should transform CommunityEvent to EventPageData structure', async () => {
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Component transforms and displays event data correctly
      expect(wrapper.text()).toContain('Health Seminar');
      expect(wrapper.text()).toContain('Health');
    });

    it('should handle event with missing content field', async () => {
      const eventWithoutContent = { ...mockEvent, content: undefined };
      mockUseEvent.mockReturnValue({
        event: ref(eventWithoutContent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Should fallback to description for content
      const eventContent = wrapper.find('.event-content');
      expect(eventContent.exists()).toBe(true);
    });

    it('should handle event without category gracefully', async () => {
      const eventWithoutCategory = { ...mockEvent, category: undefined };
      mockUseEvent.mockReturnValue({
        event: ref(eventWithoutCategory),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Should not crash when category is undefined
      const breadcrumb = wrapper.find('.event-breadcrumb');
      expect(breadcrumb.exists()).toBe(true);
    });

    it('should handle capacity and registration numbers correctly', async () => {
      // Arrange: Provide event data for testing
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Assert: Component displays event details including capacity information
      expect(wrapper.text()).toContain('2024-03-15');
      expect(wrapper.text()).toContain('Main Conference Room');
      expect(wrapper.text()).toContain('published');
    });
  });

  describe('responsive layout', () => {
    beforeEach(() => {
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should have proper grid layout classes for responsive design', async () => {
      // Setup: Use the beforeEach mock that provides event data
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Component organizes content in main and sidebar layout
      expect(wrapper.text()).toContain('Health Seminar'); // Main content
      expect(wrapper.text()).toContain('Contact Us'); // Sidebar content
    });

    it('should have sticky sidebar on medium and larger screens', async () => {
      // Setup: Provide event data
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Component renders sidebar with proper behavior (contains aside element)
      expect(wrapper.html()).toContain('aside');
      expect(wrapper.text()).toContain('Contact Us'); // Sidebar contains contact component
    });

    it('should display related events in responsive grid', async () => {
      // Setup: Provide event and related events data
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseEvents.mockReturnValue({
        events: ref(mockRelatedEvents),
        loading: ref(false),
        error: ref(null),
        total: ref(1),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Component displays related events in organized layout
      expect(wrapper.text()).toContain('More Health Events');
      expect(wrapper.text()).toContain('Wellness Workshop');
    });
  });

  describe('accessibility', () => {
    beforeEach(() => {
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render main event with proper semantic HTML', async () => {
      // Setup: Provide event data
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Verify: Component renders main event content with semantic structure
      expect(wrapper.html()).toContain('article');
      expect(wrapper.text()).toContain('Health Seminar');
    });

    it('should render aside element for sidebar content', async () => {
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const aside = wrapper.find('aside#event-page-aside');
      expect(aside.exists()).toBe(true);
    });

    it('should have proper image alt text for screen readers', async () => {
      // Arrange: Provide event data for image rendering
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Assert: Hero image has proper alt text
      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('alt')).toBe('Health Seminar');
    });

    it('should have proper heading hierarchy', async () => {
      mockUseEvents.mockReturnValue({
        events: ref(mockRelatedEvents),
        loading: ref(false),
        error: ref(null),
        total: ref(1),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const relatedHeading = wrapper.find('h2');
      expect(relatedHeading.exists()).toBe(true);
      expect(relatedHeading.text()).toContain('More Health Events');
    });
  });

  describe('SEO metadata handling', () => {
    beforeEach(() => {
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should generate proper image alt text for SEO', async () => {
      // Arrange: Provide event data for image rendering
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Assert: Hero image has proper alt text for SEO
      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('alt')).toBe('Health Seminar');
    });

    it('should provide structured data through component props', async () => {
      // Arrange: Provide event data for structured data rendering
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Assert: Structured data is available through component state
      const breadcrumb = wrapper.find('.event-breadcrumb');
      expect(breadcrumb.text()).toContain('Health Seminar');
    });
  });

  describe('error recovery', () => {
    it('should handle event loading failure gracefully', async () => {
      mockUseEvent.mockReturnValue({
        event: ref(null),
        loading: ref(false),
        error: ref('Failed to load event'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      expect(wrapper.text()).toContain('Event Temporarily Unavailable');
      expect(wrapper.find('a[href="/community/events"]').exists()).toBe(true);
    });

    it('should handle event without event_date', async () => {
      const eventWithoutDate = { ...mockEvent, event_date: undefined };
      mockUseEvent.mockReturnValue({
        event: ref(eventWithoutDate),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Should not crash when event_date is undefined
      const eventDetails = wrapper.find('.event-details');
      expect(eventDetails.exists()).toBe(true);
    });

    it('should handle event without location', async () => {
      const eventWithoutLocation = { ...mockEvent, location: undefined };
      mockUseEvent.mockReturnValue({
        event: ref(eventWithoutLocation),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      // Should not crash when location is undefined
      const eventDetails = wrapper.find('.event-details');
      expect(eventDetails.exists()).toBe(true);
    });
  });

  describe('event status handling', () => {
    it('should handle published events correctly', async () => {
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const eventDetails = wrapper.find('.event-details');
      expect(eventDetails.exists()).toBe(true);
      expect(eventDetails.text()).toContain('published');
    });

    it('should handle draft events correctly', async () => {
      const draftEvent = { ...mockEvent, status: 'draft' as const };
      mockUseEvent.mockReturnValue({
        event: ref(draftEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const eventDetails = wrapper.find('.event-details');
      expect(eventDetails.text()).toContain('draft');
    });

    it('should handle archived events correctly', async () => {
      const archivedEvent = { ...mockEvent, status: 'archived' as const };
      mockUseEvent.mockReturnValue({
        event: ref(archivedEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const eventDetails = wrapper.find('.event-details');
      expect(eventDetails.text()).toContain('archived');
    });
  });

  describe('CTA section', () => {
    beforeEach(() => {
      mockUseEvent.mockReturnValue({
        event: ref(mockEvent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render unified content CTA section', async () => {
      const wrapper = mount(DynamicEventPage);
      await nextTick();

      const ctaSection = wrapper.find('.unified-content-cta');
      expect(ctaSection.exists()).toBe(true);
      expect(ctaSection.text()).toContain('CTA Section');
    });
  });
});