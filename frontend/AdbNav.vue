<template>
  <b-navbar fixed-top shadow id="mainNav">
    <template #brand>
      <b-navbar-item tag="div">
        <img src="/static/img/logo1.png" alt="DxE" />
      </b-navbar-item>
    </template>
    <template #start>
      <b-navbar-dropdown label="Attendance" collapsible v-if="isAttendance(role)">
        <b-navbar-item href="/" :active="page === 'NewEvent'">
          New Event
        </b-navbar-item>
        <b-navbar-item href="/list_events" :active="page === 'EventList'" class="mb-2">
          All Events
        </b-navbar-item>
        <b-navbar-item
          href="/new_connection"
          :active="page === 'NewConnection'"
          v-if="isOrganizer(role)"
        >
          New Coaching
        </b-navbar-item>
        <b-navbar-item
          href="/list_connections"
          :active="page === 'ConnectionsList'"
          v-if="isOrganizer(role)"
        >
          All Coachings
        </b-navbar-item>
      </b-navbar-dropdown>
      <b-navbar-dropdown label="Groups" collapsible v-if="isOrganizer(role)">
        <b-navbar-item href="/list_geocircles" :active="page === 'GeoCirclesList'">
          Geo-Circles
        </b-navbar-item>
        <b-navbar-item href="/list_circles" :active="page === 'CirclesList'">
          Interest Circles
        </b-navbar-item>
        <b-navbar-item href="/list_working_groups" :active="page === 'WorkingGroupList'">
          Working Groups
        </b-navbar-item>
      </b-navbar-dropdown>
      <b-navbar-dropdown label="Prospects" collapsible v-if="isOrganizer(role)">
        <b-navbar-item href="/community_prospects" :active="page === 'CommunityProspects'">
          Community Prospects
        </b-navbar-item>
        <b-navbar-item href="/chapter_member_prospects" :active="page === 'ChapterMemberProspects'">
          Chapter Member Prospects
        </b-navbar-item>
        <b-navbar-item href="/organizer_prospects" :active="page === 'OrganizerProspects'">
          Organizer Prospects
        </b-navbar-item>
      </b-navbar-dropdown>
      <b-navbar-dropdown label="Activists" collapsible v-if="isOrganizer(role)">
        <b-navbar-item
          href="/chapter_member_development"
          :active="page === 'ChapterMemberDevelopment'"
        >
          Chapter Members
        </b-navbar-item>
        <b-navbar-item href="/activist_development" :active="page === 'OrganizerDevelopment'">
          Organizers
        </b-navbar-item>
        <b-navbar-item href="/leaderboard" :active="page === 'Leaderboard'">
          Leaderboard
        </b-navbar-item>
        <b-navbar-item href="list_activists" :active="page === 'ActivistList'">
          All Activists
        </b-navbar-item>
      </b-navbar-dropdown>
      <b-navbar-dropdown label="Admin" collapsible v-if="isAdmin(role)">
        <b-navbar-item href="/list_chapters" :active="page === 'ChaptersList'">
          Chapters
        </b-navbar-item>
        <b-navbar-item href="/admin/users" :active="page === 'UserList'">
          Users
        </b-navbar-item>
      </b-navbar-dropdown>
    </template>

    <template #end>
      <b-navbar-item tag="div">
        <div class="mx-2">
          <b-icon icon="account" size="is-small"></b-icon>
          {{ user }}
        </div>
        <div class="buttons">
          <a href="/logout" class="button is-light">
            Log out
          </a>
        </div>
      </b-navbar-item>
    </template>
  </b-navbar>
</template>

<script lang="ts">
import Vue from 'vue';
import Buefy from 'buefy';

Vue.use(Buefy);

export default Vue.extend({
  name: 'adb-nav',
  props: {
    page: String,
    user: String,
    role: String,
  },
  created() {},
  methods: {
    isAdmin(role: string) {
      return role === 'admin';
    },
    isOrganizer(role: string) {
      return role === 'admin' || role === 'organizer';
    },
    isAttendance(role: string) {
      return role === 'admin' || role === 'organizer' || role === 'attendance';
    },
  },
});
</script>
