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
      { text: 'Franchise Stat Tracking', link: '/save-game-setup' },
      { text: 'Team Transfer Tool', link: '/team-transfer' },
    ],
    sidebar: [
      {
        text: 'Introduction',
        items: [
          { text: 'Getting Started', link: '/getting-started' },
          { text: 'User Guide', link: '/user-guide' },
        ],
      },
      {
        text: 'Franchise Stat Tracking',
        items: [
          { text: 'Save Game Setup & Season Sync', link: '/save-game-setup' },
          { text: 'Importing from SmbExplorerCompanion', link: '/legacy-migration' },
          { text: 'Franchise Forking', link: '/franchise-forking' },
          { text: 'CSV Exports', link: '/csv-exports' },
        ],
      },
      {
        text: 'Team Transfer Tool',
        items: [
          { text: 'League Export & Import', link: '/team-transfer' },
          { text: 'Save Game Editor', link: '/save-game-editor' },
        ],
      },
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/tbrittain/smb-tools' },
    ],
  },
})
