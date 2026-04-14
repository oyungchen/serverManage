<template>
  <div class="flex items-center justify-between">
    <div class="flex items-center">
      <span
        :class="[
          'inline-flex items-center justify-center h-3 w-3 rounded-full mr-3',
          service.status === 'running' ? 'bg-green-500' : 'bg-gray-400'
        ]"
      ></span>
      <div>
        <p class="text-lg font-medium text-gray-900">{{ service.name }}</p>
        <p class="text-sm text-gray-500">
          {{ service.status === 'running' ? '运行中' : '已停止' }}
          <span v-if="service.port"> :{{ service.port }}</span>
        </p>
      </div>
    </div>

    <div class="flex items-center gap-2">
      <button
        v-if="service.status === 'running'"
        @click="$emit('stop')"
        class="bg-red-500 hover:bg-red-600 text-white px-3 py-1 rounded text-sm"
      >
        停止
      </button>
      <button
        v-else
        @click="$emit('start')"
        class="bg-green-500 hover:bg-green-600 text-white px-3 py-1 rounded text-sm"
      >
        启动
      </button>

      <button
        @click="$emit('restart')"
        :disabled="service.status !== 'running'"
        class="bg-yellow-500 hover:bg-yellow-600 text-white px-3 py-1 rounded text-sm disabled:opacity-50"
      >
        重启
      </button>

      <button
        @click="$emit('edit')"
        class="bg-gray-500 hover:bg-gray-600 text-white px-3 py-1 rounded text-sm"
      >
        编辑
      </button>

      <button
        @click="$emit('delete')"
        class="bg-gray-300 hover:bg-gray-400 text-gray-700 px-3 py-1 rounded text-sm"
      >
        删除
      </button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  service: {
    type: Object,
    required: true
  }
})

defineEmits(['start', 'stop', 'restart', 'edit', 'delete'])
</script>