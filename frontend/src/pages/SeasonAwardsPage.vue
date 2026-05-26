<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import MultiSelect from 'primevue/multiselect'
import { useToast } from 'primevue/usetoast'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  ComputeSeasonStatLeaderAwards,
  GetSeasonAwardCandidates,
  GetSeasonList,
  ListAllAwards,
  SubmitSeasonAwards,
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import AppLink from '../components/AppLink.vue'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import SeasonSelector from '../components/SeasonSelector.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'

const props = defineProps<{
  initialSeasonId?: number
}>()

// ── Local types ───────────────────────────────────────────────────────────────

// Wraps a DTO row with a mutable pendingAwardIds that the multiselect binds to.
interface BatterRow extends main.BattingCandidateDTO {
  pendingAwardIds: number[]
}
interface PitcherRow extends main.PitchingCandidateDTO {
  pendingAwardIds: number[]
}

// ── State ─────────────────────────────────────────────────────────────────────

const seasons = ref<main.SeasonSummaryDTO[]>([])
const selectedSeasonId = ref<number | null>(null)
const candidates = ref<main.SeasonAwardCandidatesDTO | null>(null)
const allAwards = ref<main.AwardDTO[]>([])

// Shared award state keyed by playerSeasonId. A player appearing in multiple sections
// (e.g. a top-10 rookie who also tops the overall list) intentionally shows the same
// staged awards in every section — this is the desired UX. Using reactive<Record>
// instead of ref<Map> gives Vue per-key tracking so only the affected row re-renders
// when a specific playerSeasonId's awards change, preventing cross-player contamination.
const pending = reactive<Record<number, number[]>>({})

const toast = useToast()

const loadingSeasons = ref(false)
const loadingCandidates = ref(false)
const submitting = ref(false)
const error = ref<string | null>(null)

const regularAwards = computed(() => allAwards.value.filter((a) => !a.isPlayoffAward && a.isUserAssignable))
const playoffAwards = computed(() => allAwards.value.filter((a) => a.isPlayoffAward && a.isUserAssignable))

const champRoundOnly = ref(false)

// Toggle switches between two server-side-sorted, server-side-limited arrays.
// No client-side filtering of a pre-truncated list.
const filteredPlayoffBatters = computed(() =>
  champRoundOnly.value ? (candidates.value?.championBatters ?? []) : (candidates.value?.playoffBatters ?? []),
)
const filteredPlayoffPitchers = computed(() =>
  champRoundOnly.value ? (candidates.value?.championPitchers ?? []) : (candidates.value?.playoffPitchers ?? []),
)

// ── Data loading ──────────────────────────────────────────────────────────────

async function loadSeasons() {
  loadingSeasons.value = true
  error.value = null
  try {
    const [list, awards] = await Promise.all([GetSeasonList(), ListAllAwards()])
    seasons.value = list ?? []
    allAwards.value = awards ?? []
    if (props.initialSeasonId && seasons.value.some((s) => s.id === props.initialSeasonId)) {
      selectedSeasonId.value = props.initialSeasonId
    } else if (seasons.value.length > 0) {
      selectedSeasonId.value = seasons.value[seasons.value.length - 1].id
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loadingSeasons.value = false
  }
}

async function loadCandidates(seasonId: number) {
  loadingCandidates.value = true
  champRoundOnly.value = false
  error.value = null
  try {
    const data = await GetSeasonAwardCandidates(seasonId)
    candidates.value = data

    // Clear previous state then seed from server data. Each playerSeasonId gets one
    // entry regardless of how many sections reference them.
    for (const k of Object.keys(pending)) delete (pending as Record<string, number[]>)[k]

    const reg = (rows: main.BattingCandidateDTO[] | main.PitchingCandidateDTO[]) => {
      for (const r of rows) {
        if (!(r.playerSeasonId in pending)) pending[r.playerSeasonId] = [...(r.awardIds ?? [])]
      }
    }
    reg(data.topBatters ?? [])
    reg(data.topPitchers ?? [])
    reg(data.topRookieBatters ?? [])
    reg(data.topRookiePitchers ?? [])
    for (const t of data.byTeam ?? []) {
      reg(t.batters ?? [])
      reg(t.pitchers ?? [])
    }
    for (const p of data.byPosition ?? []) reg(p.batters ?? [])
    reg(data.playoffBatters ?? [])
    reg(data.playoffPitchers ?? [])
    reg(data.championBatters ?? [])
    reg(data.championPitchers ?? [])
  } catch (e) {
    error.value = String(e)
  } finally {
    loadingCandidates.value = false
  }
}

// ── Actions ───────────────────────────────────────────────────────────────────

// Submit runs auto-computed stat leaders first (same as the legacy companion app),
// then persists all manual award selections in one transaction.
async function submitAwards() {
  if (selectedSeasonId.value == null) return
  submitting.value = true
  error.value = null
  try {
    await ComputeSeasonStatLeaderAwards(selectedSeasonId.value)

    const playerAwards: main.PlayerAwardEntryDTO[] = []
    for (const [psIdStr, ids] of Object.entries(pending) as [string, number[]][]) {
      playerAwards.push(new main.PlayerAwardEntryDTO({ playerSeasonId: Number(psIdStr), awardIds: ids }))
    }
    await SubmitSeasonAwards(
      new main.SubmitSeasonAwardsDTO({
        seasonId: selectedSeasonId.value,
        playerAwards,
      }),
    )
    await loadCandidates(selectedSeasonId.value)
    toast.add({ severity: 'success', summary: 'Awards saved', life: 3000 })
  } catch (e) {
    error.value = String(e)
  } finally {
    submitting.value = false
  }
}

// ── Row helpers ───────────────────────────────────────────────────────────────

function getRegularAwardIds(psId: number): number[] {
  const all = pending[psId] ?? []
  const ids = new Set(regularAwards.value.map((a) => a.id))
  return all.filter((id: number) => ids.has(id))
}

function setRegularAwardIds(psId: number, newIds: number[]) {
  const all = pending[psId] ?? []
  const regularIdSet = new Set(regularAwards.value.map((a) => a.id))
  const kept = all.filter((id: number) => !regularIdSet.has(id))
  pending[psId] = [...kept, ...newIds]
}

function getPlayoffAwardIds(psId: number): number[] {
  const all = pending[psId] ?? []
  const ids = new Set(playoffAwards.value.map((a) => a.id))
  return all.filter((id: number) => ids.has(id))
}

function setPlayoffAwardIds(psId: number, newIds: number[]) {
  const all = pending[psId] ?? []
  const playoffIdSet = new Set(playoffAwards.value.map((a) => a.id))
  const kept = all.filter((id: number) => !playoffIdSet.has(id))
  pending[psId] = [...kept, ...newIds]
}

// IP display: outs / 3 + remainder tenths (e.g. 97 outs → "32.1")
function formatIP(outs: number): string {
  if (!outs) return '0.0'
  return `${Math.floor(outs / 3)}.${outs % 3}`
}

function formatRate(v: number, digits: number): string {
  if (!v) return '—'
  return v.toFixed(digits).replace(/^0\./, '.')
}

watch(selectedSeasonId, (id) => {
  if (id != null) loadCandidates(id)
})

const { set } = useBreadcrumbs()
onMounted(() => set([{ label: 'Awards' }]))
onMounted(loadSeasons)
</script>

<template>
  <div class="awards-page">
    <!-- ── Header ────────────────────────────────────────────────────────────── -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">Awards</h1>
        <SeasonSelector v-model="selectedSeasonId" :seasons="seasons" />
      </div>
      <div class="header-actions">
        <button
          class="btn btn-primary"
          :disabled="submitting || selectedSeasonId == null"
          @click="submitAwards"
        >
          {{ submitting ? 'Saving…' : 'Submit Awards' }}
        </button>
      </div>
    </div>

    <div v-if="candidates?.autoSuggested" class="auto-suggest-banner">
      ✦ Awards pre-filled with suggestions — All-Star for top 2 per team, Silver Slugger for top
      batter per position. Edit and submit to confirm.
    </div>

    <div v-if="error" class="error-msg">{{ error }}</div>

    <LoadingSpinner v-if="loadingSeasons || loadingCandidates" />

    <EmptyState v-else-if="!candidates" message="No seasons synced yet." />

    <div v-else class="sections">

      <!-- ── Top Batting Overall ──────────────────────────────────────────── -->
      <section class="award-section">
        <h2 class="section-title">Top Batting Overall</h2>
        <div class="table-wrap">
          <table class="stats-table">
            <thead>
              <tr>
                <th class="col-player">Player</th>
                <th class="col-team">Team</th>
                <th class="col-pos">Pos</th>
                <th>AB</th><th>H</th><th>HR</th><th>RBI</th>
                <th>BB</th><th>R</th><th>SB</th>
                <th>BA</th><th>OBP</th><th>SLG</th><th>OPS</th>
                <th class="col-awards">Awards</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in candidates.topBatters" :key="r.playerSeasonId">
                <td class="col-player">
                  <AppLink :to="`/players/${r.playerId}`" class="player-link">
                    {{ r.lastName }}, {{ r.firstName }}
                  </AppLink>
                </td>
                <td class="col-team">{{ r.teamName }}</td>
                <td class="col-pos">{{ r.primaryPosition }}</td>
                <td>{{ r.atBats }}</td>
                <td>{{ r.hits }}</td>
                <td>{{ r.homeRuns }}</td>
                <td>{{ r.rbi }}</td>
                <td>{{ r.walks }}</td>
                <td>{{ r.runs }}</td>
                <td>{{ r.stolenBases }}</td>
                <td>{{ formatRate(r.ba, 3) }}</td>
                <td>{{ formatRate(r.obp, 3) }}</td>
                <td>{{ formatRate(r.slg, 3) }}</td>
                <td class="stat-highlight">{{ formatRate(r.ops, 3) }}</td>
                <td class="col-awards">
                  <MultiSelect
                    :model-value="getRegularAwardIds(r.playerSeasonId)"
                    :options="regularAwards"
                    option-label="name"
                    option-value="id"
                    placeholder="—"
                    class="award-select"
                    :max-selected-labels="2"
                    @update:model-value="(v: number[]) => setRegularAwardIds(r.playerSeasonId, v)"
                  />
                </td>
              </tr>
              <tr v-if="!candidates.topBatters?.length">
                <td colspan="15" class="empty-row">No batting data for this season.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ── Top Pitching Overall ─────────────────────────────────────────── -->
      <section class="award-section">
        <h2 class="section-title">Top Pitching Overall</h2>
        <div class="table-wrap">
          <table class="stats-table">
            <thead>
              <tr>
                <th class="col-player">Player</th>
                <th class="col-team">Team</th>
                <th class="col-pos">Role</th>
                <th>W</th><th>L</th><th>SV</th><th>IP</th>
                <th>K</th><th>BB</th><th>ERA</th>
                <th>WHIP</th><th>K/9</th><th>H/9</th>
                <th class="col-awards">Awards</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in candidates.topPitchers" :key="r.playerSeasonId">
                <td class="col-player">
                  <AppLink :to="`/players/${r.playerId}`" class="player-link">
                    {{ r.lastName }}, {{ r.firstName }}
                  </AppLink>
                </td>
                <td class="col-team">{{ r.teamName }}</td>
                <td class="col-pos">{{ r.pitcherRole }}</td>
                <td>{{ r.wins }}</td>
                <td>{{ r.losses }}</td>
                <td>{{ r.saves }}</td>
                <td>{{ formatIP(r.outsPitched) }}</td>
                <td>{{ r.strikeouts }}</td>
                <td>{{ r.walks }}</td>
                <td class="stat-highlight">{{ formatRate(r.era, 2) }}</td>
                <td>{{ formatRate(r.whip, 2) }}</td>
                <td>{{ formatRate(r.k9, 2) }}</td>
                <td>{{ formatRate(r.h9, 2) }}</td>
                <td class="col-awards">
                  <MultiSelect
                    :model-value="getRegularAwardIds(r.playerSeasonId)"
                    :options="regularAwards"
                    option-label="name"
                    option-value="id"
                    placeholder="—"
                    class="award-select"
                    :max-selected-labels="2"
                    @update:model-value="(v: number[]) => setRegularAwardIds(r.playerSeasonId, v)"
                  />
                </td>
              </tr>
              <tr v-if="!candidates.topPitchers?.length">
                <td colspan="14" class="empty-row">No pitching data for this season.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ── Top Rookie Batting ───────────────────────────────────────────── -->
      <section v-if="candidates.topRookieBatters?.length" class="award-section">
        <h2 class="section-title">Top Rookie Batting</h2>
        <div class="table-wrap">
          <table class="stats-table">
            <thead>
              <tr>
                <th class="col-player">Player</th>
                <th class="col-team">Team</th>
                <th class="col-pos">Pos</th>
                <th>AB</th><th>H</th><th>HR</th><th>RBI</th>
                <th>BB</th><th>R</th><th>SB</th>
                <th>BA</th><th>OBP</th><th>SLG</th><th>OPS</th>
                <th class="col-awards">Awards</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in candidates.topRookieBatters" :key="r.playerSeasonId">
                <td class="col-player">
                  <AppLink :to="`/players/${r.playerId}`" class="player-link">
                    {{ r.lastName }}, {{ r.firstName }}
                  </AppLink>
                </td>
                <td class="col-team">{{ r.teamName }}</td>
                <td class="col-pos">{{ r.primaryPosition }}</td>
                <td>{{ r.atBats }}</td><td>{{ r.hits }}</td><td>{{ r.homeRuns }}</td>
                <td>{{ r.rbi }}</td><td>{{ r.walks }}</td><td>{{ r.runs }}</td>
                <td>{{ r.stolenBases }}</td>
                <td>{{ formatRate(r.ba, 3) }}</td>
                <td>{{ formatRate(r.obp, 3) }}</td>
                <td>{{ formatRate(r.slg, 3) }}</td>
                <td class="stat-highlight">{{ formatRate(r.ops, 3) }}</td>
                <td class="col-awards">
                  <MultiSelect
                    :model-value="getRegularAwardIds(r.playerSeasonId)"
                    :options="regularAwards"
                    option-label="name"
                    option-value="id"
                    placeholder="—"
                    class="award-select"
                    :max-selected-labels="2"
                    @update:model-value="(v: number[]) => setRegularAwardIds(r.playerSeasonId, v)"
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ── Top Rookie Pitching ──────────────────────────────────────────── -->
      <section v-if="candidates.topRookiePitchers?.length" class="award-section">
        <h2 class="section-title">Top Rookie Pitching</h2>
        <div class="table-wrap">
          <table class="stats-table">
            <thead>
              <tr>
                <th class="col-player">Player</th>
                <th class="col-team">Team</th>
                <th class="col-pos">Role</th>
                <th>W</th><th>L</th><th>SV</th><th>IP</th>
                <th>K</th><th>BB</th><th>ERA</th>
                <th>WHIP</th><th>K/9</th><th>H/9</th>
                <th class="col-awards">Awards</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in candidates.topRookiePitchers" :key="r.playerSeasonId">
                <td class="col-player">
                  <AppLink :to="`/players/${r.playerId}`" class="player-link">
                    {{ r.lastName }}, {{ r.firstName }}
                  </AppLink>
                </td>
                <td class="col-team">{{ r.teamName }}</td>
                <td class="col-pos">{{ r.pitcherRole }}</td>
                <td>{{ r.wins }}</td><td>{{ r.losses }}</td><td>{{ r.saves }}</td>
                <td>{{ formatIP(r.outsPitched) }}</td>
                <td>{{ r.strikeouts }}</td><td>{{ r.walks }}</td>
                <td class="stat-highlight">{{ formatRate(r.era, 2) }}</td>
                <td>{{ formatRate(r.whip, 2) }}</td>
                <td>{{ formatRate(r.k9, 2) }}</td>
                <td>{{ formatRate(r.h9, 2) }}</td>
                <td class="col-awards">
                  <MultiSelect
                    :model-value="getRegularAwardIds(r.playerSeasonId)"
                    :options="regularAwards"
                    option-label="name"
                    option-value="id"
                    placeholder="—"
                    class="award-select"
                    :max-selected-labels="2"
                    @update:model-value="(v: number[]) => setRegularAwardIds(r.playerSeasonId, v)"
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ── By Team ──────────────────────────────────────────────────────── -->
      <section v-if="candidates.byTeam?.length" class="award-section">
        <h2 class="section-title">By Team</h2>
        <div class="team-tabs-wrapper">
          <div class="team-tabs-scroller">
            <div v-for="team in candidates.byTeam" :key="team.historyId" class="team-block">
              <h3 class="team-name-header">{{ team.teamName }}</h3>

              <!-- Team batters -->
              <div class="table-wrap">
                <table class="stats-table stats-table--compact">
                  <thead>
                    <tr>
                      <th class="col-player">Batter</th>
                      <th class="col-pos">Pos</th>
                      <th>AB</th><th>H</th><th>HR</th><th>RBI</th>
                      <th>BA</th><th>OPS</th>
                      <th class="col-awards">Awards</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="r in team.batters" :key="r.playerSeasonId">
                      <td class="col-player">
                        <AppLink :to="`/players/${r.playerId}`" class="player-link">
                          {{ r.lastName }}, {{ r.firstName }}
                        </AppLink>
                      </td>
                      <td class="col-pos">{{ r.primaryPosition }}</td>
                      <td>{{ r.atBats }}</td>
                      <td>{{ r.hits }}</td>
                      <td>{{ r.homeRuns }}</td>
                      <td>{{ r.rbi }}</td>
                      <td>{{ formatRate(r.ba, 3) }}</td>
                      <td class="stat-highlight">{{ formatRate(r.ops, 3) }}</td>
                      <td class="col-awards">
                        <MultiSelect
                          :model-value="getRegularAwardIds(r.playerSeasonId)"
                          :options="regularAwards"
                          option-label="name"
                          option-value="id"
                          placeholder="—"
                          class="award-select"
                          :max-selected-labels="2"
                          @update:model-value="(v: number[]) => setRegularAwardIds(r.playerSeasonId, v)"
                        />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- Team pitchers -->
              <div class="table-wrap" style="margin-top: 0.5rem">
                <table class="stats-table stats-table--compact">
                  <thead>
                    <tr>
                      <th class="col-player">Pitcher</th>
                      <th class="col-pos">Role</th>
                      <th>W</th><th>L</th><th>SV</th><th>IP</th>
                      <th>K</th><th>ERA</th>
                      <th class="col-awards">Awards</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="r in team.pitchers" :key="r.playerSeasonId">
                      <td class="col-player">
                        <AppLink :to="`/players/${r.playerId}`" class="player-link">
                          {{ r.lastName }}, {{ r.firstName }}
                        </AppLink>
                      </td>
                      <td class="col-pos">{{ r.pitcherRole }}</td>
                      <td>{{ r.wins }}</td>
                      <td>{{ r.losses }}</td>
                      <td>{{ r.saves }}</td>
                      <td>{{ formatIP(r.outsPitched) }}</td>
                      <td>{{ r.strikeouts }}</td>
                      <td class="stat-highlight">{{ formatRate(r.era, 2) }}</td>
                      <td class="col-awards">
                        <MultiSelect
                          :model-value="getRegularAwardIds(r.playerSeasonId)"
                          :options="regularAwards"
                          option-label="name"
                          option-value="id"
                          placeholder="—"
                          class="award-select"
                          :max-selected-labels="2"
                          @update:model-value="(v: number[]) => setRegularAwardIds(r.playerSeasonId, v)"
                        />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- ── By Position ──────────────────────────────────────────────────── -->
      <section v-if="candidates.byPosition?.length" class="award-section">
        <h2 class="section-title">By Position</h2>
        <div class="position-grid">
          <div v-for="pos in candidates.byPosition" :key="pos.position" class="position-block">
            <h3 class="position-header">{{ pos.position }}</h3>
            <div class="table-wrap">
              <table class="stats-table stats-table--compact">
                <thead>
                  <tr>
                    <th class="col-player">Player</th>
                    <th class="col-team">Team</th>
                    <th>AB</th><th>H</th><th>HR</th><th>RBI</th>
                    <th>BA</th><th>OPS</th>
                    <th class="col-awards">Awards</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="r in pos.batters" :key="r.playerSeasonId">
                    <td class="col-player">
                      <AppLink :to="`/players/${r.playerId}`" class="player-link">
                        {{ r.lastName }}, {{ r.firstName }}
                      </AppLink>
                    </td>
                    <td class="col-team">{{ r.teamName }}</td>
                    <td>{{ r.atBats }}</td>
                    <td>{{ r.hits }}</td>
                    <td>{{ r.homeRuns }}</td>
                    <td>{{ r.rbi }}</td>
                    <td>{{ formatRate(r.ba, 3) }}</td>
                    <td class="stat-highlight">{{ formatRate(r.ops, 3) }}</td>
                    <td class="col-awards">
                      <MultiSelect
                        :model-value="getRegularAwardIds(r.playerSeasonId)"
                        :options="regularAwards"
                        option-label="name"
                        option-value="id"
                        placeholder="—"
                        class="award-select"
                        :max-selected-labels="2"
                        @update:model-value="(v: number[]) => setRegularAwardIds(r.playerSeasonId, v)"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      <!-- ── Playoff Awards ───────────────────────────────────────────────── -->
      <section
        v-if="playoffAwards.length > 0 && (candidates.playoffBatters?.length || candidates.playoffPitchers?.length)"
        class="award-section"
      >
        <div class="section-title-row">
          <h2 class="section-title">Playoff Awards</h2>
          <label class="champ-toggle">
            <input v-model="champRoundOnly" type="checkbox" />
            Champions only
          </label>
        </div>

        <!-- Playoff batters -->
        <DataTable
          v-if="filteredPlayoffBatters.length"
          :value="filteredPlayoffBatters"
          size="small"
        >
          <Column header="Player" style="min-width: 160px">
            <template #body="{ data }">
              <span v-if="data.isChampionTeam" aria-hidden="true">🏆 </span>
              <AppLink :to="`/players/${data.playerId}`" class="player-link">
                {{ data.lastName }}, {{ data.firstName }}
              </AppLink>
            </template>
          </Column>
          <Column field="teamName" header="Team" style="min-width: 90px" />
          <Column field="primaryPosition" header="Pos" style="min-width: 50px" />
          <Column field="atBats" header="AB" />
          <Column field="hits" header="H" />
          <Column field="homeRuns" header="HR" />
          <Column field="rbi" header="RBI" />
          <Column field="walks" header="BB" />
          <Column field="runs" header="R" />
          <Column field="stolenBases" header="SB" />
          <Column header="BA">
            <template #body="{ data }">{{ formatRate(data.ba, 3) }}</template>
          </Column>
          <Column header="OBP">
            <template #body="{ data }">{{ formatRate(data.obp, 3) }}</template>
          </Column>
          <Column header="SLG">
            <template #body="{ data }">{{ formatRate(data.slg, 3) }}</template>
          </Column>
          <Column header="OPS">
            <template #body="{ data }"><strong>{{ formatRate(data.ops, 3) }}</strong></template>
          </Column>
          <Column header="Playoff Award" style="min-width: 180px">
            <template #body="{ data }">
              <MultiSelect
                :model-value="getPlayoffAwardIds(data.playerSeasonId)"
                :options="playoffAwards"
                option-label="name"
                option-value="id"
                placeholder="—"
                class="award-select"
                :max-selected-labels="2"
                @update:model-value="(v: number[]) => setPlayoffAwardIds(data.playerSeasonId, v)"
              />
            </template>
          </Column>
        </DataTable>

        <!-- Playoff pitchers -->
        <DataTable
          v-if="filteredPlayoffPitchers.length"
          :value="filteredPlayoffPitchers"
          size="small"
          style="margin-top: 0.75rem"
        >
          <Column header="Pitcher" style="min-width: 160px">
            <template #body="{ data }">
              <span v-if="data.isChampionTeam" aria-hidden="true">🏆 </span>
              <AppLink :to="`/players/${data.playerId}`" class="player-link">
                {{ data.lastName }}, {{ data.firstName }}
              </AppLink>
            </template>
          </Column>
          <Column field="teamName" header="Team" style="min-width: 90px" />
          <Column field="pitcherRole" header="Role" style="min-width: 50px" />
          <Column field="wins" header="W" />
          <Column field="losses" header="L" />
          <Column field="saves" header="SV" />
          <Column header="IP">
            <template #body="{ data }">{{ formatIP(data.outsPitched) }}</template>
          </Column>
          <Column field="strikeouts" header="K" />
          <Column field="walks" header="BB" />
          <Column header="ERA">
            <template #body="{ data }"><strong>{{ formatRate(data.era, 2) }}</strong></template>
          </Column>
          <Column header="WHIP">
            <template #body="{ data }">{{ formatRate(data.whip, 2) }}</template>
          </Column>
          <Column header="Playoff Award" style="min-width: 180px">
            <template #body="{ data }">
              <MultiSelect
                :model-value="getPlayoffAwardIds(data.playerSeasonId)"
                :options="playoffAwards"
                option-label="name"
                option-value="id"
                placeholder="—"
                class="award-select"
                :max-selected-labels="2"
                @update:model-value="(v: number[]) => setPlayoffAwardIds(data.playerSeasonId, v)"
              />
            </template>
          </Column>
        </DataTable>

        <p v-if="!filteredPlayoffBatters.length && !filteredPlayoffPitchers.length" class="section-hint">
          No players found for the selected filter.
        </p>
      </section>

    </div>

    <!-- Sticky submit bar at bottom -->
    <div class="submit-bar">
      <span v-if="candidates?.autoSuggested" class="suggest-note">
        Auto-suggested awards loaded — review and submit.
      </span>
      <button
        class="btn btn-primary btn-lg"
        :disabled="submitting || selectedSeasonId == null"
        @click="submitAwards"
      >
        {{ submitting ? 'Saving…' : 'Submit Awards' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.awards-page {
  padding: 1.5rem 2rem 5rem;
  max-width: 1400px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
  gap: 1rem;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
  white-space: nowrap;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.btn {
  padding: 0.45rem 1rem;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  white-space: nowrap;
}

.btn-primary {
  background: var(--color-accent, #4c9aff);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn-ghost {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

.btn-ghost:hover:not(:disabled) {
  background: var(--color-surface-3);
}

.btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.btn-lg {
  padding: 0.6rem 1.5rem;
  font-size: 0.9375rem;
}

.auto-suggest-banner {
  background: #1a2233;
  border: 1px solid #2d4a7a;
  color: #7aadff;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.8125rem;
  margin-bottom: 1rem;
}

.error-msg {
  color: var(--color-error, #f87171);
  font-size: 0.875rem;
  margin-bottom: 0.75rem;
}

/* ── Sections ──────────────────────────────────────────────────────────────── */

.sections {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.award-section {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0.375rem;
  border-bottom: 1px solid var(--color-border);
}

.section-title {
  font-size: 1rem;
  font-weight: 600;
  margin: 0;
}

.champ-toggle {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  cursor: pointer;
  user-select: none;
}

.champ-toggle input {
  cursor: pointer;
}


.section-hint {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  margin: 0;
}

/* ── Tables ────────────────────────────────────────────────────────────────── */

.table-wrap {
  overflow-x: auto;
}

.stats-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.stats-table th,
.stats-table td {
  padding: 0.35rem 0.5rem;
  text-align: right;
  white-space: nowrap;
  border-bottom: 1px solid var(--color-border);
}

.stats-table th {
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  background: var(--color-surface-1);
}

.stats-table--compact th,
.stats-table--compact td {
  padding: 0.25rem 0.4rem;
}

.col-player {
  text-align: left;
  min-width: 140px;
}

.col-team {
  text-align: left;
  min-width: 90px;
}

.col-pos {
  text-align: left;
  min-width: 50px;
}

.col-awards {
  text-align: left;
  min-width: 180px;
}

.player-link {
  font-weight: 500;
}

.stat-highlight {
  font-weight: 600;
  color: var(--color-text-primary);
}

.empty-row {
  text-align: center;
  color: var(--color-text-secondary);
  padding: 1rem;
}

/* ── Award selector ────────────────────────────────────────────────────────── */

.award-select {
  min-width: 160px;
  font-size: 0.8rem;
}

/* ── By-team layout ────────────────────────────────────────────────────────── */

.team-tabs-wrapper {
  overflow-x: auto;
}

.team-tabs-scroller {
  display: flex;
  gap: 2rem;
  min-width: max-content;
}

.team-block {
  min-width: 580px;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.team-name-header {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin: 0;
  padding-bottom: 0.25rem;
  border-bottom: 1px solid var(--color-border);
}

/* ── By-position layout ────────────────────────────────────────────────────── */

.position-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(480px, 1fr));
  gap: 1.5rem;
}

.position-block {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.position-header {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin: 0;
}

/* ── Sticky submit bar ─────────────────────────────────────────────────────── */

.submit-bar {
  position: fixed;
  bottom: 0;
  left: 220px; /* sidebar width */
  right: 0;
  background: var(--color-surface-1);
  border-top: 1px solid var(--color-border);
  padding: 0.75rem 2rem;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 1rem;
  z-index: 10;
}

.suggest-note {
  font-size: 0.8125rem;
  color: #7aadff;
}
</style>
