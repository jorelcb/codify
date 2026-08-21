#!/usr/bin/env bash
# Genera el fixture que exigen los quickstart de `001` y `002`.
#
# Reproduce el patrón que originó el proyecto: un README que **referencia** un documento
# hermano, de modo que quedarse en el README produce una arquitectura inventada. Es el caso
# que la herramienta existe para no fallar.
#
# Uso:  ./scripts/quickstart-fixture.sh [destino]      (por defecto /tmp/codify-fixture)

set -euo pipefail
DEST="${1:-/tmp/codify-fixture}"

rm -rf "$DEST"
mkdir -p "$DEST/docs"

# --- README: apunta al SPEC y a una URL PRIVADA (escenario de referencia no resuelta) ------
cat > "$DEST/README.md" <<'EOF'
# Plataforma de Ejecución de Workers

Servicio que orquesta la ejecución de trabajos de larga duración para el equipo de
plataforma.

La arquitectura de referencia está descrita en [docs/SPEC-30.md](docs/SPEC-30.md).
Las decisiones de despliegue viven en el wiki interno:
https://internal.example.invalid/wiki/spaces/PLAT/pages/9812734/Despliegue

## Cómo correrlo

    make run
EOF

# --- SPEC: aquí está la verdad, y CONTRADICE lo que un modelo asumiría del README ---------
# Un README que habla de "orquestar trabajos de larga duración" invita a suponer un broker
# de mensajes y arquitectura event-sourced. El SPEC dice explícitamente lo contrario: si el
# contexto generado menciona un broker, el agente se lo inventó.
cat > "$DEST/docs/SPEC-30.md" <<'EOF'
# SPEC-30 — Arquitectura de la plataforma de workers

## Decisiones tomadas

- **Motor de orquestación: Temporal.** Se evaluó Kafka y se descartó.
- **NO hay broker de mensajes.** Ni Kafka, ni RabbitMQ, ni SQS. Los workers hablan con
  Temporal directamente.
- **NO es event-sourced.** El estado vive en PostgreSQL como filas mutables; no hay log de
  eventos ni proyecciones.
- Persistencia: PostgreSQL 16.

## Fuera de alcance en F0

- Métricas de negocio: por definir.
- Multi-tenancy.
EOF

# --- Una fuente que CONTRADICE al SPEC, para el escenario de contradicción -----------------
cat > "$DEST/docs/PRD.md" <<'EOF'
# PRD — Plataforma de workers

La persistencia se resolverá con DynamoDB por el volumen esperado.
EOF

cat > "$DEST/Makefile" <<'EOF'
run:
	cargo run
EOF

echo "Fixture creado en: $DEST"
echo
echo "Qué contiene y para qué:"
echo "  README.md            → referencia a docs/SPEC-30.md y a una URL privada (S4 de 001)"
echo "  docs/SPEC-30.md      → la verdad: sin broker, no event-sourced, Temporal (S1 de 001)"
echo "  docs/PRD.md          → contradice al SPEC en la persistencia (S3 de 002)"
echo
echo "La trampa: el README invita a suponer un broker de mensajes. Si el contexto generado"
echo "menciona uno, el agente se lo inventó en vez de seguir la referencia — que es"
echo "exactamente el fallo que este producto existe para no cometer."
