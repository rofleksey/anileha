<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Add torrent</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-file
          ref="torrentFileRef"
          v-model="torrentFile"
          label="Torrent File"
          accept=".torrent"
          max-file-size="8388608"
          :rules="[ val => val || 'Required' ]">
          <template v-slot:prepend>
            <q-icon name="attach_file"/>
          </template>
        </q-file>
      </q-card-section>
      <q-card-actions align="right">
        <q-btn
          color="accent"
          :loading="postLoading"
          flat
          round
          icon="add"
          @click="onOKClick"/>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import {useDialogPluginComponent} from 'quasar'
import {ref} from 'vue';
import {postNewTorrent} from 'src/lib/post-api';
import {showError} from 'src/lib/util';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  seriesId: number;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

const torrentFileRef = ref<any>(null);

const postLoading = ref(false);
const torrentFile = ref<File | null>(null);

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  const file = torrentFile.value;
  if (!file || !torrentFileRef.value?.validate()) {
    return
  }
  postLoading.value = true;
  postNewTorrent(props.seriesId, file)
    .then(() => {
      onDialogOK();
    })
    .catch((e) => {
      showError('failed to add torrent', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
