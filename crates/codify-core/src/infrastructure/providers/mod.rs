//! Adapters de `ModelProvider` (decisión raíz D3).
//!
//! v1 cubre los backends **locales** con API OpenAI-compatible (Ollama, `llama.cpp` server).
//! Los remotos con OAuth llegan en la fase de Polish (T045).

pub mod local;
pub mod probe;
