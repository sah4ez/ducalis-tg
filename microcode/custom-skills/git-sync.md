---
name: git-sync
description: >-
  Work on top of the pre-cloned ducalis-tg workspace and sync results back to
  the host. microcode clones https://github.com/sah4ez/ducalis-tg.git (branch
  master) into /workspace via sandbox.sync BEFORE loki starts — do NOT clone/init
  yourself. Commit on vm/<sandbox>, expose via read-only git-daemon on :9418,
  host fetches+merges. Never push to remote from the VM. Use on the first git
  action and on any host↔VM sync.
---

# Git sync — work on clone, push back via daemon

Единственный правильный способ обмена кодом между VM, общим remote и host.
Подробности топологии/branching-модели — `references/git-sync.md` скилла
microcode (в репо github.com/sah4ez/microcode). Здесь — адаптация под ducalis-tg.

## Branching модель (single-VM)

- **remote `master`** — read-only из VM. Никогда не пушить в master из VM.
- **`vm/ducalis-build`** — рабочая ветка loki в VM (где `ducalis-build` =
  `sandbox.name` из манифеста). loki коммитит сюда с checkpoint'ами
  (`loki/session-*` теги внутри).
- **host** — тянет `vm/ducalis-build` через git-daemon, ревьюит, мёрджит в
  свою локальную ветку / делает PR в remote вручную.

## Workspace уже склонирован — НЕ делай git clone/init сам

Клонирование делает **microcode** через секцию `sandbox.sync` в манифесте
(`sync.enabled: true`, `remote_url: https://github.com/sah4ez/ducalis-tg.git`,
`branch: master`). К моменту старта loki `/workspace` — это **полноценный clone**
с shared history, уже на ветке `vm/ducalis-build` (clone-скрипт microcode
выполняет `git checkout -b vm/<sandbox> master`).

**loki НЕ должен:**
- `git clone ...` — уже сделано microcode;
- `git init` — loki-mode сам пропускает init если `/workspace` уже репозиторий
  (loki-mode run.sh:7115), но всё равно не запускайте clone/init в RARV-цикле;
- `git checkout master` и т.п. — работайте ТОЛЬКО на `vm/ducalis-build`.

Проверь при первом действии (read-only, без изменений):

```bash
cd /workspace
git branch --show-current          # должно быть: vm/ducalis-build
git remote -v                      # origin = https://github.com/sah4ez/ducalis-tg.git
git log --oneline -3               # shared history от master
```

Затем — обычная работа + коммиты на `vm/ducalis-build`:

```bash
git add -A && git commit -m "feat: wire transport into main.go"
```

git-daemon УЖЕ запущен в bootstrap (на 0.0.0.0:9418, экспортит /workspace).
Никаких доп. действий для отдачи результата наружу не нужно.

## Тянем результат на хост

```bash
# на хосте, из локального клона ducalis-tg:
git fetch git://localhost:9418/ vm/ducalis-build
git log --oneline FETCH_HEAD -20           # посмотреть что наделал loki
git merge FETCH_HEAD                        # или cherry-pick отдельных коммитов
# затем ревью / тесты / push в origin/master через нормальный flow
```

## sync-конфигурация (декларативная, в манифесте)

Клонирование настраивается в `sandbox.sync` (см. `platform.yaml`/`build.yaml`):

```
enabled: true
remote_url: https://github.com/sah4ez/ducalis-tg.git   (публичный репо)
branch: master
dest: /workspace
depth: 1                                               (shallow clone)
auth: (отсутствует — публичный репо, PAT не нужен)
```

Git-хост `github.com` auto-allowlisted через `sync_egress_rules()` — отдельная
запись в `network.allow` не нужна. Mount для `sync.dest` (`/workspace`)
подавляется автоматически; overlay-skills (`/workspace/skills`) и PRD
(`/workspace/PRD.md`) приходят через отдельные mounts и clone их не затирает.

**Если репо станет приватным** — раскомментируй `sync.auth` в манифесте:
```yaml
sync:
  auth:
    method: https
    token_env: GH_TOKEN        # host env var с PAT (значение НЕ инлайнится)
```
и задай на хосте `export GH_TOKEN=ghp_...`.

## Чего НЕ делать

- **НЕ** `git push` из VM в remote (master — read-only из VM; host ревьюит и пушит).
- **НЕ** `git init` в `/workspace`, если есть remote — клонируй, переиспользуй history.
- **НЕ** бинди git-daemon на 127.0.0.1 — только 0.0.0.0 (msb port-forward → eth0).
- **НЕ** коммить сгенерированный транспорт (`internal/transport/`, `pkg/client/`)
  без `make generate` — если регенерировал, коммить отдельным чистым коммитом.
