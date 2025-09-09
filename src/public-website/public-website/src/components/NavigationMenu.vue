<template>
  <div>
    <nav
      ref="navRef"
      :class="`fixed top-0 left-0 right-0 z-50 transition-all duration-300 border-b border-gray-200 bg-white dark:bg-gray-900 ${
        (isScrolled || activeDropdown) && !isMobileMenuOpen
          ? 'shadow-lg border-gray-200'
          : 'border-gray-200'
      }`"
    >
      <!-- Top Banner Section - Desktop Only -->
      <div class="hidden md:block border-b border-gray-200">
        <div class="container mx-auto px-4">
          <div class="flex items-center justify-between h-16">
            <!-- Left side - Emergency disclaimer -->
            <div class="flex items-center">
              <div class="text-xs text-gray-700 dark:text-gray-300 font-medium">
                Emergency? Call 911
              </div>
            </div>

            <!-- Right side - Actions -->
            <div class="flex items-center gap-3">
              <a
                href="tel:+1-561-555-0123"
                class="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors duration-200"
              >
                <Smartphone class="w-4 h-4 flex-shrink-0" />
                <span class="leading-none">(561) 555-0123</span>
              </a>
              <a
                href="/community/donations"
                class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-black bg-white hover:bg-gray-100 rounded transition-colors border border-gray-300"
              >
                <Heart class="w-4 h-4 flex-shrink-0" />
                <span class="leading-none">Donations</span>
              </a>
              <a
                href="/appointment"
                class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-black hover:bg-gray-800 rounded transition-colors"
              >
                <Calendar class="w-4 h-4 flex-shrink-0" />
                <span class="leading-none">Book Appointment</span>
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Navigation Section -->
      <div class="container mx-auto px-4 relative">
        <div class="flex items-center justify-between h-16">
          <!-- Logo -->
          <a
            :href="logo.url"
            class="flex items-center gap-2 lg:gap-3 text-gray-900 dark:text-gray-100 hover:opacity-80 transition-opacity"
          >
            <svg
              width="32"
              height="32"
              viewBox="0 0 286 326"
              class="w-8 h-8 lg:w-8 lg:h-8 text-blue-600 flex-shrink-0"
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
            <div class="flex flex-col min-w-0">
              <span class="font-bold text-lg leading-tight truncate">{{ logo.title }}</span>
              <span
                class="text-xs font-medium text-gray-600 dark:text-gray-400 leading-tight tracking-wide truncate"
              >
                for Regenerative Medicine
              </span>
            </div>
          </a>

          <!-- Desktop Navigation -->
          <div class="hidden md:flex items-center space-x-4">
            <template v-for="item in menu" :key="item.title">
              <div v-if="item.items" class="relative">
                <button
                  :class="`flex items-center gap-1 px-2 py-3 text-sm font-medium leading-none transition-colors duration-200 rounded ${
                    activeDropdown === item.title
                      ? 'text-gray-900 dark:text-gray-100 bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700'
                      : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-50 dark:hover:bg-gray-800/50'
                  }`"
                  @click.prevent.stop="handleDropdownClick(item.title)"
                >
                  {{ item.title }}
                  <div class="flex items-center justify-center ml-0.5">
                    <ChevronDown
                      :class="`w-4 h-4 transition-transform duration-200 flex-shrink-0 ${
                        activeDropdown === item.title
                          ? 'rotate-180 text-gray-900 dark:text-gray-100'
                          : ''
                      }`"
                    />
                  </div>
                </button>
              </div>
              <a
                v-else
                :key="item.title"
                :href="item.url"
                :class="`px-2 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-100/30 dark:hover:bg-gray-800/50 transition-colors duration-200 ${
                  item.title === 'Home' ? 'md:hidden lg:block' : ''
                }`"
              >
                {{ item.title }}
              </a>
            </template>
            <button
              @click="isSearchOpen = !isSearchOpen"
              class="flex items-center justify-center w-10 h-10 flex-shrink-0 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors duration-200 border border-gray-200 ml-2"
              aria-label="Search"
            >
              <Search class="w-5 h-5" />
            </button>
          </div>

          <!-- Mobile Controls -->
          <div class="md:hidden flex items-center space-x-2">
            <!-- Mobile Search Button -->
            <button
              @click="isSearchOpen = !isSearchOpen"
              class="flex items-center justify-center w-10 h-10 flex-shrink-0 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors duration-200 border border-gray-200"
              aria-label="Search"
            >
              <Search class="w-5 h-5" />
            </button>

            <!-- Mobile Menu Toggle -->
            <button
              class="flex items-center justify-center w-10 h-10 flex-shrink-0 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors duration-200 border border-gray-200"
              @click="isMobileMenuOpen = !isMobileMenuOpen"
              aria-label="Toggle navigation"
            >
              <X v-if="isMobileMenuOpen" class="w-6 h-6" />
              <Menu v-else class="w-6 h-6" />
            </button>
          </div>
        </div>
      </div>

      <!-- Desktop Navigation Dropdown - Full React Implementation -->
      <div
        v-if="activeDropdown && !isMobileMenuOpen"
        class="fixed inset-0 z-[45] hidden md:block"
        style="top: 128px"
      >
        <!-- Background overlay - click to close -->
        <div class="absolute inset-0 bg-black/20 backdrop-blur-sm" @click="activeDropdown = null" />

        <!-- Main dropdown container -->
        <div
          class="relative bg-white border border-gray-200 shadow-2xl w-full h-full overflow-hidden"
        >
          <div class="w-full mx-auto h-full" style="max-width: 1400px">
            <div class="md:flex h-full">
              <!-- Left Creative Section - Large screens only -->
              <div
                :class="`hidden lg:block lg:w-1/4 flex-shrink-0 h-full overflow-hidden relative ${
                  activeDropdown === 'Services'
                    ? 'bg-blue-600'
                    : activeDropdown === 'Patient Resources'
                      ? 'bg-green-600'
                      : activeDropdown === 'Community'
                        ? 'bg-amber-600'
                        : activeDropdown === 'Company'
                          ? 'bg-purple-600'
                          : 'bg-black'
                }`"
              >
                <!-- Services Professional Decoration -->
                <div v-if="activeDropdown === 'Services'" class="relative h-full">
                  <!-- Sparse Circle Stars - Services -->
                  <div class="absolute inset-0">
                    <div class="absolute inset-0">
                      <div
                        v-for="i in 35"
                        :key="`services-star-${i}`"
                        class="w-2 h-2 bg-white rounded-full absolute animate-pulse"
                        :style="{
                          left: `${Math.random() * 97}%`,
                          top: `${12 + Math.random() * 85}%`,
                          opacity: 0.08 + Math.random() * 0.3,
                          animationDelay: `${Math.random() * 8}s`,
                          animationDuration: `${8 + Math.random() * 17}s`,
                        }"
                      />
                    </div>
                  </div>
                  <!-- Brand Typography -->
                  <div class="absolute top-8 left-0 right-0 z-20 flex justify-center">
                    <div class="text-center">
                      <div class="text-xs font-semibold text-white tracking-[0.2em] mb-1">
                        INTERNATIONAL CENTER
                      </div>
                      <div class="text-[10px] font-medium text-gray-100 tracking-widest">
                        REGENERATIVE MEDICINE
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Patient Resources Professional Decoration -->
                <div v-if="activeDropdown === 'Patient Resources'" class="relative h-full">
                  <!-- Sparse Triangle Stars - Patient Resources -->
                  <div class="absolute inset-0">
                    <div class="absolute inset-0">
                      <div
                        v-for="i in 32"
                        :key="`patient-star-${i}`"
                        class="w-0 h-0 absolute animate-pulse"
                        :style="{
                          borderLeft: '6px solid transparent',
                          borderRight: '6px solid transparent',
                          borderBottom: '8px solid white',
                          left: `${Math.random() * 97}%`,
                          top: `${12 + Math.random() * 85}%`,
                          opacity: 0.08 + Math.random() * 0.3,
                          animationDelay: `${Math.random() * 8}s`,
                          animationDuration: `${8 + Math.random() * 17}s`,
                        }"
                      />
                    </div>
                  </div>
                  <!-- Brand Typography -->
                  <div class="absolute top-8 left-0 right-0 z-20 flex justify-center">
                    <div class="text-center">
                      <div class="text-xs font-semibold text-white tracking-[0.2em] mb-1">
                        INTERNATIONAL CENTER
                      </div>
                      <div class="text-[10px] font-medium text-gray-100 tracking-widest">
                        REGENERATIVE MEDICINE
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Community Professional Decoration -->
                <div v-if="activeDropdown === 'Community'" class="relative h-full">
                  <!-- Sparse Diamond Stars - Community -->
                  <div class="absolute inset-0">
                    <div class="absolute inset-0">
                      <div
                        v-for="i in 38"
                        :key="`community-star-${i}`"
                        class="w-2 h-2 bg-white transform rotate-45 absolute animate-pulse"
                        :style="{
                          left: `${Math.random() * 97}%`,
                          top: `${12 + Math.random() * 85}%`,
                          opacity: 0.08 + Math.random() * 0.3,
                          animationDelay: `${Math.random() * 8}s`,
                          animationDuration: `${8 + Math.random() * 17}s`,
                        }"
                      />
                    </div>
                  </div>
                  <!-- Brand Typography -->
                  <div class="absolute top-8 left-0 right-0 z-20 flex justify-center">
                    <div class="text-center">
                      <div class="text-xs font-semibold text-white tracking-[0.2em] mb-1">
                        INTERNATIONAL CENTER
                      </div>
                      <div class="text-[10px] font-medium text-gray-100 tracking-widest">
                        REGENERATIVE MEDICINE
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Company Professional Decoration -->
                <div v-if="activeDropdown === 'Company'" class="relative h-full">
                  <!-- Sparse Square Stars - Company -->
                  <div class="absolute inset-0">
                    <div class="absolute inset-0">
                      <div
                        v-for="i in 30"
                        :key="`company-star-${i}`"
                        class="w-2 h-2 bg-white absolute animate-pulse"
                        :style="{
                          left: `${Math.random() * 97}%`,
                          top: `${12 + Math.random() * 85}%`,
                          opacity: 0.08 + Math.random() * 0.3,
                          animationDelay: `${Math.random() * 8}s`,
                          animationDuration: `${8 + Math.random() * 17}s`,
                        }"
                      />
                    </div>
                  </div>
                  <!-- Brand Typography -->
                  <div class="absolute top-8 left-0 right-0 z-20 flex justify-center">
                    <div class="text-center">
                      <div class="text-xs font-semibold text-white tracking-[0.2em] mb-1">
                        INTERNATIONAL CENTER
                      </div>
                      <div class="text-[10px] font-medium text-gray-100 tracking-widest">
                        REGENERATIVE MEDICINE
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Right Main Dropdown - All screens -->
              <div class="flex-1 h-full bg-white">
                <div class="flex flex-col h-full">
                  <!-- Header with title and close button -->
                  <div
                    class="flex-shrink-0 pt-2 pb-4 pl-4 pr-4 sm:pt-4 sm:pb-6 sm:pl-6 sm:pr-6 lg:pt-6 lg:pb-8 lg:pl-8 lg:pr-8"
                  >
                    <div
                      class="flex items-start justify-between gap-4 pb-4 border-b border-gray-200"
                    >
                      <!-- Title -->
                      <div class="flex-1">
                        <h2 class="text-3xl font-bold text-black">{{ activeDropdown }}</h2>
                      </div>

                      <!-- Action buttons and close button -->
                      <div class="flex-shrink-0 flex items-center gap-3">
                        <!-- Services dropdown action button -->
                        <a
                          v-if="
                            activeDropdown === 'Services' && effectiveServiceCategories.length > 0
                          "
                          href="/services"
                          class="flex items-center gap-2 px-4 h-10 text-sm font-medium text-white bg-black hover:bg-gray-800 rounded transition-colors"
                          @click="handleNavigationClick"
                        >
                          <LayoutGrid class="w-4 h-4" />
                          View All Services
                        </a>

                        <!-- Patient Resources dropdown action button -->
                        <a
                          v-if="activeDropdown === 'Patient Resources'"
                          href="/patient-resources/portal"
                          class="flex items-center gap-2 px-4 h-10 text-sm font-medium text-white bg-black hover:bg-gray-800 rounded transition-colors"
                          @click="handleNavigationClick"
                        >
                          <IdCard class="w-4 h-4" />
                          Patient Portal
                        </a>

                        <!-- Community dropdown action button -->
                        <a
                          v-if="activeDropdown === 'Community'"
                          href="/community/learning-portal"
                          class="flex items-center gap-2 px-4 h-10 text-sm font-medium text-white bg-black hover:bg-gray-800 rounded transition-colors"
                          @click="handleNavigationClick"
                        >
                          <GraduationCap class="w-4 h-4" />
                          Learning Portal
                        </a>

                        <!-- Company dropdown action button -->
                        <a
                          v-if="activeDropdown === 'Company'"
                          href="/company/contact"
                          class="flex items-center gap-2 px-4 h-10 text-sm font-medium text-white bg-black hover:bg-gray-800 rounded transition-colors"
                          @click="handleNavigationClick"
                        >
                          <ClipboardList class="w-4 h-4" />
                          Contact Us
                        </a>

                        <!-- Close button -->
                        <button
                          @click="activeDropdown = null"
                          class="flex items-center justify-center w-10 h-10 text-gray-600 hover:text-gray-800 bg-white hover:bg-gray-50 rounded transition-colors duration-200 border border-gray-300"
                          aria-label="Close navigation"
                        >
                          <X class="w-6 h-6" />
                        </button>
                      </div>
                    </div>
                  </div>

                  <!-- Main Dropdown Content -->
                  <div class="flex overflow-hidden flex-1">
                    <div
                      class="w-full h-full overflow-y-auto pl-4 pr-4 sm:pl-6 sm:pr-6 lg:pl-8 lg:pr-8"
                    >
                      <!-- Services Multi-column layout -->
                      <div v-if="activeDropdown === 'Services'" class="w-full h-full">
                        <!-- Loading state -->
                        <div v-if="isLoadingServices" class="flex items-center justify-center h-64">
                          <div class="text-center">
                            <div
                              class="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"
                            ></div>
                            <h3 class="text-lg font-semibold text-gray-900 mb-2">
                              Loading Services
                            </h3>
                            <p class="text-gray-600">Please wait while we load our services...</p>
                          </div>
                        </div>
                        <!-- Services content -->
                        <div
                          v-else-if="effectiveServiceCategories.length > 0"
                          class="grid grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-6 pb-8"
                        >
                          <div
                            v-for="category in effectiveServiceCategories.slice(0, 6)"
                            :key="category.title"
                            class="flex flex-col space-y-2"
                          >
                            <!-- Category header -->
                            <div class="border-b border-gray-200 pb-3">
                              <h3 class="text-lg font-semibold text-black">{{ category.title }}</h3>
                            </div>
                            <!-- Category services -->
                            <div class="space-y-1">
                              <a
                                v-for="service in category.items"
                                :key="service.title"
                                :href="service.url"
                                class="block py-2 text-gray-700 hover:text-black hover:bg-blue-50 transition-colors border-l-4 border-l-transparent hover:border-l-blue-400"
                                @click="handleNavigationClick"
                              >
                                <div class="text-base font-medium pl-3">{{ service.title }}</div>
                              </a>
                            </div>
                          </div>
                        </div>
                        <!-- Error state -->
                        <div v-else class="flex items-center justify-center h-64">
                          <div class="text-center">
                            <h3 class="text-lg font-semibold text-gray-900 mb-2">
                              Services Temporarily Unavailable
                            </h3>
                            <p class="text-gray-600 mb-4">
                              We're unable to load service information at the moment.
                            </p>
                            <button
                              @click="loadServiceCategories"
                              class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded transition-colors"
                              :disabled="isLoadingServices"
                            >
                              {{ isLoadingServices ? 'Retrying...' : 'Retry' }}
                            </button>
                            <p v-if="apiError" class="text-xs text-red-600 mt-2">{{ apiError }}</p>
                          </div>
                        </div>
                      </div>

                      <!-- Patient Resources 3-column layout -->
                      <div v-else-if="activeDropdown === 'Patient Resources'" class="w-full h-full">
                        <div class="grid grid-cols-3 gap-x-8 gap-y-6 pb-8">
                          <!-- Treatment Care -->
                          <div class="flex flex-col space-y-2">
                            <div class="border-b border-gray-200 pb-3">
                              <h3 class="text-lg font-semibold text-black">Treatment Care</h3>
                            </div>
                            <div class="space-y-1">
                              <a
                                href="/patient-resources/pre-treatment"
                                class="block py-2 text-gray-700 hover:text-black hover:bg-green-50 transition-colors border-l-4 border-l-transparent hover:border-l-green-400"
                                @click="handleNavigationClick"
                              >
                                <div class="text-base font-medium pl-3">
                                  Pre-Treatment Preparation
                                </div>
                              </a>
                              <a
                                href="/patient-resources/post-treatment"
                                class="block py-2 text-gray-700 hover:text-black hover:bg-green-50 transition-colors border-l-4 border-l-transparent hover:border-l-green-400"
                                @click="handleNavigationClick"
                              >
                                <div class="text-base font-medium pl-3">Post-Treatment Care</div>
                              </a>
                            </div>
                          </div>
                          <!-- Patient Support -->
                          <div class="flex flex-col space-y-2">
                            <div class="border-b border-gray-200 pb-3">
                              <h3 class="text-lg font-semibold text-black">Patient Support</h3>
                            </div>
                            <div class="space-y-1">
                              <a
                                href="/patient-resources/forms"
                                class="block py-2 text-gray-700 hover:text-black hover:bg-green-50 transition-colors border-l-4 border-l-transparent hover:border-l-green-400"
                                @click="handleNavigationClick"
                              >
                                <div class="text-base font-medium pl-3">Patient Forms</div>
                              </a>
                              <a
                                href="/patient-resources/support-groups"
                                class="block py-2 text-gray-700 hover:text-black hover:bg-green-50 transition-colors border-l-4 border-l-transparent hover:border-l-green-400"
                                @click="handleNavigationClick"
                              >
                                <div class="text-base font-medium pl-3">Support Groups</div>
                              </a>
                            </div>
                          </div>
                          <!-- Financial Services -->
                          <div class="flex flex-col space-y-2">
                            <div class="border-b border-gray-200 pb-3">
                              <h3 class="text-lg font-semibold text-black">Financial Services</h3>
                            </div>
                            <div class="space-y-1">
                              <a
                                href="/patient-resources/standard-charges"
                                class="block py-2 text-gray-700 hover:text-black hover:bg-green-50 transition-colors border-l-4 border-l-transparent hover:border-l-green-400"
                                @click="handleNavigationClick"
                              >
                                <div class="text-base font-medium pl-3">Standard Charges</div>
                              </a>
                              <a
                                href="/patient-resources/insurance-billing"
                                class="block py-2 text-gray-700 hover:text-black hover:bg-green-50 transition-colors border-l-4 border-l-transparent hover:border-l-green-400"
                                @click="handleNavigationClick"
                              >
                                <div class="text-base font-medium pl-3">Insurance & Billing</div>
                              </a>
                              <a
                                href="/patient-resources/financial-options"
                                class="block py-2 text-gray-700 hover:text-black hover:bg-green-50 transition-colors border-l-4 border-l-transparent hover:border-l-green-400"
                                @click="handleNavigationClick"
                              >
                                <div class="text-base font-medium pl-3">Financial Options</div>
                              </a>
                            </div>
                          </div>
                        </div>
                      </div>

                      <!-- Community Layout -->
                      <div v-else-if="activeDropdown === 'Community'" class="w-full h-full">
                        <div class="flex flex-col lg:grid lg:grid-cols-2 lg:gap-8 h-full">
                          <!-- Menu items -->
                          <div class="space-y-1 mb-8 lg:mb-0 lg:order-2">
                            <a
                              v-for="subItem in communityMenuItems"
                              :key="subItem.title"
                              :href="subItem.url"
                              class="block py-2 text-gray-700 hover:text-black hover:bg-amber-50 transition-colors border-l-4 border-l-transparent hover:border-l-amber-400"
                              @click="handleNavigationClick"
                            >
                              <div class="text-base font-medium pl-3">{{ subItem.title }}</div>
                            </a>
                          </div>
                          <!-- Featured content placeholder -->
                          <div
                            class="space-y-6 pt-6 lg:pt-0 border-t lg:border-t-0 border-gray-200 lg:order-1"
                          >
                            <div class="p-4 sm:p-6 rounded border border-gray-200">
                              <div class="h-36 sm:h-44 flex flex-col justify-between">
                                <div class="flex-shrink-0">
                                  <span
                                    class="text-xs font-semibold text-amber-600 uppercase tracking-wide mb-3 block"
                                    >Featured Content</span
                                  >
                                  <h3
                                    class="text-lg font-bold text-gray-900 line-clamp-2 mt-1 mb-2 leading-tight"
                                  >
                                    Community Updates Coming Soon
                                  </h3>
                                </div>
                                <div class="flex-shrink-0 border-t border-gray-200 pt-3">
                                  <div class="space-y-2 text-xs">
                                    <div class="flex justify-between">
                                      <span class="font-medium text-gray-600">Status:</span>
                                      <span class="text-xs text-gray-700">In Development</span>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>

                      <!-- Company Layout -->
                      <div v-else-if="activeDropdown === 'Company'" class="w-full h-full">
                        <div class="flex flex-col lg:grid lg:grid-cols-2 lg:gap-8 h-full">
                          <!-- Menu items -->
                          <div class="space-y-1 mb-8 lg:mb-0 lg:order-2">
                            <a
                              v-for="subItem in companyMenuItems"
                              :key="subItem.title"
                              :href="subItem.url"
                              class="block py-2 text-gray-700 hover:text-black hover:bg-purple-50 transition-colors border-l-4 border-l-transparent hover:border-l-purple-400"
                              @click="handleNavigationClick"
                            >
                              <div class="text-base font-medium pl-3">{{ subItem.title }}</div>
                            </a>
                          </div>
                          <!-- Featured content placeholder -->
                          <div
                            class="space-y-6 pt-6 lg:pt-0 border-t lg:border-t-0 border-gray-200 lg:order-1"
                          >
                            <div class="p-4 sm:p-6 border border-gray-200 rounded">
                              <div class="h-36 sm:h-44 flex flex-col justify-between">
                                <div class="flex-shrink-0">
                                  <span
                                    class="text-xs font-semibold text-purple-600 uppercase tracking-wide mb-3 block"
                                    >Featured News</span
                                  >
                                  <h3
                                    class="text-lg font-bold text-gray-900 line-clamp-2 mt-1 mb-2 leading-tight"
                                  >
                                    Company News Coming Soon
                                  </h3>
                                </div>
                                <div class="flex-shrink-0 border-t border-gray-200 pt-3">
                                  <div class="space-y-2 text-xs">
                                    <div class="flex justify-between">
                                      <span class="font-medium text-gray-600">Status:</span>
                                      <span class="text-xs text-gray-700">In Development</span>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </nav>

    <!-- Search Overlay -->
    <div
      v-if="isSearchOpen"
      class="fixed inset-0 z-50 bg-black/50 flex items-start justify-center pt-20"
      @click.self="isSearchOpen = false"
    >
      <div
        class="w-full max-w-2xl mx-4 bg-white rounded-lg shadow-xl animate-in fade-in-0 slide-in-from-top-4 duration-200"
      >
        <div class="p-4">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-semibold text-gray-900">Search</h3>
            <button
              @click="isSearchOpen = false"
              class="flex items-center justify-center w-8 h-8 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors"
              aria-label="Close search"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="space-y-2">
            <input
              type="text"
              placeholder="Search treatments, articles, and case studies..."
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
              @keydown.escape="isSearchOpen = false"
            />
            <p class="text-sm text-gray-600">Search functionality coming soon...</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Mobile Navigation -->
    <div
      :class="`fixed inset-0 z-40 md:hidden ${
        isMobileMenuOpen ? 'opacity-100 visible' : 'opacity-0 invisible'
      }`"
    >
      <div class="absolute inset-0 bg-black/50" @click.self="isMobileMenuOpen = false" />
      <div
        :class="`fixed inset-0 bg-white ${
          isMobileMenuOpen ? 'translate-y-0 opacity-100' : '-translate-y-full opacity-0'
        }`"
        style="top: 64px"
        @click.stop
      >
        <div class="mobile-nav-container w-full h-full overflow-y-auto overflow-x-hidden">
          <div class="mobile-nav-top-section pt-4 px-4 pb-4 border-b border-gray-200 bg-gray-50">
            <div class="text-xs text-black text-center mb-2 font-medium">Emergency? Call 911</div>
            <div class="flex rounded border border-gray-200 mb-3 overflow-hidden">
              <a
                href="tel:+1-561-555-0123"
                class="flex-1 flex items-center justify-center gap-1.5 py-3 px-3 text-sm font-medium text-black bg-white hover:bg-gray-100 border-r border-gray-200"
                @click.stop
              >
                <Smartphone class="w-4 h-4 flex-shrink-0" />
                <span class="leading-none">(561) 555-0123</span>
              </a>
              <a
                href="/community/donations"
                class="flex-1 flex items-center justify-center gap-1.5 py-3 px-3 text-sm font-medium text-black bg-white hover:bg-gray-100"
                @click="handleNavigationClick"
              >
                <Heart class="w-4 h-4 flex-shrink-0" />
                <span class="leading-none">Donations</span>
              </a>
            </div>
            <a
              href="/appointment"
              class="flex items-center justify-center gap-1.5 w-full py-3 px-4 text-sm font-medium text-white bg-black hover:bg-gray-800 rounded transition-colors duration-200"
              @click="handleNavigationClick"
            >
              <Calendar class="w-4 h-4 flex-shrink-0" />
              <span class="leading-none">Book Appointment</span>
            </a>
          </div>

          <!-- Mobile Menu Items -->
          <template v-for="item in menu" :key="item.title">
            <div v-if="item.items">
              <div
                :class="`sticky top-0 z-10 flex items-center w-full p-4 text-left text-white text-lg font-medium ${
                  item.title === 'Company'
                    ? 'bg-purple-600'
                    : item.title === 'Community'
                      ? 'bg-amber-600'
                      : item.title === 'Patient Resources'
                        ? 'bg-green-600'
                        : 'bg-blue-600'
                }`"
              >
                {{ item.title }}
              </div>

              <div>
                <!-- Services Mobile Menu -->
                <template v-if="item.title === 'Services'">
                  <!-- Loading state for mobile -->
                  <div v-if="isLoadingServices" class="border-b border-gray-200">
                    <div class="px-4 py-6 text-center text-gray-600">
                      <div
                        class="w-6 h-6 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-2"
                      ></div>
                      <div class="text-base font-medium">Loading Services...</div>
                    </div>
                  </div>
                  <!-- Services content for mobile -->
                  <template v-else-if="effectiveServiceCategories.length > 0">
                    <div v-for="category in effectiveServiceCategories" :key="category.title">
                      <div class="border-b border-gray-200">
                        <div class="px-4 py-4 text-black text-base font-medium bg-gray-100">
                          {{ category.title }}
                        </div>
                      </div>
                      <div
                        v-for="service in category.items"
                        :key="service.title"
                        class="border-b border-gray-200"
                      >
                        <a
                          :href="service.url"
                          class="block px-4 py-4 text-black text-base"
                          @click="handleNavigationClick"
                        >
                          {{ service.title }}
                        </a>
                      </div>
                    </div>
                    <div class="border-b border-gray-200">
                      <a
                        href="/services"
                        class="block px-4 py-4 text-black text-base font-semibold"
                        @click="handleNavigationClick"
                      >
                        View All Services
                      </a>
                    </div>
                  </template>
                  <!-- Error state for mobile -->
                  <div v-else class="border-b border-gray-200">
                    <div class="px-4 py-6 text-center text-gray-600">
                      <div class="text-base font-medium mb-2">Services Temporarily Unavailable</div>
                      <div class="text-sm mb-3">Unable to load service information</div>
                      <button
                        @click="loadServiceCategories"
                        class="px-3 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded transition-colors"
                        :disabled="isLoadingServices"
                      >
                        {{ isLoadingServices ? 'Retrying...' : 'Retry' }}
                      </button>
                    </div>
                  </div>
                </template>

                <!-- Other menu items -->
                <template v-else>
                  <div
                    v-for="subItem in item.items"
                    :key="subItem.title"
                    class="border-b border-gray-200"
                  >
                    <a
                      :href="subItem.url"
                      class="block px-4 py-4 text-black text-base"
                      @click="handleNavigationClick"
                    >
                      {{ subItem.title }}
                    </a>
                  </div>
                  
                  <!-- Action items for each section -->
                  <div v-if="item.title === 'Patient Resources'" class="border-b border-gray-200">
                    <a
                      href="/patient-resources/portal"
                      class="block px-4 py-4 text-black text-base font-semibold"
                      @click="handleNavigationClick"
                    >
                      Patient Portal
                    </a>
                  </div>
                  
                  <div v-if="item.title === 'Community'" class="border-b border-gray-200">
                    <a
                      href="/community/learning-portal"
                      class="block px-4 py-4 text-black text-base font-semibold"
                      @click="handleNavigationClick"
                    >
                      Learning Portal
                    </a>
                  </div>
                  
                  <div v-if="item.title === 'Company'" class="border-b border-gray-200">
                    <a
                      href="/company/contact"
                      class="block px-4 py-4 text-black text-base font-semibold"
                      @click="handleNavigationClick"
                    >
                      Contact Us
                    </a>
                  </div>
                </template>
              </div>
            </div>
            <div v-else>
              <a
                :href="item.url"
                class="block p-4 text-black text-base font-semibold transition-colors w-full text-left border-b border-gray-200"
                @click="handleNavigationClick"
              >
                {{ item.title }}
              </a>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue';
import {
  Heart,
  FileText,
  Calendar,
  Menu,
  X,
  ChevronDown,
  Home,
  Users,
  Phone,
  IdCard,
  Search,
  Newspaper,
  ClipboardList,
  UserCheck,
  Smartphone,
  Stethoscope,
  LifeBuoy,
  CreditCard,
  Wallet,
  LayoutGrid,
  GraduationCap,
} from 'lucide-vue-next';
import { staticServiceCategories } from '../data/static-services';

// Props interface
interface LogoProps {
  url: string;
  src?: string;
  alt: string;
  title: string;
}

interface MenuItem {
  title: string;
  url: string;
  description?: string;
  icon?: string;
  items?: MenuItem[];
  image?: string;
}

interface ServiceCategory {
  title: string;
  description: string;
  items: MenuItem[];
}

interface Props {
  logo?: LogoProps;
  menu?: MenuItem[];
  serviceCategories?: ServiceCategory[];
}

// Props with defaults
const props = withDefaults(defineProps<Props>(), {
  logo: () => ({
    url: '/',
    alt: 'International Center Logo',
    title: 'International Center',
  }),
  serviceCategories: () => staticServiceCategories,
  menu: () => [],
});

// API State Management
const apiServiceCategories = ref<ServiceCategory[]>([]);
const isLoadingServices = ref(false);
const apiError = ref<string | null>(null);
const useStaticData = ref(false);

// Reactive state
const isMobileMenuOpen = ref(false);
const isSearchOpen = ref(false);
const isScrolled = ref(false);
const activeDropdown = ref<string | null>(null);
const activeSubmenu = ref<string | null>(null);
const isTouchDevice = ref(false);
const isComponentMounted = ref(false);
const savedScrollPosition = ref(0);
const isNavigating = ref(false);

// Refs
const navRef = ref<HTMLElement>();
const timeoutRef = ref<NodeJS.Timeout | null>(null);

// Computed property for effective service categories (API first, fallback to static)
const effectiveServiceCategories = computed(() => {
  if (useStaticData.value || apiError.value) {
    return staticServiceCategories;
  }
  return apiServiceCategories.value;
});

// Load navigation data using domain-specific service
async function loadServiceCategories() {
  if (isLoadingServices.value) return;

  isLoadingServices.value = true;
  apiError.value = null;

  try {
    const { loadNavigationData } = await import('../lib/navigation-data');
    const navigationData = await loadNavigationData();
    
    // Transform to expected format for navigation menu
    apiServiceCategories.value = navigationData.navigationCategories || [];
    useStaticData.value = false;
    
    console.log(`✅ Navigation menu loaded ${apiServiceCategories.value.length} service categories`);
  } catch (error) {
    console.warn('❌ Navigation API failed, falling back to static data:', error);
    apiError.value = error instanceof Error ? error.message : 'Unknown error';
    useStaticData.value = true;
  } finally {
    isLoadingServices.value = false;
  }
}

// Create dynamic menu with services from API or static data
const menu = computed(() => {
  if (props.menu && props.menu.length > 0) return props.menu;

  const serviceCategories = effectiveServiceCategories.value;

  return [
    { title: 'Home', url: '/' },
    {
      title: 'Services',
      url: '/services',
      items:
        serviceCategories.length > 0
          ? serviceCategories.flatMap(category =>
              category.items.map(service => ({
                title: service.title,
                description: service.description || `${service.title} treatment`,
                url: service.url,
              }))
            )
          : [], // Empty services menu if no data available
    },
    {
      title: 'Patient Resources',
      url: '/patient-resources',
      items: [
        {
          title: 'Pre-Treatment Preparation',
          description: 'Guidelines and instructions to prepare for your treatments',
          url: '/patient-resources/pre-treatment',
        },
        {
          title: 'Post-Treatment Care',
          description: 'Recovery protocols and aftercare instructions',
          url: '/patient-resources/post-treatment',
        },
        {
          title: 'Patient Forms',
          description: 'Download and complete necessary medical forms',
          url: '/patient-resources/forms',
        },
        {
          title: 'Support Groups',
          description: 'Connect with other patients and support resources',
          url: '/patient-resources/support-groups',
        },
        {
          title: 'Standard Charges',
          description: 'Transparent pricing information for all services and procedures',
          url: '/patient-resources/standard-charges',
        },
        {
          title: 'Insurance & Billing',
          description: 'Insurance coverage information and billing assistance',
          url: '/patient-resources/insurance-billing',
        },
        {
          title: 'Financial Options',
          description: 'Payment plans and financing options for treatments',
          url: '/patient-resources/financial-options',
        },
      ],
    },
    {
      title: 'Community',
      url: '/community',
      items: [
        {
          title: 'About Our Values',
          description: 'Our commitment to patient care and medical excellence',
          url: '/community',
        },
        {
          title: 'Research & Innovation',
          description: 'Cutting-edge research and innovative treatments',
          url: '/community/research',
        },
        {
          title: 'Certification Programs',
          description: 'Professional training through in-person and online learning',
          url: '/community/certification-programs',
        },
        {
          title: 'Donations',
          description: 'Support our mission of advancing regenerative medicine',
          url: '/community/donations',
        },
        {
          title: 'Volunteer',
          description: 'Community outreach and volunteer opportunities',
          url: '/community/volunteer',
        },
        {
          title: 'Events',
          description: 'Upcoming seminars, workshops, and community events',
          url: '/community/events',
        },
      ],
    },
    {
      title: 'Company',
      url: '/company',
      items: [
        {
          title: 'About Our Team',
          description: 'Meet our board-certified physicians and medical team',
          url: '/company/team',
        },
        {
          title: 'News & Insights',
          description: 'Company news and medical breakthroughs',
          url: '/company/news',
        },
        {
          title: 'Credentials',
          description: 'Certifications, accreditations, and professional memberships',
          url: '/company/credentials',
        },
        {
          title: 'Careers',
          description: 'Join our team of healthcare professionals',
          url: '/company/careers',
        },
      ],
    },
  ];
});

// Community menu items
const communityMenuItems = computed(() => [
  {
    title: 'About Our Values',
    description: 'Our commitment to patient care and medical excellence',
    url: '/community',
  },
  {
    title: 'Research & Innovation',
    description: 'Cutting-edge research and innovative treatments',
    url: '/community/research',
  },
  {
    title: 'Certification Programs',
    description: 'Professional training through in-person and online learning',
    url: '/community/certification-programs',
  },
  {
    title: 'Donations',
    description: 'Support our mission of advancing regenerative medicine',
    url: '/community/donations',
  },
  {
    title: 'Volunteer',
    description: 'Community outreach and volunteer opportunities',
    url: '/community/volunteer',
  },
  {
    title: 'Events',
    description: 'Upcoming seminars, workshops, and community events',
    url: '/community/events',
  },
]);

// Company menu items
const companyMenuItems = computed(() => [
  {
    title: 'About Our Team',
    description: 'Meet our board-certified physicians and medical team',
    url: '/company/team',
  },
  {
    title: 'News & Insights',
    description: 'Company news and medical breakthroughs',
    url: '/company/news',
  },
  {
    title: 'Credentials',
    description: 'Certifications, accreditations, and professional memberships',
    url: '/company/credentials',
  },
  {
    title: 'Careers',
    description: 'Join our team of healthcare professionals',
    url: '/company/careers',
  },
]);

// Event handlers
const handleDropdownClick = (title: string) => {
  // Clear any existing timeouts
  if (timeoutRef.value) {
    clearTimeout(timeoutRef.value);
    timeoutRef.value = null;
  }

  // Toggle dropdown - close if same, open if different
  const isOpening = activeDropdown.value !== title;
  const isSwitching = activeDropdown.value && activeDropdown.value !== title;

  // Reset submenu when switching between dropdowns or closing
  if (isSwitching || !isOpening) {
    activeSubmenu.value = null;
  }

  activeDropdown.value = activeDropdown.value === title ? null : title;
};

const handleDropdownLeave = () => {
  // Only handle mouse events on non-touch devices
  if (isTouchDevice.value) return;

  timeoutRef.value = setTimeout(() => {
    activeDropdown.value = null;
    activeSubmenu.value = null;
  }, 500);
};

const handleNavigationClick = () => {
  isNavigating.value = true;
  isMobileMenuOpen.value = false;
  activeDropdown.value = null;
  activeSubmenu.value = null;

  // Reset navigation flag after a short delay
  setTimeout(() => {
    isNavigating.value = false;
  }, 100);
};

const handleClickOutside = (event: MouseEvent | TouchEvent) => {
  // Don't close dropdowns if mobile menu is open or component not ready
  if (isMobileMenuOpen.value || !isComponentMounted.value) return;

  if (navRef.value && !navRef.value.contains(event.target as Node)) {
    activeDropdown.value = null;
    activeSubmenu.value = null;
  }
};

const handleScroll = () => {
  if (typeof window !== 'undefined') {
    isScrolled.value = window.scrollY > 10;
  }
};

const handleEscapeKey = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    if (isSearchOpen.value) {
      isSearchOpen.value = false;
    } else if (activeDropdown.value) {
      activeDropdown.value = null;
      activeSubmenu.value = null;
    }
  }
};

const checkTouchDevice = () => {
  if (typeof window !== 'undefined') {
    isTouchDevice.value = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
  }
};

// Lifecycle
onMounted(() => {
  isComponentMounted.value = true;

  // Load service categories from API
  loadServiceCategories();

  // Add event listeners
  if (typeof window !== 'undefined') {
    window.addEventListener('scroll', handleScroll);
    window.addEventListener('resize', checkTouchDevice);
  }

  if (typeof document !== 'undefined') {
    document.addEventListener('mousedown', handleClickOutside, { passive: true });
    document.addEventListener('touchstart', handleClickOutside, { passive: true });
    document.addEventListener('keydown', handleEscapeKey);
  }

  checkTouchDevice();
});

onUnmounted(() => {
  // Cleanup event listeners
  if (typeof window !== 'undefined') {
    window.removeEventListener('scroll', handleScroll);
    window.removeEventListener('resize', checkTouchDevice);
  }

  if (typeof document !== 'undefined') {
    document.removeEventListener('mousedown', handleClickOutside);
    document.removeEventListener('touchstart', handleClickOutside);
    document.removeEventListener('keydown', handleEscapeKey);
  }

  if (timeoutRef.value) {
    clearTimeout(timeoutRef.value);
  }

  // Reset body styles
  document.body.style.overflow = '';
  document.body.style.position = '';
  document.body.style.width = '';
  document.body.style.top = '';
});

// Watch for mobile menu changes to handle body scroll lock
watch(isMobileMenuOpen, newValue => {
  if (!isComponentMounted.value) return;

  if (newValue) {
    // Capture scroll position
    const scrollY = typeof window !== 'undefined' ? window.scrollY || window.pageYOffset || 0 : 0;
    savedScrollPosition.value = scrollY;

    // Store original styles
    const originalPosition = document.body.style.position;
    const originalTop = document.body.style.top;
    const originalOverflow = document.body.style.overflow;
    const originalWidth = document.body.style.width;

    document.body.dataset.originalPosition = originalPosition;
    document.body.dataset.originalTop = originalTop;
    document.body.dataset.originalOverflow = originalOverflow;
    document.body.dataset.originalWidth = originalWidth;

    // Apply scroll lock
    document.body.style.overflow = 'hidden';
    document.body.style.position = 'fixed';
    document.body.style.width = '100%';
    document.body.style.top = `-${scrollY}px`;
  } else {
    // Restore original styles
    const originalPosition = document.body.dataset.originalPosition || '';
    const originalTop = document.body.dataset.originalTop || '';
    const originalOverflow = document.body.dataset.originalOverflow || '';
    const originalWidth = document.body.dataset.originalWidth || '';

    document.body.style.position = originalPosition;
    document.body.style.top = originalTop;
    document.body.style.overflow = originalOverflow;
    document.body.style.width = originalWidth;

    // Clean up data attributes
    delete document.body.dataset.originalPosition;
    delete document.body.dataset.originalTop;
    delete document.body.dataset.originalOverflow;
    delete document.body.dataset.originalWidth;

    // Only restore scroll position if we're not navigating to a new page
    if (!isNavigating.value && typeof window !== 'undefined') {
      window.scrollTo(0, savedScrollPosition.value);
    }
  }
});

// Close dropdowns when mobile menu opens/closes
watch(isMobileMenuOpen, newValue => {
  if (newValue) {
    activeDropdown.value = null;
  }
});
</script>

<style scoped>
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

.mobile-nav-container {
  transition:
    transform 0.3s ease-in-out,
    opacity 0.3s ease-in-out;
}

.animate-in {
  animation: slide-in-from-top 0.2s ease-out;
}

@keyframes slide-in-from-top {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.fade-in-0 {
  animation: fade-in 0.2s ease-out;
}

@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
