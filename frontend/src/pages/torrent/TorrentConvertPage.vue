<template>
  <q-table
    style="width: 100%"
    :rows="data?.files ?? []"
    :columns="columns"
    v-model:selected="selected"
    selection="multiple"
    row-key="id"
    :loading="dataLoading">
  </q-table>
  <q-page-sticky position="bottom-right" :offset="[18, 18]">
    <q-btn
      :disable="dataLoading || selected.length === 0"
      fab
      icon="settings"
      color="accent"
      @click="onAnalysisClick"/>
  </q-page-sticky>
</template>

<script setup lang="ts">
import {computed, onMounted, ref} from 'vue';
import {TorrentFile, TorrentWithFiles} from 'src/lib/api-types';
import {fetchTorrentById} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import {useRoute} from 'vue-router';
import AnalyzeModal from 'components/modal/AnalyzeModal.vue';
import {useQuasar} from 'quasar';

const quasar = useQuasar();
const route = useRoute();
const torrentId = computed(() => Number(route.params.torrentId));

const dataLoading = ref(false);
const data = ref<TorrentWithFiles | null>();
const selected = ref<TorrentFile[]>([]);

const columns: {
  name: string;
  label: string;
  field: string | ((row: any) => any);
  required?: boolean;
  align?: 'left' | 'right' | 'center';
  sortable?: boolean;
  sort?: (a: any, b: any, rowA: any, rowB: any) => number;
  sortOrder?: 'ad' | 'da';
  format?: (val: any, row: any) => any;
  style?: string | ((row: any) => string);
  classes?: string | ((row: any) => string);
  headerStyle?: string;
  headerClasses?: string;
}[] = [
  {
    name: 'path',
    label: 'Path',
    field: 'path',
    align: 'left',
    sortable: true,
  },
]

function onAnalysisClick() {
  quasar.dialog({
    component: AnalyzeModal,
    componentProps: {
      torrentId: torrentId.value,
      fileIndices: selected.value.map((it) => it.clientIndex),
    }
  }).onOk(() => {
    refreshData();
  });
}

function refreshData() {
  dataLoading.value = true;
  fetchTorrentById(torrentId.value)
    .then((newTorrent) => {
      console.log(newTorrent);
      data.value = newTorrent;
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
