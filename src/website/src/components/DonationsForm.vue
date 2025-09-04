<template>
  <div :class="`max-w-4xl mx-auto ${className}`">
    <!-- Donation Form -->
    <Card class="p-6 transition-colors group">
      <CardHeader class="p-0 mb-6">
        <CardTitle class="text-3xl transition-colors mb-4">Support Our Mission</CardTitle>
        <div class="space-y-3">
          <p class="text-lg text-muted-foreground">
            Your donation helps us create and improve our regenerative medicine clinic, advancing patient care and expanding access to innovative treatments.
          </p>
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p class="text-blue-800 text-sm font-medium">
              🏥 Your contributions directly support:
            </p>
            <ul class="text-blue-700 text-sm mt-2 space-y-1 ml-4">
              <li>• Advanced medical equipment and technology</li>
              <li>• Clinic facility development and improvements</li>
              <li>• Research and development of new treatments</li>
              <li>• Patient care programs and accessibility initiatives</li>
            </ul>
          </div>
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <form @submit="handleDonationSubmit" class="space-y-6" novalidate>
          <!-- Donation Amount Section -->
          <div class="space-y-4">
            <div class="space-y-2">
              <Label class="text-lg font-semibold">Donation Amount *</Label>
              <p class="text-sm text-muted-foreground">Select a preset amount or enter a custom amount</p>
            </div>
            
            <!-- Preset Amount Buttons -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
              <Button
                v-for="amount in presetAmounts"
                :key="amount"
                type="button"
                :variant="donationForm.amount === amount.toString() ? 'default' : 'outline'"
                @click="selectPresetAmount(amount)"
                class="h-12 text-base font-semibold"
              >
                ${{ amount }}
              </Button>
            </div>

            <!-- Custom Amount Input -->
            <div class="space-y-2">
              <Label for="customAmount">Custom Amount</Label>
              <div class="relative">
                <span class="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground">$</span>
                <Input
                  id="customAmount"
                  name="customAmount"
                  type="number"
                  min="1"
                  step="1"
                  v-model="donationForm.amount"
                  @blur="() => handleDonationBlur('amount', donationForm.amount)"
                  :class="
                    cn(
                      'pl-7',
                      donationTouched.amount &&
                        donationErrors.amount &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="Enter amount"
                  required
                />
              </div>
              <p
                v-if="donationTouched.amount && donationErrors.amount"
                class="text-sm text-red-500"
              >
                {{ donationErrors.amount }}
              </p>
            </div>
          </div>

          <!-- Donation Frequency -->
          <div class="space-y-2">
            <Label for="frequency">Donation Frequency *</Label>
            <Select v-model="donationForm.frequency" required>
              <SelectTrigger
                id="frequency"
                name="frequency"
                :class="
                  cn(
                    donationTouched.frequency &&
                      donationErrors.frequency &&
                      'border-red-500 focus:border-red-500'
                  )
                "
              >
                <SelectValue placeholder="Select frequency" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="one-time">One-time donation</SelectItem>
                <SelectItem value="monthly">Monthly recurring</SelectItem>
                <SelectItem value="quarterly">Quarterly recurring</SelectItem>
                <SelectItem value="annually">Annual recurring</SelectItem>
              </SelectContent>
            </Select>
            <p
              v-if="donationTouched.frequency && donationErrors.frequency"
              class="text-sm text-red-500"
            >
              {{ donationErrors.frequency }}
            </p>
          </div>

          <!-- Donor Information -->
          <div class="space-y-4">
            <Label class="text-lg font-semibold">Donor Information</Label>
            
            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label for="firstName">First Name *</Label>
                <Input
                  id="firstName"
                  name="firstName"
                  v-model="donationForm.firstName"
                  @blur="() => handleDonationBlur('firstName', donationForm.firstName)"
                  :class="
                    cn(
                      donationTouched.firstName &&
                        donationErrors.firstName &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="Your first name"
                  maxlength="50"
                  required
                />
                <p
                  v-if="donationTouched.firstName && donationErrors.firstName"
                  class="text-sm text-red-500"
                >
                  {{ donationErrors.firstName }}
                </p>
              </div>

              <div class="space-y-2">
                <Label for="lastName">Last Name *</Label>
                <Input
                  id="lastName"
                  name="lastName"
                  v-model="donationForm.lastName"
                  @blur="() => handleDonationBlur('lastName', donationForm.lastName)"
                  :class="
                    cn(
                      donationTouched.lastName &&
                        donationErrors.lastName &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="Your last name"
                  maxlength="50"
                  required
                />
                <p
                  v-if="donationTouched.lastName && donationErrors.lastName"
                  class="text-sm text-red-500"
                >
                  {{ donationErrors.lastName }}
                </p>
              </div>
            </div>

            <div class="space-y-2">
              <Label for="email">Email Address *</Label>
              <Input
                id="email"
                name="email"
                type="email"
                v-model="donationForm.email"
                @blur="() => handleDonationBlur('email', donationForm.email)"
                :class="
                  cn(
                    donationTouched.email &&
                      donationErrors.email &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                placeholder="your.email@example.com"
                maxlength="100"
                required
              />
              <p
                v-if="donationTouched.email && donationErrors.email"
                class="text-sm text-red-500"
              >
                {{ donationErrors.email }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="phone">Phone Number (Optional)</Label>
              <Input
                id="phone"
                name="phone"
                type="tel"
                v-model="donationForm.phone"
                @blur="() => handleDonationBlur('phone', donationForm.phone)"
                :class="
                  cn(
                    donationTouched.phone &&
                      donationErrors.phone &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                placeholder="(555) 123-4567"
                maxlength="20"
              />
              <p
                v-if="donationTouched.phone && donationErrors.phone"
                class="text-sm text-red-500"
              >
                {{ donationErrors.phone }}
              </p>
            </div>
          </div>

          <!-- Address Information for Tax Receipt -->
          <div class="space-y-4">
            <div class="space-y-2">
              <Label class="text-lg font-semibold">Address for Tax Receipt</Label>
              <p class="text-sm text-muted-foreground">Required for donation receipts and tax documentation</p>
            </div>

            <div class="space-y-2">
              <Label for="address">Street Address *</Label>
              <Input
                id="address"
                name="address"
                v-model="donationForm.address"
                @blur="() => handleDonationBlur('address', donationForm.address)"
                :class="
                  cn(
                    donationTouched.address &&
                      donationErrors.address &&
                      'border-red-500 focus:border-red-500'
                  )
                "
                placeholder="123 Main Street"
                maxlength="100"
                required
              />
              <p
                v-if="donationTouched.address && donationErrors.address"
                class="text-sm text-red-500"
              >
                {{ donationErrors.address }}
              </p>
            </div>

            <div class="grid gap-4 md:grid-cols-3">
              <div class="space-y-2">
                <Label for="city">City *</Label>
                <Input
                  id="city"
                  name="city"
                  v-model="donationForm.city"
                  @blur="() => handleDonationBlur('city', donationForm.city)"
                  :class="
                    cn(
                      donationTouched.city &&
                        donationErrors.city &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="City"
                  maxlength="50"
                  required
                />
                <p
                  v-if="donationTouched.city && donationErrors.city"
                  class="text-sm text-red-500"
                >
                  {{ donationErrors.city }}
                </p>
              </div>

              <div class="space-y-2">
                <Label for="state">State *</Label>
                <Input
                  id="state"
                  name="state"
                  v-model="donationForm.state"
                  @blur="() => handleDonationBlur('state', donationForm.state)"
                  :class="
                    cn(
                      donationTouched.state &&
                        donationErrors.state &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="FL"
                  maxlength="2"
                  required
                />
                <p
                  v-if="donationTouched.state && donationErrors.state"
                  class="text-sm text-red-500"
                >
                  {{ donationErrors.state }}
                </p>
              </div>

              <div class="space-y-2">
                <Label for="zipCode">ZIP Code *</Label>
                <Input
                  id="zipCode"
                  name="zipCode"
                  v-model="donationForm.zipCode"
                  @blur="() => handleDonationBlur('zipCode', donationForm.zipCode)"
                  :class="
                    cn(
                      donationTouched.zipCode &&
                        donationErrors.zipCode &&
                        'border-red-500 focus:border-red-500'
                    )
                  "
                  placeholder="12345"
                  maxlength="10"
                  required
                />
                <p
                  v-if="donationTouched.zipCode && donationErrors.zipCode"
                  class="text-sm text-red-500"
                >
                  {{ donationErrors.zipCode }}
                </p>
              </div>
            </div>
          </div>

          <!-- Optional Dedication/Message -->
          <div class="space-y-2">
            <Label for="dedication">Dedication or Message (Optional)</Label>
            <Textarea
              id="dedication"
              name="dedication"
              v-model="donationForm.dedication"
              @blur="() => handleDonationBlur('dedication', donationForm.dedication)"
              :class="
                cn(
                  donationTouched.dedication &&
                    donationErrors.dedication &&
                    'border-red-500 focus:border-red-500'
                )
              "
              placeholder="In honor of... or leave a message about your support"
              maxlength="500"
              rows="3"
            />
            <p class="text-xs text-muted-foreground">
              {{ donationForm.dedication.length }}/500 characters
            </p>
            <p
              v-if="donationTouched.dedication && donationErrors.dedication"
              class="text-sm text-red-500"
            >
              {{ donationErrors.dedication }}
            </p>
          </div>

          <!-- Contact Preferences -->
          <div class="space-y-4">
            <Label class="text-lg font-semibold">Communication Preferences</Label>
            
            <div class="space-y-3">
              <div class="flex items-start space-x-3">
                <input
                  id="anonymousDonation"
                  name="anonymousDonation"
                  type="checkbox"
                  v-model="donationForm.anonymousDonation"
                  class="mt-1 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <div class="flex-1">
                  <Label for="anonymousDonation" class="text-sm font-medium cursor-pointer">
                    Make this an anonymous donation
                  </Label>
                  <p class="text-xs text-muted-foreground">
                    Your name will not be publicly disclosed in donor acknowledgments
                  </p>
                </div>
              </div>

              <div class="flex items-start space-x-3">
                <input
                  id="receiveUpdates"
                  name="receiveUpdates"
                  type="checkbox"
                  v-model="donationForm.receiveUpdates"
                  class="mt-1 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <div class="flex-1">
                  <Label for="receiveUpdates" class="text-sm font-medium cursor-pointer">
                    Receive updates about our clinic progress
                  </Label>
                  <p class="text-xs text-muted-foreground">
                    Get occasional updates about how your donation is making an impact
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Honeypot field for spam protection -->
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

          <!-- Payment Information Notice -->
          <div class="bg-amber-50 border border-amber-200 rounded-lg p-4">
            <p class="text-amber-800 text-sm font-medium">
              📋 Next Steps:
            </p>
            <p class="text-amber-700 text-sm mt-1">
              After submitting this form, you will be contacted with secure payment processing instructions and donation receipt information.
            </p>
          </div>

          <Button
            type="submit"
            class="w-full"
            size="lg"
            variant="default"
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? 'Submitting...' : 'Submit Donation Information' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
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
import { contactsClient } from '../lib/clients';

interface DonationsFormProps {
  className?: string;
}

const props = withDefaults(defineProps<DonationsFormProps>(), {
  className: '',
});

// Preset donation amounts
const presetAmounts = [25, 50, 100, 250, 500, 1000, 2500, 5000];

// Form data
const donationForm = reactive({
  amount: '',
  frequency: '',
  firstName: '',
  lastName: '',
  email: '',
  phone: '',
  address: '',
  city: '',
  state: '',
  zipCode: '',
  dedication: '',
  anonymousDonation: false,
  receiveUpdates: true,
});

// Form validation state
const donationTouched = reactive({
  amount: false,
  frequency: false,
  firstName: false,
  lastName: false,
  email: false,
  phone: false,
  address: false,
  city: false,
  state: false,
  zipCode: false,
  dedication: false,
});

const donationErrors = reactive({
  amount: null as string | null,
  frequency: null as string | null,
  firstName: null as string | null,
  lastName: null as string | null,
  email: null as string | null,
  phone: null as string | null,
  address: null as string | null,
  city: null as string | null,
  state: null as string | null,
  zipCode: null as string | null,
  dedication: null as string | null,
});

// Submission state
const isSubmitting = ref(false);
const submitStatus = ref<'idle' | 'success' | 'error'>('idle');
const submitMessage = ref('');

// Select preset amount
const selectPresetAmount = (amount: number) => {
  donationForm.amount = amount.toString();
  donationTouched.amount = true;
  donationErrors.amount = validateAmount(donationForm.amount);
};

// Validation functions
const validateAmount = (amount: string): string | null => {
  if (!amount || amount.trim() === '') {
    return 'Donation amount is required';
  }
  const numAmount = parseFloat(amount);
  if (isNaN(numAmount) || numAmount <= 0) {
    return 'Please enter a valid donation amount';
  }
  if (numAmount < 1) {
    return 'Minimum donation amount is $1';
  }
  if (numAmount > 100000) {
    return 'Maximum donation amount is $100,000';
  }
  return null;
};

const validateFrequency = (frequency: string): string | null => {
  if (!frequency || frequency.trim() === '') {
    return 'Please select a donation frequency';
  }
  const validFrequencies = ['one-time', 'monthly', 'quarterly', 'annually'];
  if (!validFrequencies.includes(frequency)) {
    return 'Please select a valid donation frequency';
  }
  return null;
};

const validateName = (name: string, fieldName: string): string | null => {
  if (!name || name.trim() === '') {
    return `${fieldName} is required`;
  }
  if (name.trim().length < 2) {
    return `${fieldName} must be at least 2 characters long`;
  }
  if (name.trim().length > 50) {
    return `${fieldName} must be 50 characters or less`;
  }
  if (!/^[a-zA-Z\s'-]+$/.test(name.trim())) {
    return `${fieldName} can only contain letters, spaces, hyphens, and apostrophes`;
  }
  return null;
};

const validateEmail = (email: string): string | null => {
  if (!email || email.trim() === '') {
    return 'Email address is required';
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email.trim())) {
    return 'Please enter a valid email address';
  }
  if (email.trim().length > 100) {
    return 'Email address must be 100 characters or less';
  }
  return null;
};

const validatePhone = (phone: string): string | null => {
  if (!phone || phone.trim() === '') {
    return null; // Phone is optional
  }
  const phoneRegex = /^[\+]?[1-9][\d]{0,15}$/;
  const cleanPhone = phone.replace(/[\s\-\(\)\.]/g, '');
  if (!phoneRegex.test(cleanPhone)) {
    return 'Please enter a valid phone number';
  }
  return null;
};

const validateAddress = (address: string): string | null => {
  if (!address || address.trim() === '') {
    return 'Street address is required for tax receipts';
  }
  if (address.trim().length < 5) {
    return 'Street address must be at least 5 characters long';
  }
  if (address.trim().length > 100) {
    return 'Street address must be 100 characters or less';
  }
  return null;
};

const validateCity = (city: string): string | null => {
  if (!city || city.trim() === '') {
    return 'City is required';
  }
  if (city.trim().length < 2) {
    return 'City must be at least 2 characters long';
  }
  if (city.trim().length > 50) {
    return 'City must be 50 characters or less';
  }
  return null;
};

const validateState = (state: string): string | null => {
  if (!state || state.trim() === '') {
    return 'State is required';
  }
  if (state.trim().length !== 2) {
    return 'Please enter a valid 2-letter state code (e.g., FL)';
  }
  if (!/^[A-Z]{2}$/i.test(state.trim())) {
    return 'State must be a 2-letter code (e.g., FL)';
  }
  return null;
};

const validateZipCode = (zipCode: string): string | null => {
  if (!zipCode || zipCode.trim() === '') {
    return 'ZIP code is required';
  }
  if (!/^\d{5}(-\d{4})?$/.test(zipCode.trim())) {
    return 'Please enter a valid ZIP code (e.g., 12345 or 12345-6789)';
  }
  return null;
};

const validateDedication = (dedication: string): string | null => {
  if (dedication && dedication.length > 500) {
    return 'Dedication message must be 500 characters or less';
  }
  return null;
};

// Handle field blur validation
const handleDonationBlur = (field: keyof typeof donationForm, value: string | boolean) => {
  donationTouched[field as keyof typeof donationTouched] = true;
  
  switch (field) {
    case 'amount':
      donationErrors.amount = validateAmount(value as string);
      break;
    case 'frequency':
      donationErrors.frequency = validateFrequency(value as string);
      break;
    case 'firstName':
      donationErrors.firstName = validateName(value as string, 'First name');
      break;
    case 'lastName':
      donationErrors.lastName = validateName(value as string, 'Last name');
      break;
    case 'email':
      donationErrors.email = validateEmail(value as string);
      break;
    case 'phone':
      donationErrors.phone = validatePhone(value as string);
      break;
    case 'address':
      donationErrors.address = validateAddress(value as string);
      break;
    case 'city':
      donationErrors.city = validateCity(value as string);
      break;
    case 'state':
      donationErrors.state = validateState(value as string);
      break;
    case 'zipCode':
      donationErrors.zipCode = validateZipCode(value as string);
      break;
    case 'dedication':
      donationErrors.dedication = validateDedication(value as string);
      break;
  }
};

// Form submission
const handleDonationSubmit = async (event: Event) => {
  event.preventDefault();
  
  // Mark all fields as touched for validation
  Object.keys(donationTouched).forEach(key => {
    donationTouched[key as keyof typeof donationTouched] = true;
  });

  // Validate all fields
  const errors: Record<string, string | null> = {
    amount: validateAmount(donationForm.amount),
    frequency: validateFrequency(donationForm.frequency),
    firstName: validateName(donationForm.firstName, 'First name'),
    lastName: validateName(donationForm.lastName, 'Last name'),
    email: validateEmail(donationForm.email),
    phone: validatePhone(donationForm.phone),
    address: validateAddress(donationForm.address),
    city: validateCity(donationForm.city),
    state: validateState(donationForm.state),
    zipCode: validateZipCode(donationForm.zipCode),
    dedication: validateDedication(donationForm.dedication),
  };

  // Update error state
  Object.keys(errors).forEach(key => {
    donationErrors[key as keyof typeof donationErrors] = errors[key];
  });

  // Check if there are any errors
  const hasErrors = Object.values(errors).some(error => error !== null);
  if (hasErrors) {
    submitStatus.value = 'error';
    submitMessage.value = 'Please correct the errors above and try again.';
    return;
  }

  // Check for honeypot spam protection
  const honeypotField = (event.target as HTMLFormElement).querySelector('[name="website"]') as HTMLInputElement;
  if (honeypotField && honeypotField.value) {
    submitStatus.value = 'error';
    submitMessage.value = 'Invalid submission detected.';
    return;
  }

  isSubmitting.value = true;
  submitStatus.value = 'idle';

  try {
    // Format donation data for submission
    const donationData = {
      type: 'donation',
      amount: parseFloat(donationForm.amount),
      frequency: donationForm.frequency,
      firstName: donationForm.firstName.trim(),
      lastName: donationForm.lastName.trim(),
      email: donationForm.email.trim(),
      phone: donationForm.phone.trim() || null,
      address: donationForm.address.trim(),
      city: donationForm.city.trim(),
      state: donationForm.state.trim().toUpperCase(),
      zipCode: donationForm.zipCode.trim(),
      dedication: donationForm.dedication.trim() || null,
      anonymousDonation: donationForm.anonymousDonation,
      receiveUpdates: donationForm.receiveUpdates,
      submittedAt: new Date().toISOString(),
    };

    // Submit using the contacts client (extend for donations)
    const response = await contactsClient.submitContact(donationData as any);
    
    if (response.success) {
      submitStatus.value = 'success';
      submitMessage.value = 'Thank you for your generous donation! We will contact you soon with payment processing instructions and receipt information.';
      
      // Reset form
      Object.keys(donationForm).forEach(key => {
        if (typeof donationForm[key as keyof typeof donationForm] === 'string') {
          (donationForm as any)[key] = '';
        } else if (typeof donationForm[key as keyof typeof donationForm] === 'boolean') {
          (donationForm as any)[key] = key === 'receiveUpdates';
        }
      });
      
      Object.keys(donationTouched).forEach(key => {
        donationTouched[key as keyof typeof donationTouched] = false;
      });
      
      Object.keys(donationErrors).forEach(key => {
        donationErrors[key as keyof typeof donationErrors] = null;
      });
    } else {
      submitStatus.value = 'error';
      submitMessage.value = response.message || 'Failed to submit donation information. Please try again.';
    }
  } catch (error) {
    console.error('Donation submission error:', error);
    submitStatus.value = 'error';
    submitMessage.value = 'Network error occurred. Please try again later.';
  } finally {
    isSubmitting.value = false;
  }
};
</script>