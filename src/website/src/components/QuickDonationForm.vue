<template>
  <div class="p-6 bg-white rounded-sm border border-gray-200">
    <div class="mb-6">
      <h3 class="text-xl font-semibold text-gray-900 mb-2">Quick Donation</h3>
      <p class="text-gray-600 text-sm">
        Make a secure online donation with instant processing via Stripe.
      </p>
    </div>

    <form @submit="handleSubmit" class="space-y-4" novalidate>
      <!-- Preset Amount Buttons -->
      <div class="space-y-3">
        <label class="text-sm font-medium text-gray-900">Amount</label>
        <div class="grid grid-cols-2 gap-3">
          <button
            v-for="amount in presetAmounts"
            :key="amount"
            type="button"
            :class="[
              'py-3 px-4 text-sm font-medium rounded border transition-colors',
              form.amount === amount.toString()
                ? 'bg-blue-600 text-white border-blue-600'
                : 'bg-white text-gray-700 border-gray-200 hover:border-gray-300'
            ]"
            @click="selectAmount(amount)"
          >
            ${{ amount }}
          </button>
        </div>
      </div>

      <!-- Custom Amount -->
      <div class="space-y-2">
        <label for="quickCustomAmount" class="text-sm font-medium text-gray-900">Custom Amount</label>
        <div class="relative">
          <span class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-600 text-sm">$</span>
          <input
            id="quickCustomAmount"
            type="number"
            min="1"
            step="1"
            v-model="form.amount"
            @blur="() => handleBlur('amount', form.amount)"
            :class="[
              'w-full pl-7 pr-3 py-2 border rounded text-sm',
              touched.amount && errors.amount
                ? 'border-red-500'
                : 'border-gray-200'
            ]"
            placeholder="Enter amount"
            required
          />
        </div>
        <p v-if="touched.amount && errors.amount" class="text-red-500 text-xs">
          {{ errors.amount }}
        </p>
      </div>

      <!-- Contact Information -->
      <div class="grid gap-3 md:grid-cols-2">
        <div class="space-y-2">
          <label for="quickFirstName" class="text-sm font-medium text-gray-900">First Name</label>
          <input
            id="quickFirstName"
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
          <label for="quickLastName" class="text-sm font-medium text-gray-900">Last Name</label>
          <input
            id="quickLastName"
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
        <label for="quickEmail" class="text-sm font-medium text-gray-900">Email Address</label>
        <input
          id="quickEmail"
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

      <!-- Anonymous Option -->
      <div class="flex items-center space-x-3">
        <input
          id="quickAnonymous"
          type="checkbox"
          v-model="form.anonymous"
          class="h-4 w-4 text-blue-600 border-gray-300 rounded"
        />
        <label for="quickAnonymous" class="text-sm text-gray-600">
          Make this donation anonymous
        </label>
      </div>

      <!-- Honeypot field -->
      <input
        type="text"
        name="website"
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
        {{ isSubmitting ? 'Processing...' : 'Donate Now' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { contactsClient } from '../lib/clients';

// Preset amounts
const presetAmounts = [25, 50, 100, 250, 500, 1000];

// Form data
const form = reactive({
  amount: '',
  firstName: '',
  lastName: '',
  email: '',
  anonymous: false,
});

const touched = reactive({
  amount: false,
  firstName: false,
  lastName: false,
  email: false,
});

const errors = reactive({
  amount: null as string | null,
  firstName: null as string | null,
  lastName: null as string | null,
  email: null as string | null,
});

// Submission state
const isSubmitting = ref(false);
const submitStatus = ref<'idle' | 'success' | 'error'>('idle');
const submitMessage = ref('');

// Validation functions
const validateAmount = (amount: string): string | null => {
  if (!amount || amount.trim() === '') {
    return 'Amount is required';
  }
  const numAmount = parseFloat(amount);
  if (isNaN(numAmount) || numAmount <= 0) {
    return 'Please enter a valid amount';
  }
  if (numAmount < 1) {
    return 'Minimum amount is $1';
  }
  if (numAmount > 50000) {
    return 'For amounts over $50,000, please use the consultation form';
  }
  return null;
};

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

// Select preset amount
const selectAmount = (amount: number) => {
  form.amount = amount.toString();
  touched.amount = true;
  errors.amount = validateAmount(form.amount);
};

// Handle field blur validation
const handleBlur = (field: keyof typeof form, value: string | boolean) => {
  touched[field as keyof typeof touched] = true;
  
  switch (field) {
    case 'amount':
      errors.amount = validateAmount(value as string);
      break;
    case 'firstName':
      errors.firstName = validateName(value as string, 'First name');
      break;
    case 'lastName':
      errors.lastName = validateName(value as string, 'Last name');
      break;
    case 'email':
      errors.email = validateEmail(value as string);
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
    amount: validateAmount(form.amount),
    firstName: validateName(form.firstName, 'First name'),
    lastName: validateName(form.lastName, 'Last name'),
    email: validateEmail(form.email),
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
  const honeypotField = (event.target as HTMLFormElement).querySelector('[name="website"]') as HTMLInputElement;
  if (honeypotField && honeypotField.value) {
    submitStatus.value = 'error';
    submitMessage.value = 'Invalid submission detected.';
    return;
  }

  isSubmitting.value = true;
  submitStatus.value = 'idle';

  try {
    const donationData = {
      type: 'quick-donation',
      amount: parseFloat(form.amount),
      firstName: form.firstName.trim(),
      lastName: form.lastName.trim(),
      email: form.email.trim(),
      anonymous: form.anonymous,
      submittedAt: new Date().toISOString(),
    };

    const response = await contactsClient.submitContact(donationData as any);
    
    if (response.success) {
      submitStatus.value = 'success';
      submitMessage.value = 'Thank you for your donation! You will be redirected to complete payment shortly.';
      
      // Reset form
      Object.keys(form).forEach(key => {
        if (typeof form[key as keyof typeof form] === 'string') {
          (form as any)[key] = '';
        } else if (typeof form[key as keyof typeof form] === 'boolean') {
          (form as any)[key] = false;
        }
      });
      
      Object.keys(touched).forEach(key => {
        touched[key as keyof typeof touched] = false;
      });
      
      Object.keys(errors).forEach(key => {
        errors[key as keyof typeof errors] = null;
      });
    } else {
      submitStatus.value = 'error';
      submitMessage.value = response.message || 'Failed to submit donation. Please try again.';
    }
  } catch (error) {
    console.error('Quick donation submission error:', error);
    submitStatus.value = 'error';
    submitMessage.value = 'Network error occurred. Please try again later.';
  } finally {
    isSubmitting.value = false;
  }
};
</script>