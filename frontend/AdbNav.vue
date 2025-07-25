<!-- When making changes to this file, be
     sure to implement the same changes in
     `frontend-v2/src/components/nav.tsx`.
-->

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
        <div
          style="display: flex; align-items: center; gap: 0.75rem; justify-content: space-between"
        >
          <div class="has-text-grey-dark" style="display: flex; align-items: center; gap: 0.5rem">
            <b-icon icon="account" size="is-small"></b-icon>
            <span>
              <span>{{ user }}</span>
              <span v-if="role !== 'admin'"> ({{ chapterName }})</span>
            </span>
          </div>
          <a href="/logout" style="color: LinkText">Log out</a>
        </div>
      </b-navbar-item>
      <div v-if="role === 'admin'" class="navbar-item">
        <b-select v-model="activeChapterId" @input="switchChapter">
          <option v-for="chapter in chapters" :key="chapter.ChapterID" :value="chapter.ChapterID">
            {{ chapter.Name }}
          </option>
        </b-select>
      </div>
    </template>
  </b-navbar>
</template>

<script lang="ts">
import Vue from 'vue';
import Buefy from 'buefy';
import navbarData from '../shared/nav.json';
import { SF_BAY_CHAPTER_ID } from './chapters';

Vue.use(Buefy);

export default Vue.extend({
  name: 'adb-nav',
  props: {
    page: String,
    user: String,
    role: String,
    chapterName: String,
    chapterId: Number,
  },
  data() {
    return {
      navbarData,
      chapters: [] as Array<{ ChapterID: number; Name: string }>,
      activeChapterId: this.chapterId,
    };
  },
  mounted() {
    if (this.role !== 'admin') {
      return;
    }
    this.fetchChapters();
  },
  methods: {
    hasAccess(roleRequired: string[] | undefined) {
      if (this.activeChapterId !== SF_BAY_CHAPTER_ID) {
        // If non-sfbay chapter is active, we only show non-sfbay items or admin items.
        return (
          !roleRequired ||
          roleRequired.indexOf('non-sfbay') !== -1 ||
          (roleRequired.indexOf('admin') !== -1 && this.role === 'admin')
        );
      }
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
    async fetchChapters() {
      try {
        const response = await fetch('/chapter/list');
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        const data = await response.json();
        this.chapters = data.chapters;
      } catch (error) {
        console.error('Error fetching chapters:', error);
      }
    },
    switchChapter(chapterId: number) {
      // TODO: Switching chapter & then using the browser's back button
      // can cause issues since the stale data from the previous chapter
      // is still displayed. Building a cache invalidation mechanism
      // for Vue might not be worth it for now, as this will be easier
      // to handle in React w/ TanStack Query. Almira has been warned
      // and is okay w/ this.
      window.location.href = `/auth/switch_chapter?chapter_id=${chapterId}`;
    },
  },
});
</script>
