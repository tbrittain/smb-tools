# Frontend — Coding Standards and UI Patterns

Standards for `frontend/`. Covers Vue/TypeScript conventions and canonical UI patterns. Read this before working on any frontend code.

---

## Coding Standards

**`<script setup>` on every component.** No Options API.

**TypeScript strict mode is non-negotiable.** No `any` types — Biome enforces this with `noExplicitAny: "error"`. `as unknown as T` is equally forbidden — it defeats the type system the same way `any` does. Fix the source type instead (use class constructors, fix prop types, add proper type guards).

**Formatting is enforced by Biome**: 2-space indent, single quotes, no semicolons, 120-char line width. Run `npm run lint:fix` before committing. Do not manually reformat.

**Component structure order within `<script setup>`**:
1. Imports
2. Props / emits definitions
3. Injected dependencies (inject, useRoute, useRouter, stores)
4. Reactive state (ref, reactive)
5. Computed properties
6. Watchers
7. Lifecycle hooks
8. Functions / event handlers

**Pinia stores** are for app-wide state (current franchise, app loading state). Component-local state stays in `ref()`/`reactive()`. Don't reach for Pinia for state that doesn't need to be shared.

**Composables** for logic reused across more than one component. Composables live in `src/composables/`.

**No inline styles.** Use scoped CSS in the component's `<style scoped>` block or PrimeVue design tokens.

**PrimeVue components first.** Before hand-rolling any UI primitive — tables, dialogs, dropdowns, paginators, tabs, checkboxes — check whether PrimeVue 4 already provides it. Use `DataTable` + `Column` for any tabular data (never `<table>`/`<thead>`/`<tbody>` by hand), `MultiSelect` for multi-value pickers, `TabView`/`TabPanel` for tabs, etc. Custom HTML primitives are only acceptable when no PrimeVue component fits and the element is too small to warrant a library import.

**Server-side pagination and filtering — always.** smb-tools can accumulate hundreds of seasons and thousands of player-seasons. Pagination and filtering must be implemented in the Go store layer (SQL `LIMIT`/`OFFSET`, `WHERE`, `ORDER BY`), never in the Vue layer by slicing or filtering a full array that was already fetched. Client-side filtering of a server-truncated result set is always wrong — if the backend returns the top 10 rows by OPS and the frontend filters by team, it silently misses players ranked 11+. The only exception: instant UI feedback for a user-typed search that debounces to a real server call; lightweight ephemeral UI state (tab selection, column sort on a fully-loaded small dataset) is acceptable.

**DataTable column widths — always `min-width`, never `width`.** PrimeVue `<Column>` elements must use `style="min-width: Xpx"` so columns can flex and scale. Never `style="width: Xpx"` — static widths break the layout on wider screens.

**Wails bindings** are imported from `../../wailsjs/go/main/App` and called as async functions. Always handle errors explicitly — Wails surfaces Go errors as rejected promises.

## Testing

- **Vitest** for unit tests on composables, utility functions, and stat calculation logic ported to the frontend.
- **Storybook** for component-level development and visual regression. Every non-trivial component in `src/components/` must have a `.stories.ts` file covering: default state, empty/zero state, loading state, and any notable variants.
- **Storybook is not optional for components.** Build the Story alongside the component — this is how component invariants are validated without a live database. See `AppLink.stories.ts` for the canonical structure.
- Vue components receive data via props, not by directly calling Wails bindings — this is what makes component testing possible without Wails.

---

## UI Patterns

Standards for UI patterns that appear throughout smb-tools. Check here first before hand-rolling a pattern that already has a canonical implementation.

### Links

#### Use `AppLink` for every link in the app

`AppLink` (`src/components/AppLink.vue`) is the single component for all clickable navigation — both internal routes and external URLs. Do not use raw `<RouterLink>`, bare `<a>` tags, or `router.push()` as a stand-in for a link.

```vue
<!-- Internal route -->
<AppLink :to="`/players/${r.playerId}`">{{ r.firstName }} {{ r.lastName }}</AppLink>

<!-- External URL (opens in new tab automatically) -->
<AppLink href="https://www.baseball-reference.com">Baseball Reference</AppLink>
```

All links share one appearance: accent color, no underline at rest, underlines on hover. There is no alternate variant.

**Props**

| Prop | Type | Description |
|------|------|-------------|
| `to` | `RouteLocationRaw` | Internal route. Renders a `RouterLink`. |
| `href` | `string` | External URL. Renders an `<a target="_blank" rel="noopener noreferrer">`. |

`to` and `href` are mutually exclusive. If neither is provided the component renders a `<span>` — useful as a conditional placeholder.

**In DataTable columns** — wrap the slot content with `AppLink` instead of a local link class:

```vue
<Column header="Player" sort-field="lastName" sortable style="min-width: 160px">
  <template #body="{ data: r }">
    <AppLink :to="`/players/${r.playerId}`">{{ r.firstName }} {{ r.lastName }}</AppLink>
    <span v-if="r.isHallOfFamer" class="hof-badge">HoF</span>
  </template>
</Column>
```

**What NOT to do**

```vue
<!-- Bad: raw RouterLink with a one-off CSS class -->
<RouterLink :to="`/players/${r.playerId}`" class="player-link">…</RouterLink>

<!-- Bad: anchor tag for internal navigation -->
<a href="/players/1">…</a>

<!-- Bad: router.push() as a link substitute -->
<button @click="router.push(`/players/${r.playerId}`)">…</button>
```

---

### Icons

smb-tools uses **PrimeIcons** (`primeicons` package). The CSS is imported globally in `src/main.ts` — no per-component import needed.

Icons are referenced by CSS class name (`pi pi-{name}`) and are most commonly passed to PrimeVue Button's `icon` prop:

```vue
<Button icon="pi pi-trash" severity="danger" outlined size="small" />
<Button label="Manage Logos" icon="pi pi-image" severity="secondary" outlined size="small" />
```

They can also be used as standalone elements:

```vue
<i class="pi pi-check" />
```

Browse the full icon set at [primevue.org/icons](https://primevue.org/icons). PrimeIcons ships as a webfont — all icons are bundled regardless of which ones are used, so there is no bundle-size benefit to limiting icon usage.

---

### Page Layout

There are two layout modes. Every page must use exactly one of them — choosing the wrong one causes either cramped grids or awkwardly wide text.

#### Medium-width pages (default)

Use for pages that are primarily text, forms, or small tables — content that reads poorly when stretched across the full viewport (e.g., Dashboard, Awards, Hall of Fame).

**Step 1** — Leave the route in `router.ts` without a `meta` key (or omit `fullWidth`):

```ts
// router.ts
{ path: '/awards', component: () => import('./pages/SeasonAwardsPage.vue') }
```

**Step 2** — Inside the page component, apply `padding: 2rem` to the root element. The shell constrains it to `max-width: 1280px` via `.page-view`:

```css
.my-page {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}
```

No inner wrappers or max-width constraints needed — the shell handles it.

#### Full-width pages

Use for pages with data-heavy grids or tables that benefit from the extra horizontal room (e.g., Team, Player, Leaderboards, Teams index).

**Step 1** — Add `meta: { fullWidth: true }` to the route in `router.ts`:

```ts
{
  path: '/teams/:teamId',
  component: () => import('./pages/TeamPage.vue'),
  props: (route) => ({ teamId: Number(route.params.teamId) }),
  meta: { fullWidth: true },
},
```

**Step 2** — Inside the page component, use a two-zone layout:

- **Header zone** — a wrapper with `padding: 2rem 2rem 0` and `max-width: 1000px`. Keeps the title and summary stats from stretching uncomfortably wide.
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

.team-content {
  padding: 2rem 2rem 0;
  max-width: 1000px;
}

.section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0 2rem;
}
</style>
```

> Both steps are required. Adding `meta: { fullWidth: true }` without the two-zone CSS produces content that spans edge-to-edge with no padding. Adding the two-zone CSS without the meta flag produces a layout still capped at 1280px.

#### New page checklist

Every new page in `src/pages/` requires all three of these:

1. **Route** — add an entry in `router.ts` (with `meta: { fullWidth: true }` if applicable).
2. **Sidebar link** — add a `<router-link>` in the `<nav class="sidebar-nav">` block in `App.vue` (top-level pages only).
3. **Breadcrumb** — call `useBreadcrumbs().set(...)` in `onMounted` so the topbar label updates on navigation.

**Top-level pages** (reachable directly from the sidebar) must also be added to `ROOT_PATHS` in `src/composables/useBreadcrumbs.ts`. Paths in that set reset the breadcrumb trail instead of pushing onto it, so navigating sidebar → Export doesn't carry over crumbs from wherever the user was before.

```ts
// useBreadcrumbs.ts
const ROOT_PATHS = new Set(['/', '/teams', '/leaderboards', ..., '/my-page'])
```

```vue
<!-- MyPage.vue -->
<script lang="ts" setup>
import { onMounted } from 'vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'

const { set } = useBreadcrumbs()
onMounted(() => set([{ label: 'My Page' }]))
</script>
```

**Detail pages** (navigated to from a list, not the sidebar) do not go in `ROOT_PATHS`. They push onto the trail, so the parent page appears as a clickable crumb:

```ts
// e.g. PlayerPage.vue — the Teams crumb is injected automatically from the trail
onMounted(() => set([{ label: playerName }]))
```

Omitting the breadcrumb call leaves the topbar blank when the user lands on the page.

---

### DataTable Column Headers

#### Tooltips on acronym headers

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

**What needs a tooltip**: any header text requiring domain knowledge to parse — single letters (W, L, R), composite abbreviations (RA, PCT), or sport-specific shorthand (OPS, ERA, FIP, smbWAR). Fully spelled-out words ("Season", "Team", "Player") do not.

---

### Traits

#### Use `TraitList` to display player traits

`TraitList` (`src/components/TraitList.vue`) is the single component for rendering a player's trait list. It handles all trait-specific formatting: positive traits are blue (`#4a9eff`), negative traits are red (`var(--color-error)`), and empty trait lists render as a muted em dash. Never hand-roll trait coloring inline.

```vue
<!-- In a DataTable column body -->
<TraitList :traits="r.traits" />

<!-- When the source may be undefined (e.g., career rows) -->
<TraitList :traits="r.traits ?? []" />
```

| Prop | Type | Description |
|------|------|-------------|
| `traits` | `string[]` | Trait names to display. Pass `[]` for players with no traits — renders `—`. |

The negative trait set is defined inside `TraitList.vue` and covers both SMB4 names and legacy SMB3 names present in migrated franchise data. Do not duplicate this set elsewhere.

**What NOT to do**

```vue
<!-- Bad: inline coloring scattered across components -->
<span :class="isNegative(t) ? 'text-red' : 'text-blue'">{{ trait }}</span>

<!-- Bad: plain join with no coloring -->
{{ r.traits.join(', ') }}
```

---

### Modals

There are two modal patterns:

- **`useConfirm()` + `<ConfirmDialog>`** — for simple one-click destructive confirmations. The global confirm service drives it; no custom template needed.
- **`<Dialog>`** — for anything with a form, tab structure, or interactive content. `TeamLogoManager` is the canonical example.

```vue
<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'

const visible = defineModel<boolean>('visible', { required: true })
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Dialog Title" :style="{ width: '560px' }">
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
- Footer pattern: **Cancel on the left** (secondary, text), **primary action on the right**. Destructive primary actions use `severity="danger"`.
- Do not put loading spinners in the dialog header — show them inside the content area.
- Only `$emit('close')` when the user explicitly closes (Cancel or a final success action). Let the parent control visibility via `v-model:visible`.

**What NOT to do**

```vue
<!-- Bad: raw <div> overlay -->
<div v-if="open" class="overlay">…</div>

<!-- Bad: confirm-service for a form -->
useConfirm().require({ message: '<form content here>' })

<!-- Bad: no width constraint -->
<Dialog v-model:visible="visible" modal header="Upload Logo">
```

---

### Toast Notifications

Use PrimeVue's toast service for non-blocking feedback after user actions — saves, deletes, uploads, and other operations where the user needs confirmation something happened.

The `<Toast />` component is mounted once in `App.vue` and listens globally. Components only need the composable:

```vue
<script setup lang="ts">
import { useToast } from 'primevue/usetoast'

const toast = useToast()

async function save() {
  await doSave()
  toast.add({ severity: 'success', summary: 'Changes saved', life: 3000 })
}
</script>
```

**Severity levels in use:**

| Severity | When to use |
|----------|-------------|
| `'success'` | Confirming a completed action (save, delete, upload, assign) |
| `'error'` | Surfacing a failure that inline error text can't reach |

**Rules:**
- Always set `life` (ms). `3000` for routine confirmations; `4000–5000` for deletions or less reversible actions.
- Keep `summary` short — one clause, no period. Omit `detail` unless extra context is genuinely needed.
- Do not show a success toast and an inline error message simultaneously for the same operation.
- Do not show a toast for read operations or navigation.

**What NOT to do**

```vue
<!-- Bad: mounting a second <Toast> inside a component -->
<Toast />

<!-- Bad: no life — toast stays on screen forever -->
toast.add({ severity: 'success', summary: 'Saved' })
```

---

### Storybook

Every non-trivial component in `src/components/` must have a `.stories.ts` file. See `AppLink.stories.ts` for the canonical structure: individual named exports per variant, plus an `AllVariants` story that shows them together in a realistic dark-background layout.
