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

## Storybook

Every non-trivial component in `src/components/` must have a `.stories.ts` file. See `AppLink.stories.ts` for the canonical structure: individual named exports per variant, plus an `AllVariants` story that shows them together in a realistic dark-background layout.
