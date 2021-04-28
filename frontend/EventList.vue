<template>
  <adb-page :title="connections ? 'All Coachings' : 'Events'">
    <b-loading :is-full-page="true" v-model="loading"></b-loading>

    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-field>
            <b-switch v-model="showFilters" type="is-primary"
            >Show filters</b-switch
            >
          </b-field>
        </div>
      </div>
    </nav>

    <nav class="level" v-if="showFilters">
      <div class="level-left">
        <div class="level-item">
          <b-field label-position="on-border" :label="connections ? 'Coach' : 'Event Name'">
            <b-input v-model="search.name" type="text">
            </b-input>
          </b-field>
        </div>
        <div class="level-item">
          <b-field label-position="on-border" :label="connections ? 'Coachees' : 'Activist'">
            <b-select v-model="search.activist">
              <option v-for="name in activistFilterOptions" :value="name" :key="name">
                {{ name }}
              </option>
            </b-select>
          </b-field>
        </div>

        <div class="level-item" v-if="!connections">
          <b-field label="Type" label-position="on-border">
            <b-select v-model="search.type">
              <option v-for="interest in [
                  {value: 'noConnections', display: 'All'},
                  {value: 'Action', display: 'Action'},
                  {value: 'Campaign Action', display: 'Campaign Action'},
                  {value: 'Community', display: 'Community'},
                  {value: 'Frontline Surveillance', display: 'Frontline Surveillance'},
                  {value: 'Meeting', display: 'Meeting'},
                  {value: 'Outreach', display: 'Outreach'},
                  {value: 'Sanctuary', display: 'Sanctuary'},
                  {value: 'Training', display: 'Training'},
                  {value: 'mpiDA', display: 'MPI: Direct Action'},
                  {value: 'mpiCOM', display: 'MPI: Community'},
                ]" :value="interest.value" :key="interest.value">
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
            <b-input v-model="search.start" type="date" ></b-input>
          </b-field>
        </div>
        <div class="level-item">
          <b-field label="To" label-position="on-border">
            <b-input v-model="search.end" type="date" ></b-input>
          </b-field>
        </div>

        <div class="level-item">
          <b-button type="is-primary" label="Filter" @click="eventListRequest"></b-button>
        </div>

      </div>
    </nav>

    <!-- TODO: build table -->

    <table class="adb-table table table-hover table-striped">
      <thead>
        <tr>
          <th class="col-xs-1"></th>
          <th class="col-xs-2">Date</th>
          <th class="col-xs-2">{{ connections ? 'Coach' : 'Name' }}</th>
          <th class="col-xs-2 hidden-xs">Type</th>
          <th class="col-xs-1 hidden-xs">Total {{ connections ? 'Coachees' : 'Attendance' }}</th>
          <th class="col-xs-4 hidden-xs">
            {{ connections ? 'Coachees' : 'Attendees' }}
            <!-- hide this span on connections page -->
            <span style="display: inline-block" v-if="!connections">
              (
              <button title="Show all attendees" class="btn btn-link" v-on:click="showAllAttendees">
                +
              </button>
              /
              <button title="Hide all attendees" class="btn btn-link" v-on:click="hideAllAttendees">
                -
              </button>
              )
            </span>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td></td>
          <td><i>Loading...</i></td>
          <td></td>
          <td hidden-xs></td>
          <td hidden-xs></td>
          <td hidden-xs></td>
        </tr>

        <tr v-if="!loading && events.length == 0">
          <td></td>
          <td><i>No events found</i></td>
          <td></td>
          <td class="hidden-xs"></td>
          <td class="hidden-xs"></td>
          <td class="hidden-xs"></td>
        </tr>

        <tr v-for="event in events" :key="event.event_id">
          <td>
            <a
              class="edit-link"
              :href="(connections ? '/update_connection/' : '/update_event/') + event.event_id"
            >
              <button class="btn btn-default glyphicon glyphicon-pencil"></button>
            </a>
            <br />
            <br />
            <div class="dropdown">
              <button
                class="btn btn-default dropdown-toggle glyphicon glyphicon-option-horizontal"
                data-toggle="dropdown"
              ></button>
              <ul class="dropdown-menu">
                <li><a v-on:click.stop="confirmDeleteEvent(event)">Delete event</a></li>
              </ul>
            </div>
          </td>
          <td nowrap>
            {{ event.event_date }}
            <span class="hidden-sm hidden-md hidden-lg hidden-xl">
              <br /><br />
              Attendance: {{ event.attendees.length }}
            </span>
          </td>
          <td>
            <b>{{ event.event_name }}</b>
          </td>
          <td nowrap class="hidden-xs">{{ connections ? 'Coaching' : event.event_type }}</td>
          <td nowrap class="hidden-xs">{{ event.attendees.length }}</td>
          <td class="hidden-xs">
            <button v-if="!connections" class="btn btn-link" v-on:click="toggleAttendees(event)">
              <span v-if="event.showAttendees">-</span> <span v-else>+</span>
              {{ connections ? 'Coachees' : 'Attendees' }}
            </button>
            <a target="_blank" class="btn btn-link" :href="event.emailLink" v-if="!connections">
              <span class="glyphicon glyphicon-envelope"></span>
            </a>
            <ul class="attendee-list" v-show="event.showAttendees">
              <li v-for="attendee in event.attendees" :key="attendee">{{ attendee }}</li>
            </ul>
          </td>
        </tr>
      </tbody>
    </table>
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage } from './flash_message';

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
  components: {
    AdbPage,
  },
  props: {
    connections: Boolean,
  },
  data() {
    // Default search from the 1st of last month to today.
    const today = new Date();
    const start = new Date(today.getFullYear(), today.getMonth() - 1, 1);

    return {
      search: {
        name: '',
        activist: '',
        start: start.toISOString().slice(0, 10),
        end: today.toISOString().slice(0, 10),
        type: 'noConnections',
      },

      showFilters: false,
      activistFilterOptions: [] as string[],

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
    eventListRequest() {
      // Always show the loading screen when the button is clicked.
      this.loading = true;

      console.log(this.search)

      $.ajax({
        url: '/event/list',
        method: 'POST',
        data: {
          event_name: this.search.name,
          event_activist: this.search.activist,
          event_date_start: this.search.start,
          event_date_end: this.search.end,
          event_type: this.connections ? 'Connection' : this.search.type,
        },
        success: (data) => {
          let parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"

          // The data must be a list of events b/c it was successful.
          let events = parsed as Event[];

          // Process event JSON to make presentable.
          for (let event of events) {
            event.showAttendees = this.connections; // Show by default on connections list.
            if (event.attendees == null) {
              event.attendees = [];
            }
            event.emailLink =
              'https://mail.google.com/mail/?view=cm&fs=1&bcc=' +
              (event.attendee_emails || []).join(',');
          }

          if (events.length == 0) {
            flashMessage('No events from server', true);
          }

          this.loading = false;
          this.events = events;
        },
        error: () => {
          flashMessage('Error connecting to server.', true);
        },
      });
    },

    confirmDeleteEvent(event: Event) {
      let confirmed = confirm(
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
            let parsed = JSON.parse(data);
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

    showAllAttendees() {
      for (let event of this.events) {
        event.showAttendees = true;
      }
    },

    hideAllAttendees() {
      for (let event of this.events) {
        event.showAttendees = false;
      }
    },

    toggleAttendees(event: Event) {
      event.showAttendees = !event.showAttendees;
    },
  },
});
</script>
