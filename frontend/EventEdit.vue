<template>
  <div class="body-wrapper event-new-content">
    <!-- TODO(mdempsky): Is this okay? Should this go somewhere else? -->
    <link rel="stylesheet" href="/static/external/awesomplete/awesomplete.css">

    <div class="title">
      <h1>{{connections ? "Maintenance Connection" : "Event"}}</h1>
    </div>
    <br>

    <div class="main">
      <form action id="eventForm" v-on:change="changed('change', -1)">
        <fieldset :disabled="loading">
          <label for="eventName" id="nameLabel">
            <b>{{connections ? "Connector" : "Event"}} name</b>
            <br>
          </label>
          <input id="eventName" class="form-control" v-model="name">
          <br>

          <template v-if="!connections">
            <label for="eventType">
              <b>Event type</b>
              <br>
            </label>
            <select id="eventType" class="form-control" v-model="type">
              <option disabled selected value>-- select an option --</option>
              <option value="Working Group">Working Group</option>
              <option value="Protest">Protest</option>
              <option value="Community">Community</option>
              <option value="Outreach">Outreach</option>
              <option value="Key Event">Key Event</option>
              <option value="Sanctuary">Sanctuary (Rescue/Work Day)</option>
            </select>
            <br>
          </template>

          <label for="eventDate">
            <b>{{connections ? "Connection" : "Event"}} date</b>
            <button
              class="btn btn-xs btn-primary"
              style="margin: 0px 10px"
              v-on:click.prevent="setDateToToday"
            >today</button>
            <br>
          </label>
          <input id="eventDate" class="form-control" type="date" v-model="date">
          <br>

          <label for="attendee1" id="attendeeLabel">
            <b>{{connections ? "Connectees" : "Attendees"}}</b>
            <br>
          </label>
          <div id="attendee-rows">
            <input
              class="attendee-input form-control"
              v-for="(attendee, index) in attendees"
              :key="index"
              v-model="attendees[index]"
              v-on:input="changed('input', index)"
              v-on:awesomplete-selectcomplete="changed('select', index)"
            >
          </div>

          <br>

          <label for="attendeeTotal">
            <b>Total attendance:</b>
          </label>
          <span id="attendeeTotal">{{attendeeCount}}</span>
          <br>
        </fieldset>
      </form>
      <br>
      <center>
        <button
          class="btn btn-success btn-lg"
          id="submit-button"
          v-on:click="save"
          :disabled="saving"
        >
          <span>Save {{connections ? "connection" : "event"}}</span>
        </button>
      </center>
      <br>
    </div>
  </div>
</template>

<script>
import * as Awesomplete from "external/awesomplete";
import { flashMessage, setFlashMessageSuccessCookie } from "flash_message";

export default {
  props: {
    connections: Boolean,
    // TODO(mdempsky): Change id to Number.
    id: String
  },
  data() {
    return {
      loading: false,
      saving: false,

      name: "",
      date: "",
      type: "",
      attendees: [],

      oldName: "",
      oldDate: "",
      oldType: "",
      oldAttendees: [],

      allActivists: [],
      allActivistsSet: new Set()
    };
  },
  computed: {
    attendeeCount() {
      let result = 0;
      for (let attendee of this.attendees) {
        if (attendee.trim() != "") {
          result++;
        }
      }
      return result;
    }
  },

  created() {
    this.updateAutocompleteNames();

    // If we're editing an existing event, fetch the data.
    if (Number(this.id) != 0) {
      this.loading = true;
      $.ajax({
        url: "/event/get/" + this.id,
        method: "GET",
        dataType: "json",
        success: data => {
          const event = data.event;
          this.name = event.event_name || "";
          this.type = event.event_type || "";
          this.date = event.event_date || "";
          this.attendees = event.attendees || [];

          this.oldName = this.name;
          this.oldType = this.type;
          this.oldDate = this.date;
          this.oldAttendees = [...this.attendees];

          this.loading = false;
          this.changed("load", -1);
        },
        error: () => {
          flashMessage("Error: could not load event", true);
        }
      });
    }

    // TODO(mdempsky): Unregister event listener when destroyed.
    window.addEventListener("beforeunload", e => {
      if (this.dirty()) {
        e.preventDefault();
        // MDN says returnValue is unused, but still required by Chrome.
        // https://developer.mozilla.org/en-US/docs/Web/Events/beforeunload
        e.returnValue = "";
      }
    });
  },

  updated() {
    this.$nextTick(() => {
      for (let row of $("#attendee-rows > input.attendee-input")) {
        new Awesomplete(row, {
          list: this.allActivists,
          sort: false
        });
      }
    });
  },

  methods: {
    setDateToToday(event) {
      const today = new Date();
      this.date = today.toISOString().slice(0, 10);
    },

    dirty() {
      if (
        this.name.trim() != this.oldName ||
        (!this.connections && this.type != this.oldType) || // Connections are always "Connection"
        this.date != this.oldDate
      ) {
        return true;
      }

      var newSet = new Set();
      for (let attendee of this.attendees) {
        attendee = attendee.trim();
        if (attendee != "") {
          newSet.add(attendee);
        }
      }
      var oldSet = new Set();
      for (let attendee of this.oldAttendees) {
        attendee = attendee.trim();
        if (attendee != "") {
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

    addRows(n) {
      for (let i = 0; i < n; i++) {
        this.attendees.push("");
      }
    },

    changed(x, y) {
      const inputs = $("#attendee-rows input.attendee-input");

      // Add more rows if there are less than 5,
      // or if the last row isn't empty.
      let more = 5 - this.attendees.length;
      if (more <= 0 && this.attendees[this.attendees.length - 1].trim() != "") {
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

      // If event came from selecting an autocomplete suggestion, then move focus to the next input.
      if (x == "select") {
        inputs.get(y + 1).focus();

        // Awesomplete fires after modifying the input element's value,
        // but before Vue has updated the attendees array. Go ahead and
        // synchronize them now.
        // TODO(mdempsky): Figure out how to handle this properly.
        this.attendees[y] = inputs.get(y).value;
      }

      // Update attendee warnings.
      // TODO(mdempsky): Let vue handle this.
      let seen = new Set();
      for (let i = 0; i < this.attendees.length; i++) {
        const name = this.attendees[i].trim();

        let warning = "";
        if (name != "") {
          if (!this.allActivistsSet.has(name)) {
            warning = "unknown";
          } else if (seen.has(name)) {
            warning = "duplicate";
          } else {
            seen.add(name);
          }
        }

        if (i < inputs.length) {
          inputs.get(i).dataset.warning = warning;
        }
      }
    },

    save() {
      const name = this.name.trim();
      const date = this.date;
      const type = this.connections ? "Connection" : this.type;
      if (name === "") {
        flashMessage("Error: Please enter event name!", true);
        return;
      }
      if (date === "") {
        flashMessage("Error: Please enter date!", true);
        return;
      }
      if (type === "") {
        flashMessage("Error: Must choose event type.", true);
        return;
      }

      let attendees = [];
      let attendeesSet = new Set();
      for (let attendee of this.attendees) {
        attendee = attendee.trim();
        if (attendee != "" && !attendeesSet.has(attendee)) {
          attendees.push(attendee);
          attendeesSet.add(attendee);
        }
      }

      if (attendees.length === 0) {
        flashMessage("Error: must enter attendees", true);
        return;
      }

      // TODO(mdempsky): Fix API backend so we don't have to compute diffs manually.
      const oldAttendeesSet = new Set(this.oldAttendees);
      let addedActivists = attendees.filter(function(activist) {
        return !oldAttendeesSet.has(activist);
      });
      let deletedActivists = this.oldAttendees.filter(function(activist) {
        return !attendeesSet.has(activist);
      });

      this.saving = true;
      $.ajax({
        url: this.connections ? "/connection/save" : "/event/save",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({
          event_id: Number(this.id),
          event_name: name,
          event_date: date,
          event_type: type,
          added_attendees: addedActivists,
          deleted_attendees: deletedActivists
        }),
        success: data => {
          this.saving = false;
          let parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }

          this.oldName = name;
          this.oldType = type;
          this.oldDate = date;
          this.oldAttendees = attendees;

          if (parsed.redirect) {
            // TODO(mdempsky): Implement as history rewrite.
            setFlashMessageSuccessCookie("Saved!");
            window.location = parsed.redirect;
          } else {
            flashMessage("Saved!", false);
          }

          // Saving the event may have created new activists,
          // which affects styling.
          this.updateAutocompleteNames();
        },
        error: () => {
          this.saving = false;
          flashMessage("Error, did not save data", true);
        }
      });
    },

    // TODO(mdempsky): Move into utility file.
    updateAutocompleteNames() {
      $.ajax({
        url: "/activist_names/get",
        method: "GET",
        dataType: "json",
        success: data => {
          var activistNames = data.activist_names;
          // Clear current activist name array and set before re-adding
          this.allActivists.length = 0;
          this.allActivistsSet.clear();
          for (let name of data.activist_names) {
            this.allActivists.push(name);
            this.allActivistsSet.add(name);
          }
          this.changed("autocomplete", -1);
        },
        error: () => {
          flashMessage("Error: could not load activist names", true);
        }
      });
    }
  }
};
</script>

<style>
.attendee-input[data-warning="duplicate"] {
  border: 2px solid red;
}

.attendee-input[data-warning="unknown"] {
  border: 2px solid yellow;
}
</style>
