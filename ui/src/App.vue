<script setup lang="ts">
import { load as loadYaml} from "js-yaml"

const onChange = (event: any) => {
  const reader = new FileReader()
  reader.onload = (ev) => {
    const text = reader.result
    console.debug("Raw file text:", text)
    const yaml = loadYaml(text, "utf8")
    console.log("YAML object", yaml)
  }
  reader.readAsText(event.target.files[0])
}
</script>

<template>
  <header>
    WAUI - WAP Tool UI
  </header>

  <main>
    <input type="file" @change="onChange" />
  </main>
</template>

<style scoped>

</style>
