<template>
  <adb-page :title="connections ? 'All Coachings' : 'Events'">
    <b-loading :is-full-page="true" v-model="loading"></b-loading>

    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-field>
            <b-switch v-model="showFilters" type="is-primary">Show filters</b-switch>
          </b-field>
        </div>
      </div>
    </nav>

    <nav class="level" v-if="showFilters">
      <div class="level-left">
        <div class="level-item">
          <b-field label-position="on-border" :label="connections ? 'Coach' : 'Event Name'">
            <b-input
              v-model="search.name"
              type="text"
              :icon="connections ? 'clipboard-account' : 'alphabetical-variant'"
            >
            </b-input>
          </b-field>
        </div>
        <div class="level-item">
          <b-field label-position="on-border" :label="connections ? 'Coachees' : 'Activist'">
            <!-- TODO: experimenting -->
            <b-taginput
              v-model="search.activist"
              :data="activistFilterOptions"
              autocomplete
              :allow-new="false"
              icon="account-outline"
              @typing="getFilteredActivists"
              maxtags="1"
              type="is-info"
              dropdown-position="bottom"
              :has-counter="false"
            ></b-taginput>
          </b-field>
        </div>

        <div class="level-item" v-if="!connections">
          <b-field label="Type" label-position="on-border">
            <b-select v-model="search.type" icon="shape">
              <option
                v-for="interest in [
                  { value: 'noConnections', display: 'All' },
                  { value: 'Action', display: 'Action' },
                  { value: 'Campaign Action', display: 'Campaign Action' },
                  { value: 'Community', display: 'Community' },
                  { value: 'Frontline Surveillance', display: 'Frontline Surveillance' },
                  { value: 'Meeting', display: 'Meeting' },
                  { value: 'Outreach', display: 'Outreach' },
                  { value: 'Sanctuary', display: 'Sanctuary' },
                  { value: 'Training', display: 'Training' },
                  { value: 'mpiDA', display: 'MPI: Direct Action' },
                  { value: 'mpiCOM', display: 'MPI: Community' },
                ]"
                :value="interest.value"
                :key="interest.value"
              >
                {{ interest.display }}
              </option>
            </b-select>
          </b-field>
        </div>
      </div>
    </nav>

    <nav class="level" v-if="showFilters">
      <div class="level-left">
        <div class="level-item">
          <b-field label="From" label-position="on-border">
            <b-input v-model="search.start" type="date" icon="calendar-start"></b-input>
          </b-field>
        </div>
        <div class="level-item">
          <b-field label="To" label-position="on-border">
            <b-input v-model="search.end" type="date" icon="calendar-end"></b-input>
          </b-field>
        </div>

        <div class="level-item">
          <b-button type="is-primary" label="Filter" @click="eventListRequest"></b-button>
        </div>
      </div>
    </nav>

    <b-table
      :data="events"
      striped
      hoverable
      default-sort="name"
      detailed
      detail-key="event_id"
      show-detail-icon
    >
      <template #empty>
        <div class="has-text-centered">No events found.</div>
      </template>

      <b-table-column v-slot="props" width="1px">
        <div style="width: 85px;">
          <b-button
            tag="a"
            :href="(connections ? '/update_connection/' : '/update_event/') + props.row.event_id"
          >
            <b-icon icon="pencil" type="is-primary"></b-icon>
          </b-button>
          <b-button @click.stop="confirmDeleteEvent(props.row)">
            <b-icon icon="delete" type="is-danger"></b-icon>
          </b-button>
        </div>
      </b-table-column>

      <b-table-column field="event_date" label="Date" v-slot="props" sortable>
        {{ props.row.event_date }}
      </b-table-column>

      <b-table-column
        field="event_name"
        :label="connections ? 'Coach' : 'Name'"
        v-slot="props"
        sortable
      >
        {{ props.row.event_name }}
      </b-table-column>

      <b-table-column field="event_type" label="Type" v-slot="props" sortable v-if="!connections">
        {{ props.row.event_type }}
      </b-table-column>

      <b-table-column
        field="attendees.length"
        label="Total Attendees"
        v-slot="props"
        sortable
        v-if="!connections"
      >
        {{ props.row.attendees.length }}
      </b-table-column>

      <b-table-column
        field="attendees[0].name"
        label="Coachees"
        v-slot="props"
        sortable
        v-if="connections"
      >
        {{ props.row.attendees.join(', ') }}
      </b-table-column>

      <template #detail="props">
        <article class="media">
          <div class="media-content">
            <div class="content">
              <b-button
                label="Email all attendees"
                type="is-info"
                icon-left="email"
                tag="a"
                :href="props.row.emailLink"
                target="_blank"
                class="mb-3"
              ></b-button>
              <br />
              <span v-if="!connections">
                <strong class="has-text-primary">{{
                  connections ? 'Coachees' : 'Attendees'
                }}</strong>
                <ul>
                  <li v-for="attendee in props.row.attendees" :key="attendee">{{ attendee }}</li>
                </ul>
              </span>
            </div>
          </div>
        </article>
      </template>
    </b-table>
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage, initializeFlashMessage } from './flash_message';
import moment from 'moment';

interface Event {
  // Supplied by server.
  event_id: number;
  event_name: string;
  event_date: string;
  event_type: string;
  attendees: string[];
  attendee_emails: string[];

  // Populated locally.
  emailLink: string;
  showAttendees: boolean;
}

export default Vue.extend({
  created() {
    initializeFlashMessage();
  },
  components: {
    AdbPage,
  },
  props: {
    connections: Boolean,
  },
  data() {
    // Default search from the 1st of last month to today.
    const today = moment().format('YYYY-MM-DD');
    const start = moment()
      .subtract(1, 'months')
      .startOf('month')
      .format('YYYY-MM-DD');

    return {
      search: {
        name: '',
        activist: [] as string[],
        start: start,
        end: today,
        type: 'noConnections',
      },

      showFilters: false,
      activistFilterOptions: [] as string[],
      filteredActivists: [] as string[],

      loading: false,
      events: [] as Event[],
    };
  },
  mounted() {
    this.eventListRequest();
    this.getActivistFilterOptions();
  },
  methods: {
    getActivistFilterOptions() {
      // TODO: add loading spinner for this method
      $.ajax({
        url: '/activist_names/get',
        method: 'GET',
        dataType: 'json',
        success: (data) => {
          this.activistFilterOptions = data.activist_names as string[];
        },
        error: () => {
          flashMessage('Error: could not load activist names', true);
        },
      });
    },
    getFilteredActivists(text: string) {
      this.filteredActivists = this.activistFilterOptions.filter((a: string) => {
        return a.toLowerCase().startsWith(text.toLowerCase());
      });
    },
    eventListRequest() {
      // Always show the loading screen when the button is clicked.
      this.loading = true;

      console.log(this.search);

      $.ajax({
        url: '/event/list',
        method: 'POST',
        data: {
          event_name: this.search.name,
          event_activist: this.search.activist[0],
          event_date_start: this.search.start,
          event_date_end: this.search.end,
          event_type: this.connections ? 'Connection' : this.search.type,
        },
        success: (data) => {
          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"

          // The data must be a list of events b/c it was successful.
          const events = parsed as Event[];

          // Process event JSON to make presentable.
          for (let event of events) {
            event.showAttendees = this.connections; // Show by default on connections list.
            event.emailLink =
              'https://mail.google.com/mail/?view=cm&fs=1&bcc=' +
              (event.attendee_emails || []).join(',');
          }

          if (events.length === 0) {
            flashMessage('No events from server', true);
          }

          this.events = events;
          this.loading = false;
        },
        error: () => {
          this.loading = false;
          flashMessage('Error connecting to server.', true);
        },
      });
    },

    confirmDeleteEvent(event: Event) {
      const confirmed = confirm(
        'Are you sure you want to delete the event "' +
          event.event_name +
          '"?\n\nPress OK to delete this event.',
      );

      if (confirmed) {
        $.ajax({
          url: '/event/delete',
          method: 'POST',
          data: {
            event_id: event.event_id,
          },
          success: (data) => {
            const parsed = JSON.parse(data);
            if (parsed.status === 'error') {
              flashMessage('Error: ' + parsed.message, true);
              return;
            }
            // status === "success"

            flashMessage('Deleted event ' + event.event_name);
            this.eventListRequest();
          },
          error: () => {
            flashMessage('Error connecting to server.', true);
          },
        });
      }
    },
  },
});
</script>
