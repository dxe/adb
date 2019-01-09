<template>
  <div id="app" class="main">
    <div class="activist-list-filters form-inline">
      <input
        v-on:input="debounceSearchInput"
        class="form-control filter-margin"
        type="text"
        placeholder="Search"
      />

      <button
        class="btn-link"
        @click="toggleShowOptions('filters')"
        v-if="view == 'all_activists' || view == 'activist_pool'"
      >
        <span v-if="showOptions !== 'filters'">+</span
        ><span v-if="showOptions === 'filters'">-</span> Filters
      </button>
      <button class="btn-link" @click="toggleShowOptions('columns')">
        <span v-if="showOptions !== 'columns'">+</span
        ><span v-if="showOptions === 'columns'">-</span> Columns
      </button>

      <span>&nbsp;&nbsp;&nbsp;&nbsp;<b>Total rows: </b></span><span id="rowCount">0</span>

      <div v-if="showOptions === 'filters'">
        <div v-if="view == 'all_activists' || view == 'activist_pool'">
          <label>Last Event From:</label>
          <input v-model="lastEventDateFrom" class="form-control filter-margin" type="date" />
        </div>
        <div v-if="view == 'all_activists' || view == 'activist_pool'">
          <label>Last Event To:</label>
          <input v-model="lastEventDateTo" class="form-control filter-margin" type="date" />
        </div>
        <!--         <div v-if="view === 'action_team'">
          <label>Action Team:</label>
          <select id="filterActionTeam" v-model="filterActionTeam" class="form-control filter-margin">
            <option>All</option>
            <option>Communications</option>
            <option>Community</option>
            <option>Development</option>
            <option>Direct Action</option>
            <option>Finance</option>
          </select>
        </div> -->
      </div>

      <div v-if="showOptions === 'columns'">
        <div v-for="column in columns">
          <span v-if="column.header !== ''">
            <input type="checkbox" :id="column.header" v-model="column.enabled" />
            <label :for="column.header">{{ column.longHeader }}</label>
          </span>
        </div>
      </div>
    </div>
    <div id="hot-table-container">
      <HotTable
        ref="hot"
        :root="root"
        :settings="hotSettings"
        :data="activists"
        :height="height"
      ></HotTable>
    </div>
    <modal
      name="activist-options-modal"
      height="auto"
      classes="no-background-color no-top"
      @opened="modalOpened"
      @closed="modalClosed"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">{{ currentActivist.name }}</h2>
          </div>
          <div class="modal-body">
            <ul class="activist-options-body">
              <!-- <li>
                <a @click="showModal('connection-modal', currentActivist, activistIndex)">Add Maintenance Connection</a>
              </li> -->
              <li>
                <a @click="showModal('merge-activist-modal', currentActivist, activistIndex)"
                  >Merge Activist</a
                >
              </li>
              <li>
                <a @click="showModal('hide-activist-modal', currentActivist, activistIndex)"
                  >Hide Activist</a
                >
              </li>
            </ul>
          </div>
        </div>
      </div>
    </modal>
    <modal
      name="merge-activist-modal"
      :height="650"
      classes="no-background-color"
      @opened="modalOpened"
      @closed="modalClosed"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header"><h2 class="modal-title">Merge activist</h2></div>
          <div class="modal-body">
            <p>Merging activists is used to combine redundant activist entries</p>
            <p>Merging this activist does two things:</p>
            <ul>
              <li>
                all of this activist&#39;s attendance data will be merged into the target activist
              </li>
              <li>this activist will be hidden</li>
            </ul>
            <p>Non-attendance data (e.g. email, location, etc) is <strong>NOT</strong> merged.</p>
            <p>Merge {{ currentActivist.name }} into another activist:</p>
            <p>
              Target activist:
              <select
                id="merge-target-activist"
                class="filter-margin"
                style="min-width: 200px"
              ></select>
            </p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-danger"
              @click="confirmMergeActivistModal"
              v-focus
            >
              Merge activist
            </button>
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
          <div class="modal-header"><h2 class="modal-title">Hide activist</h2></div>
          <div class="modal-body">
            <p>Are you sure you want to hide {{ currentActivist.name }}?</p>
            <p>
              Hiding an activist hides them from the activist list page but does not delete any
              event data associated with them. If this activist is a duplicate of another activist,
              you should merge them instead.
            </p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-danger"
              @click="confirmHideActivistModal"
              v-focus
            >
              Hide activist
            </button>
          </div>
        </div>
      </div>
    </modal>
    <!-- <modal
       name="connection-modal"
       :height="400"
       classes="no-background-color"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Add maintenance connection</h2>
          </div>
          <div class="modal-body">
            <p><b>WARNING: This feature is not yet available!</b></p>
            <br />
            <p>Activist ID: {{currentActivist.id}}</p>
            <p>Activist Name: {{currentActivist.name}}</p>
            <p>Connector: {{currentActivist.connector}}</p>
            <p>Date: <input id="connection-date" type="date"></p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Cancel</button>
          </div>
        </div>
      </div>
    </modal> -->
  </div>
</template>

<script>
import vmodal from 'vue-js-modal';
import HotTable from './external/vue-handsontable-official/HotTable.vue';
import { rewriteSettings } from './external/vue-handsontable-official/helpers';
import Vue from 'vue';
import { focus } from './directives/focus';
import { flashMessage } from './flash_message';
import { EventBus } from './EventBus';
import { initActivistSelect } from './chosen_utils';
import debounce from 'debounce';

Vue.use(vmodal);

function getDefaultColumns(view) {
  return [
    {
      header: '',
      data: {
        renderer: optionsButtonRenderer,
        readOnly: true,
        disableVisualSelection: true,
        colWidths: 35,
      },
      enabled: true,
    },
    // Standard activist fields
    {
      header: 'ID',
      longHeader: 'Activist ID',
      data: {
        type: 'numeric',
        data: 'id',
        readOnly: true,
        colWidths: 50,
      },
      enabled: false,
    },
    {
      header: 'Name',
      longHeader: 'Activist Name',
      data: {
        data: 'name',
        colWidths: 150,
      },
      enabled: true,
    },
    {
      header: 'Managing',
      longHeader: 'Managing',
      data: {
        data: 'dev_manager',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },
    {
      header: 'Points',
      longHeader: 'Leaderboard Points',
      data: {
        type: 'numeric',
        data: 'total_points',
        readOnly: true,
        colWidths: 50,
      },
      enabled: view === 'leaderboard' || view === 'action_team',
    },
    {
      header: 'Email',
      longHeader: 'Email',
      data: {
        data: 'email',
        colWidths: 150,
      },
      enabled:
        view === 'all_activists' ||
        view === 'activist_pool' ||
        view === 'activist_recruitment' ||
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development' ||
        view === 'circle_member_prospects',
    },
    {
      header: 'Phone',
      longHeader: 'Phone Number',
      data: {
        data: 'phone',
        colWidths: 100,
      },
      enabled: false,
    },
    {
      header: 'Location',
      longHeader: 'Location',
      data: {
        data: 'location',
        colWidths: 100,
      },
      enabled: view === 'action_team',
    },
    {
      header: 'Facebook',
      longHeader: 'Facebook URL',
      data: {
        data: 'facebook',
      },
      enabled:
        view === 'all_activists' || view === 'activist_recruitment' || view === 'activist_pool',
    },

    // {
    //  header: "Contacted Date",
    //  data: {
    //    data: "contacted_date",
    //    type: 'date',
    //    dateFormat: 'YYYY-MM-DD',
    //    correctFormat: true,
    //    colWidths: 100,
    //  },
    //  enabled: view === "activist_pool",
    // }, {
    //  header: "Interested",
    //  data: {
    //    data: "interested",
    //    colWidths: 100,
    //    type: 'dropdown',
    //    source: [
    //     "",
    //      "Yes",
    //      "No",
    //    ],
    //  },
    //  enabled: view === "activist_pool",
    // },

    // ActivistMembershipData
    // {
    //  header: "Recruitment Connection Date",
    //  data: {
    //    data: "meeting_date",
    //    type: 'date',
    //    dateFormat: 'YYYY-MM-DD',
    //    correctFormat: true,
    //    colWidths: 100,
    //  },
    //  enabled: view === "activist_pool",
    // },
    //{
    //  header: "Escalation",
    //  data: {
    //    data: "escalation",
    //    type: 'dropdown',
    //    colWidths: 100,
    //    source: [
    //      "",
    //      "Yes",
    //      "No",
    //    ],
    //  },
    //  enabled: view === "activist_recruitment",
    //},

    {
      header: 'Level',
      longHeader: 'Activist Level',
      data: {
        data: 'activist_level',
        readOnly: false,
        colWidths: 140,
        type: 'dropdown',
        source: ['Supporter', 'Circle Member', 'Chapter Member', 'Organizer', 'Senior Organizer'],
      },
      enabled:
        view === 'all_activists' ||
        view === 'activist_recruitment' ||
        view === 'leaderboard' ||
        view === 'action_team' ||
        view === 'organizer_prospects' ||
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development' ||
        view === 'circle_member_prospects',
    },
    {
      header: 'Applied',
      longHeader: 'Application Date',
      data: {
        type: 'date',
        data: 'dev_application_date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
        readOnly: true,
      },
      enabled:
        view === 'circle_member_prospects' ||
        view === 'chapter_member_prospects' ||
        view == 'organizer_prospects',
    },
    {
      header: 'Prosp. Organizer',
      longHeader: 'Prospective Organizer',
      data: {
        type: 'checkbox',
        data: 'prospect_organizer',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },
    {
      header: 'Prosp. Ch. Member',
      longHeader: 'Prospective Chapter Member',
      data: {
        type: 'checkbox',
        data: 'prospect_chapter_member',
        colWidths: 110,
      },
      enabled: view === 'chapter_member_prospects',
    },
    {
      header: 'Prosp. Cir. Member',
      longHeader: 'Prospective Circle Member',
      data: {
        type: 'checkbox',
        data: 'prospect_circle_member',
        colWidths: 105,
      },
      enabled: view === 'circle_member_prospects',
    },
    {
      header: 'Interest',
      longHeader: 'Interest',
      data: {
        data: 'dev_interest',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects' || view === 'circle_member_prospects',
    },
    {
      header: 'Working Groups',
      longHeader: 'Working Group Membership',
      data: {
        data: 'working_group_list',
        readOnly: true,
        colWidths: 200,
      },
      enabled: view === 'action_team',
    },
    {
      header: 'Circles',
      longHeader: 'Circle Membership',
      data: {
        data: 'circles_list',
        readOnly: true,
        colWidths: 200,
      },
      enabled: view === 'action_team',
    },
    {
      header: 'WG or Cir. Member',
      longHeader: 'Working Group or Circle Member',
      data: {
        type: 'checkbox',
        readOnly: true,
        data: 'wg_or_cir_member',
        colWidths: 100,
      },
      enabled: view === 'chapter_member_prospects' || view === 'chapter_member_development',
    },
    {
      header: 'First Event',
      longHeader: 'First Event Attended',
      data: {
        data: 'first_event_name',
        readOnly: true,
        colWidths: 200,
      },
      enabled:
        view === 'activist_pool' || view === 'activist_recruitment' || view === 'leaderboard',
    },
    {
      header: 'Last Event',
      longHeader: 'Last Event Attended',
      data: {
        data: 'last_event_name',
        readOnly: true,
        colWidths: 200,
      },
      enabled: view === 'activist_recruitment' || view === 'leaderboard',
    },
    {
      header: 'Total Events',
      longHeader: 'Total Events Attended',
      data: {
        type: 'numeric',
        data: 'total_events',
        readOnly: true,
        colWidths: 90,
      },
      enabled: view === 'leaderboard',
    },
    // {
    //   header: "Active",
    //   data: {
    //     type: "checkbox",
    //     data: "active",
    //     readOnly: true,
    //     colWidths: 55,
    //   },
    //   enabled: (view === "action_team"),
    // },
    {
      header: 'MPI',
      longHeader: 'Movement Power Index Status',
      data: {
        type: 'checkbox',
        data: 'mpi',
        readOnly: true,
        colWidths: 30,
      },
      enabled:
        view === 'action_team' ||
        view === 'activist_pool' ||
        view === 'activist_recruitment' ||
        view === 'all_activists' ||
        view === 'leaderboard' ||
        view === 'development' ||
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development',
    },
    {
      header: 'Connector',
      longHeader: 'Connector Name',
      data: {
        data: 'connector',
        colWidths: 125,
      },
      enabled: view === 'activist_pool' || view === 'action_team' || view === 'development',
    },
    {
      header: 'Last Maint. Conn.',
      longHeader: 'Last Maintenance Connection',
      data: {
        type: 'date',
        data: 'last_connection',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
        readOnly: true,
      },
      enabled: view === 'development',
    },
    {
      header: 'Workshop',
      longHeader: 'Date Attended Training: Workshop',
      data: {
        type: 'date',
        data: 'training0',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'action_team' || view === 'chapter_member_prospects',
    },
    {
      header: 'Consent&A-O',
      longHeader: 'Date Attended Training: Consent & Anti-Opression',
      data: {
        type: 'date',
        data: 'training1',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'development',
    },
    {
      header: 'Nuts&Bolts',
      longHeader: 'Date Attended Training: Nuts & Bolts',
      data: {
        type: 'date',
        data: 'training2',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'development',
    },
    {
      header: 'Exp. Cmnty',
      longHeader: 'Date Attended Training: Expanding Our Community',
      data: {
        type: 'date',
        data: 'training3',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'development',
    },
    {
      header: 'Healthy Cmnty',
      longHeader: 'Date Attended Training: Healthy Community',
      data: {
        type: 'date',
        data: 'training4',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'development',
    },
    {
      header: 'Emergent Ldshp',
      longHeader: 'Date Attended Training: Emergent Leadership',
      data: {
        type: 'date',
        data: 'training5',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'development',
    },
    {
      header: 'Histry&Vision',
      longHeader: 'Date Attended Training: History & Vision',
      data: {
        type: 'date',
        data: 'training6',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'development',
    },
    {
      header: 'First Email',
      longHeader: 'Chapter Membership: Date First Email Sent',
      data: {
        type: 'date',
        data: 'cm_first_email',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'chapter_member_prospects',
    },
    {
      header: 'Apprv. Email',
      longHeader: 'Chapter Membership: Date Approval Email Sent',
      data: {
        type: 'date',
        data: 'cm_approval_email',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'chapter_member_prospects',
    },
    {
      header: 'Warning Email',
      longHeader: 'Chapter Membership: Date Removal Warning Email Sent',
      data: {
        type: 'date',
        data: 'cm_warning_email',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'chapter_member_development',
    },
    {
      header: 'First Email',
      longHeader: 'Circle Membership: Date First Email Sent',
      data: {
        type: 'date',
        data: 'cir_first_email',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'circle_member_prospects',
    },

    {
      header: 'Point Auth',
      longHeader: 'Organizer: Date of Point Authorization',
      data: {
        type: 'date',
        data: 'dev_auth',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },
    {
      header: 'Sent Elig. Email',
      longHeader: 'Organizer: Date Eligibility Email Sent',
      data: {
        type: 'date',
        data: 'dev_email_sent',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },
    {
      header: 'Vetting',
      longHeader: 'Organizer: Vetting Completed',
      data: {
        type: 'checkbox',
        data: 'dev_vetted',
        colWidths: 50,
      },
      enabled: view === 'organizer_prospects',
    },
    {
      header: 'Interview',
      longHeader: 'Organizer: Date of Interview',
      data: {
        type: 'date',
        data: 'dev_interview',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },
    {
      header: 'Completed Onboarding Logistics',
      longHeader: 'Organizer: Date of Onboarding Logistics Completion',
      data: {
        type: 'checkbox',
        data: 'dev_onboarding',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },
  ];
}

// Constants related to list ordering
// Corresponds to the constants DescOrder and AscOrder in model/activist.go
const DescOrder = 2;
const AscOrder = 1;

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
// var statusOrder = {
//   "Current": 1,
//   "New": 2,
//   "Former": 3,
//   "No attendance": 4,
// };

// var activistLevelOrder = {
//   "activist" : 3,
//   "core_activist" : 2,
//   "organizer" : 1,
//   "senior_organizer" : 0
// };

window['showOptionsModal'] = function(row) {
  EventBus.$emit('activist-show-options-modal', row);
};

function optionsButtonRenderer(instance, td, row, col, prop, value, cellProperties) {
  td.innerHTML =
    '<button ' +
    'data-role="trigger" ' +
    'class="activist-options-btn btn btn-default btn-xs dropdown-toggle glyphicon glyphicon-option-horizontal" ' +
    'type="button" ' +
    'onclick="window.showOptionsModal(' +
    row +
    ')"></button>';
  return td;
}

function initialDateFromValue() {
  var d = new Date();
  var rawYear = d.getFullYear();
  var rawMonth = d.getMonth() + 1;

  var monthOffset = 2;
  rawMonth -= monthOffset;
  if (rawMonth <= 0) {
    // 12 + rawMonth will be the correct month from the previous year
    // because rawMonth is either 0 or negative at this point.
    rawMonth = 12 + rawMonth;
    rawYear -= 1;
  }

  var year = '' + rawYear;
  var month = rawMonth > 9 ? '' + rawMonth : '0' + rawMonth;

  var fromDate = year + '-' + month + '-01';
  return fromDate;
}

function initialDateToValue() {
  var d = new Date();
  // An ISO date looks like "2017-11-01T23:21:50.377Z", so we cut off
  // everything after the date.
  return d.toISOString().slice(0, 10);
}

function generateBooleanSortFn(field, ascending) {
  return function(a, b) {
    var order = a[field] === b[field] ? 0 : Number(a[field]) - Number(b[field]);
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateStringSortFn(field, ascending) {
  return function(a, b) {
    var order = a[field].toLowerCase() < b[field].toLowerCase() ? -1 : 1;
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateGenericSortFn(field, ascending) {
  return function(a, b) {
    var order = a[field] < b[field] ? -1 : 1;
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateDateSortFn(field, ascending) {
  return function(a, b) {
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

    var order = valueA < valueB ? -1 : 1;

    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

export default {
  name: 'activist-list',
  props: {
    // `view` is the default view to show. It can be one of:
    // "all_activists", "leaderboard", "activist_pool",
    // "activist_recruitment", or "action_team"
    view: {
      type: String,
      validator(value) {
        var validViews = [
          'all_activists',
          'leaderboard',
          'activist_pool',
          'activist_recruitment',
          'action_team',
          'development',
        ];
        return validViews.indexOf(value) !== -1;
      },
    },
  },
  methods: {
    showOptionsModal(row) {
      var activist = this.activists[row];
      this.showModal('activist-options-modal', activist, row);
    },
    showModal(modalName, activist, index) {
      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      // Show the modal in the next tick so that this code runs after
      // vue has hidden the previous modal.
      Vue.nextTick(() => {
        this.currentActivist = activist;

        if (index != undefined) {
          this.activistIndex = index; // needed for updating activist
        } else {
          this.activistIndex = -1;
        }

        this.currentModalName = modalName;
        this.$modal.show(modalName);
      });
    },
    hideModal() {
      if (this.currentModalName) {
        this.$modal.hide(this.currentModalName);
      }
      this.currentModalName = '';
      this.activistIndex = -1;
      this.currentActivist = {};
    },
    modalOpened() {
      // Add noscroll to body tag so it doesn't scroll while the modal
      // is shown.
      $(document.body).addClass('noscroll');
      this.disableConfirmButton = false;

      if (this.currentModalName == 'merge-activist-modal') {
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
    modalClosed() {
      // Allow body to scroll after modal is closed.
      $(document.body).removeClass('noscroll');
    },
    removeActivist(id) {
      var activistIndex;
      for (var i = 0; i < this.allActivists.length; i++) {
        if (this.allActivists[i].id === id) {
          activistIndex = i;
        }
      }
      if (!activistIndex) {
        throw new Error("Couldn't find activist index for activist with id: " + id);
      }
      this.allActivists = this.allActivists
        .slice(0, activistIndex)
        .concat(this.allActivists.slice(activistIndex + 1));
    },
    confirmMergeActivistModal() {
      var targetActivistName = $('#merge-target-activist').val();
      if (!targetActivistName) {
        flashMessage('Must choose an activist to merge into', true);
        return;
      }

      this.disableConfirmButton = true;
      var currentActivistID = this.currentActivist.id;

      $.ajax({
        url: '/activist/merge',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          current_activist_id: currentActivistID,
          target_activist_name: targetActivistName,
        }),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          flashMessage(this.currentActivist.name + ' was merged into ' + targetActivistName);

          // Remove activist from list.
          this.removeActivist(currentActivistID);

          this.hideModal();
        },
        error: (err) => {
          this.disableConfirmButton = false;

          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    confirmHideActivistModal() {
      this.disableConfirmButton = true;
      var currentActivistID = this.currentActivist.id;

      $.ajax({
        url: '/activist/hide',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id: currentActivistID }),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          flashMessage(this.currentActivist.name + ' was hidden');

          // Remove activist from list.
          this.removeActivist(currentActivistID);

          this.hideModal();
        },
        error: (err) => {
          this.disableConfirmButton = false;

          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    loadActivists() {
      $.ajax({
        url: '/activist/list',
        method: 'POST',
        data: JSON.stringify(this.listActivistsParameters()),
        success: (data) => {
          var parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }

          // status === "success"
          var activistList = parsed.activist_list;

          // filtering
          if (
            this.view === 'activist_pool' ||
            this.view === 'activist_recruitment' ||
            this.view === 'action_team' ||
            this.view === 'leaderboard' ||
            this.view === 'development' ||
            this.view === 'organizer_prospects' ||
            this.view === 'chapter_member_prospects' ||
            this.view === 'circle_member_prospects' ||
            this.view === 'chapter_member_development'
          ) {
            var activistListFiltered;
            activistListFiltered = activistList.filter((el) => {
              if (this.view === 'activist_pool') {
                return el.activist_level == 'Supporter';
              } else if (this.view === 'action_team') {
                var selectedActionTeam = $('#filterActionTeam :selected').text();

                if (
                  selectedActionTeam != 'All' &&
                  selectedActionTeam != '' &&
                  selectedActionTeam != null
                ) {
                  return (
                    (el.activist_level == 'Action Team' ||
                      el.activist_level == 'Organizer' ||
                      el.activist_level == 'Senior Organizer') &&
                    (el.action_team_focus == selectedActionTeam ||
                      el.action_team_focus_secondary
                        .toLowerCase()
                        .indexOf(selectedActionTeam.toLowerCase()) != -1)
                  );
                } else {
                  return (
                    el.activist_level == 'Action Team' ||
                    el.activist_level == 'Organizer' ||
                    el.activist_level == 'Senior Organizer'
                  );
                }
              } else if (this.view === 'leaderboard') {
                return el.active == 1;
              } else if (this.view === 'development') {
                return el.activist_level == 'Organizer' || el.activist_level == 'Senior Organizer';
              } else if (this.view === 'organizer_prospects') {
                return el.prospect_organizer == 1;
              } else if (this.view === 'chapter_member_prospects') {
                return el.prospect_chapter_member == 1;
              } else if (this.view === 'circle_member_prospects') {
                return el.prospect_circle_member == 1;
              } else if (this.view === 'chapter_member_development') {
                return (
                  el.activist_level == 'Chapter Member' ||
                  el.activist_level == 'Organizer' ||
                  el.activist_level == 'Senior Organizer'
                );
              } else {
                return true; // unreachable
              }
            });
            activistList = activistListFiltered;
          }

          if (activistList !== null) {
            this.allActivists = activistList;
          }
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    afterChangeCallback(changes, source) {
      if (
        source !== 'edit' &&
        source !== 'CopyPaste.paste' &&
        source !== 'UndoRedo.undo' &&
        source !== 'UndoRedo.redo'
      ) {
        return;
      }
      for (var i = 0; i < changes.length; i++) {
        var change = changes[i];
        var columnIndex = change[0];
        var columnName = change[1];
        var previousData = change[2];
        var newData = change[3];

        var activist = this.activists[columnIndex];
        (function(change) {
          // TODO: use change?
          $.ajax({
            url: '/activist/save',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(activist),
            success: (data) => {
              var parsed = JSON.parse(data);
              if (parsed.status === 'error') {
                flashMessage('Error: ' + parsed.message, true);
                return;
              }
            },
            error: (err) => {
              console.warn(err.responseText);
              flashMessage('Server error: ' + err.responseText, true);
            },
          });
        })(change);
      }
    },
    setHOTHeight() {
      var hotContainer = document.getElementById('hot-table-container');
      if (!hotContainer) {
        this.height = 500;
        return;
      }
      var y = hotContainer.getBoundingClientRect().y;
      this.height = window.innerHeight - y;
    },
    listActivistsParameters() {
      var order_field = 'last_event';
      return {
        order: DescOrder,
        order_field:
          this.view === 'leaderboard' || this.view === 'action_team'
            ? 'total_points'
            : 'last_event',
        last_event_date_to: this.lastEventDateTo,
        last_event_date_from: this.lastEventDateFrom,
      };
    },
    toggleShowOptions(optionsType) {
      if (this.showOptions === optionsType) {
        this.showOptions = '';
      } else {
        this.showOptions = optionsType;
      }
      Vue.nextTick(() => {
        this.setHOTHeight(); // Resize the spreadsheet.
      });
    },
    refreshHOTData() {
      var table = this.$refs.hot.table;
      var newSettings = {
        data: rewriteSettings(this.activists),
      };
      table.updateSettings(newSettings);
    },
    sortColumn(col) {
      var field = col.data.data;
      if (!field) {
        // Don't sort columsn with no data field (e.g. the first
        // column).
        return;
      }

      var type = col.data.type;
      var sortFunction;
      var ascending = shouldSortByAscending(field, type === 'date');
      if (type === 'date') {
        sortFunction = generateDateSortFn(field, ascending);
      } else if (type === 'numeric') {
        sortFunction = generateGenericSortFn(field, ascending);
      } else if (type === 'checkbox') {
        sortFunction = generateBooleanSortFn(field, ascending);
      } else {
        sortFunction = generateStringSortFn(field, ascending);
      }

      this.allActivists.sort(sortFunction);

      setPreviousSortData(field, ascending);

      this.refreshHOTData();
    },
    afterOnCellMouseDownCallback(event, coords, td) {
      // If the row is -1, then the user clicked on a column header.
      if (coords.row === -1) {
        // To find the column this maps to, we iterate through all the enabled columns.
        var visibleColIndex = coords.col;
        var foundCol;
        for (var i = 0; i < this.columns.length; i++) {
          var col = this.columns[i];
          if (col.enabled) {
            if (visibleColIndex === 0) {
              foundCol = col;
              break;
            }
            visibleColIndex--;
          }
        }
        if (!foundCol) {
          throw new Error('Could not find column at index ' + coords.col);
        }
        this.sortColumn(foundCol);
      }
    },
    debounceSearchInput: debounce(function(e) {
      this.search = e.target.value;
    }, 500),
  },
  data() {
    if (this.view === ('all_activists' || 'leaderboard')) {
      var initDateFrom = initialDateFromValue();
      var initDateTo = initialDateToValue();
    } else {
      var initDateFrom = '';
      var initDateTo = '';
    }

    return {
      root: 'activists-root',
      currentModalName: '',
      activistIndex: -1,
      currentActivist: {},
      disableConfirmButton: false,
      allActivists: [],
      height: 500,
      columns: getDefaultColumns(this.view),
      lastEventDateFrom: initDateFrom,
      lastEventDateTo: initDateTo,
      filterActionTeam: 'All',
      showOptions: '',
      search: '',
    };
  },
  computed: {
    hotSettings() {
      const columns = [];
      const columnHeaders = [];
      for (var i = 0; i < this.columns.length; i++) {
        var col = this.columns[i];
        if (!col.enabled) {
          continue;
        }
        columns.push(this.columns[i].data);
        columnHeaders.push(this.columns[i].header);
      }
      return {
        columns: columns,
        colHeaders: columnHeaders,
        rowHeaders: this.view === 'leaderboard',
        disableVisualSelection: false,
        multiSelect: true,
        fillHandle: false,
        afterChange: this.afterChangeCallback.bind(this),
        afterOnCellMouseDown: this.afterOnCellMouseDownCallback.bind(this),
        undo: true,
        manualColumnResize: true,
        autoColumnSize: false,
        colWidths: 200,
        viewportRowRenderingOffset: 100,
        viewportColumnRenderingOffset: 20,
        wordWrap: false,
        //fixedColumnsLeft: 2, // this causes too much havoc
      };
    },
    activists() {
      if (this.search.length < 3) {
        return this.allActivists;
      }

      // This search implementation is slow when we have lots of data.
      // Make it faster when that becomes an issue.
      var searchNormalized = this.search.trim().toLowerCase();
      var activists = [];
      for (var i = 0; i < this.allActivists.length; i++) {
        var activist = this.allActivists[i];
        if (activist.name.toLowerCase().includes(searchNormalized)) {
          activists.push(activist);
        }
      }
      return activists;
    },
  },
  watch: {
    lastEventDateFrom() {
      this.loadActivists();
    },
    lastEventDateTo() {
      this.loadActivists();
    },
    filterActionTeam() {
      this.loadActivists();
    },
  },
  created() {
    this.loadActivists();
    EventBus.$on('activist-show-options-modal', (row) => {
      this.showOptionsModal(row);
    });
    window.addEventListener('resize', () => {
      this.setHOTHeight();
    });
  },
  mounted() {
    this.setHOTHeight();
  },
  updated() {
    var rowCount = this.$refs.hot.table.countRows();
    $('#rowCount').html(rowCount);
  },
  components: {
    HotTable,
  },
  directives: {
    focus,
  },
};
</script>

<style>
.activist-options-body a {
  color: #337ab7;
  cursor: pointer;
}
#activists-root {
  overflow: scroll;
}
.activist-options-btn {
  border: 0;
}
.activist-list-filters {
  margin: 10px 25px;
}
</style>
