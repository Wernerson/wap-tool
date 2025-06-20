<template>
  <VSelect
    :items="options"
    item-value="identifier"
    item-title="identifier"
    v-model="modelValue"
    :label="control.label"
    clearable
  >
    <template v-slot:item="{ props: itemProps, item }">
      <VListItem 
      v-bind="itemProps"
      :base-color="item.raw.textColor || '#000000'"
      :style="{backgroundColor: item.raw.color || '#f0f0f0'}"
      />
    </template>
  </VSelect>
</template>

<script lang="ts" setup>
import type { ControlProps, JsonFormsSubStates } from '@jsonforms/core';
import { useJsonFormsControl } from '@jsonforms/vue';
import { computed, inject } from 'vue';
import { VListItem, VSelect } from 'vuetify/components';

const props = defineProps<ControlProps>();
const { control, handleChange } = useJsonFormsControl(props);

const jsonforms = inject<JsonFormsSubStates>('jsonforms');
if (!jsonforms?.core) throw new Error("Missing 'jsonforms.core'");

const rootData = computed(() => jsonforms?.core?.data || {});

// Complex path resolution
function resolvePath(obj: any, path: string): any {
  return path
    .replace(/\[(\w+)\]/g, '.$1') // convert indexes to properties
    .split('.')
    .filter(Boolean)
    .reduce((o, p) => (o ? o[p] : undefined), obj);
}

const sourcePath = control.value.uischema?.options?.source ?? 'items';
const options = computed(() => {
  const resolved = resolvePath(rootData.value, sourcePath);
  return Array.isArray(resolved) ? resolved : [];
});

const modelValue = computed({
  get: () => control.value.data,
  set: (val) => {
    if (val === null) {
      // Remove the key entirely
      handleChange(control.value.path, undefined);
    } else {
      handleChange(control.value.path, val);
    }
  },
});
</script>
