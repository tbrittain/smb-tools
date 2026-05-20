import { createRouter, createWebHashHistory } from 'vue-router'
import DashboardPage from './pages/DashboardPage.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', component: DashboardPage },
    {
      path: '/teams',
      component: () => import('./pages/TeamsPage.vue'),
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
      path: '/search',
      component: () => import('./pages/SearchPage.vue'),
      props: (route) => ({ q: String(route.query.q ?? '') }),
    },
  ],
})

export default router
