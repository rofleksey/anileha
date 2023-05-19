<template>
  <q-page class="full-width row items-start q-gutter-lg content-start justify-center" padding>
    <q-card
      class="series-card"
      v-for="series in data"
      :key="series.id"
      @click="router.push(`/watch/${series.episodeId}`)">
      <q-img
        :src="series.thumb"
        :ratio="1"
      />
      <q-badge color="green" floating>{{ series.episodeName }}</q-badge>
      <q-card-section>
        <div class="text-subtitle2 ellipsis">{{ series.title }}</div>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {Series} from 'src/lib/api-types';
import {BASE_URL, fetchAllSeries} from 'src/lib/get-api';
import {loadWatchLater, showError, WatchLaterItem} from 'src/lib/util';
import {useQuasar} from 'quasar';
import CreateSeriesModal from 'components/modal/CreateSeriesModal.vue';
import {useRouter} from 'vue-router';
import {useInterval} from 'src/lib/composables';

const quasar = useQuasar();
const router = useRouter();

const dataLoading = ref(false);
const data = ref<WatchLaterView[]>([]);

useInterval(refreshData, 10000);

interface WatchLaterView extends Series {
  episodeId: number;
  episodeName: string;
}

function refreshData() {
  dataLoading.value = true;
  fetchAllSeries()
    .then((newSeries) => {
      const watchLater = loadWatchLater();
      const watchLaterMap: { [key: number]: WatchLaterItem } = {};
      watchLater.forEach((it) => {
        watchLaterMap[it.seriesId] = it;
      })
      data.value = newSeries
        .filter((series) => watchLaterMap[series.id])
        .map((series) => ({
          ...series,
          thumb: `${BASE_URL}${series.thumb}`,
          episodeName: watchLaterMap[series.id].episodeName,
          episodeId: watchLaterMap[series.id].episodeId
        }));
    })
    .catch((e) => {
      showError('failed to fetch series', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
}

onMounted(() => {
  refreshData();
})

function openCreateSeriesModal() {
  quasar.dialog({
    component: CreateSeriesModal,
  }).onOk(() => {
    refreshData();
  });
}
</script>

<style lang="sass" scoped>
.series-card
  width: 200px
  cursor: pointer
  transition: opacity 0.3s ease

  &:hover
    opacity: 0.6
</style>
