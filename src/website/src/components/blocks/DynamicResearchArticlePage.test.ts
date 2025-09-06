import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref, nextTick } from 'vue';
import DynamicResearchArticlePage from './DynamicResearchArticlePage.vue';
import type { ResearchArticle } from '@/lib/clients/research/types';

// Mock composables
vi.mock('@/composables/useResearch');

// Mock child components
vi.mock('./ResearchBreadcrumb.vue', () => ({
  default: {
    name: 'ResearchBreadcrumb',
    template: '<div class="research-breadcrumb">{{ articleName }} - {{ category }}</div>',
    props: ['articleName', 'title', 'category']
  }
}));

vi.mock('./ResearchArticleContent.vue', () => ({
  default: {
    name: 'ResearchArticleContent',
    template: '<div class="research-content">{{ article.title }}</div>',
    props: ['article']
  }
}));

vi.mock('./ResearchArticleDetails.vue', () => ({
  default: {
    name: 'ResearchArticleDetails',
    template: '<div class="research-details">{{ publishedAt }} - {{ author }} - {{ readTime }}</div>',
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

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    pathname: '/community/research/regenerative-medicine-study',
  },
  writable: true
});

// Import the mocked functions 
import { useResearchArticle, useResearchArticles } from '@/composables/useResearch';

describe('DynamicResearchArticlePage', () => {
  
  // Get mocked functions
  const mockUseResearchArticle = vi.mocked(useResearchArticle);
  const mockUseResearchArticles = vi.mocked(useResearchArticles);
  
  const mockResearchArticle: ResearchArticle = {
    id: '660e8400-e29b-41d4-a716-446655440001',
    title: 'Regenerative Medicine Study',
    slug: 'regenerative-medicine-study',
    excerpt: 'Comprehensive study on regenerative medicine applications',
    content: '<h2>Study Overview</h2><p>This study examines the effectiveness of regenerative medicine treatments.</p>',
    featured_image: 'https://storage.azure.com/images/regenerative-study.jpg',
    author: 'Dr. Sarah Johnson',
    tags: ['regenerative-medicine', 'clinical-study', 'treatment'],
    status: 'published',
    featured: true,
    category: 'Clinical Research',
    client_name: 'International Center',
    industry: 'Healthcare',
    challenge: 'Evaluating treatment effectiveness',
    solution: 'Controlled clinical trial methodology',
    results: 'Significant improvement in patient outcomes',
    technologies: ['stem-cell-therapy', 'growth-factors', 'tissue-engineering'],
    gallery_images: ['https://storage.azure.com/images/study-1.jpg', 'https://storage.azure.com/images/study-2.jpg'],
    meta_title: 'Regenerative Medicine Study - International Center Research',
    meta_description: 'Comprehensive clinical study examining regenerative medicine effectiveness',
    published_at: '2024-02-15T10:00:00Z',
    createdAt: '2024-02-01T08:00:00Z',
    updatedAt: '2024-02-05T12:00:00Z'
  };

  const mockRelatedArticles: ResearchArticle[] = [
    {
      id: '660e8400-e29b-41d4-a716-446655440002',
      title: 'Stem Cell Therapy Applications',
      slug: 'stem-cell-therapy-applications',
      excerpt: 'Exploring various applications of stem cell therapy',
      content: '<p>Research on stem cell applications in medicine.</p>',
      featured_image: '',
      author: 'Dr. Michael Chen',
      tags: ['stem-cells', 'therapy'],
      status: 'published',
      featured: false,
      category: 'Clinical Research',
      technologies: ['stem-cells'],
      gallery_images: [],
      meta_title: 'Stem Cell Therapy Applications',
      meta_description: 'Research on stem cell therapy applications',
      published_at: '2024-02-10T09:00:00Z',
      createdAt: '2024-02-01T08:00:00Z',
      updatedAt: '2024-02-05T12:00:00Z'
    },
    {
      id: '660e8400-e29b-41d4-a716-446655440003',
      title: 'Tissue Engineering Advances',
      slug: 'tissue-engineering-advances',
      excerpt: 'Latest advances in tissue engineering research',
      content: '<p>Breakthrough developments in tissue engineering.</p>',
      featured_image: '',
      author: 'Dr. Emily Rodriguez',
      tags: ['tissue-engineering', 'research'],
      status: 'published',
      featured: false,
      category: 'Clinical Research',
      technologies: ['tissue-engineering'],
      gallery_images: [],
      meta_title: 'Tissue Engineering Advances',
      meta_description: 'Latest advances in tissue engineering research',
      published_at: '2024-02-08T14:00:00Z',
      createdAt: '2024-02-01T08:00:00Z',
      updatedAt: '2024-02-05T12:00:00Z'
    }
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Reset composable mocks
    mockUseResearchArticle.mockReturnValue({
      article: ref(null),
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
  });

  describe('URL slug extraction', () => {
    it('should extract slug from current URL path', async () => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();
      
      expect(mockUseResearchArticle).toHaveBeenCalledWith(
        expect.objectContaining({
          value: 'regenerative-medicine-study'
        })
      );
    });

    it('should handle empty pathname gracefully', async () => {
      window.location.pathname = '/community/research/';
      
      const wrapper = mount(DynamicResearchArticlePage);
      
      expect(mockUseResearchArticle).toHaveBeenCalledWith(
        expect.objectContaining({
          value: ''
        })
      );
    });

    it('should call useResearchArticles for related articles', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      
      expect(mockUseResearchArticles).toHaveBeenCalled();
    });
  });

  describe('loading state', () => {
    it('should display loading skeleton when article is loading', async () => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
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
      mockUseResearchArticle.mockReturnValue({
        article: ref(null),
        loading: ref(true),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      expect(wrapper.find('.research-breadcrumb').exists()).toBe(false);
      expect(wrapper.find('.research-content').exists()).toBe(false);
      expect(wrapper.find('.text-center.py-12').exists()).toBe(false);
    });
  });

  describe('error state', () => {
    it('should display error message when article fails to load', async () => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(null),
        loading: ref(false),
        error: ref('Research article not found'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const errorSection = wrapper.find('.text-center.py-12');
      expect(errorSection.exists()).toBe(true);
      expect(errorSection.text()).toContain('Research Article Temporarily Unavailable');
      expect(errorSection.text()).toContain('We\'re experiencing technical difficulties');
      
      const browseLink = errorSection.find('a[href="/community/research"]');
      expect(browseLink.exists()).toBe(true);
      expect(browseLink.text()).toBe('Browse All Research');
    });

    it('should not display content or loading when error occurs', async () => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(null),
        loading: ref(false),
        error: ref('Research article not found'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      expect(wrapper.find('.research-breadcrumb').exists()).toBe(false);
      expect(wrapper.find('.animate-pulse').exists()).toBe(false);
      expect(wrapper.find('.research-content').exists()).toBe(false);
    });
  });

  describe('article content display', () => {
    beforeEach(() => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render breadcrumb with article information', async () => {
      // Setup: Provide research article data through mocks
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      mockUseResearchArticles.mockReturnValue({
        articles: ref([mockResearchArticle]),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Verify: Component displays breadcrumb information correctly
      expect(wrapper.text()).toContain('Regenerative Medicine Study');
      expect(wrapper.text()).toContain('Clinical Research');
    });

    it('should render hero image with correct attributes', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('src')).toBe('https://storage.azure.com/images/regenerative-study.jpg');
      expect(heroImage.attributes('alt')).toContain('Regenerative Medicine Study');
      expect(heroImage.classes()).toContain('aspect-video');
    });

    it('should render fallback hero image when featured_image is not provided', async () => {
      const articleWithoutImage = { ...mockResearchArticle, featured_image: '' };
      mockUseResearchArticle.mockReturnValue({
        article: ref(articleWithoutImage),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const heroImage = wrapper.find('img');
      expect(heroImage.attributes('src')).toContain('placehold.co');
      expect(heroImage.attributes('src')).toContain(encodeURIComponent('Regenerative Medicine Study'));
    });

    it('should render research content component with transformed data', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const researchContent = wrapper.find('.research-content');
      expect(researchContent.exists()).toBe(true);
      expect(researchContent.text()).toContain('Regenerative Medicine Study');
    });

    it('should render research details with all information', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const researchDetails = wrapper.find('.research-details');
      expect(researchDetails.exists()).toBe(true);
      expect(researchDetails.text()).toContain('Dr. Sarah Johnson');
    });

    it('should render contact card in sidebar', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const contactCard = wrapper.find('.contact-card');
      expect(contactCard.exists()).toBe(true);
      expect(contactCard.text()).toBe('Contact Us');
    });
  });

  describe('related articles section', () => {
    beforeEach(() => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should display related articles section when articles are available', async () => {
      mockUseResearchArticles.mockReturnValue({
        articles: ref(mockRelatedArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const relatedSection = wrapper.find('.pt-16.lg\\:pt-20.pb-8.lg\\:pb-12');
      expect(relatedSection.exists()).toBe(true);
      
      const sectionTitle = relatedSection.find('h2');
      expect(sectionTitle.text()).toContain('More from Clinical Research');
    });

    it('should render related article cards', async () => {
      mockUseResearchArticles.mockReturnValue({
        articles: ref(mockRelatedArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const articleCards = wrapper.findAll('.article-card');
      expect(articleCards).toHaveLength(2);
      expect(articleCards[0].text()).toContain('Stem Cell Therapy Applications');
      expect(articleCards[1].text()).toContain('Tissue Engineering Advances');
    });

    it('should not display related articles section when no articles available', async () => {
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

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const relatedSection = wrapper.find('.pt-16.lg\\:pt-20.pb-8.lg\\:pb-12');
      expect(relatedSection.exists()).toBe(false);
    });

    it('should handle related articles loading failure gracefully', async () => {
      mockUseResearchArticles.mockReturnValue({
        articles: ref([]),
        loading: ref(false),
        error: ref('Failed to load related articles'),
        total: ref(0),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(0),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Should not crash and should not show related articles section
      const relatedSection = wrapper.find('.pt-16.lg\\:pt-20.pb-8.lg\\:pb-12');
      expect(relatedSection.exists()).toBe(false);
      
      // Main content should still be visible
      const researchContent = wrapper.find('.research-content');
      expect(researchContent.exists()).toBe(true);
    });
  });

  describe('data transformation', () => {
    it('should transform ResearchArticle to ResearchArticlePageData structure', async () => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Test that the transformed data structure is passed to child components
      expect(wrapper.vm.articleData).toEqual(
        expect.objectContaining({
          id: mockResearchArticle.id,
          title: mockResearchArticle.title,
          slug: mockResearchArticle.slug,
          description: mockResearchArticle.excerpt,
          content: mockResearchArticle.content,
          heroImage: expect.objectContaining({
            src: mockResearchArticle.featured_image,
            alt: expect.stringContaining('Regenerative Medicine Study')
          }),
          articleDetails: expect.objectContaining({
            publishedAt: expect.any(String),
            author: 'Dr. Sarah Johnson',
            readTime: expect.any(String)
          }),
          category: 'Clinical Research'
        })
      );
    });

    it('should handle article with missing content field', async () => {
      const articleWithoutContent = { ...mockResearchArticle, content: '' };
      mockUseResearchArticle.mockReturnValue({
        article: ref(articleWithoutContent),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      expect(wrapper.vm.articleData.content).toBe(mockResearchArticle.excerpt);
    });

    it('should handle article without category gracefully', async () => {
      const articleWithoutCategory = { ...mockResearchArticle, category: '' };
      mockUseResearchArticle.mockReturnValue({
        article: ref(articleWithoutCategory),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      expect(wrapper.vm.articleData.category).toBe('');
      expect(wrapper.vm.relatedArticlesTitle).toBe('More Research Articles');
    });

    it('should handle reading time calculation correctly', async () => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Should calculate reading time or use default
      expect(wrapper.vm.articleData.articleDetails.readTime).toMatch(/\d+ min read/);
    });
  });

  describe('responsive layout', () => {
    beforeEach(() => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should have proper grid layout classes for responsive design', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const gridContainer = wrapper.find('.mt-8.grid.gap-12.md\\:grid-cols-12.md\\:gap-8');
      expect(gridContainer.exists()).toBe(true);
      
      const mainContent = wrapper.find('.order-2.md\\:order-none.md\\:col-span-7.md\\:col-start-1.lg\\:col-span-8');
      expect(mainContent.exists()).toBe(true);
      
      const sidebar = wrapper.find('.order-1.md\\:order-none.md\\:col-span-5.lg\\:col-span-4');
      expect(sidebar.exists()).toBe(true);
    });

    it('should have sticky sidebar on medium and larger screens', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const stickySidebar = wrapper.find('.md\\:sticky.md\\:top-20');
      expect(stickySidebar.exists()).toBe(true);
    });

    it('should display related articles in responsive grid', async () => {
      mockUseResearchArticles.mockReturnValue({
        articles: ref(mockRelatedArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const relatedGrid = wrapper.find('.grid.gap-4.md\\:gap-6.lg\\:gap-8.md\\:grid-cols-2.lg\\:grid-cols-3');
      expect(relatedGrid.exists()).toBe(true);
    });
  });

  describe('accessibility', () => {
    beforeEach(() => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render main article with proper semantic HTML', async () => {
      // Arrange: Provide research article data
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Assert: Component renders semantic article element for main content
      const article = wrapper.find('article');
      expect(article.exists()).toBe(true);
      
      // Assert: Component displays research article content
      expect(wrapper.text()).toContain('Regenerative Medicine Study');
    });

    it('should render aside element for sidebar content', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const aside = wrapper.find('aside#research-article-page-aside');
      expect(aside.exists()).toBe(true);
    });

    it('should have proper image alt text for screen readers', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const heroImage = wrapper.find('img');
      const altText = heroImage.attributes('alt');
      expect(altText).toContain('Regenerative Medicine Study');
      expect(altText).toContain('International Center Research');
    });

    it('should have proper heading hierarchy', async () => {
      mockUseResearchArticles.mockReturnValue({
        articles: ref(mockRelatedArticles),
        loading: ref(false),
        error: ref(null),
        total: ref(2),
        page: ref(1),
        pageSize: ref(10),
        totalPages: ref(1),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const relatedSectionHeading = wrapper.find('h2');
      expect(relatedSectionHeading.exists()).toBe(true);
      expect(relatedSectionHeading.text()).toContain('More from Clinical Research');
    });
  });

  describe('SEO metadata handling', () => {
    beforeEach(() => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should generate proper image alt text for SEO', async () => {
      // Arrange: Provide research article data for image rendering
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      // Act: Mount component
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Assert: Hero image has proper alt text for SEO
      const heroImage = wrapper.find('img');
      const altText = heroImage.attributes('alt');
      expect(altText).toBe('Regenerative Medicine Study - International Center Research');
    });

    it('should provide structured data through component props', async () => {
      // Setup: Provide research article data through mocks
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Verify: Component displays structured article data correctly
      expect(wrapper.text()).toContain('Regenerative Medicine Study');
      expect(wrapper.text()).toContain('Clinical Research');
      expect(wrapper.text()).toContain('Dr. Sarah Johnson');
    });
  });

  describe('error recovery', () => {
    it('should handle article loading failure gracefully', async () => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(null),
        loading: ref(false),
        error: ref('Network error'),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      expect(wrapper.find('.text-center.py-12').exists()).toBe(true);
      expect(wrapper.find('a[href="/community/research"]').exists()).toBe(true);
    });

    it('should handle article without published_at', async () => {
      const articleWithoutDate = { ...mockResearchArticle, published_at: '' };
      mockUseResearchArticle.mockReturnValue({
        article: ref(articleWithoutDate),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Should not crash when published_at is missing
      const researchDetails = wrapper.find('.research-details');
      expect(researchDetails.exists()).toBe(true);
    });

    it('should handle article without author', async () => {
      const articleWithoutAuthor = { ...mockResearchArticle, author: '' };
      mockUseResearchArticle.mockReturnValue({
        article: ref(articleWithoutAuthor),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Should use default author
      expect(wrapper.vm.articleData.articleDetails.author).toBe('International Center Research Team');
    });
  });

  describe('article status handling', () => {
    it('should handle published articles correctly', async () => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const researchContent = wrapper.find('.research-content');
      expect(researchContent.exists()).toBe(true);
    });

    it('should handle draft articles correctly', async () => {
      const draftArticle = { ...mockResearchArticle, status: 'draft' as const };
      mockUseResearchArticle.mockReturnValue({
        article: ref(draftArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Component should still render draft articles (business logic can restrict visibility elsewhere)
      const researchContent = wrapper.find('.research-content');
      expect(researchContent.exists()).toBe(true);
    });

    it('should handle archived articles correctly', async () => {
      const archivedArticle = { ...mockResearchArticle, status: 'archived' as const };
      mockUseResearchArticle.mockReturnValue({
        article: ref(archivedArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });

      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      // Component should still render archived articles
      const researchContent = wrapper.find('.research-content');
      expect(researchContent.exists()).toBe(true);
    });
  });

  describe('CTA section', () => {
    beforeEach(() => {
      mockUseResearchArticle.mockReturnValue({
        article: ref(mockResearchArticle),
        loading: ref(false),
        error: ref(null),
        refetch: vi.fn()
      });
    });

    it('should render unified content CTA section', async () => {
      const wrapper = mount(DynamicResearchArticlePage);
      await nextTick();

      const ctaSection = wrapper.find('.unified-content-cta');
      expect(ctaSection.exists()).toBe(true);
      expect(ctaSection.text()).toBe('CTA Section');
    });
  });
});