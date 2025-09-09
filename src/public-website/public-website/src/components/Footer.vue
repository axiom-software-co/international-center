<template>
  <footer class="bg-black text-white">
    <div class="container">
      <!-- Desktop Layout -->
      <div class="hidden xl:block">
        <!-- Main Footer Content -->
        <div class="py-8 border-b border-gray-800">
          <div class="container">
            <!-- Desktop Layout -->
            <div class="flex items-start gap-8">
              <!-- Contact Information - Far Left -->
              <div class="flex-shrink-0 w-80 max-w-sm">
                <!-- Company Logo and Name -->
                <div class="mb-6">
                  <a href="/" class="flex items-center gap-3 hover:opacity-80 transition-opacity">
                    <svg
                      width="32"
                      height="32"
                      viewBox="0 0 286 326"
                      class="w-8 h-8 text-blue-600"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M266.006 -9.15527e-05C277.052 -9.15527e-05 286.006 8.95421 286.006 19.9999C286.006 63.3016 262.147 96.1501 235.27 120.882C208.842 145.199 175.696 164.877 154.286 179.512C133.026 194.044 104.6 210.659 80.1709 232.959C68.9387 243.212 59.5065 253.86 52.6094 265H233.995C241.496 277.403 245.806 290.607 245.997 305H40.0098C40.0049 305.333 40 305.666 40 306C40 317.046 31.0457 326 20 326C8.95431 326 0 317.046 0 306C3.61802e-05 262.353 25.8692 228.368 53.2041 203.416C80.3999 178.591 113.474 158.957 131.714 146.489C148.121 135.274 166.244 124.141 183.51 111.5H102.705C95.0362 105.785 87.4268 99.6644 80.1709 93.0409C76.0176 89.2496 72.1106 85.404 68.4863 81.4999H218.233C235.492 62.9951 246.006 43.0175 246.006 19.9999C246.006 8.95421 254.96 -9.15527e-05 266.006 -9.15527e-05Z"
                        fill="currentColor"
                        fill-opacity="0.75"
                      />
                      <path
                        d="M198.562 176.105C211.057 184.88 223.797 194.561 235.27 205.118C262.147 229.85 286.006 262.698 286.006 306C286.006 317.046 277.052 326 266.006 326C254.96 326 246.006 317.046 246.006 306C246.006 282.982 235.492 263.005 218.233 244.5H68.4863C72.1106 240.596 76.0176 236.75 80.1709 232.959C87.4272 226.335 95.0379 220.215 102.707 214.5H183.514C176.716 209.523 169.785 204.78 162.878 200.174L198.562 176.105ZM20 0C31.0457 0 40 8.95431 40 20C40 20.3338 40.0049 20.6671 40.0098 21H245.997C245.806 35.393 241.496 48.5969 233.995 61H52.6094C59.5065 72.1396 68.9387 82.7879 80.1709 93.041C100.482 111.581 123.555 126.194 142.976 138.945C139.155 141.468 135.391 143.976 131.714 146.489C125.025 151.062 116.341 156.598 106.683 163.001C90.0062 151.944 70.4265 138.305 53.2041 122.584C25.8692 97.6318 3.61802e-05 63.6471 0 20C0 8.95431 8.95431 0 20 0Z"
                        fill="currentColor"
                      />
                    </svg>
                    <div class="flex flex-col">
                      <h2 class="text-xl font-bold text-white leading-tight">International Center</h2>
                      <span class="text-xs font-medium text-gray-400 leading-tight tracking-wide">
                        for Regenerative Medicine
                      </span>
                    </div>
                  </a>
                </div>

                <div class="bg-gray-900/50 rounded p-6">
                  <table class="w-full text-sm">
                    <thead>
                      <tr>
                        <th colspan="2" class="font-semibold text-white pb-4 text-left text-lg">
                          Contact Information
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr class="border-b border-gray-800">
                        <td class="font-medium text-gray-300 py-3 pr-4">Phone</td>
                        <td class="py-3 text-right">
                          <a
                            :href="`tel:${contactInfo.mainPhone}`"
                            class="text-gray-300 hover:text-white transition-colors font-medium"
                          >
                            {{ contactInfo.mainPhone }}
                          </a>
                        </td>
                      </tr>
                      <tr class="border-b border-gray-800">
                        <td class="font-medium text-gray-300 py-3 pr-4">Email</td>
                        <td class="py-3 text-right">
                          <a
                            :href="`mailto:${contactInfo.email}`"
                            class="text-gray-300 hover:text-white transition-colors"
                          >
                            {{ contactInfo.email }}
                          </a>
                        </td>
                      </tr>
                      <tr class="last:border-b-0">
                        <td class="font-medium text-gray-300 pt-3 pr-4">Address</td>
                        <td class="pt-3 text-right">
                          <div class="text-gray-300 text-right">
                            <div>{{ contactInfo.address.street }}</div>
                            <div>
                              {{ contactInfo.address.city }}, {{ contactInfo.address.state }}
                              {{ contactInfo.address.zip }}
                            </div>
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- Services, Patient Resources, Community, and Company - Evenly spaced in remaining area -->
              <div class="flex-1 flex justify-evenly items-start min-w-0">
                <!-- Services with Dropdowns -->
                <div class="w-52 flex-shrink-0">
                  <h3 class="font-semibold text-white mb-6 text-lg">Services</h3>

                  <div class="space-y-4">
                    <!-- Service Categories with Dropdowns -->
                    <div class="space-y-3">
                      <!-- Loading state -->
                      <template v-if="isLoadingServices">
                        <div class="text-center py-4">
                          <p class="text-gray-400 text-sm">Loading services...</p>
                        </div>
                      </template>
                      <!-- Services content -->
                      <template v-else-if="serviceCategories.length > 0">
                        <div v-for="category in serviceCategories" :key="category.title">
                          <button
                            @click="toggleCategory(category.title)"
                            :class="`flex items-center justify-between w-full text-left hover:text-white transition-colors font-medium ${
                              expandedCategory === category.title ? 'text-white' : 'text-gray-400'
                            }`"
                          >
                            <span>{{ category.title }}</span>
                            <ChevronDown
                              :class="`w-4 h-4 transition-transform ${expandedCategory === category.title ? 'rotate-180' : ''}`"
                            />
                          </button>

                          <div v-if="expandedCategory === category.title" class="mt-2 space-y-3">
                            <a
                              v-for="service in category.services"
                              :key="service.name"
                              :href="service.href"
                              class="flex items-center text-gray-400 hover:text-white transition-colors font-normal"
                            >
                              <span class="text-blue-600 mr-2 text-sm font-bold">—</span>
                              {{ service.name }}
                            </a>
                          </div>
                        </div>
                      </template>
                      <!-- Error state -->
                      <div v-else class="text-center py-4">
                        <p class="text-gray-400 text-sm">Services temporarily unavailable</p>
                      </div>
                    </div>

                    <a
                      v-if="serviceCategories.length > 0"
                      href="/services"
                      class="block text-gray-300 hover:text-white transition-colors font-semibold mt-4"
                    >
                      View All Services
                    </a>
                  </div>
                </div>

                <!-- Patient Resources -->
                <div class="w-52 flex-shrink-0 ml-8">
                  <h3 class="font-semibold text-white mb-6 text-lg">Patient Resources</h3>
                  <ul class="space-y-3 mb-4">
                    <li v-for="(link, idx) in patientResourcesLinks" :key="idx">
                      <a :href="link.href" class="text-gray-400 hover:text-white transition-colors">
                        {{ link.name }}
                      </a>
                    </li>
                  </ul>

                  <a
                    href="/patient-resources/portal"
                    class="block text-gray-300 hover:text-white transition-colors font-semibold"
                  >
                    Patient Portal
                  </a>
                </div>

                <!-- Community -->
                <div class="w-48 flex-shrink-0">
                  <h3 class="font-semibold text-white mb-6 text-lg">Community</h3>
                  <ul class="space-y-3 mb-4">
                    <li v-for="(link, idx) in communityLinks" :key="idx">
                      <a :href="link.href" class="text-gray-400 hover:text-white transition-colors">
                        {{ link.name }}
                      </a>
                    </li>
                  </ul>

                  <a
                    href="/community/learning-portal"
                    class="block text-gray-300 hover:text-white transition-colors font-semibold"
                  >
                    Learning Portal
                  </a>
                </div>

                <!-- Company -->
                <div class="w-44 flex-shrink-0">
                  <h3 class="font-semibold text-white mb-6 text-lg">Company</h3>
                  <ul class="space-y-3 mb-4">
                    <li v-for="(link, idx) in companyLinks" :key="idx">
                      <a :href="link.href" class="text-gray-400 hover:text-white transition-colors">
                        {{ link.name }}
                      </a>
                    </li>
                  </ul>

                  <a
                    href="/company/contact"
                    class="block text-gray-300 hover:text-white transition-colors font-semibold"
                  >
                    Contact Us
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Mobile and Tablet Layout -->
      <div class="xl:hidden">
        <!-- Mobile First: Contact Information (Company Logo and Name) -->
        <div class="md:hidden py-8 border-b border-gray-800">
          <!-- Company Logo and Name -->
          <div class="mb-6">
            <a href="/" class="flex items-center gap-3 hover:opacity-80 transition-opacity">
              <svg
                width="32"
                height="32"
                viewBox="0 0 286 326"
                class="w-8 h-8 text-blue-600"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M266.006 -9.15527e-05C277.052 -9.15527e-05 286.006 8.95421 286.006 19.9999C286.006 63.3016 262.147 96.1501 235.27 120.882C208.842 145.199 175.696 164.877 154.286 179.512C133.026 194.044 104.6 210.659 80.1709 232.959C68.9387 243.212 59.5065 253.86 52.6094 265H233.995C241.496 277.403 245.806 290.607 245.997 305H40.0098C40.0049 305.333 40 305.666 40 306C40 317.046 31.0457 326 20 326C8.95431 326 0 317.046 0 306C3.61802e-05 262.353 25.8692 228.368 53.2041 203.416C80.3999 178.591 113.474 158.957 131.714 146.489C148.121 135.274 166.244 124.141 183.51 111.5H102.705C95.0362 105.785 87.4268 99.6644 80.1709 93.0409C76.0176 89.2496 72.1106 85.404 68.4863 81.4999H218.233C235.492 62.9951 246.006 43.0175 246.006 19.9999C246.006 8.95421 254.96 -9.15527e-05 266.006 -9.15527e-05Z"
                  fill="currentColor"
                  fill-opacity="0.75"
                />
                <path
                  d="M198.562 176.105C211.057 184.88 223.797 194.561 235.27 205.118C262.147 229.85 286.006 262.698 286.006 306C286.006 317.046 277.052 326 266.006 326C254.96 326 246.006 317.046 246.006 306C246.006 282.982 235.492 263.005 218.233 244.5H68.4863C72.1106 240.596 76.0176 236.75 80.1709 232.959C87.4272 226.335 95.0379 220.215 102.707 214.5H183.514C176.716 209.523 169.785 204.78 162.878 200.174L198.562 176.105ZM20 0C31.0457 0 40 8.95431 40 20C40 20.3338 40.0049 20.6671 40.0098 21H245.997C245.806 35.393 241.496 48.5969 233.995 61H52.6094C59.5065 72.1396 68.9387 82.7879 80.1709 93.041C100.482 111.581 123.555 126.194 142.976 138.945C139.155 141.468 135.391 143.976 131.714 146.489C125.025 151.062 116.341 156.598 106.683 163.001C90.0062 151.944 70.4265 138.305 53.2041 122.584C25.8692 97.6318 3.61802e-05 63.6471 0 20C0 8.95431 8.95431 0 20 0Z"
                  fill="currentColor"
                />
              </svg>
              <div class="flex flex-col">
                <h2 class="text-xl font-bold text-white leading-tight">International Center</h2>
                <span class="text-xs font-medium text-gray-400 leading-tight tracking-wide">
                  for Regenerative Medicine
                </span>
              </div>
            </a>
          </div>

          <div class="bg-gray-900/50 rounded p-6">
            <table class="w-full text-sm">
              <thead>
                <tr>
                  <th colspan="2" class="font-semibold text-white pb-4 text-left text-lg">
                    Contact Information
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr class="border-b border-gray-800">
                  <td class="font-medium text-gray-300 py-3 pr-4">Phone</td>
                  <td class="py-3 text-right">
                    <a
                      :href="`tel:${contactInfo.mainPhone}`"
                      class="text-gray-300 md:hover:text-white transition-colors font-medium"
                    >
                      {{ contactInfo.mainPhone }}
                    </a>
                  </td>
                </tr>
                <tr class="border-b border-gray-800">
                  <td class="font-medium text-gray-300 py-3 pr-4">Email</td>
                  <td class="py-3 text-right">
                    <a
                      :href="`mailto:${contactInfo.email}`"
                      class="text-gray-300 md:hover:text-white transition-colors"
                    >
                      {{ contactInfo.email }}
                    </a>
                  </td>
                </tr>
                <tr class="last:border-b-0">
                  <td class="font-medium text-gray-300 pt-3 pr-4">Address</td>
                  <td class="pt-3 text-right">
                    <div class="text-gray-300 text-right">
                      <div>{{ contactInfo.address.street }}</div>
                      <div>
                        {{ contactInfo.address.city }}, {{ contactInfo.address.state }}
                        {{ contactInfo.address.zip }}
                      </div>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Mobile: Services Section -->
        <div class="md:hidden py-8 border-b border-gray-800">
          <div>
            <h3 class="font-semibold text-white mb-6 text-lg">Services</h3>

            <div class="space-y-4">
              <!-- Service Categories with Dropdowns -->
              <div class="space-y-3">
                <!-- Loading state -->
                <template v-if="isLoadingServices">
                  <div class="text-center py-4">
                    <p class="text-gray-400 text-sm">Loading services...</p>
                  </div>
                </template>
                <!-- Services content -->
                <template v-else-if="serviceCategories.length > 0">
                  <div v-for="category in serviceCategories" :key="category.title">
                    <button
                      @click="toggleCategory(category.title)"
                      :class="`flex items-center justify-between w-full text-left py-2 md:hover:text-white transition-colors ${
                        expandedCategory === category.title ? 'text-white' : 'text-gray-400'
                      }`"
                    >
                      <span>{{ category.title }}</span>
                      <ChevronDown
                        :class="`w-4 h-4 transition-transform ${expandedCategory === category.title ? 'rotate-180' : ''}`"
                      />
                    </button>

                    <div v-if="expandedCategory === category.title" class="mt-2 space-y-2">
                      <a
                        v-for="service in category.services"
                        :key="service.name"
                        :href="service.href"
                        class="flex items-center py-2 text-gray-400 md:hover:text-white transition-colors"
                      >
                        <span class="text-blue-600 mr-2 text-sm font-bold">—</span>
                        {{ service.name }}
                      </a>
                    </div>
                  </div>
                </template>
                <!-- Error state -->
                <div v-else class="text-center py-4">
                  <p class="text-gray-400 text-sm">Services temporarily unavailable</p>
                </div>
              </div>

              <a
                v-if="serviceCategories.length > 0"
                href="/services"
                class="block text-gray-300 md:hover:text-white transition-colors font-semibold mt-4"
              >
                View All Services
              </a>
            </div>
          </div>
        </div>

        <!-- Tablet Grid Layout -->
        <div class="hidden md:block xl:hidden py-8">
          <!-- Top Row -->
          <div class="grid grid-cols-2 gap-8 mb-8">
            <!-- Contact Information - Top Left -->
            <div>
              <!-- Company Logo and Name -->
              <div class="mb-6">
                <a href="/" class="flex items-center gap-3 hover:opacity-80 transition-opacity">
                  <svg
                    width="32"
                    height="32"
                    viewBox="0 0 286 326"
                    class="w-8 h-8 text-blue-600"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      d="M266.006 -9.15527e-05C277.052 -9.15527e-05 286.006 8.95421 286.006 19.9999C286.006 63.3016 262.147 96.1501 235.27 120.882C208.842 145.199 175.696 164.877 154.286 179.512C133.026 194.044 104.6 210.659 80.1709 232.959C68.9387 243.212 59.5065 253.86 52.6094 265H233.995C241.496 277.403 245.806 290.607 245.997 305H40.0098C40.0049 305.333 40 305.666 40 306C40 317.046 31.0457 326 20 326C8.95431 326 0 317.046 0 306C3.61802e-05 262.353 25.8692 228.368 53.2041 203.416C80.3999 178.591 113.474 158.957 131.714 146.489C148.121 135.274 166.244 124.141 183.51 111.5H102.705C95.0362 105.785 87.4268 99.6644 80.1709 93.0409C76.0176 89.2496 72.1106 85.404 68.4863 81.4999H218.233C235.492 62.9951 246.006 43.0175 246.006 19.9999C246.006 8.95421 254.96 -9.15527e-05 266.006 -9.15527e-05Z"
                      fill="currentColor"
                      fill-opacity="0.75"
                    />
                    <path
                      d="M198.562 176.105C211.057 184.88 223.797 194.561 235.27 205.118C262.147 229.85 286.006 262.698 286.006 306C286.006 317.046 277.052 326 266.006 326C254.96 326 246.006 317.046 246.006 306C246.006 282.982 235.492 263.005 218.233 244.5H68.4863C72.1106 240.596 76.0176 236.75 80.1709 232.959C87.4272 226.335 95.0379 220.215 102.707 214.5H183.514C176.716 209.523 169.785 204.78 162.878 200.174L198.562 176.105ZM20 0C31.0457 0 40 8.95431 40 20C40 20.3338 40.0049 20.6671 40.0098 21H245.997C245.806 35.393 241.496 48.5969 233.995 61H52.6094C59.5065 72.1396 68.9387 82.7879 80.1709 93.041C100.482 111.581 123.555 126.194 142.976 138.945C139.155 141.468 135.391 143.976 131.714 146.489C125.025 151.062 116.341 156.598 106.683 163.001C90.0062 151.944 70.4265 138.305 53.2041 122.584C25.8692 97.6318 3.61802e-05 63.6471 0 20C0 8.95431 8.95431 0 20 0Z"
                      fill="currentColor"
                    />
                  </svg>
                  <div class="flex flex-col">
                    <h2 class="text-xl font-bold text-white leading-tight">International Center</h2>
                    <span class="text-xs font-medium text-gray-400 leading-tight tracking-wide">
                      for Regenerative Medicine
                    </span>
                  </div>
                </a>
              </div>

              <div class="bg-gray-900/50 rounded p-6">
                <table class="w-full text-sm">
                  <thead>
                    <tr>
                      <th colspan="2" class="font-semibold text-white pb-4 text-left text-lg">
                        Contact Information
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr class="border-b border-gray-800">
                      <td class="font-medium text-gray-300 py-3 pr-4">Phone</td>
                      <td class="py-3 text-right">
                        <a
                          :href="`tel:${contactInfo.mainPhone}`"
                          class="text-gray-300 md:hover:text-white transition-colors font-medium"
                        >
                          {{ contactInfo.mainPhone }}
                        </a>
                      </td>
                    </tr>
                    <tr class="border-b border-gray-800">
                      <td class="font-medium text-gray-300 py-3 pr-4">Email</td>
                      <td class="py-3 text-right">
                        <a
                          :href="`mailto:${contactInfo.email}`"
                          class="text-gray-300 md:hover:text-white transition-colors"
                        >
                          {{ contactInfo.email }}
                        </a>
                      </td>
                    </tr>
                    <tr class="last:border-b-0">
                      <td class="font-medium text-gray-300 pt-3 pr-4">Address</td>
                      <td class="pt-3 text-right">
                        <div class="text-gray-300 text-right">
                          <div>{{ contactInfo.address.street }}</div>
                          <div>
                            {{ contactInfo.address.city }}, {{ contactInfo.address.state }}
                            {{ contactInfo.address.zip }}
                          </div>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Services - Top Right -->
            <div>
              <h3 class="font-semibold text-white mb-6 text-lg">Services</h3>

              <div class="space-y-4">
                <!-- Service Categories with Dropdowns -->
                <div class="space-y-3">
                  <!-- Loading state -->
                  <template v-if="isLoadingServices">
                    <div class="text-center py-4">
                      <p class="text-gray-400 text-sm">Loading services...</p>
                    </div>
                  </template>
                  <!-- Services content -->
                  <template v-else-if="serviceCategories.length > 0">
                    <div v-for="category in serviceCategories" :key="category.title">
                      <button
                        @click="toggleCategory(category.title)"
                        :class="`flex items-center justify-between w-full text-left md:hover:text-white transition-colors font-medium ${
                          expandedCategory === category.title ? 'text-white' : 'text-gray-400'
                        }`"
                      >
                        <span>{{ category.title }}</span>
                        <ChevronDown
                          :class="`w-4 h-4 transition-transform ${expandedCategory === category.title ? 'rotate-180' : ''}`"
                        />
                      </button>

                      <div v-if="expandedCategory === category.title" class="mt-2 space-y-3">
                        <a
                          v-for="service in category.services"
                          :key="service.name"
                          :href="service.href"
                          class="flex items-center text-gray-400 md:hover:text-white transition-colors font-normal"
                        >
                          <span class="text-blue-600 mr-2 text-sm font-bold">—</span>
                          {{ service.name }}
                        </a>
                      </div>
                    </div>
                  </template>
                  <!-- Error state -->
                  <div v-else class="text-center py-4">
                    <p class="text-gray-400 text-sm">Services temporarily unavailable</p>
                  </div>
                </div>

                <a
                  v-if="serviceCategories.length > 0"
                  href="/services"
                  class="block text-gray-300 md:hover:text-white transition-colors font-semibold mt-4"
                >
                  View All Services
                </a>
              </div>
            </div>
          </div>

          <!-- Horizontal Divider -->
          <div class="border-t border-gray-800 mb-8"></div>

          <!-- Bottom Row -->
          <div class="flex justify-center gap-16 max-w-4xl mx-auto">
            <!-- Patient Resources - Bottom Left -->
            <div class="flex-1">
              <h3 class="font-semibold text-white mb-6 text-lg">Patient Resources</h3>

              <div>
                <ul class="space-y-3 mb-4">
                  <li v-for="(link, idx) in patientResourcesLinks" :key="idx">
                    <a :href="link.href" class="text-gray-400 md:hover:text-white transition-colors">
                      {{ link.name }}
                    </a>
                  </li>
                </ul>

                <a
                  href="/patient-resources/portal"
                  class="block text-gray-300 md:hover:text-white transition-colors font-semibold"
                >
                  Patient Portal
                </a>
              </div>
            </div>

            <!-- Community - Bottom Middle -->
            <div class="flex-1">
              <h3 class="font-semibold text-white mb-6 text-lg">Community</h3>

              <div>
                <ul class="space-y-3 mb-4">
                  <li v-for="(link, idx) in communityLinks" :key="idx">
                    <a :href="link.href" class="text-gray-400 md:hover:text-white transition-colors">
                      {{ link.name }}
                    </a>
                  </li>
                </ul>

                <a
                  href="/community/learning-portal"
                  class="block text-gray-300 md:hover:text-white transition-colors font-semibold"
                >
                  Learning Portal
                </a>
              </div>
            </div>

            <!-- Company - Bottom Right -->
            <div class="flex-1">
              <h3 class="font-semibold text-white mb-6 text-lg">Company</h3>

              <div>
                <ul class="space-y-3 mb-4">
                  <li v-for="(link, idx) in companyLinks" :key="idx">
                    <a :href="link.href" class="text-gray-400 md:hover:text-white transition-colors">
                      {{ link.name }}
                    </a>
                  </li>
                </ul>

                <a
                  href="/company/contact"
                  class="block text-gray-300 md:hover:text-white transition-colors font-semibold"
                >
                  Contact Us
                </a>
              </div>
            </div>
          </div>
        </div>

        <!-- Mobile: Patient Resources -->
        <div class="md:hidden py-8 border-b border-gray-800">
          <h3 class="font-semibold text-white mb-6 text-lg">Patient Resources</h3>

          <div>
            <ul class="space-y-3 mb-4">
              <li v-for="(link, idx) in patientResourcesLinks" :key="idx">
                <a :href="link.href" class="text-gray-400 md:hover:text-white transition-colors">
                  {{ link.name }}
                </a>
              </li>
            </ul>

            <a
              href="/patient-resources/portal"
              class="block text-gray-300 md:hover:text-white transition-colors font-semibold"
            >
              Patient Portal
            </a>
          </div>
        </div>

        <!-- Mobile: Community -->
        <div class="md:hidden py-8 border-b border-gray-800">
          <h3 class="font-semibold text-white mb-6 text-lg">Community</h3>

          <div>
            <ul class="space-y-3 mb-4">
              <li v-for="(link, idx) in communityLinks" :key="idx">
                <a :href="link.href" class="text-gray-400 md:hover:text-white transition-colors">
                  {{ link.name }}
                </a>
              </li>
            </ul>

            <a
              href="/community/learning-portal"
              class="block text-gray-300 md:hover:text-white transition-colors font-semibold"
            >
              Learning Portal
            </a>
          </div>
        </div>

        <!-- Mobile: Company -->
        <div class="md:hidden py-8 border-b border-gray-800">
          <h3 class="font-semibold text-white mb-6 text-lg">Company</h3>

          <div>
            <ul class="space-y-3 mb-4">
              <li v-for="(link, idx) in companyLinks" :key="idx">
                <a :href="link.href" class="text-gray-400 md:hover:text-white transition-colors">
                  {{ link.name }}
                </a>
              </li>
            </ul>

            <a
              href="/company/contact"
              class="block text-gray-300 md:hover:text-white transition-colors font-semibold"
            >
              Contact Us
            </a>
          </div>
        </div>

        <!-- Mobile: Quick Access -->
        <div class="md:hidden py-8">
          <h3 class="font-semibold text-white text-lg mb-4">Quick Access</h3>
          <div class="flex flex-col gap-3">
            <a
              href="/patient-resources/portal"
              class="flex items-center gap-2 px-4 py-3 text-sm font-medium text-gray-900 bg-white hover:bg-gray-100 rounded transition-colors border border-gray-300 justify-center"
            >
              <IdCard class="w-4 h-4" />
              Patient Portal
            </a>
            <a
              href="/appointment"
              class="flex items-center gap-2 px-4 py-3 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded transition-colors justify-center"
            >
              <Calendar class="w-4 h-4" />
              Book Appointment
            </a>
          </div>
        </div>
      </div>

      <!-- Footer Bottom -->
      <div class="py-8 border-t border-gray-800">
        <div class="flex flex-col md:flex-row justify-between items-center gap-6">
          <!-- Legal Links -->
          <div class="flex flex-row gap-4 sm:gap-6 order-2 md:order-1">
            <a
              href="/privacy-policy"
              class="flex items-center gap-2 text-xs text-gray-400 hover:text-gray-300 transition-colors"
            >
              <FileText class="w-4 h-4" />
              Privacy Policy
            </a>
            <a
              href="/terms-of-service"
              class="flex items-center gap-2 text-xs text-gray-400 hover:text-gray-300 transition-colors"
            >
              <FileText class="w-4 h-4" />
              Terms of Service
            </a>
          </div>

          <!-- Copyright -->
          <div class="text-center order-1 md:order-2">
            <p class="text-sm text-gray-300">© 2025 International Center. All rights reserved.</p>
          </div>

          <!-- Social Links -->
          <div class="flex items-center gap-4 order-3">
            <span class="text-xs text-gray-500 font-medium">Follow Us:</span>
            <a
              v-for="social in socialLinks"
              :key="social.label"
              :href="social.href"
              :aria-label="social.label"
              class="text-gray-500 hover:text-white transition-colors p-1"
            >
              <component :is="social.icon" class="w-5 h-5" />
            </a>
          </div>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { FileText, Linkedin, Facebook, Instagram, ChevronDown, IdCard, Calendar } from 'lucide-vue-next';

interface FooterServiceCategory {
  title: string;
  services: Array<{
    name: string;
    href: string;
  }>;
}

// State management for services
const serviceCategories = ref<FooterServiceCategory[]>([]);
const isLoadingServices = ref(false);
const expandedCategory = ref<string | null>(null);

const toggleCategory = (categoryTitle: string) => {
  expandedCategory.value = expandedCategory.value === categoryTitle ? null : categoryTitle;
};

// Load services using domain-specific client
async function loadFooterServices() {
  if (isLoadingServices.value) return;

  isLoadingServices.value = true;

  try {
    const { loadNavigationData } = await import('../lib/navigation-data');
    const navigationData = await loadNavigationData();
    
    // Transform to footer format
    serviceCategories.value = navigationData.footerCategories || [];
    
    console.log(`✅ Footer loaded ${serviceCategories.value.length} service categories`);
  } catch (error) {
    console.warn('❌ Failed to load footer services:', error);
    serviceCategories.value = [];
  } finally {
    isLoadingServices.value = false;
  }
}

// Load footer data on mount
onMounted(() => {
  loadFooterServices();
});

// Create X/Twitter icon component since it's not in lucide-vue-next
const XIcon = {
  template: `
    <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
      <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
    </svg>
  `,
};

const socialLinks = computed(() => [
  { icon: Linkedin, href: '#', label: 'LinkedIn' },
  { icon: Facebook, href: '#', label: 'Facebook' },
  { icon: Instagram, href: '#', label: 'Instagram' },
  { icon: XIcon, href: '#', label: 'X' },
]);

const communityLinks = [
  { name: 'About Our Values', href: '/community' },
  { name: 'Research & Innovation', href: '/community/research' },
  { name: 'Certification Programs', href: '/community/certification-programs' },
  { name: 'Donations', href: '/community/donations' },
  { name: 'Volunteer', href: '/community/volunteer' },
  { name: 'Events', href: '/community/events' },
];

const patientResourcesLinks = [
  { name: 'Pre-Treatment Preparation', href: '/patient-resources/pre-treatment' },
  { name: 'Post-Treatment Care', href: '/patient-resources/post-treatment' },
  { name: 'Patient Forms', href: '/patient-resources/forms' },
  { name: 'Support Groups', href: '/patient-resources/support-groups' },
  { name: 'Standard Charges', href: '/patient-resources/standard-charges' },
  { name: 'Insurance & Billing', href: '/patient-resources/insurance-billing' },
  { name: 'Financial Options', href: '/patient-resources/financial-options' },
];

const companyLinks = [
  { name: 'About Our Team', href: '/company/team' },
  { name: 'News & Insights', href: '/company/news' },
  { name: 'Credentials', href: '/company/credentials' },
  { name: 'Careers', href: '/company/careers' },
];

const contactInfo = {
  mainPhone: '(561) 555-0123',
  alternatePhone: '+1 (561) 555-5555',
  email: 'info@internationalcenter.com',
  address: {
    street: '1234 Medical Plaza Drive',
    city: 'Boynton Beach',
    state: 'FL',
    zip: '33426',
  },
};
</script>
