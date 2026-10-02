# Phase 0 — Investigación

Seis decisiones. La tercera y la cuarta son las que tienen consecuencias de seguridad, y las dos se
resuelven igual: **no hacer lo que no hace falta**.

---

## D1 · Dónde vive la detección

**Decisión.** Un **port de capacidad** en `application/`, con un adapter por fuente en
`infrastructure/deteccion/`. Lo que cruza el port son tipos propios, no rutas ni cadenas.

**Por qué ahí y no en el dominio.** La constitución separa los dos casos con una pregunta: *¿lo
nombra el Dominio?* El dominio de este producto habla de referencias, procedencia y contradicción.
**No habla de detectar**: eso es una capacidad que la aplicación necesita, como el almacén de
credenciales o el reloj.

**Por qué no dentro de `providers/`.** Detectar y proveer son cosas distintas: una encuentra
candidatos, la otra habla con ellos. Mezclarlas haría que «un proveedor» signifique dos cosas —y
este proyecto ya tuvo ese problema exacto con la palabra «proveedor» (#55).

---

## D2 · Cómo se encuentra un servidor en marcha

**Decisión.** Sondear un conjunto **declarado** de puertos de loopback con `LocalProviderProbe`,
que ya existe.

**Por qué reutilizarla y no escribir otra.** Porque **rechaza por construcción** cualquier
dirección que no sea del propio equipo:

```rust
if !LocalOpenAiCompatProvider::is_loopback(&base_url) {
    return Err(CoreError::EgressBlocked(...))
}
```

Eso es justo lo que FR-010a pide: una clasificación que se **sostiene**, no que se declara. Un
servidor encontrado por esta vía es local porque no pudo ser otra cosa.

Y además ya sabe decir por qué algo no sirve —`NotListening`, `NoModels`— con códigos estables que
la piel traduce. FR-009 pide decir **qué falta**, y la mitad del vocabulario ya está.

---

## D3 · No se ejecuta nada de lo que se encuentra — **decisión de seguridad**

**El problema.** Para saber que un `llama-server` del `PATH` es lo que dice, lo natural es
invocarlo: `llama-server --version`. Y eso es **ejecutar un binario arbitrario encontrado en el
`PATH` de la máquina**, elegido por su nombre.

**Decisión.** No se ejecuta. La detección de programas instalados reporta **presencia y ruta**, y
nada más.

**Por qué.** Lo único que hace falta saber es «está instalado», y eso lo da el `PATH`. Confirmar la
identidad invocándolo compra muy poco —un nombre más preciso en una lista— a cambio de un riesgo
que no controlamos: cualquier cosa puede llamarse `ollama` y estar antes en el `PATH`.

Es el mismo razonamiento que llevó a no leer el llavero del operador para verificar custodia: **la
comprobación cara no era necesaria para la pregunta que había que responder**, y hacerla enseñaba
un hábito que no queremos enseñar.

**Qué se pierde, dicho en voz alta.** No se puede distinguir un binario legítimo de uno que se
llama igual, ni saber su versión. Lo primero lo resuelve el usuario al usarlo; lo segundo no hace
falta para nada de este spec.

**Alternativa descartada.** Invocar con `--version` y un límite de tiempo. **Señal para
revisitarla:** que aparezca un requisito que de verdad necesite la versión, y entonces se decide
con ese requisito delante.

---

## D4 · Qué se cablea y qué solo se enseña

**La distinción que ordena todo**: lo que clasifica a un proveedor **no es dónde corre el programa,
sino dónde ocurre la inferencia**.

| | Proceso | Inferencia | ¿Sale contenido? |
|---|---|---|---|
| `llama-server`, `lms` | local | **local** | no |
| `claude` | local | **remota** | sí |

**Decisión.** Se **detecta todo**. Se **cablea como local** únicamente lo que llega por una vía que
sostiene esa propiedad —hoy, el sondeo de loopback de D2—. Lo que no puede sostenerla se enseña, se
nombra por lo que es, y **no se ofrece como local**.

**Por qué no basta declararlo.** Una lista que diga «`claude` es remoto» funciona mientras alguien
la mantenga. Una clasificación que sale de un constructor que rechaza lo que no es loopback
funciona **aunque nadie se acuerde**. Donde se puede tener lo segundo, no se acepta lo primero.

**Lo que esto NO significa.** Que un proveedor de inferencia remota sea un problema. **No lo es**:
el producto quiere los dos caminos, y la constitución acota el cero-egress estructural al camino
local. El defecto sería presentar como local algo que no lo es.

**Alternativa descartada.** Un adapter de subproceso que permita **usar** un CLI de inferencia
remota. Es trabajo real y deseable, y necesita decidir protocolo, tiempos y streaming: ciclo
propio. Este encuentra y nombra, que ya es más de lo que hay.

---

## D5 · Que detectar no cueste la apertura

**Decisión.** Las fuentes corren **a la vez**, con un presupuesto común de **3 s**. La interfaz se
pinta antes de que terminen y declara que está mirando. Lo que no llega a tiempo se abandona y se
dice.

**Por qué 3 s.** Es un número elegido, no medido: por encima se nota al abrir, y por debajo se
corre el riesgo de descartar un servidor que tarda en contestar. Queda declarado como elección en
el spec para que nadie lo lea como un dato.

**Por qué no se guarda lo detectado.** Sería una cuarta fuente de verdad sobre un hecho que cambia
solo —alguien arranca o para un servidor entre dos aperturas— y este proyecto ya ha pagado dos
veces por tener dos sitios que afirman lo mismo (#53, #55).

---

## D6 · Qué pasa con la dirección por defecto

**Decisión.** `DEFAULT_ENDPOINT` —`http://localhost:11434`— **desaparece**. La variable de entorno
sobrevive solo como anulación de desarrollo, declarada como tal, y pasa por el mismo sondeo que
todo lo demás: sin trato de favor.

**Por qué.** Enseñar una dirección hacia un servicio que puede no estar instalado es **afirmar algo
sin comprobarlo**, en un producto cuya tesis es que eso no se hace. En la máquina donde se escribió
este spec, esa dirección apunta a un Ollama que no existe, mientras hay cuatro cosas usables en el
`PATH`.

Y pedirle al usuario una variable de entorno para usar algo que ya tiene instalado es trasladarle
el trabajo de descubrir. **Ninguna herramienta comparable lo hace** — lo dijo el operador, y es
comprobable.
