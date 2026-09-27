# Phase 0 — Investigación

Seis decisiones. Tres de ellas no estaban en el spec: **aparecieron al recorrer el camino entero**,
que es lo que nadie había hecho desde `003`.

---

## D1 · La instrucción cruza como código, no como frase

**Decisión.** `Desafio::PideCredencial` deja de llevar `instrucciones: String` y pasa a llevar un
enum de dominio, `InstruccionDeCredencial`, con `code()` y `all()`. La piel traduce el código contra
el catálogo.

**Por qué este patrón y no otro.** No se inventa nada: es el que este proyecto usa ya tres veces
—`ProviderIssue`, `SessionFailure`, `ConnectionState`— y cada uno trae de regalo el test que
recorre `all()` y exige texto en los dos idiomas. Un valor nuevo sin traducir **no compila el
test**, en vez de aparecer en pantalla en el idioma equivocado.

**Lo que quita.** Este literal, hoy dentro de `connect_provider`:

```rust
DirectCredential::new("Pega tu credencial: se guarda en el almacén del sistema y no vuelve a mostrarse.")
```

Una frase en español, en el backend, que no se ve **solo porque la interfaz ignora ese campo**. El
test que persigue cadenas sueltas mira `ui/` y `strings.rs`: nunca iba a encontrarla.

**Alternativas descartadas.**

- *Que la piel ignore el campo y use su propia cadena.* Deja el campo muerto — exactamente lo que
  este spec vino a eliminar.
- *Que el núcleo reciba el idioma y devuelva la frase.* Mete el catálogo en el núcleo, que hoy vive
  entero en la piel, y obliga a pasar el idioma por firmas que no lo necesitan.
- *Quitar el campo y fijar una instrucción por vía.* Funciona hoy y se rompe en cuanto un proveedor
  necesite decir algo propio. **Señal para revisitarlo:** que `all()` se quede con un solo valor
  durante varios ciclos.

---

## D2 · Cómo se comprueba que no queda un camino sin puerta

**El problema.** La autorización delegada está construida, probada y **es inalcanzable**: la
interfaz manda siempre `delegada: false`. No es un comando huérfano —el test de `004` cazaría
eso— sino **una rama dentro de un comando**, elegida por un parámetro que la interfaz nunca varía.

**Decisión.** Un test que recoge, para cada selector de rama conocido, **los valores literales que
la interfaz puede enviar**. Si una rama no es alcanzable con ninguno, debe estar en una lista de
**caminos declarados sin puerta, con su issue al lado**. Misma forma que la lista de comandos sin
invocar que `004` estrenó: deuda con nombre, no permiso.

**Lo que NO cubre, dicho aquí y no descubierto en la revisión.** Caza *esta clase* de camino sin
puerta: un selector que la piel fija. No caza código inalcanzable en general, y no descubre
selectores que nadie le haya enseñado. Es barato, determinista y ataca el caso real; pretender más
sería la misma afirmación sin respaldo que este spec viene a corregir.

**Alternativas descartadas.**

- *Cobertura de ramas (`cargo-llvm-cov`).* Mide lo que **los tests** ejercitan, no lo que **la
  interfaz** alcanza — que es la pregunta. Un test puede llamar a la rama delegada y dejarla
  igual de inalcanzable para el usuario. Mediría bien la pregunta equivocada.
- *Construir la puerta y acabar con el problema.* Es la solución de verdad, y está fuera de alcance:
  #56 puede cambiar por dónde se entra a conectar, así que construirla ahora es apostar a que
  sobreviva.

---

## D3 · Un desafío, no un mapa

**Decisión.** El estado pasa de `Mutex<HashMap<String, DesafioPendiente>>` a
`Mutex<Option<DesafioEnCurso>>`.

**Por qué.** FR-009 pide que una cuenta no tenga dos desafíos compitiendo, y preguntarse «qué
identifica a una cuenta que todavía no existe» es una pregunta sin buena respuesta: antes de
conectarse no hay identidad, solo un formulario. Pero es que **la aplicación tiene un solo
formulario**: no se puede estar conectando dos cosas a la vez.

Con un `Option`, «dos desafíos en curso» deja de ser un estado que hay que impedir y pasa a ser uno
que **no se puede escribir**. Es la jugada que este proyecto prefiere: la garantía en el tipo, no en
la vigilancia.

**Alternativas descartadas.**

- *Mantener el mapa y limpiar por tiempo.* Deja al usuario esperando a que expire algo que ya
  abandonó, y conserva la posibilidad de dos.
- *Clave por `endpoint` + etiqueta.* Resuelve «dos por cuenta» y no «dos a la vez», que es el caso
  que ocurre: pulsar «Conectar» dos veces seguidas.

---

## D4 · Un desafío rechazado no se pierde — **defecto encontrado**

FR-006 pide que una credencial rechazada no consuma el desafío. **Hoy lo consume**, y por
construcción:

```rust
.and_then(|mut m| m.remove(&challenge_id))   // sale del mapa ANTES de intentar nada
```

`complete_connection` saca el desafío del mapa, y **después** llama al conector. Si el proveedor
rechaza la credencial, el desafío ya no está: el usuario tiene que rehacer el formulario entero por
haberse equivocado de tecla.

**Decisión.** El desafío se consume **solo al terminar bien**. Un fallo lo deja en pie, para
corregir sin volver a empezar.

**Cómo se prueba.** Un test del camino: desafío abierto → credencial rechazada → **sigue en curso**
→ credencial buena → conectado. Hoy ese test no puede pasar, y ninguno lo intentaba.

---

## D5 · Que el envío diga que está en curso — y el segundo no se cuele

**Decisión.** Mientras la credencial está en vuelo, el control de envío lo declara y **no admite un
segundo envío**. Es SC-005, y de paso cierra una carrera: dos envíos del mismo desafío, con el
segundo llegando a un desafío ya consumido.

**Por qué importa aquí más que en otras pantallas.** El defecto que originó este spec se vivió como
«pulsé y no pasó nada». Un envío sin señal es indistinguible de eso, así que la señal no es cortesía
— es lo que separa «está pasando» de «esto está roto otra vez».

**Relación con #51.** Misma familia —no saber si algo sigue vivo— pero otra superficie. Aquí la
espera es un viaje de red; allí son hasta quince minutos de sesión. No se resuelve el otro de paso.

---

## D6 · Dónde vive todo esto en la pantalla

**Decisión.** Dentro del plegable que `004` dejó, en secuencia: los campos de la cuenta → al pedir
conectar, aparecen las instrucciones y **dónde escribir la credencial**, con el control de cancelar
al lado → al terminar, la cuenta en la lista y el plegable vuelve a su estado limpio.

**Por qué no un paso aparte ni un diálogo.** `002` evita los diálogos por principio (SC-004), y
`004` decidió que esta superficie vive en la ventana principal. Sacar la credencial a otro sitio
contradiría las dos cosas por comodidad de implementación.

**Lo que se conserva.** Los 24 caracteres visibles por campo, el contenedor que envuelve, y el
techo que impide que un valor largo empuje a los demás (`004`-SC-005). Una clave de proveedor pasa
de cien caracteres con facilidad, así que ese techo deja de ser teórico.
