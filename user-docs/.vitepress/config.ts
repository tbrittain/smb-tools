import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'smb-tools',
  description: 'User documentation for smb-tools — Super Mega Baseball 4 franchise history tracker.',
  base: '/smb-tools/',
  lang: 'en-US',
  lastUpdated: true,
  head: [['link', { rel: 'icon', href: '/smb-tools/logo.png' }]],
  themeConfig: {
    logo: '/logo.png',
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Getting Started', link: '/getting-started' },
      { text: 'User Guide', link: '/user-guide' },
    ],
    sidebar: [
      {
        text: 'Guide',
        items: [
          { text: 'Getting Started', link: '/getting-started' },
          { text: 'User Guide', link: '/user-guide' },
        ],
      },
      {
        text: 'Core Workflows',
        items: [
          { text: 'Save Game Setup & Season Sync', link: '/save-game-setup' },
          { text: 'Importing from SmbExplorerCompanion', link: '/legacy-migration' },
          { text: 'Franchise Forking', link: '/franchise-forking' },
        ],
      },
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/tbrittain/smb-tools' },
    ],
  },
})
