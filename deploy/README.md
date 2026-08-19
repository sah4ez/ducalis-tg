# Deploy — скрипты для выделенной VM

Комплект для деплоя Ducalis на выделенную Linux VM (Ubuntu/Debian).

## Структура

```
deploy/
├── setup.sh      ← однократная подготовка VM (Go, Node, PostgreSQL, БД)
├── env.example   ← шаблон конфигурации (→ deploy/env)
├── deploy.sh     ← сборка + выкладка релиза + перезапуск
├── start.sh      ← запуск сервисов из current
├── stop.sh       ← graceful остановка
├── status.sh     ← состояние (PID, health, память, ошибки)
└── rollback.sh   ← откат на предыдущий релиз
```

Каталог на VM после деплоя:

```
/opt/ducalis/
├── current → releases/20250817-143000   (symlink)
├── releases/
│   ├── 20250817-120000/
│   │   ├── bin/server-public, server-admin, server-internal
│   │   └── web/ (React SPA)
│   └── 20250817-143000/
├── logs/public.log, admin.log, internal.log
└── pids/public.pid, admin.pid, internal.pid
```

## Быстрый старт (с нуля)

```bash
# 1. Клонировать репо на VM
git clone https://github.com/sah4ez/ducalis-tg.git
cd ducalis-tg

# 2. Подготовить VM (root/sudo, ~5 мин)
sudo bash deploy/setup.sh

# 3. Создать конфиг с секретами
cp deploy/env.example deploy/env
nano deploy/env
#   JWT_SECRET=$(openssl rand -hex 32)
#   ADMIN_JWT_SECRET=$(openssl rand -hex 32)
#   INTERNAL_API_KEY=$(openssl rand -hex 32)

# 4. Собрать и запустить (~2 мин)
bash deploy/deploy.sh

# 5. Проверить
bash deploy/status.sh
curl http://localhost:8080/health
```

UI откроется на `http://<VM_IP>:8080`.

## Ежедневные операции

```bash
bash deploy/status.sh              # состояние всех сервисов
bash deploy/deploy.sh              # новый релиз (git pull → deploy.sh)
bash deploy/stop.sh                # остановить всё
bash deploy/start.sh               # запустить из current
bash deploy/rollback.sh            # откат на предпоследний релиз
bash deploy/rollback.sh 20250817-120000   # откат на конкретный
tail -f /opt/ducalis/logs/public.log      # логи
```

## Автозапуск при ребуте (опционально)

systemd-юниты (файл ниже — пример для public, повторить для admin/internal):

```bash
sudo tee /etc/systemd/system/ducalis-public.service <<'EOF'
[Unit]
Description=Ducalis Public API
After=network.target postgresql.service

[Service]
Type=simple
User=ducalis
WorkingDirectory=/opt/ducalis/current
EnvironmentFile=/opt/ducalis/deploy-env
ExecStart=/opt/ducalis/current/bin/server-public
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# deploy-env — плоский KEY=VALUE (без 'source'):
sudo cp deploy/env /opt/ducalis/deploy-env

sudo systemctl daemon-reload
sudo systemctl enable --now ducalis-public
```

При использовании systemd не запускайте `deploy/start.sh` — сервисами управляет systemd:
```bash
sudo systemctl restart ducalis-public   # вместо start/stop.sh
```

## Требования к VM

| Ресурс | Минимум | Рекомендуется |
|---|---|---|
| CPU | 1 vCPU | 2 vCPU |
| RAM | 512 MB | 1 GB |
| Disk | 2 GB | 5 GB |
| OS | Ubuntu 20.04+ / Debian 11+ | Ubuntu 22.04 |

Порты: 8080 (UI+API), 8082 (admin), 8083 (internal). Внешне нужен только 8080; 8082/8083 — за firewall или по VPN.

## Что НЕ нужно на VM

- **Redis** — сервисы его не используют (env-заглушка)
- **Kafka** — аналогично, подключается при необходимости позже
- **Docker** — нативный запуск, но docker-compose.yaml работает если предпочитаете контейнеры
