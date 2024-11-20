import axios from 'axios';

export default defineNuxtPlugin((nuxtApp) => {
    const config = useRuntimeConfig();

    // 创建一个 Axios 实例
    const api = axios.create({
        baseURL: config.public.apiBase, // 使用配置中的 baseURL
        headers: {
            'Content-Type': 'application/json',
        },
    });

    // 请求拦截器来附加 JWT 令牌
    api.interceptors.request.use(
        (requestConfig) => {
            const token = localStorage.getItem('token');
            if (token) {
                requestConfig.headers.Authorization = `Bearer ${token}`;
            }
            return requestConfig;
        },
        (error) => Promise.reject(error)
    );

    return {
        provide: {
            api,
        },
    };
});
