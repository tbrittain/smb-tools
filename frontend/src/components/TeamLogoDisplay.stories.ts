import type { Meta, StoryObj } from '@storybook/vue3'
import TeamLogoDisplay from './TeamLogoDisplay.vue'

// Minimal 1×1 PNG for story previews — no live Wails binding needed.
const SAMPLE_LOGO =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='

const meta: Meta<typeof TeamLogoDisplay> = {
  title: 'Components/TeamLogoDisplay',
  component: TeamLogoDisplay,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof TeamLogoDisplay>

export const WithLogoLg: Story = {
  args: { logoUrl: SAMPLE_LOGO, size: 'lg' },
}

export const WithLogoSm: Story = {
  args: { logoUrl: SAMPLE_LOGO, size: 'sm' },
}

export const NoLogoLg: Story = {
  args: { logoUrl: '', size: 'lg' },
}

export const NoLogoSm: Story = {
  args: { logoUrl: '', size: 'sm' },
}

export const AllVariants: Story = {
  render: () => ({
    components: { TeamLogoDisplay },
    setup: () => ({ sampleLogo: SAMPLE_LOGO }),
    template: `
      <div style="display:flex;flex-direction:column;gap:1.5rem;padding:2rem;background:#0d1117">
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;width:120px;font-size:0.75rem">lg — with logo</span>
          <TeamLogoDisplay :logoUrl="sampleLogo" size="lg" />
        </div>
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;width:120px;font-size:0.75rem">lg — no logo</span>
          <TeamLogoDisplay logoUrl="" size="lg" />
        </div>
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;width:120px;font-size:0.75rem">sm — with logo</span>
          <TeamLogoDisplay :logoUrl="sampleLogo" size="sm" />
        </div>
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;width:120px;font-size:0.75rem">sm — no logo</span>
          <TeamLogoDisplay logoUrl="" size="sm" />
        </div>
      </div>
    `,
  }),
}
