<template>
  <q-stepper
    v-model="step"
    header-nav
    ref="stepper"
    color="primary"
    style="width: 100%"
    animated
  >
    <q-step
      :name="1"
      title="Select files to convert"
      icon="description"
      :done="step > 1"
      :header-nav="step > 1"
    >
      <q-table
        style="width: 100%"
        :rows="readyFiles"
        :columns="selectFilesColumns"
        v-model:selected="selectedForConversion"
        selection="multiple"
        row-key="clientIndex"
        :loading="dataLoading"
        :pagination="{rowsPerPage: 10}">
      </q-table>

      <q-stepper-navigation>
        <q-btn
          :disable="dataLoading || selectedForConversion.length === 0"
          @click="() => { step = 2 }"
          color="primary"
          label="Continue"/>
      </q-stepper-navigation>
    </q-step>

    <q-step
      :name="2"
      title="Set options"
      icon="settings"
      :done="step > 2"
      :header-nav="step > 2"
    >
      <q-table
        style="width: 100%"
        :rows="analysisData"
        :columns="analysisColumns"
        row-key="id"
        :loading="dataLoading">
        <template v-slot:body-cell-sub="props">
          <q-td :props="props">
            <template v-if="props.value.stream !== undefined">
              #{{ props.value.stream }} -
              {{ formatName(props.row.analysis.sub.find((it) => it.index === props.value.stream).name) }}
            </template>
            <template v-else-if="props.value.file !== undefined">
              file: {{ props.value.file }}
            </template>
            <template v-else>
              disabled
            </template>
            <q-btn
              @click="openSubStreamPickModal(props.row.clientIndex, props.row.analysis.sub, props.value.stream, props.value.file)"
              flat
              round
              color="orange"
              size="sm"
              icon="settings"/>
          </q-td>
        </template>
        <template v-slot:body-cell-audio="props">
          <q-td :props="props">
            <template v-if="props.value.stream !== undefined">
              #{{ props.value.stream }} -
              {{ formatName(props.row.analysis.audio.find((it) => it.index === props.value.stream).name) }}
            </template>
            <template v-else>
              disabled
            </template>
            <q-btn
              @click="openAudioStreamPickModal(props.row.clientIndex, props.row.analysis.audio, props.value.stream)"
              flat
              round
              color="orange"
              size="sm"
              icon="settings"/>
          </q-td>
        </template>
        <template v-slot:body-cell-meta="props">
          <q-td :props="props">
            {{ props.value }}
            <q-btn
              @click="openChangeMetadataModal(props.row.clientIndex, props.row.prefs.episode, props.row.prefs.season)"
              flat
              round
              color="orange"
              size="sm"
              icon="settings"/>
          </q-td>
        </template>
      </q-table>

      <q-stepper-navigation>
        <q-btn
          flat
          :disable="dataLoading"
          @click="step = 1"
          color="primary"
          label="Back"
          class="q-ml-sm"/>
        <q-btn
          @click="startConversion"
          :loading="dataLoading"
          :disable="dataLoading"
          color="primary"
          label="Convert"/>
      </q-stepper-navigation>
    </q-step>
  </q-stepper>
</template>

<script setup lang="ts">
import {computed, onMounted, ref, watch} from 'vue';
import {
  Analysis,
  BaseStream,
  ConversionPreference,
  StartConversionFileData,
  SubStream,
  TorrentFile,
  TorrentWithFiles
} from 'src/lib/api-types';
import {fetchTorrentById} from 'src/lib/get-api';
import {QuasarColumnType, showError, showSuccess} from 'src/lib/util';
import {useRoute} from 'vue-router';
import {useQuasar} from 'quasar';
import {postAnalyze, postStartConversion} from 'src/lib/post-api';
import PickAudioStreamModal from 'components/modal/PickAudioStreamModal.vue';
import ChangeMetadataModal from 'components/modal/ChangeMetadataModal.vue';
import PickSubtitleStreamModal from 'components/modal/PickSubtitleStreamModal.vue';

const quasar = useQuasar();
const route = useRoute();
const torrentId = computed(() => Number(route.params.torrentId));

interface AnalysisWithPrefs {
  path: string;
  clientIndex: number;
  analysis: Analysis;
  prefs: StartConversionFileData;
}

const step = ref(1);
const dataLoading = ref(false);
const torrentData = ref<TorrentWithFiles | null>();
const analysisData = ref<AnalysisWithPrefs[]>([]);
const selectedForConversion = ref<TorrentFile[]>([]);

function formatName(name: string) {
  if (name.trim().length === 0) {
    return '<blank>';
  }
  return name;
}

function pickSubStream(streams: SubStream[], langPref: string | null): ConversionPreference {
  if (streams.length === 0) {
    return {
      disable: true,
    }
  }
  if (langPref) {
    const langStreams = streams.filter((s) => s.lang === langPref);
    if (langStreams.length !== 0) {
      streams = langStreams;
    }
  }
  const pictureSubs = streams.filter((s) => s.textLength < 32)
    .sort((a, b) => a.size - b.size);
  if (pictureSubs.length != 0) {
    return {
      stream: pictureSubs[pictureSubs.length - 1].index
    }
  }
  const textSubs = streams.filter((s) => s.textLength > 32)
    .sort((a, b) => a.textLength - b.textLength);
  return {
    stream: textSubs[textSubs.length - 1].index
  }
}

function pickAudioStream(streams: BaseStream[], langPref: string | null): ConversionPreference {
  if (streams.length === 0) {
    return {
      disable: true,
    }
  }
  if (langPref) {
    const langStreams = streams.filter((s) => s.lang === langPref);
    if (langStreams.length !== 0) {
      streams = langStreams;
    }
  }
  const sorted = streams.sort((a, b) => a.size - b.size);
  return {
    stream: sorted[sorted.length - 1].index
  };
}

async function loadAnalysis(): Promise<AnalysisWithPrefs[]> {
  const files = selectedForConversion.value;
  const result: AnalysisWithPrefs[] = [];
  for (const file of files) {
    const analysis = await postAnalyze(torrentId.value, file.clientIndex);
    result.push({
      path: file.path,
      clientIndex: file.clientIndex,
      analysis: analysis,
      prefs: {
        index: file.clientIndex,
        sub: pickSubStream(analysis.sub, 'eng'),
        audio: pickAudioStream(analysis.audio, 'jpn'),
        season: analysis.season,
        episode: analysis.episode
      }
    });
  }
  return result;
}

function startAnalysis() {
  dataLoading.value = true;
  quasar.loading.show({
    delay: 100,
  });
  loadAnalysis()
    .then((newData) => {
      console.log(newData);
      analysisData.value = newData;
    })
    .catch((e) => {
      showError('failed to process files', e);
    })
    .finally(() => {
      dataLoading.value = false;
      quasar.loading.hide();
    });
}

function startConversion() {
  dataLoading.value = true;
  postStartConversion({
    torrentId: torrentId.value,
    files: analysisData.value.map((it) => it.prefs),
  })
    .then(() => {
      showSuccess('Conversion started')
    })
    .catch((e) => {
      showError('failed to start conversion', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
}

watch(step, () => {
  const curStep = step.value;
  if (curStep === 2) {
    startAnalysis();
  }
})

const readyFiles = computed(() => {
  const torrent = torrentData.value;
  if (!torrent) {
    return [];
  }
  return torrent.files.filter((file) => file.status === 'ready');
});

const externalSubtitleFiles = computed(() => {
  return readyFiles.value.filter((file) => {
    const lowerCase = file.path.toLowerCase();
    return lowerCase.endsWith('.srt') || lowerCase.endsWith('.ssa') || lowerCase.endsWith('.ass');
  }).map((file) => file.path);
});

const selectFilesColumns: QuasarColumnType[] = [
  {
    name: 'path',
    label: 'Path',
    field: 'path',
    align: 'left',
    sortable: true,
  },
]

const analysisColumns: QuasarColumnType[] = [
  {
    name: 'path',
    label: 'Path',
    field: 'path',
    align: 'left',
    sortable: true,
  },
  {
    name: 'audio',
    label: 'Audio',
    field: (obj: AnalysisWithPrefs) => obj.prefs.audio,
    align: 'left',
  },
  {
    name: 'sub',
    label: 'Subtitles',
    field: (obj: AnalysisWithPrefs) => obj.prefs.sub,
    align: 'left',
  },
  {
    name: 'meta',
    label: 'Meta',
    field: (obj: AnalysisWithPrefs) => `E: ${obj.prefs.episode}, S: ${obj.prefs.season}`,
    align: 'left',
  },
]

function openAudioStreamPickModal(fileIndex: number, streams: BaseStream[], curIndex: number) {
  quasar.dialog({
    component: PickAudioStreamModal,
    componentProps: {
      streams,
      curIndex,
    }
  }).onOk(({stream, file}: { stream: number | undefined, file: string | undefined }) => {
    const analysisForFile = analysisData.value.find((it) => it.clientIndex === fileIndex);
    if (!analysisForFile) {
      return
    }
    if (stream !== undefined) {
      analysisForFile.prefs.audio.stream = stream;
    }
  });
}

function openSubStreamPickModal(fileIndex: number, streams: SubStream[], curIndex: number | undefined,
                                curFile: string | undefined) {
  quasar.dialog({
    component: PickSubtitleStreamModal,
    componentProps: {
      streams,
      curIndex,
      files: externalSubtitleFiles.value,
      curFile
    }
  }).onOk(({stream, file}: { stream: number | undefined, file: string | undefined }) => {
    const analysisForFile = analysisData.value.find((it) => it.clientIndex === fileIndex);
    if (!analysisForFile) {
      return
    }
    if (stream !== undefined) {
      analysisForFile.prefs.sub.stream = stream;
      analysisForFile.prefs.sub.file = undefined;
    } else if (file !== undefined) {
      analysisForFile.prefs.sub.file = file;
      analysisForFile.prefs.sub.stream = undefined;
    }
  });
}

function openChangeMetadataModal(fileIndex: number, curEpisode: string, curSeason: string) {
  quasar.dialog({
    component: ChangeMetadataModal,
    componentProps: {
      curEpisode,
      curSeason,
    }
  }).onOk(({episode, season}: { episode: string, season: string }) => {
    const analysisForFile = analysisData.value.find((it) => it.clientIndex === fileIndex);
    if (!analysisForFile) {
      return
    }
    analysisForFile.prefs.season = season;
    analysisForFile.prefs.episode = episode;
  });
}

function refreshData() {
  dataLoading.value = true;
  fetchTorrentById(torrentId.value)
    .then((newTorrent) => {
      console.log(newTorrent);
      torrentData.value = newTorrent;
    })
    .catch((e) => {
      showError('failed to fetch torrent', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
}

onMounted(() => {
  refreshData();
})
</script>

<style lang="sass" scoped>

</style>
