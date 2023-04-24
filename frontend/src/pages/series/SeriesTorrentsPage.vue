<template>
  <q-page class="full-width" padding>
    <TorrentTable :data="data" :loading="dataLoading"/>
    <q-page-sticky position="bottom-right" :offset="[18, 18]">
      <q-fab color="accent" icon="add" direction="up">
        <q-fab-action color="primary" @click="openNewTorrentFileModal" icon="file_upload" />
        <q-fab-action color="primary" @click="openNewTorrentSearchModal" icon="search" />
      </q-fab>
    </q-page-sticky>
  </q-page>
</template>

<script setup lang="ts">
import {computed, onMounted, ref} from 'vue';
import {Torrent} from 'src/lib/api-types';
import {fetchTorrentsBySeriesId} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import {useQuasar} from 'quasar';
import {useRoute} from 'vue-router';
import NewTorrentFileModal from 'components/modal/NewTorrentFileModal.vue';
import TorrentTable from 'components/TorrentTable.vue';
import {useInterval} from 'src/lib/composables';
import NewTorrentSearchModal from 'components/modal/NewTorrentSearchModal.vue';

const quasar = useQuasar();
const route = useRoute();
const seriesId = computed(() => Number(route.params.seriesId));

const dataLoading = ref(false);
const data = ref<Torrent[]>([]);

useInterval(refreshData, 10000);

function refreshData() {
  dataLoading.value = true;
  fetchTorrentsBySeriesId(seriesId.value)
    .then((newTorrents) => {
      data.value = newTorrents;
    })
    .catch((e) => {
      showError('Failed to fetch torrents', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
}

onMounted(() => {
  refreshData();
})

function openNewTorrentFileModal() {
  quasar.dialog({
    component: NewTorrentFileModal,
    componentProps: {
      seriesId: seriesId.value,
    }
  }).onOk(() => {
    refreshData();
  });
}

function openNewTorrentSearchModal() {
  quasar.dialog({
    component: NewTorrentSearchModal,
    componentProps: {
      seriesId: seriesId.value,
    }
  }).onOk(() => {
    refreshData();
  });
}
</script>

<style lang="sass" scoped>

</style>
