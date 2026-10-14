//! `006` — qué modelos hay disponibles en esta máquina.
//!
//! Detectar y proveer son cosas distintas: una encuentra candidatos, la otra habla con ellos.
//! Están separadas a propósito, porque mezclarlas haría que «proveedor» signifique dos cosas — y
//! este proyecto ya pagó ese precio con esa palabra exacta.

use crate::domain::error::Result;
use std::time::Duration;

/// Dónde ocurre la inferencia. **No es lo mismo que dónde corre el programa.**
///
/// Un servidor local corre en la máquina e infiere en la máquina. Un agente de línea de órdenes de
/// un servicio remoto corre en la máquina e infiere **fuera**. Los dos caminos existen a propósito:
/// el producto quiere integrarse con unos y con otros. El defecto no es tener el segundo — es
/// presentarlo como el primero.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum DondeInfiere {
    Dentro,
    Fuera,
}

impl DondeInfiere {
    pub fn code(&self) -> &'static str {
        match self {
            DondeInfiere::Dentro => "inside",
            DondeInfiere::Fuera => "outside",
        }
    }
    pub fn all() -> [DondeInfiere; 2] {
        [DondeInfiere::Dentro, DondeInfiere::Fuera]
    }
}

/// Cuánto falta para poder usar un modelo. **Ordenada**: lo más cerca se ofrece primero.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum Distancia {
    /// Responde ahora mismo.
    Usable,
    /// Está en la máquina y no en marcha.
    Instalado,
    /// No está. Existe para que quien abre esto por primera vez **tenga algo que leer**.
    Ausente,
}

impl Distancia {
    pub fn code(&self) -> &'static str {
        match self {
            Distancia::Usable => "usable",
            Distancia::Instalado => "installed",
            Distancia::Ausente => "absent",
        }
    }
    pub fn all() -> [Distancia; 3] {
        [Distancia::Usable, Distancia::Instalado, Distancia::Ausente]
    }
}

/// De dónde salió un modelo. **Obligatoria**: el principio rector del proyecto es que lo que no se
/// puede verificar contra una fuente no se afirma, y una lista que aparece sola sin decir de dónde
/// salió cada fila es exactamente esa clase de afirmación.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Procedencia {
    /// Un servidor que responde en un puerto del propio equipo.
    Servidor,
    /// Un programa encontrado en el `PATH`.
    Programa,
    /// Lo puso el usuario.
    Usuario,
}

impl Procedencia {
    pub fn code(&self) -> &'static str {
        match self {
            Procedencia::Servidor => "server",
            Procedencia::Programa => "program",
            Procedencia::Usuario => "user",
        }
    }
    pub fn all() -> [Procedencia; 3] {
        [
            Procedencia::Servidor,
            Procedencia::Programa,
            Procedencia::Usuario,
        ]
    }
}

/// Algo con lo que se podría trabajar.
///
/// **No se puede construir sin procedencia**: no hay constructor que la omita. Eso lo sostiene el
/// tipo, no un test.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ModeloDisponible {
    nombre: String,
    procedencia: Procedencia,
    distancia: Distancia,
    donde_infiere: DondeInfiere,
    /// `true` solo cuando la clasificación viene de una fuente que **no pudo devolver otra cosa**.
    sostenido: bool,
}

impl ModeloDisponible {
    /// Un modelo cuya localidad está **sostenida por construcción**: llegó por una vía que rechaza
    /// cualquier dirección que no sea del propio equipo, así que no pudo ser remoto.
    pub fn sostenido_local(
        nombre: impl Into<String>,
        procedencia: Procedencia,
        distancia: Distancia,
    ) -> Self {
        Self {
            nombre: nombre.into(),
            procedencia,
            distancia,
            donde_infiere: DondeInfiere::Dentro,
            sostenido: true,
        }
    }

    /// Un modelo cuya clasificación es una **declaración**: se enseña, y no se cablea como local
    /// aunque diga inferir dentro. Donde se puede tener una garantía sostenida no se acepta una
    /// declarada.
    pub fn declarado(
        nombre: impl Into<String>,
        procedencia: Procedencia,
        distancia: Distancia,
        donde_infiere: DondeInfiere,
    ) -> Self {
        Self {
            nombre: nombre.into(),
            procedencia,
            distancia,
            donde_infiere,
            sostenido: false,
        }
    }

    pub fn nombre(&self) -> &str {
        &self.nombre
    }
    pub fn procedencia(&self) -> Procedencia {
        self.procedencia
    }
    pub fn distancia(&self) -> Distancia {
        self.distancia
    }
    pub fn donde_infiere(&self) -> DondeInfiere {
        self.donde_infiere
    }

    /// Solo lo sostenido **y** dentro. Las dos condiciones, y la primera es la que importa.
    pub fn se_puede_cablear_como_local(&self) -> bool {
        self.sostenido && self.donde_infiere == DondeInfiere::Dentro
    }
}

/// Un sitio donde mirar. Cada fuente sabe buscar una clase de cosa y **declara lo que encuentra**.
///
/// **Ninguna ejecuta lo que encuentra.** Mirar el `PATH` dice que algo está; invocarlo para
/// confirmarlo sería lanzar un binario arbitrario elegido por su nombre, para decidir si enseñar
/// una fila en una lista.
#[async_trait::async_trait]
pub trait FuenteDeDeteccion: Send + Sync {
    fn nombre(&self) -> &'static str;
    async fn buscar(&self) -> Result<Vec<ModeloDisponible>>;
}

/// Lo que la detección devolvió, y si le dio tiempo.
#[derive(Debug, Clone)]
pub struct Deteccion {
    pub modelos: Vec<ModeloDisponible>,
    /// `true` si alguna fuente no llegó a tiempo. Se **dice**, no se oculta.
    pub incompleta: bool,
}

/// El presupuesto. Número **elegido, no medido**: por encima se nota al abrir, y por debajo se
/// corre el riesgo de descartar un servidor que tarda en contestar.
pub const PRESUPUESTO: Duration = Duration::from_secs(3);

/// Corre todas las fuentes **a la vez** con un presupuesto común, y declara lo que no llegó.
pub async fn detectar(fuentes: &[std::sync::Arc<dyn FuenteDeDeteccion>]) -> Deteccion {
    detectar_con(fuentes, PRESUPUESTO).await
}

pub async fn detectar_con(
    fuentes: &[std::sync::Arc<dyn FuenteDeDeteccion>],
    presupuesto: Duration,
) -> Deteccion {
    // A la vez, no en fila: el presupuesto es común, así que una fuente lenta no consume el de
    // las demás. Con `JoinSet` y sin dependencias nuevas.
    let mut tareas = tokio::task::JoinSet::new();
    for f in fuentes {
        let f = std::sync::Arc::clone(f);
        tareas.spawn(async move { f.buscar().await.unwrap_or_default() });
    }
    let todas = async move {
        let mut salida = Vec::new();
        while let Some(r) = tareas.join_next().await {
            salida.extend(r.unwrap_or_default());
        }
        salida
    };

    match tokio::time::timeout(presupuesto, todas).await {
        Ok(resultados) => {
            let mut modelos: Vec<ModeloDisponible> = resultados;
            // Lo más cerca de poder usarse, primero.
            modelos.sort_by_key(|m| m.distancia());
            Deteccion {
                modelos,
                incompleta: false,
            }
        }
        // Lo que no llegó se abandona **diciéndolo**: una espera muda es indistinguible de que no
        // pase nada, y eso ya costó un ciclo en esta misma aplicación.
        Err(_) => Deteccion {
            modelos: Vec::new(),
            incompleta: true,
        },
    }
}
