import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'smb-tools',
  description: 'User documentation for smb-tools — Super Mega Baseball 4 franchise history tracker.',
  base: '/smb-tools/',
  lang: 'en-US',
  lastUpdated: true,
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
    ],
    sidebar: [],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/tbrittain/smb-tools' },
    ],
  },
})
