// News gRPC Adapter - Maintains compatibility with existing REST client interface
// Adapts gRPC-Web client to match the REST client API for seamless migration

import { newsGrpcClient } from '../grpc/NewsGrpcClient';
import type {
  NewsArticle,
  NewsResponse,
  NewsArticleResponse,
  GetNewsParams,
  SearchNewsParams,
  LegacyNewsResponse,
} from './types';

// Utility functions to convert between gRPC and REST response formats
function convertGrpcNewsToRest(grpcArticle: any): NewsArticle {
  return {
    id: grpcArticle.id,
    title: grpcArticle.title,
    slug: grpcArticle.slug,
    summary: grpcArticle.summary || '',
    content: grpcArticle.content || '',
    author: grpcArticle.author || '',
    published_at: grpcArticle.publishedDate,
    category: grpcArticle.category || '',
    category_id: grpcArticle.categoryData?.id || '',
    category_data: grpcArticle.categoryData ? {
      id: grpcArticle.categoryData.id,
      name: grpcArticle.categoryData.name,
      description: grpcArticle.categoryData.description || '',
      slug: grpcArticle.categoryData.slug,
      display_order: grpcArticle.categoryData.displayOrder,
      active: grpcArticle.categoryData.active,
      created_at: grpcArticle.categoryData.createdAt,
      updated_at: grpcArticle.categoryData.updatedAt,
    } : undefined,
    image_url: grpcArticle.imageUrl || '',
    featured: grpcArticle.featured,
    status: grpcArticle.status,
    tags: grpcArticle.tags || [],
    meta_title: grpcArticle.metaTitle || '',
    meta_description: grpcArticle.metaDescription || '',
    created_at: grpcArticle.createdAt,
    updated_at: grpcArticle.updatedAt,
  };
}

function convertGrpcResponseToRest(grpcResponse: any): NewsResponse {
  return {
    data: grpcResponse.articles?.map(convertGrpcNewsToRest) || [],
    total: grpcResponse.pagination?.total || 0,
    page: grpcResponse.pagination?.page || 1,
    pageSize: grpcResponse.pagination?.pageSize || 10,
    totalPages: grpcResponse.pagination?.totalPages || 0,
  };
}

function convertRestParamsToGrpc(params: GetNewsParams) {
  return {
    pagination: {
      page: params.page,
      pageSize: params.pageSize,
    },
    filter: {
      categoryId: params.category,
      featured: params.featured,
      status: 'published', // Default to published articles
    },
    sort: {
      field: params.sortBy?.includes('date') ? 'published_date' : 'created_at',
      ascending: params.sortBy?.includes('asc') || false,
    },
  };
}

function convertSearchParamsToGrpc(params: SearchNewsParams) {
  return {
    query: params.q,
    pagination: {
      page: params.page,
      pageSize: params.pageSize,
    },
    filter: {
      categoryId: params.category,
      status: 'published',
    },
    sort: {
      field: params.sortBy?.includes('date') ? 'published_date' : 'created_at',
      ascending: params.sortBy?.includes('asc') || false,
    },
  };
}

/**
 * gRPC-enabled News Client that maintains REST API compatibility
 * Drop-in replacement for the existing NewsClient
 */
export class NewsGrpcAdapter {
  constructor() {
    console.log('🚀 [NewsGrpcAdapter] Initialized with gRPC-Web backend');
  }

  /**
   * Get paginated list of news articles
   */
  async getNewsArticles(params: GetNewsParams = {}): Promise<NewsResponse> {
    try {
      const grpcRequest = convertRestParamsToGrpc(params);
      const grpcResponse = await newsGrpcClient.getNewsArticles(grpcRequest);
      return convertGrpcResponseToRest(grpcResponse);
    } catch (error) {
      console.error('❌ [NewsGrpcAdapter] Error in getNewsArticles:', error);
      throw error;
    }
  }

  /**
   * Get news article by slug
   */
  async getNewsArticleBySlug(slug: string): Promise<NewsArticle> {
    try {
      const grpcArticle = await newsGrpcClient.getNewsArticleBySlug(slug);
      return convertGrpcNewsToRest(grpcArticle);
    } catch (error) {
      console.error('❌ [NewsGrpcAdapter] Error in getNewsArticleBySlug:', error);
      throw error;
    }
  }

  /**
   * Search news articles
   */
  async searchNewsArticles(params: SearchNewsParams): Promise<NewsResponse> {
    try {
      const grpcRequest = convertSearchParamsToGrpc(params);
      const grpcResponse = await newsGrpcClient.searchNewsArticles(grpcRequest);
      return convertGrpcResponseToRest(grpcResponse);
    } catch (error) {
      console.error('❌ [NewsGrpcAdapter] Error in searchNewsArticles:', error);
      throw error;
    }
  }

  /**
   * Get featured news articles
   */
  async getFeaturedNews(limit?: number): Promise<NewsArticle[]> {
    try {
      const grpcArticles = await newsGrpcClient.getFeaturedNews(limit);
      return grpcArticles.map(convertGrpcNewsToRest);
    } catch (error) {
      console.error('❌ [NewsGrpcAdapter] Error in getFeaturedNews:', error);
      throw error;
    }
  }

  /**
   * Get news categories
   */
  async getNewsCategories(): Promise<any[]> {
    try {
      const grpcCategories = await newsGrpcClient.getNewsCategories();
      return grpcCategories.map(cat => ({
        id: cat.id,
        name: cat.name,
        description: cat.description || '',
        slug: cat.slug,
        display_order: cat.displayOrder,
        active: cat.active,
        created_at: cat.createdAt,
        updated_at: cat.updatedAt,
      }));
    } catch (error) {
      console.error('❌ [NewsGrpcAdapter] Error in getNewsCategories:', error);
      return [];
    }
  }

  /**
   * Get news articles by category
   */
  async getNewsByCategory(category: string, params: GetNewsParams = {}): Promise<NewsResponse> {
    return this.getNewsArticles({
      ...params,
      category,
    });
  }

  /**
   * Get recent news articles
   */
  async getRecentNews(limit: number = 5): Promise<NewsArticle[]> {
    try {
      const grpcArticles = await newsGrpcClient.getRecentNews(limit);
      return grpcArticles.map(convertGrpcNewsToRest);
    } catch (error) {
      console.error('❌ [NewsGrpcAdapter] Error in getRecentNews:', error);
      throw error;
    }
  }

  /**
   * Load news page data with organized categories and featured article
   * Used for client-side rendering of news hub page
   */
  async loadNewsPageData() {
    try {
      console.log('🔄 [NewsGrpcAdapter] Loading news page data via gRPC...');
      
      const result = await newsGrpcClient.loadNewsPageData();
      
      // Convert the result to match the REST API format
      return {
        articleCategories: result.articleCategories.map((category: any) => ({
          ...category,
          articles: category.articles.map(convertGrpcNewsToRest),
        })),
        featuredArticle: result.featuredArticle ? convertGrpcNewsToRest(result.featuredArticle) : null,
      };
    } catch (error) {
      console.error('❌ [NewsGrpcAdapter] Error in loadNewsPageData:', error);
      return { articleCategories: [], featuredArticle: null };
    }
  }

  /**
   * Legacy method - maintains backward compatibility
   * @deprecated Use getNewsArticles() instead
   */
  async getLegacyNewsArticles(params: GetNewsParams = {}): Promise<LegacyNewsResponse> {
    const response = await this.getNewsArticles(params);
    return {
      articles: response.data,
      total: response.total,
      page: response.page,
      pageSize: response.pageSize,
      totalPages: response.totalPages,
    };
  }

  /**
   * Health check
   */
  async healthCheck(): Promise<{ status: string; environment: string }> {
    try {
      const health = await newsGrpcClient.healthCheck();
      return {
        status: health.status,
        environment: 'gRPC',
      };
    } catch (error) {
      console.error('❌ [NewsGrpcAdapter] Health check failed:', error);
      throw error;
    }
  }
}

// Export singleton instance as drop-in replacement
export const newsClient = new NewsGrpcAdapter();