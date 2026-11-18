import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// 后台前端开发/构建配置
// base 与后端托管路径保持一致：/admin/
// 构建产物直接输出到后端静态目录，部署时无需额外拷贝
export default defineConfig({
  plugins: [react()],
  base: '/admin/',
  build: {
    outDir: '../backend/web/admin',
    emptyOutDir: true,
  },
  server: {
    port: 5174,
    proxy: {
      '/api': {
        // 本地后端（本机 8080 被微信占用，后端起在 8081）
        target: 'http://127.0.0.1:8081',
        // 线上后端：切换时注释上一行，取消下一行注释
        // target: 'https://glancelog.cn',
        changeOrigin: true,
      },
    },
  },
});
