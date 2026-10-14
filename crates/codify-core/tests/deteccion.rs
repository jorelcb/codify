//! `006` — que la aplicación vea lo que ya está instalado.
//!
//! Estos tests usan **fuentes de mentira**. Eso los hace rápidos y deterministas, y significa que
//! **no prueban el descubrimiento real**: mirar el `PATH` de verdad daría un resultado distinto en
//! cada máquina y en CI. Está declarado en el plan como renuncia, y lo cierra una persona con su
//! equipo delante.

use codify_core::application::deteccion::{
    Deteccion, Distancia, DondeInfiere, FuenteDeDeteccion, ModeloDisponible, Procedencia,
};
use codify_core::domain::error::Result;
use std::time::Duration;

/// Una fuente que devuelve lo que se le diga, sin tocar la máquina.
struct FuenteDeMentira {
    nombre: &'static str,
    modelos: Vec<ModeloDisponible>,
    tarda: Duration,
}

#[async_trait::async_trait]
impl FuenteDeDeteccion for FuenteDeMentira {
    fn nombre(&self) -> &'static str {
        self.nombre
    }
    async fn buscar(&self) -> Result<Vec<ModeloDisponible>> {
        tokio::time::sleep(self.tarda).await;
        Ok(self.modelos.clone())
    }
}

fn servidor(nombre: &str) -> ModeloDisponible {
    ModeloDisponible::sostenido_local(nombre, Procedencia::Servidor, Distancia::Usable)
}

fn programa(nombre: &str) -> ModeloDisponible {
    ModeloDisponible::declarado(
        nombre,
        Procedencia::Programa,
        Distancia::Instalado,
        DondeInfiere::Fuera,
    )
}

/// **FR-010, FR-010a, SC-005** — nada marcado como local infiere fuera del equipo.
///
/// Lo que clasifica a un proveedor **no es dónde corre el programa sino dónde ocurre la
/// inferencia**. Un servidor hallado por sondeo de loopback es local porque no pudo ser otra cosa;
/// un binario del `PATH` solo lo declara, y por eso no se cablea como local.
#[tokio::test]
async fn nada_marcado_como_local_infiere_fuera() {
    let sostenido = servidor("llama.cpp");
    assert_eq!(sostenido.donde_infiere(), DondeInfiere::Dentro);
    assert!(
        sostenido.se_puede_cablear_como_local(),
        "lo que llega por una fuente que sostiene la propiedad sí se cablea"
    );

    // Un programa del `PATH` puede ser un cliente de un servicio remoto: `claude` corre en la
    // máquina e infiere fuera. Que esté en la lista es correcto; que se cablee como local, no.
    let declarado = programa("claude");
    assert!(
        !declarado.se_puede_cablear_como_local(),
        "una clasificación declarada no basta para cablear: hay que poder sostenerla"
    );

    // Y lo que sí infiere dentro pero solo lo declara, tampoco.
    let declarado_dentro = ModeloDisponible::declarado(
        "algo",
        Procedencia::Programa,
        Distancia::Instalado,
        DondeInfiere::Dentro,
    );
    assert!(
        !declarado_dentro.se_puede_cablear_como_local(),
        "declarar que infiere dentro no es sostenerlo"
    );
}

/// **FR-002** — las dos clases de hallazgo producen la distancia que les toca.
///
/// Un servidor que responde está **usable**; un binario del `PATH` está **instalado**. Están a
/// distinta distancia de poder usarse, y confundirlos le haría creer al usuario que algo sirve
/// cuando todavía hay que arrancarlo.
#[tokio::test]
async fn cada_fuente_produce_la_distancia_que_le_toca() {
    let fuentes: Vec<std::sync::Arc<dyn FuenteDeDeteccion>> = vec![
        std::sync::Arc::new(FuenteDeMentira {
            nombre: "servidores",
            modelos: vec![servidor("llama.cpp")],
            tarda: Duration::ZERO,
        }),
        std::sync::Arc::new(FuenteDeMentira {
            nombre: "programas",
            modelos: vec![programa("claude")],
            tarda: Duration::ZERO,
        }),
    ];

    let d: Deteccion = codify_core::application::deteccion::detectar(&fuentes).await;
    assert!(!d.incompleta);
    assert_eq!(d.modelos.len(), 2);
    // Lo más cerca de poder usarse va primero.
    assert_eq!(d.modelos[0].distancia(), Distancia::Usable);
    assert_eq!(d.modelos[1].distancia(), Distancia::Instalado);
}

/// **FR-003** — el `PATH` se recorre entero, no solo las rutas habituales.
///
/// Compilar desde fuente es normal en este terreno: en la máquina donde se escribió este spec,
/// `llama-server` vive en `~/src/llama.cpp/build/bin`, fuera de todo sitio estándar. Una detección
/// que solo mire rutas conocidas no encuentra nada de eso.
#[tokio::test]
async fn el_path_se_recorre_entero() {
    use codify_core::infrastructure::deteccion::programas::ProgramasEnElPath;

    let raro = std::env::temp_dir().join(format!("codify-ruta-rara-{}", std::process::id()));
    std::fs::create_dir_all(&raro).expect("crear la ruta rara");
    let binario = raro.join("llama-server");
    std::fs::write(&binario, b"#!/bin/sh\n").expect("escribir");
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        std::fs::set_permissions(&binario, std::fs::Permissions::from_mode(0o755))
            .expect("permisos");
    }

    let fuente = ProgramasEnElPath::con_path(vec![raro.clone()]);
    let hallados = fuente.buscar().await.expect("buscar");
    std::fs::remove_dir_all(&raro).ok();
    assert!(
        hallados.iter().any(|m| m.nombre() == "llama-server"),
        "no se encontró un binario conocido en un directorio dado: {hallados:?}"
    );

    // Y la otra mitad, que es la que de verdad importa: **se mira el `PATH` entero**, no una lista
    // de rutas habituales. Sin esta comprobación el test pasaba construyendo la fuente con rutas
    // explícitas, y romper `del_entorno` no lo hacía caer — lo descubrió una inyección.
    let del_entorno = ProgramasEnElPath::del_entorno();
    let esperados: Vec<std::path::PathBuf> = std::env::var_os("PATH")
        .map(|p| std::env::split_paths(&p).collect())
        .unwrap_or_default();
    assert_eq!(
        del_entorno.directorios(),
        esperados.as_slice(),
        "no se recorre el `PATH` entero: compilar desde fuente deja los binarios fuera de toda \
         ruta habitual, y así no se encontrarían"
    );

    // Y **nada de lo que sale de esta fuente se puede cablear como local**. El test de la
    // invariante construye los modelos a mano, así que no veía si la fuente la respeta: romperla
    // no lo hacía caer. Lo descubrió una inyección.
    assert!(
        hallados.iter().all(|m| !m.se_puede_cablear_como_local()),
        "un binario del `PATH` no puede sostener dónde infiere, así que no se cablea como local"
    );
}

/// **FR-007, FR-008, SC-004** — la detección tiene un límite y lo declara.
///
/// Una fuente que no contesta no puede dejar la aplicación esperando: se abandona **diciéndolo**.
/// Una espera muda es indistinguible de que no pase nada, y eso ya costó un ciclo entero en esta
/// misma aplicación.
#[tokio::test]
async fn la_deteccion_respeta_su_presupuesto() {
    use codify_core::application::deteccion::detectar_con;

    let fuentes: Vec<std::sync::Arc<dyn FuenteDeDeteccion>> =
        vec![std::sync::Arc::new(FuenteDeMentira {
            nombre: "la que nunca contesta",
            modelos: vec![servidor("fantasma")],
            tarda: Duration::from_secs(30),
        })];

    let empezo = std::time::Instant::now();
    let d = detectar_con(&fuentes, Duration::from_millis(80)).await;

    assert!(
        empezo.elapsed() < Duration::from_secs(1),
        "la detección esperó a una fuente que no contesta: {:?}",
        empezo.elapsed()
    );
    assert!(
        d.incompleta,
        "no se dijo que la búsqueda quedó a medias: abandonar en silencio es peor que tardar"
    );
}
