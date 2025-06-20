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
import { locale } from '@/translator';
import type { GroupLayout, JsonFormsCellRendererRegistryEntry, JsonFormsRendererRegistryEntry, JsonSchema, Layout } from '@jsonforms/core';
import { DispatchRenderer } from '@jsonforms/vue';
import { useJsonForms } from '@jsonforms/vue-vuetify';
import { computed, ref } from 'vue';
import { VCard, VCardText, VCardTitle, VExpandTransition, VIcon } from 'vuetify/components';

const props = defineProps<{
  uischema: Layout; 
  schema: JsonSchema;
  path: string;
  enabled: boolean;
  renderers: JsonFormsRendererRegistryEntry[];
  cells: JsonFormsCellRendererRegistryEntry[];
}>()

const groupLayout = computed(() => props.uischema as GroupLayout)

const expanded = ref(groupLayout.value.options?.defaultOpen ?? true)

const ctxt = useJsonForms();

const label = computed(() => {
  const translate = ctxt.i18n?.translate;
  const _ = locale.value;
  const lbl = groupLayout.value.i18n;
  return translate!(lbl! + ".label", "Group", ctxt);
});

const { schema, path, enabled, renderers, cells } = props
</script>
