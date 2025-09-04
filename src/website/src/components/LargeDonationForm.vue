<template>
  <div class="p-6 bg-white rounded-sm border border-gray-200">
    <div class="mb-6">
      <h3 class="text-xl font-semibold text-gray-900 mb-2">Large Donation Consultation</h3>
      <p class="text-gray-600 text-sm">
        For donations over $10,000, discuss customized giving options and recognition opportunities.
      </p>
    </div>

    <form @submit="handleSubmit" class="space-y-4" novalidate>
      <!-- Contact Information -->
      <div class="grid gap-3 md:grid-cols-2">
        <div class="space-y-2">
          <label for="largeFirstName" class="text-sm font-medium text-gray-900">First Name</label>
          <input
            id="largeFirstName"
            v-model="form.firstName"
            @blur="() => handleBlur('firstName', form.firstName)"
            :class="[
              'w-full px-3 py-2 border rounded text-sm',
              touched.firstName && errors.firstName
                ? 'border-red-500'
                : 'border-gray-200'
            ]"
            placeholder="First name"
            maxlength="50"
            required
          />
          <p v-if="touched.firstName && errors.firstName" class="text-red-500 text-xs">
            {{ errors.firstName }}
          </p>
        </div>

        <div class="space-y-2">
          <label for="largeLastName" class="text-sm font-medium text-gray-900">Last Name</label>
          <input
            id="largeLastName"
            v-model="form.lastName"
            @blur="() => handleBlur('lastName', form.lastName)"
            :class="[
              'w-full px-3 py-2 border rounded text-sm',
              touched.lastName && errors.lastName
                ? 'border-red-500'
                : 'border-gray-200'
            ]"
            placeholder="Last name"
            maxlength="50"
            required
          />
          <p v-if="touched.lastName && errors.lastName" class="text-red-500 text-xs">
            {{ errors.lastName }}
          </p>
        </div>
      </div>

      <div class="space-y-2">
        <label for="largeEmail" class="text-sm font-medium text-gray-900">Email Address</label>
        <input
          id="largeEmail"
          type="email"
          v-model="form.email"
          @blur="() => handleBlur('email', form.email)"
          :class="[
            'w-full px-3 py-2 border rounded text-sm',
            touched.email && errors.email
              ? 'border-red-500'
              : 'border-gray-200'
          ]"
          placeholder="your.email@example.com"
          maxlength="100"
          required
        />
        <p v-if="touched.email && errors.email" class="text-red-500 text-xs">
          {{ errors.email }}
        </p>
      </div>

      <div class="space-y-2">
        <label for="largePhone" class="text-sm font-medium text-gray-900">Phone Number</label>
        <input
          id="largePhone"
          type="tel"
          v-model="form.phone"
          @blur="() => handleBlur('phone', form.phone)"
          :class="[
            'w-full px-3 py-2 border rounded text-sm',
            touched.phone && errors.phone
              ? 'border-red-500'
              : 'border-gray-200'
          ]"
          placeholder="(555) 123-4567"
          maxlength="20"
          required
        />
        <p v-if="touched.phone && errors.phone" class="text-red-500 text-xs">
          {{ errors.phone }}
        </p>
      </div>

      <!-- Donation Interest -->
      <div class="space-y-2">
        <label for="donationInterest" class="text-sm font-medium text-gray-900">Area of Interest</label>
        <select
          id="donationInterest"
          v-model="form.interest"
          @blur="() => handleBlur('interest', form.interest)"
          :class="[
            'w-full px-3 py-2 border rounded text-sm',
            touched.interest && errors.interest
              ? 'border-red-500'
              : 'border-gray-200'
          ]"
          required
        >
          <option value="">Select an area</option>
          <option value="facility-development">Facility Development</option>
          <option value="equipment-technology">Equipment & Technology</option>
          <option value="research-innovation">Research & Innovation</option>
          <option value="patient-assistance">Patient Assistance Programs</option>
          <option value="education-training">Education & Training</option>
          <option value="general-support">General Support</option>
        </select>
        <p v-if="touched.interest && errors.interest" class="text-red-500 text-xs">
          {{ errors.interest }}
        </p>
      </div>

      <!-- Estimated Amount -->
      <div class="space-y-2">
        <label for="estimatedAmount" class="text-sm font-medium text-gray-900">Estimated Donation Amount</label>
        <select
          id="estimatedAmount"
          v-model="form.amount"
          @blur="() => handleBlur('amount', form.amount)"
          :class="[
            'w-full px-3 py-2 border rounded text-sm',
            touched.amount && errors.amount
              ? 'border-red-500'
              : 'border-gray-200'
          ]"
          required
        >
          <option value="">Select range</option>
          <option value="10000-25000">$10,000 - $25,000</option>
          <option value="25000-50000">$25,000 - $50,000</option>
          <option value="50000-100000">$50,000 - $100,000</option>
          <option value="100000-250000">$100,000 - $250,000</option>
          <option value="250000-500000">$250,000 - $500,000</option>
          <option value="500000+">$500,000+</option>
        </select>
        <p v-if="touched.amount && errors.amount" class="text-red-500 text-xs">
          {{ errors.amount }}
        </p>
      </div>

      <!-- Message -->
      <div class="space-y-2">
        <label for="largeMessage" class="text-sm font-medium text-gray-900">Message (Optional)</label>
        <textarea
          id="largeMessage"
          v-model="form.message"
          @blur="() => handleBlur('message', form.message)"
          :class="[
            'w-full px-3 py-2 border rounded text-sm',
            touched.message && errors.message
              ? 'border-red-500'
              : 'border-gray-200'
          ]"
          placeholder="Tell us about your philanthropic goals or specific interests..."
          maxlength="500"
          rows="3"
        ></textarea>
        <p class="text-xs text-gray-500">
          {{ form.message.length }}/500 characters
        </p>
        <p v-if="touched.message && errors.message" class="text-red-500 text-xs">
          {{ errors.message }}
        </p>
      </div>

      <!-- Honeypot field -->
      <input
        type="text"
        name="company"
        tabindex="-1"
        autocomplete="off"
        style="position: absolute; left: -9999px; width: 1px; height: 1px; opacity: 0; visibility: hidden;"
        aria-hidden="true"
      />

      <!-- Submit Messages -->
      <div v-if="submitStatus === 'success'" class="p-3 rounded bg-green-50 border border-green-200">
        <p class="text-green-800 text-sm">{{ submitMessage }}</p>
      </div>

      <div v-if="submitStatus === 'error'" class="p-3 rounded bg-red-50 border border-red-200">
        <p class="text-red-800 text-sm">{{ submitMessage }}</p>
      </div>

      <button
        type="submit"
        :disabled="isSubmitting"
        class="w-full py-3 px-4 bg-blue-600 text-white font-medium rounded hover:bg-blue-700 transition-colors disabled:opacity-50"
      >
        {{ isSubmitting ? 'Submitting...' : 'Request Consultation' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { contactsClient } from '../lib/clients';

// Form data
const form = reactive({
  firstName: '',
  lastName: '',
  email: '',
  phone: '',
  interest: '',
  amount: '',
  message: '',
});

const touched = reactive({
  firstName: false,
  lastName: false,
  email: false,
  phone: false,
  interest: false,
  amount: false,
  message: false,
});

const errors = reactive({
  firstName: null as string | null,
  lastName: null as string | null,
  email: null as string | null,
  phone: null as string | null,
  interest: null as string | null,
  amount: null as string | null,
  message: null as string | null,
});

// Submission state
const isSubmitting = ref(false);
const submitStatus = ref<'idle' | 'success' | 'error'>('idle');
const submitMessage = ref('');

// Validation functions
const validateName = (name: string, fieldName: string): string | null => {
  if (!name || name.trim() === '') {
    return `${fieldName} is required`;
  }
  if (name.trim().length < 2) {
    return `${fieldName} must be at least 2 characters`;
  }
  if (name.trim().length > 50) {
    return `${fieldName} must be 50 characters or less`;
  }
  return null;
};

const validateEmail = (email: string): string | null => {
  if (!email || email.trim() === '') {
    return 'Email is required';
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email.trim())) {
    return 'Please enter a valid email address';
  }
  return null;
};

const validatePhone = (phone: string): string | null => {
  if (!phone || phone.trim() === '') {
    return 'Phone number is required';
  }
  const phoneRegex = /^[\+]?[1-9][\d]{0,15}$/;
  const cleanPhone = phone.replace(/[\s\-\(\)\.]/g, '');
  if (!phoneRegex.test(cleanPhone)) {
    return 'Please enter a valid phone number';
  }
  return null;
};

const validateRequired = (value: string, fieldName: string): string | null => {
  if (!value || value.trim() === '') {
    return `${fieldName} is required`;
  }
  return null;
};

const validateMessage = (message: string): string | null => {
  if (message && message.length > 500) {
    return 'Message must be 500 characters or less';
  }
  return null;
};

// Handle field blur validation
const handleBlur = (field: keyof typeof form, value: string) => {
  touched[field as keyof typeof touched] = true;
  
  switch (field) {
    case 'firstName':
      errors.firstName = validateName(value, 'First name');
      break;
    case 'lastName':
      errors.lastName = validateName(value, 'Last name');
      break;
    case 'email':
      errors.email = validateEmail(value);
      break;
    case 'phone':
      errors.phone = validatePhone(value);
      break;
    case 'interest':
      errors.interest = validateRequired(value, 'Area of interest');
      break;
    case 'amount':
      errors.amount = validateRequired(value, 'Estimated amount');
      break;
    case 'message':
      errors.message = validateMessage(value);
      break;
  }
};

// Form submission
const handleSubmit = async (event: Event) => {
  event.preventDefault();
  
  // Mark all fields as touched
  Object.keys(touched).forEach(key => {
    touched[key as keyof typeof touched] = true;
  });

  // Validate all fields
  const validationErrors: Record<string, string | null> = {
    firstName: validateName(form.firstName, 'First name'),
    lastName: validateName(form.lastName, 'Last name'),
    email: validateEmail(form.email),
    phone: validatePhone(form.phone),
    interest: validateRequired(form.interest, 'Area of interest'),
    amount: validateRequired(form.amount, 'Estimated amount'),
    message: validateMessage(form.message),
  };

  // Update error state
  Object.keys(validationErrors).forEach(key => {
    errors[key as keyof typeof errors] = validationErrors[key];
  });

  // Check for errors
  const hasErrors = Object.values(validationErrors).some(error => error !== null);
  if (hasErrors) {
    submitStatus.value = 'error';
    submitMessage.value = 'Please correct the errors above and try again.';
    return;
  }

  // Check honeypot
  const honeypotField = (event.target as HTMLFormElement).querySelector('[name="company"]') as HTMLInputElement;
  if (honeypotField && honeypotField.value) {
    submitStatus.value = 'error';
    submitMessage.value = 'Invalid submission detected.';
    return;
  }

  isSubmitting.value = true;
  submitStatus.value = 'idle';

  try {
    const consultationData = {
      type: 'large-donation-consultation',
      firstName: form.firstName.trim(),
      lastName: form.lastName.trim(),
      email: form.email.trim(),
      phone: form.phone.trim(),
      interest: form.interest,
      estimatedAmount: form.amount,
      message: form.message.trim() || null,
      submittedAt: new Date().toISOString(),
    };

    const response = await contactsClient.submitContact(consultationData as any);
    
    if (response.success) {
      submitStatus.value = 'success';
      submitMessage.value = 'Thank you for your interest! Our team will contact you within 2 business days to discuss your donation and partnership opportunities.';
      
      // Reset form
      Object.keys(form).forEach(key => {
        (form as any)[key] = '';
      });
      
      Object.keys(touched).forEach(key => {
        touched[key as keyof typeof touched] = false;
      });
      
      Object.keys(errors).forEach(key => {
        errors[key as keyof typeof errors] = null;
      });
    } else {
      submitStatus.value = 'error';
      submitMessage.value = response.message || 'Failed to submit consultation request. Please try again.';
    }
  } catch (error) {
    console.error('Large donation consultation submission error:', error);
    submitStatus.value = 'error';
    submitMessage.value = 'Network error occurred. Please try again later.';
  } finally {
    isSubmitting.value = false;
  }
};
</script>