// Content utilities for International Center Platform
// Provides content processing and formatting functions

/**
 * Calculate estimated reading time based on content
 * Assumes average reading speed of 200 words per minute
 */
export function calculateReadingTime(content: string, wordsPerMinute: number = 200): string {
  if (!content) {
    return '1 min read'; // Default for empty content
  }
  
  // Remove HTML tags and count words
  const plainText = content.replace(/<[^>]*>/g, '').trim();
  const words = plainText.split(/\s+/).filter(word => word.length > 0);
  const wordCount = words.length;
  
  if (wordCount === 0) {
    return '1 min read';
  }
  
  const minutes = Math.max(1, Math.ceil(wordCount / wordsPerMinute));
  
  return `${minutes} min read`;
}

/**
 * Generate hero image URL with fallback
 * Uses existing asset utilities but provides a content-specific interface
 */
export function generateHeroImageUrl(
  imageUrl: string | null | undefined,
  title: string,
  contentType: 'news' | 'service' | 'research' = 'news'
): string {
  // If we have a valid image URL, use it
  if (imageUrl && imageUrl.trim()) {
    return imageUrl;
  }
  
  // Generate fallback placeholder URL
  const encodedTitle = encodeURIComponent(title);
  const dimensions = '1200x600';
  const backgroundColor = '2563eb'; // Blue color
  const textColor = 'ffffff'; // White text
  
  return `https://placehold.co/${dimensions}/${backgroundColor}/${textColor}/png?text=${encodedTitle}`;
}

/**
 * Generate SEO-friendly alt text for images
 */
export function generateImageAlt(
  title: string,
  contentType: 'news' | 'service' | 'research' = 'news',
  suffix: string = 'International Center'
): string {
  const typeText = {
    news: 'News',
    service: 'Service',
    research: 'Research'
  }[contentType];
  
  return `${title} - ${suffix} ${typeText}`;
}

/**
 * Create excerpt from content
 * @param content - HTML or plain text content
 * @param maxLength - Maximum character length for excerpt
 * @param suffix - Suffix to add if content is truncated
 */
export function createExcerpt(content: string, maxLength: number = 160, suffix: string = '...'): string {
  if (!content) return '';
  
  // Remove HTML tags and clean up whitespace
  const plainText = content
    .replace(/<[^>]*>/g, '')
    .replace(/\s+/g, ' ')
    .trim();
    
  if (plainText.length <= maxLength) {
    return plainText;
  }
  
  // Find the last word boundary within the limit
  const truncated = plainText.substring(0, maxLength);
  const lastSpaceIndex = truncated.lastIndexOf(' ');
  
  if (lastSpaceIndex > maxLength * 0.8) {
    return truncated.substring(0, lastSpaceIndex) + suffix;
  }
  
  return truncated + suffix;
}

/**
 * Extract headings from HTML content for table of contents
 */
export function extractHeadings(htmlContent: string): Array<{ level: number; text: string; id: string }> {
  if (!htmlContent) return [];
  
  const headingRegex = /<h([1-6])[^>]*>([^<]+)<\/h[1-6]>/gi;
  const headings: Array<{ level: number; text: string; id: string }> = [];
  let match;
  
  while ((match = headingRegex.exec(htmlContent)) !== null) {
    const level = parseInt(match[1]);
    const text = match[2].trim();
    const id = createSlugFromText(text);
    
    headings.push({ level, text, id });
  }
  
  return headings;
}

/**
 * Create a URL-safe slug from text
 */
export function createSlugFromText(text: string): string {
  if (!text) return '';
  
  return text
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

/**
 * Validate content for required fields
 */
export function validateContent(content: {
  title?: string;
  summary?: string;
  content?: string;
}): { isValid: boolean; errors: string[] } {
  const errors: string[] = [];
  
  if (!content.title || content.title.trim().length === 0) {
    errors.push('Title is required');
  }
  
  if (content.title && content.title.length > 200) {
    errors.push('Title must be less than 200 characters');
  }
  
  if (!content.summary || content.summary.trim().length === 0) {
    errors.push('Summary is required');
  }
  
  if (content.summary && content.summary.length > 500) {
    errors.push('Summary must be less than 500 characters');
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
}

/**
 * Format author name consistently
 */
export function formatAuthorName(
  authorName: string | null | undefined,
  fallback: string = 'International Center Team'
): string {
  if (!authorName || authorName.trim().length === 0) {
    return fallback;
  }
  
  return authorName.trim();
}

/**
 * Check if content is featured based on various criteria
 */
export function shouldFeatureContent(content: {
  featured?: boolean;
  order_number?: number;
  published_at?: string;
  category_id?: string;
}): boolean {
  // Explicitly marked as featured
  if (content.featured === true) {
    return true;
  }
  
  // High priority order numbers (1-3)
  if (content.order_number && content.order_number <= 3) {
    return true;
  }
  
  // Recent important categories
  if (content.published_at && content.category_id) {
    const publishDate = new Date(content.published_at);
    const daysSincePublish = (Date.now() - publishDate.getTime()) / (1000 * 60 * 60 * 24);
    
    // Feature recent content from important categories
    if (daysSincePublish <= 7 && ['breaking-news', 'announcements'].includes(content.category_id)) {
      return true;
    }
  }
  
  return false;
}

/**
 * Parse delivery modes based on service type
 * Determines which delivery modes (mobile, outpatient, inpatient) are available for a service
 * @param slug - The service slug to analyze
 * @returns Array of delivery mode strings
 */
export function parseServiceDeliveryModes(slug: string): string[] {
  const modes: string[] = [];

  // Mobile services (can be performed at patient location)
  const mobileServices = [
    'prp-therapy',
    'exosome-therapy', 
    'peptide-therapy',
    'iv-therapy',
    'wellness',
    'immunizations',
    'telehealth',
    'annual-wellness',
    'chronic-care',
    'physical-exams',
    'immune-support'
  ];

  if (mobileServices.includes(slug)) {
    modes.push('mobile');
  }

  // Outpatient services (most services are outpatient except complex procedures)
  const inpatientOnlyServices = ['stem-cell'];
  if (!inpatientOnlyServices.includes(slug)) {
    modes.push('outpatient');
  }

  // Inpatient services (requiring facility stay or complex procedures)
  const inpatientServices = ['stem-cell', 'diagnostics', 'longevity'];
  if (inpatientServices.includes(slug)) {
    modes.push('inpatient');
  }

  // Default to outpatient if no modes determined
  return modes.length > 0 ? modes : ['outpatient'];
}

/**
 * Format event date for display
 * @param dateString - ISO date string (e.g., "2024-03-15")
 * @returns Formatted date string (e.g., "Friday, March 15, 2024")
 */
export function formatEventDate(dateString: string): string {
  if (!dateString) return '';
  
  try {
    return new Date(dateString).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  } catch (error) {
    console.warn('Invalid date string provided to formatEventDate:', dateString);
    return dateString; // Return original string if parsing fails
  }
}

/**
 * Generate event image URL with fallback placeholder
 * @param imageUrl - Original featured image URL
 * @param eventTitle - Event title for fallback placeholder
 * @returns Image URL with fallback
 */
export function generateEventImageUrl(
  imageUrl: string | null | undefined,
  eventTitle: string
): string {
  // If we have a valid image URL, use it
  if (imageUrl && imageUrl.trim()) {
    return imageUrl;
  }
  
  // Generate fallback placeholder URL with event title
  const encodedTitle = encodeURIComponent(eventTitle);
  return `https://placehold.co/800x600/e5e7eb/6b7280/png?text=${encodedTitle}`;
}

/**
 * Parse event status and determine appropriate display information
 * @param status - Event status string
 * @param capacity - Event capacity (optional)
 * @param registered - Number of registered attendees (optional)
 * @returns Object with parsed status information
 */
export function parseEventStatus(
  status: string,
  capacity?: number,
  registered?: number
): {
  displayStatus: string;
  buttonText: string;
  isEventFull: boolean;
  canRegister: boolean;
} {
  const isEventFull = capacity && registered !== undefined ? registered >= capacity : false;
  
  // Handle different status cases
  switch (status.toLowerCase()) {
    case 'published':
      return {
        displayStatus: 'Open for Registration',
        buttonText: isEventFull ? 'Event Full' : 'Join Event',
        isEventFull,
        canRegister: !isEventFull,
      };
      
    case 'registration required':
      return {
        displayStatus: 'Registration Required',
        buttonText: isEventFull ? 'Event Full' : 'Register Now',
        isEventFull,
        canRegister: !isEventFull,
      };
      
    case 'draft':
      return {
        displayStatus: 'Draft',
        buttonText: 'Not Available',
        isEventFull: false,
        canRegister: false,
      };
      
    case 'archived':
      return {
        displayStatus: 'Archived',
        buttonText: 'Event Ended',
        isEventFull: false,
        canRegister: false,
      };
      
    case 'cancelled':
      return {
        displayStatus: 'Cancelled',
        buttonText: 'Event Cancelled',
        isEventFull: false,
        canRegister: false,
      };
      
    default:
      return {
        displayStatus: status,
        buttonText: isEventFull ? 'Event Full' : 'Join Event',
        isEventFull,
        canRegister: !isEventFull,
      };
  }
}