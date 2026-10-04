//! Programas de modelos instalados, encontrados recorriendo el `PATH`.
//!
//! **No se ejecuta ninguno.** Se reporta presencia y nada más: lo único que hace falta saber es que
//! está instalado, y confirmarlo invocándolo compra un nombre más preciso a cambio de ejecutar
//! código arbitrario. Está razonado en `research.md` D3.

use crate::application::deteccion::{
    Distancia, DondeInfiere, FuenteDeDeteccion, ModeloDisponible, Procedencia,
};
use crate::domain::error::Result;
use std::path::PathBuf;

/// Lo que se sabe buscar, y **dónde infiere cada cosa**.
///
/// Que `claude` esté en esta lista es correcto: el producto quiere los dos caminos. Lo que sería un
/// defecto es presentarlo como local — corre en la máquina y **consulta un servicio remoto**.
///
/// La lista es **declarada**: encontrar algo nuevo exige enseñárselo. Eso se dice en vez de
/// aparentar que se descubre solo.
const CONOCIDOS: &[(&str, DondeInfiere)] = &[
    ("ollama", DondeInfiere::Dentro),
    ("llama-server", DondeInfiere::Dentro),
    ("llama-cli", DondeInfiere::Dentro),
    ("lms", DondeInfiere::Dentro),
    ("claude", DondeInfiere::Fuera),
    ("codex", DondeInfiere::Fuera),
];

pub struct ProgramasEnElPath {
    directorios: Vec<PathBuf>,
}

impl Default for ProgramasEnElPath {
    fn default() -> Self {
        Self::del_entorno()
    }
}

impl ProgramasEnElPath {
    /// El `PATH` **entero**, no un conjunto de rutas habituales: compilar desde fuente es normal en
    /// este terreno y deja los binarios fuera de todo sitio estándar.
    pub fn del_entorno() -> Self {
        let directorios = std::env::var_os("PATH")
            .map(|p| std::env::split_paths(&p).collect())
            .unwrap_or_default();
        Self { directorios }
    }

    pub fn con_path(directorios: Vec<PathBuf>) -> Self {
        Self { directorios }
    }

    /// Qué directorios se van a mirar. Existe para poder comprobar que son **el `PATH` entero** y
    /// no una lista escrita a mano: sin esto, un test que construya la fuente con rutas explícitas
    /// pasa igual aunque `del_entorno` mire solo dos sitios conocidos. Pasó.
    pub fn directorios(&self) -> &[PathBuf] {
        &self.directorios
    }
}

#[async_trait::async_trait]
impl FuenteDeDeteccion for ProgramasEnElPath {
    fn nombre(&self) -> &'static str {
        "programas"
    }

    async fn buscar(&self) -> Result<Vec<ModeloDisponible>> {
        let mut hallados = Vec::new();
        for dir in &self.directorios {
            for (nombre, donde) in CONOCIDOS {
                if hallados
                    .iter()
                    .any(|m: &ModeloDisponible| m.nombre() == *nombre)
                {
                    continue;
                }
                if dir.join(nombre).is_file() {
                    // **Declarado**, nunca sostenido: un binario del `PATH` no puede probar dónde
                    // infiere. Por eso no se cablea como local aunque diga inferir dentro.
                    hallados.push(ModeloDisponible::declarado(
                        *nombre,
                        Procedencia::Programa,
                        Distancia::Instalado,
                        *donde,
                    ));
                }
            }
        }
        Ok(hallados)
    }
}
