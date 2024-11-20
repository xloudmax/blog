import { createRouter, createWebHistory } from 'vue-router';
import { useNotification } from 'naive-ui';

const routes = [
    { path: '/', component: () => import('@/pages/index.vue') },
    { path: '/login', component: () => import('@/pages/Login.vue') },
    { path: '/articles', component: () => import('@/pages/Articles.vue'), meta: { requiresAuth: true } },
    { path: '/editor', component: () => import('@/pages/Editor.vue'), meta: { requiresAuth: true } },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach((to, _from, next) => {
    const token = localStorage.getItem('token');
    const notification = useNotification();

    if (to.meta.requiresAuth && !token) {
        notification.error({
            title: '访问受限',
            description: '请先登录以访问此页面！',
        });
        next({ path: '/login', query: { redirect: to.fullPath } });
    } else {
        next();
    }
});

export default router;
