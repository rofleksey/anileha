<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Add torrent from file</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-toggle
          v-model="auto"
          label="Automatic Torrent"
        />
        <q-input
          ref="audioRef"
          v-model="audioLang"
          label="Audio Language (e.g. jpn)"
          :disable="!auto"
          :rules="auto ? [ val => val.trim().length === 3 || '3 chars expected' ] : []"/>
        <q-input
          ref="subRef"
          v-model="subLang"
          label="Subtitle Language (e.g. eng)"
          :disable="!auto"
          :rules="auto ? [ val => val.trim().length === 3 || '3 chars expected' ] : []"/>
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
import {postNewTorrentFromFile} from 'src/lib/post-api';
import {showError} from 'src/lib/util';
import {AutoTorrent} from 'src/lib/api-types';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  seriesId: number;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

const torrentFileRef = ref<any>(null);
const audioRef = ref<any>(null);
const subRef = ref<any>(null);

const postLoading = ref(false);
const auto = ref(false);
const audioLang = ref('jpn');
const subLang = ref('eng');
const torrentFile = ref<File | null>(null);

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  const file = torrentFile.value;
  if (!file || !torrentFileRef.value?.validate() || !audioRef.value?.validate() || !subRef.value?.validate()) {
    return
  }
  let autoTorrent: AutoTorrent | undefined;
  if (auto.value) {
    autoTorrent = {
      audioLang: audioLang.value,
      subLang: subLang.value,
    }
  }
  postLoading.value = true;
  postNewTorrentFromFile(props.seriesId, file, autoTorrent)
    .then(() => {
      onDialogOK();
    })
    .catch((e) => {
      showError('Failed to add torrent', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
