# Contracts — Tool-schemas del agente (loop D4/D5)

Herramientas que el loop (`application/authoring_loop.rs`) expone al LLM vía `ModelProvider`. El LLM decide cuáles llamar; el loop las ejecuta contra los ports y devuelve el resultado. Schemas JSON (tool-use).

## Herramientas de ingesta (D4)

```jsonc
// list_repo — lista entradas de un directorio del repo (acotado por presupuesto)
{ "name": "list_repo", "input": { "path": "string (repo-relative, '' = raíz)" },
  "returns": "[{ name, kind: 'file'|'dir', size }]" }

// read_file — lee un archivo del repo
{ "name": "read_file", "input": { "path": "string (repo-relative)" },
  "returns": "{ content, truncated: bool }" }   // truncated => declarar lo omitido, no en silencio

// fetch_url — trae una referencia externa PÚBLICA (v1: sin auth)
{ "name": "fetch_url", "input": { "url": "string (http/https público)" },
  "returns": "{ content } | { unresolved: 'inaccessible'|'requires_auth'|'out_of_scope' }" }
```

## Herramientas de authoring

```jsonc
// propose_change — emite un ChangeProposal (el core hace el diff real y clasifica riesgo)
{ "name": "propose_change",
  "input": { "target": "artifact_kind|repo_path", "new_content": "string",
             "rationale": "string", "groundedness": "[{span, sources[]|tentative_reason}]" },
  "returns": "{ proposal_id, risk: 'low'|'high_impact', applied: bool }" }

// ask_user — hace una pregunta al usuario en el loop conversacional (surface vía Prompter)
{ "name": "ask_user", "input": { "question": "string", "suggestions?": ["string"] },
  "returns": "{ answer }" }

// note_unresolved — registra una referencia/asunto no resuelto (nunca inventar; FR-004/SC-006)
{ "name": "note_unresolved", "input": { "what": "string", "reason": "string" }, "returns": "ack" }

// finalize — declara el contexto listo para revisión/cierre
{ "name": "finalize", "input": { "summary": "string" }, "returns": "ack" }
```

## Reglas del loop
- El loop impone un **presupuesto de exploración** (nº de lecturas/fetches) y, al terminar, **declara qué no leyó** (sin truncamiento silencioso).
- `propose_change` de bajo riesgo se **auto-aplica** y se muestra (revertible); alto impacto **espera** `decide` del usuario (D6/FR-012).
- El loop **jamás** deja que el modelo fabrique contenido de una referencia no resuelta: solo `note_unresolved`.
- En modo local, `fetch_url` a hosts no-locales está **deshabilitado** (cero-egress estructural) → `unresolved: 'out_of_scope'` + evento `egress.blocked`.
