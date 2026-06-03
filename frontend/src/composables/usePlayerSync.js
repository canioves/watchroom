import { ref } from "vue";

export function usePlayerSync(playerRef, send) {
  const isRemoteUpdate = ref(false);
  function onPlayerPlay() {
    if (isRemoteUpdate.value) {
      isRemoteUpdate.value = false;
      return;
    }
    send({ type: "play", position: playerRef.value.getCurrentTime(), sentAt: Date.now() });
  }
  function onPlayerPause() {
    if (isRemoteUpdate.value) {
      isRemoteUpdate.value = false;
      return;
    }
    send({ type: "pause", position: playerRef.value.getCurrentTime(), sentAt: Date.now() });
  }
  function applyMessage(msg) {
    if (!playerRef.value) return;
    isRemoteUpdate.value = true;
    if (msg.type === "play") {
      playerRef.value.seekTo(msg.position);
      playerRef.value.play();
    }
    if (msg.type === "pause") playerRef.value.pause();
    if (msg.type === "load") {
      isRemoteUpdate.value = true;
      playerRef.value.loadVideoById(msg.videoId);
    }
  }
  return { onPlayerPlay, onPlayerPause, applyMessage };
}
