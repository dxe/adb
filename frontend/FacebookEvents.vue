<template>
  <adb-page title="Facebook Events" class="body-wrapper-extra-wide">
    <b-loading :is-full-page="true" v-model="loading"></b-loading>
    <b-table :data="events" striped hoverable default-sort="StartTime">
      <b-table-column field="StartTime" label="Date" v-slot="props" sortable>
        {{ dayjs(props.row.StartTime).format('YYYY-MM-DD') }}
      </b-table-column>
      <b-table-column field="Name" label="Name" v-slot="props" sortable>
        <a :href="'https://facebook.com/' + props.row.ID" target="_blank" rel="noreferrer">{{
          props.row.Name
        }}</a>
      </b-table-column>
      <b-table-column v-slot="props">
        <b-switch
          v-model="props.row.Featured"
          @input="(val) => featureEvent(props.row.ID, val)"
        >
          {{ props.row.Featured ? 'Featured' : 'Feature' }}
        </b-switch>
      </b-table-column>
      <b-table-column v-slot="props">
        <b-button @click="cancelEvent(props.row.ID)" icon-left="delete" type="is-danger">
          Delete
        </b-button>
      </b-table-column>
    </b-table>
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage } from './flash_message';
import { focus } from './directives/focus';
import dayjs from 'dayjs';

const SF_BAY_FACEBOOK_ID = '1377014279263790';

interface ExternalEvent {
  ID: number;
  Name: string;
  StartTime: string;
  Featured: boolean;
}

export default Vue.extend({
  name: 'facebook-events',
  computed: {},
  methods: {
    dayjs,
    featureEvent(id: number, featured: boolean) {
      const csrfToken = $('meta[name="csrf-token"]').attr('content');
      $.ajax({
        url: '/admin/external_events/feature',
        method: 'POST',
        headers: { 'X-CSRF-Token': csrfToken },
        contentType: 'application/json',
        data: JSON.stringify({ id, featured }),
        success: (data) => {
          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage('Successfully featured event.');

          // Update the event in the page state.
          this.events = this.events.map((it) =>
            it.ID === id ? { ...it, Featured: featured } : it,
          );
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    cancelEvent(id: number) {
      const csrfToken = $('meta[name="csrf-token"]').attr('content');
      $.ajax({
        url: '/admin/external_events/cancel',
        method: 'POST',
        headers: { 'X-CSRF-Token': csrfToken },
        contentType: 'application/json',
        data: JSON.stringify({ id }),
        success: (data) => {
          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage('Successfully cancelled event.');

          // Update the page state.
          const idx = this.events.findIndex((it) => it.ID === id);
          Vue.delete(this.events, idx);
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
  },
  data() {
    return {
      loading: true,
      events: [] as ExternalEvent[],
    };
  },
  created() {
    const csrfToken = $('meta[name="csrf-token"]').attr('content');
    // Get chapters
    $.ajax({
      url: `/external_events/${SF_BAY_FACEBOOK_ID}?start_time=${dayjs().format(
        'YYYY-MM-DD',
      )}T00:00:00Z`,
      method: 'GET',
      success: (data) => {
        const parsed = JSON.parse(data);
        if (parsed.status === 'error') {
          flashMessage('Error: ' + parsed.message, true);
          return;
        }
        // status === "success"
        this.loading = false;
        this.events = parsed.events;
      },
      error: (err) => {
        this.loading = false;
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
      },
    });
  },
  components: {
    AdbPage,
  },
  directives: {
    focus,
  },
});
</script>
