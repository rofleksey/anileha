<template>
  <q-page class="full-width" padding>
    <div class="interactive-overlay" @click="overlayVisible = false"
         :style="{display: overlayVisible ? undefined : 'none'}">
      Click anywhere
    </div>
    <q-toolbar class="bg-purple text-white shadow-2 rounded-borders">
      <q-btn flat :label="`Room: ${episodeData?.title ?? ''}`"/>
      <q-btn flat :label="`${watchersState.length} viewers`"/>
    </q-toolbar>
    <VideoPlayer
      ref="playerRef"
      style="margin-top: 10px"
      :src="videoSrc"
      :loading="videoLoading"
      :progress="videoProgress"
      @play="onPlay"
      @pause="onPause"
      @time="onTime"
      @seek="onSeek"
    />
  </q-page>
</template>

<script setup lang="ts">
import {computed, ComputedRef, onMounted, onUnmounted, ref, watch} from 'vue';
import {Episode, RoomState, User, WatcherState, WatcherStatePartial} from 'src/lib/api-types';
import {BASE_URL, fetchEpisodeById} from 'src/lib/get-api';
import {showError, showHint, showSuccess} from 'src/lib/util';
import {useRoute} from 'vue-router';
import {useUserStore} from 'stores/user-store';
import VideoPlayer from 'components/VideoPlayer.vue';
import {useRoomStore} from 'stores/room-store';

interface IVideoPlayer {
  seek: (time: number) => void
  setPlaying: (value: boolean) => void
  getPlaying: () => boolean
  getTimestamp: () => number
  screenshot: () => string | null
}

let ws: WebSocket | undefined = undefined;
let downloadRequest: XMLHttpRequest | undefined = undefined;
let lastObjectUrl: string | undefined = undefined;
let reconnectInterval: NodeJS.Timeout | number | undefined = undefined;

const route = useRoute();

const roomStore = useRoomStore();

const roomId = computed(() => route.query.id?.toString());
console.log(roomId.value);

const suggestedEpisodeId = computed(() => Number(route.query.episodeId));
const episodeId = ref<number | null>(suggestedEpisodeId.value || null);

const userStore = useUserStore();
const curUser: ComputedRef<User | null> = computed(() => userStore.user);

const playerRef = ref<IVideoPlayer | undefined>()

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
const overlayVisible = ref(!navigator.userActivation?.hasBeenActive);

const dataLoading = ref(false);
const videoLoading = ref(false);
const videoError = ref(false);
const videoProgress = ref(0);
const episodeData = ref<Episode | undefined>();
const watchersState = ref<WatcherState[]>([]);
const videoSrc = ref<Blob | string>('');

watch(episodeId, refreshData);

watch(roomId, () => {
  roomStore.setRoomId(roomId.value ?? '');
});

watch(suggestedEpisodeId, () => {
  if (suggestedEpisodeId.value) {
    episodeId.value = suggestedEpisodeId.value;
  }
});

function onPlay() {
  updateSelfStatus((w) => {
    w.status = 'play';
  });

  const message: WebSocketMessage<RoomState> = {
    type: 'room-state',
    message: {
      episodeId: episodeId.value,
      timestamp: playerRef.value?.getTimestamp() ?? 0,
      playing: true,
    }
  }
  ws?.send(JSON.stringify(message));
}

function onPause() {
  updateSelfStatus((w) => {
    w.status = 'pause';
  });

  const message: WebSocketMessage<RoomState> = {
    type: 'room-state',
    message: {
      episodeId: episodeId.value,
      timestamp: playerRef.value?.getTimestamp() ?? 0,
      playing: false,
    }
  }
  ws?.send(JSON.stringify(message));
}

function updateSelfStatus(callback: (watcher: WatcherState) => void) {
  const selfWatcher = watchersState.value.find((it) => it.id === curUser.value?.id)
  if (!selfWatcher) {
    return
  }
  callback(selfWatcher)
  const message: WebSocketMessage<WatcherState> = {
    type: 'user-state',
    message: selfWatcher
  }
  ws?.send(JSON.stringify(message));
}

function onTime(timestamp: number) {
  updateSelfStatus((w) => {
    w.status = 'play';
    w.timestamp = timestamp;
  });
}

function onSeek(timestamp: number) {
  updateSelfStatus((w) => {
    w.status = 'pause';
    w.timestamp = timestamp;
  });

  const roomMessage: WebSocketMessage<RoomState> = {
    type: 'room-state',
    message: {
      episodeId: episodeId.value,
      timestamp: timestamp,
      playing: false,
    }
  }

  ws?.send(JSON.stringify(roomMessage));
}

interface WebSocketMessage<T> {
  type: string;
  message: T;
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

function webSocketConnect() {
  console.log('websocket reconnecting...');

  ws = new WebSocket(`ws://${window.location.host}/room/ws/${roomId.value}`);

  ws.onopen = function () {
    if (episodeId.value) {
      let status = 'pause';
      if (videoError.value) {
        status = 'error';
      }
      if (videoLoading.value) {
        status = 'loading';
      }
      const message: WebSocketMessage<WatcherStatePartial> = {
        type: 'user-state',
        message: {
          timestamp: playerRef.value?.getTimestamp() ?? 0,
          progress: videoProgress.value,
          status: status,
        }
      }
      ws?.send(JSON.stringify(message))
    }
    console.log('websocket connected');
  }

  ws.onmessage = function (e) {
    const data = JSON.parse(e.data) as WebSocketMessage<never>;
    if (data.type === 'full-state') {
      const {room, watchers} = data.message as FullStateMessage
      playerRef.value?.setPlaying(false);
      playerRef.value?.seek(room.timestamp)
      if (!episodeId.value) {
        episodeId.value = room.episodeId;
      }
      watchersState.value = watchers;
    } else if (data.type === 'room-state') {
      const roomState = data.message as RoomState;
      console.log(roomState, watchersState);

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
      if (!episodeId.value) {
        episodeId.value = roomState.episodeId;
      }
    } else if (data.type === 'user-state') {
      const message = data.message as WatcherState
      const index = watchersState.value.findIndex((it) => it.id === message.id);
      if (index >= 0) {
        const curWatcher = watchersState.value[index];
        if (curWatcher.id !== curUser.value?.id) {
          const lastStatus = curWatcher.status
          if (lastStatus === 'loading' && message.status === 'pause') {
            showHint('User ready', curWatcher.name);
          }
        }
        watchersState.value[index] = message;
      }
    } else if (data.type === 'user-connect') {
      const message = data.message as WatcherState
      const index = watchersState.value.findIndex((it) => it.id === message.id);
      if (index >= 0) {
        watchersState.value.splice(index, 1);
      }
      playerRef.value?.setPlaying(false);
      showHint('User connected', message.name);
      watchersState.value.push(message);
    } else if (data.type === 'user-disconnect') {
      const message = data.message as IdMessage
      const index = watchersState.value.findIndex((it) => it.id === message.id);
      if (index >= 0) {
        showHint('User disconnected', watchersState.value[index].name);
        watchersState.value.splice(index, 1);
      }
      playerRef.value?.setPlaying(false);
    } else {
      console.warn('invalid message type', data);
    }
  };

  ws.onclose = function (e) {
    console.error('websocket closed', e)
    ws?.close();
    ws = undefined;
  };

  ws.onerror = function (err) {
    console.error('websocket error', err)
    ws?.close();
    ws = undefined;
  };
}

function refreshData() {
  if (!episodeId.value) {
    episodeData.value = undefined;
    return
  }
  dataLoading.value = true;
  fetchEpisodeById(episodeId.value)
    .then((newEpisode) => {
      episodeData.value = newEpisode;
      loadVideo(`${BASE_URL}${newEpisode.link}`);
      ws?.close();
      webSocketConnect();
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
  reconnectInterval = setInterval(() => {
    if (!ws) {
      webSocketConnect();
    }
  }, 3000);
})

onUnmounted(() => {
  downloadRequest?.abort();
  if (lastObjectUrl) {
    URL.revokeObjectURL(lastObjectUrl);
    lastObjectUrl = undefined;
  }
  ws?.close();
  clearInterval(reconnectInterval);
})
</script>

<style lang="sass" scoped>
.interactive-overlay
  position: fixed
  width: 100%
  height: 100%
  top: 0
  left: 0
  background: rgba(0, 0, 0, 0.5)
  display: flex
  align-items: center
  justify-content: center
  text-align: center
  color: white
  font-family: monospace
  font-size: 5em
  z-index: 10000
</style>