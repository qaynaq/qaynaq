import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from "path"

export default defineConfig({
    plugins: [react()],
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
                    if (id.includes('node_modules')) {
                        if (id.includes('react') || id.includes('react-dom')) {
                            return 'vendor-react';
                        }
                        if (id.includes('@radix-ui')) {
                            return 'vendor-ui';
                        }
                        if (id.includes('js-yaml')) {
                            return 'vendor-yaml';
                        }
                        if (id.includes('lucide-react')) {
                            return 'vendor-icons';
                        }
                        if (id.includes('@xyflow')) {
                            return 'vendor-xyflow';
                        }
                        return 'vendor-misc';
                    }
                    if (id.includes('inline-yaml-editor')) {
                        if (id.includes('components/')) {
                            return 'yaml-components';
                        }
                        if (id.includes('utils/')) {
                            return 'yaml-utils';
                        }
                        return 'yaml-editor';
                    }
                }
            }
        },
        chunkSizeWarningLimit: 400
    }
}) 