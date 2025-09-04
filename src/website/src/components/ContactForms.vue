<template>
  <div :class="`grid gap-8 lg:grid-cols-2 lg:items-stretch ${className}`">
    <!-- Business Inquiry Form -->
    <Card class="p-6 transition-colors group h-full flex flex-col">
      <CardHeader class="p-0 mb-6">
        <CardTitle class="text-2xl transition-colors mb-3">Business Inquiries</CardTitle>
        <p class="text-muted-foreground">
          Partnership opportunities and professional collaboration.
        </p>
      </CardHeader>
      <CardContent class="p-0 flex-1 flex flex-col">
        <form @submit="handleBusinessSubmit" class="flex flex-col h-full" novalidate>
          <div class="space-y-6 flex-1">
            <div class="space-y-2">
              <Label for="organizationName">Organization Name *</Label>
              <Input
                id="organizationName"
                name="organizationName"
                v-model="businessForm.organizationName"
                @blur="() => handleBusinessBlur('organizationName', businessForm.organizationName)"
                :class="
                  cn(
                    businessTouched.organizationName &&
                      businessErrors.organizationName &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                placeholder="Your company or organization name"
                maxlength="100"
                required
              />
              <p
                v-if="businessTouched.organizationName && businessErrors.organizationName"
                class="text-sm text-red-500"
              >
                {{ businessErrors.organizationName }}
              </p>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label for="contactName">Contact Name *</Label>
                <Input
                  id="contactName"
                  name="contactName"
                  v-model="businessForm.contactName"
                  @blur="() => handleBusinessBlur('contactName', businessForm.contactName)"
                  :class="
                    cn(
                      businessTouched.contactName &&
                        businessErrors.contactName &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="First and last name"
                  maxlength="50"
                  required
                />
                <p
                  v-if="businessTouched.contactName && businessErrors.contactName"
                  class="text-sm text-red-500"
                >
                  {{ businessErrors.contactName }}
                </p>
              </div>
              <div class="space-y-2">
                <Label for="title">Title/Position *</Label>
                <Input
                  id="title"
                  name="title"
                  v-model="businessForm.title"
                  @blur="() => handleBusinessBlur('title', businessForm.title)"
                  :class="
                    cn(
                      businessTouched.title &&
                        businessErrors.title &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="CEO, CTO, Director, etc."
                  maxlength="50"
                  required
                />
                <p
                  v-if="businessTouched.title && businessErrors.title"
                  class="text-sm text-red-500"
                >
                  {{ businessErrors.title }}
                </p>
              </div>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label for="business-email">Business Email Address *</Label>
                <Input
                  type="email"
                  id="business-email"
                  name="email"
                  v-model="businessForm.email"
                  @blur="() => handleBusinessBlur('email', businessForm.email)"
                  :class="
                    cn(
                      businessTouched.email &&
                        businessErrors.email &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="john@company.com"
                  maxlength="254"
                  autocomplete="email"
                  required
                />
                <p
                  v-if="businessTouched.email && businessErrors.email"
                  class="text-sm text-red-500"
                >
                  {{ businessErrors.email }}
                </p>
              </div>
              <div class="space-y-2">
                <Label for="business-phone">
                  Phone Number <span class="text-muted-foreground">(Optional)</span>
                </Label>
                <Input
                  type="tel"
                  id="business-phone"
                  name="phone"
                  v-model="businessForm.phone"
                  @input="handleBusinessPhoneInput"
                  @blur="() => handleBusinessBlur('phone', businessForm.phone)"
                  :class="
                    cn(
                      businessTouched.phone &&
                        businessErrors.phone &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="(555) 123-4567"
                  maxlength="14"
                  autocomplete="tel"
                />
                <p
                  v-if="businessTouched.phone && businessErrors.phone"
                  class="text-sm text-red-500"
                >
                  {{ businessErrors.phone }}
                </p>
              </div>
            </div>

            <div class="space-y-2">
              <Label for="inquiryType">Type of Inquiry</Label>
              <Select v-model="businessForm.inquiryType">
                <SelectTrigger>
                  <SelectValue placeholder="Select inquiry type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="partnership">Strategic Partnership</SelectItem>
                  <SelectItem value="licensing">Licensing Opportunities</SelectItem>
                  <SelectItem value="research">Research Collaboration</SelectItem>
                  <SelectItem value="technology">Technology Integration</SelectItem>
                  <SelectItem value="regulatory">Regulatory Consultation</SelectItem>
                  <SelectItem value="other">Other Business Matter</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-2">
              <Label for="industry">
                Industry Sector <span class="text-muted-foreground">(Optional)</span>
              </Label>
              <Input
                id="industry"
                name="industry"
                placeholder="Healthcare, Biotech, Medical Devices, Pharmaceuticals, etc."
                v-model="businessForm.industry"
                maxlength="50"
              />
            </div>

            <div class="space-y-2">
              <Label for="business-message">
                Message * <span class="text-muted-foreground">(Min. 20 characters)</span>
              </Label>
              <Textarea
                id="business-message"
                name="message"
                rows="4"
                placeholder="Describe your organization's interests, objectives, and how we might collaborate..."
                v-model="businessForm.message"
                @blur="() => handleBusinessBlur('message', businessForm.message)"
                :class="
                  cn(
                    businessTouched.message &&
                      businessErrors.message &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                maxlength="1500"
                required
              />
              <p
                v-if="businessTouched.message && businessErrors.message"
                class="text-sm text-red-500"
              >
                {{ businessErrors.message }}
              </p>
              <p class="text-xs text-muted-foreground">
                {{ businessForm.message.length }}/1500 characters
              </p>
            </div>
          </div>

          <!-- Honeypot field for spam protection - hidden from users -->
          <input
            type="text"
            name="website"
            tabindex="-1"
            autocomplete="off"
            style="
              position: absolute;
              left: -9999px;
              width: 1px;
              height: 1px;
              opacity: 0;
              visibility: hidden;
            "
            aria-hidden="true"
          />

          <!-- Submit status messages -->
          <div
            v-if="submitStatus.business === 'success'"
            class="p-4 rounded-lg bg-green-50 border border-green-200"
          >
            <p class="text-green-800 text-sm">{{ submitMessages.business }}</p>
          </div>

          <div
            v-if="submitStatus.business === 'error'"
            class="p-4 rounded-lg bg-red-50 border border-red-200"
          >
            <p class="text-red-800 text-sm">{{ submitMessages.business }}</p>
          </div>

          <Button type="submit" class="w-full mt-6" size="lg" :disabled="isSubmitting.business">
            {{ isSubmitting.business ? 'Submitting...' : 'Submit Business Inquiry' }}
          </Button>
        </form>
      </CardContent>
    </Card>

    <!-- Media Relations Form -->
    <Card class="p-6 transition-colors group h-full flex flex-col">
      <CardHeader class="p-0 mb-6">
        <CardTitle class="text-2xl transition-colors mb-3">Media Relations</CardTitle>
        <p class="text-muted-foreground">
          Press inquiries, interviews, and media coverage requests.
        </p>
      </CardHeader>
      <CardContent class="p-0 flex-1 flex flex-col">
        <form @submit="handleMediaSubmit" class="flex flex-col h-full" novalidate>
          <div class="space-y-6 flex-1">
            <div class="space-y-2">
              <Label for="outlet">Media Outlet *</Label>
              <Input
                id="outlet"
                name="outlet"
                v-model="mediaForm.outlet"
                @blur="() => handleMediaBlur('outlet', mediaForm.outlet)"
                :class="
                  cn(
                    mediaTouched.outlet &&
                      mediaErrors.outlet &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                placeholder="Associated Press, Reuters, Miami Herald, etc."
                maxlength="100"
                required
              />
              <p v-if="mediaTouched.outlet && mediaErrors.outlet" class="text-sm text-red-500">
                {{ mediaErrors.outlet }}
              </p>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label for="media-contactName">Reporter/Contact Name *</Label>
                <Input
                  id="media-contactName"
                  name="contactName"
                  v-model="mediaForm.contactName"
                  @blur="() => handleMediaBlur('contactName', mediaForm.contactName)"
                  :class="
                    cn(
                      mediaTouched.contactName &&
                        mediaErrors.contactName &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="First and last name"
                  maxlength="50"
                  required
                />
                <p
                  v-if="mediaTouched.contactName && mediaErrors.contactName"
                  class="text-sm text-red-500"
                >
                  {{ mediaErrors.contactName }}
                </p>
              </div>
              <div class="space-y-2">
                <Label for="media-title">Title/Position *</Label>
                <Input
                  id="media-title"
                  name="title"
                  v-model="mediaForm.title"
                  @blur="() => handleMediaBlur('title', mediaForm.title)"
                  :class="
                    cn(
                      mediaTouched.title &&
                        mediaErrors.title &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="Journalist, Editor, Producer, etc."
                  maxlength="50"
                  required
                />
                <p v-if="mediaTouched.title && mediaErrors.title" class="text-sm text-red-500">
                  {{ mediaErrors.title }}
                </p>
              </div>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label for="media-email">Professional Email Address *</Label>
                <Input
                  type="email"
                  id="media-email"
                  name="email"
                  v-model="mediaForm.email"
                  @blur="() => handleMediaBlur('email', mediaForm.email)"
                  :class="
                    cn(
                      mediaTouched.email &&
                        mediaErrors.email &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="reporter@newsoutlet.com"
                  maxlength="254"
                  autocomplete="email"
                  required
                />
                <p v-if="mediaTouched.email && mediaErrors.email" class="text-sm text-red-500">
                  {{ mediaErrors.email }}
                </p>
              </div>
              <div class="space-y-2">
                <Label for="media-phone">Direct Phone Number *</Label>
                <Input
                  type="tel"
                  id="media-phone"
                  name="phone"
                  v-model="mediaForm.phone"
                  @input="handleMediaPhoneInput"
                  @blur="() => handleMediaBlur('phone', mediaForm.phone)"
                  :class="
                    cn(
                      mediaTouched.phone &&
                        mediaErrors.phone &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="(555) 123-4567"
                  maxlength="14"
                  autocomplete="tel"
                  required
                />
                <p v-if="mediaTouched.phone && mediaErrors.phone" class="text-sm text-red-500">
                  {{ mediaErrors.phone }}
                </p>
              </div>
            </div>

            <div class="space-y-2">
              <Label for="mediaType">Media Type</Label>
              <Select v-model="mediaForm.mediaType">
                <SelectTrigger>
                  <SelectValue placeholder="Select media type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="print">Print Publication</SelectItem>
                  <SelectItem value="digital">Digital/Online Media</SelectItem>
                  <SelectItem value="television">Television</SelectItem>
                  <SelectItem value="radio">Radio</SelectItem>
                  <SelectItem value="podcast">Podcast</SelectItem>
                  <SelectItem value="medical-journal">Medical Journal</SelectItem>
                  <SelectItem value="other">Other</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-2">
              <Label for="deadline">
                Deadline <span class="text-muted-foreground">(Optional)</span>
              </Label>
              <DatePicker
                id="deadline"
                v-model="mediaForm.deadline"
                placeholder="Select deadline date"
              />
              <p class="text-xs text-muted-foreground">
                Maximum 5 years from today. Format: YYYY-MM-DD
              </p>
            </div>

            <div class="space-y-2">
              <Label for="subject">
                Story Subject/Topic *
                <span class="text-muted-foreground">(Min. 20 characters)</span>
              </Label>
              <Textarea
                id="subject"
                name="subject"
                rows="4"
                placeholder="Describe your story angle, interview questions, coverage scope, or specific information needed..."
                v-model="mediaForm.subject"
                @blur="() => handleMediaBlur('subject', mediaForm.subject)"
                :class="
                  cn(
                    mediaTouched.subject &&
                      mediaErrors.subject &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                maxlength="1500"
                required
              />
              <p v-if="mediaTouched.subject && mediaErrors.subject" class="text-sm text-red-500">
                {{ mediaErrors.subject }}
              </p>
              <p class="text-xs text-muted-foreground">
                {{ mediaForm.subject.length }}/1500 characters
              </p>
            </div>
          </div>

          <!-- Honeypot field for spam protection - hidden from users -->
          <input
            type="text"
            name="company"
            tabindex="-1"
            autocomplete="off"
            style="
              position: absolute;
              left: -9999px;
              width: 1px;
              height: 1px;
              opacity: 0;
              visibility: hidden;
            "
            aria-hidden="true"
          />

          <!-- Submit status messages -->
          <div
            v-if="submitStatus.media === 'success'"
            class="p-4 rounded-lg bg-green-50 border border-green-200"
          >
            <p class="text-green-800 text-sm">{{ submitMessages.media }}</p>
          </div>

          <div
            v-if="submitStatus.media === 'error'"
            class="p-4 rounded-lg bg-red-50 border border-red-200"
          >
            <p class="text-red-800 text-sm">{{ submitMessages.media }}</p>
          </div>

          <Button
            type="submit"
            class="w-full mt-6"
            size="lg"
            variant="default"
            :disabled="isSubmitting.media"
          >
            {{ isSubmitting.media ? 'Submitting...' : 'Submit Media Request' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import Card from '@/components/vue-ui/Card.vue';
import CardContent from '@/components/vue-ui/CardContent.vue';
import CardHeader from '@/components/vue-ui/CardHeader.vue';
import CardTitle from '@/components/vue-ui/CardTitle.vue';
import Button from '@/components/vue-ui/Button.vue';
import Input from '@/components/vue-ui/Input.vue';
import Textarea from '@/components/vue-ui/Textarea.vue';
import Select from '@/components/vue-ui/Select.vue';
import SelectContent from '@/components/vue-ui/SelectContent.vue';
import SelectItem from '@/components/vue-ui/SelectItem.vue';
import SelectTrigger from '@/components/vue-ui/SelectTrigger.vue';
import SelectValue from '@/components/vue-ui/SelectValue.vue';
import DatePicker from '@/components/vue-ui/DatePicker.vue';
import Label from '@/components/vue-ui/Label.vue';
import { cn } from '@/lib/utils';
import { contactsClient, type ContactSubmission } from '../lib/clients';

interface ContactFormsProps {
  className?: string;
}

const props = withDefaults(defineProps<ContactFormsProps>(), {
  className: '',
});

// Vue reactive state for contact submission
const submitting = ref(false);
const error = ref<string | null>(null);
const success = ref(false);

// Contact submission function
const submitContact = async (contactData: ContactSubmission) => {
  submitting.value = true;
  error.value = null;
  success.value = false;

  try {
    const response = await contactsClient.submitContact(contactData);
    if (response.success) {
      success.value = true;
      return response;
    } else {
      error.value = response.message || 'Failed to submit contact form';
      return response;
    }
  } catch (err) {
    error.value = 'Network error occurred. Please try again.';
    console.error('Contact submission error:', err);
    return { success: false, message: error.value };
  } finally {
    submitting.value = false;
  }
};

// Reset function
const reset = () => {
  submitting.value = false;
  error.value = null;
  success.value = false;
};

// Input filtering functions
const filterNameInput = (value: string): string => {
  // Allow only letters, spaces, hyphens, and apostrophes
  return value.replace(/[^a-zA-Z\s\-']/g, '');
};

const filterPhoneInput = (value: string): string => {
  // Allow only numbers, spaces, parentheses, hyphens, plus, and periods
  return value.replace(/[^0-9\s()\-+.]/g, '');
};

// Format USA phone number as user types
const formatUSAPhone = (value: string): string => {
  // Remove all non-digit characters
  const phoneNumber = value.replace(/\D/g, '');

  // Limit to 10 digits for USA numbers
  const truncated = phoneNumber.slice(0, 10);

  // Format based on length
  if (truncated.length === 0) {
    return '';
  } else if (truncated.length <= 3) {
    return `(${truncated}`;
  } else if (truncated.length <= 6) {
    return `(${truncated.slice(0, 3)}) ${truncated.slice(3)}`;
  } else {
    return `(${truncated.slice(0, 3)}) ${truncated.slice(3, 6)}-${truncated.slice(6)}`;
  }
};

// Extract clean phone number from formatted string
const cleanPhoneNumber = (value: string): string => {
  return value.replace(/\D/g, '');
};

const filterEmailInput = (value: string): string => {
  // Allow only valid email characters
  return value.replace(/[^a-zA-Z0-9@._-]/g, '').toLowerCase();
};

const filterOrganizationInput = (value: string): string => {
  // Allow letters, numbers, spaces, common punctuation for business names
  return value.replace(/[^a-zA-Z0-9\s\-'.,&()]/g, '');
};

const filterIndustryInput = (value: string): string => {
  // Allow letters, spaces, hyphens, commas for industry names
  return value.replace(/[^a-zA-Z\s\-,]/g, '');
};

// Validation functions
const validateEmail = (email: string): string | null => {
  if (!email) return 'Email is required';
  const businessEmailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
  if (!businessEmailRegex.test(email)) return 'Please enter a valid business email address';
  if (email.includes('+')) return 'Please use your primary business email';
  return null;
};

const validatePhone = (phone: string): string | null => {
  if (!phone) return null; // Phone is optional in business form
  // Extract digits only for validation
  const digitsOnly = phone.replace(/\D/g, '');
  // USA phone numbers should be exactly 10 digits
  if (digitsOnly.length !== 10) return 'Please enter a valid 10-digit USA phone number';
  // Check if it starts with a valid area code (not 0 or 1)
  if (digitsOnly[0] === '0' || digitsOnly[0] === '1') return 'Invalid USA phone number';
  return null;
};

const validateOrganization = (name: string): string | null => {
  if (!name) return 'Organization name is required';
  if (name.length < 2) return 'Organization name must be at least 2 characters';
  if (name.length > 100) return 'Organization name must be less than 100 characters';
  return null;
};

const validateContactName = (name: string): string | null => {
  if (!name) return 'Contact name is required';
  if (name.length < 2) return 'Name must be at least 2 characters';
  if (name.length > 50) return 'Name must be less than 50 characters';
  if (!/^[a-zA-Z\s\-']+$/.test(name))
    return 'Name can only contain letters, spaces, hyphens, and apostrophes';
  return null;
};

const validateTitle = (title: string): string | null => {
  if (!title) return 'Title/Position is required';
  if (title.length < 2) return 'Title must be at least 2 characters';
  if (title.length > 50) return 'Title must be less than 50 characters';
  return null;
};

const validateMessage = (message: string, minLength: number = 20): string | null => {
  if (!message) return 'Message is required';
  if (message.length < minLength) return `Message must be at least ${minLength} characters`;
  if (message.length > 1500) return 'Message must be less than 1500 characters';
  return null;
};

const validateMediaOutlet = (outlet: string): string | null => {
  if (!outlet) return 'Media outlet is required';
  if (outlet.length < 2) return 'Media outlet name must be at least 2 characters';
  if (outlet.length > 100) return 'Media outlet name must be less than 100 characters';
  return null;
};


const validateSubject = (subject: string): string | null => {
  if (!subject) return 'Story subject/topic is required';
  if (subject.length < 20) return 'Subject must be at least 20 characters';
  if (subject.length > 1500) return 'Subject must be less than 1500 characters';
  return null;
};

// Form state
const businessForm = reactive({
  organizationName: '',
  contactName: '',
  title: '',
  email: '',
  phone: '',
  inquiryType: '',
  industry: '',
  message: '',
});

const mediaForm = reactive({
  outlet: '',
  contactName: '',
  title: '',
  email: '',
  phone: '',
  mediaType: '',
  deadline: '',
  subject: '',
});

const businessErrors = reactive<Record<string, string | null>>({});
const mediaErrors = reactive<Record<string, string | null>>({});
const businessTouched = reactive<Record<string, boolean>>({});
const mediaTouched = reactive<Record<string, boolean>>({});
const isSubmitting = reactive({ business: false, media: false });

// Success and error states
const submitStatus = reactive<{
  business: 'idle' | 'success' | 'error';
  media: 'idle' | 'success' | 'error';
}>({ business: 'idle', media: 'idle' });
const submitMessages = reactive<{
  business: string;
  media: string;
}>({ business: '', media: '' });

// Business form validation
const validateBusinessForm = (): boolean => {
  // Extract clean phone number for validation
  const cleanPhone = businessForm.phone ? cleanPhoneNumber(businessForm.phone) : '';

  const errors: Record<string, string | null> = {
    organizationName: validateOrganization(businessForm.organizationName),
    contactName: validateContactName(businessForm.contactName),
    title: validateTitle(businessForm.title),
    email: validateEmail(businessForm.email),
    phone: validatePhone(cleanPhone),
    message: validateMessage(businessForm.message),
  };

  Object.assign(businessErrors, errors);
  return !Object.values(errors).some(error => error !== null);
};

// Media form validation
const validateMediaForm = (): boolean => {
  // Extract clean phone number for validation
  const cleanPhone = mediaForm.phone ? cleanPhoneNumber(mediaForm.phone) : '';

  const errors: Record<string, string | null> = {
    outlet: validateMediaOutlet(mediaForm.outlet),
    contactName: validateContactName(mediaForm.contactName),
    title: validateTitle(mediaForm.title),
    email: validateEmail(mediaForm.email),
    phone: cleanPhone ? validatePhone(cleanPhone) : 'Phone number is required for media inquiries',
    subject: validateSubject(mediaForm.subject),
  };

  // Phone is required for media inquiries
  if (!mediaForm.phone) {
    errors.phone = 'Phone number is required for media inquiries';
  }

  Object.assign(mediaErrors, errors);
  return !Object.values(errors).some(error => error !== null);
};

// Handle field blur for real-time validation
const handleBusinessBlur = (field: string, value: string) => {
  businessTouched[field] = true;
  const error = (() => {
    switch (field) {
      case 'organizationName':
        return validateOrganization(value);
      case 'contactName':
        return validateContactName(value);
      case 'title':
        return validateTitle(value);
      case 'email':
        return validateEmail(value);
      case 'phone': {
        if (!value) return null; // Phone is optional for business form
        // Extract clean phone number for validation
        const cleanPhone = cleanPhoneNumber(value);
        return validatePhone(cleanPhone);
      }
      case 'message':
        return validateMessage(value);
      default:
        return null;
    }
  })();
  businessErrors[field] = error;
};

const handleMediaBlur = (field: string, value: string) => {
  mediaTouched[field] = true;
  const error = (() => {
    switch (field) {
      case 'outlet':
        return validateMediaOutlet(value);
      case 'contactName':
        return validateContactName(value);
      case 'title':
        return validateTitle(value);
      case 'email':
        return validateEmail(value);
      case 'phone': {
        if (!value) return 'Phone number is required for media inquiries';
        // Extract clean phone number for validation
        const cleanPhone = cleanPhoneNumber(value);
        return validatePhone(cleanPhone);
      }
      case 'subject':
        return validateSubject(value);
      default:
        return null;
    }
  })();
  mediaErrors[field] = error;
};

// Phone input handlers
const handleBusinessPhoneInput = (e: Event) => {
  const input = (e.target as HTMLInputElement).value;
  // If user is deleting, allow it
  if (input.length < businessForm.phone.length) {
    businessForm.phone = input;
  } else {
    // Format as USA phone number
    const filtered = filterPhoneInput(input);
    const formatted = formatUSAPhone(filtered);
    businessForm.phone = formatted;
  }
};

const handleMediaPhoneInput = (e: Event) => {
  const input = (e.target as HTMLInputElement).value;
  // If user is deleting, allow it
  if (input.length < mediaForm.phone.length) {
    mediaForm.phone = input;
  } else {
    // Format as USA phone number
    const filtered = filterPhoneInput(input);
    const formatted = formatUSAPhone(filtered);
    mediaForm.phone = formatted;
  }
};

// Watchers for input filtering
watch(
  () => businessForm.organizationName,
  newVal => {
    businessForm.organizationName = filterOrganizationInput(newVal);
  }
);

watch(
  () => businessForm.contactName,
  newVal => {
    businessForm.contactName = filterNameInput(newVal);
  }
);

watch(
  () => businessForm.title,
  newVal => {
    businessForm.title = filterOrganizationInput(newVal);
  }
);

watch(
  () => businessForm.email,
  newVal => {
    businessForm.email = filterEmailInput(newVal);
  }
);

watch(
  () => businessForm.industry,
  newVal => {
    businessForm.industry = filterIndustryInput(newVal);
  }
);

watch(
  () => mediaForm.outlet,
  newVal => {
    mediaForm.outlet = filterOrganizationInput(newVal);
  }
);

watch(
  () => mediaForm.contactName,
  newVal => {
    mediaForm.contactName = filterNameInput(newVal);
  }
);

watch(
  () => mediaForm.title,
  newVal => {
    mediaForm.title = filterOrganizationInput(newVal);
  }
);

watch(
  () => mediaForm.email,
  newVal => {
    mediaForm.email = filterEmailInput(newVal);
  }
);

const handleBusinessSubmit = async (e: Event) => {
  e.preventDefault();
  isSubmitting.business = true;
  submitStatus.business = 'idle';
  submitMessages.business = '';

  // Mark all fields as touched for error display
  const allFields = Object.keys(businessForm);
  allFields.forEach(field => {
    businessTouched[field] = true;
  });

  if (!validateBusinessForm()) {
    isSubmitting.business = false;
    return;
  }

  // Reset hook state
  reset();

  // Prepare data for standardized Contact API submission
  const submissionData: ContactSubmission = {
    name: businessForm.contactName,
    email: businessForm.email,
    phone: businessForm.phone || undefined,
    subject: `Business Inquiry - ${businessForm.inquiryType || 'General'}`,
    message: `Organization: ${businessForm.organizationName}\nContact: ${businessForm.contactName}\nTitle: ${businessForm.title}\nInquiry Type: ${businessForm.inquiryType || 'Not specified'}\nIndustry: ${businessForm.industry || 'Not specified'}\n\nMessage:\n${businessForm.message}`,
  };

  try {
    // Submit using standardized hook
    const response = await submitContact(submissionData);

    if (response) {
      console.log('✅ Business inquiry submitted successfully:', response);

      // Set success state
      submitStatus.business = 'success';
      const referenceMsg = response.id ? ` Your reference ID is: ${response.id}` : '';
      submitMessages.business = `Thank you! Your business inquiry has been submitted successfully. We will review your request and respond within 24-48 hours.${referenceMsg}`;

      // Reset form on success
      Object.assign(businessForm, {
        organizationName: '',
        contactName: '',
        title: '',
        email: '',
        phone: '',
        inquiryType: '',
        industry: '',
        message: '',
      });
      Object.keys(businessErrors).forEach(key => (businessErrors[key] = null));
      Object.keys(businessTouched).forEach(key => (businessTouched[key] = false));
    } else {
      // Handle hook error state
      submitStatus.business = 'error';
      submitMessages.business =
        error ||
        'Unable to submit your inquiry at this time. Please try again later or contact us directly.';
    }
  } finally {
    isSubmitting.business = false;
  }
};

const handleMediaSubmit = async (e: Event) => {
  e.preventDefault();
  isSubmitting.media = true;
  submitStatus.media = 'idle';
  submitMessages.media = '';

  // Mark all fields as touched for error display
  const allFields = Object.keys(mediaForm);
  allFields.forEach(field => {
    mediaTouched[field] = true;
  });

  if (!validateMediaForm()) {
    isSubmitting.media = false;
    return;
  }

  // Reset hook state
  reset();

  // Determine urgency based on deadline
  let urgency: 'low' | 'medium' | 'high' = 'medium';
  if (mediaForm.deadline) {
    const deadlineDate = new Date(mediaForm.deadline);
    const today = new Date();
    const diffDays = Math.ceil((deadlineDate.getTime() - today.getTime()) / (1000 * 3600 * 24));

    if (diffDays <= 1) urgency = 'high';
    else if (diffDays <= 3) urgency = 'medium';
    else urgency = 'low';
  }

  // Prepare data for standardized Contact API submission
  const submissionData: ContactSubmission = {
    name: mediaForm.contactName,
    email: mediaForm.email,
    phone: mediaForm.phone,
    subject: `Media Inquiry - ${mediaForm.mediaType || 'General'}${mediaForm.deadline ? ` (Deadline: ${mediaForm.deadline})` : ''}`,
    message: `Media Outlet: ${mediaForm.outlet}\nContact: ${mediaForm.contactName}\nTitle: ${mediaForm.title}\nMedia Type: ${mediaForm.mediaType || 'Not specified'}\nDeadline: ${mediaForm.deadline || 'Not specified'}\nUrgency: ${urgency}\n\nStory Subject/Topic:\n${mediaForm.subject}`,
  };

  try {
    // Submit using standardized hook
    const response = await submitContact(submissionData);

    if (response) {
      console.log('✅ Media inquiry submitted successfully:', response);

      // Set success state
      submitStatus.media = 'success';
      const urgencyNote =
        urgency === 'high'
          ? ' Due to the urgent nature of your request, we will prioritize our response.'
          : '';
      const referenceMsg = response.id ? ` Your reference ID is: ${response.id}` : '';
      submitMessages.media = `Thank you! Your media inquiry has been submitted successfully. We will respond within 4-8 hours for standard requests.${urgencyNote}${referenceMsg}`;

      // Reset form on success
      Object.assign(mediaForm, {
        outlet: '',
        contactName: '',
        title: '',
        email: '',
        phone: '',
        mediaType: '',
        deadline: '',
        subject: '',
      });
      Object.keys(mediaErrors).forEach(key => (mediaErrors[key] = null));
      Object.keys(mediaTouched).forEach(key => (mediaTouched[key] = false));
    } else {
      // Handle hook error state
      submitStatus.media = 'error';
      submitMessages.media =
        error ||
        'Unable to submit your media inquiry at this time. Please try again later or contact us directly.';
    }
  } finally {
    isSubmitting.media = false;
  }
};
</script>
