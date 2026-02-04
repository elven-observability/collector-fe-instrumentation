# Scripts

## install.sh – Linux installer (Faro Collector)

Instala o binário do Faro Collector como serviço systemd, com configuração via env ou interativo.

### Requisitos

- Linux com systemd (Ubuntu/Debian, RHEL/CentOS/Rocky/AlmaLinux/Fedora, Amazon Linux)
- root (sudo)
- curl (o script instala se possível)

### Uso

**Interativo (perguntas no terminal):**

```bash
sudo bash scripts/install.sh
# ou
chmod +x scripts/install.sh && sudo ./scripts/install.sh
```

**Com variáveis de ambiente (CI / automatizado):**

```bash
sudo SECRET_KEY="sua-chave-com-pelo-menos-32-caracteres" \
     LOKI_URL="https://logs-prod-xxx.grafana.net" \
     LOKI_API_TOKEN="seu-token" \
     ALLOW_ORIGINS="https://app.example.com,https://*.example.com" \
     ./scripts/install.sh
```

**Instalar a partir de um binário local (air-gap / release manual):**

```bash
sudo LOCAL_BINARY=/path/to/collector-fe-instrumentation-linux-amd64 ./scripts/install.sh
```

### Variáveis de configuração

| Variável           | Obrigatório | Descrição |
|--------------------|-------------|-----------|
| `SECRET_KEY`       | Sim         | Chave para validar JWT (mín. 32 caracteres) |
| `LOKI_URL`         | Sim         | URL do Loki (ex.: `https://logs-prod-xxx.grafana.net`) |
| `LOKI_API_TOKEN`   | Sim         | Token de API do Loki |
| `ALLOW_ORIGINS`    | Sim         | Origens CORS permitidas (vírgula) |
| `PORT`             | Não         | Porta HTTP (padrão: 3000) |
| `JWT_ISSUER`       | Não         | Issuer esperado no JWT (padrão: trusted-issuer) |
| `JWT_VALIDATE_EXP` | Não         | Validar expiração do JWT: true/false (padrão: false) |

### Variáveis do instalador

| Variável             | Descrição |
|----------------------|-----------|
| `LOCAL_BINARY`       | Caminho do binário local (instala sem download) |
| `GITHUB_REPO`        | Repo no GitHub (padrão: elven/collector-fe-instrumentation) |
| `COLLECTOR_VERSION`  | Tag da release (padrão: latest) |

### Onde fica instalado

- Binário: `/opt/collector-fe-instrumentation/collector-fe-instrumentation`
- Config (env): `/etc/collector-fe-instrumentation/env`
- Serviço: `collector-fe-instrumentation` (systemd)

### Comandos úteis

```bash
systemctl status collector-fe-instrumentation
systemctl restart collector-fe-instrumentation
journalctl -u collector-fe-instrumentation -f
curl http://localhost:3000/health
```
