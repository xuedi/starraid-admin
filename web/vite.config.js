import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

// Dev: Vite serves the SPA at :5173 and proxies /api + /healthz to the Go admin
// backend (:8090) so fetches hit real data with hot reload. Built output goes to
// web/dist, which the backend serves from disk in production.
export default defineConfig({
  plugins: [tailwindcss(), svelte()],
  build: { outDir: 'dist' },
  server: {
    proxy: {
      '/api': 'http://localhost:8090',
      '/healthz': 'http://localhost:8090',
    },
  },
})
