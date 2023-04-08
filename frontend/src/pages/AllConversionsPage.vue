<template>
  <q-page class="full-width" padding>
    <ConversionTable
      title="All conversions"
      :data="data"
      :loading="dataLoading"/>
  </q-page>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {Conversion} from 'src/lib/api-types';
import {fetchAllConversions, fetchAllTorrents} from 'src/lib/get-api';
import {showError} from 'src/lib/util';
import ConversionTable from 'components/ConversionTable.vue';

const dataLoading = ref(false);
const data = ref<Conversion[]>([]);

function refreshData() {
  dataLoading.value = true;
  fetchAllConversions()
    .then((newConversions) => {
      console.log(newConversions);
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
