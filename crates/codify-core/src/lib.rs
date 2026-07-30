//! # codify-core
//!
//! Núcleo hexagonal de codify-NG. Contiene el dominio del *authoring* de contexto y la
//! orquestación del loop agéntico. **No conoce UI ni protocolo**: las pieles (Tauri hoy;
//! MCP/CLI mañana) linkean este crate y lo consumen a través de sus ports.
//!
//! Regla de Dependencia (constitución, Principio I): el código apunta **solo hacia adentro**
//! `infrastructure → application → domain`. Verificado por la fitness function
//! `tests/arch_deps.rs`.

pub mod application;
pub mod domain;
pub mod infrastructure;
