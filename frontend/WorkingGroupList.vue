<template>
  <adb-page title="Working Groups" class="body-wrapper-extra-wide">
    <b-loading :is-full-page="true" v-model="loadingActivists"></b-loading>
    <b-loading :is-full-page="true" v-model="loadingOrganizers"></b-loading>
    <b-loading :is-full-page="true" v-model="loadingWorkingGroups"></b-loading>
    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-button
            icon-left="plus"
            type="is-primary"
            @click="showModal('edit-working-group-modal')"
          >
            New Working Group
          </b-button>
        </div>
      </div>
      <div class="level-right">
        <div class="level-item">
          <b-button @click="toggleMembers" :icon-left="membersVisible ? 'eye-off' : 'eye'">
            {{ membersVisible ? 'Hide' : 'Show' }} members
          </b-button>
        </div>
      </div>
    </nav>

    <b-table :data="workingGroups" striped hoverable default-sort="name">
      <b-table-column v-slot="props" width="1px">
        <div style="width: 85px;">
          <b-button @click="showModal('edit-working-group-modal', props.row)">
            <b-icon icon="pencil" type="is-primary"></b-icon>
          </b-button>
          <b-button @click="showModal('delete-working-group-modal', props.row)">
            <b-icon icon="delete" type="is-danger"></b-icon>
          </b-button>
        </div>
      </b-table-column>

      <b-table-column field="name" label="Name" v-slot="props" sortable>
        {{ props.row.name }}
      </b-table-column>

      <b-table-column field="email" label="Email" v-slot="props" sortable>
        {{ props.row.email }}
      </b-table-column>

      <b-table-column field="type" label="Type" v-slot="props" sortable>
        {{ displayWorkingGroupType(props.row.type) }}
      </b-table-column>

      <b-table-column field="members" label="Point Person" v-slot="props">
        <!-- There should only ever be one point person -->
        <!-- TODO: calculate this somewhere else so column can be sortable -->
        <template v-for="member in props.row.members">
          <template v-if="member.point_person">
            {{ member.name }}
          </template>
        </template>
      </b-table-column>

      <b-table-column label="Total Members" v-slot="props" v-if="!membersVisible">
        <!-- TODO: calculate this somewhere else so column can be sortable -->
        {{ numberOfWorkingGroupMembers(props.row) }}
      </b-table-column>

      <b-table-column label="Members" field="members" v-slot="props" v-if="membersVisible">
        <ul v-for="member in props.row.members">
          <template v-if="!member.point_person">
            <li>{{ member.name }}</li>
          </template>
        </ul>
      </b-table-column>
    </b-table>

    <b-modal
      :active="currentModalName === 'delete-working-group-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">Delete working group</p>
        </header>
        <section class="modal-card-body">
          <p>
            Are you sure you want to delete <strong>{{ currentWorkingGroup.name }}</strong
            >?
          </p>
          <b-message type="is-warning" has-icon class="mt-3">
            Before deleting a working group, be sure to remove all members of that group.
          </b-message>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Cancel" @click="hideModal" />
          <b-button
            label="Delete"
            type="is-danger"
            :disabled="disableConfirmButton"
            @click="confirmDeleteWorkingGroupModal"
          />
        </footer>
      </div>
    </b-modal>

    <b-modal
      :active="currentModalName === 'edit-working-group-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="false"
      :width="400"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">
            {{ currentWorkingGroup.id ? 'Edit' : 'New' }}
            Working Group
          </p>
        </header>
        <section class="modal-card-body">
          <b-field label="Name" label-position="on-border">
            <b-input
              type="text"
              v-model.trim="currentWorkingGroup.name"
              icon="hammer-screwdriver"
              required
            ></b-input>
          </b-field>

          <b-field label="Email" label-position="on-border">
            <b-input
              type="email"
              v-model.trim="currentWorkingGroup.email"
              icon="email"
              required
            ></b-input>
          </b-field>

          <b-field label="Type" label-position="on-border">
            <b-select v-model="currentWorkingGroup.type" required expanded icon="shape">
              <option
                v-for="type in [
                  { name: 'working_group', display: 'Working Group' },
                  { name: 'committee', display: 'Committee' },
                ]"
                :value="type.name"
                :key="type.name"
              >
                {{ type.display }}
              </option>
            </b-select>
          </b-field>

          <b-field label="Description" label-position="on-border">
            <b-input
              type="text"
              v-model.trim="currentWorkingGroup.description"
              icon="text-box"
            ></b-input>
          </b-field>

          <b-field label="Meeting Day & Time" label-position="on-border">
            <b-input
              type="text"
              v-model.trim="currentWorkingGroup.meeting_time"
              icon="calendar-blank"
            ></b-input>
          </b-field>

          <b-field label="Meeting Location" label-position="on-border">
            <b-input
              type="text"
              v-model.trim="currentWorkingGroup.meeting_location"
              icon="map-marker"
            ></b-input>
          </b-field>

          <b-field>
            <b-switch v-model="currentWorkingGroup.visible" type="is-success">Visible</b-switch>
          </b-field>

          <b-field label="Point Person">
            <b-taginput
              v-model="currentPointPerson"
              :data="filteredOrganizers"
              autocomplete
              :allow-new="false"
              icon="crown"
              placeholder="Search by name..."
              @typing="getFilteredOrganizers"
              maxtags="1"
              type="is-info"
              dropdown-position="top"
            ></b-taginput>
          </b-field>

          <b-field label="Members">
            <b-taginput
              v-model="currentMembers"
              :data="filteredOrganizers"
              autocomplete
              :allow-new="false"
              icon="account-multiple"
              @typing="getFilteredOrganizers"
              type="is-info"
              dropdown-position="top"
            ></b-taginput>
          </b-field>

          <b-field label="Non-members on Mailing List">
            <b-taginput
              v-model="currentNonMembers"
              :data="filteredActivists"
              autocomplete
              :allow-new="false"
              icon="account-multiple"
              @typing="getFilteredActivists"
              type="is-info"
              dropdown-position="top"
            ></b-taginput>
          </b-field>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Cancel" icon-left="cancel" @click="hideModal" />
          <b-button
            label="Save"
            icon-left="floppy"
            type="is-primary"
            :disabled="disableConfirmButton"
            @click="confirmEditWorkingGroupModal"
          />
        </footer>
      </div>
    </b-modal>
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage, initializeFlashMessage } from './flash_message';
import { focus } from './directives/focus';

interface Activist {
  name: string;
  point_person?: boolean;
  non_member_on_mailing_list?: boolean;
}

interface WorkingGroup {
  id: number;
  name: string;
  email: string;
  type: string;
  visible: boolean;
  description: string;
  meeting_time: string;
  meeting_location: string;
  members: Activist[];
}

export default Vue.extend({
  name: 'working-group-list',
  methods: {
    toggleMembers() {
      this.membersVisible = !this.membersVisible;
    },
    getFilteredActivists(text: string) {
      this.filteredActivists = this.allActivists.filter((a: string) => {
        return a.toLowerCase().startsWith(text.toLowerCase());
      });
    },
    getFilteredOrganizers(text: string) {
      this.filteredOrganizers = this.allOrganizers.filter((a: string) => {
        return a.toLowerCase().startsWith(text.toLowerCase());
      });
    },
    numberOfWorkingGroupMembers(wg: WorkingGroup) {
      if (!wg.members) {
        return 0;
      }

      let count = 0;
      for (let i = 0; i < wg.members.length; i++) {
        if (!wg.members[i].non_member_on_mailing_list) {
          count++;
        }
      }

      return count;
    },
    isMobile() {
      return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
        navigator.userAgent,
      );
    },
    showModal(modalName: string, workingGroup: WorkingGroup) {
      // Hide the navbar so that the model doesn't go behind it.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'hidden';

      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      this.currentWorkingGroup = { ...workingGroup };

      if (this.currentWorkingGroup.members && this.currentWorkingGroup.members.length > 0) {
        this.currentPointPerson = this.currentWorkingGroup.members
          .filter((a: Activist) => {
            return a.point_person;
          })
          .map((a: Activist) => {
            return a.name;
          });

        this.currentMembers = this.currentWorkingGroup.members
          .filter((a: Activist) => {
            return !a.point_person && !a.non_member_on_mailing_list;
          })
          .map((a: Activist) => {
            return a.name;
          });

        this.currentNonMembers = this.currentWorkingGroup.members
          .filter((a: Activist) => {
            return a.non_member_on_mailing_list;
          })
          .map((a: Activist) => {
            return a.name;
          });
      }

      // Get the index for updating the view w/o refreshing the whole page.
      this.workingGroupIndex = this.workingGroups.findIndex((wg) => {
        return wg.id === this.currentWorkingGroup.id;
      });

      this.currentModalName = modalName;

      this.disableConfirmButton = false;
    },
    hideModal() {
      // Show the navbar.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'visible';

      this.currentModalName = '';
      this.workingGroupIndex = -1;
      this.currentWorkingGroup = {} as WorkingGroup;
      this.currentPointPerson = [] as string[];
      this.currentMembers = [] as string[];
      this.currentNonMembers = [] as string[];
    },
    confirmEditWorkingGroupModal() {
      // Rebuild the members array based on current data.
      let members = [] as Activist[];
      if (this.currentPointPerson.length > 0) {
        members.push({ name: this.currentPointPerson[0], point_person: true });
      }
      if (this.currentMembers.length > 0) {
        this.currentMembers.forEach((m: string) => {
          const memberSameAsHost =
            this.currentPointPerson.length > 0 && this.currentPointPerson[0] === m;
          if (!memberSameAsHost) {
            members.push({ name: m });
          }
        });
      }
      if (this.currentNonMembers.length > 0) {
        this.currentNonMembers.forEach((m: string) => {
          const memberSameAsHost =
            this.currentPointPerson.length > 0 && this.currentPointPerson[0] === m;
          if (!memberSameAsHost) {
            members.push({ name: m, non_member_on_mailing_list: true });
          }
        });
      }
      this.currentWorkingGroup.members = members;

      // Save working group
      this.disableConfirmButton = true;

      $.ajax({
        url: '/working_group/save',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(this.currentWorkingGroup),
        success: (data) => {
          this.disableConfirmButton = false;

          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentWorkingGroup.name + ' saved');

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
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    confirmDeleteWorkingGroupModal() {
      this.disableConfirmButton = true;

      $.ajax({
        url: '/working_group/delete',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          working_group_id: this.currentWorkingGroup.id,
        }),
        success: (data) => {
          this.disableConfirmButton = false;

          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentWorkingGroup.name + ' deleted');
          this.workingGroups.splice(this.workingGroupIndex, 1);
          this.hideModal();
        },
        error: (err) => {
          this.disableConfirmButton = false;
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    displayWorkingGroupType(type: string) {
      switch (type) {
        case 'committee':
          return 'Committee';
        case 'working_group':
          return 'Working Group';
      }
      return '';
    },
  },
  data() {
    return {
      loadingActivists: true,
      loadingOrganizers: true,
      loadingWorkingGroups: true,
      currentWorkingGroup: {} as WorkingGroup,
      workingGroups: [] as WorkingGroup[],
      workingGroupIndex: -1,
      disableConfirmButton: false,
      currentModalName: '',
      allActivists: [],
      allOrganizers: [],
      filteredActivists: [],
      filteredOrganizers: [],
      currentPointPerson: [] as string[],
      currentMembers: [] as string[],
      currentNonMembers: [] as string[],
      membersVisible: false,
    };
  },
  computed: {},
  created() {
    // Get working groups
    $.ajax({
      url: '/working_group/list',
      method: 'POST',
      success: (data) => {
        const parsed = JSON.parse(data);
        if (parsed.status === 'error') {
          flashMessage('Error: ' + parsed.message, true);
          return;
        }
        // status === "success"
        this.workingGroups = parsed.working_groups;
        this.loadingWorkingGroups = false;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
        this.loadingWorkingGroups = false;
      },
    });

    // Get organizers for members dropdown
    $.ajax({
      url: '/activist_names/get_organizers',
      method: 'GET',
      success: (data) => {
        const parsed = JSON.parse(data);
        this.allOrganizers = parsed.activist_names;
        this.loadingOrganizers = false;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
        this.loadingOrganizers = false;
      },
    });

    // Get activists for non-members on mailing list dropdown.
    $.ajax({
      url: '/activist_names/get',
      method: 'GET',
      success: (data) => {
        const parsed = JSON.parse(data);
        this.allActivists = parsed.activist_names;
        this.loadingActivists = false;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
        this.loadingActivists = false;
      },
    });
    initializeFlashMessage();
  },
  components: {
    AdbPage,
  },
  directives: {
    focus,
  },
});
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
