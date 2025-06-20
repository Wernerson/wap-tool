<template>
  <div>
    <v-menu
      v-model="menuOpen"
      :close-on-content-click="false"
      offset-y
      location="bottom end"
    >
      <template #activator="{ props }">
        <v-btn v-bind="props" icon>
            <v-icon>mdi-cog</v-icon>
        </v-btn>
      </template>

      <v-card class="pa-4" min-width="200">
        <v-select
          v-model="theme"
          :items="themes"
          :label="translation('general.settings.theme', 'Theme')"
          density="compact"
          variant="outlined"
          hide-details
        />
        <v-select
          v-model="locale"
          :items="locales"
          :label="translation('general.settings.language', 'Sprache')"
          density="compact"
          variant="outlined"
          hide-details
          class="mt-4"
        />
      </v-card>
    </v-menu>
  </div>
</template>

<script setup lang="ts">
import { usePreferredDark } from '@vueuse/core'
import { ref, watch } from 'vue'
import { useTheme } from 'vuetify'
import { VBtn, VCard, VIcon, VMenu, VSelect } from 'vuetify/components'
import { allLocales, locale, type Locale, translation } from './translator'

const allThemes = ["light", "dark"]
type Theme = typeof allThemes[number];

// Theme + locale options
const themes: Theme[] = ['light', 'dark']
const locales: { title: string; value: Locale }[] = [
  { title: 'Deutsch', value: 'de' },
  { title: "Français", value: "fr"},
  { title: "Italiano", value: "it"},
  { title: 'English', value: 'en' },
]

const preferedLocale = window.localStorage.getItem("wapLocale") || getLang();
locale.value = allLocales.includes(preferedLocale) 
  ? preferedLocale
  : "de";

const themeInstance = useTheme()
const isDark = usePreferredDark()

themeInstance.global.name.value = window.localStorage.getItem("wapTheme") || (isDark.value ? 'dark' : 'light')

// State
const menuOpen = ref(false)
const theme = ref<Theme>(allThemes.includes(themeInstance.global.name.value) 
  ? themeInstance.global.name.value 
  : "light")

watch(theme, (newTheme) => {
  window.localStorage.setItem("wapTheme", newTheme);
  themeInstance.global.name.value = newTheme;
})

// Emit or handle locale changes
watch(locale, (newLocale) => {
  console.log(`Locale changed to: ${newLocale}`);
  window.localStorage.setItem("wapLocale", locale.value);
})

function getLang() {
  let language;
  if (navigator.languages !== undefined) {
    language = navigator.languages[0]; 
  } else {
    language = navigator.language;
  }
  return language.split("-")[0];
}
</script>
