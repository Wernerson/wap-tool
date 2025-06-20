<template>
  <div>
    <v-tooltip v-if="!isValid">
      <template #activator="{ props }">
        <span v-bind="props" style="display: inline-block;">
          <v-btn :disabled="true" :prepend-icon="icon">
            {{ buttonText }}
          </v-btn>
        </span>
      </template>
      <span>
        {{ tooltipText }}
      </span>
    </v-tooltip>
    <v-btn
      v-else
      @click="onClick"
      :prepend-icon="icon"
    >
      {{ buttonText }}
    </v-btn>
  </div>
</template>

<script setup lang="ts">
import { VBtn, VTooltip } from 'vuetify/components';
defineProps({
  isValid: {
    type: Boolean,
    required: true,
  },
  buttonText: {
    type: String,
    required: true,
  },
  tooltipText: {
    type: String,
    required: true,
  },
  onClick: {
    type: Function,
    required: true,
  },
  icon: {
    type: String,
    default: 'mdi-download'
  }
})
</script>
