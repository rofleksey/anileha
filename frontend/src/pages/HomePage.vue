<template>
  <q-page class="full-width row items-start q-gutter-lg content-start justify-center" padding>
    <q-pagination
      v-if="maxPages > 0"
      v-model="page"
      color="purple"
      :max="maxPages"
      :max-pages="6"
      boundary-numbers
    />
    <q-card
      class="episode-card"
      v-for="episode in data"
      :key="episode.id"
      @click="router.push(`/watch/${episode.id}`)">
      <q-img
        :src="episode.thumb"
        :ratio="1"
      />
      <q-card-section>
        <div class="text-subtitle2 ellipsis">{{ episode.title }}</div>
      </q-card-section>
    </q-card>
    <q-pagination
      v-if="maxPages > 0"
      v-model="page"
      color="purple"
      :max="maxPages"
      :max-pages="6"
      boundary-numbers
    />
  </q-page>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {Episode} from 'src/lib/api-types';
import {fetchEpisodes} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import {useQuasar} from 'quasar';
import {useRouter} from 'vue-router';

const quasar = useQuasar();
const router = useRouter();

const dataLoading = ref(false);
const data = ref<Episode[]>([]);
const page = ref(1);
const maxPages = ref(1);

function refreshData() {
  dataLoading.value = true;
  fetchEpisodes(page.value - 1)
    .then((newData) => {
      data.value = newData.episodes;
      maxPages.value = Math.max(1, newData.maxPages)
    })
    .catch((e) => {
      showError('Failed to fetch episodes', e);
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
