<template>
  <q-page class="full-width" padding>
    <TorrentTable
      title="All torrents"
      :data="data"
      :loading="dataLoading"/>
  </q-page>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {Torrent} from 'src/lib/api-types';
import {fetchAllTorrents} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import TorrentTable from 'components/TorrentTable.vue';

const dataLoading = ref(false);
const data = ref<Torrent[]>([]);

function refreshData() {
  dataLoading.value = true;
  fetchAllTorrents()
    .then((newTorrents) => {
      data.value = newTorrents;
    })
    .catch((e) => {
      showError('failed to fetch torrents', e);
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
