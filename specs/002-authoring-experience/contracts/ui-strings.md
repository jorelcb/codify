# Contract — Catálogo de cadenas y localización

Materializa FR-016b y hace **verificable** SC-009 ("cero cadenas sin traducir").

## Dónde vive

En **Rust**: `crates/codify-app/src/strings.rs`. No en archivos JSON del frontend.

El motivo es el criterio de éxito: SC-009 exige cero cadenas sin traducir, y eso solo se puede *demostrar* si el catálogo es un dato que un test puede recorrer. Si las cadenas viven sueltas en el DOM, el criterio degrada a una revisión visual — exactamente lo que la constitución evita.

## Forma

```rust
pub struct UiStrings {
    pub locale: Locale,              // Es | En
    pub entries: BTreeMap<&'static str, &'static str>,
}

pub fn strings_for(locale: Locale) -> UiStrings;
pub fn system_locale() -> Locale;    // idioma del SO; cae a En si no es es/en
```

Se expone al frontend con el comando `ui_strings(locale)`.

## Reglas de contrato

1. **Paridad total**: el conjunto de claves de `es` y el de `en` son **idénticos**. Es el test que sostiene SC-009.
2. **Sin claves huérfanas**: toda clave del catálogo se usa en la interfaz, y toda cadena de la interfaz sale del catálogo.
3. **Claves con espacio de nombres** por superficie: `stream.*`, `artifact.*`, `provider.*`, `session.*`, `error.*`, `a11y.*`.
4. **Independiente del idioma de los artefactos**: cambiar la interfaz a inglés **no** cambia el idioma en que se genera el contexto, ni al revés (FR-016b).
5. **Cambiar de idioma no exige reiniciar**: la interfaz se vuelve a pintar con el catálogo nuevo.
6. Los textos de accesibilidad (`a11y.*`: etiquetas de región, anuncios de la corriente) son parte del catálogo, no un añadido posterior.
7. **El núcleo no redacta**: cuando el núcleo tiene que explicar algo, devuelve un **código** y la piel elige la frase (`provider.issue.<code>`). Una frase ya escrita en el núcleo tendría idioma fijo y volvería la regla 2 indemostrable — además de ser presentación colándose en la aplicación.
8. **Un solo dueño por texto**: si el JavaScript le escribe el `textContent` a un elemento, ese elemento **no lleva `data-i18n`**. Con dos dueños, `apply()` y el código se pisan al cambiar de idioma; quien pinta a mano es responsable de repintar.
9. **`hidden` gana siempre**: hace falta la regla global `[hidden] { display: none !important; }`. El atributo pone `display: none` desde la hoja del navegador y **cualquier** `display` de autor lo pisa — un panel con `display: flex` sigue visible con `hidden` puesto, y el código que lo oculta no hace nada.
10. **Quien pinta a mano expone `render()`** y el manejador de cambio de idioma lo llama. `apply()` solo alcanza al DOM marcado: sin esto, todo lo pintado con `textContent` se queda en el idioma en que se pintó. Única excepción, `stream.js`: sus bloques son **append-only** porque espejan el log de auditoría, y reescribir uno ya emitido sería falsificar lo que pasó.

## Claves mínimas por superficie

| Espacio | Cubre |
|---|---|
| `session.*` | estados, iniciar, cancelar, balance de escrituras |
| `stream.*` | rótulos de cada tipo de bloque (leído, no resuelto, contradicción, escrito, salida bloqueada) |
| `artifact.*` | vista de artefacto, **etiquetas de los tres estados de fundamento**, fuentes, motivo |
| `provider.*` | estado del backend, elección de modelo, **qué hacer cuando falta algo** |
| `error.*` | fallos accionables (FR-028) |
| `a11y.*` | nombres de región, anuncios de la corriente, texto de foco |

> Las etiquetas de `artifact.*` para grounded / tentativo / contradicción son las que garantizan que la distinción **no dependa del color** (FR-026): son la señal textual de las tres redundantes.

## Verificación

- **Test de paridad**: `strings_for(Es).entries.keys() == strings_for(En).entries.keys()`. Falla el build si alguien añade una cadena en un solo idioma.
- **Test de no vacías**: ningún valor es cadena vacía.
- **Test de códigos de estado**: cada variante de `SessionState` tiene su `session.state.<code>` en ambos idiomas. Cierra un acoplamiento que estaba sin vigilar por los dos lados — la clave se compone en una plantilla, así que el extractor de `t("literal")` no la veía.
- **Tests del contrato de interfaz** (`crates/codify-app/tests/ui_contract.rs`): que ninguna clave usada falte del catálogo y que ninguna clave del catálogo sobre; que ningún texto visible esté escrito directamente en el HTML; que ningún elemento tenga dos dueños de su texto; y que **todo `ProviderIssue` tenga frase en ambos idiomas**.
- **Manual (quickstart)**: recorrer la aplicación en cada idioma. Sigue teniendo valor —nadie automatiza «se entiende»— pero ya no es lo único: la primera pasada real encontró tres textos que se quedaban en el idioma equivocado, y los tests de arriba existen para que no vuelvan.
