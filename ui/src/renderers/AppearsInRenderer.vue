<template>
  <VSelect
    v-if="control.visible"
    :label="control.label || 'Appears In'"
    :items="options"
    v-model="localValue"
    :disabled="!control.enabled"
    multiple
    chips
    clearable
    @update:modelValue="onChange"
  />
</template>

<script lang="ts" setup>
import { inject, computed, ref, watch } from 'vue';
import { useJsonFormsControl } from '@jsonforms/vue';
import type { JsonFormsSubStates, ControlElement } from '@jsonforms/core';
import cloneDeep from 'lodash/cloneDeep';
import { VSelect } from 'vuetify/components';

// Props from JSON Forms
const props = defineProps<{
  uischema: ControlElement;
  schema: any;
  path: string;
}>()

// Get array control state
const { control, handleChange } = useJsonFormsControl(props);

// Inject JSON Forms state to access root data
const jsonforms = inject<JsonFormsSubStates>('jsonforms');
const rootData = computed(() => jsonforms?.core?.data || {});

// Extract the index path to the current event
function extractEventLocation(path: string): [number, number, number] | null {
  const match = path.match(/weeks\.(\d+)\.days\.(\d+)\.events\.(\d+)/);
  if (!match) return null;
  return match.slice(1).map(Number) as [number, number, number];
}

// Dynamically resolve available columns
const options = computed(() => {
  const location = extractEventLocation(control.value.path);
  if (!location) return [];

  const [weekIndex, dayIndex] = location;
  const columns = rootData.value?.weeks?.[weekIndex]?.days?.[dayIndex]?.columns || [];

  return Array.isArray(columns)
    ? columns
    : [];
});

// Local copy of the selected values
const localValue = ref<string[]>(cloneDeep(control.value.data || []));

// Watch and sync local value
watch(
  () => control.value.data,
  (newVal) => {
    localValue.value = cloneDeep(newVal || []);
  },
  { immediate: true }
);

// Emit changes
function onChange(newVal: string[]) {
  handleChange(control.value.path, newVal);
}
</script>
