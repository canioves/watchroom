import { ref } from 'vue'

export function useWebSocket(roomId, onMessage) {
  const status = ref('disconnected')
  let ws = null
  let reconnectTimer = null

  function connect() {
    status.value = 'connecting'
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    ws = new WebSocket(`${proto}://${location.host}/ws/${roomId}`)

    ws.onopen = () => { status.value = 'connected' }
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
      ws.send(JSON.stringify(msg))
    }
  }

  return { status, connect, disconnect, send }
}
