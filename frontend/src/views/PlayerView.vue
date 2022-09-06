<script setup>
import axios from "axios";
import { ref, watch, computed, onMounted, onUnmounted } from "vue";
import { notify } from "@kyvg/vue3-notification";
import { useRoute } from "vue-router/dist/vue-router";
import durationFormat from "format-duration";
import { clamp, throttle } from "lodash";
import sanitize from "sanitize-filename";

const route = useRoute();

let hideControlsTimeout = undefined;
let trackTimeInterval = undefined;

const playerRef = ref(null);
const videoRef = ref(null);
const sliderRef = ref(null);
const volumeSliderRef = ref(null);

const seriesName = ref("Loading...");
const episodeName = ref("Loading...");
const totalTime = ref(0);
const curTime = ref(0);
const previewTimeStr = ref("");
const previewLeft = ref(0);
const playing = ref(false);
const volume = ref(1);
const videoSrc = ref(undefined);
const showControls = ref(true);

const totalTimeStr = computed(() => durationFormat(totalTime.value * 1000));
const curTimeStr = computed(() => durationFormat(curTime.value * 1000));
const progress = computed(() => curTime.value / totalTime.value);

function startTrackingTime() {
  trackTimeInterval = setInterval(() => {
    if (!videoRef.value) {
      return;
    }
    curTime.value = videoRef.value.currentTime;
  }, 50);
}

function stopHideControlsTimer() {
  if (hideControlsTimeout) {
    clearTimeout(hideControlsTimeout);
    hideControlsTimeout = undefined;
  }
}

function startHideControlsTimer() {
  if (!showControls.value) {
    showControls.value = true;
  }
  stopHideControlsTimer();
  if (!playing.value) {
    return;
  }
  hideControlsTimeout = setTimeout(() => {
    showControls.value = false;
  }, 3000);
}

function togglePlayback(e) {
  playing.value = !playing.value;
  e.stopPropagation();
}

function movePreview(e) {
  const bounds = e.target.getBoundingClientRect();
  const maxDx = sliderRef.value.clientWidth;
  const dx = clamp(e.clientX - bounds.left, 0, maxDx);
  const previewTime = (totalTime.value * dx) / maxDx;
  previewLeft.value = dx / maxDx;
  previewTimeStr.value = durationFormat(previewTime * 1000);
  startHideControlsTimer();
  return dx / maxDx;
}

function download() {
  const link = document.createElement("a");
  link.href = videoSrc.value;
  link.setAttribute(
    "download",
    sanitize(`${seriesName.value} - ${episodeName.value}.mp4`)
  );
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

const seekThrottle = throttle((newTime) => {
  videoRef.value.currentTime = newTime;
}, 50);

function seekTo(newTime, throttle) {
  if (throttle) {
    seekThrottle(newTime);
  } else {
    videoRef.value.currentTime = newTime;
  }
  playing.value = false;
  startHideControlsTimer();
}

function seekToPreview(e) {
  const newProgress = movePreview(e);
  const newTime = newProgress * totalTime.value;
  seekTo(newTime, true);
}

function onPreviewHover(e) {
  if (e.buttons === 1 || e.buttons === 3) {
    seekToPreview(e);
  } else {
    movePreview(e);
  }
}

function seekToVolume(e) {
  const bounds = e.target.getBoundingClientRect();
  const maxDx = volumeSliderRef.value.clientWidth;
  const dx = clamp(e.clientX - bounds.left, 0, maxDx);
  volume.value = dx / maxDx;
  startHideControlsTimer();
}

function onVolumeHover(e) {
  if (e.buttons === 1 || e.buttons === 3) {
    seekToVolume(e);
  }
}

function playerKeyboardListener(e) {
  // TODO: enable throttle here?
  if (e.key === " ") {
    togglePlayback();
    e.preventDefault();
  } else if (e.key === "ArrowLeft" || e.key === "a" || e.key === "ф") {
    seekTo(
      clamp(videoRef.value.current.currentTime - 10, 0, totalTime.value),
      false,
      true
    );
  } else if (e.key === "ArrowRight" || e.key === "d" || e.key === "в") {
    seekTo(
      clamp(videoRef.value.current.currentTime + 10, 0, totalTime.value),
      false,
      true
    );
  } else if (e.key === "," || e.key === "б") {
    seekTo(
      clamp(
        videoRef.value.current.currentTime - (e.ctrlKey ? 5 : 1) / 24,
        0,
        totalTime.value
      ),
      false,
      true
    );
  } else if (e.key === "." || e.key === "ю") {
    seekTo(
      clamp(
        videoRef.value.current.currentTime + (e.ctrlKey ? 5 : 1) / 24,
        0,
        totalTime.value
      ),
      false,
      true
    );
  }
}

function toggleFullScreen(e) {
  if (
    document.fullscreenElement ||
    document.webkitFullscreenElement ||
    document.mozFullScreenElement ||
    document.msFullscreenElement
  ) {
    if (document.exitFullscreen) {
      document.exitFullscreen();
    } else if (document.mozCancelFullScreen) {
      document.mozCancelFullScreen();
    } else if (document.webkitExitFullscreen) {
      document.webkitExitFullscreen();
    } else if (document.msExitFullscreen) {
      document.msExitFullscreen();
    }
  } else {
    const element = playerRef.value;
    if (element.requestFullscreen) {
      element.requestFullscreen();
    } else if (element.mozRequestFullScreen) {
      element.mozRequestFullScreen();
    } else if (element.webkitRequestFullscreen) {
      element.webkitRequestFullscreen(Element.ALLOW_KEYBOARD_INPUT);
    } else if (element.msRequestFullscreen) {
      element.msRequestFullscreen();
    }
    element.focus();
  }
  e.stopPropagation();
}

watch(playing, (val) => {
  if (val) {
    videoRef.value.play();
    startHideControlsTimer();
  } else {
    videoRef.value.pause();
    stopHideControlsTimer();
  }
});

watch(volume, (val) => {
  videoRef.value.volume = val;
});

onUnmounted(() => {
  clearInterval(trackTimeInterval);
  trackTimeInterval = null;
});

onMounted(() => {
  axios(`/episodes/${route.params.id}`)
    .then(({ data }) => {
      axios(`/series/${data.seriesId}`)
        .then(({ data }) => {
          seriesName.value = data.name;
        })
        .catch((err) => {
          notify({
            title: "Failed to load series data",
            text: err?.response?.data?.error ?? "",
            type: "error",
          });
        });
      episodeName.value = data.name;
      videoSrc.value = data.link;
      videoRef.value.addEventListener(
        "loadeddata",
        () => {
          console.log(videoRef.value.duration);
          totalTime.value = videoRef.value.duration;
          startTrackingTime();
        },
        false
      );
    })
    .catch((err) => {
      notify({
        title: "Failed to load episode",
        text: err?.response?.data?.error ?? "",
        type: "error",
      });
    });
});
</script>

<template>
  <div
    :class="{ player: true, immersed: !showControls }"
    @mousemove="startHideControlsTimer"
    @keydown="playerKeyboardListener"
    @click="togglePlayback"
    tabIndex="0"
    ref="playerRef"
  >
    <div class="video-container">
      <video preload ref="videoRef" class="video" :src="videoSrc" />
      <div class="controls">
        <div>
          <button
            :class="{ 'play-pause-btn': true, play: playing, pause: !playing }"
            @click="togglePlayback"
          />
          <div>{{ curTimeStr }}</div>
          <div
            class="slider seeker"
            @mousemove="onPreviewHover"
            @mousedown="seekToPreview"
            @click="(e) => e.stopPropagation()"
            ref="sliderRef"
          >
            <div class="preview" :style="{ left: `${previewLeft * 100}%` }">
              {{ previewTimeStr }}
            </div>
            <div class="handle" :style="{ width: `${progress * 100}%` }" />
          </div>
          <div class="duration">{{ totalTimeStr }}</div>
          <div
            class="slider volume"
            @mousemove="onVolumeHover"
            @mousedown="seekToVolume"
            @click="(e) => e.stopPropagation()"
            ref="volumeSliderRef"
          >
            <div class="handle" :style="{ width: `${volume * 100}%` }" />
          </div>
          <button @click="download">
            <svg class="icon">
              <g>
                <polygon points="2,2 8,14 14,2" />
              </g>
            </svg>
          </button>
          <button @click="toggleFullScreen">
            <svg class="icon">
              <g>
                <polygon points="8,2 14,2 14,8 12,8 12,4 8,4" />
                <polygon points="2,8 4,8 4,12 8,12 8,14 2,14" />
              </g>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.player {
  width: 100%;
}

.video-container {
  position: relative;
  background: #000;
  max-width: 600px;
  width: 100%;
  height: 100%;
  margin: 0 auto;
}

.player:fullscreen .video-container {
  max-width: 100%;
}

.video {
  width: 100%;
  height: 100%;
  display: block;
  outline: none;
  animation: fade-in 0.3s;
}

.controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 50px;
  transition: opacity 0.3s;
  animation: fade-in 0.3s;
  background: linear-gradient(
    0deg,
    rgba(0, 0, 0, 0.8),
    rgba(0, 0, 0, 0.4) 25%,
    rgba(0, 0, 0, 0.2) 50%,
    rgba(0, 0, 0, 0.1) 75%,
    transparent
  );
}

.player.immersed .controls {
  opacity: 0;
}

.player.immersed {
  cursor: none;
}

.controls > div {
  display: flex;
  align-items: center;
  padding-left: 16px;
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
}

.controls > div > * {
  margin: 0 8px;
}

.controls button.play-pause-btn {
  background-position: 50%;
  background-repeat: no-repeat;
  background-image: url(data:image/svg+xml;base64,PHN2ZyB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgMTYgMTYiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDE2IDE2OyIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+CiAgICA8cGF0aCBmaWxsPSJ3aGl0ZSIgZD0iTSAzLDIgTCAxMyw4IEwgMywxNCBaIj48L3BhdGg+Cjwvc3ZnPgo=);
  background-size: 16px;
  width: 32px;
  height: 48px;
}

.controls button.play-pause-btn.play {
  background-image: url(data:image/svg+xml;base64,PHN2ZyB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgMTYgMTYiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDE2IDE2OyIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+CiAgICA8cGF0aCBmaWxsPSJ3aGl0ZSIgZD0iTSAyLDIgTCA2LDIgTCA2LDE0IEwgMiwxNCBaIj48L3BhdGg+CiAgICA8cGF0aCBmaWxsPSJ3aGl0ZSIgZD0iTSAxMCwyIEwgMTQsMiBMIDE0LDE0IEwgMTAsMTQgWiI+PC9wYXRoPgo8L3N2Zz4K);
}

.controls button {
  margin: 0;
  padding: 16px 8px;
  background: none;
  box-shadow: none;
  border: none;
  transition: -webkit-filter 0.15s;
  transition: filter 0.15s;
  cursor: pointer;
}

.controls button:hover {
  -webkit-filter: drop-shadow(0 0 10px #fff);
  filter: drop-shadow(0 0 10px white);
}

button {
  font: inherit;
  color: inherit;
  background-color: hsla(0, 0%, 100%, 0.05);
  border: 1px solid rgba(255, 255, 255, 0);
  padding: 0.5em 1em;
  border-radius: 0.25em;
  box-shadow: inset 0 0 0 0 rgb(255 255 255 / 5%);
  transition: box-shadow 0.3s, border 0.15s ease-in;
  outline: none;
  min-height: 2em;
  overflow: hidden;
  display: inline-flex;
  justify-content: center;
  align-items: center;
  cursor: default;
}

.controls .seeker {
  flex-grow: 1;
}

.slider {
  position: relative;
  padding: 1.5em 0;
  box-sizing: initial;
  background: hsla(0, 0%, 100%, 0.2);
  background-clip: content-box;
  height: 2px;
  cursor: pointer;
  margin: 0 16px;
}

.slider .preview {
  position: absolute;
  bottom: 100%;
  width: 96px;
  margin-left: -46px;
  opacity: 0;
  height: auto;
  background: transparent;
  text-align: center;
  transition: opacity 0.15s;
}

.slider:hover .preview {
  opacity: 1;
}

.slider:not(.dragging) .handle {
  transition: width 0.15s;
}

.slider > .handle {
  position: relative;
  height: 2px;
  max-width: 100%;
  background: #e53232;
  z-index: 1;
}

.slider .handle:before {
  content: "";
  position: absolute;
  height: 16px;
  width: 16px;
  background: hsla(0, 0%, 100%, 0.2);
  border-radius: 50%;
  right: -8px;
  top: -7px;
  transition: background 0.15s;
}

.slider:active .handle:before {
  background: hsla(0, 0%, 100%, 0.5);
}

.slider .handle:after {
  content: "";
  position: absolute;
  height: 4px;
  width: 4px;
  background: #fff;
  border-radius: 50%;
  right: -2px;
  top: -1px;
}

.controls .volume {
  width: 64px;
}

.icon {
  width: 1em;
  height: 1em;
  fill: #fff;
}

button > svg {
  flex: none;
}

.player:fullscreen {
  width: 100% !important;
}

.player:-webkit-full-screen {
  width: 100% !important;
}
</style>
