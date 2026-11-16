import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  	plugins: [react()],
	resolve:{
		alias:{
			'@':resolve(__dirname, 'src')
		}
	},
	server: {
		open: true,//启动项目自动弹出浏览器
		host: "0.0.0.0",
		port: 3300,//启动端口
		proxy: {
			'/api': {
				// target: 'http://10.10.103.96:2234',
				target: 'http://localhost:9900',
				changeOrigin: true,
				rewrite: (path) => path.replace(/^\/api/, '')
			},
		}
	}
})
