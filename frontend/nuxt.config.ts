import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers';

export default defineNuxtConfig({
  compatibilityDate: '2024-11-18',
  devtools: { enabled: true },

  // 环境变量配置
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE || 'http://localhost:8080', // 公开的 API 基础路径
    },
  },

  // 开发服务器配置
  devServer: {
    port: 11451,
    host: 'localhost',
  },

  // 加载 Nuxt 模块
  modules: [
    'nuxtjs-naive-ui',
    '@nuxt/content',
  ],

  // 注册插件
  plugins: [
    '~/plugins/v-md-editor.js',
  ],

  // Vite 插件配置
  vite: {
    plugins: [
      AutoImport({
        imports: [
          {
            'naive-ui': [
              'useDialog',
              'useMessage',
              'useNotification',
              'useLoadingBar',
            ],
          },
        ],
      }),
      Components({
        resolvers: [NaiveUiResolver()],
      }),
    ],
  },

  // 全局样式文件
  css: [
    '~/assets/css/tailwind.css',
    '@kangc/v-md-editor/lib/style/base-editor.css',
    '@kangc/v-md-editor/lib/theme/style/vuepress.css',
  ],

  // PostCSS 配置
  postcss: {
    plugins: {
      tailwindcss: {},
      autoprefixer: {},
    },
  },
});
