<template>
  <div id="app" class="main">
    <button class="btn btn-default" @click="showModal('edit-working-group-modal')">Add New Working Group</button>
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
       name="delete-working-group-modal"
       height="auto"
       classes="no-background-color no-top"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Delete working group</h2>
          </div>
          <div class="modal-body">
            <p>Are you sure you want to delete the working group {{currentWorkingGroup.name}}?</p>
            <p>Before you delete a working group, you need to remove all members of that working group.</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button type="button" v-bind:disabled="disableConfirmButton" class="btn btn-danger" @click="confirmDeleteWorkingGroupModal">Delete working group</button>
          </div>
        </div>
      </div>
    </modal>
    </modal>
    <modal
       name="edit-working-group-modal"
       height="auto"
       classes="no-background-color no-top"
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
                <label for="point-person">Point person: </label>
                <div class="select-row" v-for="(member, index) in currentWorkingGroup.members">
                  <template v-if="member.point_person">
                    <basic-select
                       :options="activistOptions"
                       :selected-option="memberOption(member)"
                       :extra-data="{index: index, pointPerson: true}"
                       inheritStyle="min-width: 500px"
                       @select="onMemberSelect">
                    </basic-select>
                    <button type="button" class="select-row-btn btn btn-sm btn-danger" @click="removeMember(index)"> - </button>
                  </template>
                </div>
                <button v-if="showAddPointPerson" type="button" class="btn btn-sm" @click="addPointPerson">Add point person</button>
              </p>
              <p>
                <label for="members">Members: </label>
                <div class="select-row" v-for="(member, index) in currentWorkingGroup.members">
                  <template v-if="!member.point_person && !member.non_member_on_mailing_list">
                    <basic-select
                       :options="activistOptions"
                       :selected-option="memberOption(member)"
                       :extra-data="{index: index}"
                       inheritStyle="min-width: 500px"
                       @select="onMemberSelect">
                    </basic-select>
                    <button type="button" class="select-row-btn btn btn-sm btn-danger" @click="removeMember(index)"> - </button>
                  </template>
                </div>
                <button type="button" class="btn btn-sm" @click="addMember">Add member</button>
              </p>
              <p>
                <label for="non-members">Non-members on the mailing list: </label>
                <div class="select-row" v-for="(member, index) in currentWorkingGroup.members">
                  <template v-if="member.non_member_on_mailing_list">
                    <basic-select
                       :options="activistOptions"
                       :selected-option="memberOption(member)"
                       :extra-data="{index: index, nonMemberOnMailingList: true}"
                       inheritStyle="min-width: 500px"
                       @select="onMemberSelect">
                    </basic-select>
                    <button type="button" class="select-row-btn btn btn-sm btn-danger" @click="removeMember(index)"> - </button>
                  </template>
                </div>
                <button type="button" class="btn btn-sm" @click="addNonMember">Add non-member to mailing list</button>
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
import BasicSelect from 'external/search-select/BasicSelect.vue';

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
      // First, check for duplicate activists because that's the most
      // likely error.
      if (this.currentWorkingGroup.members) {
        var members = this.currentWorkingGroup.members;
        var memberNameMap = {};
        for (var i = 0; i < members.length; i++) {
          if (members[i].name in memberNameMap) {
            flashMessage("Error: Cannot have duplicate members: " + members[i].name, true);
            return;
          }
          memberNameMap[members[i].name] = true;
        }
      }

      // Save working group
      this.disableConfirmButton = true;

      $.ajax({
        url: "/working_group/save",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify(this.currentWorkingGroup),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentWorkingGroup.name + " saved");

          if (this.workingGroupIndex === -1) {
            // New working group, insert at the top
            this.workingGroups = [parsed.working_group].concat(this.workingGroups);
          } else {
            // We edited an existing working group, replace their row.
            Vue.set(this.workingGroups, this.workingGroupIndex, parsed.working_group);
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
    confirmDeleteWorkingGroupModal: function() {
      this.disableConfirmButton = true;

      $.ajax({
        url: "/working_group/delete",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({
          working_group_id: this.currentWorkingGroup.id,
        }),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentWorkingGroup.name + " deleted");
          this.workingGroups.splice(this.workingGroupIndex, this.workingGroupIndex + 1);
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
      $(document.body).addClass('noscroll');
      this.disableConfirmButton = false;
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
      if (this.currentWorkingGroup.members === undefined) {
        Vue.set(this.currentWorkingGroup, 'members', []);
      }
      this.currentWorkingGroup.members.push({name: ''});
    },
    addPointPerson: function() {
      if (this.currentWorkingGroup.members === undefined) {
        Vue.set(this.currentWorkingGroup, 'members', []);
      }
      this.currentWorkingGroup.members.push({name: '', point_person: true});
    },
    addNonMember: function() {
      if (this.currentWorkingGroup.members === undefined) {
        Vue.set(this.currentWorkingGroup, 'members', []);
      }
      this.currentWorkingGroup.members.push({name: '', non_member_on_mailing_list: true});
    },
    removeMember: function(index) {
      this.currentWorkingGroup.members.splice(index, 1);
    },
    memberOption: function(member) {
      return {text: member.name};
    },
    onMemberSelect: function(selected, extraData) {
      var index = extraData.index;
      Vue.set(this.currentWorkingGroup.members, index, {
        name: selected.text,
        point_person: !!extraData.pointPerson,
        non_member_on_mailing_list: !!extraData.nonMemberOnMailingList,
      });
    },
  },
  data() {
    return {
      currentWorkingGroup: {},
      workingGroups: [],
      workingGroupIndex: -1,
      disableConfirmButton: false,
      currentModalName: '',
      activistOptions: [],
    }
  },
  computed: {
    showAddPointPerson: function() {
      if (!this.currentWorkingGroup) {
        return false; // doesn't matter
      }
      if (this.currentWorkingGroup && !this.currentWorkingGroup.members) {
        return true;
      }

      var members = this.currentWorkingGroup.members;
      var numPointPeople = 0;
      for (var i = 0; i < members.length; i++) {
        if (members[i].point_person) {
          numPointPeople++;
        }
      }

      return numPointPeople < 1;
    },
  },
  created() {
    // Get working groups
    $.ajax({
      url: "/working_group/list",
      method: "POST",
      success: (data) => {
        var parsed = JSON.parse(data);
        if (parsed.status === "error") {
          flashMessage("Error: " + parsed.message, true);
          return;
        }
        // status === "success"
        this.workingGroups = parsed.working_groups;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage("Server error: " + err.responseText, true);
      },
    });

    // Get activists for members dropdown.
    $.ajax({
      url: "/activist_names/get",
      method: "GET",
      success: (data) => {
        var parsed = JSON.parse(data);

        // Convert activist_names to a format usable by basic-select.
        var options = [];
        for (var i = 0; i < parsed.activist_names.length; i++) {
          options.push({text: parsed.activist_names[i]});
        }
        this.activistOptions = options;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage("Server error: " + err.responseText, true);
      },
    });
  },
  components: {
    Dropdown,
    BasicSelect,
  },
  directives: {
    focus,
  },
}

</script>

<style>
  .select-row {
    margin: 5px 0;
  }

  .select-row-btn {
    margin: 0 5px;
  }
</style>
