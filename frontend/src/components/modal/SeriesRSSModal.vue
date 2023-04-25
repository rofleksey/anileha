<template>
  <q-dialog ref="dialogRef" @hide="onDialogHide">
    <q-card class="q-dialog-plugin card">
      <q-card-section>
        <div class="text-h6">Series RSS settings</div>
      </q-card-section>
      <q-card-section class="q-pt-none">
        <q-toggle
          v-model="rssEnabled"
          label="Enable RSS Scan"
        />
        <q-input
          ref="providerRef"
          v-model="providerStr"
          label="Provider (e.g. nyaa)"
          :disable="!rssEnabled"
          :rules="rssEnabled ? [ val => val.trim().length > 0 || 'Required' ] : []"/>
        <q-input
          ref="includeRef"
          v-model="includeStr"
          label="Include words (separated by space)"
          :disable="!rssEnabled"
          :rules="rssEnabled ? [ val => val.trim().length > 0 || 'Required' ] : []"/>
        <q-input
          ref="excludeRef"
          v-model="excludeStr"
          label="Exclude words (separated by space)"
          :disable="!rssEnabled"/>
        <q-toggle
          v-model="singleFile"
          :disable="!rssEnabled"
          label="Accept torrents with single file only"
        />
        <q-input
          ref="audioRef"
          v-model="audioLang"
          label="Audio Language (e.g. jpn)"
          :disable="!rssEnabled"
          :rules="rssEnabled ? [ val => val.trim().length === 3 || '3 chars expected' ] : []"/>
        <q-input
          ref="subRef"
          v-model="subLang"
          label="Subtitle Language (e.g. eng)"
          :disable="!rssEnabled"
          :rules="rssEnabled ? [ val => val.trim().length === 3 || '3 chars expected' ] : []"/>
      </q-card-section>
      <q-card-actions align="right">
        <q-btn
          color="accent"
          :loading="postLoading"
          flat
          round
          icon="downloading"
          :disable="!rssEnabled"
          @click="onOldClick"/>
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
import {postAddTorrentsFromQuery, postSetSeriesQuery} from 'src/lib/post-api';
import {showError, showSuccess} from 'src/lib/util';
import {AutoTorrent, SeriesQueryServer, SetSeriesQueryRequestData} from 'src/lib/api-types';

const {dialogRef, onDialogHide, onDialogOK} = useDialogPluginComponent()

interface Props {
  seriesId: number;
  query: SeriesQueryServer | null;
}

const props = defineProps<Props>()

defineEmits([
  ...useDialogPluginComponent.emits
])

const audioRef = ref<any>(null);
const subRef = ref<any>(null);
const providerRef = ref<any>(null);
const includeRef = ref<any>(null);
const excludeRef = ref<any>(null);

const postLoading = ref(false);
const rssEnabled = ref(!!props.query);
const providerStr = ref(props.query?.provider ?? 'nyaa');
const includeStr = ref(props.query?.include?.join(' ') ?? '');
const excludeStr = ref(props.query?.exclude?.join(' ') ?? '');
const singleFile = ref(props.query?.singleFile ?? true);
const audioLang = ref(props.query?.auto?.audioLang ?? 'jpn');
const subLang = ref(props.query?.auto?.subLang ?? 'eng');

function onOldClick() {
  if (postLoading.value) {
    return;
  }
  if (!rssEnabled.value || !audioRef.value?.validate() || !subRef.value?.validate() || !providerRef.value?.validate()
    || !includeRef.value?.validate()) {
    return
  }
  const autoTorrent: AutoTorrent = {
    audioLang: audioLang.value,
    subLang: subLang.value,
  }
  const data: SetSeriesQueryRequestData = {
    provider: providerStr.value.trim(),
    include: includeStr.value.trim(),
    exclude: excludeStr.value.trim(),
    singleFile: singleFile.value,
    auto: autoTorrent,
  }
  postLoading.value = true;
  postAddTorrentsFromQuery(props.seriesId, data)
    .then(() => {
      showSuccess('Torrents added')
      onDialogOK();
    })
    .catch((e) => {
      showError('Failed to add torrents from query', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}

function onOKClick() {
  if (postLoading.value) {
    return;
  }
  if (!audioRef.value?.validate() || !subRef.value?.validate() || !providerRef.value?.validate()
    || !includeRef.value?.validate()) {
    return
  }
  if (!rssEnabled.value) {
    postLoading.value = true;
    postSetSeriesQuery(props.seriesId, null)
      .then(() => {
        showSuccess('Series query disabled')
        onDialogOK();
      })
      .catch((e) => {
        showError('Failed to disable series query', e);
      })
      .finally(() => {
        postLoading.value = false;
      });
    return;
  }
  const autoTorrent: AutoTorrent = {
    audioLang: audioLang.value,
    subLang: subLang.value,
  }
  const data: SetSeriesQueryRequestData = {
    provider: providerStr.value.trim(),
    include: includeStr.value.trim(),
    exclude: excludeStr.value.trim(),
    singleFile: singleFile.value,
    auto: autoTorrent,
  }
  postLoading.value = true;
  postSetSeriesQuery(props.seriesId, data)
    .then(() => {
      showSuccess('Series query applied')
      onDialogOK();
    })
    .catch((e) => {
      showError('Failed to disable series query', e);
    })
    .finally(() => {
      postLoading.value = false;
    });
}
</script>

<style lang="sass" scoped>

</style>
