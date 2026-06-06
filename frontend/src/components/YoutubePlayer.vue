<template>
  <div class="player-wrap">
    <div id="yt-player"></div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, watch } from "vue";
const props = defineProps({ videoId: String, startPosition: { type: Number, default: 0 }, autoPlay: { type: Boolean, default: false } });
const emits = defineEmits(["play", "pause", "seek", "ended"]);
let player = null;

let lastKnownTime = 0;
let pausedPoll = null;

function initPlayer() {
  player = new YT.Player("yt-player", {
    videoId: props.videoId,
    events: {
      onReady: onReady,
      onStateChange: onStateChange,
    },
  });
}

function onReady(event) {
  player = event.target;
  if (props.startPosition > 0) {
    player.seekTo(props.startPosition, true);
  }
  if (props.autoPlay) player.playVideo();
}

function onStateChange(event) {
  if (event.data === YT.PlayerState.PLAYING) {
    clearInterval(pausedPoll);
    pausedPoll = null;
    lastKnownTime = player.getCurrentTime();
    emits("play");
  }
  if (event.data === YT.PlayerState.PAUSED) {
    lastKnownTime = player.getCurrentTime();
    emits("pause");
    pausedPoll = setInterval(() => {
      const t = player?.getCurrentTime() ?? 0;
      if (Math.abs(t - lastKnownTime) > 1) {
        lastKnownTime = t;
        emits("seek", t);
      }
    }, 500);
  }
  if (event.data === YT.PlayerState.LOAD) emits("load");
  if (event.data === YT.PlayerState.ENDED) emits("ended")
}

function play() {
  player?.playVideo();
}

function pause() {
  player?.pauseVideo();
}

function seekTo(sec) {
  player?.seekTo(sec, true);
}

function getCurrentTime() {
  return player?.getCurrentTime();
}

function loadVideoById(id) {
  return player?.loadVideoById(id);
}

function setLastKnownTime(t) {
  lastKnownTime = t;
}

watch(
  () => props.videoId,
  (id) => {
    if (player && id) player.loadVideoById(id);
  },
);

onMounted(() => {
  if (window.YT?.Player) {
    initPlayer();
  } else {
    const prev = window.onYouTubeIframeAPIReady;
    window.onYouTubeIframeAPIReady = () => {
      prev?.();
      initPlayer();
    };
  }
});

onUnmounted(() => {
  player?.destroy();
  player = null;
});

defineExpose({ play, pause, seekTo, getCurrentTime, loadVideoById, setLastKnownTime });
</script>

<style scoped>
.player-wrap {
  width: 100%;
  height: 100%;
}
.player-wrap :deep(iframe),
#yt-player {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
