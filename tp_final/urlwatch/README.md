# URLWatch

Microservice de vérification d'URLs en masse. Vérifie les URLs en parallèle (pool borné), agrège les résultats et les expose via une API REST.

## Prérequis

- Go 1.22+

## Build & Run

```bash
# Build
go build ./...

# Lancer le serveur (port 8080 par défaut)
go run ./cmd/urlwatch

# Changer le port
PORT=9090 go run ./cmd/urlwatch

# Activer les logs debug
LOG_LEVEL=debug go run ./cmd/urlwatch
```

## Tests

```bash
# Tests simples
go test ./...

# Avec le race detector (recommandé)
go test -race ./...

# Verbose
go test -v -race ./...
```

## Exemples curl

### Créer un lot de vérifications

```bash
curl -s -X POST http://localhost:8080/v1/checks \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://go.dev", "https://pkg.go.dev", "https://exemple.invalid"],
    "options": { "concurrency": 4, "timeout_ms": 3000 }
  }' | jq .
```

Réponse `201 Created` :

```json
{
  "batch_id": "b_4f3c1a",
  "created_at": "2026-06-22T10:00:00Z",
  "summary": { "total": 3, "up": 2, "down": 1, "duration_ms": 812 },
  "results": [
    { "url": "https://go.dev", "status_code": 200, "ok": true, "latency_ms": 120 },
    { "url": "https://pkg.go.dev", "status_code": 200, "ok": true, "latency_ms": 95 },
    { "url": "https://exemple.invalid", "ok": false, "error": "dial tcp: lookup exemple.invalid: no such host", "latency_ms": 501 }
  ]
}
```

### Récupérer un lot par ID

```bash
curl -s http://localhost:8080/v1/checks/b_4f3c1a | jq .
```

Erreur 404 si l'ID est inconnu :

```json
{ "error": { "code": "batch_not_found", "message": "aucun lot avec l'id b_xyz" } }
```

### Health check

```bash
curl -s http://localhost:8080/healthz
# {"status":"ok"}
```

## Structure

```
cmd/urlwatch/       — point d'entrée, câblage des dépendances
internal/domain/    — types métier, interfaces, erreurs
internal/checker/   — HTTPChecker + MockChecker (tests)
internal/pool/      — worker pool borné (fan-out/fan-in)
internal/store/     — store en mémoire (RWMutex)
internal/api/       — handlers HTTP, middleware, routeur
```
