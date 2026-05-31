/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./views/**/*.templ",
    "./views/**/*.go",
    "./internal/handlers/**/*.go",
    "./static/js/**/*.js",
  ],
  theme: {
    extend: {
      colors: {
        // Mapeo de tus colores actuales a Tailwind
        'app-bg': '#0c0d0f',
        'app-bg-elev': '#131416',
        'app-bg-surf': '#18191d',
        'app-accent': '#3b82f6',
      }
    },
  },
  plugins: [],
  darkMode: 'class', // Soporte para el tema oscuro que ya tienes
}
