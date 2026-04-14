<template>
  <div class="min-h-screen bg-gray-100">
    <header class="bg-white shadow">
      <div class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8 flex justify-between items-center">
        <h1 class="text-3xl font-bold text-gray-900">ServerManage</h1>
        <button
          @click="startAll"
          :disabled="loading"
          class="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded disabled:opacity-50"
        >
          启动全部
        </button>
      </div>
    </header>

    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <ServiceList
        :services="services"
        @refresh="loadServices"
        @start="handleStart"
        @stop="handleStop"
        @restart="handleRestart"
        @edit="editService"
        @delete="handleDelete"
      />

      <div class="mt-4 flex gap-4">
        <button
          @click="showModal = true"
          class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
        >
          + 添加服务
        </button>
      </div>

      <ServiceModal
        v-if="showModal"
        :service="editingService"
        @close="closeModal"
        @save="saveService"
      />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getServices, createService, updateService, deleteService, startService, stopService, restartService } from './api'
import ServiceList from './components/ServiceList.vue'
import ServiceModal from './components/ServiceModal.vue'

const services = ref([])
const showModal = ref(false)
const editingService = ref(null)
const loading = ref(false)

const loadServices = async () => {
  try {
    const res = await getServices()
    services.value = res.data
  } catch (e) {
    console.error('Failed to load services:', e)
  }
}

const handleStart = async (id) => {
  await startService(id)
  await loadServices()
}

const handleStop = async (id) => {
  await stopService(id)
  await loadServices()
}

const handleRestart = async (id) => {
  await restartService(id)
  await loadServices()
}

const startAll = async () => {
  loading.value = true
  for (const svc of services.value) {
    if (svc.status === 'stopped') {
      await startService(svc.id)
    }
  }
  await loadServices()
  loading.value = false
}

const editService = (service) => {
  editingService.value = service
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  editingService.value = null
}

const saveService = async (data) => {
  if (editingService.value) {
    await updateService(editingService.value.id, data)
  } else {
    await createService(data)
  }
  await loadServices()
  closeModal()
}

const handleDelete = async (id) => {
  if (confirm('确定要删除这个服务吗？')) {
    await deleteService(id)
    await loadServices()
  }
}

onMounted(() => {
  loadServices()
})
</script>