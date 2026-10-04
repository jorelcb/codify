//! Servidores de modelos que ya están escuchando en este equipo.
//!
//! Es la **única** fuente que sostiene la clasificación por construcción: `LocalProviderProbe`
//! rechaza cualquier dirección que no sea del propio equipo, así que lo que llega por aquí no pudo
//! ser remoto. Donde se puede tener una garantía sostenida no se acepta una declarada.

use crate::application::deteccion::{Distancia, FuenteDeDeteccion, ModeloDisponible, Procedencia};
use crate::application::ports::ProviderDiscovery;
use crate::domain::error::Result;
use crate::infrastructure::providers::probe::LocalProviderProbe;

/// Puertos de loopback donde miran las herramientas del ramo. **Declarados**: añadir uno es
/// añadirlo aquí.
pub const PUERTOS: &[(&str, u16)] = &[("ollama", 11434), ("llama.cpp", 8080), ("lm-studio", 1234)];

#[derive(Default)]
pub struct ServidoresEnEsteEquipo;

#[async_trait::async_trait]
impl FuenteDeDeteccion for ServidoresEnEsteEquipo {
    fn nombre(&self) -> &'static str {
        "servidores"
    }

    async fn buscar(&self) -> Result<Vec<ModeloDisponible>> {
        let mut hallados = Vec::new();
        for (nombre, puerto) in PUERTOS {
            let url = format!("http://127.0.0.1:{puerto}");
            // Si esto no se puede construir es porque la dirección no es del propio equipo, y
            // entonces no tenemos nada que ofrecer: la garantía va antes que el hallazgo.
            let Ok(sonda) = LocalProviderProbe::new(&url) else {
                continue;
            };
            // Un servidor que responde sin modelos lo dice `ProviderIssue::NoModels`, así que ese
            // caso límite queda cubierto por lo que ya hay.
            let estado = sonda.probe().await;
            {
                if estado.issue.is_none() && !estado.models.is_empty() {
                    hallados.push(ModeloDisponible::sostenido_local(
                        *nombre,
                        Procedencia::Servidor,
                        Distancia::Usable,
                    ));
                }
            }
        }
        Ok(hallados)
    }
}
