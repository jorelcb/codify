# Phase 1 — Modelo

Una entidad cambia de forma, otra nace, y un enum sustituye a una frase.

---

## Desafío en curso

Una conexión empezada y no terminada.

| | |
|---|---|
| **Cuántos** | **Cero o uno.** No es una regla que se vigila: es la forma del estado (`Option`, no mapa) |
| **Qué guarda** | identificador · vía · conector · tier · etiqueta · a dónde apunta la cuenta |
| **Qué NO guarda** | **la credencial**. Entra con el envío, va al almacén del sistema, y no vuelve |
| **Nace** | al pedir conectar |
| **Muere** | al completarse bien · al cancelarse · al ser reemplazado por uno nuevo |
| **Sobrevive** | a una credencial rechazada, y a cualquier otro fallo (FR-006) |

**Transiciones**

```
        pedir conectar
   ∅ ──────────────────► en curso
   ▲                      │  │  │
   │   cancelar           │  │  │  credencial rechazada
   └──────────────────────┘  │  └──────────────► en curso   (no se pierde)
   ▲                         │
   │   completado bien       │
   └─────────────────────────┘   → la cuenta pasa a conectada

        pedir conectar estando en curso  →  el anterior se reemplaza
```

**La invariante que cambia de sitio.** Hoy «no puede haber dos» sería una comprobación. Con la
forma nueva, dos desafíos a la vez **no se pueden escribir**. Es la diferencia entre impedir un
estado y no poder nombrarlo.

---

## Instrucción de credencial

Lo que el desafío le dice al usuario: dónde encontrar su credencial.

| | |
|---|---|
| **Qué viaja** | un **código estable**, no una frase |
| **Quién lo redacta** | la piel, contra el catálogo, en el idioma elegido |
| **Por qué** | el núcleo sabe *qué* hay que decir; en qué idioma decirlo solo lo sabe quien conoce el idioma. Una frase redactada en el núcleo solo puede estar en uno |
| **Se comprueba** | recorriendo `all()`: todo valor tiene texto en los dos idiomas, o el test no pasa |

Es el patrón que `ProviderIssue`, `SessionFailure` y `ConnectionState` ya usan. Lo nuevo es
**dónde**: la frase que se retira estaba en el backend y nunca llegó a verse, porque la interfaz
ignoraba el campo que la llevaba.

---

## Vía de conexión

Cómo se obtiene el secreto. Lo que cambia entre vías es **cómo se obtiene**, no qué se hace con él:
custodia, uso y revocación son idénticas, y por eso la frontera está donde termina esa diferencia.

| Vía | Qué pide | Alcanzable desde la interfaz |
|---|---|---|
| Credencial directa | una clave, una sola vez | **Sí** — lo que este ciclo entrega |
| Autorización delegada | un código que el usuario lleva fuera | **No: no tiene puerta** |

La segunda está construida y probada desde `003`, y **ningún usuario puede llegar a ella**: la
interfaz fija siempre la primera. Este ciclo no le construye la puerta; lo que hace es que deje de
ser invisible — queda declarada, con su issue, y un test impide que la lista crezca sin que nadie
se entere (FR-012).
