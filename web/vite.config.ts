import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Dev: Vite на :5173, /api проксируется на server-public (:8080).
// Prod: build → web/dist/, отдаётся server-public (fiber Static).
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
