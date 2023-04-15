<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide" @keyup.enter="onOKClick">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Upload episode</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-input
          ref="titleRef"
          v-model="title"
          label="Title"
          :rules="[ val => val.trim().length > 0 || 'Required' ]"/>
        <q-input
          ref="seasonRef"
          v-model="season"
          label="Season"
          clearable />
        <q-input
          ref="episodeRef"
          v-model="episode"
          label="Episode"
          clearable />
        <q-file
          ref="fileRef"
          v-model="file"
          label="Video File"
          accept="video/mp4"
          :rules="[ val => val || 'Required' ]">
          <template v-slot:prepend>
            <q-icon name="attach_file"/>
          </template>
        </q-file>
      </q-card-section>
      <q-inner-loading :showing="postLoading">
        <slot>
          <q-circular-progress
            show-value
            style="margin: 0"
            class="text-light-blue q-ma-md"
            :value="Math.floor(100 * postProgress)"
            track-color="grey-9"
            size="xl"
            color="light-blue"
          />
        </slot>
      </q-inner-loading>
      <q-card-actions align="right">
        <q-btn
          color="accent"
          :loading="postLoading"
          flat
          round
          icon="done"
          @click="onOKClick"/>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import {useDialogPluginComponent} from 'quasar'
import {ref} from 'vue';
import {postNewEpisode} from 'src/lib/post-api';
import {showError, showSuccess} from 'src/lib/util';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  seriesId: number | null;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

const titleRef = ref<any>(null);
const seasonRef = ref<any>(null);
const episodeRef = ref<any>(null);
const fileRef = ref<any>(null);

const postLoading = ref(false);
const postProgress = ref(0);
const title = ref('');
const season = ref('');
const episode = ref('');
const file = ref<File | null>(null);

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  if (!titleRef.value?.validate() || !fileRef.value?.validate() || !file.value) {
    return
  }
  postLoading.value = true;
  postNewEpisode(props.seriesId, file.value, title.value, season.value, episode.value, (e) => {
    if (e.progress) {
      postProgress.value = e.progress;
    } else if (e.total) {
      postProgress.value = e.loaded / e.total;
    }
  })
    .then(() => {
      showSuccess('Episode uploaded');
      onDialogOK();
    })
    .catch((e) => {
      showError('Failed to upload episode', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
