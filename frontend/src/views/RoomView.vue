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
        <button class="queue-btn" @click="addToQueue" title="Add to queue">+</button>
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
        @seek="onPlayerSeek"
        @ended="onPlayerEnded"
      />
    </div>
    <div class="queue-area" v-if="queue.length">
      <button class="queue-skip" @click="skipNext">Next</button>
      <div class="queue-list">
        <div v-for="(item, i) in queue" :key="item.id + i" class="queue-item">
          <img :src="`https://img.youtube.com/vi/${item.id}/mqdefault.jpg`" alt="" />
          <span class="queue-num">{{ i + 1 }}</span>
        </div>
      </div>
    </div>
    <aside class="sidebar">
      <div class="participants">
        <h3>Participants ({{ participants.length }})</h3>
        <ul>
          <li v-for="p in participants" :key="p.userId">
            {{ p.userId === myId ? `${p.nickname} (you)` : p.nickname }}
          </li>
        </ul>
      </div>
      <div class="chat">
        <div class="chat-messages" ref="chatEl">
          <div v-for="m in messages" :key="m.id" class="chat-msg">
            <span class="chat-nick">{{ m.nickname }}</span>
            <span class="chat-text">{{ m.text }}</span>
          </div>
        </div>
        <div class="chat-input">
          <input v-model="chatInput" placeholder="Message..." @keydown.enter="submitChat" />
          <button @click="submitChat">→</button>
        </div>
      </div>
    </aside>

    <div class="log">
      <p v-for="(entry, i) in log" :key="i">{{ entry }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, onMounted, onUnmounted, useTemplateRef } from "vue";
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

const messages = ref([]);
const chatInput = ref("");
const chatEl = useTemplateRef("chatEl");

const queue = ref([]);

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
      autoPlay.value = true;
      startPosition.value = 0;
      currentVideoId.value = data.videoId;
    } else if (data.type === "chat") {
      messages.value.push({ id: Date.now(), nickname: data.nickname, text: data.text });
      nextTick(() => {
        if (chatEl.value) chatEl.value.scrollTop = chatEl.value.scrollHeight;
      });
    } else if (data.type === "queue_add") {
      queue.value.push({ id: data.videoId });
    } else if (data.type === "queue_advance") {
      if (currentVideoId.value !== data.videoId) {
        queue.value.shift();
        autoPlay.value = true;
        startPosition.value = 0;
        currentVideoId.value = data.videoId;
      }
    } else if (["play", "pause", "seek"].includes(data.type)) {
      applyMessage(data);
    }
  },
  () => {
    participants.value = [];
    myId.value = "";
  },
);

const { onPlayerPlay, onPlayerPause, onPlayerSeek, applyMessage } = usePlayerSync(player, send);

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
    autoPlay.value = true;
    startPosition.value = 0;
    currentVideoId.value = id;
    send({ type: "load", videoId: id });
  }
}

function submitChat() {
  const text = chatInput.value.trim();
  if (!text) return;
  messages.value.push({ id: Date.now(), nickname, text });
  nextTick(() => {
    if (chatEl.value) chatEl.value.scrollTop = chatEl.value.scrollHeight;
  });
  send({ type: "chat", text });
  chatInput.value = "";
}

function addToQueue() {
  const id = parseVideoId(videoInput.value);
  queue.value.push({ id });
  send({ type: "queue_add", videoId: id });
  videoInput.value = "";
}

function onPlayerEnded() {
  if (queue.value.length) skipNext();
}

function skipNext() {
  if (!queue.value.length) return;
  const next = queue.value[0].id;
  send({ type: "queue_advance", videoId: next });
  send({ type: "load", videoId: next });
  queue.value.shift();
  autoPlay.value = true;
  startPosition.value = 0;
  currentVideoId.value = next;
}

onMounted(connect);
onUnmounted(disconnect);
</script>

<style scoped>
.room {
  display: grid;
  grid-template-rows: auto 1fr auto auto;
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
.queue-btn {
  background: #2d2d2d !important;
  border: 1px solid #444 !important;
  color: #ccc !important;
  font-size: 1.1rem !important;
  padding: 6px 12px !important;
}
.queue-btn:hover { background: #3a3a3a !important; }
.player-area {
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  min-height: 0;
  grid-row: 2;
  grid-column: 1;
}
.queue-area {
  grid-row: 3;
  grid-column: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  background: #111;
  border-radius: 8px;
  padding: 8px 12px;
  min-width: 0;
}
.queue-skip {
  flex-shrink: 0;
  padding: 6px 12px;
  border-radius: 6px;
  border: none;
  background: #2d2d2d;
  color: #ccc;
  font-size: 0.85rem;
  cursor: pointer;
  white-space: nowrap;
}
.queue-skip:hover { background: #3a3a3a; }
.queue-list {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  flex: 1;
  padding-bottom: 2px;
}
.queue-item {
  position: relative;
  flex-shrink: 0;
  cursor: pointer;
}
.queue-item img {
  width: 120px;
  height: 68px;
  object-fit: cover;
  border-radius: 4px;
  display: block;
  border: 1px solid #333;
}
.queue-num {
  position: absolute;
  top: 4px;
  left: 4px;
  background: rgba(0,0,0,0.7);
  color: #fff;
  font-size: 0.7rem;
  padding: 1px 5px;
  border-radius: 3px;
}
.empty-player {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #555;
  font-size: 0.95rem;
}
.sidebar {
  background: #1a1a1a;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  grid-row: 2 / 4;
  grid-column: 2;
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
.chat {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  margin-top: 12px;
  border-top: 1px solid #2d2d2d;
  padding-top: 10px;
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
}
.chat-msg {
  display: flex;
  flex-direction: column;
  gap: 1px;
}
.chat-nick {
  font-size: 0.7rem;
  color: #888;
}
.chat-text {
  font-size: 0.85rem;
  color: #e0e0e0;
  word-break: break-word;
}
.chat-input {
  display: flex;
  gap: 6px;
  margin-top: 8px;
  flex-shrink: 0;
}
.chat-input input {
  flex: 1;
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid #333;
  background: #111;
  color: #e0e0e0;
  font-size: 0.85rem;
  min-width: 0;
}
.chat-input button {
  padding: 6px 12px;
  border-radius: 6px;
  border: none;
  background: #e53e3e;
  color: #fff;
  cursor: pointer;
  font-size: 1rem;
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

@media (max-width: 640px) {
  .room {
    grid-template-columns: 1fr;
    grid-template-rows: auto auto auto auto;
    height: auto;
    min-height: 100dvh;
    padding: 12px;
    gap: 10px;
  }
  header {
    flex-wrap: wrap;
    row-gap: 10px;
  }
  .load-form {
    margin-left: 0;
    width: 100%;
  }
  .load-form input {
    flex: 1;
    width: auto;
  }
  .player-area {
    aspect-ratio: 16 / 9;
    min-height: unset;
  }
  .sidebar {
    max-height: 260px;
  }
  .participants ul {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 12px;
  }
  .log {
    max-height: 80px;
  }
}
</style>
