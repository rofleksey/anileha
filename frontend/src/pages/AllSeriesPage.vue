<template>
  <q-page class="full-width row items-start q-gutter-lg content-start justify-center" padding>
    <q-card
      class="series-card"
      v-for="series in data"
      :key="series.id"
      @click="router.push(`/series/${series.id}/episodes`)">
      <q-img
        :src="series.thumb"
        :ratio="1"
      />
      <q-card-section>
        <div class="text-subtitle2 ellipsis">{{ series.title }}</div>
      </q-card-section>
    </q-card>
    <q-page-sticky position="bottom-right" :offset="[18, 18]" v-if="curUser?.roles?.includes('admin')">
      <q-btn fab icon="add" color="accent" @click="openCreateSeriesModal"/>
    </q-page-sticky>
  </q-page>
</template>

<script setup lang="ts">
import {computed, ComputedRef, onMounted, ref} from 'vue';
import {Series, User} from 'src/lib/api-types';
import {BASE_URL, fetchAllSeries} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import {useQuasar} from 'quasar';
import CreateSeriesModal from 'components/modal/CreateSeriesModal.vue';
import {useRouter} from 'vue-router';
import {useUserStore} from 'stores/user-store';
import {useInterval} from 'src/lib/composables';

const quasar = useQuasar();
const router = useRouter();

const userStore = useUserStore();
const curUser: ComputedRef<User | null> = computed(() => userStore.user);

const dataLoading = ref(false);
const data = ref<Series[]>([]);

useInterval(refreshData, 10000);

function refreshData() {
  dataLoading.value = true;
  fetchAllSeries()
    .then((newSeries) => {
      data.value = newSeries.map((series) => ({
        ...series,
        thumb: `${BASE_URL}${series.thumb}`
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
