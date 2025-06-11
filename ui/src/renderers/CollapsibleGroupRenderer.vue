<template>
  <VCard class="mb-4" elevation="2">
    <VCardTitle
      @click="expanded = !expanded"
      class="cursor-pointer d-flex justify-space-between align-center"
    >
      {{ label }}
      <VIcon :icon="expanded ? 'mdi-chevron-up' : 'mdi-chevron-down'" />
    </VCardTitle>

    <VExpandTransition>
      <VCardText v-show="expanded">
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
      </VCardText>
    </VExpandTransition>
  </VCard>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { VExpandTransition, VIcon, VCard, VCardText, VCardTitle } from 'vuetify/components'
import { DispatchRenderer } from '@jsonforms/vue'
import type { GroupLayout, LayoutProps } from '@jsonforms/core'

const expanded = ref(true)

const props = defineProps<LayoutProps>()

const groupLayout = computed(() => props.uischema as GroupLayout)

const label = computed(() => {
  const lbl = groupLayout.value.label
  return lbl && lbl.trim() !== '' ? lbl : 'Group'
})

const { schema, path, enabled, renderers, cells } = props
</script>
