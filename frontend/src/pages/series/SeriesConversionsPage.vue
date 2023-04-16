<template>
  <q-page class="full-width" padding>
    <ConversionTable :data="data" :loading="dataLoading"/>
  </q-page>
</template>

<script setup lang="ts">
import {computed, onMounted, ref} from 'vue';
import {Conversion} from 'src/lib/api-types';
import {fetchConversionsBySeriesId} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import {useRoute} from 'vue-router';
import ConversionTable from 'components/ConversionTable.vue';
import {useInterval} from 'src/lib/composables';

const route = useRoute();
const seriesId = computed(() => Number(route.params.seriesId));

const dataLoading = ref(false);
const data = ref<Conversion[]>([]);

useInterval(refreshData, 10000);

function refreshData() {
  dataLoading.value = true;
  fetchConversionsBySeriesId(seriesId.value)
    .then((newConversions) => {
      data.value = newConversions;
    })
    .catch((e) => {
      showError('failed to fetch conversions', e);
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
