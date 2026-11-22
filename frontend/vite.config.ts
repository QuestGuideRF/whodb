import fs from 'fs';
import {defineConfig} from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import type { Plugin, ResolvedConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite'
import yamlPlugin from './plugins/vite-plugin-yaml';

const baseHrefPlaceholder = '__WHODB_BASE_HREF__';
const frontendEditionMarkerFile = '.whodb-edition';

const resolveBuildSourcemap = (): boolean | 'hidden' | 'inline' => {
  switch (process.env.WHODB_BUILD_SOURCEMAP) {
    case 'true':
      return true;
    case 'false':
      return false;
    case 'hidden':
    case 'inline':
      return process.env.WHODB_BUILD_SOURCEMAP;
    default:
      return process.env.NODE_ENV === 'production' ? 'hidden' : true;
  }
};

const htmlMetaPlugin = () => {
  const title = 'Clidey WhoDB';
  const description = 'WhoDB is the next-generation database explorer';
  return {
    name: 'html-meta',
    transformIndexHtml(html: string) {
      return html
        .replace('%VITE_APP_TITLE%', title)
        .replace('%VITE_APP_DESCRIPTION%', description);
    }
  };
};

const frontendEditionPlugin = (edition: string): Plugin => {
  let resolvedConfig: ResolvedConfig | null = null;

  return {
    name: 'frontend-edition',
    apply: 'build',
    configResolved(config) {
      resolvedConfig = config;
    },
    closeBundle() {
      if (resolvedConfig == null) {
        return;
      }
      const outDir = path.resolve(resolvedConfig.root, resolvedConfig.build.outDir);
      fs.writeFileSync(path.join(outDir, frontendEditionMarkerFile), `${edition}\n`);
    },
  };
};

export default defineConfig(async ({command}) => {

  let istanbulPlugin = null;
  if (process.env.NODE_ENV === 'test') {
    try {
      const { default: istanbul } = await import('vite-plugin-istanbul');

      istanbulPlugin = istanbul({
        requireEnv: false,
        include: ['src/**/*.{js,jsx,ts,tsx}'],
        exclude: [
          'node_modules',
          '**/*.d.ts',
          '**/*.test.{js,jsx,ts,tsx}',
          '**/*.spec.{js,jsx,ts,tsx}',
          'src/generated/**',
          'src/index.tsx'
        ],
        cwd: process.cwd(),
      });
    } catch (e) {
      console.warn('Failed to load vite-plugin-istanbul:', e);
    }
  }

    return {
      plugins: [
        yamlPlugin(),
        react(),
        tailwindcss(),
        {
          name: 'base-href-placeholder',
          transformIndexHtml(html: string) {
            if (command === 'build') {
              return html;
            }
            return html.replace(baseHrefPlaceholder, '/');
          }
        },
        htmlMetaPlugin(),
        frontendEditionPlugin('ce'),
        istanbulPlugin
      ].filter(Boolean),

    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
        '@graphql': path.resolve(__dirname, './src/generated/graphql.ts'),
      },
    },
    server: {
      port: parseInt(process.env.VITE_PORT || '3000'),
      open: process.env.NODE_ENV !== 'test',
      proxy: {
        '/api': {
          target: `http://localhost:${process.env.VITE_BACKEND_PORT || '8080'}`,
          changeOrigin: true,
        },
      },
    },
    base: command === 'build' ? './' : '/',
    build: {
      outDir: 'build',
      sourcemap: resolveBuildSourcemap(),
      chunkSizeWarningLimit: 1000,
    },
    define: {
      'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV),
      '__APP_VERSION__': JSON.stringify(process.env.VITE_APP_VERSION || 'development'),
    },
  };
});