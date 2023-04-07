<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Analysis</div>
      </q-card-section>
      <q-card-section class="q-pt-none">

      </q-card-section>
      <q-card-actions align="right">
        <q-btn
          color="accent"
          :loading="dataLoading"
          flat
          round
          icon="settings_backup_restore"
          @click="onOKClick"/>
      </q-card-actions>
      <q-inner-loading :showing="dataLoading">
        <q-spinner-gears size="50px" color="primary"/>
      </q-inner-loading>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import {useDialogPluginComponent} from 'quasar'
import {onMounted, ref} from 'vue';
import {postAnalyze} from 'src/lib/post-api';
import {showError} from 'src/lib/util';
import {Analysis} from 'src/lib/api-types';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  torrentId: number;
  fileIndices: number[];
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

const dataLoading = ref(false);
const data = ref<Analysis[]>([]);

function onOKClick() {
  console.log('ok');
}

async function loadAnalysis() {
  for (const index of props.fileIndices) {
    const analysis = await postAnalyze(props.torrentId, index);
    data.value.push(analysis);
  }
}

onMounted(() => {
  dataLoading.value = true;
  loadAnalysis()
    .then(() => {
      console.log(data.value);
    })
    .catch((e) => {
      showError('failed to process files', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
})
</script>

<style lang="sass" scoped>

</style>
