import type { Meta, StoryObj } from '@storybook/vue3'
import TraitList from './TraitList.vue'

const meta: Meta<typeof TraitList> = {
  title: 'Components/TraitList',
  component: TraitList,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof TraitList>

export const Positive: Story = {
  args: { traits: ['Clutch'] },
}

export const Negative: Story = {
  args: { traits: ['Choker'] },
}

export const Mixed: Story = {
  args: { traits: ['Clutch', 'Injury Prone'] },
}

export const Empty: Story = {
  args: { traits: [] },
}

export const AllVariants: Story = {
  render: () => ({
    components: { TraitList },
    template: `
      <div style="display:flex;flex-direction:column;gap:1rem;padding:1.5rem;background:#0d1117;font-size:0.9375rem">
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;min-width:120px;font-size:0.75rem">positive only</span>
          <TraitList :traits="['Clutch']" />
        </div>
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;min-width:120px;font-size:0.75rem">negative only</span>
          <TraitList :traits="['Choker']" />
        </div>
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;min-width:120px;font-size:0.75rem">mixed (max 2)</span>
          <TraitList :traits="['Clutch', 'Injury Prone']" />
        </div>
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;min-width:120px;font-size:0.75rem">empty</span>
          <TraitList :traits="[]" />
        </div>
      </div>
    `,
  }),
}
