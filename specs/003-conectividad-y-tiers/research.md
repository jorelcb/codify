# Research — Conectividad y reparto de modelos

Fase 0. Las decisiones de auth y de alcance del egress las cerró la sesión de clarificación del
2026-08-26 y viven en `spec.md`; aquí van las técnicas, con lo que se descartó y **la señal que
justificaría revisitarlo**.

## D1 — El modo, en el tipo del constructor

- **Decisión**: `CoreBuilder<M>` con `M ∈ {Local, Hybrid}`. El método que acepta un proveedor
  remoto existe **solo** en `CoreBuilder<Hybrid>`. Un grafo local no puede contener un adapter de
  red porque no hay forma de escribirlo.
- **Rationale**: FR-008 pide «estructuralmente imposible». Un rechazo en tiempo de ejecución dice
  «no lo hace»; un método que no existe dice «no puede». Es la diferencia entre una comprobación
  que hay que recordar mantener y una que el compilador mantiene sola.
- **Alternativas**:
  - **Solo comprobación en runtime** (lo de hoy) — correcta pero insuficiente una vez existe un
    adapter capaz de salir: la garantía pasaría a depender de que nadie construya el adapter por
    otra vía. **Se mantiene igualmente**, como defensa en profundidad.
  - **Fitness function que lea el código del composition root** — analizar texto para comprobar
    que el adapter remoto solo se instancia bajo el modo híbrido. Frágil: un renombrado la engaña,
    y no distingue una llamada muerta de una viva.
  - **Feature flag de compilación** (`--features remote`) — daría una garantía aún más fuerte
    —el código ni se compila—, pero obligaría a distribuir **dos binarios** y a que el usuario
    eligiera el modo al instalar, no al usar. Contradice FR-008a y SC-007.
- **Revisitar si**: aparece la necesidad de un tercer modo, o si el parámetro de tipo empieza a
  propagarse más allá del constructor. Si `AuthoringDeps` acabara genérico, el coste habría
  superado al beneficio.

## D2 — Un adapter remoto genérico, no uno por proveedor

- **Decisión**: un solo adapter contra API compatible con OpenAI, parametrizado por endpoint,
  credencial y tier declarado. Cubre la mayoría de proveedores y pasarelas sin código por cada uno.
- **Rationale**: `001`-FR-016 exige no acoplarse a un proveedor específico, y el adapter local ya
  demostró que el patrón funciona: `LocalOpenAiCompatProvider` sirve Ollama y `llama.cpp` sin
  saber cuál es cuál.
- **Alternativas**: un adapter por proveedor (más fiel a cada API, multiplica el mantenimiento y
  acopla el spec a nombres comerciales). **Revisitar si** un proveedor que el usuario quiera exige
  un contrato que el genérico no pueda expresar.

## D3 — El keyring del sistema, sin plan B

- **Decisión**: `keyring` contra el almacén del sistema operativo. Si no está disponible, se dice
  y se ofrece seguir en local (FR-004). **No hay respaldo en archivo.**
- **Rationale**: un respaldo en archivo cifrado solo mueve el problema —¿dónde vive la llave?— y
  daría una promesa de custodia que el producto no puede sostener. Es preferible no ofrecer la
  función que ofrecerla peor de lo que el usuario supondrá.
- **Alternativas**: archivo cifrado con clave derivada de una contraseña del usuario (añade una
  contraseña más que recordar, y un modo de fallo nuevo); guardar en claro (descartado sin
  discusión). **Revisitar si** aparece demanda real en entornos Linux sin *keyring* — sabiendo que
  la respuesta correcta puede seguir siendo «usa el modo local».

## D4 — Dos vías de conexión tras una sola frontera

- **Decisión**: un port `AccountConnector` con dos implementaciones —device-flow y credencial
  directa—, elegidas según lo que el proveedor ofrezca. El resto del sistema no sabe cuál se usó.
- **Rationale**: la diferencia entre las dos vías es de **cómo se obtiene** el secreto, no de qué
  se hace con él. Custodia, uso y revocación son idénticos, así que la frontera va donde termina
  esa diferencia.
- **Alternativas**: dos caminos separados de punta a punta (duplica la custodia y el ciclo de
  vida, que es justo lo que no cambia). **Revisitar si** un proveedor exige renovar el token de
  una forma que el otro camino no contemple.

## Unknowns técnicos resueltos

- **Device flow**: `oauth2` con `DeviceAuthorizationUrl`. La aplicación muestra código y URL; el
  sondeo tiene su propio timeout y **no** reutiliza el de generación de 900 s — esperar a un
  humano y esperar a un modelo son cosas distintas.
- **Tier de un proveedor remoto**: declarado al conectar, no adivinado. El sistema no puede saber
  si un endpoint sirve un modelo caro o barato, y fingir que sí produciría un reparto arbitrario.
- **Revocación**: desconectar borra del keyring y **rearma el grafo**, que es lo que hace que
  SC-006 sea inmediato sin reiniciar.
