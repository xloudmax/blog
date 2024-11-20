<template>
  <div :class="['register', { 'dark': isDark }]">
    <h2 class="text-center text-2xl font-bold mb-6 dark:text-white">注册</h2>

    <!-- 注册表单 -->
    <form v-if="!isVerifying" @submit.prevent="submitRegistration">
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
        <label for="email" class="block text-sm dark:text-white">邮箱:</label>
        <input
            type="email"
            id="email"
            v-model="email"
            placeholder="请输入邮箱"
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
      <div class="mb-4">
        <label for="inviteCode" class="block text-sm dark:text-white">邀请码(可选):</label>
        <input
            type="text"
            id="inviteCode"
            v-model="inviteCode"
            placeholder="请输入邀请码"
            class="w-full p-2 rounded border border-gray-300 dark:bg-gray-700 dark:text-white"
        />
      </div>
      <button type="submit" class="w-full p-2 mt-4 bg-green-500 text-white rounded hover:bg-green-600">注册</button>
    </form>

    <!-- 验证码表单 -->
    <form v-else @submit.prevent="verifyEmailCode">
      <h2 class="text-center text-2xl font-bold mb-6 dark:text-white">邮箱验证</h2>
      <div class="mb-4">
        <label for="verificationCode" class="block text-sm dark:text-white">验证码:</label>
        <input
            type="text"
            id="verificationCode"
            v-model="verificationCode"
            placeholder="请输入邮箱验证码"
            class="w-full p-2 rounded border border-gray-300 dark:bg-gray-700 dark:text-white"
        />
      </div>
      <button type="submit" class="w-full p-2 mt-4 bg-blue-500 text-white rounded hover:bg-blue-600">验证</button>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useNuxtApp } from '#app';

const username = ref('');
const email = ref('');
const password = ref('');
const inviteCode = ref(''); // 保存邀请码
const isDark = ref(false);
const isVerifying = ref(false); // 控制是否显示验证码表单
const verificationCode = ref('');
const userEmail = ref(''); // 保存注册时填写的邮箱

const { $api } = useNuxtApp(); // 使用 Nuxt 的 API 插件

// 提交注册表单
const submitRegistration = async () => {
  try {
    const response = await $api.post('http://localhost:8080/api/register', {
      username: username.value,
      email: email.value,
      password: password.value,
      invite_code: inviteCode.value, // 将邀请码发送到后端
    });
    console.log('注册成功，请检查邮箱验证码:', response.data);
    alert('注册成功！请检查邮箱并输入验证码');

    // 保存邮箱并显示验证码输入表单
    userEmail.value = email.value;
    isVerifying.value = true;
  } catch (error) {
    console.error('注册失败:', error);
    alert(error.response?.data?.message || '注册失败');
  }
};

// 验证邮箱验证码
const verifyEmailCode = async () => {
  try {
    const response = await $api.post('/api/verify_email', {
      email: userEmail.value,
      code: verificationCode.value,
    });
    console.log('邮箱验证成功:', response.data);
    alert('邮箱验证成功！');
    isVerifying.value = false; // 隐藏验证码输入表单
  } catch (error) {
    console.error('邮箱验证失败:', error);
    alert(error.response?.data?.message || '邮箱验证失败');
  }
};
</script>

<style scoped>
.register {
  @apply max-w-md mx-auto p-8 rounded-lg shadow-lg transition-colors duration-300;
  background-color: #fff;
  color: #000;
}

.dark .register {
  @apply bg-gray-800 text-white;
}
</style>
