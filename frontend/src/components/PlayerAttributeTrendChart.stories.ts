import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import PlayerAttributeTrendChart from './PlayerAttributeTrendChart.vue'

const meta: Meta<typeof PlayerAttributeTrendChart> = {
  title: 'Charts/PlayerAttributeTrendChart',
  component: PlayerAttributeTrendChart,
}
export default meta

type Story = StoryObj<typeof PlayerAttributeTrendChart>

function makeSeason(
  seasonNum: number,
  power: number,
  contact: number,
  speed: number,
  fielding: number,
  arm: number,
  velocity = 0,
  junk = 0,
  accuracy = 0,
): main.PlayerAttributeSeasonDTO {
  const pct = (v: number) => (v > 0 ? Math.round(((v - 40) / 59) * 100) : undefined)
  return new main.PlayerAttributeSeasonDTO({
    seasonNum,
    seasonId: seasonNum,
    power,
    contact,
    speed,
    fielding,
    arm,
    velocity,
    junk,
    accuracy,
    powerPct: pct(power),
    contactPct: pct(contact),
    speedPct: pct(speed),
    fieldingPct: pct(fielding),
    armPct: pct(arm),
    velocityPct: velocity > 0 ? pct(velocity) : undefined,
    junkPct: junk > 0 ? pct(junk) : undefined,
    accuracyPct: accuracy > 0 ? pct(accuracy) : undefined,
    lgAvgPower: 65,
    lgAvgContact: 63,
    lgAvgSpeed: 62,
    lgAvgFielding: 64,
    lgAvgArm: 61,
    lgAvgVelocity: velocity > 0 ? 67 : 0,
    lgAvgJunk: junk > 0 ? 64 : 0,
    lgAvgAccuracy: accuracy > 0 ? 66 : 0,
  })
}

// A batter who peaks mid-career then declines
const batterSeasons = [
  makeSeason(1, 55, 62, 70, 68, 60),
  makeSeason(2, 65, 68, 72, 70, 63),
  makeSeason(3, 75, 74, 75, 73, 66),
  makeSeason(4, 82, 79, 74, 75, 68),
  makeSeason(5, 85, 82, 72, 77, 70),
  makeSeason(6, 80, 78, 68, 74, 67),
  makeSeason(7, 73, 72, 63, 71, 64),
]

// A pitcher whose pitch quality grows while batter attrs stay modest
const pitcherSeasons = [
  makeSeason(1, 45, 42, 40, 50, 48, 65, 58, 62),
  makeSeason(2, 46, 43, 41, 51, 48, 70, 63, 67),
  makeSeason(3, 47, 44, 42, 52, 49, 76, 68, 72),
  makeSeason(4, 48, 45, 43, 53, 50, 81, 73, 77),
  makeSeason(5, 48, 46, 43, 54, 50, 85, 78, 81),
]

// Edge case: only one season
const singleSeason = [makeSeason(1, 72, 68, 65, 70, 64)]

export const BatterRaw: Story = {
  args: { seasons: batterSeasons, isPitcher: false },
}

export const BatterPercentile: Story = {
  args: { seasons: batterSeasons, isPitcher: false },
  play: async ({ canvasElement }) => {
    // Click the Percentile button to start in percentile mode.
    const btn = canvasElement.querySelector<HTMLButtonElement>('.mode-btn:last-child')
    btn?.click()
  },
}

export const PitcherRaw: Story = {
  args: { seasons: pitcherSeasons, isPitcher: true },
}

export const SingleSeason: Story = {
  args: { seasons: singleSeason, isPitcher: false },
}

export const NoSeasons: Story = {
  args: { seasons: [], isPitcher: false },
}

export const AllVariants: Story = {
  render: () => ({
    components: { PlayerAttributeTrendChart },
    template: `
      <div style="background:#1e1e2e;padding:2rem;display:flex;flex-direction:column;gap:3rem">
        <div>
          <h3 style="color:#cdd6f4;font-size:0.875rem;margin:0 0 0.75rem">Batter — 7 seasons</h3>
          <PlayerAttributeTrendChart :seasons="batter" :is-pitcher="false" />
        </div>
        <div>
          <h3 style="color:#cdd6f4;font-size:0.875rem;margin:0 0 0.75rem">Pitcher — 5 seasons</h3>
          <PlayerAttributeTrendChart :seasons="pitcher" :is-pitcher="true" />
        </div>
        <div>
          <h3 style="color:#cdd6f4;font-size:0.875rem;margin:0 0 0.75rem">Single season</h3>
          <PlayerAttributeTrendChart :seasons="single" :is-pitcher="false" />
        </div>
      </div>
    `,
    setup() {
      return { batter: batterSeasons, pitcher: pitcherSeasons, single: singleSeason }
    },
  }),
}
