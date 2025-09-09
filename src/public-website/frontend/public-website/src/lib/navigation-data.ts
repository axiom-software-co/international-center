import { servicesClient, newsClient, researchClient } from './clients';

export interface NavigationService {
  title: string;
  url: string;
  description?: string;
}

export interface NavigationServiceCategory {
  title: string;
  description: string;
  items: NavigationService[];
}

export interface FooterServiceCategory {
  title: string;
  services: Array<{
    name: string;
    href: string;
  }>;
}

export interface HeroService {
  title: string;
  url: string;
}

export interface HeroServicesData {
  primaryCareServices: HeroService[];
  regenerativeTherapies: HeroService[];
  primaryCategoryName: string;
  secondaryCategoryName: string;
}

// Icon mappings for different service types (used by components)
export const getServiceIcon = (serviceSlug: string) => {
  const iconMap = {
    'prp-therapy': 'Heart',
    'exosome-therapy': 'Stethoscope',
    'peptide-therapy': 'Pill',
    'iv-therapy': 'Heart',
    'joint-pain': 'Stethoscope',
    wellness: 'Heart',
    diagnostics: 'Stethoscope',
    'stem-cell': 'Heart',
    'hormone-therapy': 'TestTube',
    'anti-aging': 'Heart',
    'sports-medicine': 'Stethoscope',
    'pain-management': 'Shield',
    'weight-management': 'Heart',
    'immune-support': 'Stethoscope',
    'cognitive-enhancement': 'Brain',
    'sexual-wellness': 'Heart',
    cardiovascular: 'Stethoscope',
    metabolic: 'TrendingUp',
    longevity: 'Heart',
  };
  return iconMap[serviceSlug as keyof typeof iconMap] || 'Heart';
};

/**
 * Transform database services into navigation service categories
 * Uses actual database category information with fallback mapping for display names
 */
export function transformToNavigationCategories(
  dbServices: any[],
  categories: any[]
): NavigationServiceCategory[] {
  if (!dbServices || dbServices.length === 0) {
    return [];
  }

  // Group services by their actual database categories
  const categoryMap = new Map<string, NavigationService[]>();

  // Create a mapping from category IDs to category names
  const categoryIdToName: { [key: number]: string } = {};
  categories.forEach((category: any) => {
    categoryIdToName[category.id] = category.name;
    categoryMap.set(category.name, []);
  });

  // Define category descriptions (matching published website structure)
  const categoryDescriptions = {
    'Primary Care': 'Comprehensive primary care and preventive health services',
    'Primary Care Services': 'Comprehensive primary care and preventive health services',
    'Regenerative Medicine': 'Advanced cellular and biological treatments for healing and recovery',
    'Regenerative Therapies': 'Advanced cellular and biological treatments for healing and recovery',
    'Pain Management': 'Comprehensive solutions for chronic and acute pain relief',
    'Wellness & Prevention': 'Preventive care and overall health optimization services',
    'Diagnostics': 'Advanced diagnostic testing and health assessments',
    'Specialized Care': 'Targeted treatments for specific health conditions',
    'Other Services': 'Additional healthcare services and treatments',
  };

  // Add fallback category for uncategorized services
  categoryMap.set('Other Services', []);

  // Organize services by their actual database categories
  dbServices.forEach(service => {
    // Use actual category information from database
    let categoryName = 'Other Services';
    if (service.category_id && categoryIdToName[service.category_id]) {
      categoryName = categoryIdToName[service.category_id];
    } else if (service.category?.name) {
      categoryName = service.category.name;
    }

    // Ensure category exists in map
    if (!categoryMap.has(categoryName)) {
      categoryMap.set(categoryName, []);
    }

    categoryMap.get(categoryName)!.push({
      title: service.title || 'Unknown Service',
      url: `/services/${service.slug || 'unknown'}`,
      description: service.description || '',
    });
  });

  // Sort services within each category by featured status then alphabetically
  categoryMap.forEach(services => {
    services.sort((a, b) => {
      const aService = dbServices.find(s => s.title === a.title);
      const bService = dbServices.find(s => s.title === b.title);
      
      // Sort by featured first, then alphabetically
      if (aService?.featured && !bService?.featured) return -1;
      if (!aService?.featured && bService?.featured) return 1;
      return a.title.localeCompare(b.title);
    });
  });

  // Define preferred category order for navigation display
  const navigationCategoryOrder = [
    'Primary Care',
    'Primary Care Services',
    'Regenerative Medicine', 
    'Regenerative Therapies',
    'Pain Management',
    'Diagnostics',
    'Wellness & Prevention',
    'Specialized Care',
    'Other Services',
  ];

  // Convert to navigation format with preferred ordering, filtering empty categories
  return Array.from(categoryMap.entries())
    .filter(([_, items]) => items.length > 0)
    .map(([categoryName, items]) => ({
      title: categoryName,
      description: categoryDescriptions[categoryName as keyof typeof categoryDescriptions] || '',
      items,
    }))
    .sort((a, b) => {
      const aIndex = navigationCategoryOrder.indexOf(a.title);
      const bIndex = navigationCategoryOrder.indexOf(b.title);
      
      // If both are in the order array, sort by index
      if (aIndex !== -1 && bIndex !== -1) {
        return aIndex - bIndex;
      }
      // If only one is in the order array, prioritize it
      if (aIndex !== -1) return -1;
      if (bIndex !== -1) return 1;
      // If neither is in the order array, sort alphabetically
      return a.title.localeCompare(b.title);
    });
}

/**
 * Transform database services into footer service categories
 * Uses actual database category information instead of hardcoded slug matching
 */
export function transformToFooterCategories(
  dbServices: any[],
  categories: any[]
): FooterServiceCategory[] {
  if (!dbServices || dbServices.length === 0) {
    return [];
  }

  // Group services by their actual database categories
  const categoryMap = new Map<string, Array<{ name: string; href: string }>>();

  // Create a mapping from category IDs to category names
  const categoryIdToName: { [key: number]: string } = {};
  categories.forEach((category: any) => {
    categoryIdToName[category.id] = category.name;
    categoryMap.set(category.name, []);
  });

  // Add fallback category for uncategorized services
  categoryMap.set('Other Services', []);

  // Organize services by their actual database categories
  dbServices.forEach(service => {
    // Use actual category information from database
    let categoryName = 'Other Services';
    if (service.category_id && categoryIdToName[service.category_id]) {
      categoryName = categoryIdToName[service.category_id];
    } else if (service.category?.name) {
      categoryName = service.category.name;
    }

    // Ensure category exists in map
    if (!categoryMap.has(categoryName)) {
      categoryMap.set(categoryName, []);
    }

    categoryMap.get(categoryName)!.push({
      name: service.title || 'Unknown Service',
      href: `/services/${service.slug || 'unknown'}`,
    });
  });

  // Sort services within each category by featured status then alphabetically
  categoryMap.forEach(services => {
    services.sort((a, b) => {
      const aService = dbServices.find(s => s.title === a.name);
      const bService = dbServices.find(s => s.title === b.name);
      
      // Sort by featured first, then alphabetically
      if (aService?.featured && !bService?.featured) return -1;
      if (!aService?.featured && bService?.featured) return 1;
      return a.name.localeCompare(b.name);
    });
  });

  // Define preferred category order for footer display
  const footerCategoryOrder = [
    'Primary Care',
    'Regenerative Medicine',
    'Pain Management',
    'Diagnostics',
    'Specialized Care',
    'Other Services',
  ];

  // Convert to footer format with preferred ordering, filtering empty categories
  return Array.from(categoryMap.entries())
    .filter(([_, services]) => services.length > 0)
    .map(([categoryName, services]) => ({
      title: categoryName,
      services,
    }))
    .sort((a, b) => {
      const aIndex = footerCategoryOrder.indexOf(a.title);
      const bIndex = footerCategoryOrder.indexOf(b.title);
      
      // If both are in the order array, sort by index
      if (aIndex !== -1 && bIndex !== -1) {
        return aIndex - bIndex;
      }
      // If only one is in the order array, prioritize it
      if (aIndex !== -1) return -1;
      if (bIndex !== -1) return 1;
      // If neither is in the order array, sort alphabetically
      return a.title.localeCompare(b.title);
    });
}

/**
 * Load services data for navigation and footer components
 * Used in Layout.astro during SSR
 */
export async function loadNavigationData() {
  try {
    console.log('🔄 Loading navigation data from Content API...');

    // Fetch both services and categories using new domain client
    const [servicesResponse, categoriesResponse] = await Promise.all([
      servicesClient.getServices({ pageSize: 100 }),
      servicesClient.getServiceCategories(),
    ]);

    const services = servicesResponse.services || servicesResponse.data || [];
    const categories = categoriesResponse.categories || categoriesResponse.data || [];

    if (!services || !Array.isArray(services) || services.length === 0) {
      console.warn('⚠️ No services returned, navigation will show empty state');
      return {
        navigationCategories: [],
        footerCategories: [],
      };
    }

    console.log(`✅ Loaded ${services.length} services for navigation`);
    console.log(`✅ Loaded ${categories.length} categories for navigation`);

    return {
      navigationCategories: transformToNavigationCategories(services, categories),
      footerCategories: transformToFooterCategories(services, categories),
    };
  } catch (error) {
    console.error('❌ Failed to load navigation data:', error);
    return {
      navigationCategories: [],
      footerCategories: [],
    };
  }
}

/**
 * Load top 4 services from featured categories for hero section
 * Used for home page hero component
 */
export async function loadHeroServicesData(): Promise<HeroServicesData> {
  try {
    console.log('🔄 Loading hero services data from Content API...');

    // First fetch categories to find the featured ones
    const categoriesResponse = await servicesClient.getServiceCategories();
    const categories = categoriesResponse.categories || categoriesResponse.data || [];

    if (!categories || !Array.isArray(categories) || categories.length === 0) {
      console.warn('⚠️ No categories returned, hero will show empty state');
      return {
        primaryCareServices: [],
        regenerativeTherapies: [],
        primaryCategoryName: 'Featured Category 1',
        secondaryCategoryName: 'Featured Category 2',
      };
    }

    // Find featured categories
    const featured1Category = categories.find(cat => cat.featured1 === true);
    const featured2Category = categories.find(cat => cat.featured2 === true);

    if (!featured1Category || !featured2Category) {
      console.warn('⚠️ No featured categories found, using fallback');
      return {
        primaryCareServices: [],
        regenerativeTherapies: [],
        primaryCategoryName: 'Featured Category 1',
        secondaryCategoryName: 'Featured Category 2',
      };
    }

    console.log(
      `✅ Found featured categories: ${featured1Category.name} (featured1), ${featured2Category.name} (featured2)`
    );

    // Fetch top 4 services from each featured category
    const [primaryServicesResponse, secondaryServicesResponse] = await Promise.all([
      servicesClient.getServices({
        category: featured1Category.slug,
        pageSize: 4,
      }),
      servicesClient.getServices({
        category: featured2Category.slug,
        pageSize: 4,
      }),
    ]);

    const primaryServices = primaryServicesResponse.services || primaryServicesResponse.data || [];
    const secondaryServices = secondaryServicesResponse.services || secondaryServicesResponse.data || [];

    // Transform to hero service format
    const primaryCareServices = primaryServices.map(service => ({
      title: service.title || 'Unknown Service',
      url: `/services/${service.slug || 'unknown'}`,
    }));

    const regenerativeTherapies = secondaryServices.map(service => ({
      title: service.title || 'Unknown Service',
      url: `/services/${service.slug || 'unknown'}`,
    }));

    console.log(
      `✅ Hero services loaded: ${primaryCareServices.length} from ${featured1Category.name}, ${regenerativeTherapies.length} from ${featured2Category.name}`
    );

    return {
      primaryCareServices,
      regenerativeTherapies,
      primaryCategoryName: featured1Category.name,
      secondaryCategoryName: featured2Category.name,
    };
  } catch (error) {
    console.error('❌ Failed to load hero services data:', error);
    return {
      primaryCareServices: [],
      regenerativeTherapies: [],
      primaryCategoryName: 'Featured Category 1',
      secondaryCategoryName: 'Featured Category 2',
    };
  }
}

/**
 * Load recent content for home page sections
 * Used for research articles and news sections during SSR
 */
export async function loadRecentContentData() {
  try {
    console.log('🔄 Loading recent content data from Content API...');

    // Fetch recent research articles and news articles using new domain clients
    const [researchArticlesResponse, newsResponse] = await Promise.all([
      researchClient.getResearchArticles({ pageSize: 3 }),
      newsClient.getNewsArticles({ pageSize: 3 }),
    ]);

    const researchArticles = researchArticlesResponse.articles || researchArticlesResponse.data || [];
    const newsArticles = newsResponse.articles || newsResponse.data || [];

    console.log(
      `✅ Loaded recent content: ${researchArticles?.length || 0} research articles, ${newsArticles?.length || 0} news articles`
    );

    return {
      researchArticles: researchArticles || [],
      newsArticles: newsArticles || [],
    };
  } catch (error) {
    console.error('❌ Failed to load recent content data:', error);
    return {
      researchArticles: [],
      newsArticles: [],
    };
  }
}


/**
 * Load services data for services page
 * Used for client-side rendering of services page
 */
export async function loadServicesPageData() {
  try {
    console.log('🔄 Loading services page data from REST API through Public Gateway...');

    // Fetch both services and categories from the REST API
    const [servicesResponse, categoriesResponse] = await Promise.all([
      servicesClient.getServices({ pageSize: 100 }),
      servicesClient.getServiceCategories(),
    ]);

    // Handle REST API response format
    if (!servicesResponse.success) {
      console.error('❌ Services API request failed:', servicesResponse.message);
      return { serviceCategories: [] };
    }

    if (!categoriesResponse.success) {
      console.error('❌ Categories API request failed:', categoriesResponse.message);
      return { serviceCategories: [] };
    }

    const allServices = servicesResponse.data || [];
    const categories = categoriesResponse.data || [];

    if (!allServices || !Array.isArray(allServices) || allServices.length === 0) {
      console.warn('⚠️ No services returned from API');
      return { serviceCategories: [] };
    }

    console.log(`✅ Loaded ${allServices.length} services from API`);
    console.log(`✅ Loaded ${categories.length} categories from API`);

    // Create a map for category organization
    const categoryMap = new Map();

    // Create a mapping from category IDs to category names for easier lookup
    const categoryIdToName: { [key: number]: string } = {};
    categories.forEach((category: any) => {
      categoryIdToName[category.id] = category.name;
      categoryMap.set(category.name, {
        id: category.id,
        title: category.name,
        description: category.description || '',
        services: [],
      });
    });

    // Add uncategorized category for services without a category
    categoryMap.set('Other Services', {
      id: null,
      title: 'Other Services',
      description: 'Additional healthcare services and treatments',
      services: [],
    });

    // Organize services by their categories
    allServices.forEach((service: any) => {
      // Use category_id to find the category name, then fall back to the category object or 'Other Services'
      let categoryName = 'Other Services';
      if (service.category_id && categoryIdToName[service.category_id]) {
        categoryName = categoryIdToName[service.category_id];
      } else if (service.category?.name) {
        categoryName = service.category.name;
      }

      const categoryData = categoryMap.get(categoryName);

      if (categoryData) {
        const serviceForDisplay = {
          name: service.title || 'Unknown Service',
          href: `/services/${service.slug || ''}`,
          description: service.description || '',
          duration: service.duration || '45-90 minutes',
          available: service.available === true,
          featured: service.featured || false,
          // Pass through the actual delivery_modes array for dynamic processing
          delivery_modes: Array.isArray(service.delivery_modes) ? service.delivery_modes : [],
        };

        categoryData.services.push(serviceForDisplay);
      }
    });

    // Convert to array and filter empty categories, sort services within categories
    const serviceCategories = Array.from(categoryMap.values())
      .filter((category: any) => category.services.length > 0)
      .map((category: any) => ({
        ...category,
        services: category.services.sort((a: any, b: any) => {
          // Sort by featured first, then by name
          if (a.featured && !b.featured) return -1;
          if (!a.featured && b.featured) return 1;
          return a.name.localeCompare(b.name);
        }),
      }))
      .sort((a: any, b: any) => {
        // Sort categories by predefined order (using actual API category names)
        const categoryOrder = [
          'Primary Care',
          'Regenerative Medicine',
          'Pain Management',
          'Diagnostics',
          'Specialized Care',
          'Other Services',
        ];
        const aIndex = categoryOrder.indexOf(a.title);
        const bIndex = categoryOrder.indexOf(b.title);

        // If both are in the order array, sort by index
        if (aIndex !== -1 && bIndex !== -1) {
          return aIndex - bIndex;
        }
        // If only one is in the order array, prioritize it
        if (aIndex !== -1) return -1;
        if (bIndex !== -1) return 1;
        // If neither is in the order array, sort alphabetically
        return a.title.localeCompare(b.title);
      });

    console.log(`✅ Organized services into ${serviceCategories.length} categories`);

    return { serviceCategories };
  } catch (error) {
    console.error('❌ Failed to load services page data:', error);
    return { serviceCategories: [] };
  }
}
