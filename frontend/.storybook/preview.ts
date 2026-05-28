import '@fontsource/ibm-plex-sans/400.css'
import '@fontsource/ibm-plex-sans/600.css'
import '@fontsource/ibm-plex-sans/700.css'
import '@fontsource/ibm-plex-mono/400.css'
import Aura from '@primeuix/themes/aura'
import PrimeVue from 'primevue/config'
import { createMemoryHistory, createRouter } from 'vue-router'
import type { Preview } from '@storybook/vue3-vite'
import { setup } from '@storybook/vue3-vite'
import '../src/assets/tokens.css'
import '../src/style.css'


// Minimal router so RouterLink works in stories without throwing
const storybookRouter = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
})

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
  // Wrap every story in a .dark div so CSS variables resolve correctly
  decorators: [
    (story) => ({
      components: { story },
      // .dark activates PrimeVue's dark mode selector and makes --color-*
      // variables available. The wrapper is otherwise unstyled so Storybook's
      // own canvas controls the background.
      template: '<div class="dark"><story /></div>',
    }),
  ],
  parameters: {
    backgrounds: {
      default: 'dark',
      values: [{ name: 'dark', value: '#0d1117' }],
    },
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
