<template>

  <div id="app" class="main">
    <table id="activist-list" class="adb-table table table-hover table-striped tablesorter">
      <thead>
        <tr>
          <th></th>
          <th @click="sortBy('name')">Name</th>
          <th @click="sortBy('email')">Email</th>
          <th @click="sortBy('phone')">Phone</th>
          <th @click="sortByDate('first_event')">First Event</th>
          <th @click="sortByDate('last_event')">Last Event</th>
          <th @click="sortByStatus('status')">Status</th>
          <th>Level</th>
        </tr>
      </thead>
      <tbody id="activist-list-body">
        <tr v-for="(activist, index) in activists">
          <td><button class="btn btn-default glyphicon glyphicon-pencil" @click="showModal(activist, index)"></button></td>
          <td>{{activist.name}}</td>
          <td>{{activist.email}}</td>
          <td>{{activist.phone}}</td>
          <td>{{activist.first_event}}</td>
          <td>{{activist.last_event}}</td>
          <td>{{activist.status}}</td>
          <td>{{displayActivistLevel(activist.activist_level)}}</td>
        </tr>
      </tbody>
    </table>
    <modal
       name="edit-activist-modal"
       :height="830"
       classes="no-background-color"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Edit activist</h2>
          </div>
          <div class="modal-body">
            <form action="" id="editActivistForm">
              <p><label for="name">Name: </label><input class="form-control" type="text" v-model.trim="currentActivist.name" id="name" /></p>
              <p><label for="email">Email: </label><input class="form-control" type="text" v-model.trim="currentActivist.email" id="email" /></p>
              <p><label for="chapter">Chapter: </label><input class="form-control" type="text" v-model.trim="currentActivist.chapter" id="chapter"></p>
              <p><label for="phone">Phone: </label><input class="form-control" type="text" v-model.trim="currentActivist.phone" id="phone"></p>
              <p><label for="location">Location: </label><input class="form-control" type="text" v-model.trim="currentActivist.location" id="location"></p>
              <p><label for="facebook">Facebook: </label><input class="form-control" type="text" v-model.trim="currentActivist.facebook" id="facebook"></p>
              <p><label for="challenge_level">Activist Level: </label>
                <select id="activist_level" class="form-control" v-model="currentActivist.activist_level">
                  <option value="activist">Activist</option>
                  <option value="core_activist">Core Activist</option>
                  <option value="organizer">Organizer</option>
                  <option value="senior_organizer">Senior Organizer</option>
                </select>
              </p>
              <p><label for="core">Core/Staff:&nbsp;</label><input class="form-check-input" type="checkbox" v-model="currentActivist.core_staff" :true-value="1" :false-value="0" id="core"></p>
              <p><label for="exclude">Exclude from Leaderboard:&nbsp;</label><input class="form-check-input" type="checkbox" v-model="currentActivist.exclude_from_leaderboard" :true-value="1" :false-value="0" id="exclude"></p>
              <p><label for="pledge">Liberation Pledge:&nbsp;</label><input class="form-check-input" type="checkbox" v-model="currentActivist.liberation_pledge" :true-value="1" :false-value="0" id="pledge"></p>
              <p><label for="globalteam">Global Team Member:&nbsp;</label><input class="form-check-input" type="checkbox" v-model="currentActivist.global_team_member" :true-value="1" :false-value="0" id="globalteam"></p>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button type="button" class="btn btn-success" @click="saveModal">Save changes</button>
          </div>
        </div>
      </div>
    </modal>
  </div>

</template>

<script>
// Library from here: https://github.com/euvl/vue-js-modal
import vmodal from 'vue-js-modal';
import Vue from 'vue';
import {flashMessage} from 'flash_message';

Vue.use(vmodal);

// Store the data of the previous sort.
var previousSortData = {
  field: null,
  ascending: null,
};

// Uses previousSortData to determine whether the next sort should be
// ascending.
//
// If sortByDate is true, then the default is to sort by descending.
// Otherwise, the default is to sort by ascending.
function shouldSortByAscending(field, sortByDate) {
  if (field == previousSortData.field) {
    return !previousSortData.ascending;
  }

  if (sortByDate) {
    return false;
  }
  return true;
}

// Call this after every sort.
function setPreviousSortData(field, ascending) {
  previousSortData.field = field;
  previousSortData.ascending = ascending;
}


// Must be kept in sync with the list in model/model.go
var statusOrder = {
  "Current": 1,
  "New": 2,
  "Former": 3,
  "No attendance": 4,
};

export default {
  name: 'activist-list',
  methods: {
    showModal: function(activist, index) {
      // Make shallow copy of selected activist to prevent persisting unsaved
      // edits at the view layer when closing modal
      this.currentActivist = $.extend({}, activist);
      this.activistIndex = index; // needed for updating activist
      this.$modal.show('edit-activist-modal');
    },
    hideModal: function() {
      this.$modal.hide('edit-activist-modal');
      this.currentActivist = {};
    },
    saveModal: function() {
      $.ajax({
        url: "/activist/save",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify(this.currentActivist),
        success: function(data) {
          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          // status === "success""
          Vue.set(this.activists, this.activistIndex, parsed.activist);
          flashMessage("Saved!", false);
        }.bind(this),
        error : function() {
          flashMessage("Error Connecting to Server", true);
        }
      });
    },
    modalOpened: function() {
      // Add noscroll to body tag so it doesn't scroll while the modal
      // is shown.
      $(document.body).addClass('noscroll');
    },
    modalClosed: function() {
      // Allow body to scroll after modal is closed.
      $(document.body).removeClass('noscroll');
    },
    setActivists: function(activistsData) {
      this.activists = activistsData;
    },
    sortBy: function(field) {
      var ascending = shouldSortByAscending(field);

      this.activists.sort(function(a,b) {
        var order = (a[field].toLowerCase() < b[field].toLowerCase()) ? -1 : 1;
        if (ascending) {
          return order;
        }
        return -1 * order;
      });

      setPreviousSortData(field, ascending);
    },
    sortByDate: function(field) {
      var ascending = shouldSortByAscending(field, true);

      this.activists.sort(function(a, b) {
        // Always sort empty values to the bottom, no matter the
        // order.
        if (!a[field]) {
          return 1;
        }
        if (!b[field]) {
          return -1;
        }

        var valueA = new Date(a[field]).getTime();
        var valueB = new Date(b[field]).getTime();

        var order = (valueA < valueB) ? -1 : 1;

        if (ascending) {
          return order;
        }
        return -1 * order;
      });

      setPreviousSortData(field, ascending);
    },
    sortByStatus: function(field) {
      var ascending = shouldSortByAscending(field);

      this.activists.sort(function(a, b) {
        var valueA = statusOrder[a[field]];
        var valueB = statusOrder[b[field]];

        var order = (valueA < valueB) ? -1 : 1;

        if (ascending) {
          return order;
        }
        return -1 * order;
      });

      setPreviousSortData(field, ascending);
    },
    displayActivistLevel: function(activistLevel) {
      var displayValue = "";

      switch(activistLevel) {
        case "activist":
          displayValue = "Activist";
          break;
        case "core_activist":
          displayValue = "Core Activist";
          break;
        case "organizer":
          displayValue = "Organizer";
          break;
        case "senior_organizer":
          displayValue = "Senior Organizer"
          break;
      }

      return displayValue;
    }
  },
  data() {
    return {
      currentActivist: {},
      activists: [],
      activistIndex: -1,
    };
  },
  created() {
    $.ajax({
      url: "/activist/list",
      success: function(data) {
        var parsed = JSON.parse(data);
        if (parsed.status === "error") {
          flashMessage("Error: " + parsed.message, true);
          return;
        }
        // status === "success"

        this.setActivists(parsed);
      }.bind(this),
      error: function() {
        flashMessage("Error connecting to server.", true);
      },
    });
  }
}
</script>

<style>

  .no-background-color {
    background-color: transparent !important;
  }

  .noscroll {
    overflow: hidden;
  }

  .v--modal-overlay {
    overflow-y: scroll;
    z-index: 1040 !important;
  }

</style>
