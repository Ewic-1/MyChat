<template>
  <div class="auth-shell">
    <section class="auth-shell__brand">
      <p class="auth-shell__eyebrow">MYCHAT ENTERPRISE</p>
      <h1 class="auth-shell__title">{{ title }}</h1>
      <p class="auth-shell__subtitle">{{ subtitle }}</p>
      <ul class="auth-shell__highlights">
        <li>统一会话中心</li>
        <li>企业级权限与审计</li>
        <li>稳定实时通信体验</li>
      </ul>
    </section>

    <section class="auth-shell__card">
      <slot />
      <slot name="footer" />
    </section>
  </div>
</template>

<script>
export default {
  name: 'AuthShell',
  props: {
    title: {
      type: String,
      default: '欢迎使用 MyChat',
    },
    subtitle: {
      type: String,
      default: '安全、高效、现代化的企业沟通平台',
    },
  },
}
</script>
