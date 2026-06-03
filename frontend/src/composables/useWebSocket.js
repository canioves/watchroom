import { ref } from 'vue'

export function useWebSocket(roomId, nickname, onMessage, onReconnect) {
  const status = ref('disconnected')
  let ws = null
  let reconnectTimer = null
  let hasConnectedBefore = false

  function connect() {
    status.value = 'connecting'
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    ws = new WebSocket(`${proto}://${location.host}/ws/${roomId}?nickname=${encodeURIComponent(nickname)}`)

    ws.onopen = () => {
      status.value = 'connected'
      if (hasConnectedBefore) onReconnect?.()
      hasConnectedBefore = true
    }
    ws.onmessage = (e) => onMessage(e.data)
    ws.onclose = () => {
      status.value = 'disconnected'
      reconnectTimer = setTimeout(connect, 3000)
    }
    ws.onerror = () => ws.close()
  }

  function disconnect() {
    clearTimeout(reconnectTimer)
    ws?.close()
  }

  function send(msg) {
    if (ws?.readyState === WebSocket.OPEN) {
      console.log(msg)
      ws.send(JSON.stringify(msg))
    }
  }

  return { status, connect, disconnect, send }
}
