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
             "rationale": "string",
             // FR-006a: `sources` NO basta. Cada span grounded exige `quotes`: fragmentos
             // TEXTUALES de la fuente. El núcleo comprueba que aparecen en el material leído
             // y degrada a tentativo lo que no se sostiene (FR-006c).
             "groundedness": "[{span, sources[]+quotes[]|tentative_reason}]" },
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

## Regla que gobierna la procedencia (FR-006a/b/c)

**Declarar una fuente no la verifica.** El núcleo comprueba que la cita textual aparece en el
material leído; si no, el span **se degrada a tentativo** con el motivo — no se descarta.

Aplica igual a las contradicciones: afirmar que dos fuentes chocan exige una cita comprobable
**de cada una**. Sin ellas, la contradicción no se afirma.

### Cómo se compara la cita

La comparación **normaliza mayúsculas y espacios** —cualquier racha de espacios, tabuladores o
saltos de línea cuenta como uno—, y nada más. El modelo reproduce el sentido de una frase, no su
maquetación: exigir el byte exacto haría que un salto de línea de más invalidara una cita
legítima. Lo que **no** se normaliza son las palabras; sin eso, la comprobación dejaría de
distinguir lo que la fuente dice de lo que se le atribuye.

Dos consecuencias que conviene tener presentes:

- una cita por debajo de **12 caracteres** no verifica nada: un fragmento así aparece en
  cualquier texto;
- el identificador de la fuente se empareja con tolerancia (`docs/SPEC-30.md` vale por
  `SPEC-30.md`), porque el modelo abrevia rutas. Es tolerancia sobre **qué** fuente, nunca sobre
  qué dice.

> Nace de un fallo real (2026-08-23): el sistema registró «[PRD vs Makefile] el Makefile solo
> soporta PostgreSQL 16» sobre un `Makefile` de dos líneas que no menciona PostgreSQL. La fuente
> **sí se había leído** — lo inventado era lo que se le atribuía, así que comprobar «¿se leyó?»
> no habría bastado.
