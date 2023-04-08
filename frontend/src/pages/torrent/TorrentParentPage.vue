<template>
  <q-page class="row items-start q-gutter-lg content-start justify-center" padding>
    <q-toolbar class="bg-purple text-white shadow-2 rounded-borders">
      <q-btn flat :label="title"/>
      <q-btn
        v-if="data?.status !== 'download'"
        flat
        round
        icon="delete"
        @click="onDeleteClick"/>
      <q-space/>
      <q-tabs :model-value="tabName" @update:model-value="onTabChange" shrink>
        <q-tab name="download" label="Download"/>
        <q-tab name="convert" label="Convert"/>
      </q-tabs>
    </q-toolbar>
    <router-view></router-view>
  </q-page>
</template>

<script setup lang="ts">
import {computed, onMounted, ref, watch} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {fetchTorrentById} from 'src/lib/get-api';
import {showError, showSuccess} from 'src/lib/util';
import {useQuasar} from 'quasar';
import {deleteTorrent} from 'src/lib/delete-api';
import {TorrentWithFiles} from 'src/lib/api-types';

const router = useRouter();
const route = useRoute();
const torrentId = computed(() => Number(route.params.torrentId));
const tabName = computed(() => route.name?.toString().replace('torrent-', ''));

const quasar = useQuasar();

const data = ref<TorrentWithFiles | undefined>();
const title = computed(() => data.value?.name ?? '');

function onTabChange(value: string) {
  router.replace(`/torrent/${torrentId.value}/${value}`)
}

function onDeleteClick() {
  quasar.dialog({
    title: 'Confirm',
    message: 'Do you really want to delete this torrent?',
    cancel: true,
  }).onOk(() => {
    deleteTorrent(torrentId.value)
      .then(() => {
        showSuccess('Torrent deleted', `Successfully deleted torrent ${title.value}`);
        router.push('/torrents');
      })
      .catch((e) => {
        showError('failed to delete torrent', e);
      })
  })
}

watch(torrentId, reloadData);

function reloadData() {
  fetchTorrentById(torrentId.value)
    .then((newTorrent) => {
      data.value = newTorrent;
    })
    .catch((e) => {
      showError('failed to fetch torrent', e);
    })
}

onMounted(() => {
  reloadData();
})
</script>

<style lang="sass" scoped>
</style>
