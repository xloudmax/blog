<template>
  <NConfigProvider :theme="darkTheme" :theme-overrides="computedThemeOverrides">
    <NLayout id="app" class="flex flex-col min-h-screen font-mono pb-12 dark:bg-black dark:text-white bg-white text-black">
      <!-- 加载模板 -->
      <div
          v-if="isLoading"
          :class="[
          'fixed inset-0 flex items-center justify-center z-50 transition-opacity duration-300',
          isDark ? 'bg-black bg-opacity-90' : 'bg-white bg-opacity-90',
        ]"
      >
        <div class="text-lg font-bold" :class="isDark ? 'text-white' : 'text-black'">加载中...</div>
      </div>

      <!-- 顶部带菜单和暗黑模式切换 -->
      <NLayoutHeader class="shadow-md bg-white dark:bg-black flex justify-between items-center p-4">
        <!-- NMenu 组件，基于 isLoggedIn 的动态 key -->
        <NMenu
            :key="menuKey"
            :options="menuOptions"
            @update:value="handleMenuClick"
            mode="horizontal"
            :class="isDark ? 'text-white' : 'text-black'"
        />
        <!-- 暗黑模式切换按钮 -->
        <NButton
            @click="toggleDarkMode"
            class="ml-4"
            :style="{
            backgroundColor: isDark ? '#333333' : '#f0f0f0',
            color: isDark ? '#ffffff' : '#000000',
          }"
            type="default"
            text
        >
          {{ isDark ? '白天' : '夜间' }}
        </NButton>
      </NLayoutHeader>

      <!-- 主内容区域 -->
      <NLayoutContent class="main-content flex-1 p-8 relative overflow-auto bg-white dark:bg-black">
        <NuxtPage @login-success="onLoginSuccess" />
      </NLayoutContent>

      <!-- 底部 -->
      <NLayoutFooter
          class="footer fixed bottom-0 left-0 w-full text-center p-4 text-sm z-10 bg-gray-100 dark:bg-gray-900 text-black dark:text-white"
      >
        <p>© 2024 Xloudmax</p>
      </NLayoutFooter>
    </NLayout>
  </NConfigProvider>
</template>

<script setup>
import {
  NConfigProvider,
  NLayout,
  NLayoutHeader,
  NLayoutContent,
  NLayoutFooter,
  NMenu,
  NButton,
  darkTheme,
} from 'naive-ui';
import { useRouter } from 'vue-router';
import { ref, computed, provide, onMounted, watch } from 'vue';

const router = useRouter();

// 用户登录状态
const isLoggedIn = ref(false);

// 菜单选项
const loggedOutMenuOptions = [
  { label: '主页', key: '/' },
  { label: '登录', key: '/login' },
  { label: '注册', key: '/register' },
];

const loggedInMenuOptions = [
  {label: '主页', key: '/'},
  {label: '文章', key: '/articles'},
  {label: '编辑器', key: '/editor'},
  {label: '登出', key: 'logout'},
];

const menuOptions = ref(loggedOutMenuOptions);
const menuKey = ref(0);

// 根据登录状态更新菜单
watch(isLoggedIn, (newValue) => {
  menuOptions.value = newValue ? loggedInMenuOptions : loggedOutMenuOptions;
  menuKey.value += 1;
});

// 登录成功
const onLoginSuccess = () => {
  isLoggedIn.value = true;
  router.push('/');
};

// 菜单点击事件
const handleMenuClick = (key) => {
  if (key === 'logout') {
    logoutUser();
  } else {
    router.push(key);
  }
};

// 检查登录状态
const checkLoginStatus = () => {
  isLoggedIn.value = !!localStorage.getItem('token');
};

// 用户登出
const logoutUser = () => {
  localStorage.removeItem('token');
  isLoggedIn.value = false;
  router.push('/');
};

// 暗黑模式
const isDark = ref(false);
const toggleDarkMode = () => {
  isDark.value = !isDark.value;
  document.documentElement.classList.toggle('dark', isDark.value);
  localStorage.setItem('isDark', isDark.value);
};

// 页面加载时设置
onMounted(() => {
  const savedDarkMode = localStorage.getItem('isDark') === 'true';
  isDark.value = savedDarkMode;
  document.documentElement.classList.toggle('dark', savedDarkMode);
  checkLoginStatus();
});

// 提供暗黑模式状态到子组件
provide('isDark', isDark);
provide('toggleDarkMode', toggleDarkMode);

// 动态主题
const computedThemeOverrides = computed(() => ({
  common: {
    textColorBase: isDark.value ? '#ffffff' : '#000000',
    bodyColor: isDark.value ? '#000000' : '#ffffff',
  },
  Layout: {
    headerColor: isDark.value ? '#000000' : '#ffffff',
  },
  Menu: {
    itemTextColor: isDark.value ? '#ffffff' : '#000000',
    itemTextColorActive: isDark.value ? '#e0e0e0' : '#333333',
  },
}));

// 加载状态
const isLoading = ref(false);
onMounted(async () => {
  isLoading.value = true;
  await new Promise((resolve) => setTimeout(resolve, 1500));
  isLoading.value = false;
});
</script>

<style scoped>
body {
  @apply transition-colors duration-300;
  transition: background-color 0.3s, color 0.3s;
}
</style>
