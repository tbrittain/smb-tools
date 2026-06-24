import '@fontsource/ibm-plex-sans/400.css'
import '@fontsource/ibm-plex-sans/600.css'
import '@fontsource/ibm-plex-sans/700.css'
import '@fontsource/ibm-plex-mono/400.css'
import Aura from '@primeuix/themes/aura'
import PrimeVue from 'primevue/config'
import { createMemoryHistory, createRouter } from 'vue-router'
import type { Preview } from '@storybook/vue3-vite'
import { setup } from '@storybook/vue3-vite'
import 'primeicons/primeicons.css'
import '../src/assets/tokens.css'
import '../src/style.css'


// Minimal router so RouterLink works in stories without throwing
const storybookRouter = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
})

// Wails injects window.runtime only inside the desktop webview. Stub it so any
// component calling wailsjs/runtime functions (BrowserOpenURL, etc.) no-ops
// in Storybook instead of throwing "Cannot read properties of undefined".
;(window as unknown as { runtime: unknown }).runtime ??= new Proxy(
  {},
  { get: () => () => {} },
)

// PrimeVue's dark-mode CSS variables are emitted scoped to `:root.dark`
// (matching main.ts, which adds 'dark' to document.documentElement) — a
// `.dark` class on a wrapper div further down the tree doesn't satisfy that
// selector, so PrimeVue components (DataTable, Paginator, etc.) silently keep
// their light-theme `--p-*` tokens even though our own --color-* tokens
// (a plain `.dark` selector) pick up correctly from the decorator below.
document.documentElement.classList.add('dark')

setup((app) => {
  app.use(PrimeVue, {
    theme: {
      preset: Aura,
      options: {
        darkModeSelector: '.dark',
        cssLayer: false,
      },
    },
  })
  app.use(storybookRouter)
})

const preview: Preview = {
  decorators: [
    (story) => ({
      components: { story },
      // The background is set inline (rather than relying on the backgrounds
      // addon's `parameters.backgrounds`, which isn't registered in main.ts)
      // so components styled for dark text are legible by default instead of
      // sitting on Storybook's white canvas. 'dark' itself is applied to
      // document.documentElement above, not here — see that comment.
      template: '<div style="background:#0d1117;min-height:100vh;padding:1rem"><story /></div>',
    }),
  ],
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    a11y: {
      test: 'todo',
    },
  },
}

export default preview
