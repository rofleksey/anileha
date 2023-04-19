<template>
  <div
    :class="{player: true, loading: props.loading, immersed: !showControls}"
    @mousemove="restartHideControlsTimer"
    @keydown="playerKeyboardListener"
    @click="togglePlayback"
    tabIndex="-1"
    ref="playerRef">
    <canvas
      ref="canvasRef"
      style="display: none"
      :width="videoWidth"
      :height="videoHeight"></canvas>
    <div class="video-container">
      <q-inner-loading :showing="props.loading">
        <q-circular-progress
          show-value
          style="margin: 0"
          class="text-light-blue q-ma-md"
          :value="Math.floor(100 * props.progress)"
          track-color="grey-9"
          size="xl"
          color="light-blue"
        />
      </q-inner-loading>
      <video
        preload
        draggable="false"
        ref="videoRef"
        class="video"
        :poster="props.poster"
        :src="props.src"/>
      <slot :playing="playing"></slot>
      <div class="controls">
        <div>
          <button
            :class="{'play-pause-btn': true, play: playing, pause: !playing}"
            @click="togglePlayback"/>
          <div>{{ videoTimestampStr }}</div>
          <div
            class="slider seeker"
            @mousemove="onPreviewHover"
            @mousedown="seekToPreview"
            @click="(e)=> { e.stopPropagation() }"
            ref="sliderRef">
            <div class="preview"
                 :style="{left: `${previewLeft * 100}%`}">
              {{ previewTimestampStr }}
            </div>
            <div class="handle"
                 :style="{width: `${progress * 100}%`}"/>
          </div>
          <div class="duration">{{ totalDurationStr }}</div>
          <div
            class="slider volume"
            @mousemove="onVolumeHover"
            @mousedown="seekToVolume"
            @click="(e) => { e.stopPropagation() }"
            ref="volumeSliderRef">
            <div class="handle"
                 :style="{width: `${volume * 100}%`}"/>
          </div>
          <button @click="toggleFullscreen">
            <svg class="icon">
              <g>
                <polygon points="8,2 14,2 14,8 12,8 12,4 8,4"/>
                <polygon points="2,8 4,8 4,12 8,12 8,14 2,14"/>
              </g>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref, watch} from 'vue';
import {clamp, throttle} from 'lodash';
import formatDuration from 'format-duration';
import {useInterval} from 'src/lib/composables';

interface Props {
  src: Blob | string;
  poster: string;
  loading?: boolean;
  progress?: number;
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'time', timestamp: number): void
  (e: 'seek', timestamp: number): void
  (e: 'play'): void
  (e: 'pause'): void
}>()

let hideControlsTimer: NodeJS.Timeout | number | undefined = undefined;

useInterval(updateTime, 100)

const canvasRef = ref<HTMLCanvasElement | undefined>()
const playerRef = ref<HTMLDivElement | undefined>();
const videoRef = ref<HTMLVideoElement | undefined>();
const sliderRef = ref<HTMLDivElement | undefined>();
const volumeSliderRef = ref<HTMLDivElement | undefined>();
const videoWidth = ref(1920)
const videoHeight = ref(1080)
const playing = ref(false);
const videoTimestamp = ref(0);
const previewTimestamp = ref(0);
const totalDuration = ref(1);
const previewLeft = ref(0);
const volume = ref(1);
const showControls = ref(true);

const totalDurationStr = computed(() => formatDuration(totalDuration.value * 1000));
const previewTimestampStr = computed(() => formatDuration(previewTimestamp.value * 1000));
const videoTimestampStr = computed(() => formatDuration(videoTimestamp.value * 1000));
const progress = computed(() => videoTimestamp.value / totalDuration.value);

watch(videoRef, () => {
  const video = videoRef.value;
  if (!video) {
    return;
  }
  video.controls = false;
  video.addEventListener('loadeddata', () => {
    totalDuration.value = video.duration;
    videoWidth.value = video.width;
    videoHeight.value = video.height;
  }, false);
});

watch(playing, () => {
  if (playing.value) {
    videoRef.value?.play();
    restartHideControlsTimer()
  } else {
    videoRef.value?.pause();
    stopHideControlsTimer();
  }
});

watch(volume, () => {
  const video = videoRef.value;
  if (!video) {
    return;
  }
  video.volume = volume.value;
})

function stopHideControlsTimer() {
  if (hideControlsTimer !== undefined) {
    clearTimeout(hideControlsTimer);
  }
}

function restartHideControlsTimer() {
  if (!showControls.value) {
    showControls.value = true;
  }
  if (!playing.value) {
    return;
  }
  stopHideControlsTimer();
  hideControlsTimer = setTimeout(() => {
    showControls.value = false;
  }, 3000);
}

function updateTime() {
  const video = videoRef.value;
  if (!video) {
    return;
  }
  const curTime = video.currentTime
  videoTimestamp.value = curTime;
  emit('time', curTime)
}

function playerKeyboardListener(e: KeyboardEvent) {
  // TODO: enable throttle here?
  const video = videoRef.value;
  if (!video) {
    return;
  }
  if (e.key === ' ') {
    togglePlayback();
    e.preventDefault();
  } else if (e.key === 'ArrowLeft' || e.key === 'a' || e.key === 'ф') {
    seekTo(clamp(video.currentTime - 10, 0, totalDuration.value), false);
  } else if (e.key === 'ArrowRight' || e.key === 'd' || e.key === 'в') {
    seekTo(clamp(video.currentTime + 10, 0, totalDuration.value), false);
  } else if (e.key === ',' || e.key === 'б') {
    seekTo(clamp(video.currentTime - (e.ctrlKey ? 5 : 1) / 24, 0, totalDuration.value), false);
  } else if (e.key === '.' || e.key === 'ю') {
    seekTo(clamp(video.currentTime + (e.ctrlKey ? 5 : 1) / 24, 0, totalDuration.value), false);
  }
}

function togglePlayback(e?: MouseEvent) {
  const newValue = !playing.value
  playing.value = newValue;
  if (newValue) {
    emit('play')
  } else {
    emit('pause')
  }
  e?.stopPropagation();
}

function movePreview(e: MouseEvent) {
  const bounds = (e.target as HTMLDivElement).getBoundingClientRect();
  const maxDx = (sliderRef.value as HTMLDivElement).clientWidth;
  const dx = clamp(e.clientX - bounds.left, 0, maxDx);
  previewTimestamp.value = totalDuration.value * dx / maxDx;
  previewLeft.value = dx / maxDx;
  restartHideControlsTimer();
  return dx / maxDx;
}

const seekThrottle = throttle((newTime: number) => {
  const video = videoRef.value;
  if (!video) {
    return;
  }
  video.currentTime = newTime;
  emit('seek', newTime)
}, 100);

function seekTo(newTime: number, throttle: boolean, remote?: boolean) {
  const video = videoRef.value;
  if (!video) {
    return;
  }

  if (throttle) {
    seekThrottle(newTime);
  } else {
    video.currentTime = newTime;
    if (!remote) {
      emit('seek', newTime)
    }
  }

  videoTimestamp.value = video.currentTime;
  if (!remote) {
    playing.value = false;
  }

  restartHideControlsTimer();
}

function onPreviewHover(e: MouseEvent) {
  if (e.buttons === 1 || e.buttons === 3) {
    seekToPreview(e);
  } else {
    movePreview(e);
  }
}

function seekToPreview(e: MouseEvent) {
  const newProgress = movePreview(e);
  const newTime = newProgress * totalDuration.value;
  seekTo(newTime, true);
}

function onVolumeHover(e: MouseEvent) {
  if (e.buttons === 1 || e.buttons === 3) {
    seekToVolume(e);
  }
}

function seekToVolume(e: MouseEvent) {
  const bounds = (e.target as HTMLDivElement).getBoundingClientRect();
  const maxDx = (volumeSliderRef.value as HTMLDivElement).clientWidth;
  const dx = clamp(e.clientX - bounds.left, 0, maxDx);
  volume.value = dx / maxDx;
  restartHideControlsTimer();
}

function toggleFullscreen(e?: MouseEvent) {
  if (document.fullscreenElement) {
    if (document.exitFullscreen) {
      document.exitFullscreen();
    }
  } else {
    const player = playerRef.value;
    if (!player) {
      return
    }
    if (player.requestFullscreen) {
      player.requestFullscreen();
    }
    player.focus();
  }
  e?.stopPropagation();
}

defineExpose({
  seek: (time: number) => seekTo(time, false, true),
  setPlaying: (value: boolean) => {
    console.log(`video player - set playing - ${value}`)
    playing.value = value;
  },
  getTimestamp: () => {
    return videoTimestamp.value
  },
  getPlaying: () => {
    return playing.value
  },
  screenshot: () => {
    const video = videoRef.value;
    if (!video) {
      return null;
    }
    const canvas = canvasRef.value;
    if (!canvas) {
      return null;
    }
    const ctx = canvas.getContext('2d');
    if (!ctx) {
      return null;
    }
    ctx.fillRect(0, 0, videoWidth.value, videoHeight.value);
    ctx.drawImage(video, 0, 0, videoWidth.value, videoHeight.value);
    return canvas.toDataURL('image/jpeg');
  }
})
</script>

<style lang="sass" scoped>
.player
  user-select: none
  user-drag: none

.loading
  pointer-events: none

.video-container
  position: relative
  background: #000
  max-width: 800px
  width: 100%
  height: 100%
  margin: 0 auto

.player:fullscreen .video-container
  max-width: 100%

.video
  width: 100%
  height: 100%
  outline: none
  animation: fade-in .3s

.controls
  position: absolute
  bottom: 0
  left: 0
  right: 0
  height: 50px
  transition: opacity .3s
  animation: fade-in .3s
  background: linear-gradient(0deg, rgba(0, 0, 0, 0.8), rgba(0, 0, 0, 0.4) 25%, rgba(0, 0, 0, 0.2) 50%, rgba(0, 0, 0, 0.1) 75%, transparent)

.player.immersed
  .controls
    opacity: 0

  cursor: none

.controls
  > div
    display: flex
    align-items: center
    padding-left: 16px
    position: absolute
    bottom: 0
    left: 0
    right: 0

    > *
      margin: 0 8px

  button
    &.play-pause-btn
      background-position: 50%
      background-repeat: no-repeat
      background-image: url(data:image/svg+xml;base64,PHN2ZyB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgMTYgMTYiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDE2IDE2OyIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+CiAgICA8cGF0aCBmaWxsPSJ3aGl0ZSIgZD0iTSAzLDIgTCAxMyw4IEwgMywxNCBaIj48L3BhdGg+Cjwvc3ZnPgo=)
      background-size: 16px
      width: 32px
      height: 48px

      &.play
        background-image: url(data:image/svg+xml;base64,PHN2ZyB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgMTYgMTYiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDE2IDE2OyIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+CiAgICA8cGF0aCBmaWxsPSJ3aGl0ZSIgZD0iTSAyLDIgTCA2LDIgTCA2LDE0IEwgMiwxNCBaIj48L3BhdGg+CiAgICA8cGF0aCBmaWxsPSJ3aGl0ZSIgZD0iTSAxMCwyIEwgMTQsMiBMIDE0LDE0IEwgMTAsMTQgWiI+PC9wYXRoPgo8L3N2Zz4K)

    margin: 0
    padding: 16px 8px
    background: none
    box-shadow: none
    border: none
    transition: -webkit-filter .15s
    transition: filter .15s
    cursor: pointer

    &:hover
      -webkit-filter: drop-shadow(0 0 10px #fff)
      filter: drop-shadow(0 0 10px white)

button
  font: inherit
  color: inherit
  background-color: hsla(0, 0%, 100%, 0.05)
  border: 1px solid rgba(255, 255, 255, 0)
  padding: 0.5em 1em
  border-radius: 0.25em
  box-shadow: inset 0 0 0 0 rgb(255 255 255 / 5%)
  transition: box-shadow .3s, border .15s ease-in
  outline: none
  min-height: 2em
  overflow: hidden
  display: inline-flex
  justify-content: center
  align-items: center
  cursor: default

.controls .seeker
  flex-grow: 1

.slider
  position: relative
  padding: 1.5em 0
  box-sizing: initial
  background: hsla(0, 0%, 100%, 0.2)
  background-clip: content-box
  height: 2px
  cursor: pointer
  margin: 0 16px

  .preview
    position: absolute
    bottom: 100%
    width: 96px
    margin-left: -46px
    opacity: 0
    height: auto
    background: transparent
    text-align: center
    transition: opacity .15s

  &:hover .preview
    opacity: 1

  &:not(.dragging) .handle
    transition: width .15s

  > .handle
    position: relative
    height: 2px
    max-width: 100%
    background: #7e57c2
    z-index: 1

  .handle:before
    content: ""
    position: absolute
    height: 16px
    width: 16px
    background: hsla(0, 0%, 100%, 0.2)
    border-radius: 50%
    right: -8px
    top: -7px
    transition: background .15s

  &:active .handle:before
    background: hsla(0, 0%, 100%, 0.5)

  .handle:after
    content: ""
    position: absolute
    height: 4px
    width: 4px
    background: #fff
    border-radius: 50%
    right: -2px
    top: -1px

.controls .volume
  width: 64px

.icon
  width: 1em
  height: 1em
  fill: #fff

button > svg
  flex: none

.player
  &:fullscreen, &:-webkit-full-screen
    width: 100% !important

.player:fullscreen
  &.immersed .userList
    opacity: 0.5

  .userList
    transform: scale(1)
    opacity: 1

.dots
  white-space: nowrap
  overflow: hidden
  -o-text-overflow: ellipsis
  text-overflow: ellipsis
  max-width: 100%
  display: inline-block
</style>
