<template>
  <div class="rounded border bg-card p-6">
    <h3 class="text-lg font-semibold mb-4">Questions?</h3>
    
    <div class="space-y-3">
      <a
        href="tel:+1-561-555-0123"
        class="flex items-center gap-2 text-sm font-medium text-primary hover:text-primary/90 transition-colors"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect width="14" height="20" x="5" y="2" rx="2" ry="2"/>
          <path d="M12 18h.01"/>
        </svg>
        (561) 555-0123
      </a>
      
      <a
        href="mailto:info@internationalcenter.com"
        class="flex items-center gap-2 text-sm font-medium text-primary hover:text-primary/90 transition-colors"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect width="20" height="16" x="2" y="4" rx="2"/>
          <path d="m22 7-10 5L2 7"/>
        </svg>
        info@internationalcenter.com
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
// No props needed for this component
</script>