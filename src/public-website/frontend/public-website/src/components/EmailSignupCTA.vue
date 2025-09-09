<template>
  <section class="pt-0 pb-0">
    <div
      class="relative overflow-hidden bg-gradient-to-r from-blue-600 to-blue-700 dark:from-blue-700 dark:to-blue-800"
    >
      <div class="container mx-auto px-4">
        <div class="relative p-8 md:p-10 lg:p-12">
          <div class="flex flex-col lg:flex-row items-center gap-6 lg:gap-8">
            <!-- Content -->
            <div class="flex-1 text-center lg:text-left space-y-3">
              <h3 class="text-xl md:text-2xl font-semibold text-white">{{ title }}</h3>
              <p
                v-if="description"
                class="text-white/90 text-sm md:text-base max-w-2xl mx-auto lg:mx-0"
              >
                {{ description }}
              </p>
            </div>

            <!-- Form -->
            <form @submit="handleSubmit" class="w-full max-w-md relative">
              <!-- Honeypot field - hidden from users but visible to bots -->
              <input
                type="text"
                name="website"
                v-model="honeypot"
                style="
                  position: absolute;
                  left: -9999px;
                  width: 1px;
                  height: 1px;
                  opacity: 0;
                  pointer-events: none;
                "
                tabindex="-1"
                aria-hidden="true"
              />

              <!-- Input and Button Row -->
              <div class="flex flex-col sm:flex-row gap-3">
                <Input
                  type="email"
                  v-model="email"
                  @blur="handleEmailBlur"
                  :placeholder="placeholderText"
                  :class="
                    cn(
                      'flex-1 bg-white/10 text-white placeholder:text-white/60 focus:bg-white/20 border-white/20 focus:border-white/40 transition-all duration-200',
                      touched && validationError && 'border-red-500 focus:border-red-500'
                    )
                  "
                  :disabled="status === 'loading' || status === 'success'"
                  autocomplete="email"
                  :aria-invalid="touched && !!validationError"
                />
                <Button
                  type="submit"
                  :disabled="status === 'loading' || status === 'success'"
                  class="bg-white text-blue-600 hover:bg-gray-100 font-medium px-6 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
                >
                  {{
                    status === 'loading'
                      ? 'Subscribing...'
                      : status === 'success'
                        ? 'Subscribed!'
                        : buttonText
                  }}
                </Button>
              </div>

              <!-- Status Messages -->
              <div v-if="touched && validationError" class="mt-2">
                <p class="text-sm text-red-300">{{ validationError }}</p>
              </div>
              <div v-if="status === 'success' && message" class="mt-2">
                <p class="text-sm text-green-300">{{ message }}</p>
              </div>
              <div v-if="status === 'error' && message" class="mt-2">
                <p class="text-sm text-red-300">{{ message }}</p>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- Decorative Elements -->
      <div
        class="absolute top-0 right-0 w-64 h-64 bg-white/5 rounded-full blur-3xl -mr-32 -mt-32"
      ></div>
      <div
        class="absolute bottom-0 left-0 w-48 h-48 bg-white/5 rounded-full blur-2xl -ml-24 -mb-24"
      ></div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import Input from '@/components/vue-ui/Input.vue';
import Button from '@/components/vue-ui/Button.vue';
import { cn } from '@/lib/utils';
import { newsletterClient, type NewsletterSubscriptionData } from '@/lib/clients';

interface EmailSignupCTAProps {
  title: string;
  description: string;
  buttonText?: string;
  placeholderText?: string;
  contentType?: 'case-studies' | 'news'; // Made optional since we're making it generic
}

const props = withDefaults(defineProps<EmailSignupCTAProps>(), {
  buttonText: 'Subscribe',
  placeholderText: 'Enter your email',
});

// Reactive state
const email = ref('');
const status = ref<'idle' | 'loading' | 'success' | 'error'>('idle');
const message = ref('');
const touched = ref(false);
const validationError = ref('');
const honeypot = ref('');

// Comprehensive email validation
const validateEmail = (email: string): { isValid: boolean; message: string } => {
  // Remove whitespace
  const trimmedEmail = email.trim();

  // Check if empty
  if (!trimmedEmail) {
    return { isValid: false, message: '' };
  }

  // Check basic format
  const basicEmailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!basicEmailRegex.test(trimmedEmail)) {
    return { isValid: false, message: 'Please enter a valid email format' };
  }

  // More comprehensive email validation
  const comprehensiveEmailRegex =
    /^[a-zA-Z0-9.!#$%&'*+\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
  if (!comprehensiveEmailRegex.test(trimmedEmail)) {
    return { isValid: false, message: 'Please enter a valid email address' };
  }

  // Check for common typos in domain
  const commonTypos = [
    { pattern: /@gmial\./, suggestion: 'Did you mean @gmail.com?' },
    { pattern: /@gmai\./, suggestion: 'Did you mean @gmail.com?' },
    { pattern: /@yahooo\./, suggestion: 'Did you mean @yahoo.com?' },
    { pattern: /@yaho\./, suggestion: 'Did you mean @yahoo.com?' },
    { pattern: /@outlok\./, suggestion: 'Did you mean @outlook.com?' },
    { pattern: /@hotmial\./, suggestion: 'Did you mean @hotmail.com?' },
  ];

  for (const typo of commonTypos) {
    if (typo.pattern.test(trimmedEmail)) {
      return { isValid: false, message: typo.suggestion };
    }
  }

  // Check for minimum length
  if (trimmedEmail.length < 5) {
    return { isValid: false, message: 'Email address is too short' };
  }

  // Check for maximum length
  if (trimmedEmail.length > 254) {
    return { isValid: false, message: 'Email address is too long' };
  }

  // Check local part length (before @)
  const [localPart] = trimmedEmail.split('@');
  if (localPart.length > 64) {
    return { isValid: false, message: 'Email username is too long' };
  }

  // Check for consecutive dots
  if (/\.{2,}/.test(trimmedEmail)) {
    return { isValid: false, message: 'Email cannot contain consecutive dots' };
  }

  // Check if starts or ends with dot
  if (trimmedEmail.startsWith('.') || trimmedEmail.endsWith('.')) {
    return { isValid: false, message: 'Email cannot start or end with a dot' };
  }

  // All validations passed
  return { isValid: true, message: 'Valid email address' };
};

// Handle email blur validation
const handleEmailBlur = () => {
  touched.value = true;

  if (!email.value.trim()) {
    validationError.value = 'Email address is required';
    return;
  }

  const validation = validateEmail(email.value);
  if (!validation.isValid) {
    validationError.value = validation.message;
  } else {
    validationError.value = '';
  }
};

// Watch email changes for real-time validation
const handleEmailChange = () => {
  // Clear validation error if user starts typing after seeing error
  if (touched.value && validationError.value) {
    const validation = validateEmail(email.value);
    if (validation.isValid) {
      validationError.value = '';
    }
  }

  // Clear any submit error messages
  if (status.value === 'error') {
    status.value = 'idle';
    message.value = '';
  }
};

// Watch email for changes
watch(email, handleEmailChange);

const handleSubmit = async (e: Event) => {
  e.preventDefault();

  const trimmedEmail = email.value.trim();

  if (!trimmedEmail) {
    touched.value = true;
    validationError.value = 'Email address is required';
    return;
  }

  // Final validation before submit
  const validation = validateEmail(trimmedEmail);
  if (!validation.isValid) {
    touched.value = true;
    validationError.value = validation.message;
    return;
  }

  status.value = 'loading';
  validationError.value = '';

  // Submit newsletter subscription to Newsletter API
  try {
    // Check honeypot field (anti-spam measure)
    if (honeypot.value.trim() !== '') {
      status.value = 'error';
      message.value = 'Submission blocked. Please try again.';
      return;
    }

    const subscriptionData: NewsletterSubscriptionData = {
      email: trimmedEmail,
      source: 'website_newsletter_signup',
      contentType: 'all', // Subscribe to all content types (news and case studies)
    };

    // Subscribe using standardized newsletter client
    const response = await newsletterClient.subscribe(subscriptionData);

    if (response.success) {
      status.value = 'success';
      message.value = response.message || 'Successfully subscribed to updates!';
    } else {
      status.value = 'error';
      message.value =
        response.message || 'Unable to subscribe at the moment. Please try again later.';
    }

    if (response.success) {
      email.value = '';
      touched.value = false;
      validationError.value = '';

      // Reset success message after 5 seconds
      setTimeout(() => {
        status.value = 'idle';
        message.value = '';
      }, 5000);
    }
  } catch (error) {
    console.error('Newsletter subscription error:', error);
    status.value = 'error';
    message.value = 'Unable to subscribe at the moment. Please try again later.';
  }
};
</script>
