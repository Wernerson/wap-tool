import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'

import { mdi, aliases as mdiAliases } from 'vuetify/iconsets/mdi';
import { createVuetify } from 'vuetify';
import { mdiIconAliases } from '@jsonforms/vue-vuetify';
import '@mdi/font/css/materialdesignicons.css';


const vuetify = createVuetify({
  icons: {
      defaultSet: 'mdi',
      sets: {
        mdi,
      },
      aliases: { ...mdiAliases, ...mdiIconAliases },
    },
    defaults: {}
});

createApp(App)
.use(vuetify).mount('#app')
