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
        </tr>
      </thead>
      <tbody id="activist-list-body">
        <tr v-for="activist in activists">
          <td><button class="btn btn-default glyphicon glyphicon-pencil" @click="showModal(activist)"></button></td>
          <td>{{activist.name}}</td>
          <td>{{activist.email}}</td>
          <td>{{activist.phone}}</td>
          <td>{{activist.first_event}}</td>
          <td>{{activist.last_event}}</td>
          <td>{{activist.status}}</td>
        </tr>
      </tbody>
    </table>
    <modal
       name="edit-activist-modal"
       :height="800"
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
              <p><label for="name">Name: </label><input class="form-control" type="text" :value="currentActivist.name" id="name" /></p>
              <p><label for="email">Email: </label><input class="form-control" type="text" :value="currentActivist.email" id="email" /></p>
              <p><label for="chapter">Chapter: </label><input class="form-control" type="text" :value="currentActivist.chapter" id="chapter"></p>
              <p><label for="phone">Phone: </label><input class="form-control" type="text" :value="currentActivist.phone" id="phone"></p>
              <p><label for="location">Location: </label><input class="form-control" type="text" :value="currentActivist.location" id="location"></p>
              <p><label for="facebook">Facebook: </label><input class="form-control" type="text" :value="currentActivist.facebook" id="facebook"></p>
              <p><label for="core">Core/Staff:&nbsp;</label><input class="form-check-input" type="checkbox" :checked="currentActivist.core_staff" id="core"></p>
              <p><label for="exclude">Exclude from Leaderboard:&nbsp;</label><input class="form-check-input" type="checkbox" :checked="currentActivist.exclude_from_leaderboard" id="exclude"></p>
              <p><label for="pledge">Liberation Pledge:&nbsp;</label><input class="form-check-input" type="checkbox" :checked="currentActivist.liberation_pledge" id="pledge"></p>
              <p><label for="globalteam">Global Team Member:&nbsp;</label><input class="form-check-input" type="checkbox" :checked="currentActivist.global_team_member" id="globalteam"></p>
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

function saveActivistEdits(activistInfo) {
  console.log(activistInfo);
  $.ajax({
    url: "/activist/save",
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify(activistInfo),
    success: function(data) {
      // TODO finish filling this in
        var parsed = JSON.parse(data);
        if (parsed.status === "error") {
          flashMessage("Error: " + parsed.message, true);
          return;
        }
        flashMessage("Saved!", false); // not working yet
        // refresh html
        // status === "success"
      
    },
    error: function() {
      // TODO Handle this appropriately
      console.log("Internal Error");
    }
  });

}


export default {
  name: 'activist-list',
  methods: {
    showModal: function(activist) {
      this.currentActivist = activist;
      this.$modal.show('edit-activist-modal');
    },
    hideModal: function() {
      this.currentActivist = null;
      this.$modal.hide('edit-activist-modal');
    },
    saveModal: function() {
      // TODO Grab checkbox information
      var activistInfo = {
        "id" : this.currentActivist.id,
        "name" : $("#name")[0].value,
        "email" : $("#email")[0].value,
        "chapter" : $("#chapter")[0].value,
        "phone" : $("#phone")[0].value,
        "location" : $("#location")[0].value,
        "facebook" : $("#facebook")[0].value,
        "core_staff" : $("#core")[0].checked ? 1 : 0,
        "exclude_from_leaderboard" : $("#exclude")[0].checked ? 1 : 0,
        "liberation_pledge" : $("#pledge")[0].checked ? 1 : 0,
        "global_team_member" : $("#globalteam")[0].checked ? 1 : 0,
      };
      saveActivistEdits(activistInfo);
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
  },
  data() {
    return {
      currentActivist: {},
      activists: [],
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
