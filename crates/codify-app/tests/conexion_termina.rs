//! `005` — que conectar una cuenta llegue a su fin.
//!
//! Estos tests recorren **el camino**, no sus piezas. La distinción importa aquí más que en otros
//! sitios: el defecto que originó este spec vivía entre las piezas —cada una probada, ninguna
//! conectada con la siguiente— y sobrevivió a dos ciclos con la suite en verde.
//!
//! Por eso se ejercita el ciclo de vida a través de `AppState`, que es donde el cableado ocurre de
//! verdad. Probar `conexion_desde` y `DirectCredential` por separado ya se hacía, y no bastó
//! (issue #48).

use codify_app::commands::{AppState, FalloDeConexion};
use codify_core::application::ports::{CredentialStore, ReferenciaDeCredencial, Secreto, Tier};
use codify_core::domain::error::Result;
use std::sync::{Arc, Mutex};

/// Almacén en memoria. **Ningún test de este archivo toca el llavero real**: comprobar que el
/// camino termina no requiere escribir secretos en la máquina de quien corre la suite, y un test
/// que lo hiciera dejaría rastro en el llavero de todo el que ejecute `cargo test`.
#[derive(Default)]
struct AlmacenDeMentira {
    guardado: Mutex<Vec<String>>,
    disponible: bool,
}

impl AlmacenDeMentira {
    fn disponible() -> Arc<Self> {
        Arc::new(Self {
            disponible: true,
            ..Default::default()
        })
    }
    fn ausente() -> Arc<Self> {
        Arc::new(Self::default())
    }
    fn cuantos(&self) -> usize {
        self.guardado.lock().map(|g| g.len()).unwrap_or(0)
    }
}

#[async_trait::async_trait]
impl CredentialStore for AlmacenDeMentira {
    fn disponible(&self) -> bool {
        self.disponible
    }
    /// El secreto **no se inspecciona**: no hay forma de sacarlo de `Secreto`, y está bien que no
    /// la haya. Al test le basta con que llegara uno.
    async fn guardar(&self, r: &ReferenciaDeCredencial, _s: Secreto) -> Result<()> {
        if let Ok(mut g) = self.guardado.lock() {
            g.push(format!("{r:?}"));
        }
        Ok(())
    }
    async fn obtener(&self, _r: &ReferenciaDeCredencial) -> Result<Option<Secreto>> {
        Ok(None)
    }
    async fn borrar(&self, _r: &ReferenciaDeCredencial) -> Result<()> {
        Ok(())
    }
}

fn estado() -> AppState {
    AppState::default().con_almacen(AlmacenDeMentira::disponible())
}

/// **FR-002, FR-003** — el camino feliz llega hasta el final.
///
/// Es el test que, de haber existido, habría hecho innecesario este spec: `complete_connection`
/// estaba entero desde `003` y nadie lo llamaba.
#[tokio::test]
async fn conectar_una_cuenta_termina_en_la_lista() {
    let estado = estado();
    let desafio = estado
        .abrir_desafio(
            "mi-proveedor",
            "https://api.ejemplo.com/v1",
            Tier::Cheap,
            false,
        )
        .await
        .expect("abrir");

    assert_eq!(desafio.kind, "credencial");
    assert!(
        desafio.instructions.is_some(),
        "el desafío debe decir qué se espera del usuario"
    );

    let conexion = estado
        .completar_desafio(&desafio.challenge_id, Some("sk-de-prueba".into()))
        .await
        .expect("el camino feliz debe terminar en una conexión");

    assert_eq!(conexion.label, "mi-proveedor");
    assert_eq!(
        estado.conexiones().len(),
        1,
        "la cuenta debe quedar en la lista, sin reiniciar"
    );
    assert!(
        !estado.hay_desafio(),
        "el desafío se consume al terminar bien"
    );
}

/// **FR-006** — una credencial rechazada no pierde el desafío.
///
/// Hoy lo pierde, y por construcción: `complete_connection` saca el desafío del estado **antes**
/// de intentar completarlo. Equivocarse de tecla obliga a rehacer el formulario entero.
#[tokio::test]
async fn una_credencial_rechazada_no_pierde_el_desafio() {
    let estado = estado();
    let desafio = estado
        .abrir_desafio(
            "mi-proveedor",
            "https://api.ejemplo.com/v1",
            Tier::Cheap,
            false,
        )
        .await
        .expect("abrir");

    let fallo = estado
        .completar_desafio(&desafio.challenge_id, None)
        .await
        .expect_err("sin credencial no puede haber conexión");
    assert_eq!(fallo, FalloDeConexion::CredencialVacia);

    assert!(
        estado.hay_desafio(),
        "el desafío sigue en pie tras un rechazo: se corrige sin volver a empezar"
    );

    estado
        .completar_desafio(&desafio.challenge_id, Some("sk-buena".into()))
        .await
        .expect("el mismo desafío debe poder completarse al segundo intento");
    assert_eq!(estado.conexiones().len(), 1);
}

/// **FR-009, SC-006** — no hay dos desafíos a la vez.
///
/// Pulsar «Conectar» dos veces dejaba dos dentro del estado, y solo salía el que se completara.
#[tokio::test]
async fn pedir_conectar_dos_veces_deja_uno_solo() {
    let estado = estado();
    let primero = estado
        .abrir_desafio("uno", "https://api.uno.com/v1", Tier::Cheap, false)
        .await
        .expect("abrir");
    let segundo = estado
        .abrir_desafio("dos", "https://api.dos.com/v1", Tier::Heavy, false)
        .await
        .expect("abrir");

    assert_ne!(primero.challenge_id, segundo.challenge_id);
    assert!(estado.hay_desafio());

    let fallo = estado
        .completar_desafio(&primero.challenge_id, Some("sk".into()))
        .await
        .expect_err("el primero fue reemplazado y ya no está en curso");
    assert_eq!(fallo, FalloDeConexion::DesafioNoEnCurso);

    estado.abandonar_desafio();
    assert!(
        !estado.hay_desafio(),
        "tras abandonar no queda ningún desafío en curso"
    );
}

/// **FR-005, `003`-FR-004** — sin almacén se dice, y no se recurre a otro sitio.
///
/// Cubre además el caso límite de que el almacén **desaparezca a mitad**: había cuando se abrió el
/// desafío y no cuando se completa. Es un fallo distinto de no haberlo tenido nunca, y el usuario
/// merece saber cuál de los dos le tocó.
#[tokio::test]
async fn sin_almacen_se_dice_y_no_se_guarda_en_otro_sitio() {
    let almacen = AlmacenDeMentira::ausente();
    let estado = AppState::default().con_almacen(Arc::clone(&almacen) as Arc<dyn CredentialStore>);
    let desafio = estado
        .abrir_desafio(
            "mi-proveedor",
            "https://api.ejemplo.com/v1",
            Tier::Cheap,
            false,
        )
        .await
        .expect("abrir");

    let fallo = estado
        .completar_desafio(&desafio.challenge_id, Some("sk-de-prueba".into()))
        .await
        .expect_err("sin almacén no puede completarse");

    assert_eq!(fallo, FalloDeConexion::SinAlmacen);
    assert_eq!(
        almacen.cuantos(),
        0,
        "no se guardó nada: FR-004 dice avisar, **no** buscar otro sitio donde dejar el secreto"
    );
    assert_eq!(estado.conexiones().len(), 0);
}

/// **FR-011** — el secreto no se guarda por duplicado en el estado de la aplicación.
///
/// Lo único que persiste es la **referencia** con la que pedírselo al almacén. Un `Secreto` que
/// viviera también en `AppState` sería una copia que nadie limpia.
#[tokio::test]
async fn la_credencial_no_queda_en_el_estado() {
    let estado = estado();
    let desafio = estado
        .abrir_desafio(
            "mi-proveedor",
            "https://api.ejemplo.com/v1",
            Tier::Cheap,
            false,
        )
        .await
        .expect("abrir");
    estado
        .completar_desafio(&desafio.challenge_id, Some("sk-secretisima".into()))
        .await
        .expect("conectar");

    let volcado = format!("{:?}", estado.conexiones());
    assert!(
        !volcado.contains("sk-secretisima"),
        "la credencial aparece en el estado de conexiones: {volcado}"
    );
}

/// Un conector que rechaza siempre. Es la única forma de provocar el caso **literal** de FR-006:
/// que el proveedor diga que no. El de la credencial vacía sale antes de hablar con nadie, así que
/// probaba otra cosa — lo descubrió una inyección que no hizo caer ningún test.
struct ConectorQueRechaza;

#[async_trait::async_trait]
impl codify_core::application::ports::AccountConnector for ConectorQueRechaza {
    async fn iniciar(&self) -> Result<codify_core::application::ports::Desafio> {
        Ok(codify_core::application::ports::Desafio::PideCredencial {
            instruccion: codify_core::application::ports::InstruccionDeCredencial::PegarClave,
        })
    }
    async fn completar(
        &self,
        _d: &codify_core::application::ports::Desafio,
        _r: Option<Secreto>,
    ) -> Result<Secreto> {
        Err(codify_core::domain::error::CoreError::Unauthorized(
            "el proveedor no acepta esa clave".into(),
        ))
    }
}

/// **FR-006, el caso literal** — el proveedor rechaza y el desafío **sigue en pie**.
#[tokio::test]
async fn un_rechazo_del_proveedor_no_pierde_el_desafio() {
    let estado = AppState::default()
        .con_almacen(AlmacenDeMentira::disponible())
        .con_conector(Arc::new(ConectorQueRechaza));
    let desafio = estado
        .abrir_desafio(
            "mi-proveedor",
            "https://api.ejemplo.com/v1",
            Tier::Cheap,
            false,
        )
        .await
        .expect("abrir");

    let fallo = estado
        .completar_desafio(&desafio.challenge_id, Some("sk-mala".into()))
        .await
        .expect_err("el proveedor rechaza");
    assert_eq!(fallo, FalloDeConexion::CredencialRechazada);
    assert!(fallo.conserva_el_desafio());
    assert!(
        estado.hay_desafio(),
        "un rechazo del proveedor no puede obligar a rehacer el formulario"
    );
    assert_eq!(estado.conexiones().len(), 0);
}
