import path from 'path'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      $public: path.resolve(__dirname, './public'),
      $shared: path.resolve(__dirname, '../shared'),
    },
  },
})
