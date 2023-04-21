<template>
  <q-page class="full-width" padding>
    <q-tree class="col-12 col-sm-6"
            :nodes="nodes"
            node-key="id"
            tick-strategy="leaf"
            v-model:selected="selection"
            v-model:ticked="ticked"
            v-model:expanded="expanded"
    />
    <q-page-sticky position="bottom-right" :offset="[18, 18]">
      <q-btn
        v-if="data?.status !== 'download'"
        :disable="dataLoading || ticked.length === 0 || data?.status === 'analysis'"
        fab
        icon="play_arrow"
        color="accent"
        @click="onStartClick"/>
      <q-btn
        v-else
        :disable="dataLoading"
        fab
        icon="stop"
        color="accent"
        @click="onStopClick"/>
    </q-page-sticky>
  </q-page>
</template>

<script setup lang="ts">
import {computed, onMounted, ref} from 'vue';
import {TorrentWithFiles} from 'src/lib/api-types';
import {fetchTorrentById} from 'src/lib/get-api';
import {showError, showSuccess} from 'src/lib/util';
import {useRoute} from 'vue-router';
import {postStartTorrent, postStopTorrent} from 'src/lib/post-api';

const route = useRoute();
const torrentId = computed(() => Number(route.params.torrentId));

interface Node {
  id: number;
  label: string;
  children: Node[];
}

const dataLoading = ref(false);
const data = ref<TorrentWithFiles | null>();

const selection = ref<number[]>([]);
const ticked = ref<number[]>([]);
const expanded = ref<number[]>([]);

const nodes = computed(() => {
  const torrent = data.value;
  if (!torrent) {
    return [];
  }
  const rootNodes: Node[] = [];
  let miscCounter = -1;
  torrent.files.forEach((file) => {
    const pathSplit = file.path.split('/');
    let curNodeArray = rootNodes;
    pathSplit.forEach((pathItem, index) => {
      let nodeItem = curNodeArray.find((arrayItem) => arrayItem.label === pathItem);
      if (!nodeItem) {
        const newNode = {
          id: (index === pathSplit.length - 1) ? file.clientIndex : miscCounter--,
          label: pathItem,
          disabled: torrent.status === 'download' || torrent.status === 'analysis',
          children: [],
        }
        curNodeArray.push(newNode);
        nodeItem = newNode;
      }
      curNodeArray = nodeItem.children;
    });
  });
  return rootNodes;
});

function refreshData() {
  dataLoading.value = true;
  fetchTorrentById(torrentId.value)
    .then((newTorrent) => {
      data.value = newTorrent;
      ticked.value = newTorrent.files
        .filter((file) => file.selected)
        .map((file) => file.clientIndex);
    })
    .catch((e) => {
      showError('failed to fetch torrent', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
}

function onStopClick() {
  dataLoading.value = true;
  postStopTorrent(torrentId.value)
    .then(() => {
      showSuccess('Torrent stopped');
      refreshData();
    })
    .catch((e) => {
      showError('failed to stop torrent', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
}

function onStartClick() {
  dataLoading.value = true;
  postStartTorrent(torrentId.value, ticked.value)
    .then(() => {
      showSuccess('Torrent started');
      refreshData();
    })
    .catch((e) => {
      showError('failed to start torrent', e);
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
