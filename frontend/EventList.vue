<template>
  <adb-page
    :title="connections ? 'All Maintenance Connections' : 'Events'"
    class="event-list-content"
  >
    <form class="form-inline hidden-xs" v-on:submit.prevent="eventListRequest">
      <label for="event-name">{{ connections ? 'Connector' : 'Event Name' }}:</label>
      <input
        id="event-name"
        class="form-control filter-margin"
        style="width: 100%"
        v-model="search.name"
      />

      <label for="event-activist">{{ connections ? 'Connectee' : 'Activist' }}:</label>
      <select id="event-activist" class="filter-margin" style="width: 100%"></select>

      <label for="event-date-start">From:</label>
      <input
        id="event-date-start"
        class="form-control filter-margin"
        type="date"
        v-model="search.start"
      />

      <label for="event-date-end">To:</label>
      <input
        id="event-date-end"
        class="form-control filter-margin"
        type="date"
        v-model="search.end"
      />

      <template v-if="!connections">
        <label for="event-type">Type:</label>
        <select id="event-type" class="form-control filter-margin" v-model="search.type">
          <option value="noConnections">All</option>
          <option value="Action">Action</option>
          <option value="Circle">Circle</option>
          <option value="Community">Community</option>
          <option value="Frontline Surveillance">Frontline Surveillance</option>
          <option value="Meeting">Meeting</option>
          <option value="Meeting">Outreach</option>
          <option value="Sanctuary">Sanctuary</option>
          <option value="Meeting">Training</option>
        </select>
      </template>

      <button type="submit" id="event-date-filter" class="btn btn-primary filter-margin">
        Filter
      </button>
    </form>
    <br />

    <table class="adb-table table table-hover table-striped">
      <thead>
        <tr>
          <th class="col-xs-1"></th>
          <th class="col-xs-2">Date</th>
          <th class="col-xs-2">{{ connections ? 'Connector' : 'Name' }}</th>
          <th class="col-xs-2 hidden-xs">Type</th>
          <th class="col-xs-1 hidden-xs">Total {{ connections ? 'Connectees' : 'Attendance' }}</th>
          <th class="col-xs-4 hidden-xs">
            Attendees
            <span style="display: inline-block">
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
          <td><i>No data</i></td>
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
          <td nowrap>{{ event.event_date }}</td>
          <td class="event-name">
            <b>{{ event.event_name }}</b>
          </td>
          <td nowrap class="hidden-xs">{{ event.event_type }}</td>
          <td nowrap class="hidden-xs">{{ event.attendees.length }}</td>
          <td class="hidden-xs">
            <button class="show-attendees btn btn-link" v-on:click="toggleAttendees(event)">
              <span v-if="event.showAttendees">-</span> <span v-else>+</span> Attendees
            </button>
            <a target="_blank" class="btn btn-link" :href="event.emailLink">
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
import { initActivistSelect } from './chosen_utils';

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
        start: start.toISOString().slice(0, 10),
        end: today.toISOString().slice(0, 10),
        type: 'noConnections',
      },

      loading: false,
      events: [] as Event[],
    };
  },
  mounted() {
    initActivistSelect('#event-activist');
    this.eventListRequest();
  },
  methods: {
    eventListRequest() {
      // Always show the loading screen when the button is clicked.
      this.loading = true;

      $.ajax({
        url: '/event/list',
        method: 'POST',
        data: {
          event_name: this.search.name,
          event_activist: $('#event-activist').val(),
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
