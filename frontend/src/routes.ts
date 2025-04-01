import {createMemoryHistory, createRouter, RouterOptions} from 'vue-router'

import HomeView from '@/pages/HomeView.vue'
import AboutView from '@/pages/AboutView.vue'

const routes: RouterOptions["routes"] = [
    { path: '/', component: HomeView },
    { path: '/about', component: AboutView },
]

const router = createRouter({
    history: createMemoryHistory(),
    routes,
})

export { router }
