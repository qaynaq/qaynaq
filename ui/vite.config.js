import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import viteCompression from 'vite-plugin-compression'
import path from "path"

export default defineConfig({
    plugins: [
        react(),
        viteCompression({
            algorithm: 'brotliCompress',
            ext: '.br',
            deleteOriginFile: false,
            threshold: 1024,
            filter: /\.(js|css|html|svg|json)$/i,
        }),
        viteCompression({
            algorithm: 'gzip',
            ext: '.gz',
            deleteOriginFile: false,
            threshold: 1024,
            filter: /\.(js|css|html|svg|json)$/i,
        }),
    ],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        }
    },
    server: {
        proxy: {
            '/api': 'http://localhost:8080',
            '/auth': 'http://localhost:8080',
            '/connections/oauth': 'http://localhost:8080',
        }
    },
    build: {
        rollupOptions: {
            output: {
                manualChunks: (id) => {
                    if (!id.includes('node_modules')) {
                        return;
                    }
                    if (id.includes('/lucide-react/')) {
                        return 'vendor-icons';
                    }
                    if (id.includes('/js-yaml/')) {
                        return 'vendor-yaml';
                    }
                },
            },
        },
        chunkSizeWarningLimit: 600,
    },
})
