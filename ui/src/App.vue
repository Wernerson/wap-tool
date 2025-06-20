<script setup lang="ts">
import { JsonForms, type JsonFormsChangeEvent } from "@jsonforms/vue";
import { dump as dumpYaml, load as loadYaml } from "js-yaml";
import { computed, ref } from "vue";
import schema from "../../schema/wap.json";
import uischema from "./uischhema.json";

import '@mdi/font/css/materialdesignicons.css';

import type { ErrorObject } from "ajv";
import { VApp, VBtn, VContainer, VRow } from "vuetify/components";
import { ajv, localizeAjvErrors } from "./ajv";
import ButtonDisabledTooltip from "./ButtonDisabledTooltip.vue";
import CustomUpload from "./CustomUpload.vue";
import DebugRenderer from "./DebugRenderer.vue";
import { printPdf } from "./print";
import { renderers } from "./renderers/renderers";
import SettingsComponent from "./SettingsComponent.vue";
import { locale, translation } from "./translator";
import { parseTime } from "./utils";

Storage.prototype.setObject = function(key: string, value: object) {
    this.setItem(key, JSON.stringify(value));
}

Storage.prototype.getObject = function(key: string) : object {
    var value = this.getItem(key);
    return value && JSON.parse(value);
}

const data = ref(window.localStorage.getObject("currentWAP") || {});

const errors = ref<ErrorObject[]>([]);
const isValid = computed(() => errors.value.length === 0);

// This functions sets the "title" property of weeks and days, to be displayed in the UI. 
// These are not used in the PDF
const dayOptions :  Intl.DateTimeFormatOptions = { weekday: "long", year: 'numeric', month: 'numeric', day: 'numeric' }
const processData = (value: any) => {
  if (!("meta"in value) || !("firstDay" in value.meta) || !("weeks" in value)) return;

  let firstDate = new Date(value.meta.firstDay);
  for (const week of value.weeks) {
    week.title = "Woche vom " + firstDate.toLocaleDateString("de-CH", dayOptions);
    if ("days" in week) {
      let firstWeekDay = new Date(firstDate);
      for (const day of week.days) {
        day.title = firstWeekDay.toLocaleDateString("de-CH", dayOptions);
        firstWeekDay.setDate(firstWeekDay.getDate() + 1);
      }
    }
    firstDate.setDate(firstDate.getDate() + 7);
  }
}

// Sorts the events in every day by their starting time
const sortData = (data: any) => {
  if (!("weekks" in data)) return;
  for (const week of data.weeks) {
    for (const day of week.days) {
      if (!("events" in day)) continue
      day.events.sort(compareEvents)
    }
  }
}

const compareEvents = (a: any, b: any) => {
  if (!("start" in a) || !("start" in b)) {
    return 0;
  }
  const dateA = parseTime(a.start);
  const dateB = parseTime(b.start);
  return dateA.valueOf() - dateB.valueOf();
}

const onFileChange = (event: any) => {
  const reader = new FileReader()
  reader.onload = (_ev) => {
    const text = reader.result
    if (!text || text instanceof ArrayBuffer) {
      throw "Uploaded file is not a string";
    }
    console.debug("Raw file text:", text)
    const yaml = loadYaml(text)
    console.log("YAML object", yaml)
    sortData(yaml)
    processData(yaml)
    data.value = yaml
  }
  reader.readAsText(event.target.files[0])
}

const getYaml = () => {
  const yamlString = dumpYaml(data.value)
  return yamlString
}

const onClearCliked = (_event: any) => {
  data.value = {};
}

const onDownloadClicked = (_event: any) => {
  const yamlString = getYaml();
  const blob = new Blob([yamlString], { type: 'text/yaml;charset=utf-8' });
  downloadFile("WAP.yml", blob)
}

const onFormChange = (event: JsonFormsChangeEvent) => {
  const value = event.data;
  if (event.errors) {
    localizeAjvErrors(event.errors, locale.value);
    errors.value = event.errors;
  } else {
    errors.value = [];
  }

  insertMissingDays(value);
  processData(value);
  data.value = value;
  window.localStorage.setObject("currentWAP", data.value);
};

// When a new week is added, all 7 days are automaically added as well
const insertMissingDays = (value: any) => {
  if (!("weeks" in value)) return;
  for (const week of value["weeks"]) {
    if (!("days" in week)) week["days"] = [];
    const diffDays = 7 - week["days"].length;
    for (let i = 0; i < diffDays; i++) week["days"].push({});
  }
}

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
  printPdf(data.value);
}
</script>

<template>
  <v-app>
    <SettingsComponent />
    <v-container class="ga-4">
      <v-row justify="center">
        <h1>WAUI - WAP Tool UI</h1>
      </v-row>
      <v-row justify="center">
        <CustomUpload 
          :onChange="onFileChange"
          accept=".yml, .yaml"
        />
      </v-row>
      <v-row justify="center" class="ma-4">
        <v-btn prepend-icon="mdi-close"
        @click="onClearCliked">{{ translation("general.emptyWap", "WAP leeren") }}</v-btn>
      </v-row>
      <v-row justify="center" class="ga-4">
        <ButtonDisabledTooltip 
          :isValid="isValid"
          :buttonText="translation('general.download', 'YAML')"
          :tooltipText="translation('general.errorsExist', 'Errors exist')"
          :onClick="onDownloadClicked"
        />
        <ButtonDisabledTooltip 
          :isValid="isValid"
          :buttonText="translation('general.downloadPdf', 'YAML')"
          :tooltipText="translation('general.errorsExist', 'Errors exist')"
          :onClick="onConvertClicked"
        />
      </v-row>
    </v-container>

    <div class="myform">
      <JsonForms
      :data="data"
      :renderers="renderers"
      :schema="schema"
      :uischema="uischema"
      @change="onFormChange" 
      :i18n="{locale: locale, translate: translation}"
      :ajv="ajv"
      valid/>
    </div>
    <v-container>
      <DebugRenderer 
      :data="data"/>
    </v-container>
  </v-app>
</template>

<style scoped>
@import '@jsonforms/vue-vuetify/lib/jsonforms-vue-vuetify.css';
</style>
