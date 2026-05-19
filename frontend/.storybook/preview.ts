import Aura from '@primeuix/themes/aura'
import PrimeVue from 'primevue/config'
import type { Preview } from '@storybook/vue3-vite'
import { setup } from '@storybook/vue3-vite'
import '../src/assets/tokens.css'
import '../src/style.css'

// Register PrimeVue globally so any story using PrimeVue components works
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
