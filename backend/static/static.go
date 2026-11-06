// Package static embebe los archivos estáticos del sitio en el binario.
package static

import _ "embed"

// Styles es la hoja de estilos de la aplicación (styles.css).
//
//go:embed styles.css
var Styles []byte
