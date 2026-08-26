//! Capa de Infraestructura: adapters concretos y composition root.
//! **Único lugar** con I/O, red y dependencias de terceros.

pub mod cancel;
pub mod composition;
pub mod diff;
pub mod providers;
pub mod repo;
pub mod secrets;
