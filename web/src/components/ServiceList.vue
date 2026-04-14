<template>
  <div class="bg-white shadow overflow-hidden sm:rounded-md">
    <ul class="divide-y divide-gray-200">
      <li v-for="service in services" :key="service.id" class="px-4 py-4 sm:px-6">
        <ServiceItem
          :service="service"
          @start="$emit('start', service.id)"
          @stop="$emit('stop', service.id)"
          @restart="$emit('restart', service.id)"
          @edit="$emit('edit', service)"
          @delete="$emit('delete', service.id)"
        />
      </li>
      <li v-if="services.length === 0" class="px-4 py-8 text-center text-gray-500">
        暂无服务，请添加服务
      </li>
    </ul>
  </div>
</template>

<script setup>
import ServiceItem from './ServiceItem.vue'

defineProps({
  services: {
    type: Array,
    default: () => []
  }
})

defineEmits(['refresh', 'start', 'stop', 'restart', 'edit', 'delete'])
</script>