import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  root: __dirname,
  plugins: [react()],
  server: {
    proxy: {
      '/metrics': process.env.GATEWAY_PROXY_TARGET || 'http://localhost:8080',
      '/alerts': process.env.GATEWAY_PROXY_TARGET || 'http://localhost:8080',
      '/users': process.env.GATEWAY_PROXY_TARGET || 'http://localhost:8080',
      '/auth/': process.env.GATEWAY_PROXY_TARGET || 'http://localhost:8080',
      '/sessions': process.env.GATEWAY_PROXY_TARGET || 'http://localhost:8080',
    },
  },
  build: {
    outDir: path.resolve(__dirname, '../../../dist/dashboard'),
    emptyOutDir: true,
  },
  test: {
    environment: 'jsdom',
    environmentOptions: {
      jsdom: {
        url: 'http://localhost:3000/',
      },
    },
    globals: true,
    setupFiles: [path.resolve(__dirname, 'test-setup.js')],
  },
});
