<template>
  <adb-page :title="title" wide class="activist-list-content">
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
        v-if="view == 'all_activists' || view == 'activist_pool' || view == 'community_prospects'"
      >
        <span v-if="showOptions !== 'filters'">+</span
        ><span v-if="showOptions === 'filters'">-</span> Filters
      </button>
      <button class="btn-link" id="colFilterBtn" @click="toggleShowOptions('columns')">
        <span v-if="showOptions !== 'columns'">+</span
        ><span v-if="showOptions === 'columns'">-</span> Columns
      </button>

      <span>&nbsp;&nbsp;&nbsp;&nbsp;<b>Total rows: </b></span>

      <span v-if="!loading" id="rowCount">0</span>

      <span v-if="loading"><i>Loading...</i></span>

      <div v-if="showOptions === 'filters'">
        <div v-if="view == 'all_activists' || view == 'activist_pool'">
          <label>Last Event From:</label>
          <input v-model="lastEventDateFrom" class="form-control filter-margin" type="date" />
        </div>
        <div v-if="view == 'all_activists' || view == 'activist_pool'">
          <label>Last Event To:</label>
          <input v-model="lastEventDateTo" class="form-control filter-margin" type="date" />
        </div>
        <div v-if="view === 'community_prospects'">
          <label>Interest:</label>
          <select id="filterInterest" v-model="filterInterest" class="form-control filter-margin">
            <option>All</option>
            <option>Sanctuary</option>
            <option>Community</option>
            <option>Outreach</option>
            <option>Protest</option>
            <option>Trainings</option>
          </select>
        </div>
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
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import vmodal from 'vue-js-modal';
import HotTable from './external/vue-handsontable-official/HotTable.vue';
import { rewriteSettings } from './external/vue-handsontable-official/helpers';
import AdbPage from './AdbPage.vue';
import { focus } from './directives/focus';
import { flashMessage } from './flash_message';
import { EventBus } from './EventBus';
import { initActivistSelect } from './chosen_utils';
import debounce from 'debounce';

Vue.use(vmodal);

interface Activist {
  id: number;
  name: string;
  activist_level: string;
  active: number;
  prospect_organizer: number;
  prospect_chapter_member: number;
  circle_agreement: number;
  dev_interest: string;
  circle_interest: number;
  circles_list: string;
  total_events: number;
  source: string;
  interest_date: string;
  facebook: string;

  // To appease our clunky sorting functions.
  // TODO(mdempsky): Remove.
  [key: string]: any;
}

interface Column {
  header: string;
  longHeader?: string;
  data: Handsontable.GridSettings;
  enabled: boolean;
}

function emailValidator(value: string, callback: Function) {
  // Create a delay between the validation, and the
  // resulting state of the Column. i.e. When the user
  // clicks away from the column (or hits enter) the row
  // will be saved if valid, or the focus will be sent back
  // to the invalid column; the timeout here makes for a smoother
  // UI transition whichever the result of the validation.
  setTimeout(function() {
    if (
      /^([0-9a-zA-Z]([-.\w]*[0-9a-zA-Z])*@(([0-9a-zA-Z])+([-\w]*[0-9a-zA-Z])*\.)+[a-zA-Z]{2,9})$/.test(
        value,
      )
    ) {
      callback(true);
    } else {
      callback(false);
    }
  }, 250);
}

function getDefaultColumns(view: string): Column[] {
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
      header: 'Notes',
      longHeader: 'Notes',
      data: {
        data: 'notes',
        colWidths: 100,
      },
      enabled:
        view === 'organizer_prospects' ||
        view === 'development' ||
        view === 'circle_member_prospects' ||
        view === 'chapter_member_development' ||
        view === 'chapter_member_prospects',
    },
    {
      header: 'Managing',
      longHeader: 'Managing',
      data: {
        data: 'dev_manager',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects' || view === 'senior_organizer_prospects',
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
        validator: (email: string, cb: Function) => {
          // Allow saving empty email
          if (!email) {
            return cb(true);
          }

          emailValidator(email, (isValid: boolean) => {
            // Show invalid email error message
            if (!isValid) {
              flashMessage('Email field is invalid. Please provide a valid email address.', true);
            }

            return cb(isValid);
          });
        },
        allowInvalid: false,
      },
      enabled:
        view === 'all_activists' ||
        view === 'activist_pool' ||
        view === 'activist_recruitment' ||
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development' ||
        view === 'circle_member_prospects' ||
        view === 'circle_members' ||
        view === 'community_prospects',
    },
    {
      header: 'Phone',
      longHeader: 'Phone Number',
      data: {
        data: 'phone',
        colWidths: 100,
      },
      enabled:
        view === 'community_prospects' ||
        view === 'all_activists' ||
        view === 'chapter_member_prospects',
    },
    {
      header: 'Birthday',
      longHeader: 'Birthday',
      data: {
        data: 'dob',
        colWidths: 100,
        type: 'date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
      },
      enabled: view === 'all_activists',
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
        view === 'circle_member_prospects' ||
        view === 'senior_organizer_prospects' ||
        view === 'circle_members',
    },

    {
      header: 'Source',
      longHeader: 'Source',
      data: {
        data: 'source',
        colWidths: 100,
      },
      enabled: view === 'community_prospects',
    },

    {
      header: 'Interest Date',
      longHeader: 'Date Interest Form Submitted',
      data: {
        type: 'date',
        data: 'interest_date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
        readOnly: true,
      },
      enabled: view === 'circle_member_prospects' || view === 'community_prospects',
    },

    {
      header: 'Circle Interest',
      longHeader: 'Circle Interest',
      data: {
        type: 'dropdown',
        source: [true, false],
        data: 'circle_interest',
        colWidths: 80,
      },
      enabled: view === 'circle_member_prospects',
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
        view === 'chapter_member_prospects' ||
        view === 'organizer_prospects' ||
        view === 'senior_organizer_prospects',
    },

    {
      header: 'Application',
      longHeader: 'Application Type',
      data: {
        data: 'dev_application_type',
        colWidths: 80,
        readOnly: true,
      },
      enabled:
        view === 'chapter_member_prospects' ||
        view === 'organizer_prospects' ||
        view === 'senior_organizer_prospects',
    },

    {
      header: 'Circle Agreement',
      longHeader: 'Circle Agreement',
      data: {
        type: 'checkbox',
        data: 'circle_agreement',
        colWidths: 105,
      },
      enabled: view === 'circle_member_prospects' || view === 'circle_members',
    },

    {
      header: 'Prosp. Ch. Member',
      longHeader: 'Prospective Chapter Member',
      data: {
        type: 'dropdown',
        source: [true, false],
        data: 'prospect_chapter_member',
        colWidths: 110,
      },
      enabled: false,
    },

    {
      header: 'Prosp. Organizer',
      longHeader: 'Prospective Organizer',
      data: {
        type: 'dropdown',
        source: [true, false],
        data: 'prospect_organizer',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },

    {
      header: 'Prosp. Sen. Org.',
      longHeader: 'Prospective Senior Organizer',
      data: {
        type: 'dropdown',
        source: [true, false],
        data: 'prospect_senior_organizer',
        colWidths: 100,
        readOnly: true,
      },
      enabled: view === 'senior_organizer_prospects' || view === 'development',
    },

    {
      header: 'Contacted',
      longHeader: 'Circle Membership: Date Contacted',
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
      header: 'Interests',
      longHeader: 'Interests',
      data: {
        data: 'dev_interest',
        colWidths: 100,
      },
      enabled:
        view === 'organizer_prospects' ||
        view === 'circle_member_prospects' ||
        view === 'community_prospects',
    },
    {
      header: 'Working Groups',
      longHeader: 'Working Group Membership',
      data: {
        data: 'working_group_list',
        readOnly: true,
        colWidths: 200,
      },
      enabled: view === 'action_team' || view === 'organizer_prospects' || view === 'development',
    },
    {
      header: 'Circles',
      longHeader: 'Circle Membership',
      data: {
        data: 'circles_list',
        readOnly: true,
        colWidths: 200,
      },
      enabled:
        view === 'circle_members' || view === 'organizer_prospects' || view === 'development',
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
        view === 'activist_pool' ||
        view === 'activist_recruitment' ||
        view === 'leaderboard' ||
        view === 'community_prospects',
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
      header: 'Last Circle',
      longHeader: 'Last Circle Attended',
      data: {
        data: 'last_circle',
        type: 'date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        readOnly: true,
        colWidths: 100,
      },
      enabled: view === 'circle_members',
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
      enabled: view === 'leaderboard' || view === 'community_prospects',
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
        view === 'chapter_member_development' ||
        view === 'organizer_prospects' ||
        view === 'senior_organizer_development' ||
        view === 'development',
    },
    {
      header: 'Connector',
      longHeader: 'Connector Name',
      data: {
        data: 'connector',
        colWidths: 125,
      },
      enabled:
        view === 'activist_pool' ||
        view === 'action_team' ||
        view === 'development' ||
        view === 'senior_organizer_development',
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
      enabled:
        view === 'action_team' ||
        view === 'chapter_member_prospects' ||
        view === 'organizer_prospects',
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
      enabled: view === 'development' || view === 'organizer_prospects',
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
      enabled: view === 'development' || view === 'organizer_prospects',
    },
    {
      header: 'Talk About AR',
      longHeader: 'Date Attended Training: How To Talk About Animal Rights',
      data: {
        type: 'date',
        data: 'training3',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'development' || view === 'organizer_prospects',
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
      enabled: view === 'development' || view === 'organizer_prospects',
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
      enabled: view === 'development' || view === 'organizer_prospects',
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
      enabled: view === 'development' || view === 'organizer_prospects',
    },

    {
      header: 'Quiz',
      longHeader: 'Organizer: Quiz',
      data: {
        type: 'date',
        data: 'dev_quiz',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects' || view === 'development',
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
      header: 'MPI Email',
      longHeader: 'Chapter Membership: Date MPI Warning Email Sent',
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
      header: 'Vision Wall',
      longHeader: 'Chapter Member: Added to Vision Wall',
      data: {
        data: 'vision_wall',
        colWidths: 80,
        type: 'dropdown',
        source: ['', 'Yes', 'Declined'],
      },
      enabled: view === 'chapter_member_development',
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
      header: 'Onboarding',
      longHeader: 'Organizer: Date of Onboarding Logistics Completion',
      data: {
        type: 'checkbox',
        data: 'dev_onboarding',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },

    {
      header: 'Point Auth.',
      longHeader: 'Senior Organizer: Point Auth',
      data: {
        type: 'date',
        data: 'so_auth',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'senior_organizer_prospects',
    },
    {
      header: 'Sent to Core',
      longHeader: 'Senior Organizer: Application Sent to Core',
      data: {
        type: 'date',
        data: 'so_core',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'senior_organizer_prospects',
    },
    {
      header: 'Agreement',
      longHeader: 'Senior Organizer: Signed Agreement',
      data: {
        type: 'checkbox',
        data: 'so_agreement',
        colWidths: 70,
      },
      enabled: view === 'senior_organizer_prospects',
    },
    {
      header: 'Training',
      longHeader: 'Senior Organizer: Completed Training',
      data: {
        type: 'date',
        data: 'so_training',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'senior_organizer_prospects',
    },
    {
      header: 'Quiz',
      longHeader: 'Senior Organizer: Passed Quiz',
      data: {
        type: 'date',
        data: 'so_quiz',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'senior_organizer_prospects',
    },
    {
      header: 'Connectees',
      longHeader: 'Senior Organizer: Connectees (3+)',
      data: {
        data: 'so_connector',
        colWidths: 120,
      },
      enabled: view === 'senior_organizer_prospects' || view === 'senior_organizer_development',
    },
    {
      header: 'Onboarding',
      longHeader: 'Senior Organizer: Completed Onboarding Logistics',
      data: {
        type: 'checkbox',
        data: 'so_onboarding',
        colWidths: 70,
      },
      enabled: view === 'senior_organizer_prospects',
    },

    {
      header: 'Close Ties',
      longHeader: 'Referral: Close Ties',
      data: {
        data: 'referral_friends',
        colWidths: 100,
      },
      enabled: false,
    },

    {
      header: 'Referral',
      longHeader: 'Referral: Encouraged to Apply',
      data: {
        data: 'referral_apply',
        colWidths: 100,
      },
      enabled: view === 'community_prospects',
    },

    {
      header: 'Referral Outlet',
      longHeader: 'Referral: How did you hear about us?',
      data: {
        data: 'referral_outlet',
        colWidths: 100,
      },
      enabled: false,
    },
  ];
}

// Constants related to list ordering
// Corresponds to the constants DescOrder and AscOrder in model/activist.go
const DescOrder = 2;
const AscOrder = 1;

var previousSortData = {
  field: '',
  ascending: false,
};

// Uses previousSortData to determine whether the next sort should be
// ascending.
//
// If sortByDate is true, then the default is to sort by descending.
// Otherwise, the default is to sort by ascending.
function shouldSortByAscending(field: string, sortByDate: boolean) {
  if (field == previousSortData.field) {
    return !previousSortData.ascending;
  }

  if (sortByDate) {
    return false;
  }
  return true;
}

// Call this after every sort.
function setPreviousSortData(field: string, ascending: boolean) {
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

(window as any).showOptionsModal = function(row: number) {
  EventBus.$emit('activist-show-options-modal', row);
};

function optionsButtonRenderer(
  instance: any,
  td: HTMLElement,
  row: number,
  col: number,
  prop: any,
  value: any,
  cellProperties: any,
) {
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

function generateBooleanSortFn(field: string, ascending: boolean) {
  return function(a: Activist, b: Activist) {
    var order = a[field] === b[field] ? 0 : Number(a[field]) - Number(b[field]);
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateStringSortFn(field: string, ascending: boolean) {
  return function(a: Activist, b: Activist) {
    var order = a[field].toLowerCase() < b[field].toLowerCase() ? -1 : 1;
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateGenericSortFn(field: string, ascending: boolean) {
  return function(a: Activist, b: Activist) {
    var order = a[field] < b[field] ? -1 : 1;
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateDateSortFn(field: string, ascending: boolean) {
  return function(a: Activist, b: Activist) {
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

export default Vue.extend({
  name: 'activist-list',
  props: {
    title: String,
    // `view` is the default view to show. It can be one of:
    // "all_activists", "leaderboard", "activist_pool",
    // "activist_recruitment", or "action_team"
    view: {
      type: String,
      /*validator(value) {
        var validViews = [
          'all_activists',
          'leaderboard',
          'activist_pool',
          'activist_recruitment',
          'action_team',
          'development',
        ];
        return validViews.indexOf(value) !== -1;
      },*/
    },
  },
  methods: {
    showOptionsModal(row: number) {
      var activist = this.activists[row];
      this.showModal('activist-options-modal', activist, row);
    },
    showModal(modalName: string, activist: Activist, index: number) {
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
      this.currentActivist = {} as Activist;
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
        var interval: number;
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
    removeActivist(id: number) {
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
          //this.removeActivist(currentActivistID);

          //this.hideModal();

          // Force view to refresh in order to pick
          // up changes to target activist data.
          // TOOD: Handle this better. Perhaps return updated
          // activist from backend call.
          // This is necessary because it doesn't appear that
          // the Vue.js component is handling data refreshing
          // properly; e.g. the "facebook" field is not updated.
          location.reload();
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
      this.loading = true;

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

          // frontend filtering

          if (this.view === 'community_prospects') {
            var selectedInterest = $('#filterInterest :selected').text();

            // only need to filer if an interest is selected
            if (selectedInterest != 'All' && selectedInterest != '' && selectedInterest != null) {
              var activistListFiltered;
              activistListFiltered = activistList.filter((el: Activist) => {
                return el.dev_interest.toLowerCase().indexOf(selectedInterest.toLowerCase()) != -1;
              });
              activistList = activistListFiltered;
            }
          }

          if (activistList !== null) {
            this.allActivists = activistList;
            this.loading = false;
          }
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    afterChangeCallback(changes: any[], source: string) {
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
      var y = (hotContainer.getBoundingClientRect() as DOMRect).y;
      this.height = window.innerHeight - y;
    },
    listActivistsParameters() {
      var order_field = 'last_event';

      if (this.view === 'leaderboard') {
        order_field = 'total_points';
      } else if (this.view === 'community_prospects') {
        order_field = 'interest_date';
      } else {
        order_field = 'last_event';
      }

      return {
        order: DescOrder,
        order_field: order_field,
        last_event_date_to: this.lastEventDateTo,
        last_event_date_from: this.lastEventDateFrom,
        filter: this.view, // this passes view to the backend, where filtering will now take place
      };
    },
    toggleShowOptions(optionsType: string) {
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
      var table = this.hotTable;
      var newSettings = {
        data: rewriteSettings(this.activists),
      };
      table.updateSettings(newSettings, false);
    },
    sortColumn(col: Column) {
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
    afterOnCellMouseDownCallback(event: any, coords: any, td: any) {
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
    // TODO(mdempsky): Remove "this: any".
    debounceSearchInput: debounce(function(this: any, e: Event) {
      this.search = (e.target as HTMLInputElement).value;
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
      currentActivist: {} as Activist,
      disableConfirmButton: false,
      allActivists: [] as Activist[],
      height: 500,
      columns: getDefaultColumns(this.view),
      lastEventDateFrom: initDateFrom,
      lastEventDateTo: initDateTo,
      filterInterest: 'All',
      showOptions: '',
      search: '',
      loading: false,
    };
  },
  computed: {
    hotSettings(): object {
      const columns: Handsontable.GridSettings[] = [];
      const columnHeaders: string[] = [];
      for (var i = 0; i < this.columns.length; i++) {
        var col = this.columns[i];
        if (!col.enabled) {
          continue;
        }
        columns.push(this.columns[i].data);
        columnHeaders.push(this.columns[i].header);
      }

      if (this.view === 'development' || this.view === 'organizer_prospects') {
        var fixedCol = 3;
      } else {
        var fixedCol = 0;
      }

      return {
        columns: columns,
        colHeaders: columnHeaders,
        rowHeaders:
          this.view === 'leaderboard' ||
          this.view === 'chapter_member_development' ||
          this.view === 'chapter_member_prospects',
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
        fixedColumnsLeft: fixedCol, // this causes havoc
      };
    },
    hotTable(): Handsontable {
      return (this.$refs.hot as any).table as Handsontable;
    },
    activists(): Activist[] {
      if (this.search.length < 3) {
        return this.allActivists;
      }

      // This search implementation is slow when we have lots of data.
      // Make it faster when that becomes an issue.
      var searchNormalized = this.search.trim().toLowerCase();
      var activists: Activist[] = [];
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
    filterInterest() {
      this.loadActivists();
    },
  },
  created() {
    this.loadActivists();
    EventBus.$on('activist-show-options-modal', (row: number) => {
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
    var rowCount = this.hotTable.countRows();
    if (rowCount == 0) {
      $('#rowCount').html(String('No data'));
    } else {
      $('#rowCount').html(String(rowCount));
    }
  },
  components: {
    AdbPage,
    HotTable,
  },
  directives: {
    focus,
  },
});
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
