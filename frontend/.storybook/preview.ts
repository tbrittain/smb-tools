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
      template: `
        <div class="dark" style="background: var(--color-bg); min-height: 100vh; padding: 1.5rem;">
          <story />
        </div>
      `,
    }),
  ],
  parameters: {
    backgrounds: { disable: true }, // we control bg via .dark wrapper
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
