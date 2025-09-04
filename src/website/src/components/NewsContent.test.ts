import { describe, it, expect, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import NewsContent from './NewsContent.vue';
import type { NewsArticle } from '@/lib/clients/news/types';

describe('NewsContent', () => {
  const mockNewsArticle: NewsArticle = {
    news_id: '123',
    title: 'Breaking Healthcare News',
    summary: 'Important healthcare developments',
    slug: 'breaking-healthcare-news',
    publishing_status: 'published',
    category_id: '456',
    author_name: 'Dr. Sarah Johnson',
    author_email: 'sarah.johnson@example.com',
    content: '<h2>Revolutionary Medical Breakthrough</h2><p>Our medical team has achieved significant advances in patient care through:</p><ul><li>Advanced diagnostic techniques</li><li>Personalized treatment plans</li><li>Integrated care coordination</li><li>Patient education programs</li></ul><p>These developments will improve outcomes for thousands of patients.</p>',
    image_url: 'https://storage.azure.com/images/healthcare-breakthrough.jpg',
    featured: true,
    order_number: 1,
    published_at: '2024-01-15T10:00:00Z',
    created_on: '2024-01-10T08:30:00Z',
    created_by: 'editorial-team',
    modified_on: '2024-01-12T14:20:00Z',
    modified_by: 'content-reviewer',
    is_deleted: false,
    id: '123'
  };

  it('should render news article title and summary', () => {
    const wrapper = mount(NewsContent, {
      props: {
        article: mockNewsArticle
      }
    });

    expect(wrapper.find('h1').text()).toBe('Breaking Healthcare News');
    expect(wrapper.find('.news-summary').text()).toBe('Important healthcare developments');
  });

  it('should render inline HTML content from PostgreSQL storage', () => {
    const wrapper = mount(NewsContent, {
      props: {
        article: mockNewsArticle
      }
    });

    const contentElement = wrapper.find('.content-html');
    expect(contentElement.exists()).toBe(true);
    
    // Check that HTML content is rendered
    expect(contentElement.html()).toContain('<h2>Revolutionary Medical Breakthrough</h2>');
    expect(contentElement.html()).toContain('<ul>');
    expect(contentElement.html()).toContain('<li>Advanced diagnostic techniques</li>');
    expect(contentElement.html()).toContain('<li>Personalized treatment plans</li>');
  });

  it('should display author information when provided', () => {
    const wrapper = mount(NewsContent, {
      props: {
        article: mockNewsArticle,
        showAuthor: true
      }
    });

    const authorElement = wrapper.find('.news-author');
    expect(authorElement.exists()).toBe(true);
    expect(authorElement.text()).toContain('Dr. Sarah Johnson');
  });

  it('should display publication date when provided', () => {
    const wrapper = mount(NewsContent, {
      props: {
        article: mockNewsArticle,
        showPublishedDate: true
      }
    });

    const dateElement = wrapper.find('.news-date');
    expect(dateElement.exists()).toBe(true);
    expect(dateElement.text()).toContain('2024');
  });

  it('should display featured badge for featured articles', () => {
    const wrapper = mount(NewsContent, {
      props: {
        article: mockNewsArticle,
        showFeatured: true
      }
    });

    const featuredBadge = wrapper.find('.featured-badge');
    expect(featuredBadge.exists()).toBe(true);
    expect(featuredBadge.text()).toBe('Featured');
  });

  it('should display news image when image_url is provided', () => {
    const wrapper = mount(NewsContent, {
      props: {
        article: mockNewsArticle,
        showImage: true
      }
    });

    const imageElement = wrapper.find('.news-image img');
    expect(imageElement.exists()).toBe(true);
    expect(imageElement.attributes('src')).toBe('https://storage.azure.com/images/healthcare-breakthrough.jpg');
    expect(imageElement.attributes('alt')).toBe('Breaking Healthcare News');
    expect(imageElement.attributes('loading')).toBe('lazy');
  });

  it('should handle articles without content gracefully', () => {
    const articleWithoutContent: NewsArticle = {
      ...mockNewsArticle,
      content: undefined
    };

    const wrapper = mount(NewsContent, {
      props: {
        article: articleWithoutContent
      }
    });

    const contentElement = wrapper.find('.content-html');
    expect(contentElement.exists()).toBe(false);
    
    // Should still render title and summary
    expect(wrapper.find('h1').text()).toBe('Breaking Healthcare News');
    expect(wrapper.find('.news-summary').text()).toBe('Important healthcare developments');
  });

  it('should handle articles without image_url gracefully', () => {
    const articleWithoutImage: NewsArticle = {
      ...mockNewsArticle,
      image_url: undefined
    };

    const wrapper = mount(NewsContent, {
      props: {
        article: articleWithoutImage,
        showImage: true
      }
    });

    const imageElement = wrapper.find('.news-image img');
    expect(imageElement.exists()).toBe(false);
  });

  it('should handle articles without author information gracefully', () => {
    const articleWithoutAuthor: NewsArticle = {
      ...mockNewsArticle,
      author_name: undefined,
      author_email: undefined
    };

    const wrapper = mount(NewsContent, {
      props: {
        article: articleWithoutAuthor,
        showAuthor: true
      }
    });

    const authorElement = wrapper.find('.news-author');
    expect(authorElement.exists()).toBe(false);
  });

  it('should apply proper CSS classes for semantic structure', () => {
    const wrapper = mount(NewsContent, {
      props: {
        article: mockNewsArticle
      }
    });

    expect(wrapper.find('article.news-content').exists()).toBe(true);
    expect(wrapper.find('header.news-header').exists()).toBe(true);
    expect(wrapper.find('.news-title').exists()).toBe(true);
    expect(wrapper.find('.news-summary').exists()).toBe(true);
    expect(wrapper.find('.news-content-body').exists()).toBe(true);
  });

  it('should sanitize HTML content to prevent XSS attacks', () => {
    const maliciousArticle: NewsArticle = {
      ...mockNewsArticle,
      content: '<h2>Safe Content</h2><script>alert("xss")</script><p>More safe content</p>'
    };

    const wrapper = mount(NewsContent, {
      props: {
        article: maliciousArticle
      }
    });

    const contentHtml = wrapper.find('.content-html').html();
    expect(contentHtml).toContain('<h2>Safe Content</h2>');
    expect(contentHtml).toContain('<p>More safe content</p>');
    // Should not contain script tags (this will be implemented in the component)
    expect(contentHtml).not.toContain('<script>');
  });

  it('should provide proper accessibility attributes', () => {
    const wrapper = mount(NewsContent, {
      props: {
        article: mockNewsArticle,
        showImage: true
      }
    });

    const articleElement = wrapper.find('article');
    const imageElement = wrapper.find('img');

    expect(articleElement.attributes('role')).toBe('article');
    expect(imageElement.attributes('alt')).toBe('Breaking Healthcare News');
    expect(imageElement.attributes('loading')).toBe('lazy');
  });
});