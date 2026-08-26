//! Capa de Aplicación. Orquesta el loop de authoring dependiendo **solo de ports**
//! (constitución, Principio I + decisión raíz D5). No conoce adapters concretos.

pub mod authoring_loop;
pub mod connections;
pub mod deps;
pub mod ingest;
pub mod ports;
pub mod refine;
pub mod service;
