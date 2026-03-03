import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { setupAxiosAuth } from './plugins/axiosAuth'
// 引入'https://webrtc.github.io/adapter/adapter-latest.js'
// import 'https://webrtc.github.io/adapter/adapter-latest.js'
// import '@/assets/css/font.css'
import '@/assets/css/chat.css'

const app = createApp(App)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
// 全局注入 axios 鉴权逻辑：自动带 Bearer、401 时触发 refresh。
setupAxiosAuth(store, router)
app.use(store).use(router).use(ElementPlus).mount('#app')
