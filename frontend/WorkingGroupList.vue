<template>
  <div id="app" class="main">
    <table id="working-group-list" class="adb-table table table-hover table-striped">
      <thead>
        <tr>
          <th></th>
          <th></th>
          <th>Name</th>
          <th>Email</th>
          <th>Type</th>
          <th>Total Members</th>
          <th>Members</th>
        </tr>
      </thead>
      <tbody id="working-group-list-body">
        <tr v-for="(workingGroup, index) in workingGroups">
          <td><button class="btn btn-default glyphicon glyphicon-pencil" @click="showModal('edit-working-group-modal', workingGroup, index)"></button></td>
          <td>
            <dropdown>
              <button data-role="trigger" class="btn btn-default dropdown-toggle glyphicon glyphicon-option-horizontal" type="button">
              </button>
              <template slot="dropdown">
                <li><a @click="showModal('delete-working-group-modal', workingGroup, index)">Delete Working Group</a></li>
              </template>
            </dropdown>
          </td>
          <td>{{workingGroup.name}}</td>
          <td>{{workingGroup.email}}</td>
          <td>{{displayWorkingGroupType(workingGroup.type)}}</td>
          <td>{{workingGroup.members ? workingGroup.members.length : 0}}</td>
          <td>
            <ul v-for="member in workingGroup.members">
              <li>{{member.name}}</li>
            </ul>
          </td>
          <td></td>
        </tr>
      </tbody>
    </table>
    <modal
       name="edit-working-group-modal"
       :height="800"
       classes="no-background-color"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title" v-if="currentWorkingGroup.id">Edit working group</h2>
            <h2 class="modal-title" v-if="!currentWorkingGroup.id">New working group</h2>
          </div>
          <div class="modal-body">
            <form action="" id="editWorkingGroupForm">
              <p><label for="name">Name: </label><input class="form-control" type="text" v-model.trim="currentWorkingGroup.name" id="name" v-focus  /></p>
              <p><label for="email">Email: </label><input class="form-control" type="text" v-model.trim="currentWorkingGroup.email" id="email" /></p>
              <p>
                <label for="type">Type: </label>
                <select id="type" class="form-control" v-model="currentWorkingGroup.type">
                  <option value="working_group">Working Group</option>
                  <option value="committee">Committee</option>
                </select>
              </p>
              <p>
                <label for="members">Members: </label>
                <div v-for="(member, index) in currentWorkingGroup.members">
                  <input v-model="member.name" /> <button type="button" class="btn btn-sm" @click="removeMember(index)"> - </button>
                </div>
                <button type="button" class="btn btn-sm" @click="addMember">Add member</button>
              </p>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button type="button" v-bind:disabled="disableConfirmButton" class="btn btn-success" @click="confirmEditWorkingGroupModal">Save changes</button>
          </div>
        </div>
      </div>
    </modal>
  </div>
</template>

<script>
import vmodal from 'vue-js-modal';
import Vue from 'vue';
import {flashMessage} from 'flash_message';
import {Dropdown} from 'uiv';
import {initActivistSelect} from 'chosen_utils';
import {focus} from 'directives/focus';

Vue.use(vmodal);

export default {
  name: 'working-group-list',
  methods: {
    showModal: function(modalName, workingGroup, index) {
      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      this.currentWorkingGroup = $.extend(true, {}, workingGroup);

      if (index != undefined) {
        this.workingGroupIndex = index;
      } else {
        this.workingGroupIndex = -1;
      }

      this.currentModalName = modalName;
      this.$modal.show(modalName);
    },
    hideModal: function() {
      if (this.currentModalName) {
        this.$modal.hide(this.currentModalName);
      }
      this.currentModalName = '';
      this.workingGroupIndex = -1;
      this.currentWorkingGroup = {};
    },
    confirmEditWorkingGroupModal: function() {
      this.disableConfirmButton = true;

      // SAMER Make ajax call.
    },
    modalOpened: function() {
      $(document.body).addClass('noscroll');
      this.disableConfirmButton = false;

      if (this.currentModalName == "edit-working-group-modal") {
        // For some reason, even though this function is supposed to
        // fire after the modal is visible on the dom, the modal isn't
        // there. Vue.nextTick doesn't work for some reason, so we're
        // just going to keep calling setTimeout until the modal shows
        // up.
        var interval;
        var fn = () => {
          if ($('#add-member')[0]) {
            clearInterval(interval);
            initActivistSelect('#add-member');
          }
        };
        interval = setInterval(fn, 50);
      }
    },
    modalClosed: function() {
      $(document.body).removeClass('noscroll');
    },
    displayWorkingGroupType: function(type) {
      switch (type) {
        case "committee":
          return "Committee";
        case "working_group":
          return "Working Group";
      }
      return "";
    },
    addMember: function() {
      this.currentWorkingGroup.members.push({name: ''});
    },
    removeMember: function(index) {
      this.currentWorkingGroup.members.splice(index, 1);
    }
  },
  data() {
    return {
      currentWorkingGroup: {},
      workingGroups: [],
      workingGroupIndex: -1,
      disableConfirmButton: false,
      currentModalName: '',
    }
  },
  created() {
    // SAMER ajax call
    //this.workingGroups...
    this.workingGroups = [
      {
        name: "Tech (Best)",
        email: "tech@directactioneverywhere.com",
        type: "working_group",
        members: [{name: "Samer Masterson"}, {name:"Cameron Bailey"}, {name:"Jake Hobbs"}],
      },
      {
        name: "Development",
        email: "development@directactioneverywhere.com",
        type: "committee",
        members: [],
      }
    ];
  },
  components: {
    Dropdown,
  },
  directives: {
    focus,
  },
}

</script>
