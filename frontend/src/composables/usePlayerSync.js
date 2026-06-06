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

  function onPlayerSeek(time) {
    if (isRemoteUpdate.value) {
      isRemoteUpdate.value = false;
      return;
    }
    send({ type: "seek", position: time });
  }

  function applyMessage(msg) {
    if (!playerRef.value) return;
    isRemoteUpdate.value = true;
    if (msg.type === "play") {
      playerRef.value.seekTo(msg.position);
      playerRef.value.play();
    }
    if (msg.type === "pause") {
      playerRef.value.pause();
    }
    if (msg.type === "seek") {
      playerRef.value.setLastKnownTime(msg.position);
      playerRef.value.seekTo(msg.position);
      isRemoteUpdate.value = false;
    }
  }
  return { onPlayerPlay, onPlayerPause, onPlayerSeek, applyMessage };
}
