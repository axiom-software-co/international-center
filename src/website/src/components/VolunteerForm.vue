<template>
  <div :class="`max-w-2xl mx-auto ${className}`">
    <!-- Volunteer Application Form -->
    <Card class="p-6 transition-colors group h-full flex flex-col">
      <CardHeader class="p-0 mb-6">
        <CardTitle class="text-2xl transition-colors mb-3">{{ title || 'Volunteer Application' }}</CardTitle>
        <p v-if="description" class="text-muted-foreground">
          {{ description }}
        </p>
      </CardHeader>
      <CardContent class="p-0 flex-1 flex flex-col">
        <form @submit="handleVolunteerSubmit" class="flex flex-col h-full" novalidate>
          <div class="space-y-6 flex-1">
            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label for="firstName">First Name *</Label>
                <Input
                  id="firstName"
                  name="firstName"
                  v-model="volunteerForm.firstName"
                  @blur="() => handleVolunteerBlur('firstName', volunteerForm.firstName)"
                  :class="
                    cn(
                      volunteerTouched.firstName &&
                        volunteerErrors.firstName &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="First name"
                  maxlength="50"
                  required
                />
                <p
                  v-if="volunteerTouched.firstName && volunteerErrors.firstName"
                  class="text-sm text-red-500"
                >
                  {{ volunteerErrors.firstName }}
                </p>
              </div>
              <div class="space-y-2">
                <Label for="lastName">Last Name *</Label>
                <Input
                  id="lastName"
                  name="lastName"
                  v-model="volunteerForm.lastName"
                  @blur="() => handleVolunteerBlur('lastName', volunteerForm.lastName)"
                  :class="
                    cn(
                      volunteerTouched.lastName &&
                        volunteerErrors.lastName &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="Last name"
                  maxlength="50"
                  required
                />
                <p
                  v-if="volunteerTouched.lastName && volunteerErrors.lastName"
                  class="text-sm text-red-500"
                >
                  {{ volunteerErrors.lastName }}
                </p>
              </div>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label for="volunteer-email">Email Address *</Label>
                <Input
                  type="email"
                  id="volunteer-email"
                  name="email"
                  v-model="volunteerForm.email"
                  @blur="() => handleVolunteerBlur('email', volunteerForm.email)"
                  :class="
                    cn(
                      volunteerTouched.email &&
                        volunteerErrors.email &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="your.email@example.com"
                  maxlength="254"
                  autocomplete="email"
                  required
                />
                <p
                  v-if="volunteerTouched.email && volunteerErrors.email"
                  class="text-sm text-red-500"
                >
                  {{ volunteerErrors.email }}
                </p>
              </div>
              <div class="space-y-2">
                <Label for="volunteer-phone">Phone Number *</Label>
                <Input
                  type="tel"
                  id="volunteer-phone"
                  name="phone"
                  v-model="volunteerForm.phone"
                  @input="handleVolunteerPhoneInput"
                  @blur="() => handleVolunteerBlur('phone', volunteerForm.phone)"
                  :class="
                    cn(
                      volunteerTouched.phone &&
                        volunteerErrors.phone &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="(555) 123-4567"
                  maxlength="14"
                  autocomplete="tel"
                  required
                />
                <p
                  v-if="volunteerTouched.phone && volunteerErrors.phone"
                  class="text-sm text-red-500"
                >
                  {{ volunteerErrors.phone }}
                </p>
              </div>
            </div>

            <div class="space-y-2">
              <Label for="age">Age *</Label>
              <Input
                type="number"
                id="age"
                name="age"
                v-model="volunteerForm.age"
                @blur="() => handleVolunteerBlur('age', volunteerForm.age)"
                :class="
                  cn(
                    volunteerTouched.age &&
                      volunteerErrors.age &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                placeholder="18"
                min="18"
                max="100"
                required
              />
              <p
                v-if="volunteerTouched.age && volunteerErrors.age"
                class="text-sm text-red-500"
              >
                {{ volunteerErrors.age }}
              </p>
              <p class="text-xs text-muted-foreground">
                Must be 18 or older to volunteer
              </p>
            </div>

            <div class="space-y-2">
              <Label for="volunteerInterest">Areas of Interest *</Label>
              <Select v-model="volunteerForm.volunteerInterest" required>
                <SelectTrigger>
                  <SelectValue placeholder="Select volunteer area" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="patient-support">Patient Support</SelectItem>
                  <SelectItem value="community-outreach">Community Outreach</SelectItem>
                  <SelectItem value="research-support">Research Support</SelectItem>
                  <SelectItem value="administrative-support">Administrative Support</SelectItem>
                  <SelectItem value="multiple">Multiple Areas</SelectItem>
                  <SelectItem value="other">Other</SelectItem>
                </SelectContent>
              </Select>
              <p
                v-if="volunteerTouched.volunteerInterest && volunteerErrors.volunteerInterest"
                class="text-sm text-red-500"
              >
                {{ volunteerErrors.volunteerInterest }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="availability">Weekly Availability *</Label>
              <Select v-model="volunteerForm.availability" required>
                <SelectTrigger>
                  <SelectValue placeholder="Select time commitment" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="2-4-hours">2-4 hours per week</SelectItem>
                  <SelectItem value="4-8-hours">4-8 hours per week</SelectItem>
                  <SelectItem value="8-16-hours">8-16 hours per week</SelectItem>
                  <SelectItem value="16-hours-plus">16+ hours per week</SelectItem>
                  <SelectItem value="flexible">Flexible/As needed</SelectItem>
                </SelectContent>
              </Select>
              <p
                v-if="volunteerTouched.availability && volunteerErrors.availability"
                class="text-sm text-red-500"
              >
                {{ volunteerErrors.availability }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="experience">
                Relevant Experience <span class="text-muted-foreground">(Optional)</span>
              </Label>
              <Textarea
                id="experience"
                name="experience"
                rows="3"
                placeholder="Healthcare experience, volunteer work, education, or relevant skills..."
                v-model="volunteerForm.experience"
                maxlength="1000"
              />
              <p class="text-xs text-muted-foreground">
                {{ volunteerForm.experience.length }}/1000 characters
              </p>
            </div>

            <div class="space-y-2">
              <Label for="motivation">
                Why do you want to volunteer with us? *
                <span class="text-muted-foreground">(Min. 20 characters)</span>
              </Label>
              <Textarea
                id="motivation"
                name="motivation"
                rows="4"
                placeholder="Tell us about your motivation to volunteer and what you hope to contribute..."
                v-model="volunteerForm.motivation"
                @blur="() => handleVolunteerBlur('motivation', volunteerForm.motivation)"
                :class="
                  cn(
                    volunteerTouched.motivation &&
                      volunteerErrors.motivation &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                maxlength="1500"
                required
              />
              <p
                v-if="volunteerTouched.motivation && volunteerErrors.motivation"
                class="text-sm text-red-500"
              >
                {{ volunteerErrors.motivation }}
              </p>
              <p class="text-xs text-muted-foreground">
                {{ volunteerForm.motivation.length }}/1500 characters
              </p>
            </div>

            <div class="space-y-2">
              <Label for="schedule">
                Preferred Schedule <span class="text-muted-foreground">(Optional)</span>
              </Label>
              <Textarea
                id="schedule"
                name="schedule"
                rows="2"
                placeholder="Preferred days/times, any scheduling constraints..."
                v-model="volunteerForm.schedule"
                maxlength="500"
              />
              <p class="text-xs text-muted-foreground">
                {{ volunteerForm.schedule.length }}/500 characters
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
            v-if="submitStatus === 'success'"
            class="p-4 rounded-lg bg-green-50 border border-green-200"
          >
            <p class="text-green-800 text-sm">{{ submitMessage }}</p>
          </div>

          <div
            v-if="submitStatus === 'error'"
            class="p-4 rounded-lg bg-red-50 border border-red-200"
          >
            <p class="text-red-800 text-sm">{{ submitMessage }}</p>
          </div>

          <Button type="submit" class="w-full mt-6" size="lg" :disabled="isSubmitting">
            {{ isSubmitting ? 'Submitting...' : 'Submit Volunteer Application' }}
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
import Label from '@/components/vue-ui/Label.vue';
import { cn } from '@/lib/utils';
import { contactsClient, type ContactSubmission } from '../lib/clients';

interface VolunteerFormProps {
  className?: string;
  title?: string;
  description?: string;
}

const props = withDefaults(defineProps<VolunteerFormProps>(), {
  className: '',
  title: 'Volunteer Application',
});

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

// Validation functions
const validateEmail = (email: string): string | null => {
  if (!email) return 'Email is required';
  const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
  if (!emailRegex.test(email)) return 'Please enter a valid email address';
  return null;
};

const validatePhone = (phone: string): string | null => {
  if (!phone) return 'Phone number is required';
  // Extract digits only for validation
  const digitsOnly = phone.replace(/\D/g, '');
  // USA phone numbers should be exactly 10 digits
  if (digitsOnly.length !== 10) return 'Please enter a valid 10-digit USA phone number';
  // Check if it starts with a valid area code (not 0 or 1)
  if (digitsOnly[0] === '0' || digitsOnly[0] === '1') return 'Invalid USA phone number';
  return null;
};

const validateName = (name: string): string | null => {
  if (!name) return 'This field is required';
  if (name.length < 2) return 'Name must be at least 2 characters';
  if (name.length > 50) return 'Name must be less than 50 characters';
  if (!/^[a-zA-Z\s\-']+$/.test(name))
    return 'Name can only contain letters, spaces, hyphens, and apostrophes';
  return null;
};

const validateAge = (age: string): string | null => {
  if (!age) return 'Age is required';
  const ageNum = parseInt(age);
  if (isNaN(ageNum)) return 'Please enter a valid age';
  if (ageNum < 18) return 'Must be 18 or older to volunteer';
  if (ageNum > 100) return 'Please enter a valid age';
  return null;
};

const validateMotivation = (motivation: string): string | null => {
  if (!motivation) return 'Please tell us why you want to volunteer';
  if (motivation.length < 20) return 'Please provide at least 20 characters';
  if (motivation.length > 1500) return 'Message must be less than 1500 characters';
  return null;
};

const validateRequired = (value: string, fieldName: string): string | null => {
  if (!value) return `${fieldName} is required`;
  return null;
};

// Form state
const volunteerForm = reactive({
  firstName: '',
  lastName: '',
  email: '',
  phone: '',
  age: '',
  volunteerInterest: '',
  availability: '',
  experience: '',
  motivation: '',
  schedule: '',
});

const volunteerErrors = reactive<Record<string, string | null>>({});
const volunteerTouched = reactive<Record<string, boolean>>({});
const isSubmitting = ref(false);

// Success and error states
const submitStatus = ref<'idle' | 'success' | 'error'>('idle');
const submitMessage = ref('');

// Form validation
const validateVolunteerForm = (): boolean => {
  // Extract clean phone number for validation
  const cleanPhone = volunteerForm.phone ? cleanPhoneNumber(volunteerForm.phone) : '';

  const errors: Record<string, string | null> = {
    firstName: validateName(volunteerForm.firstName),
    lastName: validateName(volunteerForm.lastName),
    email: validateEmail(volunteerForm.email),
    phone: validatePhone(cleanPhone),
    age: validateAge(volunteerForm.age),
    volunteerInterest: validateRequired(volunteerForm.volunteerInterest, 'Areas of Interest'),
    availability: validateRequired(volunteerForm.availability, 'Weekly Availability'),
    motivation: validateMotivation(volunteerForm.motivation),
  };

  Object.assign(volunteerErrors, errors);
  return !Object.values(errors).some(error => error !== null);
};

// Handle field blur for real-time validation
const handleVolunteerBlur = (field: string, value: string) => {
  volunteerTouched[field] = true;
  const error = (() => {
    switch (field) {
      case 'firstName':
      case 'lastName':
        return validateName(value);
      case 'email':
        return validateEmail(value);
      case 'phone': {
        if (!value) return 'Phone number is required';
        // Extract clean phone number for validation
        const cleanPhone = cleanPhoneNumber(value);
        return validatePhone(cleanPhone);
      }
      case 'age':
        return validateAge(value);
      case 'volunteerInterest':
        return validateRequired(value, 'Areas of Interest');
      case 'availability':
        return validateRequired(value, 'Weekly Availability');
      case 'motivation':
        return validateMotivation(value);
      default:
        return null;
    }
  })();
  volunteerErrors[field] = error;
};

// Phone input handler
const handleVolunteerPhoneInput = (e: Event) => {
  const input = (e.target as HTMLInputElement).value;
  // If user is deleting, allow it
  if (input.length < volunteerForm.phone.length) {
    volunteerForm.phone = input;
  } else {
    // Format as USA phone number
    const filtered = filterPhoneInput(input);
    const formatted = formatUSAPhone(filtered);
    volunteerForm.phone = formatted;
  }
};

// Watchers for input filtering
watch(
  () => volunteerForm.firstName,
  newVal => {
    volunteerForm.firstName = filterNameInput(newVal);
  }
);

watch(
  () => volunteerForm.lastName,
  newVal => {
    volunteerForm.lastName = filterNameInput(newVal);
  }
);

watch(
  () => volunteerForm.email,
  newVal => {
    volunteerForm.email = filterEmailInput(newVal);
  }
);

const handleVolunteerSubmit = async (e: Event) => {
  e.preventDefault();
  isSubmitting.value = true;
  submitStatus.value = 'idle';
  submitMessage.value = '';

  // Mark all fields as touched for error display
  const allFields = Object.keys(volunteerForm);
  allFields.forEach(field => {
    volunteerTouched[field] = true;
  });

  if (!validateVolunteerForm()) {
    isSubmitting.value = false;
    return;
  }

  // Prepare data for standardized Contact API submission
  const submissionData: ContactSubmission = {
    name: `${volunteerForm.firstName} ${volunteerForm.lastName}`,
    email: volunteerForm.email,
    phone: volunteerForm.phone,
    subject: `Volunteer Application - ${volunteerForm.volunteerInterest}`,
    message: `Volunteer Application Details:

Name: ${volunteerForm.firstName} ${volunteerForm.lastName}
Age: ${volunteerForm.age}
Areas of Interest: ${volunteerForm.volunteerInterest}
Weekly Availability: ${volunteerForm.availability}

Relevant Experience:
${volunteerForm.experience || 'None specified'}

Motivation:
${volunteerForm.motivation}

Preferred Schedule:
${volunteerForm.schedule || 'None specified'}`,
  };

  try {
    // Submit using contacts client
    const response = await contactsClient.submitContact(submissionData);

    if (response.success) {
      console.log('✅ Volunteer application submitted successfully:', response);

      // Set success state
      submitStatus.value = 'success';
      const referenceMsg = response.id ? ` Your reference ID is: ${response.id}` : '';
      submitMessage.value = `Thank you for your interest in volunteering with us! Your application has been submitted successfully. We will review your application and contact you within 3-5 business days to discuss next steps.${referenceMsg}`;

      // Reset form on success
      Object.assign(volunteerForm, {
        firstName: '',
        lastName: '',
        email: '',
        phone: '',
        age: '',
        volunteerInterest: '',
        availability: '',
        experience: '',
        motivation: '',
        schedule: '',
      });
      Object.keys(volunteerErrors).forEach(key => (volunteerErrors[key] = null));
      Object.keys(volunteerTouched).forEach(key => (volunteerTouched[key] = false));
    } else {
      // Handle error state
      submitStatus.value = 'error';
      submitMessage.value =
        response.message ||
        'Unable to submit your volunteer application at this time. Please try again later or contact us directly.';
    }
  } catch (err) {
    console.error('Volunteer application submission error:', err);
    submitStatus.value = 'error';
    submitMessage.value = 'Network error occurred. Please try again later or contact us directly.';
  } finally {
    isSubmitting.value = false;
  }
};
</script>