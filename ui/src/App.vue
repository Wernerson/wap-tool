<script setup lang="ts">
import { ref, provide } from "vue";
import { JsonForms, JsonFormsChangeEvent } from "@jsonforms/vue";
import { defaultStyles, mergeStyles, vanillaRenderers } from "@jsonforms/vue-vanilla";
import schema from "../../schema/wap.json"
import { load as loadYaml, dump as dumpYaml} from "js-yaml"

const renderers = Object.freeze([
  ...vanillaRenderers,
  // here you can add custom renderers
]);

const uischema = {
  type: "VerticalLayout",
  elements: [
    {
      type: "VerticalLayout",
      elements: [
        {
          type: "Control",
          scope: "#/properties/meta",
        },
      ],
    },
  ],
};

const data = ref({})
const onFileChange = (event: any) => {
  const reader = new FileReader()
  reader.onload = (ev) => {
    const text = reader.result
    console.debug("Raw file text:", text)
    const yaml = loadYaml(text, "utf8")
    console.log("YAML object", yaml)
    data.value = yaml
  }
  reader.readAsText(event.target.files[0])
}

const onDownloadClicked = (event: any) => {
  console.log(data)
  const yamlString = dumpYaml(data.value)
  // Create a Blob from YAML string
  const blob = new Blob([yamlString], { type: 'text/yaml;charset=utf-8' });

  // Create a temporary link element
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'WAP.yaml'; // file name

  // Append to body, trigger click, and remove
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);

  // Release object URL
  URL.revokeObjectURL(url);
}

const onFormChange = (event: JsonFormsChangeEvent) => {
  data.value = event.data;
};
</script>

<template>
  <header>
    <h1>WAUI - WAP Tool UI</h1>
    <input type="file" @change="onFileChange" accept=".yml, .yaml"/>
    <button @click="onDownloadClicked">Download</button>
  </header>

  <div class="myform">
    <JsonForms :data="data" :renderers="renderers" :schema="schema" :uischema="uischema" @change="onFormChange" />
  </div>
  <pre>{{ data }}</pre>
</template>

<style scoped>

</style>
