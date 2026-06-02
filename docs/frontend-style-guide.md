# Frontend Style Guide

Standards for UI patterns that appear throughout smb-tools. When adding new UI, check here first before hand-rolling a pattern that already has a canonical implementation.

---

## Links

### Use `AppLink` for every link in the app

`AppLink` (`src/components/AppLink.vue`) is the single component for all clickable navigation — both internal routes and external URLs. Do not use raw `<RouterLink>`, bare `<a>` tags, or `router.push()` as a stand-in for a link.

```vue
<!-- Internal route -->
<AppLink :to="`/players/${r.playerId}`">{{ r.firstName }} {{ r.lastName }}</AppLink>

<!-- External URL (opens in new tab automatically) -->
<AppLink href="https://www.baseball-reference.com">Baseball Reference</AppLink>
```

All links share one appearance: accent color, no underline at rest, underlines on hover. There is no alternate variant.

### Props

| Prop | Type | Description |
|------|------|-------------|
| `to` | `RouteLocationRaw` | Internal route. Renders a `RouterLink`. |
| `href` | `string` | External URL. Renders an `<a target="_blank" rel="noopener noreferrer">`. |

`to` and `href` are mutually exclusive. If neither is provided the component renders a `<span>` — useful as a conditional placeholder.

### In DataTable columns

Wrap the slot content with `AppLink` instead of a local link class:

```vue
<Column header="Player" sort-field="lastName" sortable style="min-width: 160px">
  <template #body="{ data: r }">
    <AppLink :to="`/players/${r.playerId}`">{{ r.firstName }} {{ r.lastName }}</AppLink>
    <span v-if="r.isHallOfFamer" class="hof-badge">HoF</span>
  </template>
</Column>
```

### What NOT to do

```vue
<!-- Bad: raw RouterLink with a one-off CSS class -->
<RouterLink :to="`/players/${r.playerId}`" class="player-link">…</RouterLink>

<!-- Bad: anchor tag for internal navigation -->
<a href="/players/1">…</a>

<!-- Bad: router.push() as a link substitute -->
<button @click="router.push(`/players/${r.playerId}`)">…</button>

<!-- Bad: dynamic component trick -->
<component :is="to ? RouterLink : 'span'" :to="to">…</component>
```

Each of these scatters link styling and makes global link changes require hunting through every component file.

---

## Page Layout

There are two layout modes. Every page must use exactly one of them — choosing the wrong one causes either cramped grids or awkwardly wide text.

### Medium-width pages (default)

Use for pages that are primarily text, forms, or small tables — content that reads poorly when stretched across the full viewport (e.g., Dashboard, Awards, Hall of Fame).

**Step 1** — Leave the route in `router.ts` without a `meta` key (or omit `fullWidth`):

```ts
// router.ts
{ path: '/awards', component: () => import('./pages/SeasonAwardsPage.vue') }
```

**Step 2** — Inside the page component, apply `padding: 2rem` uniformly to the root element. The shell already constrains it to `max-width: 1280px` via `.page-view`:

```css
.my-page {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}
```

No inner wrappers or max-width constraints are needed — the shell handles it.

---

### Full-width pages

Use for pages that contain data-heavy grids or tables that benefit from the extra horizontal room (e.g., Team, Player, Leaderboards, Teams index).

**Step 1** — Add `meta: { fullWidth: true }` to the route in `router.ts`. This removes the `max-width: 1280px` cap from the shell's `.page-view`:

```ts
// router.ts
{
  path: '/teams/:teamId',
  component: () => import('./pages/TeamPage.vue'),
  props: (route) => ({ teamId: Number(route.params.teamId) }),
  meta: { fullWidth: true },
},
```

**Step 2** — Inside the page component, use a two-zone layout:

- **Header zone** — a wrapper with `padding: 2rem 2rem 0` and `max-width: 1000px`. Keeps the title and summary stats from stretching uncomfortably wide on large monitors.
- **Grid zones** — sections with `padding: 0 2rem` only, so the DataTable fills the full available width.

```vue
<template>
  <div class="team-page">
    <div class="team-content">
      <header class="page-header">…</header>
    </div>

    <section class="section">
      <h3>Season History</h3>
      <DataTable …>…</DataTable>
    </section>
  </div>
</template>

<style scoped>
.team-page {
  padding-bottom: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

/* Constrained header — prevents title/stats from spanning the full monitor width */
.team-content {
  padding: 2rem 2rem 0;
  max-width: 1000px;
}

/* Grid sections — full-width with breathing room from the edges */
.section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0 2rem;
}
</style>
```

> Both steps are required. Adding `meta: { fullWidth: true }` without the two-zone CSS produces content that spans edge-to-edge with no padding. Adding the two-zone CSS without the meta flag produces a layout constrained to 1280px that defeats the purpose of the split.

---

## DataTable Column Headers

### Tooltips on acronym headers

Any column header that uses an acronym, abbreviation, or non-obvious shorthand must include a `title` tooltip spelling out the full name. Use the `#header` slot — when this slot is present, omit the `header` prop entirely:

```vue
<Column field="runsFor" sortable style="width: 55px">
  <template #header><span title="Runs Scored">R</span></template>
</Column>
```

When a column also has a `#body` slot, both coexist on the same `<Column>`:

```vue
<Column field="winPct" sortable style="width: 68px">
  <template #header><span title="Win Percentage">PCT</span></template>
  <template #body="{ data }">{{ fmtPct(data.winPct) }}</template>
</Column>
```

**What needs a tooltip**: any header text that requires domain knowledge to parse — single letters (W, L, R), composite abbreviations (RA, PCT, PW, PL), or sport-specific shorthand (OPS, ERA, FIP, smbWAR). Fully spelled-out words ("Season", "Team", "Player") do not need tooltips.

---

## Traits

### Use `TraitList` to display player traits

`TraitList` (`src/components/TraitList.vue`) is the single component for rendering a player's trait list. It handles all trait-specific formatting: positive traits are blue (`#4a9eff`), negative traits are red (`var(--color-error)`), and empty trait lists render as a muted em dash. Never hand-roll trait coloring inline.

```vue
<!-- In a DataTable column body -->
<TraitList :traits="r.traits" />

<!-- When the source may be undefined (e.g., career rows) -->
<TraitList :traits="r.traits ?? []" />
```

The component accepts exactly one prop:

| Prop | Type | Description |
|------|------|-------------|
| `traits` | `string[]` | Trait names to display. Pass `[]` for players with no traits — renders `—`. |

The negative trait set is defined inside `TraitList.vue` and covers both SMB4 names and legacy SMB3 names present in migrated franchise data. Do not duplicate this set elsewhere.

### What NOT to do

```vue
<!-- Bad: inline coloring scattered across components -->
<span :class="isNegative(t) ? 'text-red' : 'text-blue'">{{ trait }}</span>

<!-- Bad: plain join with no coloring -->
{{ r.traits.join(', ') }}
```

---

## Modals

### Use PrimeVue `Dialog` for forms and multi-step flows

There are two modal patterns:

- **`useConfirm()` + `<ConfirmDialog>`** — for simple, one-click destructive confirmations ("Are you sure you want to delete this franchise?"). The global confirm service drives it; no custom template needed.
- **`<Dialog>`** — for anything with a form, tab structure, or interactive content. `TeamLogoManager` is the canonical example.

```vue
<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'

const visible = defineModel<boolean>('visible', { required: true })
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Dialog Title" :style="{ width: '560px' }">
    <!-- content -->
    <p>Dialog body goes here.</p>

    <template #footer>
      <Button label="Cancel" severity="secondary" text @click="visible = false" />
      <Button label="Confirm" @click="handleConfirm" />
    </template>
  </Dialog>
</template>
```

**Rules:**
- Always set `modal` so the backdrop is shown.
- Set an explicit `width` in `:style` — dialogs without a width stretch unpredictably on wide monitors.
- Footer pattern: **Cancel on the left** (secondary, text), **primary action on the right**. For destructive primary actions use `severity="danger"`.
- Do not put loading spinners in the dialog header — show them inside the content area.
- Do not `$emit('close')` on every interaction — only emit when the user explicitly closes (Cancel or a final success action). Let the parent control visibility via `v-model:visible`.

### What NOT to do

```vue
<!-- Bad: raw <div> overlay -->
<div v-if="open" class="overlay">…</div>

<!-- Bad: confirm-service for a form -->
useConfirm().require({ message: '<form content here>' })

<!-- Bad: no width constraint -->
<Dialog v-model:visible="visible" modal header="Upload Logo">
```

---

## Storybook

Every non-trivial component in `src/components/` must have a `.stories.ts` file. See `AppLink.stories.ts` for the canonical structure: individual named exports per variant, plus an `AllVariants` story that shows them together in a realistic dark-background layout.
