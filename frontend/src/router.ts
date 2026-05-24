import { createRouter, createWebHashHistory } from 'vue-router'
import DashboardPage from './pages/DashboardPage.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', component: DashboardPage },
    {
      path: '/teams',
      component: () => import('./pages/TeamsPage.vue'),
      meta: { fullWidth: true },
    },
    {
      path: '/teams/:teamId',
      component: () => import('./pages/TeamPage.vue'),
      props: (route) => ({ teamId: Number(route.params.teamId) }),
    },
    {
      path: '/teams/:teamId/seasons/:historyId',
      component: () => import('./pages/TeamSeasonPage.vue'),
      props: (route) => ({
        teamId: Number(route.params.teamId),
        historyId: Number(route.params.historyId),
      }),
    },
    {
      path: '/players/:playerId',
      component: () => import('./pages/PlayerPage.vue'),
      props: (route) => ({ playerId: Number(route.params.playerId) }),
    },
    {
      path: '/leaderboards',
      component: () => import('./pages/LeaderboardsPage.vue'),
      meta: { fullWidth: true },
    },
    {
      path: '/awards',
      component: () => import('./pages/SeasonAwardsPage.vue'),
      props: (route) => ({
        initialSeasonId: route.query.seasonId ? Number(route.query.seasonId) : undefined,
      }),
    },
    {
      path: '/hall-of-fame',
      component: () => import('./pages/HallOfFamePage.vue'),
    },
    {
      path: '/setup',
      component: () => import('./pages/SetupPage.vue'),
    },
    {
      path: '/migrate-legacy',
      component: () => import('./pages/LegacyMigrationPage.vue'),
    },
  ],
})

export default router
