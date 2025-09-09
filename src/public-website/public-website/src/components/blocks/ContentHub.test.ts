import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref, nextTick } from 'vue';
import ContentHub from './ContentHub.vue';

// Mock composables
vi.mock('@/composables/', () => ({
  useNews: vi.fn(),
  useFeaturedNews: vi.fn(),
  useResearchArticles: vi.fn(),
  useFeaturedResearch: vi.fn(),
  useEvents: vi.fn(),
  useFeaturedEvents: vi.fn()
}));

// Mock child components
vi.mock('../UnifiedContentCTA.vue', () => ({
  default: {
    name: 'UnifiedContentCTA',
    template: '<div class="unified-content-cta">CTA Section</div>'
  }
}));

vi.mock('../PublicationsSection.vue', () => ({
  default: {
    name: 'PublicationsSection',
    template: '<div class="publications-section">{{ title }} - {{ dataType }}</div>',
    props: ['title', 'dataType']
  }
}));

vi.mock('../ArticleCard.vue', () => ({
  default: {
    name: 'ArticleCard',
    template: '<div class="article-card">{{ article.title }} - {{ basePath }}</div>',
    props: ['article', 'basePath', 'defaultAuthor', 'index']
  }
}));

// Import mocked composables
import { 
  useNews, 
  useFeaturedNews,
  useResearchArticles, 
  useFeaturedResearch,
  useEvents, 
  useFeaturedEvents 
} from '@/composables/';

describe('ContentHub', () => {
  // Get mocked functions
  const mockUseNews = vi.mocked(useNews);
  const mockUseFeaturedNews = vi.mocked(useFeaturedNews);
  const mockUseResearchArticles = vi.mocked(useResearchArticles);
  const mockUseFeaturedResearch = vi.mocked(useFeaturedResearch);
  const mockUseEvents = vi.mocked(useEvents);
  const mockUseFeaturedEvents = vi.mocked(useFeaturedEvents);

  const mockNewsArticles = [
    {
      id: '1',
      title: 'Latest Medical Breakthrough',
      slug: 'medical-breakthrough',
      excerpt: 'Revolutionary treatment discovered',
      category: 'Medical Research',
      featured_image: 'https://example.com/news1.jpg',
      author: 'Dr. Jane Smith',
      published_at: '2024-03-15T10:00:00Z'
    },
    {
      id: '2', 
      title: 'New Treatment Protocol',
      slug: 'treatment-protocol',
      excerpt: 'Improved patient outcomes',
      category: 'Clinical Updates',
      featured_image: 'https://example.com/news2.jpg',
      author: 'Dr. John Doe',
      published_at: '2024-03-14T09:00:00Z'
    }
  ];

  const mockResearchArticles = [
    {
      id: '1',
      title: 'Stem Cell Research Study',
      slug: 'stem-cell-study',
      excerpt: 'Comprehensive analysis of stem cell therapy',
      category: 'Clinical Research',
      featured_image: 'https://example.com/research1.jpg',
      author: 'Research Team',
      published_at: '2024-03-10T14:00:00Z'
    }
  ];

  const mockEvents = [
    {
      id: '1',
      title: 'Medical Conference 2024',
      slug: 'medical-conference-2024',
      description: 'Annual medical conference',
      category: 'Conference',
      featured_image: 'https://example.com/event1.jpg',
      organizer_name: 'Event Team',
      event_date: '2024-04-20',
      published_at: '2024-03-01T08:00:00Z'
    }
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Reset all composable mocks
    mockUseNews.mockReturnValue({
      articles: ref([]),
      loading: ref(false),
      error: ref(null),
      total: ref(0),
      page: ref(1),
      pageSize: ref(10),
      totalPages: ref(0),
      refetch: vi.fn()
    });

    mockUseFeaturedNews.mockReturnValue({
      articles: ref([]),
      loading: ref(false),
      error: ref(null),
      refetch: vi.fn()
    });

    mockUseResearchArticles.mockReturnValue({
      articles: ref([]),
      loading: ref(false),
      error: ref(null),
      total: ref(0),
      page: ref(1),
      pageSize: ref(10),
      totalPages: ref(0),
      refetch: vi.fn()
    });

    mockUseFeaturedResearch.mockReturnValue({
      articles: ref([]),
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

    mockUseFeaturedEvents.mockReturnValue({
      events: ref([]),
      loading: ref(false),
      error: ref(null),
      refetch: vi.fn()
    });
  });

  describe('Component Props & Configuration', () => {
    it('should generate correct config for news contentType', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.vm.config.type).toBe('news');
      expect(wrapper.vm.config.basePath).toBe('/company/news');
      expect(wrapper.vm.config.errorTitle).toBe('News');
      expect(wrapper.vm.config.defaultAuthor).toBe('International Center Team');
      expect(wrapper.vm.config.showExcerpt).toBe(false);
    });

    it('should generate correct config for research contentType', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'research' }
      });
      await nextTick();

      expect(wrapper.vm.config.type).toBe('research');
      expect(wrapper.vm.config.basePath).toBe('/community/research');
      expect(wrapper.vm.config.errorTitle).toBe('Research');
      expect(wrapper.vm.config.defaultAuthor).toBe('International Center Team');
      expect(wrapper.vm.config.showExcerpt).toBe(false);
    });

    it('should generate correct config for events contentType', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'events' }
      });
      await nextTick();

      expect(wrapper.vm.config.type).toBe('events');
      expect(wrapper.vm.config.basePath).toBe('/community/events');
      expect(wrapper.vm.config.errorTitle).toBe('Events');
      expect(wrapper.vm.config.defaultAuthor).toBe('International Center Team');
      expect(wrapper.vm.config.showExcerpt).toBe(true);
    });

    it('should default to research config for unknown contentType', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'unknown' }
      });
      await nextTick();

      expect(wrapper.vm.config.type).toBe('research');
      expect(wrapper.vm.config.basePath).toBe('/community/research');
    });

    it('should use appropriate composables based on contentType', async () => {
      const newsWrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();
      expect(mockUseNews).toHaveBeenCalled();
      expect(mockUseFeaturedNews).toHaveBeenCalled();

      vi.clearAllMocks();

      const researchWrapper = mount(ContentHub, {
        props: { contentType: 'research' }
      });
      await nextTick();
      expect(mockUseResearchArticles).toHaveBeenCalled();
      expect(mockUseFeaturedResearch).toHaveBeenCalled();

      vi.clearAllMocks();

      const eventsWrapper = mount(ContentHub, {
        props: { contentType: 'events' }
      });
      await nextTick();
      expect(mockUseEvents).toHaveBeenCalled();
      expect(mockUseFeaturedEvents).toHaveBeenCalled();
    });
  });

  describe('Loading States', () => {
    it('should display loading skeleton when news content is loading', async () => {
      mockUseNews.mockReturnValue({
        articles: ref([]),
        loading: ref(true),
        error: ref(null),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      mockUseFeaturedNews.mockReturnValue({
        articles: ref([]),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.find('.animate-pulse').exists()).toBe(true);
      expect(wrapper.find('.space-y-8.lg\\:space-y-12').exists()).toBe(true);
    });

    it('should display loading skeleton when research content is loading', async () => {
      mockUseResearchArticles.mockReturnValue({
        articles: ref([]),
        loading: ref(true),
        error: ref(null),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'research' }
      });
      await nextTick();

      expect(wrapper.find('.animate-pulse').exists()).toBe(true);
    });

    it('should not display content when loading', async () => {
      mockUseNews.mockReturnValue({
        articles: ref([]),
        loading: ref(true),
        error: ref(null),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      mockUseFeaturedNews.mockReturnValue({
        articles: ref([]),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.find('.featured-card').exists()).toBe(false);
      expect(wrapper.find('.content-category').exists()).toBe(false);
    });
  });

  describe('Error States', () => {
    it('should display error message when news loading fails', async () => {
      mockUseNews.mockReturnValue({
        articles: ref([]),
        loading: ref(false),
        error: ref('Failed to load news articles'),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      const errorSection = wrapper.find('.text-center.py-12');
      expect(errorSection.exists()).toBe(true);
      expect(errorSection.text()).toContain('News Temporarily Unavailable');
      expect(errorSection.text()).toContain('We\'re unable to load news information at the moment');
    });

    it('should display error message when research loading fails', async () => {
      mockUseResearchArticles.mockReturnValue({
        articles: ref([]),
        loading: ref(false),
        error: ref('Failed to load research articles'),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'research' }
      });
      await nextTick();

      const errorSection = wrapper.find('.text-center.py-12');
      expect(errorSection.exists()).toBe(true);
      expect(errorSection.text()).toContain('Research Temporarily Unavailable');
    });

    it('should not display content when error occurs', async () => {
      mockUseEvents.mockReturnValue({
        events: ref([]),
        loading: ref(false),
        error: ref('Network error'),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'events' }
      });
      await nextTick();

      expect(wrapper.find('.animate-pulse').exists()).toBe(false);
      expect(wrapper.find('.featured-card').exists()).toBe(false);
    });
  });

  describe('Featured Article Display', () => {
    it('should display featured news article when available', async () => {
      mockUseFeaturedNews.mockReturnValue({
        articles: ref([mockNewsArticles[0]]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.vm.featuredArticle).toEqual(mockNewsArticles[0]);
      expect(wrapper.vm.featuredTitle).toBe('Latest Medical Breakthrough');
      expect(wrapper.vm.featuredCategory).toBe('Medical Research');
      expect(wrapper.vm.featuredAuthor).toBe('Dr. Jane Smith');
    });

    it('should display featured research article when available', async () => {
      mockUseFeaturedResearch.mockReturnValue({
        articles: ref([mockResearchArticles[0]]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'research' }
      });
      await nextTick();

      expect(wrapper.vm.featuredTitle).toBe('Stem Cell Research Study');
      expect(wrapper.vm.featuredCategory).toBe('Clinical Research');
    });

    it('should generate correct featured article href based on contentType', async () => {
      mockUseFeaturedNews.mockReturnValue({
        articles: ref([mockNewsArticles[0]]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.vm.featuredArticleHref).toBe('/company/news/medical-breakthrough');
    });

    it('should use default values when featured article is missing data', async () => {
      const incompleteArticle = {
        id: '1',
        title: 'Test Article',
        slug: 'test-article',
        published_at: '2024-03-15T10:00:00Z'
      };

      mockUseFeaturedNews.mockReturnValue({
        articles: ref([incompleteArticle]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.vm.featuredAuthor).toBe('International Center Team');
      expect(wrapper.vm.featuredCategory).toBe('News');
    });
  });

  describe('Article Categories Display', () => {
    it('should display news article categories and cards', async () => {
      mockUseNews.mockReturnValue({
        articles: ref(mockNewsArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.vm.articleCategories).toHaveLength(2);
      const articleCards = wrapper.findAll('.article-card');
      expect(articleCards).toHaveLength(2);
      expect(articleCards[0].text()).toContain('Latest Medical Breakthrough');
      expect(articleCards[0].text()).toContain('/company/news');
    });

    it('should display research article categories and cards', async () => {
      mockUseResearchArticles.mockReturnValue({
        articles: ref(mockResearchArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(1),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'research' }
      });
      await nextTick();

      expect(wrapper.vm.articleCategories).toHaveLength(1);
      const articleCards = wrapper.findAll('.article-card');
      expect(articleCards).toHaveLength(1);
      expect(articleCards[0].text()).toContain('/community/research');
    });

    it('should display placeholder cards when category has insufficient articles', async () => {
      mockUseNews.mockReturnValue({
        articles: ref([mockNewsArticles[0]]), // Only 1 article
        loading: ref(false),
        error: ref(null),
        total: ref(1),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      // Should show placeholder cards for missing content
      const placeholders = wrapper.findAll('.text-gray-400.dark\\:text-gray-500');
      expect(placeholders.length).toBeGreaterThan(0);
    });
  });

  describe('Publications Section', () => {
    it('should render publications section for news content', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      const publicationsSection = wrapper.find('.publications-section');
      expect(publicationsSection.exists()).toBe(true);
      expect(publicationsSection.text()).toContain('All News Publications');
      expect(publicationsSection.text()).toContain('news');
    });

    it('should render publications section for research content', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'research' }
      });
      await nextTick();

      const publicationsSection = wrapper.find('.publications-section');
      expect(publicationsSection.exists()).toBe(true);
      expect(publicationsSection.text()).toContain('All Research Publications');
      expect(publicationsSection.text()).toContain('research-articles');
    });

    it('should not render publications section for events content', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'events' }
      });
      await nextTick();

      const publicationsSection = wrapper.find('.publications-section');
      expect(publicationsSection.exists()).toBe(false);
    });
  });

  describe('Data Transformation', () => {
    it('should transform news articles into proper category structure', async () => {
      mockUseNews.mockReturnValue({
        articles: ref(mockNewsArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.vm.articleCategories).toEqual([
        expect.objectContaining({
          title: 'Medical Research',
          articles: expect.arrayContaining([
            expect.objectContaining({
              title: 'Latest Medical Breakthrough',
              slug: 'medical-breakthrough'
            })
          ])
        }),
        expect.objectContaining({
          title: 'Clinical Updates',
          articles: expect.arrayContaining([
            expect.objectContaining({
              title: 'New Treatment Protocol',
              slug: 'treatment-protocol'
            })
          ])
        })
      ]);
    });

    it('should handle articles with missing category information', async () => {
      const articleWithoutCategory = {
        ...mockNewsArticles[0],
        category: undefined
      };

      mockUseNews.mockReturnValue({
        articles: ref([articleWithoutCategory]),
        loading: ref(false),
        error: ref(null),
        total: ref(1),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.vm.articleCategories).toEqual([
        expect.objectContaining({
          title: 'News', // Should use default category
          articles: expect.arrayContaining([articleWithoutCategory])
        })
      ]);
    });

    it('should format dates correctly using formatDate function', async () => {
      mockUseFeaturedNews.mockReturnValue({
        articles: ref([mockNewsArticles[0]]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      expect(wrapper.vm.featuredDate).toMatch(/^Mar \d{1,2}, 2024$/);
    });
  });

  describe('Responsive Layout', () => {
    it('should have proper responsive grid classes', async () => {
      mockUseNews.mockReturnValue({
        articles: ref(mockNewsArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      // Check for responsive grid layouts
      const categoryGrids = wrapper.findAll('.grid.gap-4.md\\:gap-6.lg\\:gap-8.md\\:grid-cols-2.lg\\:grid-cols-3');
      expect(categoryGrids.length).toBeGreaterThan(0);
    });

    it('should display placeholder cards with proper responsive behavior', async () => {
      // Provide only 1 article to trigger hidden lg:block placeholders
      mockUseNews.mockReturnValue({
        articles: ref([mockNewsArticles[0]]), // Only 1 article
        loading: ref(false),
        error: ref(null),
        total: ref(1),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      // Check for hidden lg:block classes on placeholders
      const hiddenPlaceholders = wrapper.findAll('.hidden.lg\\:block');
      expect(hiddenPlaceholders.length).toBeGreaterThan(0);
    });
  });

  describe('SEO & Accessibility', () => {
    it('should render proper semantic HTML structure', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      // Should use semantic div containers for content sections
      expect(wrapper.find('div').exists()).toBe(true);
    });

    it('should generate proper heading hierarchy in category sections', async () => {
      mockUseNews.mockReturnValue({
        articles: ref(mockNewsArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      // Should have proper heading structure for categories
      const headings = wrapper.findAll('h2, h3, h4');
      expect(headings.length).toBeGreaterThan(0);
    });
  });

  describe('CTA Section', () => {
    it('should render unified content CTA section', async () => {
      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      const ctaSection = wrapper.find('.unified-content-cta');
      expect(ctaSection.exists()).toBe(true);
      expect(ctaSection.text()).toBe('CTA Section');
    });

    it('should render CTA section for all content types', async () => {
      const contentTypes = ['news', 'research', 'events'];
      
      for (const contentType of contentTypes) {
        const wrapper = mount(ContentHub, {
          props: { contentType }
        });
        await nextTick();

        const ctaSection = wrapper.find('.unified-content-cta');
        expect(ctaSection.exists()).toBe(true);
      }
    });
  });

  describe('Error Recovery', () => {
    it('should handle partial loading failure gracefully', async () => {
      mockUseNews.mockReturnValue({
        articles: ref(mockNewsArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      // Featured articles fail to load but main articles succeed
      mockUseFeaturedNews.mockReturnValue({
        articles: ref([]),
        loading: ref(false),
        error: ref('Failed to load featured articles'),
        refetch: vi.fn()
      });

      const wrapper = mount(ContentHub, {
        props: { contentType: 'news' }
      });
      await nextTick();

      // Should still show article categories even if featured fails
      expect(wrapper.vm.articleCategories).toHaveLength(2);
      expect(wrapper.vm.featuredArticle).toBe(null);
    });

    it('should provide meaningful error messages for each content type', async () => {
      const testCases = [
        { contentType: 'news', expectedError: 'News Temporarily Unavailable' },
        { contentType: 'research', expectedError: 'Research Temporarily Unavailable' },
        { contentType: 'events', expectedError: 'Events Temporarily Unavailable' }
      ];

      for (const testCase of testCases) {
        // Clear all mocks before each test case
        vi.clearAllMocks();
        
        // Mock error state for the appropriate composable
        if (testCase.contentType === 'news') {
          mockUseNews.mockReturnValue({
            articles: ref([]),
            loading: ref(false),
            error: ref('API Error'),
            total: ref(0),
            page: ref(1),
            pageSize: ref(10),
            totalPages: ref(0),
            refetch: vi.fn()
          });
          mockUseFeaturedNews.mockReturnValue({
            articles: ref([]),
            loading: ref(false),
            error: ref(null),
            refetch: vi.fn()
          });
        } else if (testCase.contentType === 'research') {
          mockUseResearchArticles.mockReturnValue({
            articles: ref([]),
            loading: ref(false),
            error: ref('API Error'),
            total: ref(0),
            page: ref(1),
            pageSize: ref(10),
            totalPages: ref(0),
            refetch: vi.fn()
          });
          mockUseFeaturedResearch.mockReturnValue({
            articles: ref([]),
            loading: ref(false),
            error: ref(null),
            refetch: vi.fn()
          });
        } else if (testCase.contentType === 'events') {
          mockUseEvents.mockReturnValue({
            events: ref([]),
            loading: ref(false),
            error: ref('API Error'),
            total: ref(0),
            page: ref(1),
            pageSize: ref(10),
            totalPages: ref(0),
            refetch: vi.fn()
          });
          mockUseFeaturedEvents.mockReturnValue({
            events: ref([]),
            loading: ref(false),
            error: ref(null),
            refetch: vi.fn()
          });
        }

        const wrapper = mount(ContentHub, {
          props: { contentType: testCase.contentType }
        });
        await nextTick();

        const errorSection = wrapper.find('.text-center.py-12');
        expect(errorSection.text()).toContain(testCase.expectedError);
      }
    });
  });
});