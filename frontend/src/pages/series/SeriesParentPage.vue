<template>
  <q-page class="row items-start q-gutter-lg content-start justify-center" padding>
    <q-toolbar class="bg-purple text-white shadow-2 rounded-borders">
      <q-btn flat :label="title"/>
      <q-btn
        v-if="curUser?.roles?.includes('admin')"
        flat
        round
        icon="delete"
        @click="onDeleteClick"/>
      <q-btn
        v-if="curUser?.roles?.includes('admin') && tabName === 'episodes'"
        flat
        round
        icon="upload"
        @click="openUploadModal"/>
      <q-space/>
      <q-tabs :model-value="tabName" @update:model-value="onTabChange" shrink>
        <q-tab name="episodes" label="Episodes"/>
        <q-tab v-if="curUser?.roles?.includes('admin')" name="torrents" label="Torrents"/>
        <q-tab v-if="curUser?.roles?.includes('admin')" name="conversions" label="Conversions"/>
      </q-tabs>
    </q-toolbar>
    <router-view></router-view>
  </q-page>
</template>

<script setup lang="ts">
import {computed, ComputedRef, onMounted, ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {fetchSeriesById} from 'src/lib/get-api';
import {showError, showSuccess} from 'src/lib/util';
import {useQuasar} from 'quasar';
import {deleteSeries} from 'src/lib/delete-api';
import {useUserStore} from 'stores/user-store';
import {User} from 'src/lib/api-types';
import NewEpisodeModal from 'components/modal/NewEpisodeModal.vue';

const quasar = useQuasar();
const router = useRouter();
const route = useRoute();
const seriesId = computed(() => Number(route.params.seriesId));
const tabName = computed(() => route.name?.toString().replace('series-', ''));

const userStore = useUserStore();
const curUser: ComputedRef<User | null> = computed(() => userStore.user);

const title = ref('');

function onTabChange(value: string) {
  router.replace(`/series/${seriesId.value}/${value}`)
}

function openUploadModal() {
  quasar.dialog({
    component: NewEpisodeModal,
    componentProps: {
      seriesId: seriesId.value,
    }
  });
}

function onDeleteClick() {
  quasar.dialog({
    title: 'Confirm',
    message: 'Do you really want to delete this series?',
    cancel: true,
  }).onOk(() => {
    deleteSeries(seriesId.value)
      .then(() => {
        showSuccess('Series deleted', `Successfully deleted series ${title.value}`);
        router.push('/series');
      })
      .catch((e) => {
        showError('failed to delete series', e);
      })
  })
}

function reloadTitle() {
  fetchSeriesById(seriesId.value)
    .then((series) => {
      title.value = series.title;
    })
    .catch((e) => {
      showError('failed to fetch series', e);
    })
}

onMounted(() => {
  reloadTitle();
})
</script>

<style lang="sass" scoped>
</style>
