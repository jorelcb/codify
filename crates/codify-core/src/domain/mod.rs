//! Capa de Dominio. Pura: sin I/O, sin red, sin proveedores, sin framework.
//! Define QUÉ necesita el authoring, nunca CÓMO se realiza.

pub mod audit;
pub mod change;
pub mod context;
pub mod error;
pub mod ports;
pub mod reference;
pub mod session;
pub mod write;
