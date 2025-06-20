import { isControl, isObjectArrayControl, isPrimitiveArrayControl, rankWith, uiTypeIs } from "@jsonforms/core";
import { extendedVuetifyRenderers } from "@jsonforms/vue-vuetify";

import { markRaw } from "vue";
import AppearsInRenderer from "./AppearsInRenderer.vue";
import CategoryPickerRenderer from "./CategoryPickerRenderer.vue";
import CollapsibleGroupRenderer from "./CollapsibleGroupRenderer.vue";
import ColorPickerRenderer from "./ColorPickerRenderer.vue";
import arrayRenderer from "./CustomArrayLayoutRenderer.vue";
import eventsArrayRenderer from "./CustomEventsArrayRenderer.vue";
import primitiveArrayRenderer from "./CustomPrimitiveArrayLayoutRenderer.vue";

export const renderers = markRaw([
  ...extendedVuetifyRenderers,
  { tester: rankWith(20, uiTypeIs("EventList")), renderer: eventsArrayRenderer},
  { tester: rankWith(10, isObjectArrayControl), renderer: arrayRenderer},
  { tester: rankWith(10, isPrimitiveArrayControl), renderer: primitiveArrayRenderer},
  { tester: rankWith(3, uiTypeIs("CollapsibleGroup")), renderer: CollapsibleGroupRenderer },
  { tester: rankWith(3, uiTypeIs("ColorPicker")), renderer: ColorPickerRenderer},
  { tester: rankWith(20, uiTypeIs("CategoryPicker")), renderer: markRaw(CategoryPickerRenderer)},
  { tester: rankWith(
  20, (uischema) =>
    isControl(uischema) &&
    (uischema as any).scope?.endsWith('appearsIn')
), renderer: AppearsInRenderer}
]);