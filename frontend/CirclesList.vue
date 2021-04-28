<template>
  <adb-page :title="title === 'CirclesList' ? 'Circles' : 'Geo-Circles'">
    <b-loading :is-full-page="true" v-model="loadingActivists"></b-loading>
    <b-loading :is-full-page="true" v-model="loadingCircles"></b-loading>
    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-button icon-left="plus" @click="showModal('edit-circle-modal')">
            New {{ title === 'CirclesList' ? 'Circle' : 'Geo-Circle' }}
          </b-button>
        </div>
      </div>
      <div class="level-right">
        <div class="level-item">
          <b-button
            v-if="title === 'GeoCirclesList'"
            @click="toggleMembers"
            :icon-left="membersVisible ? 'eye-off' : 'eye'"
          >
            {{ membersVisible ? 'Hide' : 'Show' }} members
          </b-button>
        </div>
      </div>
    </nav>

    <b-table :data="circleGroups" striped hoverable default-sort="name">
      <b-table-column v-slot="props" width="1px">
        <div style="width: 85px;">
          <b-button @click="showModal('edit-circle-modal', props.row)">
            <b-icon icon="pencil" type="is-primary"></b-icon>
          </b-button>
          <b-button @click="showModal('delete-circle-modal', props.row)">
            <b-icon icon="delete" type="is-danger"></b-icon>
          </b-button>
        </div>
      </b-table-column>

      <b-table-column field="name" label="Name" v-slot="props" sortable>
        {{ props.row.name }}
      </b-table-column>

      <b-table-column field="members" label="Host" v-slot="props">
        <!-- There should only ever be one point person -->
        <!-- TODO: calculate this somewhere else so column can be sortable -->
        <template v-for="member in props.row.members">
          <template v-if="member.point_person">
            {{ member.name }}
          </template>
        </template>
      </b-table-column>

      <b-table-column
        field="last_meeting"
        label="Last Event"
        v-slot="props"
        sortable
        v-if="title === 'CirclesList'"
      >
        <span class="tag" :class="colorLastMeeting(props.row.last_meeting)">
          {{ props.row.last_meeting || 'None' }}
        </span>
      </b-table-column>

      <b-table-column
        label="Total Members"
        v-slot="props"
        v-if="title === 'GeoCirclesList' && !membersVisible"
      >
        <!-- TODO: calculate this somewhere else so column can be sortable -->
        {{ numberOfCircleGroupMembers(props.row) }}
      </b-table-column>

      <b-table-column
        label="Members"
        field="members"
        v-slot="props"
        v-if="title === 'GeoCirclesList' && membersVisible"
      >
        <ul v-for="member in props.row.members">
          <template v-if="!member.point_person">
            <li>{{ member.name }}</li>
          </template>
        </ul>
      </b-table-column>
    </b-table>

    <b-modal
      :active="currentModalName === 'delete-circle-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">Delete circle</p>
        </header>
        <section class="modal-card-body">
          <p>
            Are you sure you want to delete <strong>{{ currentCircleGroup.name }}</strong
            >?
          </p>
          <b-message type="is-warning" has-icon class="mt-3">
            Before deleting a circle, be sure to remove all members of that circle.
          </b-message>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Cancel" @click="hideModal" />
          <b-button
            label="Delete"
            type="is-danger"
            v-bind:disabled="disableConfirmButton"
            @click="confirmDeleteCircleGroupModal"
          />
        </footer>
      </div>
    </b-modal>

    <b-modal
      :active="currentModalName === 'edit-circle-modal'"
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
            {{ currentCircleGroup.id ? 'Edit' : 'New' }}
            {{ title === 'CirclesList' ? 'Circle' : 'Geo-Circle' }}
          </p>
        </header>
        <section class="modal-card-body">
          <b-field label="Name" label-position="on-border">
            <b-input
              type="text"
              v-model.trim="currentCircleGroup.name"
              icon="circle-slice-8"
              required
            ></b-input>
          </b-field>

          <b-field label="Type" label-position="on-border" hidden>
            <b-input type="text" v-model.trim="currentCircleGroup.type" required></b-input>
          </b-field>

          <b-field
            :label="'Description' + (title === 'GeoCirclesList' ? ' or Notes' : '')"
            label-position="on-border"
          >
            <b-input
              type="text"
              v-model.trim="currentCircleGroup.description"
              icon="text-box"
            ></b-input>
          </b-field>

          <b-field
            label="Meeting Day & Time"
            label-position="on-border"
            v-if="title === 'CirclesList'"
          >
            <b-input
              type="text"
              v-model.trim="currentCircleGroup.meeting_time"
              icon="calendar-blank"
            ></b-input>
          </b-field>

          <b-field
            label="Meeting Location"
            label-position="on-border"
            v-if="title === 'CirclesList'"
          >
            <b-input
              type="text"
              v-model.trim="currentCircleGroup.meeting_location"
              icon="map-marker"
            ></b-input>
          </b-field>

          <b-field label="Coordinates" label-position="on-border">
            <b-input
              type="text"
              v-model.trim="currentCircleGroup.coords"
              icon="ruler"
              :disabled="title === 'GeoCirclesList'"
            ></b-input>
          </b-field>

          <b-field>
            <b-switch v-model="currentCircleGroup.visible" type="is-success"
              >Visible to public</b-switch
            >
          </b-field>

          <b-field label="Host">
            <b-taginput
              v-model="currentCircleHost"
              :data="filteredActivists"
              autocomplete
              :allow-new="false"
              icon="crown"
              placeholder="Search by name..."
              @typing="getFilteredActivists"
              maxtags="1"
              type="is-info"
              dropdown-position="top"
            ></b-taginput>
          </b-field>

          <b-field label="Members" v-if="title === 'GeoCirclesList'">
            <b-taginput
              v-model="currentCircleMembers"
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
            v-bind:disabled="disableConfirmButton"
            @click="confirmEditCircleGroupModal"
          />
        </footer>
      </div>
    </b-modal>
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage } from './flash_message';
import { focus } from './directives/focus';
import moment from "moment";

interface Activist {
  name: string;
  point_person?: boolean;
  non_member_on_mailing_list?: boolean;
}

interface Circle {
  id: number;
  name: string;
  description: string;
  meeting_time: string;
  meeting_location: string;
  coords: string;
  visible: boolean;
  members: Activist[];
  last_meeting: string;
  type: string;
}

export default Vue.extend({
  name: 'circle-list',
  props: {
    title: String,
  },
  methods: {
    isMobile() {
      return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    },
    colorLastMeeting(text: string) {
      const time = moment(text);
      let c = '';
      if (time.isValid()) {
        c = 'is-danger';
      }
      if (time.isAfter(moment().add(-32, 'day'))) {
        c = 'is-warning';
      }
      if (time.isAfter(moment().add(-8, 'day'))) {
        c = 'is-success';
      }
      return c;
    },
    getFilteredActivists(text: string) {
      this.filteredActivists = this.allActivists.filter((a: string) => {
        return a.toLowerCase().startsWith(text.toLowerCase());
      });
    },
    toggleMembers() {
      this.membersVisible = !this.membersVisible;
    },
    showModal(modalName: string, circleGroup: Circle) {
      // Hide the navbar so that the model doesn't go behind it.
      const mainNav = document.getElementById("mainNav");
      if (mainNav) mainNav.style.visibility = "hidden";

      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      this.currentCircleGroup = { ...circleGroup };

      if (this.currentCircleGroup.members && this.currentCircleGroup.members.length > 0) {
        this.currentCircleHost = this.currentCircleGroup.members
          .filter((a: Activist) => {
            return a.point_person;
          })
          .map((a: Activist) => {
            return a.name;
          });

        this.currentCircleMembers = this.currentCircleGroup.members
          .filter((a: Activist) => {
            return !a.point_person;
          })
          .map((a: Activist) => {
            return a.name;
          });
      }

      // always set the type based on the page we are on
      this.currentCircleGroup.type = this.title === 'CirclesList' ? 'circle' : 'geo-circle';

      // Get the index for updating the view w/o refreshing the whole page.
      this.circleGroupIndex = this.circleGroups.findIndex((c) => {
        return c.id === this.currentCircleGroup.id;
      });

      this.currentModalName = modalName;

      this.disableConfirmButton = false;
    },
    hideModal() {
      // Show the navbar.
      const mainNav = document.getElementById("mainNav");
      if (mainNav) mainNav.style.visibility = "visible";

      this.currentModalName = '';
      this.circleGroupIndex = -1;
      this.currentCircleGroup = {} as Circle;
      this.currentCircleHost = [] as string[];
      this.currentCircleMembers = [] as string[];
    },
    confirmEditCircleGroupModal() {
      // Rebuild the members array based on current data.
      let members = [] as Activist[];
      if (this.currentCircleHost.length > 0) {
        members.push({ name: this.currentCircleHost[0], point_person: true });
      }
      if (this.currentCircleMembers.length > 0) {
        this.currentCircleMembers.forEach((m: string) => {
          const memberSameAsHost =
            this.currentCircleHost.length > 0 && this.currentCircleHost[0] === m;
          if (!memberSameAsHost) {
            members.push({ name: m });
          }
        });
      }
      this.currentCircleGroup.members = members;

      // Save working group
      this.disableConfirmButton = true;

      $.ajax({
        url: '/circle/save',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(this.currentCircleGroup),
        success: (data) => {
          this.disableConfirmButton = false;

          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentCircleGroup.name + ' saved');

          if (this.circleGroupIndex === -1) {
            // New circle, insert at the top
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
    confirmDeleteCircleGroupModal() {
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

          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentCircleGroup.name + ' deleted');
          this.circleGroups.splice(this.circleGroupIndex, 1);
          this.hideModal();
        },
        error: (err) => {
          this.disableConfirmButton = false;
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    numberOfCircleGroupMembers(circleGroup: Circle) {
      if (!circleGroup.members) {
        return 0;
      }

      let count = 0;
      for (let i = 0; i < circleGroup.members.length; i++) {
        if (!circleGroup.members[i].non_member_on_mailing_list) {
          count++;
        }
      }

      return count;
    },
  },
  data() {
    return {
      loadingCircles: true,
      loadingActivists: true,
      currentCircleGroup: {} as Circle,
      circleGroups: [] as Circle[],
      circleGroupIndex: -1,
      disableConfirmButton: false,
      currentModalName: '',
      membersVisible: false,
      allActivists: [],
      filteredActivists: [],
      currentCircleHost: [] as string[],
      currentCircleMembers: [] as string[],
    };
  },
  computed: {},
  created() {
    // Get circles.
    $.ajax({
      url: '/circle/list',
      method: 'POST',
      success: (data) => {
        const parsed = JSON.parse(data);
        if (parsed.status === 'error') {
          flashMessage('Error: ' + parsed.message, true);
          return;
        }
        // status === "success"
        this.circleGroups = parsed.circle_groups.filter((c: Circle) => {
          return this.title === 'GeoCirclesList' ? c.type === 'geo-circle' : c.type === 'circle';
        });
        this.loadingCircles = false;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
        this.loadingCircles = false;
      },
    });

    // Get activists for members dropdown.
    $.ajax({
      url: '/activist_names/get_chaptermembers',
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
  },
  components: {
    AdbPage,
  },
  directives: {
    focus,
  },
});
</script>
