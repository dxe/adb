<template>

  <div id="app" class="main">
    <table id="activist-list" class="adb-table table table-hover table-striped tablesorter">
      <thead>
        <tr>
          <th></th>
          <th>Name</th>
          <th>Email</th>
          <th>Phone</th>
          <th>First Event</th>
          <th>Last Event</th>
          <th>Status</th>
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
    <modal name="edit-activist-modal" classes="no-background-color">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Edit activist</h2>
          </div>
          <div class="modal-body">
            <label for="name">Name: </label><input class="form-control" type="text" :value="currentActivist.name" role="name" /> <br/>
            <label for="email">Email: </label><input class="form-control" type="text" :value="currentActivist.email" role="email" /> <br/>
          </div>
        </div>
      </div>
    </modal>
  </div>

</template>

<script>
import vmodal from 'vue-js-modal';
import Vue from 'vue';

Vue.use(vmodal);

export default {
  name: 'activist-list',
  methods: {
    showModal: function(activist) {
      this.currentActivist = activist;
      this.$modal.show('edit-activist-modal');
    },
    setActivists: function(activistsData) {
      this.activists = activistsData;
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
