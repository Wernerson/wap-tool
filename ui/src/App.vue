<script setup lang="ts">
import {computed, ref} from "vue"
import { load as loadYaml} from "js-yaml"

type State = {
  meta: Object
}

const state = ref(null)
const onChange = (event: any) => {
  const reader = new FileReader()
  reader.onload = (ev) => {
    const text = reader.result
    console.debug("Raw file text:", text)
    const yaml = loadYaml(text, "utf8")
    console.log("YAML object", yaml)
    state.value = yaml
  }
  reader.readAsText(event.target.files[0])
}
</script>

<template>
  <header>
    <h1>WAUI - WAP Tool UI</h1>
    <input type="file" @change="onChange" />
    <button>Download</button>
  </header>

  <main v-if="state">
    <section>
      <label for="author">Autor:</label>
      <input v-model="state.meta.author" placeholder="Autor" />
      {{ state.meta.author }}
    </section>
  </main>
</template>

<style scoped>

</style>
