<template>
  <adb-page :title="connections ? 'Coaching' : 'Event'" narrow class="event-new-content">
    <b-loading :is-full-page="true" v-model="loading"></b-loading>
    <b-loading :is-full-page="true" v-model="loadingActivists"></b-loading>
    <form action id="eventForm" v-on:change="changed('change', -1)" autocomplete="off">
      <fieldset :disabled="loading">
        <b-field :label="connections ? 'Coach' : 'Event' + ' name'">
          <b-input
            type="text"
            v-model="name"
            :icon="connections ? 'clipboard-account' : 'alphabetical-variant'"
            required
          />
        </b-field>

        <div v-if="!connections">
          <b-field label="Type">
            <b-select icon="shape" v-model="type" expanded :required="!connections">
              <option
                v-for="eventType in [
                  'Action',
                  'Campaign Action',
                  'Community',
                  'Frontline Surveillance',
                  'Meeting',
                  'Outreach',
                  'Animal Care',
                  'Training',
                ]"
                :value="eventType"
                :key="eventType"
              >
                {{ eventType }}
              </option>
            </b-select>
          </b-field>
        </div>

        <b-field label="Date" class="mt-3">
          <b-input type="date" v-model="date" expanded icon="calendar" />
          <p class="control">
            <b-button v-on:click.prevent="setDateToToday">today</b-button>
          </p>
        </b-field>

        <b-field class="my-4" v-if="shouldShowSuppressSurveyCheckbox()">
          <b-switch v-model="suppressSurvey" type="is-info"> Don't send survey </b-switch>
        </b-field>

        <b-field :label="connections ? 'Coachees' : 'Attendees'" id="attendee-rows">
          <div v-for="(attendee, index) in attendees" class="control has-icons-right">
            <input
              class="attendee-input input"
              :key="index"
              v-model="attendees[index]"
              v-on:input="changed('input', index)"
              v-on:keyup.9="changed('tab', index)"
              v-on:awesomplete-selectcomplete="changed('select', index)"
            />
            <b-icon
              v-if="attendee"
              :icon="getAttendeeStatusIcon(attendee).name"
              :type="getAttendeeStatusIcon(attendee).color"
              class="is-right"
            ></b-icon>
          </div>
        </b-field>
      </fieldset>
    </form>

    <div class="is-flex is-justify-content-space-evenly">
      <b-button
        class="is-large is-primary my-5"
        v-on:click="save"
        :disabled="saving"
        icon-left="floppy"
      >
        Save
      </b-button>
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Total attendees</p>
          <p class="title">{{ attendeeCount }}</p>
        </div>
      </div>
    </div>
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import * as Awesomplete from 'awesomplete';
import {
  flashMessage,
  initializeFlashMessage,
  setFlashMessageSuccessCookie,
} from './flash_message';
import * as dayjs from 'dayjs';

// Like Awesomplete.FILTER_CONTAINS, but internal whitespace matches anything.
function nameFilter(text: string, input: string) {
  return RegExp(Awesomplete.$.regExpEscape(input.trim()).replace(/ +/g, '.*'), 'i').test(text);
}

interface Icon {
  name: string;
  color: string;
}

export default Vue.extend({
  components: {
    AdbPage,
  },
  props: {
    connections: Boolean,
    // TODO(mdempsky): Change id to Number.
    id: String,
    chapterName: String,
  },
  data() {
    return {
      loading: false,
      loadingActivists: false,
      saving: false,

      name: '',
      date: '',
      type: '',
      attendees: [] as string[],
      suppressSurvey: this.chapterName != 'SF Bay Area',

      oldName: '',
      oldDate: '',
      oldType: '',
      oldAttendees: [] as string[],
      oldSuppressSurvey: false,

      allActivists: [] as string[],
      allActivistsSet: new Set<string>(),
      allActivistsFull: {} as { [name: string]: any },
      showIndicatorForAttendee: {} as any,
    };
  },
  computed: {
    attendeeCount() {
      let result = 0;
      for (let attendee of this.attendees) {
        if (attendee.trim() != '') {
          result++;
        }
      }
      return result;
    },
  },

  created() {
    this.updateAutocompleteNames();

    // If we're editing an existing event, fetch the data.
    if (Number(this.id) != 0) {
      this.loading = true;
      $.ajax({
        url: '/event/get/' + this.id,
        method: 'GET',
        dataType: 'json',
        success: (data) => {
          const event = data.event;
          this.name = event.event_name || '';
          this.type = event.event_type || '';
          this.date = event.event_date || '';
          this.attendees = event.attendees || [];
          this.suppressSurvey = event.suppress_survey || false;

          // ensure we show the indicators for each attendee
          for (let i = 0; i < this.attendees.length; i++) {
            this.showIndicatorForAttendee[JSON.stringify(this.attendees[i])] = true;
          }
          this.$forceUpdate();

          this.oldName = this.name;
          this.oldType = this.type;
          this.oldDate = this.date;
          this.oldAttendees = [...this.attendees];
          this.oldSuppressSurvey = this.suppressSurvey;

          this.loading = false;
          this.changed('load', -1);
        },
        error: () => {
          flashMessage('Error: could not load event', true);
        },
      });
    }

    // TODO(mdempsky): Unregister event listener when destroyed.
    window.addEventListener('beforeunload', (e) => {
      if (this.dirty()) {
        // TODO(mdempsky): Remove after figuring out Safari issue.
        console.log(
          'Event data looks dirty',
          JSON.stringify({
            new: {
              name: this.name,
              type: this.type,
              date: this.date,
              attendees: this.attendees,
              suppressSurvey: this.suppressSurvey,
            },
            old: {
              name: this.oldName,
              type: this.oldType,
              date: this.oldDate,
              attendees: this.oldAttendees,
              suppressSurvey: this.suppressSurvey,
            },
          }),
        );

        e.preventDefault();
        // MDN says returnValue is unused, but still required by Chrome.
        // https://developer.mozilla.org/en-US/docs/Web/Events/beforeunload
        e.returnValue = '';
      }
    });

    initializeFlashMessage();
  },

  updated() {
    this.$nextTick(() => {
      for (let row of $('#attendee-rows > div > input')) {
        new Awesomplete(row, {
          filter: nameFilter,
          list: this.allActivists,
          sort: false,
          // TODO(mdempsky): Update @types/awesomplete to know about tabSelect.
          tabSelect: true,
        } as Awesomplete.Options);
      }
    });
  },

  methods: {
    getAttendeeStatusIcon(attendee: string): Icon {
      if (this.checkForDuplicate(attendee)) {
        return { name: 'numeric-2-box-outline', color: 'is-danger' };
      }
      if (!this.allActivistsSet.has(attendee)) {
        return { name: 'account-plus', color: 'is-primary' };
      }
      if (this.hasEmail(attendee) && this.hasPhone(attendee)) {
        return { name: 'check', color: 'is-success' };
      }
      if (!this.hasEmail(attendee)) {
        return { name: 'email-off', color: 'is-info' };
      }
      if (!this.hasPhone(attendee)) {
        return { name: 'phone-off', color: 'is-info' };
      }
      return {} as Icon;
    },

    setDateToToday() {
      this.date = dayjs().format('YYYY-MM-DD');
    },

    dirty() {
      if (
        this.name.trim() != this.oldName ||
        (!this.connections && this.type != this.oldType) || // Connections are always "Connection"
        this.date != this.oldDate
      ) {
        return true;
      }

      const newSet = new Set<string>();
      for (let attendee of this.attendees) {
        attendee = attendee.trim();
        if (attendee != '') {
          newSet.add(attendee);
        }
      }
      const oldSet = new Set<string>();
      for (let attendee of this.oldAttendees) {
        attendee = attendee.trim();
        if (attendee != '') {
          oldSet.add(attendee);
        }
      }

      if (oldSet.size != newSet.size) {
        return true;
      }
      for (let attendee of oldSet) {
        if (!newSet.has(attendee)) {
          return true;
        }
      }

      return false;
    },

    addRows(n: number) {
      for (let i = 0; i < n; i++) {
        this.attendees.push('');
      }
    },

    changed(x: string, y: number) {
      const inputs = $('#attendee-rows input');

      // Add more rows if there are less than 3,
      // or if the last row isn't empty.
      let more = 5 - this.attendees.length;
      if (more <= 0 && this.attendees[this.attendees.length - 1].trim() != '') {
        more = 1;
      }
      if (more >= 1) {
        this.addRows(more);

        // Restore focus to where it was before.
        // TODO(mdempsky): Why is this?
        if (y >= 0) {
          inputs.get(y).focus();
        }
      }

      for (let i = 0; i < this.attendees.length; i++) {
        this.showIndicatorForAttendee[JSON.stringify(this.attendees[i])] = true;
      }
      this.$forceUpdate();

      // If event came from selecting an autocomplete suggestion,
      // then move focus to the next input.
      if (x == 'select') {
        // If the user selected an option with "tab", then the browser
        // is going to advance the focus automatically. If we set focus
        // to y+1 now, then the tab event will instead set focus to y+2.
        // By waiting until next tick, the tab event (if any) has already
        // been processed, and we're guaranteed to assign focus to y+1.
        this.$nextTick(() => {
          inputs.get(y + 1).focus();
        });

        // Awesomplete fires after modifying the input element's value,
        // but before Vue has updated the attendees array. Go ahead and
        // synchronize them now.
        // TODO(mdempsky): Figure out how to handle this properly.
        this.attendees[y] = (inputs.get(y) as HTMLInputElement).value;
      }
    },

    checkForDuplicate(value: string) {
      const found = this.attendees.filter((a) => {
        return a === value;
      });
      return found.length > 1;
    },

    save() {
      const name = this.name.trim();
      const date = this.date;
      const type = this.connections ? 'Connection' : this.type;
      const suppressSurvey = this.suppressSurvey;
      if (name === '') {
        flashMessage('Error: Please enter event name!', true);
        return;
      }
      if (date === '') {
        flashMessage('Error: Please enter date!', true);
        return;
      }
      if (type === '') {
        flashMessage('Error: Must choose event type.', true);
        return;
      }

      let attendees: string[] = [];
      let attendeesSet = new Set();
      for (let attendee of this.attendees) {
        attendee = attendee.trim();
        if (attendee != '' && !attendeesSet.has(attendee)) {
          // check that attendee has first & last name
          if (attendee.indexOf(' ') == -1) {
            flashMessage(`Error: Attendees must have first and last name: "${attendee}"`, true);
            return;
          }
          attendees.push(attendee);
          attendeesSet.add(attendee);
        }
      }

      if (attendees.length === 0) {
        flashMessage('Error: must enter attendees', true);
        return;
      }

      // TODO(mdempsky): Fix API backend so we don't have to compute diffs manually.
      const oldAttendeesSet = new Set(this.oldAttendees);
      let addedActivists = attendees.filter(function (activist) {
        return !oldAttendeesSet.has(activist);
      });
      let deletedActivists = this.oldAttendees.filter(function (activist) {
        return !attendeesSet.has(activist);
      });

      this.saving = true;
      $.ajax({
        url: this.connections ? '/connection/save' : '/event/save',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          event_id: Number(this.id),
          event_name: name,
          event_date: date,
          event_type: type,
          added_attendees: addedActivists,
          deleted_attendees: deletedActivists,
          suppress_survey: suppressSurvey,
        }),
        success: (data) => {
          this.saving = false;
          let parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }

          this.oldName = name;
          this.oldType = type;
          this.oldDate = date;
          this.oldAttendees = attendees;
          this.oldSuppressSurvey = suppressSurvey;

          // TODO(mdempsky): Remove after figuring out Safari issue.
          if (this.dirty()) {
            console.log(
              'Oops, still dirty after save!',
              JSON.stringify({
                new: {
                  name: this.name,
                  type: this.type,
                  date: this.date,
                  attendees: this.attendees,
                  suppressSurvey: this.suppressSurvey,
                },
                old: {
                  name: this.oldName,
                  type: this.oldType,
                  date: this.oldDate,
                  attendees: this.oldAttendees,
                  suppressSurvey: this.oldSuppressSurvey,
                },
              }),
            );
          }

          if (parsed.redirect) {
            // TODO(mdempsky): Implement as history rewrite.
            setFlashMessageSuccessCookie('Saved!');
            window.location = parsed.redirect;
          } else {
            flashMessage('Saved!', false);
          }

          // Saving the event may have created new activists,
          // which affects styling.
          this.updateAutocompleteNames();
        },
        error: () => {
          this.saving = false;
          flashMessage('Error, did not save data', true);
        },
      });
    },

    // TODO(mdempsky): Move into utility file.
    updateAutocompleteNames() {
      this.loadingActivists = true;
      $.ajax({
        url: '/activist/list_basic',
        method: 'GET',
        dataType: 'json',
        success: (data) => {
          const activistData = data.activists;
          // Clear current activist name array and set before re-adding
          this.allActivists.length = 0;
          this.allActivistsSet.clear();
          this.allActivistsFull = {};
          for (let activist of activistData) {
            this.allActivistsFull[activist.name] = activist;
            this.allActivists.push(activist.name);
            this.allActivistsSet.add(activist.name);
          }

          this.changed('autocomplete', -1);
          this.loadingActivists = false;
        },
        error: () => {
          flashMessage('Error: could not load activist names', true);
          this.loadingActivists = false;
        },
      });
    },
    hasEmail(name: string) {
      if (!name) {
        return;
      }

      if (!this.allActivistsFull) {
        return;
      }

      let activistFull = this.allActivistsFull[name];

      return activistFull && activistFull.email;
    },
    hasPhone(name: string) {
      if (!name) {
        return;
      }

      if (!this.allActivistsFull) {
        return;
      }

      let activistFull = this.allActivistsFull[name];

      return activistFull && activistFull.phone;
    },
    shouldShowSuppressSurveyCheckbox() {
      if (this.chapterName != 'SF Bay Area') return false;
      // only show checkbox if a survey will be sent for this event
      if (
        this.type === 'Action' ||
        this.type === 'Campaign Action' ||
        this.type === 'Community' ||
        this.type === 'Animal Care'
      )
        return true;
      if (this.name.toLowerCase().includes('chapter meeting')) return true;
      if (this.name.toLowerCase().includes('popup') && this.type === 'Community') return true;
      if (this.name.toLowerCase().includes('meetup') && this.type === 'Community') return true;
      return false;
    },
  },
});
</script>

<style>
@import url('~awesomplete/awesomplete.css');

.awesomplete {
  display: block;
}

.awesomplete mark {
  padding: 0;
}

#attendee-rows {
  width: 100%;
}
.event-new-content input {
  width: 100%;
  margin-bottom: 2px;
}
.event-new-content select {
  width: 100%;
}
.awesomplete > ul {
  font-size: 18px;
}
</style>
