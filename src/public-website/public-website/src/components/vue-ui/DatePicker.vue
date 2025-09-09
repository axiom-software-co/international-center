<template>
  <div class="relative">
    <Popover v-model:open="open">
      <PopoverTrigger>
        <Button
          :id="id"
          variant="outline"
          type="button"
          :class="
            cn(
              'w-full justify-start text-left font-normal',
              !selectedDate && 'text-muted-foreground',
              props.class
            )
          "
        >
          <CalendarIcon class="mr-2 h-4 w-4" />
          {{ selectedDate ? formatDate(selectedDate) : placeholder }}
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-auto p-0" align="start">
        <Calendar
          :selected="selectedDate"
          :disabled="isDateDisabled"
          :min-date="minDateObj"
          :max-date="maxDateObj"
          @select="handleSelect"
        />
      </PopoverContent>
    </Popover>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import Button from './Button.vue';
import Calendar from './Calendar.vue';
import Popover from './Popover.vue';
import PopoverTrigger from './PopoverTrigger.vue';
import PopoverContent from './PopoverContent.vue';
import { Calendar as CalendarIcon } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

export interface DatePickerProps {
  modelValue?: string;
  placeholder?: string;
  class?: string;
  id?: string;
  minDate?: string;
  maxDate?: string;
}

const props = withDefaults(defineProps<DatePickerProps>(), {
  placeholder: 'Select date',
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();
const open = ref(false);

// Helper functions
const formatDate = (date: Date): string => {
  return date.toLocaleDateString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

const parseDate = (dateString: string): Date | undefined => {
  if (!dateString) return undefined;
  const date = new Date(dateString + 'T00:00:00');
  return isNaN(date.getTime()) ? undefined : date;
};

const formatDateString = (date: Date): string => {
  return date.toISOString().split('T')[0];
};

// Date constraints
const today = new Date();
today.setHours(0, 0, 0, 0);
const currentYear = new Date().getFullYear();
const maxYear = currentYear + 5;
const defaultMaxDate = new Date(maxYear, 11, 31);

const selectedDate = computed(() => parseDate(props.modelValue || ''));
const minDateObj = computed(() => (props.minDate ? parseDate(props.minDate) : today));
const maxDateObj = computed(() => (props.maxDate ? parseDate(props.maxDate) : defaultMaxDate));

const isDateDisabled = (date: Date) => {
  if (minDateObj.value && date < minDateObj.value) return true;
  if (maxDateObj.value && date > maxDateObj.value) return true;
  return false;
};

const handleSelect = (date: Date | undefined) => {
  if (date) {
    const dateString = formatDateString(date);
    emit('update:modelValue', dateString);
    open.value = false;
  }
};
</script>
