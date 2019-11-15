<template>
  <div :class="wrapperClass">
    <div class="title">
      <h1>{{ title }}</h1>
    </div>

    <div class="description">{{ description }}</div>

    <br />

    <div class="main"><slot></slot></div>
  </div>
</template>

<script lang="ts">
import Vue from 'vue';
import * as Sentry from '@sentry/browser';
import * as Integrations from '@sentry/integrations';

Sentry.init({
  dsn: 'https://1bc5adff2f574d5390f085353326f0d5@sentry.io/1820807',
  integrations: [new Integrations.Vue({ Vue, attachProps: true })],
});

Sentry.captureException(new Error('Something broke'));

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
