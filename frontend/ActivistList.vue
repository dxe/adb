<template>
  <adb-page :title="title" :description="description" wide>
    <b-loading :is-full-page="true" v-model="loading"></b-loading>

    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-field label-position="on-border" label="Search by name">
            <b-input v-on:input="debounceSearchInput" type="text" icon="account-outline"></b-input>
          </b-field>
        </div>
        <div class="level-item">
          <b-button @click="showColumnsModal" type="is-primary" icon-left="table-headers-eye">
            Choose visible columns
          </b-button>
        </div>
        <div class="level-item" v-if="view === 'all_activists' || view === 'community_prospects'">
          <b-field>
            <b-switch v-model="showFilters" type="is-primary">Show filters</b-switch>
          </b-field>
        </div>
      </div>
      <div class="level-right">
        <div class="level-item px-1" v-if="view === 'chapter_member_development'">
          <b-button
            label="Export CSV for Spoke"
            type="is-info"
            icon-left="download"
            tag="a"
            href="/csv/chapter_member_spoke"
          ></b-button>
        </div>
        <div class="level-item px-1" v-if="view === 'community_prospects'">
          <b-button
            label="Export CSV for HubSpot"
            type="is-info"
            icon-left="download"
            tag="a"
            href="/csv/community_prospects_hubspot"
          ></b-button>
        </div>
        <div class="level-item has-text-centered px-1">
          <div>
            <p class="heading">Total</p>
            <p class="title">{{ rowCount }}</p>
          </div>
        </div>
      </div>
    </nav>

    <nav class="level mb-6">
      <div class="level-left">
        <div class="level-item" v-if="showFilters && view === 'all_activists'">
          <b-field label="Last Event From" label-position="on-border">
            <b-input v-model="lastEventDateFrom" type="date" icon="calendar-start"></b-input>
          </b-field>
        </div>
        <div class="level-item" v-if="showFilters && view === 'all_activists'">
          <b-field label="Last Event To" label-position="on-border">
            <b-input v-model="lastEventDateTo" type="date" icon="calendar-end"></b-input>
          </b-field>
        </div>
        <div class="level-item" v-if="showFilters && view === 'community_prospects'">
          <b-field label="Interest" label-position="on-border">
            <b-select v-model="filterInterest" icon="filter">
              <option
                v-for="interest in [
                  'All',
                  'Sanctuary',
                  'Community',
                  'Outreach',
                  'Protest',
                  'Trainings',
                ]"
                :value="interest"
                :key="interest"
              >
                {{ interest }}
              </option>
            </b-select>
          </b-field>
        </div>
      </div>
    </nav>

    <div id="container hot-table-container">
      <HotTable
        ref="hot"
        :root="root"
        :settings="hotSettings"
        :data="activists"
        :height="height"
      ></HotTable>
    </div>

    <b-modal
      :active="currentModalName === 'activist-options-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ currentActivist.name }}</p>
        </header>
        <section class="modal-card-body">
          <b-button
            @click="showModal('merge-activist-modal', currentActivist, activistIndex)"
            label="Merge Activist"
            type="is-primary"
          ></b-button>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Close" @click="hideModal" />
        </footer>
      </div>
    </b-modal>

    <b-modal
      :active="currentModalName === 'merge-activist-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">Merge activist</p>
        </header>
        <section class="modal-card-body">
          <p class="mb-2">Merging activists is used to combine redundant activist entries.</p>
          <p>Merging this activist does two things:</p>
          <ul class="mb-2">
            <li>
              - All of {{ currentActivist.name }}'s attendance data will be merged into the target
              activist.
            </li>
            <li>- {{ currentActivist.name }} will be hidden.</li>
          </ul>
          <p class="mb-5">
            Non-attendance data (e.g. email, location, etc) is <strong>NOT</strong> merged.
          </p>
          <p class="mb-2">
            <strong>Merge {{ currentActivist.name }} into another activist:</strong>
          </p>
          <p>
            Target activist:
            <b-select v-model="mergeTarget">
              <option v-for="name in activistMergeOptions" :value="name" :key="name">
                {{ name }}
              </option>
            </b-select>
          </p>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Cancel" @click="hideModal" />
          <b-button
            type="is-danger"
            :disabled="disableConfirmButton"
            @click="confirmMergeActivistModal"
          >
            Merge activist
          </b-button>
        </footer>
      </div>
    </b-modal>

    <b-modal
      :active="currentModalName === 'columns-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">Choose columns to display</p>
        </header>
        <section class="modal-card-body">
          <div v-for="(column, idx) in columns">
            <span v-if="column.header !== ''">
              <span v-if="columns[idx].category !== columns[idx - 1].category"
                ><p class="mt-3 mb-1">
                  <strong class="has-text-primary">{{ column.category }}</strong>
                </p></span
              >
              <b-field>
                <b-switch :id="column.header" v-model="column.enabled" type="is-primary" />
                {{ column.longHeader }}
              </b-field>
            </span>
          </div>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Close" @click="hideModal" />
        </footer>
      </div>
    </b-modal>
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import HotTable from './external/vue-handsontable-official/HotTable.vue';
import { rewriteSettings } from './external/vue-handsontable-official/helpers';
import AdbPage from './AdbPage.vue';
import { focus } from './directives/focus';
import {
  flashMessage,
  initializeFlashMessage,
  setFlashMessageSuccessCookie,
} from './flash_message';
import { EventBus } from './EventBus';
import debounce from 'debounce';
import moment from 'moment';

interface Activist {
  id: number;
  name: string;
  activist_level: string;
  active: number;
  prospect_organizer: number;
  prospect_chapter_member: number;
  dev_interest: string;
  circle_interest: number;
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
  category?: string;
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
    return callback(
      /^([0-9a-zA-Z]([-.\w]*[0-9a-zA-Z])*@(([0-9a-zA-Z])+([-\w]*[0-9a-zA-Z])*\.)+[a-zA-Z]{2,9})$/.test(
        value,
      ),
    );
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
    // Basic Info
    {
      header: 'ID',
      longHeader: 'ID',
      category: 'Basic Info',
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
      longHeader: 'Name',
      category: 'Basic Info',
      data: {
        data: 'name',
        colWidths: 150,
      },
      enabled: true,
    },
    {
      header: 'SMS Name',
      longHeader: 'SMS Name (First Name or Nickname)',
      category: 'Basic Info',
      data: {
        data: 'preferred_name',
        colWidths: 90,
      },
      enabled: view === 'chapter_member_development',
    },
    {
      header: 'Email',
      longHeader: 'Email',
      category: 'Basic Info',
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
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development' ||
        view === 'community_prospects',
    },
    {
      header: 'Phone',
      longHeader: 'Phone Number',
      category: 'Basic Info',
      data: {
        data: 'phone',
        colWidths: 100,
      },
      enabled:
        view === 'community_prospects' ||
        view === 'all_activists' ||
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development',
    },
    {
      header: 'Discord ID',
      longHeader: 'Discord ID',
      category: 'Basic Info',
      data: {
        data: 'discord_id',
        colWidths: 50,
      },
      enabled: false,
    },
    {
      header: 'Facebook',
      longHeader: 'Facebook URL',
      category: 'Basic Info',
      data: {
        data: 'facebook',
      },
      enabled: view === 'all_activists',
    },
    {
      header: 'Level',
      longHeader: 'Level',
      category: 'Basic Info',
      data: {
        data: 'activist_level',
        colWidths: 140,
        type: 'dropdown',
        source: ['Supporter', 'Chapter Member', 'Organizer', 'Non-Local', 'Global Network Member'],
      },
      enabled:
        view === 'all_activists' ||
        view === 'leaderboard' ||
        view === 'organizer_prospects' ||
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development',
    },
    {
      header: 'Birthday',
      longHeader: 'Birthday',
      category: 'Basic Info',
      data: {
        data: 'dob',
        colWidths: 100,
        type: 'date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
      },
      enabled: view === 'all_activists',
    },
    // Location
    {
      header: 'Street Address',
      longHeader: 'Street Address',
      category: 'Location',
      data: {
        data: 'street_address',
        colWidths: 100,
      },
      enabled: false,
    },
    {
      header: 'City',
      longHeader: 'City',
      category: 'Location',
      data: {
        data: 'city',
        colWidths: 100,
      },
      enabled: false,
    },
    {
      header: 'State',
      longHeader: 'State',
      category: 'Location',
      data: {
        data: 'state',
        colWidths: 100,
      },
      enabled: false,
    },
    {
      header: 'Zip Code',
      longHeader: 'Zip Code',
      category: 'Location',
      data: {
        data: 'location',
        colWidths: 100,
      },
      enabled: view === 'all_activists',
    },
    // Referral Info
    {
      header: 'Source',
      longHeader: 'Source',
      category: 'Referral Info',
      data: {
        data: 'source',
        colWidths: 100,
      },
      enabled: view === 'community_prospects',
    },
    {
      header: 'Interest Date',
      longHeader: 'Date Interest Form Submitted',
      category: 'Referral Info',
      data: {
        type: 'date',
        data: 'interest_date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
        readOnly: true,
      },
      enabled: view === 'community_prospects',
    },
    {
      header: 'Close Ties',
      longHeader: 'Close Ties',
      category: 'Referral Info',
      data: {
        data: 'referral_friends',
        colWidths: 100,
      },
      enabled: false,
    },

    {
      header: 'Referral',
      longHeader: 'Encouraged to Apply',
      category: 'Referral Info',
      data: {
        data: 'referral_apply',
        colWidths: 100,
      },
      enabled: view === 'community_prospects',
    },

    {
      header: 'Referral Outlet',
      longHeader: 'How did you hear about DxE?',
      category: 'Referral Info',
      data: {
        data: 'referral_outlet',
        colWidths: 100,
      },
      enabled: false,
    },
    // Application Info
    {
      header: 'Applied',
      longHeader: 'Application Date',
      category: 'Application Info',
      data: {
        type: 'date',
        data: 'dev_application_date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
        readOnly: true,
      },
      enabled: view === 'chapter_member_prospects' || view === 'organizer_prospects',
    },

    {
      header: 'Application',
      longHeader: 'Application Type',
      category: 'Application Info',
      data: {
        data: 'dev_application_type',
        colWidths: 80,
        readOnly: true,
      },
      enabled: view === 'chapter_member_prospects' || view === 'organizer_prospects',
    },

    {
      header: 'Prsp. Ch. Mem.',
      longHeader: 'Prospective Chapter Member',
      category: 'Application Info',
      data: {
        type: 'dropdown',
        source: [true, false],
        data: 'prospect_chapter_member',
        colWidths: 60,
      },
      enabled: view === 'chapter_member_prospects',
    },

    {
      header: 'Prsp. Organizer',
      longHeader: 'Prospective Organizer',
      category: 'Application Info',
      data: {
        type: 'dropdown',
        source: [true, false],
        data: 'prospect_organizer',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
    },
    // Circle Info
    {
      header: 'Geo-Circle',
      longHeader: 'Geo-Circle Membership',
      category: 'Circle Info',
      data: {
        data: 'geo_circles',
        readOnly: true,
        colWidths: 100,
      },
      enabled: view === 'chapter_member_development',
    },
    {
      header: 'Last Circle',
      longHeader: 'Last Circle Attended',
      category: 'Circle Info',
      data: {
        data: 'last_circle',
        type: 'date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        readOnly: true,
        colWidths: 100,
      },
      enabled: false,
    },
    // Event Attendance
    {
      header: 'First Event',
      longHeader: 'First Event',
      category: 'Event Attendance',
      data: {
        data: 'first_event_name',
        readOnly: true,
        colWidths: 200,
      },
      enabled: view === 'leaderboard' || view === 'community_prospects' || view === 'study',
    },
    {
      header: 'Last Event',
      longHeader: 'Last Event',
      category: 'Event Attendance',
      data: {
        data: 'last_event_name',
        readOnly: true,
        colWidths: 200,
      },
      enabled: view === 'leaderboard' || view === 'study',
    },
    {
      header: 'Total Events',
      longHeader: 'Total Events',
      category: 'Event Attendance',
      data: {
        type: 'numeric',
        data: 'total_events',
        readOnly: true,
        colWidths: 90,
      },
      enabled: view === 'leaderboard' || view === 'community_prospects' || view === 'study',
    },
    {
      header: 'Points',
      longHeader: 'Leaderboard Points',
      category: 'Event Attendance',
      data: {
        type: 'numeric',
        data: 'total_points',
        readOnly: true,
        colWidths: 50,
      },
      enabled: view === 'leaderboard',
    },
    {
      header: 'MPI',
      longHeader: 'MPI Status',
      category: 'Event Attendance',
      data: {
        type: 'checkbox',
        data: 'mpi',
        readOnly: true,
        colWidths: 30,
      },
      enabled:
        view === 'all_activists' ||
        view === 'leaderboard' ||
        view === 'development' ||
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development' ||
        view === 'organizer_prospects' ||
        view === 'development',
    },
    {
      header: 'DA&C Current Month',
      longHeader: 'Attended both a Direct Action & a Community event in current month',
      category: 'Event Attendance',
      data: {
        data: 'mpp_requirements',
        colWidths: 80,
        readOnly: true,
      },
      enabled: view === 'chapter_member_development',
    },
    // Trainings
    {
      header: 'Workshop',
      longHeader: 'Workshop',
      category: 'Trainings',
      data: {
        type: 'date',
        data: 'training0',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'chapter_member_prospects' || view === 'organizer_prospects',
    },
    {
      header: 'Consent & Oppress',
      longHeader: 'Consent & Anti-Oppression',
      category: 'Trainings',
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
      header: 'Prpsful Cmnty',
      longHeader: 'Building Purposeful Communities',
      category: 'Trainings',
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
      header: 'Ldshp & Mgmt',
      longHeader: 'Leadership and Management',
      category: 'Trainings',
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
      header: 'Vision & Strat',
      longHeader: 'Vision and Strategy',
      category: 'Trainings',
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
      header: 'Tier 2',
      longHeader: 'Tier II Protest',
      category: 'Trainings',
      data: {
        type: 'date',
        data: 'training_protest',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: false,
    },
    {
      header: 'Consent Refresh',
      longHeader: 'Consent Refresher Quiz',
      category: 'Trainings',
      data: {
        type: 'date',
        data: 'consent_quiz',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: false,
    },
    // Development
    {
      header: 'Quiz',
      longHeader: 'Organizer Quiz',
      category: 'Development',
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
      header: 'Coach',
      longHeader: 'Coach',
      category: 'Development',
      data: {
        data: 'connector',
        colWidths: 125,
      },
      enabled: view === 'development',
    },
    {
      header: 'Last Coaching',
      longHeader: 'Last Coaching',
      category: 'Development',
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
    // Chapter Membership
    {
      header: 'First Text',
      longHeader: 'First SMS Message Sent',
      category: 'Chapter Membership',
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
      longHeader: 'Approval Email Sent',
      category: 'Chapter Membership',
      data: {
        type: 'date',
        data: 'cm_approval_email',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        colWidths: 100,
      },
      enabled: view === 'chapter_member_prospects' || view === 'chapter_member_development',
    },
    {
      header: 'Vision Wall',
      longHeader: 'Vision Wall',
      category: 'Chapter Membership',
      data: {
        data: 'vision_wall',
        colWidths: 80,
        type: 'dropdown',
        source: ['', 'Yes', 'Declined'],
      },
      enabled: view === 'chapter_member_development',
    },
    {
      header: 'Voting Agreement',
      longHeader: 'Voting Agreement',
      category: 'Chapter Membership',
      data: {
        type: 'checkbox',
        data: 'voting_agreement',
        colWidths: 50,
      },
      enabled: view === 'chapter_member_development',
    },
    // Other
    {
      header: 'Notes',
      longHeader: 'Notes',
      category: 'Other',
      data: {
        data: 'notes',
        colWidths: 100,
      },
      enabled:
        view === 'organizer_prospects' ||
        view === 'development' ||
        view === 'chapter_member_development' ||
        view === 'chapter_member_prospects',
    },
    {
      header: 'Interests',
      longHeader: 'Interests',
      category: 'Other',
      data: {
        data: 'dev_interest',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects' || view === 'community_prospects',
    },
    {
      header: 'Hiatus',
      longHeader: 'Hiatus',
      category: 'Other',
      data: {
        type: 'checkbox',
        data: 'hiatus',
        colWidths: 50,
      },
      enabled: view === 'chapter_member_development',
    },
  ];
}

// Constants related to list ordering
// Corresponds to the constants DescOrder and AscOrder in model/activist.go
const AscOrder = 1;
const DescOrder = 2;

let previousSortData = {
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

  return !sortByDate;
}

// Call this after every sort.
function setPreviousSortData(field: string, ascending: boolean) {
  previousSortData.field = field;
  previousSortData.ascending = ascending;
}

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
  td.innerHTML = `<button
      data-role="trigger"
      class="mdi mdi-dots-horizontal activist-options-btn"
      type="button"
      onclick="window.showOptionsModal(${row})"></button>`;
  return td;
}

// Returns the first of the previous month.
function initialDateFromValue() {
  return moment()
    .subtract(1, 'months')
    .startOf('month')
    .format('YYYY-MM-DD');
}

// Returns the current date.
function initialDateToValue() {
  return moment().format('YYYY-MM-DD');
}

function generateBooleanSortFn(field: string, ascending: boolean) {
  return function(a: Activist, b: Activist) {
    const order = a[field] === b[field] ? 0 : Number(a[field]) - Number(b[field]);
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateStringSortFn(field: string, ascending: boolean) {
  return function(a: Activist, b: Activist) {
    const order = a[field].toLowerCase() < b[field].toLowerCase() ? -1 : 1;
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateGenericSortFn(field: string, ascending: boolean) {
  return function(a: Activist, b: Activist) {
    const order = a[field] < b[field] ? -1 : 1;
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

    const valueA = new Date(a[field]).getTime();
    const valueB = new Date(b[field]).getTime();

    const order = valueA < valueB ? -1 : 1;

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
    description: String,
    // `view` is the default view to show.
    view: {
      type: String,
    },
  },
  methods: {
    isMobile() {
      return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
        navigator.userAgent,
      );
    },
    showOptionsModal(row: number) {
      const activist = this.activists[row];
      this.showModal('activist-options-modal', activist, row);
    },
    showColumnsModal() {
      this.showModal('columns-modal');
    },
    showModal(modalName: string, activist: Activist = {} as Activist, index: number = 0) {
      // Hide the navbar so that the model doesn't go behind it.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'hidden';

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

        if (this.currentModalName == 'merge-activist-modal') {
          this.getActivistMergeOptions(this.currentActivist.name);
        }

        this.disableConfirmButton = false;
      });
    },
    hideModal() {
      // Show the navbar.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'visible';

      this.currentModalName = '';
      this.activistIndex = -1;
      this.currentActivist = {} as Activist;
      this.mergeTarget = '';
      this.activistMergeOptions = [];
    },
    confirmMergeActivistModal() {
      const targetActivistName = this.mergeTarget;
      if (!targetActivistName) {
        flashMessage('Must choose an activist to merge into', true);
        return;
      }

      this.disableConfirmButton = true;
      const currentActivistID = this.currentActivist.id;

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

          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          setFlashMessageSuccessCookie(
            this.currentActivist.name + ' was merged into ' + targetActivistName,
          );

          // Force page to refresh in order to pick
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
    loadActivists() {
      this.loading = true;

      $.ajax({
        url: '/activist/list',
        method: 'POST',
        data: JSON.stringify(this.listActivistsParameters()),
        success: (data) => {
          const parsed = JSON.parse(data);

          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }

          // status === "success"
          let activistList = parsed.activist_list;

          // frontend filtering
          if (this.view === 'community_prospects') {
            // TODO: handle this w/ Vue instead of jquery
            const selectedInterest = $('#filterInterest :selected').text();

            // only need to filer if an interest is selected
            if (selectedInterest != 'All' && selectedInterest != '' && selectedInterest != null) {
              let activistListFiltered: Activist[];
              activistListFiltered = activistList.filter((a: Activist) => {
                return a.dev_interest.toLowerCase().indexOf(selectedInterest.toLowerCase()) != -1;
              });
              activistList = activistListFiltered;
            }
          }

          if (activistList !== null) {
            this.allActivists = activistList;
          }

          this.loading = false;
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    debounceLoadActivists: debounce(function(this: any) {
      this.loadActivists();
    }, 500),
    afterChangeCallback(changes: any[], source: string) {
      if (
        source !== 'edit' &&
        source !== 'CopyPaste.paste' &&
        source !== 'UndoRedo.undo' &&
        source !== 'UndoRedo.redo'
      ) {
        return;
      }
      for (let i = 0; i < changes.length; i++) {
        const change = changes[i];
        const columnIndex = change[0];
        const columnName = change[1];
        const previousData = change[2];
        const newData = change[3];

        const activist = this.activists[columnIndex];
        (function(change) {
          // TODO: use change?
          $.ajax({
            url: '/activist/save',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(activist),
            success: (data) => {
              const parsed = JSON.parse(data);
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
      const hotContainer = document.getElementById('hot-table-container');
      if (!hotContainer) {
        this.height = 500;
        return;
      }
      const y = (hotContainer.getBoundingClientRect() as DOMRect).y;
      this.height = window.innerHeight - y;
    },
    listActivistsParameters() {
      let order_field: string;
      switch (this.view) {
        case 'leaderboard':
          order_field = 'total_points';
          break;
        case 'community_prospects':
          order_field = 'interest_date';
          break;
        default:
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
    refreshHOTData() {
      const table = this.hotTable;
      const newSettings = {
        data: rewriteSettings(this.activists),
      };
      table.updateSettings(newSettings, false);
    },
    sortColumn(col: Column) {
      const field = col.data.data;
      if (!field) {
        // Don't sort columns with no data field (e.g. the first
        // column).
        return;
      }

      const type = col.data.type;
      const ascending = shouldSortByAscending(field, type === 'date');

      let sortFunction;
      switch (type) {
        case 'date':
          sortFunction = generateDateSortFn(field, ascending);
          break;
        case 'numeric':
          sortFunction = generateGenericSortFn(field, ascending);
          break;
        case 'checkbox':
          sortFunction = generateBooleanSortFn(field, ascending);
          break;
        default:
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
        let visibleColIndex = coords.col;
        let foundCol;
        for (let i = 0; i < this.columns.length; i++) {
          let col = this.columns[i];
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
    debounceSearchInput: debounce(function(this: any, text: string) {
      this.search = text;
    }, 500),
    getActivistMergeOptions(ignoreActivistName?: string) {
      this.loading = true;
      $.ajax({
        url: '/activist_names/get',
        method: 'GET',
        dataType: 'json',
        success: (data) => {
          const activistNames = data.activist_names as string[];
          // TODO: maybe just check that it doesn't match before merging if this is slow to filter?
          this.activistMergeOptions = activistNames.filter((a) => {
            return a != ignoreActivistName;
          });
          this.loading = false;
        },
        error: () => {
          flashMessage('Error: could not load activist names', true);
          this.loading = false;
        },
      });
    },
  },
  data() {
    let initDateFrom = '';
    let initDateTo = '';
    if (this.view === ('all_activists' || 'leaderboard')) {
      initDateFrom = initialDateFromValue();
      initDateTo = initialDateToValue();
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
      filterRadius: '5',
      showFilters: false,
      search: '',
      loading: false,
      rowCount: 0,
      mergeTarget: '',
      activistMergeOptions: [] as string[],
    };
  },
  computed: {
    hotSettings(): object {
      const columns: Handsontable.GridSettings[] = [];
      const columnHeaders: string[] = [];
      for (let i = 0; i < this.columns.length; i++) {
        const col = this.columns[i];
        if (!col.enabled) {
          continue;
        }
        columns.push(this.columns[i].data);
        columnHeaders.push(this.columns[i].header);
      }

      const fixedCol = this.view === 'development' || this.view === 'organizer_prospects' ? 3 : 0;

      return {
        columns: columns,
        colHeaders: columnHeaders,
        rowHeaders:
          this.view === 'leaderboard' ||
          this.view === 'chapter_member_development' ||
          this.view === 'chapter_member_prospects' ||
          this.view === 'all_activists',
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
      // This search implementation is slow when we have lots of data.
      // Make it faster when that becomes an issue.
      if (this.search.length > 2) {
        const searchNameNormalized = this.search.trim().toLowerCase();
        return this.allActivists.filter((a) => {
          return a.name.toLowerCase().includes(searchNameNormalized);
        });
      }

      return this.allActivists;
    },
  },
  watch: {
    lastEventDateFrom() {
      this.debounceLoadActivists();
    },
    lastEventDateTo() {
      this.debounceLoadActivists();
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
    initializeFlashMessage();
  },
  mounted() {
    this.setHOTHeight();
  },
  updated() {
    this.rowCount = this.hotTable.countRows();
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
.modal {
  /* To make sure it's on top of HoT */
  z-index: 200;
}
#activists-root {
  overflow: scroll;
}
.activist-options-btn {
  border: 0;
  background-color: white;
  font-size: 14px;
}
.colHeader {
  font-size: 0.8em;
}
.htDimmed input[type='checkbox'] {
  filter: invert(20%);
}
.htNoWrap {
  font-size: 0.9em;
}
</style>
