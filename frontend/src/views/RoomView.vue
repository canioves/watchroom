<template>
  <div class="room">
    <header>
      <span class="room-id"
        >Room: <strong>{{ roomId }}</strong></span
      >
      <span class="status" :class="status">{{ status }}</span>
      <input v-model="videoInput" placeholder="YT URL" />
      <button @click="loadVideo">Load</button>
    </header>

    <div class="placeholder">
      <YoutubePlayer
        ref="player"
        :video-id="currentVideoId"
        :start-position="startPosition"
        :auto-play="autoPlay"
        @play="onPlayerPlay"
        @pause="onPlayerPause"
      />
    </div>

    <aside class="participants">
      <h3>Participants ({{ participants.length }})</h3>
      <ul>
        <li v-for="p in participants" :key="p">
          {{ p === myId ? `${p} (you)` : p }}
        </li>
      </ul>
    </aside>

    <div class="log">
      <p v-for="(entry, i) in log" :key="i">{{ entry }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, useTemplateRef } from "vue";
import { useRoute } from "vue-router";
import { useWebSocket } from "../composables/useWebSocket";
import YoutubePlayer from "../components/YoutubePlayer.vue";
import { usePlayerSync } from "../composables/usePlayerSync.js";

const route = useRoute();
const roomId = route.params.id;

const myId = ref("");
const participants = ref([]);
const log = ref([]);

const videoInput = ref("");
const currentVideoId = ref("");

const startPosition = ref(0);
const autoPlay = ref(false);

const player = useTemplateRef("player");

const { status, connect, disconnect, send } = useWebSocket(roomId, (msg) => {
  const data = JSON.parse(msg);
  if (data.type === "sync") {
    myId.value = data.userId;
    if (data.videoId) {
      startPosition.value = data.position;
      currentVideoId.value = data.videoId;
      if (data.isPlaying) {
        autoPlay.value = true;
      }
    }
    participants.value = [data.userId];
    log.value.push(`[sync] connected as ${data.userId}`);
  } else if (data.type === "join") {
    participants.value.push(data.userId);
    log.value.push(`[join] ${data.userId}`);
  } else if (data.type === "leave") {
    participants.value = participants.value.filter((p) => p !== data.userId);
    log.value.push(`[leave] ${data.userId}`);
  } else if (["play", "pause", "seek", "load"].includes(data.type)) {
    applyMessage(data);
  }
});

const { onPlayerPlay, onPlayerPause, applyMessage } = usePlayerSync(player, send);

function parseVideoId(input) {
  try {
    const url = new URL(input);
    if (url.hostname === "youtu.be") return url.pathname.slice(1);
    return url.searchParams.get("v");
  } catch {
    return input.trim();
  }
}

function loadVideo() {
  const id = parseVideoId(videoInput.value);
  if (id) {
    currentVideoId.value = id;
    send({ type: "load", videoId: id });
  }
}

onMounted(connect);
onUnmounted(disconnect);
</script>

<style scoped>
.room {
  display: grid;
  grid-template-rows: auto 1fr auto;
  grid-template-columns: 1fr 220px;
  gap: 16px;
  padding: 16px;
  height: 100vh;
}
header {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid #222;
}
.room-id {
  font-size: 0.9rem;
  color: #888;
}
.status {
  font-size: 0.8rem;
  padding: 2px 8px;
  border-radius: 4px;
}
.status.connected {
  background: #276749;
}
.status.connecting {
  background: #744210;
}
.status.disconnected {
  background: #63171b;
}
.placeholder {
  background: #1a1a1a;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #555;
}
.participants {
  background: #1a1a1a;
  border-radius: 8px;
  padding: 12px;
}
.participants h3 {
  font-size: 0.85rem;
  color: #888;
  margin-bottom: 8px;
}
.participants ul {
  list-style: none;
}
.participants li {
  font-size: 0.9rem;
  padding: 4px 0;
}
.log {
  grid-column: 1 / -1;
  background: #111;
  border-radius: 6px;
  padding: 8px 12px;
  max-height: 120px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 0.8rem;
  color: #6ee7b7;
}
</style>
