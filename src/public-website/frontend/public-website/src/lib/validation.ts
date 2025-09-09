/**
 * Form Validation Utilities
 *
 * Lightweight validation system for admin forms.
 * Avoids overengineering while providing essential validation.
 */

export interface ValidationRule {
  test: (value: string) => boolean;
  message: string;
}

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

/**
 * Common validation rules
 */
export const ValidationRules = {
  required: (fieldName: string): ValidationRule => ({
    test: (value: string) => value.trim().length > 0,
    message: `${fieldName} is required`,
  }),

  email: (): ValidationRule => ({
    test: (value: string) => {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return emailRegex.test(value);
    },
    message: 'Please enter a valid email address',
  }),

  minLength: (min: number): ValidationRule => ({
    test: (value: string) => value.length >= min,
    message: `Must be at least ${min} characters long`,
  }),

  maxLength: (max: number): ValidationRule => ({
    test: (value: string) => value.length <= max,
    message: `Must be no more than ${max} characters long`,
  }),

  password: (): ValidationRule => ({
    test: (value: string) => {
      // Basic password requirements for admin users
      return value.length >= 6 && /[a-zA-Z]/.test(value) && /\d/.test(value);
    },
    message: 'Password must be at least 6 characters with letters and numbers',
  }),

  slug: (): ValidationRule => ({
    test: (value: string) => {
      const slugRegex = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
      return slugRegex.test(value);
    },
    message: 'Must be lowercase letters, numbers, and hyphens only',
  }),

  url: (): ValidationRule => ({
    test: (value: string) => {
      try {
        new URL(value);
        return true;
      } catch {
        return false;
      }
    },
    message: 'Please enter a valid URL',
  }),
};

/**
 * Validate a field against multiple rules
 */
export function validateField(value: string, rules: ValidationRule[]): ValidationResult {
  const errors: string[] = [];

  for (const rule of rules) {
    if (!rule.test(value)) {
      errors.push(rule.message);
    }
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

/**
 * Validate an entire form
 */
export function validateForm(
  formData: Record<string, string>,
  fieldRules: Record<string, ValidationRule[]>
): Record<string, ValidationResult> {
  const results: Record<string, ValidationResult> = {};

  for (const [fieldName, rules] of Object.entries(fieldRules)) {
    const value = formData[fieldName] || '';
    results[fieldName] = validateField(value, rules);
  }

  return results;
}

/**
 * Check if entire form validation passed
 */
export function isFormValid(validationResults: Record<string, ValidationResult>): boolean {
  return Object.values(validationResults).every(result => result.isValid);
}

/**
 * Extract all errors from form validation
 */
export function getFormErrors(validationResults: Record<string, ValidationResult>): string[] {
  const allErrors: string[] = [];

  for (const result of Object.values(validationResults)) {
    allErrors.push(...result.errors);
  }

  return allErrors;
}

/**
 * Client-side form validation helper
 * Attaches validation to a form element
 */
export function attachFormValidation(
  form: HTMLFormElement,
  fieldRules: Record<string, ValidationRule[]>,
  options: {
    validateOnSubmit?: boolean;
    validateOnBlur?: boolean;
    showInlineErrors?: boolean;
  } = {}
) {
  const { validateOnSubmit = true, validateOnBlur = false, showInlineErrors = true } = options;

  // Add validation on blur if requested
  if (validateOnBlur) {
    Object.keys(fieldRules).forEach(fieldName => {
      const field = form.querySelector(`[name="${fieldName}"]`) as HTMLInputElement;
      if (field) {
        field.addEventListener('blur', () => {
          validateFormField(field, fieldRules[fieldName], showInlineErrors);
        });
      }
    });
  }

  // Add validation on submit
  if (validateOnSubmit) {
    form.addEventListener('submit', e => {
      const formData = new FormData(form);
      const data: Record<string, string> = {};

      for (const [key, value] of formData.entries()) {
        data[key] = value.toString();
      }

      const validationResults = validateForm(data, fieldRules);

      if (!isFormValid(validationResults)) {
        e.preventDefault();

        // Show validation errors
        if (showInlineErrors) {
          showFormValidationErrors(form, validationResults);
        }

        // Focus first invalid field
        const firstInvalidField = Object.keys(validationResults).find(
          fieldName => !validationResults[fieldName].isValid
        );

        if (firstInvalidField) {
          const field = form.querySelector(`[name="${firstInvalidField}"]`) as HTMLInputElement;
          field?.focus();
        }
      }
    });
  }
}

/**
 * Validate individual form field
 */
function validateFormField(field: HTMLInputElement, rules: ValidationRule[], showErrors: boolean) {
  const result = validateField(field.value, rules);

  if (showErrors) {
    // Remove existing error messages
    const existingError = field.parentElement?.querySelector('.validation-error');
    if (existingError) {
      existingError.remove();
    }

    // Add new error message if validation failed
    if (!result.isValid && result.errors.length > 0) {
      const errorElement = document.createElement('div');
      errorElement.className = 'validation-error text-sm text-red-400 mt-1';
      errorElement.textContent = result.errors[0]; // Show first error
      field.parentElement?.appendChild(errorElement);
    }
  }

  // Add/remove error styling
  if (result.isValid) {
    field.classList.remove('border-red-500', 'border-red-400');
  } else {
    field.classList.add('border-red-500');
  }
}

/**
 * Show form validation errors
 */
function showFormValidationErrors(
  form: HTMLFormElement,
  validationResults: Record<string, ValidationResult>
) {
  // Clear existing errors
  form.querySelectorAll('.validation-error').forEach(el => el.remove());

  // Show errors for each field
  Object.entries(validationResults).forEach(([fieldName, result]) => {
    if (!result.isValid) {
      const field = form.querySelector(`[name="${fieldName}"]`) as HTMLInputElement;
      if (field) {
        validateFormField(field, [], true);

        // Add error message
        const errorElement = document.createElement('div');
        errorElement.className = 'validation-error text-sm text-red-400 mt-1';
        errorElement.textContent = result.errors[0];
        field.parentElement?.appendChild(errorElement);
      }
    }
  });
}

// ===================================================================
// ENHANCED CONTENT MANAGEMENT VALIDATION
// Extends existing validation system with security and UX features
// ===================================================================

export interface ContentFormData {
  title: string;
  slug: string;
  type: string;
  content: string;
  excerpt: string;
  status: string;
  featured: boolean;
}

export interface EnhancedValidationResult {
  isValid: boolean;
  errors: Record<string, string>;
  sanitizedData?: ContentFormData;
  warnings?: string[];
}

/**
 * Input sanitization for content management
 * Prevents XSS while preserving intended content structure
 */
export class ContentSanitizer {
  /**
   * Sanitize plain text fields (titles, excerpts)
   */
  static sanitizeText(input: string): string {
    if (!input || typeof input !== 'string') return '';

    return input
      .replace(/[<>]/g, '') // Remove angle brackets to prevent HTML injection
      .replace(/[\u0000-\u001f\u007f-\u009f]/g, '') // Remove control characters
      .trim()
      .substring(0, 1000); // Reasonable length limit
  }

  /**
   * Sanitize and normalize URL slugs
   */
  static sanitizeSlug(input: string): string {
    if (!input || typeof input !== 'string') return '';

    return input
      .toLowerCase()
      .replace(/[^\w\s-]/g, '') // Only alphanumeric, spaces, hyphens
      .replace(/[\s_-]+/g, '-') // Replace spaces/underscores with single hyphens
      .replace(/^-+|-+$/g, '') // Remove leading/trailing hyphens
      .substring(0, 100); // URL-safe length limit
  }

  /**
   * Sanitize markdown content (preserve formatting, remove dangerous elements)
   */
  static sanitizeMarkdown(input: string): string {
    if (!input || typeof input !== 'string') return '';

    return (
      input
        // Remove script tags and their content
        .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
        // Remove dangerous event handlers from any HTML tags
        .replace(/\s*on\w+\s*=\s*["'][^"']*["']/gi, '')
        // Remove javascript: and data: protocols from links and images
        .replace(/(href|src)\s*=\s*["']?(javascript|data):[^"'>\s]*/gi, '$1="#blocked"')
        // Remove iframe, object, embed tags
        .replace(/<(iframe|object|embed|form)\b[^>]*>.*?<\/\1>/gi, '')
        .trim()
        .substring(0, 50000)
    ); // Reasonable content length limit
  }
}

/**
 * Enhanced content validation with business rules
 */
export class ContentValidator {
  /**
   * Validate complete content form with sanitization
   */
  static async validateContentForm(
    rawData: Partial<ContentFormData>
  ): Promise<EnhancedValidationResult> {
    const errors: Record<string, string> = {};
    const warnings: string[] = [];

    // Extract and provide defaults
    const data: ContentFormData = {
      title: rawData.title || '',
      slug: rawData.slug || '',
      type: rawData.type || '',
      content: rawData.content || '',
      excerpt: rawData.excerpt || '',
      status: rawData.status || 'Draft',
      featured: Boolean(rawData.featured),
    };

    // Title validation
    if (!data.title.trim()) {
      errors.title = 'Title is required';
    } else if (data.title.length < 3) {
      errors.title = 'Title must be at least 3 characters long';
    } else if (data.title.length > 200) {
      errors.title = 'Title must be less than 200 characters';
    } else if (/[<>]/.test(data.title)) {
      errors.title = 'Title cannot contain HTML tags';
    }

    // Slug validation with enhanced rules
    if (!data.slug.trim()) {
      errors.slug = 'URL slug is required';
    } else {
      const normalizedSlug = ContentSanitizer.sanitizeSlug(data.slug);
      if (normalizedSlug !== data.slug) {
        warnings.push('Slug will be automatically normalized to URL-safe format');
      }

      const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
      if (!slugPattern.test(normalizedSlug)) {
        errors.slug = 'Slug must contain only lowercase letters, numbers, and hyphens';
      } else if (normalizedSlug.length < 3) {
        errors.slug = 'Slug must be at least 3 characters long';
      } else if (normalizedSlug.length > 100) {
        errors.slug = 'Slug must be less than 100 characters';
      }
    }

    // Content type validation
    if (!data.type.trim()) {
      errors.type = 'Content type is required';
    } else if (!/^[a-z-]+$/.test(data.type)) {
      errors.type = 'Invalid content type format';
    }

    // Content validation with enhanced security checks
    if (!data.content.trim()) {
      errors.content = 'Content is required';
    } else if (data.content.length < 10) {
      errors.content = 'Content must be at least 10 characters long';
    } else if (data.content.length > 50000) {
      errors.content = 'Content must be less than 50,000 characters';
    } else {
      // Check for potentially dangerous content
      if (/<script/i.test(data.content)) {
        warnings.push('Script tags will be removed from content for security');
      }
      if (/javascript:/i.test(data.content)) {
        warnings.push('JavaScript URLs will be blocked for security');
      }
    }

    // Excerpt validation (optional)
    if (data.excerpt && data.excerpt.length > 500) {
      errors.excerpt = 'Excerpt must be less than 500 characters';
    }

    // Status validation
    const validStatuses = ['Draft', 'Published', 'Archived'];
    if (!validStatuses.includes(data.status)) {
      errors.status = 'Invalid status selected';
    }

    // Sanitize data if validation passes main checks
    let sanitizedData: ContentFormData | undefined;
    if (Object.keys(errors).length === 0) {
      sanitizedData = {
        title: ContentSanitizer.sanitizeText(data.title),
        slug: ContentSanitizer.sanitizeSlug(data.slug),
        type: data.type.trim(),
        content: ContentSanitizer.sanitizeMarkdown(data.content),
        excerpt: data.excerpt ? ContentSanitizer.sanitizeText(data.excerpt) : '',
        status: data.status,
        featured: data.featured,
      };
    }

    return {
      isValid: Object.keys(errors).length === 0,
      errors,
      sanitizedData,
      warnings: warnings.length > 0 ? warnings : undefined,
    };
  }
}

/**
 * Rate limiting for form submissions (MVP implementation)
 */
export class SubmissionRateLimiter {
  private static submissions = new Map<string, number[]>();
  private static readonly MAX_SUBMISSIONS = 10; // Max submissions per window
  private static readonly WINDOW_MS = 15 * 60 * 1000; // 15 minutes

  /**
   * Check if request should be rate limited
   */
  static checkRateLimit(identifier: string): { allowed: boolean; remainingAttempts: number } {
    const now = Date.now();
    const windowStart = now - this.WINDOW_MS;

    // Get existing submissions for this identifier
    let timestamps = this.submissions.get(identifier) || [];

    // Remove old submissions outside the window
    timestamps = timestamps.filter(timestamp => timestamp > windowStart);

    // Check if limit exceeded
    const allowed = timestamps.length < this.MAX_SUBMISSIONS;
    const remainingAttempts = Math.max(0, this.MAX_SUBMISSIONS - timestamps.length);

    // Record this attempt if allowed
    if (allowed) {
      timestamps.push(now);
      this.submissions.set(identifier, timestamps);
    }

    return { allowed, remainingAttempts };
  }
}

/**
 * Enhanced error messaging for better UX
 */
export class ErrorMessageHandler {
  /**
   * Convert technical errors to user-friendly messages
   */
  static getUserFriendlyError(error: any): string {
    if (!error) return 'An unexpected error occurred';

    const message = error.message || error.toString();

    // Content-specific error mappings
    const errorMappings: Record<string, string> = {
      'Authentication required': 'Please log in to continue',
      'Access denied': 'You do not have permission to perform this action',
      'Resource not found': 'The requested content could not be found',
      'Request timeout': 'The request took too long - please try again',
      'Network error': 'Connection problem - please check your internet and try again',
      ValidationException: 'Please check your input and try again',
      DuplicateSlugException: 'This URL slug is already in use - please choose a different one',
      InvalidContentTypeException: 'The selected content type is not valid',
      ContentTooLargeException: 'Your content is too large - please reduce the size and try again',
      UnsupportedFormatException: 'This file format is not supported',
      QuotaExceededException: 'You have reached your content limit',
    };

    // Check for exact matches
    for (const [key, friendlyMessage] of Object.entries(errorMappings)) {
      if (message.includes(key)) {
        return friendlyMessage;
      }
    }

    // Handle HTTP status codes
    if (message.includes('HTTP 400')) {
      return 'Invalid request - please check your input and try again';
    } else if (message.includes('HTTP 401')) {
      return 'Please log in to continue';
    } else if (message.includes('HTTP 403')) {
      return 'You do not have permission to perform this action';
    } else if (message.includes('HTTP 404')) {
      return 'The requested content could not be found';
    } else if (message.includes('HTTP 409')) {
      return 'This content conflicts with existing data - please check for duplicates';
    } else if (message.includes('HTTP 413')) {
      return 'Your content is too large - please reduce the size and try again';
    } else if (message.includes('HTTP 429')) {
      return 'Too many requests - please wait a moment before trying again';
    } else if (message.includes('HTTP 5')) {
      return 'Server error - please try again later or contact support';
    }

    // Fallback message
    return 'An unexpected error occurred - please try again';
  }
}

/**
 * Form state preservation for server-side rendering
 * Maintains form data when validation fails
 */
export interface FormState {
  values: Record<string, any>;
  errors: Record<string, string>;
  warnings?: string[];
  timestamp: number;
}

export class FormStateManager {
  /**
   * Create state parameter for URL preservation
   */
  static createStateParam(
    formData: ContentFormData,
    errors: Record<string, string>,
    warnings?: string[]
  ): string {
    const state: FormState = {
      values: formData,
      errors,
      warnings,
      timestamp: Date.now(),
    };

    try {
      // Use base64 encoding for URL safety
      const encodedState = Buffer.from(JSON.stringify(state)).toString('base64');
      return `?state=${encodeURIComponent(encodedState)}`;
    } catch (error) {
      console.warn('Failed to create form state parameter:', error);
      return '';
    }
  }

  /**
   * Restore form state from URL
   */
  static restoreFromUrl(url: URL): FormState | null {
    try {
      const stateParam = url.searchParams.get('state');
      if (!stateParam) return null;

      const decodedState = Buffer.from(decodeURIComponent(stateParam), 'base64').toString('utf-8');
      const state: FormState = JSON.parse(decodedState);

      // Check if state is too old (prevent URL manipulation attacks)
      const maxAge = 30 * 60 * 1000; // 30 minutes
      if (Date.now() - state.timestamp > maxAge) {
        return null;
      }

      return state;
    } catch (error) {
      console.warn('Failed to restore form state:', error);
      return null;
    }
  }
}
