<template>
  <q-dialog full-width ref="dialogRef" @hide="onDialogHide" @keyup.enter="onDialogOK">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Conversion Logs</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-editor
          readonly
          v-model="logs"
          min-height="5rem"/>
      </q-card-section>
      <q-card-actions align="right">
        <q-btn
          color="accent"
          :loading="postLoading"
          flat
          round
          icon="done"
          @click="onDialogOK"/>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import {useDialogPluginComponent} from 'quasar'
import {onMounted, ref} from 'vue';
import {showError} from 'src/lib/util';
import {fetchConversionLogs} from 'src/lib/get-api';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  conversionId: number;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

const postLoading = ref(false);
const logs = ref('');

const bracketsLineRegex = /^(\[.*?])/;
const mapLineRegex = /^(.*?)\s?: (.*?)$/;

function formatLogs(text: string): string {
  text = text.split('\n').map((line) => {
    line = line.trim();
    if (bracketsLineRegex.test(line)) {
      return line.replace(bracketsLineRegex, '<b><span style="color: deepskyblue">$1</span></b>')
    }
    if (mapLineRegex.test(line)) {
      return line.replace(mapLineRegex,
        '<i><span style="color: mediumpurple">$1</span></i> : <span style="color: mediumpurple">$2</span>')
    }
    if (line.startsWith('frame=')) {
      return `<span style="color: green">${line}</span>`
    }
    if (line.startsWith('Metadata:') || line.startsWith('Input') || line.startsWith('Stream')
      || line.startsWith('Output')) {
      return `<span style="color: greenyellow">${line}</span>`
    }
    return `<span style="color: gray">${line}</span>`;
  }).join('<br>');
  return text
}

onMounted(() => {
  postLoading.value = true;
  fetchConversionLogs(props.conversionId)
    .then((text) => {
      logs.value = formatLogs(text);
    })
    .catch((e) => {
      showError('Failed to fetch conversion logs', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
})
</script>
