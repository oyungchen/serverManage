<template>
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg shadow-xl p-6 w-full max-w-md">
      <h2 class="text-xl font-bold mb-4">{{ service ? '编辑服务' : '添加服务' }}</h2>

      <form @submit.prevent="handleSubmit">
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">服务名称</label>
            <input
              v-model="form.name"
              type="text"
              required
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">启动脚本</label>
            <input
              v-model="form.startScript"
              type="text"
              required
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">停止脚本</label>
            <input
              v-model="form.stopScript"
              type="text"
              required
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">重启脚本</label>
            <input
              v-model="form.restartScript"
              type="text"
              required
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">工作目录</label>
            <input
              v-model="form.workDir"
              type="text"
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">端口</label>
            <input
              v-model.number="form.port"
              type="number"
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div class="flex items-center">
            <input
              v-model="form.autoStart"
              type="checkbox"
              id="autoStart"
              class="h-4 w-4 text-blue-600 border-gray-300 rounded"
            />
            <label for="autoStart" class="ml-2 block text-sm text-gray-900">
              开机自启动
            </label>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button
            type="button"
            @click="$emit('close')"
            class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
          >
            取消
          </button>
          <button
            type="submit"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            保存
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  service: Object
})

const emit = defineEmits(['close', 'save'])

const form = ref({
  name: '',
  startScript: '',
  stopScript: '',
  restartScript: '',
  workDir: '',
  port: 0,
  autoStart: false
})

watch(() => props.service, (newVal) => {
  if (newVal) {
    form.value = { ...newVal }
  }
}, { immediate: true })

const handleSubmit = () => {
  emit('save', { ...form.value })
}
</script>