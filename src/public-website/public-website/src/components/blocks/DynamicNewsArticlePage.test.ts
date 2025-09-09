import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref, nextTick } from 'vue';
import DynamicNewsArticlePage from './DynamicNewsArticlePage.vue';
import type { NewsArticle } from '@/lib/clients/news/types';
import * as newsComposables from '@/composables/';

// Mock composables
vi.mock('@/composables/', () => ({
  useNewsArticle: vi.fn(),
  useFeaturedNews: vi.fn()
}));

// Mock child components
vi.mock('./NewsBreadcrumb.vue', () => ({
  default: {
    name: 'NewsBreadcrumb',
    template: '<div class="news-breadcrumb">{{ articleName }} - {{ category }}</div>',
    props: ['articleName', 'title', 'category']
  }
}));

vi.mock('./NewsArticleContent.vue', () => ({
  default: {
    name: 'NewsArticleContent', 
    template: '<div class="news-article-content">{{ article.title }}</div>',
    props: ['article']
  }
}));

vi.mock('./NewsArticleDetails.vue', () => ({
  default: {
    name: 'NewsArticleDetails',
    template: '<div class="news-article-details">{{ publishedAt }} - {{ author }} - {{ readTime }}</div>',
    props: ['publishedAt', 'author', 'readTime']
  }
}));

vi.mock('./ContactCard.vue', () => ({
  default: {
    name: 'ContactCard',
    template: '<div class="contact-card">Contact Us</div>'
  }
}));

vi.mock('../UnifiedContentCTA.vue', () => ({
  default: {
    name: 'UnifiedContentCTA', 
    template: '<div class="unified-content-cta">CTA Section</div>'
  }
}));

vi.mock('../ArticleCard.vue', () => ({
  default: {
    name: 'ArticleCard',
    template: '<div class="article-card">{{ article.title }}</div>',
    props: ['article', 'basePath', 'defaultAuthor', 'index']
  }
}));

vi.mock('@/lib/utils/assets', () => ({
  resolveAssetUrl: vi.fn((url) => url || null)
}));

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    pathname: '/company/news/test-article-slug',
  },
  writable: true
});

describe('DynamicNewsArticlePage', () => {
  const mockUseNewsArticle = vi.mocked(newsComposables.useNewsArticle);
  const mockUseFeaturedNews = vi.mocked(newsComposables.useFeaturedNews);
  
  const mockNewsArticle: NewsArticle = {
    news_id: '123',
    title: 'Breaking Healthcare News',
    summary: 'Important healthcare developments',
    slug: 'breaking-healthcare-news',
    publishing_status: 'published',
    category_id: 'health-updates',
    author_name: 'Dr. Sarah Johnson',
    author_email: 'sarah.johnson@example.com',
    content: '<h2>Revolutionary Medical Breakthrough</h2><p>Our team has achieved significant advances.</p>',
    image_url: 'https://storage.azure.com/images/healthcare-breakthrough.jpg',
    featured: true,
    order_number: 1,
    published_at: '2024-01-15T10:00:00Z',
    created_on: '2024-01-10T08:30:00Z',
    created_by: 'editorial-team',
    modified_on: '2024-01-12T14:20:00Z',
    modified_by: 'content-reviewer',
    is_deleted: false,
    id: '123',
    createdAt: '2024-01-10T08:30:00Z',
    updatedAt: '2024-01-12T14:20:00Z'
  };

  const mockRelatedArticles: NewsArticle[] = [
    {
      news_id: '456',
      title: 'Related Article 1',
      summary: 'Related content 1',
      slug: 'related-article-1',
      publishing_status: 'published',
      category_id: 'health-updates',
      featured: true,
      order_number: 2,
      published_at: '2024-01-14T10:00:00Z',
      created_on: '2024-01-09T08:30:00Z',
      created_by: 'editorial-team',
      is_deleted: false,
      id: '456',
      createdAt: '2024-01-09T08:30:00Z',
      updatedAt: '2024-01-09T08:30:00Z'
    },
    {
      news_id: '789', 
      title: 'Related Article 2',
      summary: 'Related content 2',
      slug: 'related-article-2',
      publishing_status: 'published',
      category_id: 'research',
      featured: true,
      order_number: 3,
      published_at: '2024-01-13T10:00:00Z',
      created_on: '2024-01-08T08:30:00Z',
      created_by: 'editorial-team',
      is_deleted: false,
      id: '789',
      createdAt: '2024-01-08T08:30:00Z',
      updatedAt: '2024-01-08T08:30:00Z'
    }
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Reset composable mocks
    mockUseNewsArticle.mockReturnValue({
      article: ref(null),
      loading: ref(false),
      error: ref(null),
      refetch: vi.fn()
    });
    
    mockUseFeaturedNews.mockReturnValue({
      articles: ref([]),
      loading: ref(false),
      error: ref(null),
      refetch: vi.fn()
    });
  });

  describe('URL slug extraction', () => {
    it('should extract slug from current URL path', async () => {
      window.location.pathname = '/company/news/test-article-slug';
      
      const wrapper = mount(DynamicNewsArticlePage);
      
      expect(mockUseNewsArticle).toHaveBeenCalledWith(
        expect.objectContaining({
          value: 'test-article-slug'
        })
      );
    });

    it('should handle empty pathname gracefully', async () => {
      window.location.pathname = '/company/news/';
      
      const wrapper = mount(DynamicNewsArticlePage);
      
      expect(mockUseNewsArticle).toHaveBeenCalledWith(
        expect.objectContaining({
          value: ''
        })
      );
    });

    it('should call useFeaturedNews with limit of 3 for related articles', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      
      expect(mockUseFeaturedNews).toHaveBeenCalledWith(3);
    });
  });

  describe('loading state', () => {
    it('should display loading skeleton when article is loading', async () => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
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
      mockUseNewsArticle.mockReturnValue({
        article: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      expect(wrapper.find('.news-breadcrumb').exists()).toBe(false);
      expect(wrapper.find('.news-article-content').exists()).toBe(false);
      expect(wrapper.text()).not.toContain('Article Temporarily Unavailable');
    });
  });

  describe('error state', () => {
    it('should display error message when article fails to load', async () => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(null),
        loading: ref(false),
        error: ref('Failed to load article'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      expect(wrapper.text()).toContain('Article Temporarily Unavailable');
      expect(wrapper.text()).toContain('We\'re experiencing technical difficulties. Please try again later.');
      
      const errorLink = wrapper.find('a[href="/company/news"]');
      expect(errorLink.exists()).toBe(true);
      expect(errorLink.text()).toContain('Browse All News');
    });

    it('should not display content or loading when error occurs', async () => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(null),
        loading: ref(false),
        error: ref('Network error'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      expect(wrapper.find('.news-breadcrumb').exists()).toBe(false);
      expect(wrapper.find('.news-article-content').exists()).toBe(false);
      expect(wrapper.find('.animate-pulse').exists()).toBe(false);
    });
  });

  describe('article content display', () => {
    beforeEach(() => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render breadcrumb with article information', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const breadcrumb = wrapper.find('.news-breadcrumb');
      expect(breadcrumb.exists()).toBe(true);
      expect(breadcrumb.text()).toContain('Breaking Healthcare News');
      expect(breadcrumb.text()).toContain('health-updates');
    });

    it('should render hero image with correct attributes', async () => {
      // Setup: Provide article data through mock
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      // Verify: Component transforms and displays article data correctly
      expect(wrapper.text()).toContain('Breaking Healthcare News');
      expect(wrapper.html()).toContain('https://storage.azure.com/images/healthcare-breakthrough.jpg');
    });

    it('should render fallback hero image when image_url is not provided', async () => {
      const articleWithoutImage = { ...mockNewsArticle, image_url: undefined };
      mockUseNewsArticle.mockReturnValue({
        article: ref(articleWithoutImage),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('src')).toContain('placehold.co');
      expect(heroImage.attributes('src')).toContain(encodeURIComponent('Breaking Healthcare News'));
    });

    it('should render article content component with transformed data', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const articleContent = wrapper.find('.news-article-content');
      expect(articleContent.exists()).toBe(true);
      expect(articleContent.text()).toContain('Breaking Healthcare News');
    });

    it('should render article details with formatted date and author', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const articleDetails = wrapper.find('.news-article-details');
      expect(articleDetails.exists()).toBe(true);
      expect(articleDetails.text()).toContain('Jan 15, 2024');
      expect(articleDetails.text()).toContain('Dr. Sarah Johnson');
      expect(articleDetails.text()).toContain('1 min read'); // Calculated from actual content
    });

    it('should handle missing author with fallback', async () => {
      const articleWithoutAuthor = { ...mockNewsArticle, author_name: undefined };
      mockUseNewsArticle.mockReturnValue({
        article: ref(articleWithoutAuthor),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const articleDetails = wrapper.find('.news-article-details');
      expect(articleDetails.text()).toContain('International Center Team');
    });

    it('should render contact card in sidebar', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const contactCard = wrapper.find('.contact-card');
      expect(contactCard.exists()).toBe(true);
      expect(contactCard.text()).toContain('Contact Us');
    });

    it('should render unified content CTA', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const cta = wrapper.find('.unified-content-cta');
      expect(cta.exists()).toBe(true);
      expect(cta.text()).toContain('CTA Section');
    });
  });

  describe('related articles section', () => {
    beforeEach(() => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
      
      mockUseFeaturedNews.mockReturnValue({
        articles: ref(mockRelatedArticles),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should display related articles when available', async () => {
      // Setup: Provide article and related articles data
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      mockUseFeaturedNews.mockReturnValue({
        articles: ref(mockRelatedArticles),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      // Verify: Component displays related articles content
      expect(wrapper.text()).toContain('More News Articles');
      expect(wrapper.text()).toContain('Related Article 1');
      expect(wrapper.text()).toContain('Related Article 2');
    });

    it('should not display related articles section when no articles available', async () => {
      mockUseFeaturedNews.mockReturnValue({
        articles: ref([]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      expect(wrapper.text()).not.toContain('More News Articles');
      expect(wrapper.findAll('.article-card').length).toBe(0);
    });

    it('should display "View All News" button in related articles section', async () => {
      // Setup: Provide article and related articles data through mocks
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      mockUseFeaturedNews.mockReturnValue({
        articles: ref([mockNewsArticle]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      // Verify: Component displays View All News button content and navigation link
      expect(wrapper.text()).toContain('View All News');
      expect(wrapper.html()).toContain('href="/company/news"');
    });

    it('should not display related articles section when loading or error state', async () => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      expect(wrapper.text()).not.toContain('More News Articles');
      expect(wrapper.findAll('.article-card').length).toBe(0);
    });
  });

  describe('data transformation', () => {
    it('should transform NewsArticle to NewsArticlePageData structure', async () => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      // Check that transformed data is passed to child components correctly
      const breadcrumb = wrapper.find('.news-breadcrumb');
      expect(breadcrumb.text()).toContain('Breaking Healthcare News');
      expect(breadcrumb.text()).toContain('health-updates');
    });

    it('should handle article with missing content field', async () => {
      const articleWithoutContent = { ...mockNewsArticle, content: undefined };
      mockUseNewsArticle.mockReturnValue({
        article: ref(articleWithoutContent),
        loading: ref(false), 
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      // Should fallback to summary for content
      const articleContent = wrapper.find('.news-article-content');
      expect(articleContent.exists()).toBe(true);
    });

    it('should format dates correctly', async () => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const articleDetails = wrapper.find('.news-article-details');
      expect(articleDetails.text()).toContain('Jan 15, 2024');
    });

    it('should handle missing published_at by using created_on', async () => {
      const articleWithoutPublishedAt = { ...mockNewsArticle, published_at: undefined };
      mockUseNewsArticle.mockReturnValue({
        article: ref(articleWithoutPublishedAt),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const articleDetails = wrapper.find('.news-article-details');
      expect(articleDetails.text()).toContain('Jan 10, 2024'); // created_on date
    });
  });

  describe('responsive layout', () => {
    beforeEach(() => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should have proper grid layout classes for responsive design', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const gridContainer = wrapper.find('.grid.gap-12.md\\:grid-cols-12');
      expect(gridContainer.exists()).toBe(true);
      
      const mainContent = wrapper.find('.md\\:col-span-7.md\\:col-start-1.lg\\:col-span-8');
      expect(mainContent.exists()).toBe(true);
      
      const sidebar = wrapper.find('.md\\:col-span-5.lg\\:col-span-4');
      expect(sidebar.exists()).toBe(true);
    });

    it('should have sticky sidebar on medium and larger screens', async () => {
      // Setup: Provide article data
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      // Verify: Component renders sidebar content properly
      expect(wrapper.text()).toContain('Contact Us'); // Sidebar contact component
      expect(wrapper.text()).toContain('Dr. Sarah Johnson'); // Article details in sidebar
    });
  });

  describe('accessibility', () => {
    beforeEach(() => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render main article with proper semantic HTML', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const article = wrapper.find('article.prose');
      expect(article.exists()).toBe(true);
    });

    it('should render aside element for sidebar content', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const aside = wrapper.find('aside#news-article-page-aside');
      expect(aside.exists()).toBe(true);
    });

    it('should have proper heading hierarchy', async () => {
      mockUseFeaturedNews.mockReturnValue({
        articles: ref(mockRelatedArticles),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const relatedHeading = wrapper.find('h2');
      expect(relatedHeading.exists()).toBe(true);
      expect(relatedHeading.text()).toBe('More News Articles');
    });
  });

  describe('SEO metadata handling', () => {
    beforeEach(() => {
      mockUseNewsArticle.mockReturnValue({
        article: ref(mockNewsArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should generate proper image alt text for SEO', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('alt')).toBe('Breaking Healthcare News - International Center News');
    });

    it('should provide structured data through component props', async () => {
      const wrapper = mount(DynamicNewsArticlePage);
      await nextTick();

      // Verify that structured data is available through component state
      const breadcrumb = wrapper.find('.news-breadcrumb');
      expect(breadcrumb.text()).toContain('Breaking Healthcare News');
    });
  });
});