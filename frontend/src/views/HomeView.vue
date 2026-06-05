<template>
  <div class="home">
    <h1>WatchRoom</h1>

    <div class="card">
      <button :disabled="loading" @click="createRoom">
        {{ loading ? 'Creating...' : 'Create room' }}
      </button>
    </div>

    <div class="card">
      <input v-model="nickname" placeholder="Nickname">
      <input v-model="joinId" placeholder="Room ID" maxlength="8" />
      <button :disabled="!joinId" @click="joinRoom">Join</button>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const joinId = ref('')
const loading = ref(false)
const error = ref('')

const nickname = ref('')

async function createRoom() {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch('/api/rooms', { method: 'POST' })
    if (!res.ok) throw new Error('Server error')
    const { roomId } = await res.json()
    router.push(`/room/${roomId}?nickname=${encodeURIComponent(nickname.value)}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function joinRoom() {
  error.value = ''
  const res = await fetch(`/api/rooms/${joinId.value}`)
  if (!res.ok) {
    error.value = 'Room not found'
    return
  }
  router.push(`/room/${joinId.value}?nickname=${encodeURIComponent(nickname.value)}`)
}
</script>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
  padding-top: 120px;
}
h1 { font-size: 2rem; letter-spacing: 0.05em; }
.card {
  display: flex;
  gap: 8px;
}
input {
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid #333;
  background: #1a1a1a;
  color: #e0e0e0;
  font-size: 1rem;
  width: 160px;
}
button {
  padding: 10px 20px;
  border-radius: 6px;
  border: none;
  background: #e53e3e;
  color: #fff;
  font-size: 1rem;
  cursor: pointer;
}
button:disabled { opacity: 0.5; cursor: default; }
.error { color: #fc8181; }

@media (max-width: 640px) {
  .home {
    padding: 60px 16px 0;
  }
  .card {
    flex-direction: column;
    width: 100%;
    max-width: 320px;
  }
  input {
    width: 100%;
    box-sizing: border-box;
  }
  button {
    width: 100%;
  }
}
</style>
