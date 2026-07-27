import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

// Dev-server proxy mirrors the production Caddyfile's routing exactly:
// /api, /auth, /widget go to the Go backend, everything else is the SPA.
// This keeps the session cookie same-origin in both environments, which
// matters because it's SameSite=Lax and would otherwise be dropped on
// cross-origin fetch/XHR calls from the Vite dev server's own port.
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.svg', 'apple-touch-icon.png'],
      manifest: {
        name: 'Spotify Insights',
        short_name: 'Insights',
        description: 'A personal, always-on Spotify Wrapped.',
        theme_color: '#141217',
        background_color: '#141217',
        display: 'standalone',
        start_url: '/',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png' },
          { src: '/icons/icon-512-maskable.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
    }),
  ],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/auth': 'http://localhost:8080',
      '/widget': 'http://localhost:8080',
    },
  },
})
