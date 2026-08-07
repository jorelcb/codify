//! Motor de diffs y política de riesgo (US2 del spec 001).
//!
//! Ambos son **adapters**: implementan ports que el core nombra, y se cablean en el
//! composition root. El dominio no sabe que existe `similar`.

pub mod engine;
pub mod risk;
