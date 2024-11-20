<template>
  <div
      :class="[isDark ? 'dark bg-black text-white' : 'bg-white text-black', 'flex items-center justify-center min-h-screen p-4']"
  >
    <div class="w-full max-w-4xl flex flex-col h-[100vh]">
      <!-- 文件夹选择框 -->
      <n-select
          v-model:value="selectedFolder"
          :options="folderOptions"
          placeholder="选择文件夹"
          class="mb-4"
      />

      <!-- 新建文件夹按钮 -->
      <div class="mb-4 flex items-center">
        <n-input
            v-model:value="newFolderName"
            placeholder="新建文件夹名称"
            class="mr-2"
        />
        <n-button @click="handleCreateFolder">新建文件夹</n-button>
      </div>

      <!-- 文件标题输入框 -->
      <n-input
          v-model:value="title"
          placeholder="请输入文件标题"
          class="mb-4"
      />

      <!-- Markdown 编辑器 -->
      <v-md-editor
          v-model="markdownContent"
          class="w-full flex-1 editor github-markdown-body"
          :disabled-menus="[]"
          @save="handleSave"
          @upload-image="handleUploadImage"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted, nextTick, watch } from 'vue';
import axios from 'axios';
import { NInput, NSelect, NButton } from 'naive-ui';
import katex from 'katex';
import 'katex/dist/katex.min.css';

// 获取全局的 isDark 状态
const isDark = inject('isDark');
const markdownContent = ref(''); // Markdown 内容绑定
const title = ref(''); // 文件标题绑定
const folderOptions = ref([]); // 文件夹选项
const selectedFolder = ref(''); // 用户选择的文件夹
const newFolderName = ref(''); // 新建文件夹的名称

// 渲染 KaTeX 函数
function renderKatex() {
  const elements = document.querySelectorAll('.github-markdown-body');
  elements.forEach((el) => {
    el.innerHTML = el.innerHTML.replace(/<br>/g, ''); // 去除多余的 <br> 标签
    try {
      el.innerHTML = el.innerHTML.replace(/\$\$([^$]+)\$\$/g, (_, math) => katex.renderToString(math, { displayMode: true }));
      el.innerHTML = el.innerHTML.replace(/\$([^$]+)\$/g, (_, math) => katex.renderToString(math, { displayMode: false }));
    } catch (error) {
      console.error('KaTeX 渲染错误:', error);
    }
  });
}

// 监控 markdownContent 内容变化时渲染 KaTeX
watch(markdownContent, async () => {
  await nextTick();
  renderKatex();
});

// 获取文件夹列表
async function fetchFolders() {
  const token = localStorage.getItem('token');
  try {
    const response = await axios.get('http://localhost:8080/api/upload/folders', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    folderOptions.value = response.data.folders.map(folder => ({ label: folder, value: folder }));
    console.log("Folder options loaded:", folderOptions.value); // 调试信息
  } catch (error) {
    console.error("获取文件夹失败:", error);
  }
}

// 创建新文件夹
async function handleCreateFolder() {
  if (!newFolderName.value) {
    alert("请输入文件夹名称！");
    return;
  }
  const token = localStorage.getItem('token');
  try {
    // 使用正确的 JSON 字段名 `folder`
    await axios.post('http://localhost:8080/api/upload/folders', { folder: newFolderName.value }, {
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    });
    newFolderName.value = '';
    fetchFolders(); // 刷新文件夹列表
    alert("文件夹创建成功！");
  } catch (error) {
    console.error("文件夹创建失败:", error);
    alert("文件夹创建失败，请重试！");
  }
}

// 保存并上传 Markdown 文件
async function handleSave() {
  if (!markdownContent.value) {
    alert("编辑器内容为空！");
    return;
  }
  if (!title.value) {
    alert("请输入文件标题！");
    return;
  }
  if (!selectedFolder.value) {
    alert("请选择一个文件夹！");
    return;
  }

  console.log("Selected folder:", selectedFolder.value); // 确认文件夹选择

  const blob = new Blob([markdownContent.value], { type: 'text/markdown' });
  const file = new File([blob], `${title.value}.md`, { type: 'text/markdown' });
  const formData = new FormData();
  formData.append('file', file);
  formData.append('title', title.value);
  formData.append('folder', selectedFolder.value); // 将文件夹添加到 FormData

  const token = localStorage.getItem('token');
  try {
    const response = await axios.post('http://localhost:8080/api/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
        Authorization: `Bearer ${token}`,
      },
    });
    alert("文件上传成功！");
  } catch (error) {
    console.error("上传失败:", error);
    alert("文件上传失败，请重试！");
  }
}

// 组件加载时获取文件夹列表
onMounted(() => {
  fetchFolders();
  document.addEventListener('keydown', (event) => {
    if (event.ctrlKey && event.key === 's') {
      event.preventDefault();
      handleSave();
    }
  });
});
</script>

<style scoped>
@import 'katex/dist/katex.min.css';
</style>
