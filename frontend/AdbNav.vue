<template>
  <b-navbar fixed-top shadow id="mainNav">
    <template #brand>
      <b-navbar-item tag="div">
        <img src="/static/img/logo1.png" alt="DxE" />
      </b-navbar-item>
    </template>
    <template #start>
      <template v-for="dropdown in navbarData.items">
        <b-navbar-dropdown
          v-if="hasAccess(dropdown.roleRequired)"
          :label="dropdown.label"
          :key="dropdown.label"
          collapsible
        >
          <b-navbar-item
            v-for="item in dropdown.items"
            :key="item.href"
            v-if="hasAccess(item.roleRequired)"
            :href="item.href"
            :active="page === item.page"
            :class="{ 'mb-2': item.separatorBelow }"
          >
            {{ item.label }}
          </b-navbar-item>
        </b-navbar-dropdown>
      </template>
    </template>
    <template #end>
      <b-navbar-item tag="div">
        <div class="mx-2 has-text-grey-dark">
          <b-icon icon="account" size="is-small" class="mx-1"></b-icon>
          {{ user }} ({{ chapter }})
        </div>
        <div>
          <a href="/logout" style="color: inherit">Log out</a>
        </div>
      </b-navbar-item>
    </template>
  </b-navbar>
</template>

<script lang="ts">
import Vue from 'vue';
import Buefy from 'buefy';
import navbarData from './nav.json';

Vue.use(Buefy);

export default Vue.extend({
  name: 'adb-nav',
  props: {
    page: String,
    user: String,
    role: String,
    chapter: String,
  },
  data() {
    return {
      navbarData,
    };
  },
  methods: {
    hasAccess(roleRequired: string[] | undefined) {
      return (
        !roleRequired ||
        roleRequired.some((it) =>
          it === 'admin'
            ? this.role === 'admin'
            : it === 'organizer'
              ? this.role === 'admin' || this.role === 'organizer'
              : it === 'attendance'
                ? this.role === 'admin' || this.role === 'organizer' || this.role === 'attendance'
                : it === 'non-sfbay'
                  ? this.role === 'non-sfbay'
                  : false,
        )
      );
    },
  },
});
</script>
