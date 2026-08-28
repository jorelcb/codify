//! Escenario S2/S4/S6 de `003` — comprobar **fuera de la aplicación** que el secreto no está.
//!
//! `#[ignore]`: escribe en el llavero real. Lo corre quien valida, no CI.

use codify_core::application::ports::{CredentialStore, ReferenciaDeCredencial, Secreto};
use codify_core::infrastructure::secrets::keyring::SystemKeyring;

#[tokio::test]
#[ignore = "escribe en el llavero del usuario"]
async fn guarda_una_marca_rastreable() {
    let store = SystemKeyring::new();
    let r = ReferenciaDeCredencial::new("prueba-fuga");
    store.guardar(&r, Secreto::new("sk-marca-de-fuga-1787854978")).await.expect("guardar");
    println!("GUARDADA");
}

#[tokio::test]
#[ignore = "limpia lo que dejó el anterior"]
async fn borra_la_marca() {
    let store = SystemKeyring::new();
    store.borrar(&ReferenciaDeCredencial::new("prueba-fuga")).await.expect("borrar");
    println!("BORRADA");
}
