<template>

  <div id="app" class="main">
    <button class="btn btn-default" @click="showModal('edit-activist-modal')">Add New Activist</button>
    <table id="activist-list" class="adb-table table table-hover table-striped">
      <thead>
        <tr>
          <th></th>
          <th></th>
          <th @click="sortBy('name')">Name</th>
          <th @click="sortBy('email')">Email</th>
          <th @click="sortBy('phone')">Phone</th>
          <th @click="sortByDate('first_event')">First Event</th>
          <th @click="sortByDate('last_event')">Last Event</th>
          <th @click="sortByStatus('status')">Status</th>
          <th @click="sortByLevel('activist_level')">Level</th>
        </tr>
      </thead>
      <tbody id="activist-list-body">
        <tr v-for="(activist, index) in activists">
          <td><button class="btn btn-default glyphicon glyphicon-pencil" @click="showModal('edit-activist-modal', activist, index)"></button></td>
          <td>
            <dropdown>
              <button data-role="trigger" class="btn btn-default dropdown-toggle glyphicon glyphicon-option-horizontal" type="button">
              </button>
              <template slot="dropdown">
                <li><a @click="showModal('merge-activist-modal', activist, index)">Merge Activist</a></li>
                <li><a @click="showModal('hide-activist-modal', activist, index)">Hide Activist</a></li>
              </template>
            </dropdown>
          </td>
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
       name="merge-activist-modal"
       :height="650"
       classes="no-background-color"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Merge activist</h2>
          </div>
          <div class="modal-body">
            <p>Merging activists is used to combine redundant activist entries</p>
            <p>
              Merging this activist does two things:
            </p>
            <ul>
              <li>all of this activist&#39;s attendance data will be merged into the target activist</li>
              <li>this activist will be hidden</li>
            </ul>
            <p>
              Non-attendance data (e.g. email, location, etc) is <strong>NOT</strong> merged.
            </p>
            <p>Merge {{currentActivist.name}} into another activist:</p>
            <p>
              Target activist: <select id="merge-target-activist" class="filter-margin" style="min-width: 200px"></select>
            </p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button type="button" v-bind:disabled="disableConfirmButton" class="btn btn-danger" @click="confirmMergeActivistModal" v-focus>Merge activist</button>
          </div>
        </div>
      </div>
    </modal>
    <modal
       name="hide-activist-modal"
       :height="400"
       classes="no-background-color"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Hide activist</h2>
          </div>
          <div class="modal-body">
            <p>Are you sure you want to hide {{currentActivist.name}}?</p>
            <p>Hiding an activist hides them from the activist list page but does not delete any event data associated with them. If this activist is a duplicate of another activist, you should merge them instead.</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button type="button" v-bind:disabled="disableConfirmButton" class="btn btn-danger" @click="confirmHideActivistModal" v-focus>Hide activist</button>
          </div>
        </div>
      </div>
    </modal>
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
            <h2 class="modal-title" v-if="currentActivist.id">Edit activist</h2>
            <h2 class="modal-title" v-if="!currentActivist.id">New activist</h2>
          </div>
          <div class="modal-body">
            <form action="" id="editActivistForm">
              <p><label for="name">Name: </label><input class="form-control" type="text" v-model.trim="currentActivist.name" id="name" v-focus /></p>
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
            <button type="button" v-bind:disabled="disableConfirmButton" class="btn btn-success" @click="confirmEditActivistModal">Save changes</button>
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
import {Dropdown} from 'uiv';
import {initActivistSelect} from 'chosen_utils';
import {focus} from 'directives/focus';

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

var activistLevelOrder = {
  "activist" : 3,
  "core_activist" : 2,
  "organizer" : 1,
  "senior_organizer" : 0
};

export default {
  name: 'activist-list',
  methods: {
    showModal: function(modalName, activist, index) {
      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      // Make shallow copy of selected activist to prevent persisting unsaved
      // edits at the view layer when closing modal
      this.currentActivist = $.extend({}, activist);

      if (index != undefined) {
        this.activistIndex = index; // needed for updating activist
      } else {
        this.activistIndex = -1;
      }

      this.currentModalName = modalName;
      this.$modal.show(modalName);
    },
    hideModal: function() {
      if (this.currentModalName) {
        this.$modal.hide(this.currentModalName);
      }
      this.currentModalName = '';
      this.activistIndex = -1;
      this.currentActivist = {};
    },
    confirmMergeActivistModal: function() {
      var targetActivistName = $("#merge-target-activist").val();
      if (!targetActivistName) {
        flashMessage("Must choose an activist to merge into", true);
        return;
      }

      this.disableConfirmButton = true;

      $.ajax({
        url: "/activist/merge",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({
          current_activist_id: this.currentActivist.id,
          target_activist_name: targetActivistName,
        }),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          flashMessage(this.currentActivist.name + " was merged into " + targetActivistName);

          // Remove activist from list.
          this.activists = this.activists.slice(0, this.activistIndex).concat(
            this.activists.slice(this.activistIndex+1));
          this.hideModal();
        },
        error: () => {
          this.disableConfirmButton = false;

          console.warn(err.responseText);
          flashMessage("Server error: " + err.responseText, true);
        },
      });
    },
    confirmHideActivistModal: function() {
      this.disableConfirmButton = true;

      $.ajax({
        url: "/activist/hide",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({id: this.currentActivist.id}),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          flashMessage(this.currentActivist.name + " was hidden");

          // Remove activist from list.
          this.activists = this.activists.slice(0, this.activistIndex).concat(
            this.activists.slice(this.activistIndex+1));
          this.hideModal();
        },
        error: () => {
          this.disableConfirmButton = false;

          console.warn(err.responseText);
          flashMessage("Server error: " + err.responseText, true);
        },
      });
    },
    confirmEditActivistModal: function() {
      // Disable the save button when the user clicks it so they don't
      // try to save twice. Re-enable it when we get any response back
      // from the server (even an error).
      this.disableConfirmButton = true;

      $.ajax({
        url: "/activist/save",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify(this.currentActivist),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentActivist.name + " saved");

          if (this.activistIndex === -1) {
            // We're getting a new activist, insert them at the top.
            this.activists = [parsed.activist].concat(this.activists);
          } else {
            // We edited an existing activist, replace their row in
            // `activists`.
            Vue.set(this.activists, this.activistIndex, parsed.activist);
          }

          this.hideModal();
        },
        error: (err) => {
          this.disableConfirmButton = false;

          console.warn(err.responseText);
          flashMessage("Server error: " + err.responseText, true);
        },
      });
    },
    modalOpened: function() {
      // Add noscroll to body tag so it doesn't scroll while the modal
      // is shown.
      $(document.body).addClass('noscroll');
      this.disableConfirmButton = false;

      if (this.currentModalName == "merge-activist-modal") {
        // For some reason, even though this function is supposed to
        // fire after the modal is visible on the dom, the modal isn't
        // there. Vue.nextTick doesn't work for some reason, so we're
        // just going to keep calling setTimeout until the modal shows
        // up.
        var interval;
        var fn = () => {
          if ($('#merge-target-activist')[0]) {
            clearInterval(interval);
            initActivistSelect('#merge-target-activist', this.currentActivist.name);
          }
        };
        interval = setInterval(fn, 50);
      }

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
      this.sortByStatusOrLevel(field, statusOrder);
    },
    sortByLevel: function(field) {
      this.sortByStatusOrLevel(field, activistLevelOrder);
    },
    sortByStatusOrLevel: function(field, sortOrder) {
      var ascending = shouldSortByAscending(field);

      this.activists.sort(function(a, b) {
        var valueA = sortOrder[a[field]];
        var valueB = sortOrder[b[field]];

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
      disableConfirmButton: false,
      currentModalName: '',
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
  },
  components: {
    Dropdown,
  },
  directives: {
    focus,
  },
}
</script>
