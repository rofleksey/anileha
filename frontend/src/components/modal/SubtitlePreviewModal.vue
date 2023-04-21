<template>
  <q-dialog full-width ref="dialogRef" @hide="onDialogHide" @keyup.enter="onDialogOK">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Subtitle Text</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-editor
          readonly
          v-model="subText"
          min-height="5rem" />
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
import {postSubText} from 'src/lib/post-api';
import {showError} from 'src/lib/util';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  torrentId: number;
  fileIndex: number;
  stream: number;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

const postLoading = ref(false);
const subText = ref('');

function formatSubs(text: string): string {
  return text.replaceAll('\n', '<br>');
}

onMounted(() => {
  postLoading.value = true;
  postSubText(props.torrentId, props.fileIndex, props.stream)
    .then((text) => {
      console.log(text);
      subText.value = formatSubs(text);
    })
    .catch((e) => {
      showError('Failed to fetch subtitle text', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
})
</script>
