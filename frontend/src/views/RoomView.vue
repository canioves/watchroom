<template>
  <div class="room">
    <header>
      <span class="room-id">
        <strong>{{ roomId }}</strong>
        <button class="copy-btn" @click="copyLink" :title="copied ? 'Copied!' : 'Copy invite link'">
          {{ copied ? "✓" : "Copy link" }}
        </button>
      </span>
      <span class="status" :class="status">{{ status }}</span>
      <div class="load-form">
        <input v-model="videoInput" placeholder="YouTube URL or ID" @keydown.enter="loadVideo" />
        <button @click="loadVideo">Load</button>
      </div>
    </header>

    <div class="player-area">
      <div v-if="!currentVideoId" class="empty-player">Paste a YouTube URL above to start watching</div>
      <YoutubePlayer
        v-else
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
        <li v-for="p in participants" :key="p.userId">
          {{ p.userId === myId ? `${p.nickname} (you)` : p.nickname }}
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
const nickname = route.query.nickname || "anon";

const myId = ref("");
const participants = ref([]);
const log = ref([]);

const videoInput = ref("");
const currentVideoId = ref("");

const startPosition = ref(0);
const autoPlay = ref(false);

const player = useTemplateRef("player");

const copied = ref(false);
function copyLink() {
  navigator.clipboard.writeText(roomId);
  copied.value = true;
  setTimeout(() => (copied.value = false), 2000);
}

const { status, connect, disconnect, send } = useWebSocket(
  roomId,
  nickname,
  (msg) => {
    const data = JSON.parse(msg);
    if (data.type === "sync") {
      myId.value = data.userId;
      if (data.videoId) {
        startPosition.value = data.position;
        if (data.isPlaying) {
          autoPlay.value = true;
        }
        currentVideoId.value = data.videoId;
      }
      participants.value = [{ userId: data.userId, nickname: data.nickname }];
      log.value.push(`[sync] connected as ${data.nickname}`);
    } else if (data.type === "join") {
      participants.value.push({ userId: data.userId, nickname: data.nickname });
      log.value.push(`[join] ${data.nickname}`);
    } else if (data.type === "leave") {
      participants.value = participants.value.filter((p) => p.userId !== data.userId);
      log.value.push(`[leave] ${data.nickname}`);
    } else if (data.type === "load") {
      autoPlay.value = true
      startPosition.value = 0
      currentVideoId.value = data.videoId
    } else if (["play", "pause", "seek"].includes(data.type)) {
      applyMessage(data);
    }
  },
  () => {
    participants.value = [];
    myId.value = "";
  },
);

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
    autoPlay.value = true
    startPosition.value = 0
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
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  color: #888;
}
.copy-btn {
  padding: 2px 10px;
  font-size: 0.75rem;
  background: #2d2d2d;
  border: 1px solid #444;
  border-radius: 4px;
  color: #ccc;
  cursor: pointer;
}
.copy-btn:hover {
  background: #3a3a3a;
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
.load-form {
  display: flex;
  gap: 8px;
  margin-left: auto;
}
.load-form input {
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid #333;
  background: #1a1a1a;
  color: #e0e0e0;
  font-size: 0.9rem;
  width: 260px;
}
.load-form button {
  padding: 6px 16px;
  border-radius: 6px;
  border: none;
  background: #e53e3e;
  color: #fff;
  font-size: 0.9rem;
  cursor: pointer;
}
.player-area {
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  min-height: 0;
}
.empty-player {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #555;
  font-size: 0.95rem;
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
