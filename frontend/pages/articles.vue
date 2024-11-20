<template>
  <div
      :class="[isDark ? 'dark bg-black text-white' : 'bg-white text-black', 'flex flex-col min-h-screen']"
  >
    <div class="flex flex-col items-center justify-start h-full w-full max-w-4xl mx-auto">
      <!-- 顶部导航栏 -->
      <div class="sticky top-0 z-10 shadow-md w-full">
        <nav class="navbar flex justify-between items-center px-4 py-2 bg-white dark:bg-black border-b dark:border-gray-700">
          <n-button
              text
              v-if="showViewer"
              @click="returnToArticleList"
              class="transition-colors duration-300"
              :class="[isDark ? 'text-white' : 'text-black']"
          >
            返回文章列表
          </n-button>
        </nav>

        <!-- 文件夹选择框 -->
        <div class="folder-selection px-4 py-2 border-t dark:border-gray-700">
          <n-select
              v-model="selectedFolder"
              :options="folderOptions"
              placeholder="选择文件夹"
              @update:value="fetchMarkdownFiles"
              class="w-full"
              :style="{
              backgroundColor: isDark ? '#222' : '#fff',
              color: isDark ? '#fff' : '#000',
              borderColor: isDark ? '#444' : '#ddd'
            }"
          />
        </div>
      </div>

      <!-- 主内容区域 -->
      <div class="flex flex-1 w-full items-center justify-center px-4 py-6">
        <!-- 加载指示器 -->
        <div v-if="isLoading" class="loading-indicator">
          <n-spin size="large" />
        </div>

        <!-- 文章列表 -->
        <div
            v-else-if="!showViewer"
            class="article-list grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 w-full"
        >
          <div
              v-for="file in fileOptions"
              :key="file.value"
              class="article-card"
          >
            <n-card
                title="点击查看文章"
                hoverable
                @click="viewMarkdown(file.value)"
                :style="{
                backgroundColor: isDark ? '#333' : '#fff',
                color: isDark ? '#fff' : '#000',
                borderColor: isDark ? '#444' : '#ddd'
              }"
                class="cursor-pointer transition-all duration-300 transform hover:scale-105"
            >
              <p>{{ file.label }}</p>
            </n-card>
          </div>
        </div>

        <!-- 文章查看 -->
        <div v-else class="article-viewer w-full">
          <div v-if="isLoading" class="loading-indicator">
            <n-spin size="large" />
          </div>
          <div v-else>
            <div
                v-html="parsedHtml"
                class="markdown-body mx-auto border rounded-lg shadow-lg p-6"
            ></div>
            <n-button
                block
                type="primary"
                class="mt-6 transition-colors duration-300"
                :style="{
                backgroundColor: isDark ? '#444' : '#f5f5f5',
                color: isDark ? '#fff' : '#000'
              }"
                @click="returnToArticleList"
            >
              返回列表
            </n-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from "vue";
import axios from "axios";
import { marked } from "marked";
import hljs from "highlight.js";
import "highlight.js/styles/github-dark.css"; // 替换为其他样式
import "github-markdown-css/github-markdown-light.css";
import "github-markdown-css/github-markdown-dark.css";

// 配置 marked 和 highlight.js
marked.setOptions({
  highlight: (code, lang) => {
    if (lang && hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang }).value;
    }
    return hljs.highlightAuto(code).value;
  },
});

// 从根组件继承暗黑模式状态
const isDark = inject("isDark");

// 状态变量
const folderOptions = ref([]);
const fileOptions = ref([]);
const selectedFolder = ref("");
const parsedHtml = ref("");
const isLoading = ref(false);
const showViewer = ref(false);

// 获取文件夹列表
async function fetchFolders() {
  const token = localStorage.getItem("token");
  isLoading.value = true;
  try {
    const response = await axios.get("http://localhost:8080/api/upload/folders", {
      headers: { Authorization: `Bearer ${token}` },
    });
    folderOptions.value = response.data.folders.map((folder) => ({
      label: folder,
      value: folder,
    }));
  } catch (error) {
    console.error("文件夹加载失败:", error);
    alert("文件夹加载失败，请检查后端服务");
  } finally {
    isLoading.value = false;
  }
}

// 获取 Markdown 文件列表
async function fetchMarkdownFiles(folder) {
  if (!folder) return;
  selectedFolder.value = folder;
  const token = localStorage.getItem("token");
  isLoading.value = true;
  try {
    const response = await axios.get(`http://localhost:8080/api/markdown/files/${folder}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    fileOptions.value = response.data.files.map((file) => ({
      label: file,
      value: file,
    }));
  } catch (error) {
    console.error("Markdown 文件列表加载失败:", error);
    alert("文件列表加载失败，请检查后端服务");
  } finally {
    isLoading.value = false;
  }
}

// 查看 Markdown 文件
async function viewMarkdown(file) {
  if (!file || !selectedFolder.value) {
    alert("请先选择文件夹！");
    return;
  }
  const token = localStorage.getItem("token");
  isLoading.value = true;
  try {
    const response = await axios.get(`http://localhost:8080/api/markdown/${file}`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { folder: selectedFolder.value },
    });
    parsedHtml.value = marked.parse(response.data.content);
    showViewer.value = true;
  } catch (error) {
    console.error("Markdown 加载失败:", error);
    alert("无法加载文章内容！");
  } finally {
    isLoading.value = false;
  }
}

// 返回文章列表
function returnToArticleList() {
  showViewer.value = false;
}

// 初始化加载
onMounted(fetchFolders);
</script>

<style scoped>
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
}

.article-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.article-card {
  cursor: pointer;
}

.article-viewer {
  width: 100%;
  max-width: 800px;
  margin: auto;
}

.loading-indicator {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.markdown-body pre {
  background-color: #f6f8fa; /* 调整为合适的背景色 */
  padding: 1rem;
  border-radius: 8px;
  overflow-x: auto;
}

.dark .markdown-body pre {
  background-color: #282c34;
}

.markdown-body {
  max-width: 800px;
  margin: 0 auto;
  background-color: #f9f9f9;
  color: #333;
  padding: 1rem;
  border-radius: 8px;
  overflow-wrap: break-word;
  transition: background-color 0.3s, color 0.3s;
}

.dark .markdown-body {
  background-color: #1a1a1a;
  color: white;
}
</style>
