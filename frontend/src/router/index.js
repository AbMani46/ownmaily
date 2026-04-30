import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  { path: '/login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
  { path: '/setup', component: () => import('@/views/SetupWizardView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', component: () => import('@/views/DashboardView.vue'), meta: { title: 'Dashboard' } },
      { path: 'subscribers', component: () => import('@/views/SubscribersView.vue'), meta: { title: 'Subscribers' } },
      { path: 'subscribers/:id', component: () => import('@/views/SubscriberDetailView.vue'), meta: { title: 'Subscriber' } },
      { path: 'lists', component: () => import('@/views/ListsView.vue'), meta: { title: 'Lists' } },
      { path: 'lists/:id', component: () => import('@/views/ListDetailView.vue'), meta: { title: 'List' } },
      { path: 'tags', component: () => import('@/views/TagsView.vue'), meta: { title: 'Tags' } },
      { path: 'campaigns', component: () => import('@/views/CampaignsView.vue'), meta: { title: 'Campaigns' } },
      { path: 'campaigns/new', component: () => import('@/views/CampaignEditView.vue'), meta: { title: 'New Campaign' } },
      { path: 'campaigns/:id/edit', component: () => import('@/views/CampaignEditView.vue'), meta: { title: 'Edit Campaign' } },
      { path: 'campaigns/:id/stats', component: () => import('@/views/CampaignStatsView.vue'), meta: { title: 'Campaign Stats' } },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) {
    return '/login'
  }
})

export default router
