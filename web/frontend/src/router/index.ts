import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router'; // Use type-only import
import HomeView from '../views/HomeView.vue'; // Assume HomeView exists for listing
import RepoDetailView from '../views/RepoDetailView.vue'; // Assume RepoDetailView exists for viewing content
import RepoFormView from '../views/RepoFormView.vue'; // Assume RepoFormView exists for add/edit

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'Home',
    component: HomeView,
    meta: { title: 'Monitored Repositories' }
  },
  {
    path: '/repo/:id/view', // Route for viewing aggregated content
    name: 'RepoDetail',
    component: RepoDetailView,
    props: true, // Pass route params as component props (id)
    meta: { title: 'View Content' }
  },
  {
    path: '/repo/new', // Route for adding a new repository
    name: 'RepoAdd',
    component: RepoFormView,
    meta: { title: 'Add Repository' }
  },
  {
    path: '/repo/:id/edit', // Route for editing an existing repository
    name: 'RepoEdit',
    component: RepoFormView,
    props: true, // Pass route params as component props (id)
    meta: { title: 'Edit Repository' }
  },
  // Add a catch-all route for 404 if needed
  // {
  //   path: '/:catchAll(.*)',
  //   name: 'NotFound',
  //   component: () => import('../views/NotFoundView.vue') // Assume NotFoundView exists
  // }
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL), // Use history mode
  routes,
});

// Optional: Add navigation guard to update document title
router.beforeEach((to, _from, next) => { // Prefix 'from' with underscore
  const title = to.meta.title as string | undefined;
  document.title = title ? `${title} - SyncDocs` : 'SyncDocs';
  next();
});

export default router;
