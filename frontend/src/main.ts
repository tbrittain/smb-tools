import 'primeicons/primeicons.css'
import '@fontsource/ibm-plex-sans/400.css'
import '@fontsource/ibm-plex-sans/600.css'
import '@fontsource/ibm-plex-sans/700.css'
import '@fontsource/ibm-plex-mono/400.css'
import Aura from '@primeuix/themes/aura'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import Tooltip from 'primevue/tooltip'
import { createApp } from 'vue'
import { LogFrontendError } from '../wailsjs/go/main/App'
import App from './App.vue'
import router from './router'
import './assets/tokens.css'
import './style.css'

document.documentElement.classList.add('dark')

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ConfirmationService)
app.use(ToastService)
app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '.dark',
      cssLayer: false,
    },
  },
})

app.directive('tooltip', Tooltip)

app.config.errorHandler = (err, _instance, info) => {
  const msg = err instanceof Error ? err.message : String(err)
  const stack = err instanceof Error ? (err.stack ?? '') : ''
  LogFrontendError(msg, stack, info ?? '')
}

window.addEventListener('unhandledrejection', (event) => {
  const r = event.reason
  const msg = r instanceof Error ? r.message : String(r)
  const stack = r instanceof Error ? (r.stack ?? '') : ''
  LogFrontendError(msg, stack, 'unhandledrejection')
})

app.mount('#app')
