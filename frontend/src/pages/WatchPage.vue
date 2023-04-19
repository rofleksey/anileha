<template>
  <q-page class="full-width" padding>
    <q-toolbar class="bg-purple text-white shadow-2 rounded-borders">
      <q-btn flat :label="data?.title ?? ''"/>
      <q-btn
        flat
        round
        icon="download"
        @click="onDownloadClick"/>
      <q-btn
        v-if="curUser && episodeId"
        flat
        round
        icon="group"
        @click="onGroupWatch"/>
      <q-btn
        v-if="curUser?.roles?.includes('admin')"
        flat
        round
        icon="image"
        @click="onRefreshThumbClick"/>
      <q-btn
        v-if="curUser?.roles?.includes('admin')"
        flat
        round
        icon="delete"
        @click="onDeleteClick"/>
    </q-toolbar>
    <VideoPlayer
      style="margin-top: 10px"
      :src="videoSrc"
      :poster="posterSrc"
    />
  </q-page>
</template>

<script setup lang="ts">
import {computed, ComputedRef, onMounted, ref, watch} from 'vue';
import {Episode, User} from 'src/lib/api-types';
import {BASE_URL, fetchEpisodeById} from 'src/lib/get-api';
import sanitize from 'sanitize-filename';
import {showError, showSuccess} from 'src/lib/util';
import {useRoute, useRouter} from 'vue-router';
import {useUserStore} from 'stores/user-store';
import VideoPlayer from 'components/VideoPlayer.vue';
import {deleteEpisode} from 'src/lib/delete-api';
import {useQuasar} from 'quasar';
import {refreshEpisodeThumb} from 'src/lib/post-api';
import {nanoid} from 'nanoid';
import {useRoomStore} from 'stores/room-store';

const quasar = useQuasar();
const router = useRouter();
const route = useRoute();
const episodeId = computed(() => Number(route.params.episodeId));

const userStore = useUserStore();
const curUser: ComputedRef<User | null> = computed(() => userStore.user);

const roomStore = useRoomStore();
const roomId = computed(() => roomStore.roomId);

const dataLoading = ref(false);
const data = ref<Episode | undefined>();

const videoSrc = computed(() => {
  const episode = data.value;
  if (!episode) {
    return '';
  }
  return `${BASE_URL}${episode.link}`
});

const posterSrc = computed(() => {
  const episode = data.value;
  if (!episode) {
    return '';
  }
  return `${BASE_URL}${episode.thumb}`
});

function onGroupWatch() {
  roomStore.setEpisodeId(episodeId.value);

  router.push({
    path: '/room',
    query: {
      id: roomId.value,
    }
  });
}

function onRefreshThumbClick() {
  quasar.dialog({
    title: 'Confirm',
    message: 'Do you really want to refresh thumbnail?',
    cancel: true,
  }).onOk(() => {
    refreshEpisodeThumb(episodeId.value)
      .then(() => {
        showSuccess('Episode thumbnail refreshed');
        router.back();
      })
      .catch((e) => {
        showError('failed to refresh episode thumbnail', e);
      });
  })
}

function onDownloadClick() {
  const link = document.createElement('a');
  link.href = videoSrc.value;
  link.setAttribute(
    'download',
    sanitize(`${data.value?.title ?? 'blank'}.mp4`)
  );
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

function onDeleteClick() {
  quasar.dialog({
    title: 'Confirm',
    message: 'Do you really want to delete this episode?',
    cancel: true,
  }).onOk(() => {
    deleteEpisode(episodeId.value)
      .then(() => {
        showSuccess('Episode deleted', `Successfully deleted episode ${data.value?.title ?? ''}`);
        router.push('/');
      })
      .catch((e) => {
        showError('failed to delete episode', e);
      })
  })
}

function refreshData() {
  dataLoading.value = true;
  fetchEpisodeById(episodeId.value)
    .then((newEpisode) => {
      data.value = newEpisode;
    })
    .catch((e) => {
      showError('failed to fetch episode', e);
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
