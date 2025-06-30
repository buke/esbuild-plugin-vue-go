<script setup lang="ts">
defineProps<{
  msg: string;
}>();
</script>

<template>
  <div class="greetings">
    <h1 class="green">{{ msg }}</h1>
    <h3>
      You’ve successfully created a project with
      <a href="https://vite.dev/" target="_blank" rel="noopener">Vite</a> +
      <a href="https://vuejs.org/" target="_blank" rel="noopener">Vue 3</a>. What's next?
    </h3>
  </div>
</template>

<style scoped>
h1 {
  font-weight: 500;
  font-size: 2.6rem;
  position: relative;
  top: -10px;
}

h3 {
  font-size: 1.2rem;
}

.greetings h1,
.greetings h3 {
  text-align: center;
}

@media (min-width: 1024px) {
  .greetings h1,
  .greetings h3 {
    text-align: left;
  }
}
</style>
<!-- 
<style lang="scss" scoped>
@import '../styles/variables.scss';

.greetings {
  color: $primary-color;
  font-size: $font-size-base;
  border: 2px dashed $primary-color;
  padding: $padding-base;
  border-radius: $border-radius-base;

  .primary-btn {
    background: $primary-color;
    color: #fff;
    border: none;
    border-radius: $border-radius-base;
    padding: 0.5em 1.5em;
    cursor: pointer;
    font-size: 1em;
    margin-top: 1em;
    transition: background 0.2s;

    &:hover {
      background: $secondary-color;
    }
  }
}
</style> -->
