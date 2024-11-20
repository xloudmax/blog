<template>
  <div :class="['login', { 'dark': isDark }]">
    <h2 class="text-center text-2xl font-bold mb-6 dark:text-white">登录</h2>
    <form @submit.prevent="loginUser">
      <div class="mb-4">
        <label for="username" class="block text-sm dark:text-white">用户名:</label>
        <input
            type="text"
            id="username"
            v-model="username"
            placeholder="请输入用户名"
            class="w-full p-2 rounded border border-gray-300 dark:bg-gray-700 dark:text-white"
        />
      </div>
      <div class="mb-4">
        <label for="password" class="block text-sm dark:text-white">密码:</label>
        <input
            type="password"
            id="password"
            v-model="password"
            placeholder="请输入密码"
            class="w-full p-2 rounded border border-gray-300 dark:bg-gray-700 dark:text-white"
        />
      </div>
      <button
          type="submit"
          class="w-full p-2 mt-4 bg-blue-500 text-white rounded hover:bg-blue-600"
      >
        登录
      </button>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useNuxtApp, useState } from '#app';

const username = ref('');
const password = ref('');
const isDark = ref(false);
const router = useRouter();

// 登录状态管理
const isLoggedIn = useState('isLoggedIn', () => false);

// 获取 Axios 实例
const { $api } = useNuxtApp();

const loginUser = async () => {
  try {
    const response = await $api.post('/api/login', {
      username: username.value,
      password: password.value,
    });

    const token = response.data.data.token;
    localStorage.setItem('token', token); // 保存令牌到 localStorage
    isLoggedIn.value = true; // 更新登录状态
    console.log('User logged in:', response.data);
    alert('登录成功');

    router.push('/');
  } catch (error) {
    console.error('Error logging in user:', error);
    alert(error.response?.data?.message || '登录失败');
  }
};
</script>

<style scoped>
.login {
  @apply max-w-md mx-auto p-8 rounded-lg shadow-lg transition-colors duration-300;
  background-color: #fff;
  color: #000;
}
.dark .login {
  @apply bg-gray-800 text-white;
}
</style>
