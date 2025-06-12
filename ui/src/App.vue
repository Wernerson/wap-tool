<script setup lang="ts">
import { ref, markRaw } from "vue";
import { JsonForms, type JsonFormsChangeEvent } from "@jsonforms/vue";
import { extendedVuetifyRenderers } from '@jsonforms/vue-vuetify';
import schema from "../../schema/wap.json"
import { load as loadYaml, dump as dumpYaml} from "js-yaml";

import '@mdi/font/css/materialdesignicons.css';
import CollapsibleGroupRenderer from "./renderers/CollapsibleGroupRenderer.vue";
import ColorPickerRenderer from "./renderers/ColorPickerRenderer.vue";
import { and, formatIs, isControl, optionIs, rankWith, schemaMatches, schemaTypeIs, uiTypeIs, type ControlElement, type JsonSchema } from "@jsonforms/core";
import CategoryPickerRenderer from "./renderers/CategoryPickerRenderer.vue";
import AppearsInRenderer from "./renderers/AppearsInRenderer.vue";

const renderers = markRaw([
  ...extendedVuetifyRenderers,
  { tester: rankWith(3, uiTypeIs("CollapsibleGroup")), renderer: CollapsibleGroupRenderer },
  { tester: rankWith(3, uiTypeIs("ColorPicker")), renderer: ColorPickerRenderer},
  { tester: rankWith(3, uiTypeIs("CategoryPicker")), renderer: markRaw(CategoryPickerRenderer)},
  { tester: rankWith(
  10, // Adjust rank to override default array renderer
  (uischema, schema) =>
    isControl(uischema) &&
    (uischema as any).scope?.endsWith('appearsIn')
), renderer: AppearsInRenderer}
  // here you can add custom renderers
]);

const timePickerOptions = {
                    format: "time",
                    ampm: false,
                    timeFormat: "HH:mm",
                    timeSaveFormat: "HH:mm"
                  }

const uischema = {
  type: "VerticalLayout",
  elements: [
    {
      type: "CollapsibleGroup",
      label: "Informationen",
      elements: [
        {
          type: "HorizontalLayout",
          elements: [
            {
              type: "Control",
              scope: "#/properties/meta/properties/title"
            },
            {
              type: "Control",
              scope: "#/properties/meta/properties/unit"
            },
          ]
        },
        {
          type: "HorizontalLayout",
          elements: [
            {
              type: "Control",
              scope: "#/properties/meta/properties/author"
            },
            {
              type: "Control",
              scope: "#/properties/meta/properties/version"
            },
          ]
        },
      ],
    },
    {
      type: "CollapsibleGroup",
      label: "Zeiten",
      elements: [
        {
          type: "VerticalLayout",
          elements: [
            {
              type: "HorizontalLayout",
              elements: [
                {
                  type: "Control",
                  scope: "#/properties/meta/properties/startTime",
                  options: timePickerOptions
                },
                {
                  type: "Control",
                  scope: "#/properties/meta/properties/endTime",
                  options: timePickerOptions
                },
              ]
            },
            {
              type: "HorizontalLayout",
              elements: [
                {
                  type: "Control",
                  scope: "#/properties/meta/properties/firstDay",
                  options: {
                    format: "date",
                    dateFormat: "YYYY-MM-DD",
                    dateSaveFormat: "YYYY-MM-DD",
                  }
                },
              ]
            }
          ]
        }
      ]
    },
    {
      type: "CollapsibleGroup",
      label: "Kategorien",
      options: {
        defaultOpen: false
      },
      elements: [
        {
          type: "Control",
          scope: "#/properties/categories",
          options: {
            detail: {
              type: "VerticalLayout",
              elements: [
                {
                  type: "HorizontalLayout",
                  elements: [
                    {
                      type: "Control",
                      scope: "#/properties/identifier"
                    },
                    {
                      type: "Control",
                      scope: "#/properties/comment"
                    }
                  ]
                },
                {
                  type: "HorizontalLayout",
                  elements: [
                    {
                      type: "ColorPicker",
                      scope: "#/properties/color",
                      options: {
                        defaultColor: "#f0f0f0"
                      }
                    },
                    {
                      type: "ColorPicker",
                      scope: "#/properties/textColor",
                      options: {
                        defaultColor: "#000000"
                      }
                    }
                  ]
                }
              ]
            }
          }
        }
      ]
    },
    {
      type: "HorizontalLayout",
      elements: [
        {
          type: "ListWithDetails",
          scope: "#/properties/weeks",
          options: {
            detail: {
              type: "VerticalLayout",
              elements: [
                {
                  type: "CollapsibleGroup",
                  label: "Wochenbemerkungen",
                  elements: [
                    {
                      type: "Control",
                      scope: "#/properties/remarks",
                      label: "",
                      options: {
                        showSortButtons: true
                      }
                    },
                  ]
                },
                {
                  type: "Control",
                  scope: "#/properties/days",
                  options: {
                    detail: {
                    type: "VerticalLayout",
                    elements: [
                      {
                        type: "Control",
                        scope: "#/properties/columns"
                      },
                      {
                        type: "Control",
                        scope: "#/properties/remarks"
                      },
                      {
                        type: "Control",
                        scope: "#/properties/events",
                        options: {
                          detail: {
                            type: "VerticalLayout",
                            elements: [
                              {
                                type: "HorizontalLayout",
                                elements: [
                                  {
                                    type: "Control",
                                    scope: "#/properties/title"
                                  },
                                  {
                                    type: "Control",
                                    scope: "#/properties/description",
                                  },
                                ]
                              },
                              {
                                type: "HorizontalLayout",
                                elements: [
                                  {
                                    type: "Control",
                                    scope: "#/properties/start",
                                    options: timePickerOptions
                                  },
                                  {
                                    type: "Control",
                                    scope: "#/properties/end",
                                    options: timePickerOptions
                                  },
                                ]
                              },
                              {
                                type: "HorizontalLayout",
                                elements: [
                                  {
                                    type: "CategoryPicker",
                                    scope: "#/properties/category",
                                    options: {
                                      source: "categories"
                                    }
                                  },
                                ]
                              },
                              {
                                type: "HorizontalLayout",
                                elements: [
                                  {
                                    type: "Control",
                                    scope: "#/properties/openEnd"
                                  },
                                  {
                                    type: "Control",
                                    scope: "#/properties/footnote"
                                  },
                                  {
                                    type: "Control",
                                    scope: "#/properties/forceHorizontalText"
                                  }
                                ]
                              },
                              {
                                type: "HorizontalLayout",
                                elements: [
                                  {
                                    type: "Control",
                                    scope: "#/properties/appearsIn"
                                  }
                                ]
                              }
                            ]
                          }
                        }
                      }
                    ]
                  }
                  }
                }
              ]
            },
          },
        },
      ],
    },
  ],
};

const data = ref({})

const processData = (value: any) => {
  if ("meta" in value && "firstDay" in value.meta && "weeks" in value) {
    let firstDate = new Date(value.meta.firstDay);
    for (const week of value.weeks) {
      week.title = "Woche vom " + firstDate.toLocaleDateString("de-CH", {weekday: "long", year: 'numeric', month: 'numeric', day: 'numeric' });
      if ("days" in week) { 
        let firstWeekDay = new Date(firstDate);
        for (const day of week.days) {
          day.title = firstWeekDay.toLocaleDateString("de-CH", {weekday: "long", year: 'numeric', month: 'numeric', day: 'numeric' });
          firstWeekDay.setDate(firstWeekDay.getDate() + 1);
        }
      }
      firstDate.setDate(firstDate.getDate() + 7);
    }
  }
}

const sortData = (data: any) => {
  if ("weeks" in data) {
    for (const week of data.weeks) {
      for (let day of week.days) {
        if ("events" in day) {
          const today = new Date().toISOString().split('T')[0]; // e.g., "2025-06-11"
          day.events.sort((a: any, b: any) => {
            if (!("start" in a) || !("start" in b)) {
              return 0;
            }
            const dateA = new Date(`${today}T${a.start}:00`);
            const dateB = new Date(`${today}T${b.start}:00`);
            return dateA.valueOf() - dateB.valueOf();
          })
        }
      }
    }
  }
}

const onFileChange = (event: any) => {
  const reader = new FileReader()
  reader.onload = (_ev) => {
    const text = reader.result
    console.debug("Raw file text:", text)
    const yaml = loadYaml(text, "utf8")
    console.log("YAML object", yaml)
    sortData(yaml)
    processData(yaml)
    data.value = yaml
  }
  reader.readAsText(event.target.files[0])
}

const onDownloadClicked = (_event: any) => {
  console.log(data)
  const yamlString = dumpYaml(data.value)
  // Create a Blob from YAML string
  const blob = new Blob([yamlString], { type: 'text/yaml;charset=utf-8' });
  downloadFile("WAP.yml", blob)
}

const onFormChange = (event: JsonFormsChangeEvent) => {
  const value = event.data;
  processData(value);
  console.log(value)
  data.value = value;
};

function downloadFile(filename: string, file: Blob) {
  // Create a temporary link element
    const url = URL.createObjectURL(file);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename; // file name

    // Append to body, trigger click, and remove
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);

    // Release object URL
    URL.revokeObjectURL(url);
}

async function onConvertClicked(_event: any) {
  const formData = new FormData();

  const yamlString = dumpYaml(data.value)
  // Create a Blob from YAML string
  const blob = new Blob([yamlString], { type: 'text/yaml;charset=utf-8' });

  formData.append("file", blob);

  try {
    const response = await fetch("http://localhost:8080/upload", {
      method: "POST",
      body: formData,
    });
    
    downloadFile("WAP.pdf", await response.blob());
  } catch (e) {
    console.error(e);
  }
}
</script>

<template>
  <v-app>
    <header>
      <h1>WAUI - WAP Tool UI</h1>
      <input type="file" @change="onFileChange" accept=".yml, .yaml"/>
      <button @click="onDownloadClicked">Download</button>
      <button @click="onConvertClicked">Convert to PDF</button>
    </header>

    <div class="myform">
      <JsonForms 
      :data="data" 
      :renderers="renderers" 
      :schema="schema" 
      :uischema="uischema"
      @change="onFormChange" />
    </div>
    <pre>{{ data }}</pre>
  </v-app>
</template>

<style scoped>
@import '@jsonforms/vue-vuetify/lib/jsonforms-vue-vuetify.css';
</style>
