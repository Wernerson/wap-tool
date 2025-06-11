<template>
  <v-card class="mb-4" elevation="2">
    <v-card-title
      @click="expanded = !expanded"
      class="cursor-pointer d-flex justify-between align-center"
    >
      {{ label }}
      <i 
      class="mdi"
      :class="{ 'mdi-chevron-up': expanded, 'mdi-chevron-down': !expanded }"></i>
    </v-card-title>

    <v-expand-transition>
      <v-card-text v-if="expanded">
        <component
          v-for="(element, index) in groupLayout.elements"
          :key="index"
          :is="DispatchRenderer"
          :schema="schema"
          :uischema="element"
          :path="path"
          :enabled="enabled"
          :renderers="renderers"
          :cells="cells"
        />
      </v-card-text>
    </v-expand-transition>
  </v-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { DispatchRenderer } from '@jsonforms/vue'
import type { GroupLayout, LayoutProps } from '@jsonforms/core'
import { mdiChevronDown, mdiChevronUp } from '@mdi/js';

const expanded = ref(true)

const props = defineProps<LayoutProps>()

const groupLayout = computed(() => props.uischema as GroupLayout)

const label = computed(() => {
  const lbl = groupLayout.value.label
  return lbl && lbl.trim() !== '' ? lbl : 'Group'
})

const { schema, path, enabled, renderers, cells } = props
</script>
