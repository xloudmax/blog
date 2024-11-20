<template>
  <div>
    <h1>{{ post.title }}</h1>
    <nuxt-content :document="post" />
  </div>
</template>

<script setup>
import { useContent } from '@nuxt/content'
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const post = ref(null)
const route = useRoute()

onMounted(async () => {
  const { findOne } = useContent()
  post.value = await findOne('blog', route.params.slug)
})
</script>
