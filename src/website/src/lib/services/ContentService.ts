// Content Service - Business logic for cross-domain content operations
// Orchestrates multiple domain clients to provide unified content functionality

import { servicesClient } from '../clients/services/ServicesClient';
import { newsClient } from '../clients/news/NewsClient';
import { researchClient } from '../clients/research/ResearchClient';
import type { Service } from '../clients/services/types';
import type { NewsArticle } from '../clients/news/types';
import type { ResearchArticle } from '../clients/research/types';

export interface ContentItem {
  id: string;
  title: string;
  slug: string;
  excerpt: string;
  image: string;
  type: 'service' | 'news' | 'research';
  category?: string;
  published_at?: string;
  featured: boolean;
}

export interface FeaturedContent {
  services: Service[];
  news: NewsArticle[];
  research: ResearchArticle[];
}

export interface SearchResults {
  services: Service[];
  news: NewsArticle[];
  research: ResearchArticle[];
  total: number;
}

export class ContentService {
  /**
   * Get all featured content across domains
   */
  async getFeaturedContent(limit: number = 3): Promise<FeaturedContent> {
    try {
      const [services, news, research] = await Promise.allSettled([
        servicesClient.getFeaturedServices(limit),
        newsClient.getFeaturedNews(limit),
        researchClient.getFeaturedResearch(limit),
      ]);

      return {
        services: services.status === 'fulfilled' ? services.value : [],
        news: news.status === 'fulfilled' ? news.value : [],
        research: research.status === 'fulfilled' ? research.value : [],
      };
    } catch (error) {
      console.error('Error fetching featured content:', error);
      return { services: [], news: [], research: [] };
    }
  }

  /**
   * Get recent content across all domains
   */
  async getRecentContent(limit: number = 6): Promise<ContentItem[]> {
    try {
      const [newsResponse, researchResponse] = await Promise.allSettled([
        newsClient.getRecentNews(limit),
        researchClient.getRecentResearch(limit),
      ]);

      const content: ContentItem[] = [];

      // Add news articles
      if (newsResponse.status === 'fulfilled') {
        content.push(
          ...newsResponse.value.map(article => this.transformNewsToContentItem(article))
        );
      }

      // Add research articles
      if (researchResponse.status === 'fulfilled') {
        content.push(
          ...researchResponse.value.map(article => this.transformResearchToContentItem(article))
        );
      }

      // Sort by published date and limit
      return content
        .sort((a, b) => {
          const dateA = a.published_at ? new Date(a.published_at).getTime() : 0;
          const dateB = b.published_at ? new Date(b.published_at).getTime() : 0;
          return dateB - dateA;
        })
        .slice(0, limit);
    } catch (error) {
      console.error('Error fetching recent content:', error);
      return [];
    }
  }

  /**
   * Search across all domains
   */
  async searchContent(
    query: string,
    params: { page?: number; pageSize?: number } = {}
  ): Promise<SearchResults> {
    if (!query.trim()) {
      return { services: [], news: [], research: [], total: 0 };
    }

    try {
      const [servicesResult, newsResult, researchResult] = await Promise.allSettled([
        servicesClient.searchServices({ q: query, ...params }),
        newsClient.searchNewsArticles({ q: query, ...params }),
        // Research domain doesn't have search, so we'll get all and filter client-side
        researchClient.getResearchArticles(params).then(response => ({
          articles: (response.articles || []).filter(
            (article: any) =>
              article.title.toLowerCase().includes(query.toLowerCase()) ||
              article.excerpt.toLowerCase().includes(query.toLowerCase()) ||
              (article.tags || []).some((tag: any) =>
                tag.toLowerCase().includes(query.toLowerCase())
              )
          ),
          total: 0,
          page: 1,
          pageSize: 10,
          totalPages: 1,
        })),
      ]);

      const services = servicesResult.status === 'fulfilled' ? servicesResult.value.data : [];
      const news = newsResult.status === 'fulfilled' ? newsResult.value.data : [];
      const research = researchResult.status === 'fulfilled' ? researchResult.value.articles : [];

      return {
        services,
        news,
        research,
        total: services.length + news.length + research.length,
      };
    } catch (error) {
      console.error('Error searching content:', error);
      return { services: [], news: [], research: [], total: 0 };
    }
  }

  /**
   * Get content by category across domains
   */
  async getContentByCategory(
    category: string,
    params: { page?: number; pageSize?: number } = {}
  ): Promise<SearchResults> {
    try {
      const [newsResult, researchResult] = await Promise.allSettled([
        newsClient.getNewsByCategory(category, params),
        researchClient.getResearchByCategory(category, params),
      ]);

      const news = newsResult.status === 'fulfilled' ? newsResult.value.data : [];
      const research = researchResult.status === 'fulfilled' ? researchResult.value.data : [];

      return {
        services: [], // Services don't have categories in the same way
        news,
        research,
        total: news.length + research.length,
      };
    } catch (error) {
      console.error('Error fetching content by category:', error);
      return { services: [], news: [], research: [], total: 0 };
    }
  }

  /**
   * Get related content for a given item
   */
  async getRelatedContent(item: ContentItem, limit: number = 3): Promise<ContentItem[]> {
    try {
      let relatedContent: ContentItem[] = [];

      // For news articles, get other articles in the same category
      if (item.type === 'news' && item.category) {
        const related = await newsClient.getNewsByCategory(item.category, { pageSize: limit + 1 });
        relatedContent = related.data
          .filter(article => article.id !== item.id)
          .slice(0, limit)
          .map(article => this.transformNewsToContentItem(article));
      }

      // For research articles, get other articles in the same category
      if (item.type === 'research' && item.category) {
        const related = await researchClient.getResearchByCategory(item.category, {
          pageSize: limit + 1,
        });
        relatedContent = related.data
          .filter(article => article.id !== item.id)
          .slice(0, limit)
          .map(article => this.transformResearchToContentItem(article));
      }

      // If we don't have enough related content, fill with recent content
      if (relatedContent.length < limit) {
        const recent = await this.getRecentContent(limit * 2);
        const additional = recent
          .filter(content => content.id !== item.id && content.type !== item.type)
          .slice(0, limit - relatedContent.length);
        relatedContent.push(...additional);
      }

      return relatedContent.slice(0, limit);
    } catch (error) {
      console.error('Error fetching related content:', error);
      return [];
    }
  }

  /**
   * Get all available categories across domains
   */
  async getAllCategories(): Promise<string[]> {
    try {
      const [newsCategories, researchCategories] = await Promise.allSettled([
        Promise.resolve([]), // News categories not implemented yet
        researchClient.getResearchCategories(),
      ]);

      const categories = new Set<string>();

      if (newsCategories.status === 'fulfilled') {
        newsCategories.value.forEach(cat => categories.add(cat));
      }

      if (researchCategories.status === 'fulfilled') {
        researchCategories.value.forEach(cat => categories.add(cat));
      }

      return Array.from(categories).sort();
    } catch (error) {
      console.error('Error fetching categories:', error);
      return [];
    }
  }

  // Helper methods to transform domain objects to ContentItem
  private transformNewsToContentItem(article: NewsArticle): ContentItem {
    return {
      id: article.id,
      title: article.title,
      slug: article.slug,
      excerpt: article.excerpt,
      image: article.featured_image,
      type: 'news',
      category: article.category,
      published_at: article.published_at,
      featured: article.featured,
    };
  }

  private transformResearchToContentItem(article: ResearchArticle): ContentItem {
    return {
      id: article.id,
      title: article.title,
      slug: article.slug,
      excerpt: article.excerpt,
      image: article.featured_image,
      type: 'research',
      category: article.category,
      published_at: article.published_at,
      featured: article.featured,
    };
  }

  private transformServiceToContentItem(service: Service): ContentItem {
    return {
      id: service.id,
      title: service.title,
      slug: service.slug,
      excerpt: service.description,
      image: service.image,
      type: 'service',
      featured: false, // Services don't have featured flag in the same way
    };
  }
}

// Export singleton instance
export const contentService = new ContentService();
