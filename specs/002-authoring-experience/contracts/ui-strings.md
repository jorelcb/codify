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
- **Manual (quickstart)**: recorrer la aplicación en cada idioma y confirmar que no aparece ninguna clave cruda ni texto en el idioma equivocado.
