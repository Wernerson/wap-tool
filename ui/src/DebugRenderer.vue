<template>
    <VCard class="mb-4" elevation="2">
    <VCardTitle
      @click="expanded = !expanded"
      class="cursor-pointer d-flex justify-space-between align-center"
    >
      Debug Info
      <VIcon :icon="expanded ? 'mdi-chevron-up' : 'mdi-chevron-down'" />
    </VCardTitle>

    <VExpandTransition>
      <VCardText v-show="expanded">
        <pre>{{ props.data }}</pre>
      </VCardText>
    </VExpandTransition>
  </VCard>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { VCard, VCardText, VCardTitle, VExpandTransition, VIcon } from 'vuetify/components';

const props = defineProps(["data"]);
const expanded = ref(false);
</script>