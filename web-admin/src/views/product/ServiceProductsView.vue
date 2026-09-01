<script setup lang="ts">
import MerchantProductsTab from '@/views/merchant/components/MerchantProductsTab.vue'
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">服务管理</h1>
        <p class="page-subtitle">维护康养套餐、陪诊服务等服务项目信息，服务订单下单时需绑定服务对象档案。</p>
      </div>
    </div>
    <MerchantProductsTab mode="service" />
  </div>
</template>
