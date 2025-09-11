import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import ChangeLogView from '../views/ChangeLogView.vue'
import AlertsView from '../views/AlertsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/changelog',
      name: 'changelog',
      component: ChangeLogView,
    },
    {
      path: '/alerts',
      name: 'alerts',
      component: AlertsView,
    },
  ],
})

export default router
