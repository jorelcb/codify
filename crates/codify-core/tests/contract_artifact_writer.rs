//! **Contract test del port `ArtifactWriter`** (T007), contra el adapter real y el fake.
//!
//! Este port es el que cierra la deuda de que el núcleo generaba contexto y nunca lo
//! entregaba. Las reglas que se asertan son de dominio: no salirse del repositorio, declarar
//! lo ocurrido, y que **un fallo aislado no arrastre al resto**.

mod fakes;

use codify_core::application::ports::ArtifactWriter;
use codify_core::infrastructure::repo::writer::FsArtifactWriter;
use fakes::FakeArtifactWriter;
use std::path::PathBuf;

fn temp_repo(tag: &str) -> PathBuf {
    let dir = std::env::temp_dir().join(format!("codify-writer-{tag}-{}", std::process::id()));
    let _ = std::fs::remove_dir_all(&dir);
    std::fs::create_dir_all(&dir).unwrap();
    dir
}

async fn artifact_writer_contract(writer: &dyn ArtifactWriter, label: &str) {
    // 1. Escribe y el registro dice que llegó al disco.
    let record = writer.write("AGENTS.md", "# Contexto").await;
    assert!(record.reached_disk(), "[{label}] debe llegar al disco");
    assert_eq!(record.path, "AGENTS.md");
    assert_eq!(record.bytes, "# Contexto".len());

    // 2. Lo escrito se puede releer.
    let back = writer.read_existing("AGENTS.md").await.unwrap();
    assert_eq!(back.as_deref(), Some("# Contexto"), "[{label}] relectura");

    // 3. Un archivo que no existe no es un error: es `None`.
    assert_eq!(writer.read_existing("no-existe.md").await.unwrap(), None);

    // 4. Crea directorios intermedios (`context/` puede no existir).
    let nested = writer.write("context/CONTEXT.md", "arquitectura").await;
    assert!(
        nested.reached_disk(),
        "[{label}] debe crear el directorio intermedio"
    );
    assert_eq!(
        writer
            .read_existing("context/CONTEXT.md")
            .await
            .unwrap()
            .as_deref(),
        Some("arquitectura")
    );
}

#[tokio::test]
async fn artifact_writer_contract_holds_for_the_real_fs_adapter() {
    let dir = temp_repo("contract");
    let writer = FsArtifactWriter::new(&dir);
    artifact_writer_contract(&writer, "fs").await;
    // Comprobación independiente del port: el archivo existe de verdad.
    assert!(dir.join("context/CONTEXT.md").exists());
    let _ = std::fs::remove_dir_all(&dir);
}

#[tokio::test]
async fn artifact_writer_contract_holds_for_the_fake() {
    artifact_writer_contract(&FakeArtifactWriter::new(), "fake").await;
}

/// El agente no puede escribir fuera del repositorio objetivo.
#[tokio::test]
async fn real_writer_refuses_to_escape_the_repository() {
    let dir = temp_repo("escape");
    let writer = FsArtifactWriter::new(&dir);

    for escape in ["../fuera.md", "/tmp/absoluto.md", "context/../../fuera.md"] {
        let record = writer.write(escape, "x").await;
        assert!(!record.reached_disk(), "'{escape}' no debe escribirse");
    }
    let _ = std::fs::remove_dir_all(&dir);
}

/// Un fallo aislado se registra y **no** impide escribir el resto.
#[tokio::test]
async fn a_single_failure_does_not_drag_the_rest_down() {
    let writer = FakeArtifactWriter::new().failing_on("context/CONTEXT.md");

    let ok = writer.write("AGENTS.md", "a").await;
    let bad = writer.write("context/CONTEXT.md", "b").await;
    let after = writer.write("context/DEVELOPMENT_GUIDE.md", "c").await;

    assert!(ok.reached_disk());
    assert!(!bad.reached_disk(), "el fallo se registra, no se oculta");
    assert!(
        bad.summary().contains("falló"),
        "el motivo viaja: {}",
        bad.summary()
    );
    assert!(
        after.reached_disk(),
        "el fallo anterior no debe abortar lo siguiente"
    );
}
