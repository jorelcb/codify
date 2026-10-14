# Phase 1 — Modelo

Tres conceptos, y uno de ellos existe para que una afirmación del producto se pueda sostener.

---

## Modelo disponible

Algo con lo que se podría trabajar.

| | |
|---|---|
| **Nombre** | Lo que el usuario reconoce |
| **Procedencia** | De dónde salió. **Obligatoria** |
| **Distancia** | Cuánto falta para poder usarlo |
| **Dónde infiere** | Dentro del equipo, o fuera |

**La procedencia no es decorado.** El principio rector del proyecto es que lo que no se puede
verificar contra una fuente no se afirma. Una lista de modelos que aparece sola, sin decir de dónde
salió cada uno, es exactamente la clase de afirmación que este producto existe para no hacer.

---

## Distancia

Cuánto falta para que un modelo se pueda usar. **Ordenada**, porque lo más cerca se ofrece primero.

| | Qué significa | Qué se le dice al usuario |
|---|---|---|
| **Usable** | Responde ahora mismo | nada: se usa |
| **Instalado** | Está en la máquina y no en marcha | qué falta para arrancarlo |
| **Ausente** | No está | qué podría instalar |

«Ausente» existe para que el caso de quien abre esto por primera vez **tenga algo que decir** en
vez de una dirección que no significa nada.

---

## Dónde infiere

**Es lo que clasifica a un proveedor, y no es lo mismo que dónde corre el programa.**

| | Proceso | Inferencia |
|---|---|---|
| Un servidor local | en la máquina | **en la máquina** |
| Un agente de línea de órdenes de un servicio remoto | en la máquina | **fuera** |

Los dos caminos existen a propósito: el producto quiere integrarse con modelos que no sacan nada
del equipo **y** con modelos de inferencia remota. **El defecto no es tener el segundo — es
presentarlo como el primero.**

**Cómo se sostiene.** Un modelo se marca como «infiere dentro» solo si llegó por una fuente que
**no pudo** devolver otra cosa: el sondeo de loopback falla al construirse con una dirección que no
sea del propio equipo. Donde eso no se puede sostener, la clasificación es una **declaración**, y
entonces el modelo se enseña pero **no se cablea como local**.

---

## Fuente de detección

Un sitio donde mirar. Cada una sabe buscar una clase de cosa y **declara lo que encuentra**.

| Fuente | Qué busca | Distancia que produce | ¿Sostiene la clasificación? |
|---|---|---|---|
| **Servidores** | puertos de loopback conocidos | Usable | **Sí**, por construcción |
| **Programas** | el `PATH` | Instalado | No: declara |

**Ninguna fuente ejecuta lo que encuentra** ([research D3](./research.md)). Mirar el `PATH` dice
que algo está; invocarlo para confirmarlo sería lanzar un binario arbitrario elegido por su nombre.

**La lista de qué buscar es conocida y declarada.** Encontrar algo nuevo exige enseñárselo, y eso
se dice en vez de aparentar que se descubre solo.
