<template>
  <v-card
    v-if="control.visible"
    :class="styles.arrayList.root"
    v-bind="vuetifyProps('v-card')"
    style="padding-bottom: 16px"
  >
    <v-card-text v-bind="vuetifyProps('v-card-text')">
      <v-container
        justify-space-around
        align-content-center
        :class="styles.arrayList.container"
        style="padding-block: 0"
      >
        <v-row justify="center">
          <v-expansion-panels
            accordion
            v-bind="expansionPanelsProps"
            v-model="currentlyExpanded"
          >
            <v-expansion-panel
              v-for="(_element, index) in control.data"
              :key="`${control.path}-${control.data.length}-${index}`"
              :class="styles.arrayList.item"
            >
              <v-expansion-panel-title
                :class="`${styles.arrayList.itemHeader} py-0`"
              >
                <v-container :class="`${styles.arrayList.itemContainer} py-0`">
                  <v-row
                    :style="`display: grid; grid-template-columns: ${
                      !hideAvatar ? 'min-content' : ''
                    } auto min-content ${
                      appliedOptions.showSortButtons
                        ? 'min-content min-content'
                        : ''
                    }`"
                  >
                    <v-col v-if="!hideAvatar" align-self="center" class="pl-0">
                      <validation-badge
                        overlap
                        bordered
                        :errors="childErrors(index)"
                      >
                        <v-avatar size="40" aria-label="Index" color="primary">
                          <span class="primary--text text--lighten-5">{{
                            index + 1
                          }}</span></v-avatar
                        >
                      </validation-badge>
                    </v-col>

                    <v-col
                      align-self="center"
                      :class="`pl-0 text-truncate ${styles.arrayList.itemLabel}`"
                      >{{ childLabelForIndex(index) }}</v-col
                    >
                    <v-col
                      align-self="center"
                      v-if="appliedOptions.showSortButtons"
                    >
                      <v-tooltip bottom>
                        <template v-slot:activator="{ props }">
                          <v-btn
                            v-bind="props"
                            icon
                            variant="text"
                            elevation="0"
                            small
                            class="v-expansion-panel-title__icon"
                            :aria-label="translations.upAriaLabel"
                            :disabled="index <= 0 || !control.enabled"
                            :class="styles.arrayList.itemMoveUp"
                            @click="moveUpClick($event, index)"
                          >
                            <v-icon class="notranslate">{{
                              icons.current.value.itemMoveUp
                            }}</v-icon>
                          </v-btn>
                        </template>
                        {{ translations.up }}
                      </v-tooltip>
                    </v-col>
                    <v-col
                      align-self="center"
                      v-if="appliedOptions.showSortButtons"
                    >
                      <v-tooltip bottom>
                        <template v-slot:activator="{ props }">
                          <v-btn
                            v-bind="props"
                            icon
                            variant="text"
                            elevation="0"
                            small
                            class="v-expansion-panel-title__icon"
                            :aria-label="translations.downAriaLabel"
                            :disabled="
                              index >= dataLength - 1 || !control.enabled
                            "
                            :class="styles.arrayList.itemMoveDown"
                            @click="moveDownClick($event, index)"
                          >
                            <v-icon class="notranslate">{{
                              icons.current.value.itemMoveDown
                            }}</v-icon>
                          </v-btn>
                        </template>
                        {{ translations.down }}
                      </v-tooltip>
                    </v-col>
                    <v-col align-self="center">
                      <v-tooltip bottom>
                        <template v-slot:activator="{ props }">
                          <v-btn
                            v-bind="props"
                            icon
                            variant="text"
                            elevation="0"
                            small
                            class="v-expansion-panel-title__icon"
                            :aria-label="translations.removeAriaLabel"
                            :class="styles.arrayList.itemDelete"
                            :disabled="
                              !control.enabled ||
                              (appliedOptions.restrict &&
                                control.arraySchema !== undefined &&
                                control.arraySchema.minItems !== undefined &&
                                dataLength <= control.arraySchema.minItems)
                            "
                            @click.stop="suggestToDelete = index"
                          >
                            <v-icon class="notranslate">{{
                              icons.current.value.itemDelete
                            }}</v-icon>
                          </v-btn>
                        </template>
                        {{ translations.removeTooltip }}
                      </v-tooltip>
                    </v-col>
                  </v-row>
                </v-container>
              </v-expansion-panel-title>
              <v-expansion-panel-text :class="styles.arrayList.itemContent">
                <dispatch-renderer
                  :schema="control.schema"
                  :uischema="foundUISchema"
                  :path="composePaths(control.path, `${index}`)"
                  :enabled="control.enabled"
                  :renderers="control.renderers"
                  :cells="control.cells"
                />
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-row>
      </v-container>
      <v-col
      class="d-flex justify-center align-center flex-column">
          <slot
              name="toolbar-elements"
              :addClass="styles.arrayList.addButton"
              :addDisabled="addDisabled"
              :addClick="addButtonClick"
              :control="control"
              :appliedOptions="appliedOptions"
              :styles="styles"
              :icons="icons"
          >
              <v-tooltip bottom>
                  <template v-slot:activator="{ props }">
                      <v-btn
                      icon
                      variant="text"
                      elevation="0"
                      small
                      :aria-label="control.translations.addAriaLabel"
                      v-bind="props"
                      :class="styles.arrayList.addButton"
                      :disabled="addDisabled"
                      @click="addButtonClick"
                      >
                      <v-icon>{{ icons.current.value.itemAdd }}</v-icon>
                      </v-btn>
                  </template>
                  {{ translations.addTooltip }}
              </v-tooltip>
          </slot>
          <div 
          v-for="item in control.childErrors"
          class="text-red">
              {{ item.message }}
          </div>
      </v-col>
      <v-container v-if="dataLength === 0" :class="styles.arrayList.noData" style="padding-block: 0">
        {{ translations.noDataMessage }}
      </v-container>
    </v-card-text>
    <v-card-actions
      v-if="$slots.actions"
      class="pb-8"
      v-bind="vuetifyProps('v-card-actions')"
    >
      <slot
        name="actions"
        :addClass="styles.arrayList.addButton"
        :addDisabled="addDisabled"
        :addClick="addButtonClick"
        :control="control"
        :appliedOptions="appliedOptions"
        :styles="styles"
        :icons="icons"
      >
      </slot>
    </v-card-actions>
    <v-dialog
      :model-value="suggestToDelete !== null"
      max-width="600"
      @keydown.esc="suggestToDelete = null"
      @click:outside="suggestToDelete = null"
    >
      <v-card>
        <v-card-title class="text-h5">
          {{ translations.deleteDialogTitle }}
        </v-card-title>

        <v-card-text>
          {{ translations.deleteDialogMessage }}
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>

          <v-btn variant="text" @click="suggestToDelete = null">
            {{ translations.deleteDialogDecline }}</v-btn
          >
          <v-btn
            variant="text"
            ref="confirm"
            @click="
              removeItemsClick(
                suggestToDelete === null ? null : [suggestToDelete],
              );
              suggestToDelete = null;
            "
          >
            {{ translations.deleteDialogAccept }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts">
import { translation } from '@/translator';
import {
  composePaths,
  createDefaultValue,
  findUISchema,
  getControlPath,
  type ControlElement,
  type UISchemaElement,
} from '@jsonforms/core';
import {
  DispatchRenderer,
  rendererProps,
  useJsonFormsArrayControl,
  type RendererProps,
} from '@jsonforms/vue';
import { useIcons, useNested, useVuetifyArrayControl, ValidationBadge, ValidationIcon } from '@jsonforms/vue-vuetify';
import type { ErrorObject } from 'ajv';
import merge from 'lodash/merge';
import { computed, defineComponent, ref } from 'vue';
import {
  VAvatar,
  VBtn,
  VCard,
  VCardActions,
  VCardText,
  VCardTitle,
  VCol,
  VContainer,
  VDialog,
  VExpansionPanel,
  VExpansionPanels,
  VExpansionPanelText,
  VExpansionPanelTitle,
  VIcon,
  VRow,
  VSpacer,
  VToolbar,
  VToolbarTitle,
  VTooltip,
} from 'vuetify/components';

const arrayRenderer = defineComponent({
  name: 'array-layout-renderer',
  components: {
    DispatchRenderer,
    VCard,
    VCardActions,
    VCardTitle,
    VCardText,
    VAvatar,
    VDialog,
    VRow,
    VCol,
    VToolbar,
    VToolbarTitle,
    VTooltip,
    VIcon,
    VBtn,
    VSpacer,
    VExpansionPanels,
    VExpansionPanel,
    VExpansionPanelTitle,
    VExpansionPanelText,
    VContainer,
    ValidationIcon,
    ValidationBadge,
  },
  props: {
    ...rendererProps<ControlElement>(),
  },
  setup(props: RendererProps<ControlElement>) {
    const control = useVuetifyArrayControl(useJsonFormsArrayControl(props));
    const currentlyExpanded = ref<null | number>(
      control.appliedOptions.value.initCollapsed ? null : 0,
    );
    const expansionPanelsProps = computed(() =>
      merge(
        { flat: false, focusable: true },
        control.vuetifyProps('v-expansion-panels'),
      ),
    );
    const suggestToDelete = ref<null | number>(null);
    // indicate to our child renderers that we are increasing the "nested" level
    useNested('array');

    const icons = useIcons();

    const t = computed(() => translation.value);
    const keyPrefix = control.control.value.i18nKeyPrefix;
    const t_prefix = computed(() => (key: string, defaultValue: string) => t.value(keyPrefix + "." + key, defaultValue));

    const translations = computed(() => ({
      up: t.value('general.array.up', 'Move up'),
      down: t.value('general.array.down', 'Move down'),
      addTooltip: t_prefix.value('addTooltip', 'Add item'),
      removeTooltip: t_prefix.value('removeTooltip', 'Remove item'),
      deleteDialogTitle: t_prefix.value('deleteDialogTitle', 'Delete item'),
      deleteDialogMessage: t_prefix.value('deleteDialogMessage', 'Are you sure?'),
      deleteDialogAccept: t_prefix.value('deleteDialogAccept', 'Yes'),
      deleteDialogDecline: t_prefix.value('deleteDialogDecline', 'No'),
      upAriaLabel: t_prefix.value('upAriaLabel', 'Move item up'),
      downAriaLabel: t_prefix.value('downAriaLabel', 'Move item down'),
      removeAriaLabel: t_prefix.value('removeAriaLabel', 'Remove item'),
      noDataMessage: t_prefix.value('noDataMessage', 'No data'),
    }));
    
    return {
      ...control,
      currentlyExpanded,
      expansionPanelsProps,
      suggestToDelete,
      icons,
      translations
    };
  },
  computed: {
    addDisabled(): boolean {
      return (
        !this.control.enabled ||
        (this.appliedOptions.restrict &&
          this.control.arraySchema !== undefined &&
          this.control.arraySchema.maxItems !== undefined &&
          this.dataLength >= this.control.arraySchema.maxItems)
      );
    },
    dataLength(): number {
      return this.control.data ? this.control.data.length : 0;
    },
    foundUISchema(): UISchemaElement {
      return findUISchema(
        this.control.uischemas,
        this.control.schema,
        this.control.uischema.scope,
        this.control.path,
        undefined,
        this.control.uischema,
        this.control.rootSchema,
      );
    },
    hideAvatar(): boolean {
      return !!this.appliedOptions.hideAvatar;
    },
  },
  methods: {
    composePaths,
    createDefaultValue,
    addButtonClick() {
      this.addItem(
        this.control.path,
        createDefaultValue(this.control.schema, this.control.rootSchema),
      )();
      if (!this.appliedOptions.collapseNewItems && this.control.data?.length) {
        this.currentlyExpanded = this.dataLength - 1;
      }
    },
    moveUpClick(event: Event, toMove: number): void {
      event.stopPropagation();
      this.moveUp?.(this.control.path, toMove)();
    },
    moveDownClick(event: Event, toMove: number): void {
      event.stopPropagation();
      this.moveDown?.(this.control.path, toMove)();
    },
    removeItemsClick(toDelete: number[] | null): void {
      if (toDelete !== null) {
        this.removeItems?.(this.control.path, toDelete)();
      }
    },
    childErrors(index: number): ErrorObject[] {
      return this.control.childErrors.filter((e) => {
        const errorDataPath = getControlPath(e);
        return errorDataPath.startsWith(
          this.composePaths(this.control.path, `${index}`),
        );
      });
    },
  },
});

export default arrayRenderer;
</script>

<style scoped>
.notranslate {
  transform: none !important;
}

:deep(.v-toolbar__content) {
  padding-left: 0;
}
</style>