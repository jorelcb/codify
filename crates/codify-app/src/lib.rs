//! Piel de escritorio de codify-NG (fase 1).
//!
//! Linkea `codify-core` **in-process**. Es un adaptador: invoca casos de uso, traduce sus
//! resultados a DTOs y reenvía el log de auditoría del núcleo a la ventana como eventos.
//! Ninguna regla de dominio vive aquí.

pub mod adapters;
pub mod commands;

pub fn run() {
    tauri::Builder::default()
        .manage(commands::AppState::default())
        .invoke_handler(tauri::generate_handler![
            commands::start_session,
            commands::session_state,
            commands::set_locale,
        ])
        .run(tauri::generate_context!())
        .expect("no se pudo arrancar la aplicación");
}
