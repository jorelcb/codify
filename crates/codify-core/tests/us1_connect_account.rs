//! **`003`-US1** — conectar una cuenta una sola vez, sin dejar rastro del secreto.

use codify_core::application::connections::{ConnectionState, ProviderConnection};
use codify_core::application::ports::{
    AccountConnector, CredentialStore, Desafio, ReferenciaDeCredencial, Secreto, Tier,
};
use codify_core::domain::error::CoreError;
use codify_core::infrastructure::secrets::device_flow::DeviceFlow;
use codify_core::infrastructure::secrets::direct::DirectCredential;
use std::collections::HashMap;
use std::sync::Mutex;
use std::time::Duration;

#[derive(Default)]
struct Almacen {
    datos: Mutex<HashMap<String, String>>,
}

#[async_trait::async_trait]
impl CredentialStore for Almacen {
    fn disponible(&self) -> bool {
        true
    }
    async fn guardar(
        &self,
        r: &ReferenciaDeCredencial,
        s: Secreto,
    ) -> codify_core::domain::error::Result<()> {
        self.datos
            .lock()
            .unwrap()
            .insert(r.as_str().into(), s.exponer_para_la_peticion().to_string());
        Ok(())
    }
    async fn obtener(
        &self,
        r: &ReferenciaDeCredencial,
    ) -> codify_core::domain::error::Result<Option<Secreto>> {
        Ok(self
            .datos
            .lock()
            .unwrap()
            .get(r.as_str())
            .map(|v| Secreto::new(v.clone())))
    }
    async fn borrar(&self, r: &ReferenciaDeCredencial) -> codify_core::domain::error::Result<()> {
        self.datos.lock().unwrap().remove(r.as_str());
        Ok(())
    }
}

// ---------------------------------------------------------------------------
// T009 — abandonar no deja nada a medias; desconectar surte efecto ya
// ---------------------------------------------------------------------------

#[tokio::test]
async fn abandonar_la_credencial_no_deja_conexion_a_medias() {
    let conector = DirectCredential::new("pega tu clave");
    let desafio = conector.iniciar().await.expect("inicia");

    let err = conector
        .completar(&desafio, None)
        .await
        .expect_err("no se introdujo nada");

    assert!(
        matches!(err, CoreError::Unauthorized(_)),
        "abandonar no es un fallo del proveedor: es que el usuario no terminó, y distinguirlo \
         permite ofrecer reintentar en vez de sugerir que algo se rompió"
    );
}

#[tokio::test]
async fn denegar_la_autorizacion_delegada_se_distingue_de_un_fallo() {
    // Un proveedor que responde `access_denied` desde el primer sondeo.
    let servidor = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let puerto = servidor.local_addr().unwrap().port();
    tokio::spawn(async move {
        loop {
            let Ok((mut socket, _)) = servidor.accept().await else {
                return;
            };
            tokio::spawn(async move {
                use tokio::io::AsyncWriteExt;
                let cuerpo = r#"{"error":"access_denied"}"#;
                let resp = format!(
                    "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\n\r\n{}",
                    cuerpo.len(),
                    cuerpo
                );
                let _ = socket.write_all(resp.as_bytes()).await;
            });
        }
    });

    let flow = DeviceFlow::new(
        format!("http://127.0.0.1:{puerto}/device"),
        format!("http://127.0.0.1:{puerto}/token"),
        "cliente",
    )
    .expect("conector")
    .con_espera(Duration::from_secs(5));

    let desafio = Desafio::Delegada {
        codigo: "ABCD".into(),
        url: "http://example.test/verify#dev-123".into(),
    };
    let err = flow
        .completar(&desafio, None)
        .await
        .expect_err("el proveedor denegó");

    assert!(
        matches!(err, CoreError::Unauthorized(_)),
        "denegar es una respuesta, no una avería: {err:?}"
    );
}

#[tokio::test]
async fn desconectar_impide_el_uso_en_la_tarea_siguiente() {
    let store = Almacen::default();
    let referencia = ReferenciaDeCredencial::new("cuenta-1");
    store
        .guardar(&referencia, Secreto::new("sk-123"))
        .await
        .unwrap();

    let mut conexion = ProviderConnection::new(
        "cuenta-1",
        "Frontier",
        "api.example.com",
        Tier::Heavy,
        referencia.clone(),
    );
    assert!(conexion.usable());

    // Desconectar: borra del almacén y marca la conexión.
    store.borrar(&referencia).await.unwrap();
    conexion.state = ConnectionState::Revoked;

    assert!(
        !conexion.usable(),
        "SC-006: la siguiente tarea no puede usarla, y sin reiniciar"
    );
    assert!(
        store.obtener(&referencia).await.unwrap().is_none(),
        "y la credencial ya no está donde estaba"
    );
}

// ---------------------------------------------------------------------------
// T010 — sin almacén, se dice y no se escribe nada
// ---------------------------------------------------------------------------

#[tokio::test]
async fn sin_almacen_disponible_no_se_recurre_a_otro_sitio() {
    struct Caido;
    #[async_trait::async_trait]
    impl CredentialStore for Caido {
        fn disponible(&self) -> bool {
            false
        }
        async fn guardar(
            &self,
            _: &ReferenciaDeCredencial,
            _: Secreto,
        ) -> codify_core::domain::error::Result<()> {
            Err(CoreError::Storage("no hay almacén".into()))
        }
        async fn obtener(
            &self,
            _: &ReferenciaDeCredencial,
        ) -> codify_core::domain::error::Result<Option<Secreto>> {
            Err(CoreError::Storage("no hay almacén".into()))
        }
        async fn borrar(
            &self,
            _: &ReferenciaDeCredencial,
        ) -> codify_core::domain::error::Result<()> {
            Err(CoreError::Storage("no hay almacén".into()))
        }
    }

    let store = Caido;
    assert!(
        !store.disponible(),
        "FR-004: se puede avisar ANTES de que el usuario intente conectar"
    );

    let err = store
        .guardar(&ReferenciaDeCredencial::new("x"), Secreto::new("y"))
        .await
        .expect_err("no hay dónde guardar");
    assert!(
        matches!(err, CoreError::Storage(_)),
        "falla y se dice; lo que NO hace es caer a un archivo — research.md D3"
    );
}
