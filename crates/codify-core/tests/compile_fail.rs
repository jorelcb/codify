//! Arnés de los tests que deben **fallar al compilar**.
//!
//! `003`-FR-008 pide que el egress en modo local sea *estructuralmente imposible*, no
//! simplemente que no ocurra. Una aserción normal no puede expresar eso: comprobaría una
//! ejecución, y una ejecución limpia es compatible con que la ruta exista.
//!
//! Lo que se comprueba aquí es que **el programa que la usaría no existe** — no compila. Es el
//! enunciado más fuerte de «imposible» que este lenguaje permite escribir, y el único que no
//! depende de que nadie olvide mantener una comprobación.

#[test]
fn el_grafo_local_no_admite_un_proveedor_remoto() {
    let t = trybuild::TestCases::new();
    t.compile_fail("tests/compile_fail/local_no_admite_remoto.rs");
}
