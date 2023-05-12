<template>
  <q-page class="full-width" padding>
    <q-toolbar class="bg-purple text-white shadow-2 rounded-borders">
      <q-btn flat :label="`Room: ${episodeData?.title ?? ''}`"/>
      <q-btn flat :label="`${watchersState.length} viewers`"/>
      <q-btn
        v-if="episodeIndex >= 0"
        flat
        round
        icon="skip_previous"
        :disable="episodeIndex === 0"
        @click="changePageEpisode(episodeListData![episodeIndex - 1].id)"/>
      <q-btn
        v-if="episodeIndex >= 0"
        flat
        round
        icon="skip_next"
        :disable="episodeIndex === episodeListData?.length - 1"
        @click="changePageEpisode(episodeListData![episodeIndex + 1].id)"/>
    </q-toolbar>
    <InteractiveOverlay>
      Click to enable
    </InteractiveOverlay>
    <VideoPlayer
      ref="playerRef"
      style="margin-top: 10px"
      :src="videoSrc"
      :loading="videoLoading"
      :progress="videoProgress"
      pause-on-seek
      @play="onPlay"
      @pause="onPause"
      @time="onTime"
      @seek="onSeek"
    >
      <template #default="{playing}">
        <div :class="{RoomUsers: true, playing: playing}">
          <q-avatar rounded size="xl" v-for="watcher in watchersState" :key="watcher.id">
            <img :src="watcher.thumb" :alt="watcher.name"/>
            <q-badge floating rounded :color="watcherIconColor(watcher)">
              {{ watcherIconText(watcher) }}
            </q-badge>
          </q-avatar>
        </div>
      </template>
    </VideoPlayer>
  </q-page>
</template>

<script setup lang="ts">
import {computed, ComputedRef, onUnmounted, ref, watch} from 'vue';
import {Episode, RoomState, User, WatcherState, WatcherStatePartial} from 'src/lib/api-types';
import {BASE_URL, fetchEpisodeById, fetchEpisodesBySeriesId} from 'src/lib/get-api';
import {showError, showHint, showSuccess} from 'src/lib/util';
import {useRoute, useRouter} from 'vue-router';
import {useUserStore} from 'stores/user-store';
import VideoPlayer from 'components/VideoPlayer.vue';
import InteractiveOverlay from 'components/InteractiveOverlay.vue';
import formatDuration from 'format-duration';
import {useWebSocket} from 'src/lib/ws';

interface IVideoPlayer {
  seek: (time: number) => void
  setPlaying: (value: boolean) => void
  getPlaying: () => boolean
  getTimestamp: () => number
  screenshot: () => string | null
}

let downloadRequest: XMLHttpRequest | undefined = undefined;
let lastObjectUrl: string | undefined = undefined;

const route = useRoute();
const router = useRouter();

const roomId = computed(() => route.query.id?.toString());
const pageEpisodeId = computed(() => Number(route.query.episodeId?.toString()));

const userStore = useUserStore();
const curUser: ComputedRef<User | null> = computed(() => userStore.user);

const playerRef = ref<IVideoPlayer | undefined>()

const dataLoading = ref(false);
const videoLoading = ref(false);
const videoError = ref(false);
const videoProgress = ref(0);
const videoEpisodeId = ref<number | null>(null);
const episodeData = ref<Episode | undefined>();
const episodeListData = ref<Episode[] | undefined>();
const watchersState = ref<WatcherState[]>([]);
const videoSrc = ref<Blob | string>('');

const episodeIndex = computed(() => {
  if (!episodeListData.value || !videoEpisodeId.value) {
    return -1;
  }
  return episodeListData.value?.findIndex((it) => it.id === videoEpisodeId.value) ?? -1;
});

function changePageEpisode(newEpisodeId: number) {
  if (pageEpisodeId.value !== newEpisodeId) {
    router.replace({
      path: '/room',
      query: {
        id: roomId.value,
        episodeId: newEpisodeId,
      }
    });
  }
  videoEpisodeId.value = newEpisodeId;
}

const {sendWs} = useWebSocket({
  url: `ws://${window.location.host}/room/ws/${roomId.value}`,
  onConnect: () => {
    if (pageEpisodeId.value) {
      const selfWatcher = watchersState.value.find((it) => it.id === curUser.value?.id)
      if (!selfWatcher) {
        return
      }

      sendWs<WatcherStatePartial>('user-state', {
        timestamp: selfWatcher.timestamp,
        progress: selfWatcher.progress,
        status: selfWatcher.status,
      });
    }
  },
  onMessage: (type, message) => {
    if (type === 'full-state') {
      console.log(message);
      const {room, watchers} = message as FullStateMessage

      playerRef.value?.setPlaying(false);
      playerRef.value?.seek(room.timestamp)
      if (room.episodeId) {
        changePageEpisode(room.episodeId);
      } else {
        changePageEpisode(pageEpisodeId.value);
      }
      watchersState.value = watchers;

      sendWs<RoomState>('room-state', {
        episodeId: room.episodeId || pageEpisodeId.value,
        timestamp: -1,
        playing: false,
      });
    } else if (type === 'room-state') {
      console.log(message);
      const roomState = message as RoomState;
      const localPlaying = playerRef.value?.getPlaying() ?? false;
      const initiator = watchersState.value.find((it) => it.id === roomState.initiatorId);
      if (initiator && initiator.id !== curUser.value?.id) {
        if (localPlaying && !roomState.playing) {
          showHint('Paused', `by ${initiator.name}`);
        }
        if (!localPlaying && roomState.playing) {
          showHint('Resumed', `by ${initiator.name}`);
        }
      }
      playerRef.value?.setPlaying(roomState.playing);
      playerRef.value?.seek(roomState.timestamp)
      if (roomState.episodeId) {
        changePageEpisode(roomState.episodeId);
      }
    } else if (type === 'user-state') {
      const newState = message as WatcherState
      const index = watchersState.value.findIndex((it) => it.id === newState.id);
      if (index >= 0) {
        const curWatcher = watchersState.value[index];
        if (curWatcher.id !== curUser.value?.id) {
          const lastStatus = curWatcher.status
          if (lastStatus === 'loading' && newState.status === 'pause') {
            showHint('User ready', curWatcher.name);
          }
        }
        watchersState.value[index] = newState;
      }
    } else if (type === 'user-connect') {
      const newState = message as WatcherState
      const index = watchersState.value.findIndex((it) => it.id === newState.id);
      if (index >= 0) {
        watchersState.value.splice(index, 1);
      }
      playerRef.value?.setPlaying(false);
      showHint('User connected', newState.name);
      watchersState.value.push(newState);
    } else if (type === 'user-disconnect') {
      console.log(type, message);
      const idMsg = message as IdMessage
      const index = watchersState.value.findIndex((it) => it.id === idMsg.id);
      if (index >= 0) {
        showHint('User disconnected', watchersState.value[index].name);
        watchersState.value.splice(index, 1);
      }
      playerRef.value?.setPlaying(false);
    } else {
      console.warn('invalid message type', type);
    }
  },
})

watch(videoEpisodeId, refreshData);

function watcherIconText(watcher: WatcherState) {
  if (watcher.status === 'loading') {
    return `${Math.floor(watcher.progress * 100)} %`
  }
  if (watcher.status === 'play' || watcher.status === 'pause') {
    return formatDuration(watcher.timestamp * 1000);
  }
  return watcher.status;
}

function watcherIconColor(watcher: WatcherState) {
  if (watcher.status === 'loading') {
    return 'purple';
  }
  if (watcher.status === 'error') {
    return 'red';
  }
  if (watcher.status === 'play') {
    return 'green'
  }
  if (watcher.status === 'pause') {
    return 'blue'
  }
  return 'gray'
}

function onPlay() {
  console.log('onPlay');
  updateSelfStatus((w) => {
    w.status = 'play';
  });

  sendWs<RoomState>('room-state', {
    episodeId: pageEpisodeId.value,
    timestamp: playerRef.value?.getTimestamp() ?? 0,
    playing: true,
  });
}

function onPause() {
  console.log('onPause');
  updateSelfStatus((w) => {
    w.status = 'pause';
  });

  sendWs<RoomState>('room-state', {
    episodeId: pageEpisodeId.value,
    timestamp: playerRef.value?.getTimestamp() ?? 0,
    playing: false,
  });
}

function updateSelfStatus(callback: (watcher: WatcherState) => void) {
  const selfWatcher = watchersState.value.find((it) => it.id === curUser.value?.id)
  if (!selfWatcher) {
    return
  }
  callback(selfWatcher)
  sendWs<WatcherState>('user-state', selfWatcher);
}

function onTime(timestamp: number) {
  if (videoLoading.value || videoError.value) {
    return
  }
  updateSelfStatus((w) => {
    w.status = playerRef.value?.getPlaying() ? 'play' : 'pause';
    w.timestamp = timestamp;
  });
}

function onSeek(timestamp: number) {
  updateSelfStatus((w) => {
    w.status = 'pause';
    w.timestamp = timestamp;
  });

  sendWs<RoomState>('room-state', {
    episodeId: pageEpisodeId.value,
    timestamp: timestamp,
    playing: false,
  });
}

interface IdMessage {
  id: number;
}

interface FullStateMessage {
  room: RoomState;
  watchers: WatcherState[];
}

function loadVideo(src: string) {
  videoLoading.value = true;
  videoProgress.value = 0;
  downloadRequest?.abort();
  if (lastObjectUrl) {
    URL.revokeObjectURL(lastObjectUrl);
    lastObjectUrl = undefined;
  }
  downloadRequest = new XMLHttpRequest();
  downloadRequest.open('GET', src, true);
  downloadRequest.responseType = 'blob';
  downloadRequest.onload = function () {
    if (this.status === 200) {
      videoLoading.value = false;
      lastObjectUrl = URL.createObjectURL(this.response);
      videoSrc.value = lastObjectUrl;
      updateSelfStatus((w) => {
        w.status = 'pause';
      })
      showSuccess('Video downloaded')
    } else {
      videoLoading.value = false;
      videoError.value = true;
      updateSelfStatus((w) => {
        w.status = 'error';
      })
      showError('Video download error', {})
    }
  }
  downloadRequest.onerror = () => {
    videoLoading.value = false;
    videoError.value = true;
    updateSelfStatus((w) => {
      w.status = 'error';
    })
    showError('Video download error', {})
  }
  downloadRequest.onprogress = (e) => {
    videoLoading.value = true;
    videoError.value = false;
    const progress = e.loaded / e.total;
    videoProgress.value = progress;
    updateSelfStatus((w) => {
      w.status = 'loading';
      w.progress = progress;
    });
  }
  downloadRequest.send();
}

function refreshData() {
  if (!pageEpisodeId.value) {
    episodeData.value = undefined;
    return
  }
  dataLoading.value = true;
  fetchEpisodeById(pageEpisodeId.value)
    .then((newEpisode) => {
      episodeData.value = newEpisode;
      loadVideo(`${BASE_URL}${newEpisode.link}`);

      if (newEpisode.seriesId) {
        fetchEpisodesBySeriesId(newEpisode.seriesId).then((newEpisodeList) => {
          episodeListData.value = newEpisodeList
        }).catch((e) => {
          showError('Failed to fetch episode list', e);
        })
      }

      updateSelfStatus((w) => {
        w.timestamp = 0;
      });

      sendWs<RoomState>('room-state', {
        episodeId: pageEpisodeId.value,
        timestamp: -1,
        playing: false,
      });
    })
    .catch((e) => {
      showError('Failed to fetch episode', e);
    })
    .finally(() => {
      dataLoading.value = false;
    });
}

onUnmounted(() => {
  downloadRequest?.abort();
  if (lastObjectUrl) {
    URL.revokeObjectURL(lastObjectUrl);
    lastObjectUrl = undefined;
  }
})
</script>

<style lang="sass" scoped>
.RoomUsers
  pointer-events: none
  position: absolute
  left: 0
  top: 0
  margin: 10px
  z-index: 100
  display: flex
  flex-direction: column
  align-items: start
  justify-content: center
  gap: 20px
  transition: opacity 0.3s ease

  &.playing
    opacity: 0.5
</style>
