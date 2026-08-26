//! Suite de contrato de `CredentialStore` (`003`-T007/T008).
//!
//! Una sola suite contra **dos** adapters: el keyring del sistema y un doble en memoria. Es el
//! patrón que ya usan los demás ports de este repositorio, y lo que hace que sustituir el
//! adapter sea una decisión y no una apuesta.
//!
//! El adapter real va tras `#[ignore]`: escribe en el llavero del usuario. Correrlo es una
//! decisión de quien lo ejecuta, y CI no la toma por él.

use codify_core::application::ports::{CredentialStore, ReferenciaDeCredencial, Secreto};
use codify_core::domain::audit::{AuditEvent, AuditKind};
use codify_core::infrastructure::secrets::keyring::SystemKeyring;
use std::collections::HashMap;
use std::sync::Mutex;

/// Doble en memoria. Existe para que la suite corra en CI sin tocar el llavero real.
#[derive(Default)]
struct AlmacenEnMemoria {
    datos: Mutex<HashMap<String, String>>,
    caido: bool,
}

impl AlmacenEnMemoria {
    fn caido() -> Self {
        Self {
            caido: true,
            ..Default::default()
        }
    }
}

#[async_trait::async_trait]
impl CredentialStore for AlmacenEnMemoria {
    fn disponible(&self) -> bool {
        !self.caido
    }
    async fn guardar(
        &self,
        r: &ReferenciaDeCredencial,
        s: Secreto,
    ) -> codify_core::domain::error::Result<()> {
        self.datos.lock().unwrap().insert(
            r.as_str().to_string(),
            s.exponer_para_la_peticion().to_string(),
        );
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
// El contrato — una sola definición, dos adapters
// ---------------------------------------------------------------------------

async fn el_contrato(store: &dyn CredentialStore, sufijo: &str) {
    let r = ReferenciaDeCredencial::new(format!("contrato-{sufijo}"));
    let _ = store.borrar(&r).await;

    assert!(
        store.obtener(&r).await.unwrap().is_none(),
        "una referencia sin guardar no devuelve nada"
    );

    store
        .guardar(&r, Secreto::new("valor-secreto-123"))
        .await
        .expect("guardar");
    assert_eq!(
        store
            .obtener(&r)
            .await
            .unwrap()
            .map(|s| s.exponer_para_la_peticion().to_string()),
        Some("valor-secreto-123".into())
    );

    store.borrar(&r).await.expect("borrar");
    assert!(store.obtener(&r).await.unwrap().is_none());

    store
        .borrar(&r)
        .await
        .expect("borrar dos veces no es un error: desconectar es idempotente");
}

#[tokio::test]
async fn el_doble_en_memoria_cumple_el_contrato() {
    el_contrato(&AlmacenEnMemoria::default(), "memoria").await;
}

#[tokio::test]
#[ignore = "escribe en el llavero del usuario; se corre a mano"]
async fn el_keyring_del_sistema_cumple_el_contrato() {
    el_contrato(&SystemKeyring::new(), "sistema").await;
}

// ---------------------------------------------------------------------------
// T008 — el secreto no aparece donde podría filtrarse
// ---------------------------------------------------------------------------

/// `003`-FR-002 prohíbe que la credencial llegue a un registro. La forma fiable de cumplirlo no
/// es acordarse de no imprimirla, sino que **imprimirla no sirva de nada**.
#[test]
fn el_secreto_no_aparece_al_formatearlo() {
    let s = Secreto::new("sk-esto-no-debe-verse");
    let formateado = format!("{s:?}");

    assert!(
        !formateado.contains("sk-esto-no-debe-verse"),
        "el secreto se filtró al formatearlo: {formateado}"
    );
    assert_eq!(formateado, "Secreto(<redactado>)");
}

/// Y dentro de una estructura anidada, que es como se filtran de verdad: nadie imprime el
/// secreto a propósito — imprime lo que lo contiene.
#[test]
fn el_secreto_tampoco_se_filtra_anidado() {
    #[derive(Debug)]
    #[allow(dead_code)]
    struct Peticion {
        endpoint: String,
        credencial: Secreto,
    }

    let p = Peticion {
        endpoint: "https://api.example.com".into(),
        credencial: Secreto::new("sk-esto-no-debe-verse"),
    };
    assert!(!format!("{p:?}").contains("sk-esto-no-debe-verse"));
}

/// Ni en un evento de auditoría, que es el sitio por el que un secreto acabaría en disco.
#[test]
fn el_secreto_no_llega_al_registro_de_auditoria() {
    let s = Secreto::new("sk-esto-no-debe-verse");
    let evento = AuditEvent::new(
        "2026-08-26T00:00:00Z",
        AuditKind::SessionFailed,
        format!("{s:?}"),
    );
    assert!(!evento.payload.contains("sk-esto-no-debe-verse"));
}

/// Un almacén caído se declara **sin escribir nada** (FR-004).
#[tokio::test]
async fn un_almacen_caido_se_declara_antes_de_intentar_guardar() {
    let store = AlmacenEnMemoria::caido();
    assert!(
        !store.disponible(),
        "FR-004 exige poder avisar ANTES de que el usuario intente conectar"
    );
}
