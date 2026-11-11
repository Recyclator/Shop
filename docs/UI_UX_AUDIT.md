# Auditoría UI/UX — Shop / Nexora

Fecha: 2026-11-11. Base: `0f66a7d8772ec02ab46ac0c3fc4f2a1309efe503`.

## Alcance y evidencia

Inspección del layout, 11 plantillas de página, navegación compartida, CSS, notificaciones, validación y patrones JavaScript de productos, POS, separados y permisos. Go/Fiber + templ, JavaScript nativo, Alpine local y HTMX externo; Tailwind genera utilidades. Se conserva la identidad monocromática y los dos temas del área administrativa, así como la identidad azul del acceso. No hay cambios de API, permisos, modelos ni reglas de negocio.

Esto es una mejora incremental comprobada mediante compilación y DOM; **no es una certificación visual, WCAG ni una validación de extremo a extremo**. Las páginas se renderizaron sin base de datos; no se simularon resultados de ventas ni de autenticación real.

## Hallazgos y resolución

| Severidad | Evidencia / causa raíz | Acción y estado |
| --- | --- | --- |
| Alta | `views/pages/products.templ` y `views/partials/topbar.templ` duplicaban `search-input`; `static/js/products.js` selecciona el primero. Reproducido en HTML renderizado. | ID separado para la barra superior; el selector existente apunta al buscador de productos. Prueba de regresión. |
| Alta | `Base.confirmDialog`: listener global de Enter aceptaba la acción con foco fuera del diálogo. Reproducido en DOM con el código original. Sin aislamiento ni restitución del foco. | Diálogo HTML nativo con título y descripción, foco inicial en Cancelar, Escape y restitución; resolución única. Comprobados los manejadores en DOM; aislamiento nativo y teclado real pendientes. |
| Alta | `Base` bloqueaba zoom; `Topbar` usaba un div clicable. Hechos comprobados. | Zoom habilitado, enlace para saltar al contenido y botón de usuario con estado expandido y Escape. |
| Alta | `static/css/app-ui.css` estaba vacío aunque el layout lo carga y el proyecto usa utilidades. Hecho comprobado. | Asset regenerado con el build existente. Impacto visual exacto pendiente de navegador. |
| Media | Etiquetas sin `for`, errores compartidos entre campos hermanos en `FormValidator.showError`, sin relación accesible. | Asociaciones explícitas en plantillas; errores por campo mediante WeakMap, `aria-invalid` y `aria-describedby`, preservando ayudas existentes. 56 asociaciones verificadas en el HTML renderizado. |
| Media | `--text-mut` carecía de contraste suficiente; placeholders reducían además su opacidad. | Ajustados neutros conservando matiz; comprobación matemática ≥4.5:1 sobre las cuatro superficies sólidas de cada tema. No implica aprobación de todos los colores, transparencias ni estados. |
| Media | `Toast.show`, confirmaciones y alerta de acceso interpolaban mensajes en HTML. | Texto mediante `textContent`; pruebas de contenido literal. Riesgo de inyección según procedencia del mensaje: no se demostró explotación de extremo a extremo. |
| Media | Toast sin nombre de cierre ni anuncio; ancho fijo y expiración durante interacción. | Roles status/alert, cierre nombrado, ancho adaptable, pausa del temporizador por foco/hover y limpieza al cerrar. Anuncio real en lector pendiente. |
| Media | CSS de formularios repetido (40/44 px), objetivos de cierre pequeños, ausencia de reduced-motion compartido. | Eliminada definición contradicha; token de altura 44 px; foco visible; cierres ≥32 px; movimiento reducido. |
| Media | Cabeceras/paginación rígidas y área segura móvil no contemplada. | Wrap, límites de filtros, fuentes móviles, espacio para navegación inferior y regiones desplazables en tablas. Pendiente de inspección a 320/375/768/1440 px. |

No se confirmó ningún hallazgo crítico dentro de este alcance. Las hipótesis de problemas de rendimiento no se presentan como resultados medidos.

## Decisiones y archivos

- `static/css/app-main.css`, `app-components.css`: mismos fondos, tipografías y estructura; neutros más legibles, controles y estados compartidos, ajuste de selector móvil para conservar avatar.
- `views/layout/base.templ`: diálogo nativo sin dependencia, semántica principal y zoom. Conserva contrato Promise<boolean> y wrappers existentes.
- `views/partials/topbar.templ`: disclosure con botón nativo, sin imponer un patrón ARIA menu que requiera navegación adicional.
- `views/pages/*.templ`: etiquetas y cierres; acceso por botones nativos y autocompletado en login/registro; tablas con región desplazable cuando corresponde.
- `static/js/validator.js`, `toast.js`: correcciones compartidas sin cambiar criterios de validación ni operaciones de negocio.
- `*_templ.go`: regenerados con templ v0.3.1020, versión declarada en go.mod. Cambios adicionales en código generado proceden de la actualización del generador desde v0.3.1001.
- `tests/ui/`: renderizador aislado y seis pruebas de DOM. Ninguna dependencia nueva de la aplicación.

## Verificación ejecutada

- Referencia inicial: `go test ./...` pasó antes de modificar plantillas generadas.
- `make build`: Tailwind v4.3.5, templ v0.3.1020 y compilación Go correctos.
- `go test ./...`, `go vet ./...`: correctos tras los cambios.
- Sintaxis de los siete archivos JavaScript, scripts inline en 11 páginas y `git diff --check`: correctos.
- Seis pruebas DOM: HTML/IDs/etiquetas, independencia de errores, mensajes literales, confirmación/cancelación y foco, disclosure y contraste de tokens: correctas.
- El test de diálogo usa stubs para showModal/close en jsdom: no valida top layer, tabulación nativa ni distribución visual.
- Intentos de Chromium: descarga Playwright incompleta; Chromium del sistema no inicia (`socket() failed: Operation not permitted`). Sin capturas, consola de la aplicación ejecutada, Lighthouse ni lector de pantalla.
- Rendimiento: no se midieron LCP, INP ni CLS; no se afirma aceleración. Restaurar el CSS generado aumenta los bytes respecto al asset vacío y corrige una omisión de build, no es una optimización de transferencia.

### Reproducir

Usar Go 1.25, Node, templ v0.3.1020 y las dependencias existentes. El npm lock está ignorado por el repositorio; la reproducibilidad exacta de Tailwind es deuda previa.

```sh
npm install
make build
go test ./...
go vet ./...
mkdir -p /tmp/shop-ui-fixtures
go run tests/ui/render.go /tmp/shop-ui-fixtures
npm install --prefix /tmp/shop-ui-tools --ignore-scripts jsdom@31.0.0
JSDOM_MODULE=/tmp/shop-ui-tools/node_modules/jsdom UI_FIXTURES=/tmp/shop-ui-fixtures node --test tests/ui/dom.test.cjs
```

## Pendientes priorizados y veredicto del diff

1. **Alta / esfuerzo medio:** verificar escritorio/móvil y ambos temas, con zoom 200–400%, teclado, lector de pantalla y datos reales; especial atención al CSS restaurado y diálogo nativo. Mantener PR en borrador hasta esta revisión.
2. **Alta / esfuerzo medio:** extender gestión de foco y semántica a modales específicos de edición y POS; revisar listados/autocomplete y etiquetas generadas dinámicamente en productos. No están cubiertos por la corrección del confirmador compartido.
3. **Media / esfuerzo bajo:** decidir el comportamiento de la búsqueda global (carece de implementación propia); errores y estados vacíos explícitos en `roles-permissions.js`, donde existen catch vacíos.
4. **Media / esfuerzo medio:** revisar contraste integral, alertas temporizadas, botones iconográficos y estilos inline de POS/productos; consolidar gradualmente, con capturas de referencia.
5. **Media / esfuerzo medio:** fijar toolchain/lockfile, automatizar pruebas de navegador y medir recursos externos, fuentes y Core Web Vitals antes de optimizar.

**Veredicto:** cambios coherentes con el alcance y comprobaciones automáticas ejecutadas sin fallos; no se modificó lógica de negocio. Listo para revisión técnica, **pendiente de validación visual e integración para aprobar su puesta en producción**.
