<template>
  <div class="grid gap-8 lg:grid-cols-2 lg:items-start">
    <!-- Quick Donation Form (Stripe) -->
    <QuickDonationForm />

    <!-- Large Donation Contact Form -->
    <LargeDonationForm />
  </div>
</template>

<script setup lang="ts">
import QuickDonationForm from './QuickDonationForm.vue';
import LargeDonationForm from './LargeDonationForm.vue';
</script>