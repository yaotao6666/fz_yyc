import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import { setupRouter } from './router'
import { createPinia } from 'pinia'
import './styles/index.scss'

const app = createApp(App)
const pinia = createPinia()
const router = setupRouter(pinia)

app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.mount('#app')
