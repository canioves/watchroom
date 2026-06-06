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
cp .env.example .env
go run ./cmd/server
```

**Frontend**
```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Open `http://localhost:5173`.

## Production deploy

```bash
cp .env.example .env
# edit .env — set ALLOWED_ORIGIN=https://yourdomain.com
```

Get a Let's Encrypt certificate:
```bash
certbot certonly --standalone -d yourdomain.com
```

Update `frontend/nginx.conf` — replace `YOURDOMAIN` with your domain.

```bash
docker compose up -d --build
```

## Environment variables

| Variable         | Description                        | Example                      |
|------------------|------------------------------------|------------------------------|
| `PORT`           | Backend port                       | `8080`                       |
| `ALLOWED_ORIGIN` | Frontend origin for WebSocket CORS | `https://yourdomain.com`     |
