import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  base: './', // Use relative paths for GitHub Pages
  build: {
    outDir: 'dist-demo',
    rollupOptions: {
      input: {
        main: path.resolve(__dirname, 'index.html'),
      },
    },
  },
  define: {
    'process.env.DEMO_MODE': JSON.stringify('true'),
  },
});