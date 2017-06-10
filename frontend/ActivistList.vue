<template>

  <div id="app" class="main">
    <table id="activist-list" class="adb-table table table-hover table-striped tablesorter">
      <thead>
        <tr>
          <th></th>
          <th @click="sortBy('name')">Name</th>
          <th @click="sortBy('email')">Email</th>
          <th @click="sortBy('phone')">Phone</th>
          <th @click="sortBy('first_event')">First Event</th>
          <th @click="sortBy('last_event')">Last Event</th>
          <th @click="sortBy('status')">Status</th>
        </tr>
      </thead>
      <tbody id="activist-list-body">
        <tr v-for="activist in activists" @click="showModal(activist)">
          <td><button class="btn btn-default glyphicon glyphicon-pencil"></button></td>
          <td>{{activist.name}}</td>
          <td>{{activist.email}}</td>
          <td>{{activist.phone}}</td>
          <td>{{activist.first_event}}</td>
          <td>{{activist.last_event}}</td>
          <td>{{activist.status}}</td>
        </tr>
      </tbody>
    </table>
    <modal name="edit-activist-modal" :height="800" classes="no-background-color">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Edit activist</h2>
          </div>
          <div class="modal-body">
            <p><label for="name">Name: </label><input class="form-control" type="text" :value="currentActivist.name" id="name" /></p>
            <p><label for="email">Email: </label><input class="form-control" type="text" :value="currentActivist.email" id="email" /></p>
            <p><label for="chapter">Chapter: </label><input class="form-control" type="text" :value="currentActivist.chapter" id="chapter"></p>
            <p><label for="phone">Phone: </label><input class="form-control" type="text" :value="currentActivist.phone" id="phone"></p>
            <p><label for="location">Location: </label><input class="form-control" type="text" :value="currentActivist.location" id="location"></p>
            <p><label for="facebook">Facebook: </label><input class="form-control" type="text" :value="currentActivist.facebook" id="facebook"></p>
            <p><label for="core">Core/Staff:&nbsp;</label><input class="form-check-input" type="checkbox" id="core"></p>
            <p><label for="exclude">Exclude from Leaderboard:&nbsp;</label><input class="form-check-input" type="checkbox" id="exclude"></p>
            <p><label for="pledge">Liberation Pledge:&nbsp;</label><input class="form-check-input" type="checkbox" id="pledge"></p>
            <p><label for="globalteam">Global Team Member:&nbsp;</label><input class="form-check-input" type="checkbox" id="globalteam"></p>
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
      console.log('not implemented');
    },
    setActivists: function(activistsData) {
      this.activists = activistsData;
    },
    sortBy: function(field) {
      var ascending;
      if (field == previousSortData.field) {
        ascending = !previousSortData.ascending;
      } else {
        ascending = true;
      }

      this.activists.sort(function(a,b) {
        if (ascending) {
          return (a[field].toLowerCase() < b[field].toLowerCase()) ? -1 : 1;
        } else {
          return (a[field].toLowerCase() > b[field].toLowerCase()) ? -1 : 1;
        }
      });
      previousSortData.field = field;
      previousSortData.ascending = ascending;
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

</style>
