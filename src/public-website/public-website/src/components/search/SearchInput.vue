<template>
  <div class="relative">
    <!-- Search Input -->
    <div class="relative">
      <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <SearchIcon class="h-5 w-5 text-gray-400" />
      </div>
      <input
        ref="inputRef"
        v-model="searchQuery"
        type="search"
        :placeholder="placeholder"
        :class="[
          'block w-full pl-10 pr-12 py-2 border border-gray-300 rounded-lg',
          'focus:ring-2 focus:ring-blue-500 focus:border-transparent',
          'dark:bg-gray-800 dark:border-gray-600 dark:text-white',
          'dark:placeholder-gray-400 dark:focus:ring-blue-400',
          inputClass
        ]"
        :disabled="loading"
        @keyup.enter="handleSearch"
        @input="handleInput"
        @focus="showSuggestions = true"
        @blur="handleBlur"
      />
      
      <!-- Clear button -->
      <div v-if="searchQuery" class="absolute inset-y-0 right-0 flex items-center">
        <button
          type="button"
          class="h-full px-3 text-gray-400 hover:text-gray-600 focus:outline-none"
          @click="clearSearch"
        >
          <XMarkIcon class="h-4 w-4" />
        </button>
      </div>
      
      <!-- Loading indicator -->
      <div v-if="loading" class="absolute inset-y-0 right-0 flex items-center pr-3">
        <div class="animate-spin rounded-full h-4 w-4 border-2 border-blue-500 border-t-transparent"></div>
      </div>
    </div>

    <!-- Suggestions Dropdown -->
    <div
      v-if="showSuggestions && (suggestions.length > 0 || searchHistory.length > 0)"
      class="absolute z-50 w-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg max-h-64 overflow-y-auto"
    >
      <!-- Search Suggestions -->
      <div v-if="suggestions.length > 0" class="py-1">
        <div class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
          Suggestions
        </div>
        <button
          v-for="suggestion in suggestions"
          :key="`suggestion-${suggestion}`"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 focus:bg-gray-100 dark:focus:bg-gray-700 focus:outline-none"
          @click="selectSuggestion(suggestion)"
        >
          <div class="flex items-center">
            <SearchIcon class="h-4 w-4 text-gray-400 mr-2" />
            <span class="text-sm text-gray-900 dark:text-white">{{ suggestion }}</span>
          </div>
        </button>
      </div>

      <!-- Divider -->
      <div v-if="suggestions.length > 0 && searchHistory.length > 0" class="border-t border-gray-200 dark:border-gray-600"></div>

      <!-- Search History -->
      <div v-if="searchHistory.length > 0" class="py-1">
        <div class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
          Recent Searches
        </div>
        <button
          v-for="historyItem in searchHistory.slice(0, 5)"
          :key="`history-${historyItem}`"
          type="button"
          class="w-full text-left px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 focus:bg-gray-100 dark:focus:bg-gray-700 focus:outline-none"
          @click="selectSuggestion(historyItem)"
        >
          <div class="flex items-center">
            <ClockIcon class="h-4 w-4 text-gray-400 mr-2" />
            <span class="text-sm text-gray-900 dark:text-white">{{ historyItem }}</span>
          </div>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue';
import { SearchIcon, XMarkIcon, ClockIcon } from 'lucide-vue-next';
import { useQuickSearch, useSearchHistory } from '../../composables/search/useUnifiedSearch';

interface Props {
  modelValue?: string;
  placeholder?: string;
  inputClass?: string;
  autoFocus?: boolean;
  showSuggestions?: boolean;
  debounceMs?: number;
}

interface Emits {
  (e: 'update:modelValue', value: string): void;
  (e: 'search', query: string): void;
  (e: 'clear'): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: 'Search...',
  inputClass: '',
  autoFocus: false,
  showSuggestions: true,
  debounceMs: 300,
});

const emit = defineEmits<Emits>();

// Template refs
const inputRef = ref<HTMLInputElement>();

// Local state
const searchQuery = ref(props.modelValue);
const showSuggestions = ref(false);
const suggestionTimeout = ref<NodeJS.Timeout>();

// Search functionality
const { loading, suggestions, getSuggestions } = useQuickSearch();
const { history: searchHistory } = useSearchHistory();

// Watch for external model changes
watch(() => props.modelValue, (newValue) => {
  searchQuery.value = newValue;
});

// Watch for search query changes
watch(searchQuery, (newQuery) => {
  emit('update:modelValue', newQuery);
  
  // Debounce suggestions
  if (suggestionTimeout.value) {
    clearTimeout(suggestionTimeout.value);
  }
  
  if (newQuery.trim() && props.showSuggestions) {
    suggestionTimeout.value = setTimeout(() => {
      getSuggestions(newQuery.trim());
    }, props.debounceMs);
  }
});

// Event handlers
const handleSearch = (): void => {
  if (searchQuery.value.trim()) {
    emit('search', searchQuery.value.trim());
    showSuggestions.value = false;
  }
};

const handleInput = (): void => {
  if (props.showSuggestions) {
    showSuggestions.value = true;
  }
};

const handleBlur = (): void => {
  // Delay hiding suggestions to allow for clicks
  setTimeout(() => {
    showSuggestions.value = false;
  }, 200);
};

const selectSuggestion = (suggestion: string): void => {
  searchQuery.value = suggestion;
  showSuggestions.value = false;
  nextTick(() => {
    emit('search', suggestion);
  });
};

const clearSearch = (): void => {
  searchQuery.value = '';
  showSuggestions.value = false;
  emit('clear');
  
  // Focus input after clearing
  nextTick(() => {
    inputRef.value?.focus();
  });
};

// Focus input on mount if autoFocus is true
onMounted(() => {
  if (props.autoFocus) {
    nextTick(() => {
      inputRef.value?.focus();
    });
  }
});
</script>