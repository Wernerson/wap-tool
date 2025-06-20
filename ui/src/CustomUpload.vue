<template>
    <v-container>
            <v-file-upload 
            density="compact"
            v-bind="$attrs"
            :title="translation('general.upload.title', 'WAP hier hochladen')"
            @change="onFileChange" 
            :accept="props.accept"
            v-model="files"
            />
        </v-container>
        <template>
        <v-dialog
        v-model="dialog"
        width="auto">
            <v-card
            max-width="400"
            prepend-icon="mdi-file-alert"
            :title="translation('general.upload.dialogTitle', 'Dateiformat nicht unterstützt')"
            :text="translation('general.upload.dialogText', 'Es sind nur Dateien im yaml Format zugelassen')"
            >
                <template v-slot:actions>
                    <v-btn
                    class="ms-auto"
                    :text="translation('general.upload.dialogConfirm', 'Ok')"
                    @click="dialog = false"
                    />
                </template>
            </v-card>
        </v-dialog>
        </template>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { VBtn, VCard, VContainer, VDialog } from 'vuetify/components';
import { VFileUpload } from 'vuetify/labs/VFileUpload';
import { translation } from './translator';

const props = defineProps(["onChange", "accept"]);

const files = ref([])
const dialog = ref(false);

const onFileChange = (event: any) => {
    const accept = props.accept;
    if (accept) {
        const fileExtensions = accept.split(",").map((x: string) => x.trim());
        const filename = event.target.files[0].name;
        let valid = fileExtensions.some((ext: string) => filename.endsWith(ext));
        if (valid) {
            props.onChange(event);
            return;
        }
        files.value = [];
        dialog.value = true;
    }
};
</script>