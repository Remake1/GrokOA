import { createRouter, createWebHistory } from "vue-router";
import HomeView from "@/views/HomeView.vue";
import AboutView from "@/views/AboutView.vue";
import AuthView from "@/views/AuthView.vue";
import { useAuth } from "@/auth/useAuth";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView,
    },
    {
      path: "/about",
      name: "about",
      component: AboutView,
    },
    {
      path: "/auth",
      name: "auth",
      component: AuthView,
      meta: { guest: true },
    },
  ],
});

router.beforeEach((to) => {
  const { isAuthenticated } = useAuth();

  // Guest-only route (e.g. /auth) — redirect authenticated users away
  if (to.meta.guest && isAuthenticated.value) {
    return { name: "home" };
  }

  // Everything else requires auth
  if (!to.meta.guest && !isAuthenticated.value) {
    return { name: "auth" };
  }
});

export default router;
