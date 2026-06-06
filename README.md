# WatchRoom

Watch YouTube videos together in sync.

## Features

- Synchronized playback (play / pause / seek)
- Video queue with auto-advance
- Chat
- Late-joiner sync — newcomers start from the current position
- Mobile friendly

## Stack

- **Backend** — Go, gorilla/mux, gorilla/websocket
- **Frontend** — Vue 3, Vite
- **Infra** — Docker, nginx, Let's Encrypt

## Local development

**Backend**
```bash
cd backend
go run ./cmd/server
```

**Frontend**
```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`.

## Production deploy

**1. Get a TLS certificate**

```bash
certbot certonly --standalone -d yourdomain.com
```

**2. Configure environment variables**

```bash
cp .env.example .env
```

Edit `.env`:

```
DOMAIN=yourdomain.com
ALLOWED_ORIGIN=https://yourdomain.com
```

**3. Start**

```bash
docker compose up -d --build
```

## Environment variables

| Variable         | Description                        | Example                      |
|------------------|------------------------------------|------------------------------|
| `DOMAIN`         | Your domain (used for TLS paths)   | `yourdomain.com`             |
| `ALLOWED_ORIGIN` | Frontend origin for WebSocket CORS | `https://yourdomain.com`     |
| `PORT`           | Backend port (default `8080`)      | `8080`                       |
