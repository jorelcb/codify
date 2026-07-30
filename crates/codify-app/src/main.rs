// Oculta la consola en Windows para builds de release.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    codify_app::run()
}
