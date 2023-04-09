<template>
  <q-page class="full-width row items-start q-gutter-lg content-start justify-center" padding>
    <q-card
      class="episode-card"
      v-for="episode in data"
      :key="episode.id"
      @click="router.push(`/watch/${episode.id}`)">
      <q-img
        :src="`${BASE_URL}${episode.thumb}`"
        :ratio="1"
      />
      <q-card-section>
        <div class="text-subtitle2 ellipsis">{{ episode.episode }}</div>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import {computed, onMounted, ref} from 'vue';
import {Episode} from 'src/lib/api-types';
import {BASE_URL, fetchEpisodesBySeriesId} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import {useRoute, useRouter} from 'vue-router';

const router = useRouter();
const route = useRoute();
const seriesId = computed(() => Number(route.params.seriesId));

const dataLoading = ref(false);
const data = ref<Episode[]>([]);

function refreshData() {
  dataLoading.value = true;
  fetchEpisodesBySeriesId(seriesId.value)
    .then((newEpisodes) => {
      console.log(newEpisodes);
      data.value = newEpisodes;
    })
    .catch((e) => {
      showError('failed to fetch episodes', e);
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
.episode-card
  width: 200px
  cursor: pointer
  transition: opacity 0.3s ease

  &:hover
    opacity: 0.6
</style>
