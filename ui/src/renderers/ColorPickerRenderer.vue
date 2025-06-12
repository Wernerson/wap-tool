<template>
  <div class="color-picker-renderer">
    <label v-if="control?.label" class="block mb-1">{{ control.label }}</label>
    <Sketch
      :model-value="color"
      @update:model-value="onColorChange"
      :disabled="!control.enabled"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Sketch } from '@ckpack/vue-color';
import { useJsonFormsControl } from '@jsonforms/vue';
import type { ControlProps } from '@jsonforms/core';

const props = defineProps<ControlProps>();
const { control, handleChange } = useJsonFormsControl(props);

const defaultColor = control.value.uischema.options?.defaultColor || '#ff0000';

// Local reactive color value
const color = ref(control.value.data || defaultColor);

// Emit change to JSONForms
function onColorChange(event: any) {
  const newColor = event.hex;
  color.value = newColor;
  if (control.value.path) {
    handleChange(control.value.path, newColor);
  }
}
</script>

<style scoped>
.color-picker-renderer {
  margin: 1em 0;
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: baseline;
  gap: 2em;
}
</style>
