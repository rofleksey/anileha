<template>
  <q-page class="full-width" padding>
    <q-toolbar class="bg-purple text-white shadow-2 rounded-borders">
      <q-btn flat :label="episodeData?.title ?? ''"/>
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
      ref="playerRef"
      style="margin-top: 10px"
      :src="videoSrc"
      :poster="posterSrc"
      @canplay.once="onCanPlay"
    />
  </q-page>
</template>

<script setup lang="ts">
import {computed, ComputedRef, onMounted, ref} from 'vue';
import {Episode, User} from 'src/lib/api-types';
import {BASE_URL, fetchEpisodeById, fetchEpisodesBySeriesId} from 'src/lib/get-api';
import sanitize from 'sanitize-filename';
import {saveWatchLater, showError, showSuccess} from 'src/lib/util';
import {useRoute, useRouter} from 'vue-router';
import {useUserStore} from 'stores/user-store';
import VideoPlayer from 'components/VideoPlayer.vue';
import {deleteEpisode} from 'src/lib/delete-api';
import {useQuasar} from 'quasar';
import {refreshEpisodeThumb} from 'src/lib/post-api';
import {useInterval} from 'src/lib/composables';
import {nanoid} from 'nanoid';

const quasar = useQuasar();
const router = useRouter();
const route = useRoute();

const playerRef = ref<any>();

const episodeId = computed(() => Number(route.params.episodeId));

const userStore = useUserStore();
const curUser: ComputedRef<User | null> = computed(() => userStore.user);

const dataLoading = ref(false);
const episodeData = ref<Episode | undefined>();

const videoSrc = computed(() => {
  const episode = episodeData.value;
  if (!episode) {
    return '';
  }
  return `${BASE_URL}${episode.link}`
});

const posterSrc = computed(() => {
  const episode = episodeData.value;
  if (!episode) {
    return '';
  }
  return `${BASE_URL}${episode.thumb}`
});

function onGroupWatch() {
  router.push({
    path: '/room',
    query: {
      id: nanoid(6),
      episodeId: episodeId.value,
    }
  });
}

function onCanPlay() {
  console.log('on can play');
  const storedTimeStr = localStorage.getItem(`episode-timestamp-${episodeId.value}`);
  if (storedTimeStr) {
    playerRef.value?.seek(Number(storedTimeStr));
  }
}

function serializeTimestamp() {
  if (!episodeId.value || !playerRef.value?.getPlaying()) {
    return
  }
  const seriesId = episodeData.value?.seriesId;
  const timestamp = playerRef.value?.getTimestamp();
  if (seriesId && timestamp) {
    saveWatchLater(seriesId, episodeId.value, episodeData.value?.episode ?? '');
    localStorage.setItem(`episode-timestamp-${episodeId.value}`, timestamp);
  }
}

useInterval(serializeTimestamp, 10000);

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
    sanitize(`${episodeData.value?.title ?? 'blank'}.mp4`)
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
        showSuccess('Episode deleted', `Successfully deleted episode ${episodeData.value?.title ?? ''}`);
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
      episodeData.value = newEpisode;
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
