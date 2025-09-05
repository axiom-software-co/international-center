// Generic Breadcrumb Composable
// Provides consistent breadcrumb functionality across all domains

import { computed } from 'vue';
import { useNavigationStore } from '../../stores/navigation';
import type { DomainConfig } from '../../lib/navigation/domains';

export interface BreadcrumbItem {
  label: string;
  href?: string;
  current?: boolean;
}

export interface BreadcrumbProps {
  parentPath: string;
  parentLabel: string;
  itemName: string;
  title: string;
  category?: string;
}

export interface UseBreadcrumbResult {
  breadcrumbProps: BreadcrumbProps;
  breadcrumbItems: BreadcrumbItem[];
  domainConfig: DomainConfig | null;
}

/**
 * Generic breadcrumb composable for any domain
 */
export function useBreadcrumb(
  domain: string,
  itemName: string,
  title: string,
  category?: string
): UseBreadcrumbResult {
  const navigationStore = useNavigationStore();
  
  const domainConfig = computed(() => {
    return navigationStore.getDomainConfig(domain);
  });

  const breadcrumbProps = computed((): BreadcrumbProps => {
    return navigationStore.generateBreadcrumbProps(domain, itemName, title, category);
  });

  const breadcrumbItems = computed((): BreadcrumbItem[] => {
    const config = domainConfig.value;
    if (!config) return [];

    const items: BreadcrumbItem[] = [
      {
        label: 'Home',
        href: '/',
      },
      {
        label: config.parentLabel,
        href: config.parentPath,
      },
      {
        label: itemName,
        current: true,
      },
    ];

    return items;
  });

  return {
    breadcrumbProps: breadcrumbProps.value,
    breadcrumbItems: breadcrumbItems.value,
    domainConfig: domainConfig.value,
  };
}

/**
 * Domain-specific breadcrumb composables for convenience
 */
export function useEventBreadcrumb(eventName: string, title: string, category?: string) {
  return useBreadcrumb('events', eventName, title, category);
}

export function useNewsBreadcrumb(articleName: string, title: string, category?: string) {
  return useBreadcrumb('news', articleName, title, category);
}

export function useResearchBreadcrumb(articleName: string, title: string, category?: string) {
  return useBreadcrumb('research', articleName, title, category);
}

export function useServiceBreadcrumb(serviceName: string, title: string, category?: string) {
  return useBreadcrumb('services', serviceName, title, category);
}