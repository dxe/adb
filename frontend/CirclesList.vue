<template>
  <div id="app" class="main">
    <button class="btn btn-default" @click="showModal('edit-circle-modal')">
      <span class="glyphicon glyphicon-plus"></span>&nbsp;&nbsp;Add New Circle
    </button>
    &nbsp;&nbsp;&nbsp;&nbsp;
    <button
      id="showMem"
      class="btn btn-default"
      onclick="$('.wgMembers').show(); $('#showMem').hide(); $('#hideMem').show();"
    >
      <span class="glyphicon glyphicon-eye-open"></span>&nbsp;&nbsp;Show members
    </button>
    <button
      id="hideMem"
      class="btn btn-default"
      onclick="$('.wgMembers').hide(); $('#showMem').show(); $('#hideMem').hide();"
      style="display: none;"
    >
      <span class="glyphicon glyphicon-eye-close"></span>&nbsp;&nbsp;Hide members
    </button>

    <table id="working-group-list" class="adb-table table table-hover table-striped">
      <thead>
        <tr>
          <th></th>
          <th></th>
          <th>Name</th>
          <th>Email</th>
          <th>Total Members</th>
          <th>Host</th>
          <th class="wgMembers">Members</th>
        </tr>
      </thead>
      <tbody id="working-group-list-body">
        <tr v-for="(circleGroup, index) in circleGroups">
          <td>
            <button
              class="btn btn-default glyphicon glyphicon-pencil"
              @click="showModal('edit-circle-modal', circleGroup, index)"
            ></button>
          </td>
          <td>
            <dropdown>
              <button
                data-role="trigger"
                class="btn btn-default dropdown-toggle glyphicon glyphicon-option-horizontal"
                type="button"
              ></button>
              <template slot="dropdown">
                <li>
                  <a @click="showModal('delete-circle-modal', circleGroup, index)">Delete Circle</a>
                </li>
              </template>
            </dropdown>
          </td>
          <td>{{ circleGroup.name }}</td>
          <td>{{ circleGroup.email }}</td>
          <td>{{ numberOfCircleGroupMembers(circleGroup) }}</td>
          <td>
            <!-- There should only ever be one point person -->
            <template v-for="member in circleGroup.members">
              <template v-if="member.point_person">
                <p>{{ member.name }}</p>
              </template>
            </template>
          </td>
          <td>
            <ul v-for="member in circleGroup.members">
              <template v-if="!member.point_person && !member.non_member_on_mailing_list">
                <li class="wgMembers">{{ member.name }}</li>
              </template>
            </ul>
          </td>
          <td></td>
        </tr>
      </tbody>
    </table>
    <modal
      name="delete-circle-modal"
      height="auto"
      classes="no-background-color no-top"
      @opened="modalOpened"
      @closed="modalClosed"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header"><h2 class="modal-title">Delete Circle</h2></div>
          <div class="modal-body">
            <p>Are you sure you want to delete the Circle, {{ currentCircleGroup.name }}?</p>
            <p>Before you delete a Circle, you need to remove all members of that Circle.</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-danger"
              @click="confirmDeleteCircleGroupModal"
            >
              Delete Circle
            </button>
          </div>
        </div>
      </div>
    </modal>
    <modal
      name="edit-circle-modal"
      height="auto"
      classes="no-background-color no-top"
      @opened="modalOpened"
      @closed="modalClosed"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title" v-if="currentCircleGroup.id">Edit Circle</h2>
            <h2 class="modal-title" v-if="!currentCircleGroup.id">New Circle</h2>
          </div>
          <div class="modal-body">
            <form action="" id="editCircleGroupForm">
              <p>
                <label for="name">Name: </label
                ><input
                  class="form-control"
                  type="text"
                  v-model.trim="currentCircleGroup.name"
                  id="name"
                  v-focus
                />
              </p>
              <p>
                <label for="email">Email: </label
                ><input
                  class="form-control"
                  type="text"
                  v-model.trim="currentCircleGroup.email"
                  id="email"
                />
              </p>

              <!--
              <p
                <label for="type">Type: </label>
                <select id="type" class="form-control" v-model="currentCircleGroup.type">
                  <option value="circle">Circle</option>
                </select>
              </p>
            -->

              <p>
                <label for="description">Description: </label
                ><input
                  class="form-control"
                  type="text"
                  v-model.trim="currentCircleGroup.description"
                  id="description"
                />
              </p>
              <p>
                <label for="meeting_time">Meeting Day & Time: </label
                ><input
                  class="form-control"
                  type="text"
                  v-model.trim="currentCircleGroup.meeting_time"
                  id="meeting_time"
                />
              </p>
              <p>
                <label for="meeting_location">Meeting Location: </label
                ><input
                  class="form-control"
                  type="text"
                  v-model.trim="currentCircleGroup.meeting_location"
                  id="meeting_location"
                />
              </p>
              <p>
                <label for="coords">Coordinates: </label
                ><input
                  class="form-control"
                  type="text"
                  v-model.trim="currentCircleGroup.coords"
                  id="coords"
                />
              </p>
              <p>
                <label for="visible">Visible on application: </label
                ><input
                  class="form-control"
                  type="checkbox"
                  v-model.trim="currentCircleGroup.visible"
                  id="visible"
                />
              </p>

              <hr />

              <p><label for="point-person">Host: </label></p>
              <div class="select-row" v-for="(member, index) in currentCircleGroup.members">
                <template v-if="member.point_person">
                  <basic-select
                    :options="activistOptions"
                    :selected-option="memberOption(member)"
                    :extra-data="{ index: index, pointPerson: true }"
                    inheritStyle="min-width: 500px"
                    @select="onMemberSelect"
                  >
                  </basic-select>
                  <button
                    type="button"
                    class="select-row-btn btn btn-sm btn-danger"
                    @click="removeMember(index)"
                  >
                    -
                  </button>
                </template>
              </div>
              <button
                v-if="showAddPointPerson"
                type="button"
                class="btn btn-sm"
                @click="addPointPerson"
              >
                Add point person
              </button>
              <p><label for="members">Members: </label></p>
              <div class="select-row" v-for="(member, index) in currentCircleGroup.members">
                <template v-if="!member.point_person && !member.non_member_on_mailing_list">
                  <basic-select
                    :options="activistOptions"
                    :selected-option="memberOption(member)"
                    :extra-data="{ index: index }"
                    inheritStyle="min-width: 500px"
                    @select="onMemberSelect"
                  >
                  </basic-select>
                  <button
                    type="button"
                    class="select-row-btn btn btn-sm btn-danger"
                    @click="removeMember(index)"
                  >
                    -
                  </button>
                </template>
              </div>
              <button type="button" class="btn btn-sm" @click="addMember">Add member</button>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-success"
              @click="confirmEditCircleGroupModal"
            >
              Save changes
            </button>
          </div>
        </div>
      </div>
    </modal>
  </div>
</template>

<script>
import vmodal from 'vue-js-modal';
import Vue from 'vue';
import { flashMessage } from 'flash_message';
import { Dropdown } from 'uiv';
import { initActivistSelect } from 'chosen_utils';
import { focus } from 'directives/focus';
import BasicSelect from 'external/search-select/BasicSelect.vue';

Vue.use(vmodal);

export default {
  name: 'circle-list',
  methods: {
    showModal: function(modalName, circleGroup, index) {
      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      this.currentCircleGroup = $.extend(true, {}, circleGroup);

      if (index != undefined) {
        this.circleGroupIndex = index;
      } else {
        this.circleGroupIndex = -1;
      }

      this.currentModalName = modalName;
      this.$modal.show(modalName);
    },
    hideModal: function() {
      if (this.currentModalName) {
        this.$modal.hide(this.currentModalName);
      }
      this.currentModalName = '';
      this.circleGroupIndex = -1;
      this.currentCircleGroup = {};
    },
    confirmEditCircleGroupModal: function() {
      // First, check for duplicate activists because that's the most
      // likely error.
      if (this.currentCircleGroup.members) {
        var members = this.currentCircleGroup.members;
        var memberNameMap = {};
        for (var i = 0; i < members.length; i++) {
          if (members[i].name in memberNameMap) {
            flashMessage('Error: Cannot have duplicate members: ' + members[i].name, true);
            return;
          }
          memberNameMap[members[i].name] = true;
        }
      }

      // Save working group
      this.disableConfirmButton = true;

      $.ajax({
        url: '/circle/save',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(this.currentCircleGroup),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentCircleGroup.name + ' saved');

          if (this.circleGroupIndex === -1) {
            // New working group, insert at the top
            this.circleGroups = [parsed.circle].concat(this.circleGroups);
          } else {
            // We edited an existing circle, replace their row.
            Vue.set(this.circleGroups, this.circleGroupIndex, parsed.circle);
          }

          this.hideModal();
        },
        error: (err) => {
          this.disableConfirmButton = false;
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    confirmDeleteCircleGroupModal: function() {
      this.disableConfirmButton = true;

      $.ajax({
        url: '/circle/delete',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          circle_id: this.currentCircleGroup.id,
        }),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentCircleGroup.name + ' deleted');
          this.circleGroups.splice(this.circleGroupIndex, this.circleGroupIndex + 1);
          this.hideModal();
        },
        error: (err) => {
          this.disableConfirmButton = false;
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
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
    displaycircleGroupType: function(type) {
      switch (type) {
        case 'circle':
          return 'Circle';
      }
      return '';
    },
    addMember: function() {
      if (this.currentCircleGroup.members === undefined) {
        Vue.set(this.currentCircleGroup, 'members', []);
      }
      this.currentCircleGroup.members.push({ name: '' });
    },
    addPointPerson: function() {
      if (this.currentCircleGroup.members === undefined) {
        Vue.set(this.currentCircleGroup, 'members', []);
      }
      this.currentCircleGroup.members.push({ name: '', point_person: true });
    },
    addNonMember: function() {
      if (this.currentCircleGroup.members === undefined) {
        Vue.set(this.currentCircleGroup, 'members', []);
      }
      this.currentCircleGroup.members.push({ name: '', non_member_on_mailing_list: true });
    },
    removeMember: function(index) {
      this.currentCircleGroup.members.splice(index, 1);
    },
    memberOption: function(member) {
      return { text: member.name };
    },
    onMemberSelect: function(selected, extraData) {
      var index = extraData.index;
      Vue.set(this.currentCircleGroup.members, index, {
        name: selected.text,
        point_person: !!extraData.pointPerson,
        non_member_on_mailing_list: !!extraData.nonMemberOnMailingList,
      });
    },
    numberOfCircleGroupMembers: function(circleGroup) {
      if (!circleGroup.members) {
        return 0;
      }

      var count = 0;
      for (var i = 0; i < circleGroup.members.length; i++) {
        if (!circleGroup.members[i].non_member_on_mailing_list) {
          count++;
        }
      }

      return count;
    },
  },
  data() {
    return {
      currentCircleGroup: {},
      circleGroups: [],
      circleGroupIndex: -1,
      disableConfirmButton: false,
      currentModalName: '',
      activistOptions: [],
    };
  },
  computed: {
    showAddPointPerson: function() {
      if (!this.currentCircleGroup) {
        return false; // doesn't matter
      }
      if (this.currentCircleGroup && !this.currentCircleGroup.members) {
        return true;
      }

      var members = this.currentCircleGroup.members;
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
    // Get circles
    $.ajax({
      url: '/circle/list',
      method: 'POST',
      success: (data) => {
        var parsed = JSON.parse(data);
        if (parsed.status === 'error') {
          flashMessage('Error: ' + parsed.message, true);
          return;
        }
        // status === "success"
        this.circleGroups = parsed.working_groups;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
      },
    });

    // Get activists for members dropdown.
    $.ajax({
      url: '/activist_names/get',
      method: 'GET',
      success: (data) => {
        var parsed = JSON.parse(data);

        // Convert activist_names to a format usable by basic-select.
        var options = [];
        for (var i = 0; i < parsed.activist_names.length; i++) {
          options.push({ text: parsed.activist_names[i] });
        }
        this.activistOptions = options;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
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
};
</script>

<style>
.select-row {
  margin: 5px 0;
}

.select-row-btn {
  margin: 0 5px;
}

.wgMembers {
  display: none;
}
</style>
