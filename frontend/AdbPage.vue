<template>
  <div :class="wrapperClass">
    <div class="title">
      <h1>{{ title }}</h1>
      <br />
    </div>

    <div class="main"><slot></slot></div>
  </div>
</template>

<script lang="ts">
import Vue from 'vue';

export default Vue.extend({
  props: {
    title: String,
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
