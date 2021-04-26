<template>
  <section :class="wrapperClass">
    <div class="container mb-5">
      <h1 class="title">{{ title }}</h1>
      <p v-if="description" class="description">{{ description }}</p>
    </div>

    <div class="container main"><slot></slot></div>
  </section>
</template>

<script lang="ts">
import Vue from 'vue';

export default Vue.extend({
  props: {
    title: String,
    description: String,
    wide: Boolean,
    narrow: Boolean,
  },

  computed: {
    wrapperClass(): string {
      if (this.wide) {
        return 'body-wrapper-extra-wide';
      }
      if (this.narrow) {
        return 'body-wrapper';
      }
      return 'body-wrapper-wide';
    },
  },
});
</script>
