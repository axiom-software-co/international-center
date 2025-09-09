// Metadata Service - SEO and metadata aggregation
// Provides unified metadata generation for all pages

import { servicesClient } from '../clients/services/ServicesClient';
import { newsClient } from '../clients/news/NewsClient';
import { researchClient } from '../clients/research/ResearchClient';
import type { Service } from '../clients/services/types';
import type { NewsArticle } from '../clients/news/types';
import type { ResearchArticle } from '../clients/research/types';

export interface PageMetadata {
  title: string;
  description: string;
  keywords: string[];
  ogTitle?: string;
  ogDescription?: string;
  ogImage?: string;
  ogType?: 'website' | 'article';
  twitterCard?: 'summary' | 'summary_large_image';
  canonicalUrl?: string;
  alternateUrls?: Array<{ hreflang: string; href: string }>;
  jsonLd?: object;
  robots?: string;
  publishedTime?: string;
  modifiedTime?: string;
  author?: string;
  section?: string;
}

export interface SiteMetadata {
  siteName: string;
  siteDescription: string;
  siteUrl: string;
  defaultImage: string;
  twitterHandle: string;
  facebookAppId?: string;
  organizationJsonLd: object;
}

export class MetadataService {
  private readonly siteMetadata: SiteMetadata = {
    siteName: 'International Center',
    siteDescription:
      'Advanced regenerative medicine and wellness solutions for optimal health and healing.',
    siteUrl: 'https://internationalcenter.com',
    defaultImage: '/og-image.svg',
    twitterHandle: '@InternationalCenter',
    organizationJsonLd: {
      '@context': 'https://schema.org',
      '@type': 'MedicalOrganization',
      name: 'International Center',
      description:
        'Advanced regenerative medicine and wellness solutions for optimal health and healing.',
      url: 'https://internationalcenter.com',
      logo: 'https://internationalcenter.com/icon-simple.svg',
      contactPoint: {
        '@type': 'ContactPoint',
        telephone: '+1-555-0123',
        contactType: 'customer service',
        availableLanguage: 'English',
      },
      address: {
        '@type': 'PostalAddress',
        addressCountry: 'US',
      },
      sameAs: [
        'https://www.facebook.com/internationalcenter',
        'https://twitter.com/internationalcenter',
        'https://www.linkedin.com/company/internationalcenter',
      ],
    },
  };

  /**
   * Get metadata for home page
   */
  getHomeMetadata(): PageMetadata {
    return {
      title: `${this.siteMetadata.siteName} - Advanced Regenerative Medicine`,
      description: this.siteMetadata.siteDescription,
      keywords: ['regenerative medicine', 'wellness', 'health', 'therapy', 'treatment'],
      ogTitle: `${this.siteMetadata.siteName} - Advanced Regenerative Medicine`,
      ogDescription: this.siteMetadata.siteDescription,
      ogImage: this.siteMetadata.defaultImage,
      ogType: 'website',
      twitterCard: 'summary_large_image',
      canonicalUrl: this.siteMetadata.siteUrl,
      jsonLd: this.siteMetadata.organizationJsonLd,
    };
  }

  /**
   * Get metadata for services listing page
   */
  async getServicesMetadata(): Promise<PageMetadata> {
    try {
      const services = await servicesClient.getServices({ pageSize: 10 });
      const serviceNames = services.data.map((s: any) => s.title);

      return {
        title: `Services - ${this.siteMetadata.siteName}`,
        description: `Explore our comprehensive range of advanced medical services including ${serviceNames.slice(0, 3).join(', ')} and more.`,
        keywords: ['medical services', 'treatments', 'therapy', ...serviceNames],
        ogTitle: `Medical Services - ${this.siteMetadata.siteName}`,
        ogDescription: `Advanced medical services for optimal health and wellness.`,
        ogImage: this.siteMetadata.defaultImage,
        ogType: 'website',
        twitterCard: 'summary_large_image',
        canonicalUrl: `${this.siteMetadata.siteUrl}/services`,
        jsonLd: this.generateServicesJsonLd(services.data),
      };
    } catch (error) {
      console.error('Error generating services metadata:', error);
      return this.getFallbackServicesMetadata();
    }
  }

  /**
   * Get metadata for individual service page
   */
  async getServiceMetadata(slug: string): Promise<PageMetadata> {
    try {
      const service = await servicesClient.getServiceBySlug(slug);

      return {
        title: `${service.title} - ${this.siteMetadata.siteName}`,
        description: service.meta_description || service.description,
        keywords: [service.title, ...service.technologies, ...service.features],
        ogTitle: service.meta_title || service.title,
        ogDescription: service.meta_description || service.description,
        ogImage: service.image || this.siteMetadata.defaultImage,
        ogType: 'article',
        twitterCard: 'summary_large_image',
        canonicalUrl: `${this.siteMetadata.siteUrl}/services/${service.slug}`,
        jsonLd: this.generateServiceJsonLd(service),
        section: 'Services',
      };
    } catch (error) {
      console.error('Error generating service metadata:', error);
      return this.getFallbackMetadata('Service', `/services/${slug}`);
    }
  }

  /**
   * Get metadata for news listing page
   */
  async getNewsMetadata(): Promise<PageMetadata> {
    try {
      const news = await newsClient.getNewsArticles({ pageSize: 5 });
      const recentTitles = news.data.map((a: any) => a.title);

      return {
        title: `Company News - ${this.siteMetadata.siteName}`,
        description: `Stay updated with the latest news and announcements from ${this.siteMetadata.siteName}.`,
        keywords: ['company news', 'announcements', 'updates', 'healthcare news'],
        ogTitle: `Company News - ${this.siteMetadata.siteName}`,
        ogDescription: `Latest news and updates from ${this.siteMetadata.siteName}.`,
        ogImage: this.siteMetadata.defaultImage,
        ogType: 'website',
        twitterCard: 'summary_large_image',
        canonicalUrl: `${this.siteMetadata.siteUrl}/company/news`,
        jsonLd: this.generateNewsListingJsonLd(news.data),
      };
    } catch (error) {
      console.error('Error generating news metadata:', error);
      return this.getFallbackMetadata('Company News', '/company/news');
    }
  }

  /**
   * Get metadata for individual news article
   */
  async getNewsArticleMetadata(slug: string): Promise<PageMetadata> {
    try {
      const article = await newsClient.getNewsArticleBySlug(slug);

      return {
        title: `${article.title} - ${this.siteMetadata.siteName}`,
        description: article.meta_description || article.excerpt,
        keywords: [article.title, article.category, ...article.tags],
        ogTitle: article.meta_title || article.title,
        ogDescription: article.meta_description || article.excerpt,
        ogImage: article.featured_image || this.siteMetadata.defaultImage,
        ogType: 'article',
        twitterCard: 'summary_large_image',
        canonicalUrl: `${this.siteMetadata.siteUrl}/company/news/${article.slug}`,
        jsonLd: this.generateNewsArticleJsonLd(article),
        publishedTime: article.published_at,
        modifiedTime: article.updated_at,
        author: article.author,
        section: 'News',
      };
    } catch (error) {
      console.error('Error generating news article metadata:', error);
      return this.getFallbackMetadata('News Article', `/company/news/${slug}`);
    }
  }

  /**
   * Get metadata for research listing page
   */
  async getResearchMetadata(): Promise<PageMetadata> {
    try {
      const research = await researchClient.getResearchArticles({ pageSize: 5 });

      return {
        title: `Research & Innovation - ${this.siteMetadata.siteName}`,
        description: `Explore our latest research, case studies, and innovations in regenerative medicine.`,
        keywords: ['medical research', 'case studies', 'innovation', 'clinical trials'],
        ogTitle: `Research & Innovation - ${this.siteMetadata.siteName}`,
        ogDescription: `Latest research and innovations in regenerative medicine.`,
        ogImage: this.siteMetadata.defaultImage,
        ogType: 'website',
        twitterCard: 'summary_large_image',
        canonicalUrl: `${this.siteMetadata.siteUrl}/community/research-innovation`,
        jsonLd: this.generateResearchListingJsonLd(research.data),
      };
    } catch (error) {
      console.error('Error generating research metadata:', error);
      return this.getFallbackMetadata('Research & Innovation', '/community/research-innovation');
    }
  }

  /**
   * Get metadata for individual research article
   */
  async getResearchArticleMetadata(slug: string): Promise<PageMetadata> {
    try {
      const article = await researchClient.getResearchArticleBySlug(slug);

      return {
        title: `${article.title} - ${this.siteMetadata.siteName}`,
        description: article.meta_description || article.excerpt,
        keywords: [article.title, article.category, ...article.tags, ...article.technologies],
        ogTitle: article.meta_title || article.title,
        ogDescription: article.meta_description || article.excerpt,
        ogImage: article.featured_image || this.siteMetadata.defaultImage,
        ogType: 'article',
        twitterCard: 'summary_large_image',
        canonicalUrl: `${this.siteMetadata.siteUrl}/community/research-innovation/${article.slug}`,
        jsonLd: this.generateResearchArticleJsonLd(article),
        publishedTime: article.published_at,
        modifiedTime: article.updated_at,
        author: article.author,
        section: 'Research',
      };
    } catch (error) {
      console.error('Error generating research article metadata:', error);
      return this.getFallbackMetadata('Research Article', `/community/research-innovation/${slug}`);
    }
  }

  /**
   * Generate JSON-LD for services listing
   */
  private generateServicesJsonLd(services: Service[]): object {
    return {
      '@context': 'https://schema.org',
      '@type': 'ItemList',
      name: 'Medical Services',
      description: 'Comprehensive medical and wellness services',
      itemListElement: services.map((service, index) => ({
        '@type': 'ListItem',
        position: index + 1,
        item: {
          '@type': 'MedicalProcedure',
          name: service.title,
          description: service.description,
          url: `${this.siteMetadata.siteUrl}/services/${service.slug}`,
        },
      })),
    };
  }

  /**
   * Generate JSON-LD for individual service
   */
  private generateServiceJsonLd(service: Service): object {
    return {
      '@context': 'https://schema.org',
      '@type': 'MedicalProcedure',
      name: service.title,
      description: service.description,
      url: `${this.siteMetadata.siteUrl}/services/${service.slug}`,
      image: service.image,
      provider: {
        '@type': 'MedicalOrganization',
        name: this.siteMetadata.siteName,
        url: this.siteMetadata.siteUrl,
      },
    };
  }

  /**
   * Generate JSON-LD for news article
   */
  private generateNewsArticleJsonLd(article: NewsArticle): object {
    return {
      '@context': 'https://schema.org',
      '@type': 'NewsArticle',
      headline: article.title,
      description: article.excerpt,
      image: article.featured_image,
      datePublished: article.published_at,
      dateModified: article.updated_at,
      author: {
        '@type': 'Person',
        name: article.author,
      },
      publisher: {
        '@type': 'Organization',
        name: this.siteMetadata.siteName,
        logo: {
          '@type': 'ImageObject',
          url: `${this.siteMetadata.siteUrl}/icon-simple.svg`,
        },
      },
      mainEntityOfPage: `${this.siteMetadata.siteUrl}/company/news/${article.slug}`,
    };
  }

  /**
   * Generate JSON-LD for research article
   */
  private generateResearchArticleJsonLd(article: ResearchArticle): object {
    return {
      '@context': 'https://schema.org',
      '@type': 'ScholarlyArticle',
      headline: article.title,
      description: article.excerpt,
      image: article.featured_image,
      datePublished: article.published_at,
      dateModified: article.updated_at,
      author: {
        '@type': 'Person',
        name: article.author,
      },
      publisher: {
        '@type': 'Organization',
        name: this.siteMetadata.siteName,
      },
      mainEntityOfPage: `${this.siteMetadata.siteUrl}/community/research-innovation/${article.slug}`,
      keywords: article.tags.join(', '),
    };
  }

  private generateNewsListingJsonLd(articles: NewsArticle[]): object {
    return {
      '@context': 'https://schema.org',
      '@type': 'Blog',
      name: `${this.siteMetadata.siteName} News`,
      description: 'Latest news and updates',
      url: `${this.siteMetadata.siteUrl}/company/news`,
      blogPost: articles.map(article => ({
        '@type': 'BlogPosting',
        headline: article.title,
        url: `${this.siteMetadata.siteUrl}/company/news/${article.slug}`,
        datePublished: article.published_at,
        author: {
          '@type': 'Person',
          name: article.author,
        },
      })),
    };
  }

  private generateResearchListingJsonLd(articles: ResearchArticle[]): object {
    return {
      '@context': 'https://schema.org',
      '@type': 'ItemList',
      name: 'Research & Innovation',
      description: 'Latest research and case studies',
      itemListElement: articles.map((article, index) => ({
        '@type': 'ListItem',
        position: index + 1,
        item: {
          '@type': 'ScholarlyArticle',
          name: article.title,
          url: `${this.siteMetadata.siteUrl}/community/research-innovation/${article.slug}`,
          datePublished: article.published_at,
        },
      })),
    };
  }

  private getFallbackServicesMetadata(): PageMetadata {
    return {
      title: `Services - ${this.siteMetadata.siteName}`,
      description: 'Comprehensive medical and wellness services for optimal health.',
      keywords: ['medical services', 'treatments', 'therapy'],
      ogTitle: `Services - ${this.siteMetadata.siteName}`,
      ogDescription: 'Comprehensive medical and wellness services.',
      ogImage: this.siteMetadata.defaultImage,
      ogType: 'website',
      canonicalUrl: `${this.siteMetadata.siteUrl}/services`,
    };
  }

  private getFallbackMetadata(pageTitle: string, path: string): PageMetadata {
    return {
      title: `${pageTitle} - ${this.siteMetadata.siteName}`,
      description: this.siteMetadata.siteDescription,
      keywords: [pageTitle.toLowerCase(), 'health', 'wellness'],
      ogTitle: `${pageTitle} - ${this.siteMetadata.siteName}`,
      ogDescription: this.siteMetadata.siteDescription,
      ogImage: this.siteMetadata.defaultImage,
      ogType: 'website',
      canonicalUrl: `${this.siteMetadata.siteUrl}${path}`,
    };
  }
}

// Export singleton instance
export const metadataService = new MetadataService();
