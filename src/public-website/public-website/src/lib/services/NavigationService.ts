// Navigation Service - Business logic for site navigation and menu generation
// Provides dynamic navigation based on available content

import { servicesClient } from '../clients/services/ServicesClient';
import { newsClient } from '../clients/news/NewsClient';
import { researchClient } from '../clients/research/ResearchClient';
import type { Service } from '../clients/services/types';

export interface NavigationItem {
  label: string;
  href: string;
  children?: NavigationItem[];
  external?: boolean;
  featured?: boolean;
}

export interface NavigationMenu {
  services: NavigationItem[];
  content: NavigationItem[];
  company: NavigationItem[];
}

export class NavigationService {
  private cache: {
    menu?: NavigationMenu;
    lastUpdated?: number;
  } = {};

  private readonly CACHE_DURATION = 300000; // 5 minutes

  /**
   * Get complete navigation menu
   */
  async getNavigationMenu(): Promise<NavigationMenu> {
    // Return cached version if still valid
    if (this.cache.menu && this.cache.lastUpdated) {
      const now = Date.now();
      if (now - this.cache.lastUpdated < this.CACHE_DURATION) {
        return this.cache.menu;
      }
    }

    try {
      const [servicesMenu, contentMenu] = await Promise.allSettled([
        this.generateServicesMenu(),
        this.generateContentMenu(),
      ]);

      const menu: NavigationMenu = {
        services: servicesMenu.status === 'fulfilled' ? servicesMenu.value : [],
        content: contentMenu.status === 'fulfilled' ? contentMenu.value : [],
        company: this.getStaticCompanyMenu(),
      };

      // Cache the result
      this.cache.menu = menu;
      this.cache.lastUpdated = Date.now();

      return menu;
    } catch (error) {
      console.error('Error generating navigation menu:', error);
      return this.getFallbackMenu();
    }
  }

  /**
   * Generate services navigation menu dynamically
   */
  private async generateServicesMenu(): Promise<NavigationItem[]> {
    try {
      const services = await servicesClient.getServices({ pageSize: 20 });

      const servicesMenu: NavigationItem[] = [
        {
          label: 'All Services',
          href: '/services',
        },
      ];

      // Add individual services
      services.data
        .sort((a: any, b: any) => a.sort_order - b.sort_order)
        .forEach((service: any) => {
          servicesMenu.push({
            label: service.title,
            href: `/services/${service.slug}`,
            featured: service.sort_order <= 3, // Mark first 3 as featured
          });
        });

      return servicesMenu;
    } catch (error) {
      console.error('Error generating services menu:', error);
      return [
        {
          label: 'All Services',
          href: '/services',
        },
      ];
    }
  }

  /**
   * Generate content navigation menu dynamically
   */
  private async generateContentMenu(): Promise<NavigationItem[]> {
    try {
      const [newsCategories, researchCategories] = await Promise.allSettled([
        Promise.resolve([]), // News categories not implemented yet
        researchClient.getResearchCategories(),
      ]);

      const contentMenu: NavigationItem[] = [
        {
          label: 'Company News',
          href: '/company/news',
          children: [],
        },
        {
          label: 'Research & Innovation',
          href: '/community/research-innovation',
          children: [],
        },
      ];

      // Add news categories as children
      if (newsCategories.status === 'fulfilled' && newsCategories.value.length > 0) {
        const newsMenuItem = contentMenu.find(item => item.href === '/company/news');
        if (newsMenuItem) {
          newsMenuItem.children = newsCategories.value.map((category: any) => ({
            label: category,
            href: `/company/news?category=${encodeURIComponent(category)}`,
          }));
        }
      }

      // Add research categories as children
      if (researchCategories.status === 'fulfilled' && researchCategories.value.length > 0) {
        const researchMenuItem = contentMenu.find(
          item => item.href === '/community/research-innovation'
        );
        if (researchMenuItem) {
          researchMenuItem.children = researchCategories.value.map(category => ({
            label: category,
            href: `/community/research-innovation?category=${encodeURIComponent(category)}`,
          }));
        }
      }

      return contentMenu;
    } catch (error) {
      console.error('Error generating content menu:', error);
      return [
        {
          label: 'Company News',
          href: '/company/news',
        },
        {
          label: 'Research & Innovation',
          href: '/community/research-innovation',
        },
      ];
    }
  }

  /**
   * Get static company navigation menu
   */
  private getStaticCompanyMenu(): NavigationItem[] {
    return [
      {
        label: 'About',
        href: '/company',
      },
      {
        label: 'Team',
        href: '/company/team',
      },
      {
        label: 'Contact',
        href: '/company/contact',
      },
      {
        label: 'Patient Resources',
        href: '/patient-resources',
      },
    ];
  }

  /**
   * Get fallback menu when dynamic generation fails
   */
  private getFallbackMenu(): NavigationMenu {
    return {
      services: [
        {
          label: 'All Services',
          href: '/services',
        },
      ],
      content: [
        {
          label: 'Company News',
          href: '/company/news',
        },
        {
          label: 'Research & Innovation',
          href: '/community/research-innovation',
        },
      ],
      company: this.getStaticCompanyMenu(),
    };
  }

  /**
   * Get breadcrumb navigation for a given path
   */
  getBreadcrumbs(pathname: string): NavigationItem[] {
    const segments = pathname.split('/').filter(Boolean);
    const breadcrumbs: NavigationItem[] = [
      {
        label: 'Home',
        href: '/',
      },
    ];

    let currentPath = '';
    for (const segment of segments) {
      currentPath += `/${segment}`;

      // Convert segment to readable label
      const label = this.segmentToLabel(segment);

      breadcrumbs.push({
        label,
        href: currentPath,
      });
    }

    return breadcrumbs;
  }

  /**
   * Convert URL segment to readable label
   */
  private segmentToLabel(segment: string): string {
    return segment
      .split('-')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  }

  /**
   * Clear navigation cache (useful for testing or after content updates)
   */
  clearCache(): void {
    this.cache = {};
  }

  /**
   * Get services for dropdown menu (cached for performance)
   */
  async getServicesForDropdown(limit: number = 6): Promise<Service[]> {
    try {
      const services = await servicesClient.getServices({ pageSize: limit });
      return services.data.sort((a: any, b: any) => a.sort_order - b.sort_order);
    } catch (error) {
      console.error('Error fetching services for dropdown:', error);
      return [];
    }
  }
}

// Export singleton instance
export const navigationService = new NavigationService();
