<template>
  <adb-page :title="title" :description="description" wide>
    <b-loading :is-full-page="true" v-model="loading"></b-loading>

    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-field label-position="on-border" label="Search by name">
            <b-input v-model="search" type="text" icon="account-outline"></b-input>
          </b-field>
        </div>
        <div class="level-item">
          <b-button @click="showColumnsModal" type="is-primary" icon-left="table-headers-eye">
            Choose visible columns
          </b-button>
        </div>
        <div
          class="level-item"
          v-if="
            view === 'all_activists' ||
            view === 'community_prospects' ||
            view === 'new_activists' ||
            view === 'new_activists_pending_workshop'
          "
        >
          <b-field>
            <b-switch v-model="showFilters" type="is-primary">Show date filters</b-switch>
          </b-field>
        </div>
        <div class="level-item" v-if="view === 'community_prospects_followup'">
          <b-field label="Due Date" label-position="on-border">
            <b-select v-model="upcomingFollowupsOnly">
              <option
                v-for="t in [
                  { value: false, text: 'Today or Overdue' },
                  { value: true, text: 'Upcoming' },
                ]"
                :value="t.value"
                :key="t.value"
              >
                {{ t.text }}
              </option>
            </b-select>
          </b-field>
        </div>
        <div class="level-item" v-if="view === 'community_prospects_followup'">
          <b-field>
            <b-switch v-model="assignedToCurrentUser" type="is-primary"
              >Only show prospects assigned to me</b-switch
            >
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
        <!--        TODO: add date params to url-->
        <div class="level-item px-1" v-if="view === 'all_activists'">
          <b-button
            label="Export Supporter CSV for Spoke"
            type="is-info"
            icon-left="download"
            tag="a"
            :href="`/csv/all_activists_spoke?start_date=${lastEventDateFrom}&end_date=${lastEventDateTo}`"
          ></b-button>
        </div>
        <div class="level-item px-1" v-if="view === 'new_activists'">
          <b-button
            label="Export CSV for Spoke"
            type="is-info"
            icon-left="download"
            tag="a"
            :href="`/csv/new_activists_spoke?start_date=${lastEventDateFrom}&end_date=${lastEventDateTo}`"
          ></b-button>
        </div>
        <div class="level-item px-1" v-if="view === 'new_activists_pending_workshop'">
          <b-button
            label="Export CSV for Spoke"
            type="is-info"
            icon-left="download"
            tag="a"
            :href="`/csv/new_activists_pending_workshop_spoke?start_date=${lastEventDateFrom}&end_date=${lastEventDateTo}`"
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
        <div
          class="level-item"
          v-if="
            showFilters &&
            (view === 'all_activists' ||
              view === 'new_activists' ||
              view === 'new_activists_pending_workshop')
          "
        >
          <b-field label="Last Event From" label-position="on-border">
            <b-input v-model="lastEventDateFrom" type="date" icon="calendar-start"></b-input>
          </b-field>
        </div>
        <div
          class="level-item"
          v-if="
            showFilters &&
            (view === 'all_activists' ||
              view === 'new_activists' ||
              view === 'new_activists_pending_workshop')
          "
        >
          <b-field label="Last Event To" label-position="on-border">
            <b-input v-model="lastEventDateTo" type="date" icon="calendar-end"></b-input>
          </b-field>
        </div>

        <div class="level-item" v-if="showFilters && view === 'community_prospects'">
          <b-field label="Interest Date From" label-position="on-border">
            <b-input v-model="interestDateFrom" type="date" icon="calendar-start"></b-input>
          </b-field>
        </div>
        <div class="level-item" v-if="showFilters && view === 'community_prospects'">
          <b-field label="Interest Date To" label-position="on-border">
            <b-input v-model="interestDateTo" type="date" icon="calendar-end"></b-input>
          </b-field>
        </div>
      </div>
    </nav>

    <div id="container hot-table-container">
      <HotTable
        ref="hot"
        :root="root"
        :settings="hotSettings"
        :data="allActivists"
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
          <div class="mb-5">
            <h1 class="subtitle has-text-primary mb-4" style="font-weight: 500">Info</h1>

            <b-field label="Pronouns" label-position="on-border" v-if="currentActivist.pronouns">
              <b-input
                type="text"
                custom-class="has-text-dark"
                v-model="currentActivist.pronouns"
                expanded
                disabled
              ></b-input>
            </b-field>

            <b-field label="Email" label-position="on-border" v-if="currentActivist.email">
              <b-input
                type="text"
                custom-class="has-text-dark"
                v-model="currentActivist.email"
                expanded
                disabled
              ></b-input>
            </b-field>

            <b-field label="Phone" label-position="on-border" v-if="currentActivist.phone">
              <b-input
                type="text"
                custom-class="has-text-dark"
                v-model="currentActivist.phone"
                expanded
                disabled
              ></b-input>
            </b-field>

            <b-field label="Source" label-position="on-border" v-if="currentActivist.source">
              <b-input
                type="text"
                custom-class="has-text-dark"
                v-model="currentActivist.source"
                expanded
                disabled
              ></b-input>
            </b-field>

            <b-field
              label="Interest Date"
              label-position="on-border"
              v-if="currentActivist.interest_date"
            >
              <b-input
                type="text"
                custom-class="has-text-dark"
                v-model="currentActivist.interest_date"
                expanded
                disabled
              ></b-input>
            </b-field>

            <b-field label="Notes" label-position="on-border" v-if="currentActivist.notes">
              <b-input
                type="textarea"
                custom-class="has-text-dark"
                v-model="currentActivist.notes"
                expanded
                disabled
              ></b-input>
            </b-field>
          </div>

          <div class="mb-5">
            <h1 class="subtitle has-text-primary mb-4" style="font-weight: 500">Interactions</h1>

            <div v-if="currentActivist.interactions && currentActivist.interactions.length === 0">
              None
            </div>

            <div class="card mb-5" v-for="interaction in currentActivist.interactions">
              <header class="card-header">
                <p class="card-header-title" style="width: 100%">
                  {{ interaction.method ? interaction.method : 'Interaction' }} by
                  {{ interaction.user_name }}
                </p>
                <p class="has-text-grey pt-3" style="font-size: 0.8em; width: 120px">
                  {{ formatInteractionDate(interaction.timestamp) }}
                </p>
              </header>

              <div class="card-content">
                <div class="content">
                  <p>{{ interaction.notes }}</p>
                  <p v-if="interaction.outcome">
                    <span class="tag is-info">Outcome: {{ interaction.outcome }}</span>
                  </p>
                </div>
              </div>

              <footer class="card-footer">
                <b-button
                  icon-left="pencil"
                  class="is-small is-primary is-inverted mt-1 card-footer-item"
                  @click="
                    showModal('edit-interaction-modal', currentActivist, activistIndex, interaction)
                  "
                  >Edit</b-button
                >
                &nbsp;
                <b-button
                  icon-left="delete"
                  class="is-small is-danger is-inverted mt-1 card-footer-item"
                  @click="deleteInteraction(interaction)"
                  >Delete</b-button
                >
              </footer>
            </div>
          </div>
        </section>
        <footer class="modal-card-foot">
          <b-button icon-left="close" label="Close" @click="hideModal" />
          <b-button
            icon-left="plus"
            @click="showModal('edit-interaction-modal', currentActivist, activistIndex)"
            label="Add interaction"
            type="is-primary"
          ></b-button>
          <b-button
            icon-left="eye-off-outline"
            @click="confirmHideActivist(currentActivist.id)"
            label="Hide"
            type="is-danger"
          ></b-button>
          <b-button
            icon-left="merge"
            @click="showModal('merge-activist-modal', currentActivist, activistIndex)"
            label="Merge"
            type="is-warning"
          ></b-button>
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
            <b-taginput
              v-model="mergeTarget"
              :data="filteredActivistMergeOptions"
              autocomplete
              :allow-new="false"
              icon="account-outline"
              @typing="getFilteredActivistMergeOptions"
              maxtags="1"
              type="is-info"
              dropdown-position="top"
              :has-counter="false"
            ></b-taginput>
          </p>
        </section>
        <footer class="modal-card-foot">
          <b-button icon-left="cancel" label="Cancel" @click="hideModal" />
          <b-button
            icon-left="alert"
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
      :active="currentModalName === 'edit-interaction-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">Interaction: {{ currentActivist.name }}</p>
        </header>
        <section class="modal-card-body">
          <b-field label="Timestamp" label-position="on-border" v-if="currentInteraction.timestamp">
            <b-input type="text" v-model="currentInteraction.timestamp" expanded disabled></b-input>
          </b-field>

          <b-field label="Method" label-position="on-border">
            <b-select v-model.trim="currentInteraction.method" expanded placeholder="Choose one">
              <option v-for="x in ['SMS', 'Call', 'Email']" :value="x" :key="x">
                {{ x }}
              </option>
            </b-select>
          </b-field>

          <b-field label="Outcome" label-position="on-border">
            <b-select v-model.trim="currentInteraction.outcome" expanded>
              <option
                v-for="x in ['No answer', 'Had conversation', 'Wrong number', 'Left message']"
                :value="x"
                :key="x"
              >
                {{ x }}
              </option>
            </b-select>
          </b-field>

          <b-field label="Notes" label-position="on-border">
            <b-input type="textarea" v-model.trim="currentInteraction.notes" expanded></b-input>
          </b-field>

          <div v-if="!currentInteraction.id">
            <b-field>
              <b-checkbox v-model.trim="currentInteraction.assign_self">Assign to me</b-checkbox>
            </b-field>
            <b-field label="Follow-up" custom-class="has-text-primary"> </b-field>
            <b-field>
              <b-checkbox
                v-model="currentInteraction.reset_followup"
                :disabled="currentInteraction.set_followup"
                >Clear follow-up date</b-checkbox
              >
            </b-field>
            <b-field>
              <b-checkbox
                v-model="currentInteraction.set_followup"
                :disabled="currentInteraction.reset_followup"
              >
                <b-field>
                  <p class="control">
                    <b-button label="Follow-up in" class="is-static has-text-grey-dark" />
                  </p>
                  <b-numberinput
                    v-model="currentInteraction.followup_days"
                    controls-position="compact"
                    min="1"
                    max="365"
                  />
                  <p class="control">
                    <b-button label="days" class="is-static has-text-grey-dark" />
                  </p>
                </b-field>
              </b-checkbox>
            </b-field>
          </div>
        </section>
        <footer class="modal-card-foot">
          <b-button icon-left="cancel" label="Cancel" @click="hideModal" />
          <b-button
            icon-left="floppy"
            type="is-primary"
            :disabled="disableConfirmButton"
            @click="confirmEditInteractionModal"
          >
            Save
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
          <div
            v-for="(columns, category) in groupBy(columns, 'category')"
            class="mb-3"
            v-if="category !== 'Options'"
          >
            <div class="has-text-primary mb-1 has-text-weight-bold">{{ category }}</div>
            <div v-for="column in columns">
              <b-field class="mb-1">
                <b-switch :id="column.header" v-model="column.enabled" type="is-primary" />
                {{ column.longHeader }}
              </b-field>
            </div>
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
import * as dayjs from 'dayjs';

interface Activist {
  id: number;
  name: string;
  pronouns: string;
  email: string;
  phone: string;
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
  followup_date: string;
  notes: string;
  interactions: Interaction[];

  // To appease our clunky sorting functions.
  // TODO(mdempsky): Remove.
  [key: string]: any;
}

interface Interaction {
  id: number;
  activist_id: number;
  user_id: number;
  user_name: string;
  timestamp: string;
  method: string;
  outcome: string;
  notes: string;
  reset_followup: boolean;
  set_followup: boolean;
  followup_days: number;
  assign_self: boolean;
}

interface Column {
  header: string;
  longHeader?: string;
  category?: string;
  data: Handsontable.GridSettings;
  enabled: boolean;
  showForAllChapters?: boolean;
}

function emailValidator(value: string, callback: Function) {
  // Create a delay between the validation, and the
  // resulting state of the Column. i.e. When the user
  // clicks away from the column (or hits enter) the row
  // will be saved if valid, or the focus will be sent back
  // to the invalid column; the timeout here makes for a smoother
  // UI transition whichever the result of the validation.
  setTimeout(function () {
    return callback(
      /^([0-9a-zA-Z]([-.\w]*[0-9a-zA-Z])*@(([0-9a-zA-Z])+([-\w]*[0-9a-zA-Z])*\.)+[a-zA-Z]{2,9})$/.test(
        value,
      ),
    );
  }, 250);
}

function getColumnsForChapter(chapter: string, view: string): Column[] {
  const cols = getDefaultColumns(chapter, view);
  if (chapter === 'SF Bay Area') return cols;
  return cols.filter((c) => {
    return !!c.showForAllChapters && c.showForAllChapters;
  });
}

function getDefaultColumns(chapter: string, view: string): Column[] {
  return [
    {
      header: '',
      category: 'Options',
      data: {
        renderer: optionsButtonRenderer,
        readOnly: true,
        disableVisualSelection: true,
        colWidths: 35,
      },
      enabled: true,
      showForAllChapters: true,
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
      showForAllChapters: true,
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
      showForAllChapters: true,
    },
    {
      header: 'Pronouns',
      longHeader: 'Pronouns',
      category: 'Basic Info',
      data: {
        data: 'pronouns',
        colWidths: 50,
      },
      enabled: true,
      showForAllChapters: true,
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
        view === 'community_prospects' ||
        view === 'community_prospects_followup' ||
        view === 'new_activists' ||
        view === 'new_activists_pending_workshop',
      showForAllChapters: true,
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
        view === 'community_prospects_followup' ||
        view === 'all_activists' ||
        view === 'chapter_member_prospects' ||
        view === 'chapter_member_development' ||
        view === 'new_activists' ||
        view === 'new_activists_pending_workshop',
      showForAllChapters: true,
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
      enabled:
        (view === 'all_activists' ||
          view === 'new_activists' ||
          view === 'new_activists_pending_workshop') &&
        chapter === 'SF Bay Area',
      showForAllChapters: true,
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
        view === 'chapter_member_development' ||
        view === 'new_activists' ||
        view === 'new_activists_pending_workshop',
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
      enabled: view === 'all_activists' && chapter === 'SF Bay Area',
      showForAllChapters: true,
    },
    {
      header: 'Accessibility',
      longHeader: 'Accessibility Needs',
      category: 'Basic Info',
      data: {
        data: 'accessibility',
        colWidths: 50,
      },
      enabled: false,
      showForAllChapters: true,
    },
    {
      header: 'Language',
      longHeader: 'Primary Language',
      category: 'Basic Info',
      data: {
        data: 'language',
        colWidths: 50,
      },
      enabled: false,
      showForAllChapters: true,
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
        view === 'chapter_member_prospects' ||
        view === 'community_prospects' ||
        view === 'community_prospects_followup',
      showForAllChapters: true,
    },
    {
      header: 'Interests',
      longHeader: 'Interests',
      category: 'Other',
      data: {
        data: 'dev_interest',
        colWidths: 100,
      },
      enabled: view === 'organizer_prospects',
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
      showForAllChapters: true,
    },
    {
      header: 'City',
      longHeader: 'City',
      category: 'Location',
      data: {
        data: 'city',
        colWidths: 100,
      },
      enabled: view === 'all_activists',
      showForAllChapters: true,
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
      showForAllChapters: true,
    },
    {
      header: 'Zip Code',
      longHeader: 'Zip Code',
      category: 'Location',
      data: {
        data: 'location',
        colWidths: 100,
      },
      enabled:
        (view === 'all_activists' || view === 'new_activists_pending_workshop') &&
        chapter === 'SF Bay Area',
      showForAllChapters: true,
    },
    // Prospect Info
    {
      header: 'Assigned To',
      longHeader: 'Assigned To',
      category: 'Prospect Info',
      data: {
        type: 'dropdown',
        // TODO: try to only get the user list only on page load instead of everytime the dropdown is clicked
        //@ts-ignore
        source: function (query: any, process: any) {
          $.ajax({
            url: '/user/list',
            success: (data) => {
              const parsed = JSON.parse(data);

              let users = parsed
                .filter((user: any) => {
                  return (
                    user.disabled === false &&
                    user.roles &&
                    (user.roles.includes('admin') || user.roles.includes('organizer'))
                  );
                })
                .map((user: any) => {
                  return user.name;
                })
                .sort();

              users.unshift('');

              if (query.length > 0) {
                users = users.filter((user: any) => {
                  return user.toLowerCase().startsWith(query.toLowerCase());
                });
              }
              process(users);
            },
            error: () => {
              flashMessage('Error getting user names from server.', true);
            },
          });
        },
        mode: 'strict',
        data: 'assigned_to_name',
        colWidths: 100,
      },
      enabled: view === 'community_prospects' || view === 'community_prospects_followup',
    },
    {
      header: 'Follow-up',
      longHeader: 'Date to Follow-up',
      category: 'Prospect Info',
      data: {
        data: 'followup_date',
        colWidths: 110,
        type: 'date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
      },
      enabled: view === 'community_prospects' || view === 'community_prospects_followup',
    },
    {
      header: 'Interactions',
      longHeader: 'Total Number of Interactions',
      category: 'Prospect Info',
      data: {
        data: 'total_interactions',
        colWidths: 80,
        type: 'numeric',
        readOnly: true,
      },
      enabled: view === 'community_prospects' || view === 'community_prospects_followup',
    },
    {
      header: 'Last Interaction',
      longHeader: 'Last Interaction Date',
      category: 'Prospect Info',
      data: {
        data: 'last_interaction_date',
        colWidths: 100,
        type: 'date',
        dateFormat: 'YYYY-MM-DD',
        correctFormat: true,
        readOnly: true,
      },
      enabled: view === 'community_prospects' || view === 'community_prospects_followup',
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
      enabled: view === 'community_prospects' || view === 'community_prospects_followup',
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
      enabled: view === 'community_prospects' || view === 'community_prospects_followup',
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
      enabled: false,
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
    // Event Attendance
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
      enabled:
        view === 'leaderboard' ||
        view === 'community_prospects' ||
        view === 'community_prospects_followup' ||
        view === 'study' ||
        view === 'new_activists_pending_workshop' ||
        chapter !== 'SF Bay Area',
      showForAllChapters: true,
    },
    {
      header: 'First Event',
      longHeader: 'First Event',
      category: 'Event Attendance',
      data: {
        data: 'first_event_name',
        readOnly: true,
        colWidths: 200,
      },
      enabled:
        view === 'leaderboard' ||
        view === 'community_prospects' ||
        view === 'community_prospects_followup' ||
        view === 'study' ||
        chapter !== 'SF Bay Area',
      showForAllChapters: true,
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
      enabled:
        view === 'leaderboard' ||
        view === 'study' ||
        view === 'new_activists_pending_workshop' ||
        chapter !== 'SF Bay Area',
      showForAllChapters: true,
    },
    {
      header: 'Mo. Since Last Action',
      longHeader: 'Months Since Last Action',
      category: 'Event Attendance',
      data: {
        type: 'numeric',
        data: 'months_since_last_action',
        readOnly: true,
        colWidths: 50,
      },
      enabled: view === 'chapter_member_development',
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
        view === 'development' ||
        view === 'new_activists' ||
        view === 'new_activists_pending_workshop',
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
      enabled:
        view === 'chapter_member_prospects' ||
        view === 'organizer_prospects' ||
        view === 'new_activists_pending_workshop',
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
      enabled: view === 'chapter_member_prospects',
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

(window as any).showOptionsModal = function (row: number) {
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
function initialDateFromValue(months: number = 1) {
  return dayjs().subtract(months, 'months').startOf('month').format('YYYY-MM-DD');
}

// Returns the current date.
function initialDateToValue() {
  return dayjs().format('YYYY-MM-DD');
}

function generateBooleanSortFn(field: string, ascending: boolean) {
  return function (a: Activist, b: Activist) {
    const order = a[field] === b[field] ? 0 : Number(a[field]) - Number(b[field]);
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateStringSortFn(field: string, ascending: boolean) {
  return function (a: Activist, b: Activist) {
    const order = a[field].toLowerCase() < b[field].toLowerCase() ? -1 : 1;
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateGenericSortFn(field: string, ascending: boolean) {
  return function (a: Activist, b: Activist) {
    const order = a[field] < b[field] ? -1 : 1;
    if (ascending) {
      return order;
    }
    return -1 * order;
  };
}

function generateDateSortFn(field: string, ascending: boolean) {
  return function (a: Activist, b: Activist) {
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
    chapterName: String,
  },
  methods: {
    groupBy(objectArray: Column[], property: any) {
      // TODO: don't use "any"
      return objectArray.reduce((acc: any, obj: any) => {
        const key = obj[property];
        if (!acc[key]) {
          acc[key] = [];
        }
        // Add object to list for given key's value
        acc[key].push(obj);
        return acc;
      }, {});
    },
    isMobile() {
      return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
        navigator.userAgent,
      );
    },
    showOptionsModal(row: number) {
      const activist = this.allActivists[row];
      this.showModal('activist-options-modal', activist, row);
      this.loading = true;

      $.ajax({
        url: '/interaction/list',
        method: 'POST',
        data: JSON.stringify({ activist_id: activist.id }),
        success: (data) => {
          const parsed = JSON.parse(data);

          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            this.loading = false;
            return;
          }

          // status === "success"
          this.currentActivist.interactions = parsed.interactions;
          this.loading = false;
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
          this.loading = false;
        },
      });
    },
    showColumnsModal() {
      this.showModal('columns-modal');
    },
    showModal(
      modalName: string,
      activist: Activist = {} as Activist,
      index: number = -1,
      interaction: Interaction = {} as Interaction,
    ) {
      // Hide the navbar so that the model doesn't go behind it.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'hidden';

      this.currentModalName = modalName;
      this.currentActivist = activist;
      this.activistIndex = index; // needed for updating activist
      this.currentInteraction = interaction;

      if (this.currentModalName === 'edit-interaction-modal' && !this.currentInteraction.id) {
        // default values for new interactions
        this.currentInteraction.followup_days = 3;
        this.currentInteraction.assign_self = true;
        this.currentInteraction.method = 'SMS';
      }

      if (this.currentModalName == 'merge-activist-modal') {
        this.getActivistMergeOptions(this.currentActivist.name);
      }

      this.disableConfirmButton = false;
    },
    hideModal() {
      if (
        this.view === 'community_prospects_followup' &&
        this.currentModalName === 'activist-options-modal'
      ) {
        // If we are only supposed to be showing activists with a past or current follow-up date,
        // then hide them if their follow-up date is now in the future.
        if (
          !this.upcomingFollowupsOnly &&
          (this.currentActivist.followup_date === '' ||
            this.currentActivist.followup_date > dayjs().format('YYYY-MM-DD'))
        ) {
          Vue.delete(this.allActivists, this.activistIndex);
        }
        // If we are only supposed to be showing activists with a future follow-up date,
        // then hide them if their follow-up date is blank.
        if (this.upcomingFollowupsOnly && this.currentActivist.followup_date === '') {
          Vue.delete(this.allActivists, this.activistIndex);
        }
      }

      this.currentModalName = '';
      this.activistIndex = -1;
      this.currentActivist = {} as Activist;
      this.currentInteraction = {} as Interaction;
      this.mergeTarget = [] as Activist[];
      this.activistMergeOptions = [];

      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'visible';
    },
    confirmMergeActivistModal() {
      if (this.mergeTarget.length === 0) {
        flashMessage('Must choose an activist to merge into', true);
        return;
      }
      const targetActivistName = this.mergeTarget[0];

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
          // TODO: Handle this better. Perhaps return updated
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
    confirmEditInteractionModal() {
      this.loading = true;
      this.disableConfirmButton = true;

      const interactionData = this.currentInteraction;
      interactionData.activist_id = this.currentActivist.id;

      $.ajax({
        url: '/interaction/save',
        method: 'POST',
        data: JSON.stringify(interactionData),
        success: (data) => {
          const parsed = JSON.parse(data);

          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            this.loading = false;
            return;
          }

          // status === "success"
          Vue.set(this.allActivists, this.activistIndex, parsed.activist);
          this.refreshHOTData();
          flashMessage('Saved interaction!', false);
          this.loading = false;
          this.disableConfirmButton = false;
          this.showOptionsModal(this.activistIndex);
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
          this.loading = false;
          this.disableConfirmButton = false;
        },
      });
    },
    deleteInteraction(interaction: Interaction) {
      this.loading = true;

      $.ajax({
        url: '/interaction/delete',
        method: 'POST',
        data: JSON.stringify({ id: interaction.id, activist_id: interaction.activist_id }),
        success: (data) => {
          const parsed = JSON.parse(data);

          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            this.loading = false;
            return;
          }

          // status === "success"
          Vue.set(this.allActivists, this.activistIndex, parsed.activist);
          this.refreshHOTData();
          flashMessage('Deleted interaction!', false);
          this.loading = false;
          this.showOptionsModal(this.activistIndex);
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
          this.loading = false;
        },
      });
    },
    getFilteredActivistMergeOptions(text: string) {
      this.filteredActivistMergeOptions = this.activistMergeOptions.filter((a: string) => {
        return a.toLowerCase().startsWith(text.toLowerCase());
      });
    },
    confirmHideActivist(activistID: number) {
      let confirm = window.confirm(
        'WARNING: Hiding this activist will make them inaccessible unless they are unhidden by Tech. Are you sure you want to hide this activist?',
      );

      if (confirm) {
        $.ajax({
          url: '/activist/hide',
          method: 'POST',
          data: JSON.stringify({ id: activistID }),
          success: (data) => {
            const parsed = JSON.parse(data);

            if (parsed.status === 'error') {
              flashMessage('Error: ' + parsed.message, true);
              this.loading = false;
              return;
            }

            // status === "success"
            // TODO: hide the activist using their index in allActivist instead of reloading the page
            this.hideModal();
            location.reload();
            this.loading = false;
          },
          error: (err) => {
            console.warn(err.responseText);
            flashMessage('Server error: ' + err.responseText, true);
            this.loading = false;
          },
        });
      }
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

          // If the user makes multiple requests within a short timespan,
          // it's possible that results will be returned from the server
          // out of order. In order to always show date from the last
          // request made, we check the timestamp.
          if (parsed.request_timestamp < this.lastRequestTimestamp) {
            this.loading = false;
            return;
          }
          this.allActivists = parsed.activist_list ? parsed.activist_list : [];
          this.refreshHOTData();
          this.lastRequestTimestamp = parsed.request_timestamp;
          this.loading = false;

          if (this.allActivists.length === 0) {
            flashMessage('No activists found matching given criteria', true);
          }
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    debounceLoadActivists: debounce(function (this: any) {
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

        const activist = this.allActivists[columnIndex];
        (function (change) {
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
        case 'community_prospects_followup':
          order_field = 'followup_date';
          break;
        default:
          order_field = 'last_event';
      }

      return {
        order: DescOrder,
        order_field: order_field,
        last_event_date_to: this.lastEventDateTo,
        last_event_date_from: this.lastEventDateFrom,
        interest_date_from: this.interestDateFrom,
        interest_date_to: this.interestDateTo,
        name: this.search,
        filter: this.view,
        assigned_to_current_user: this.assignedToCurrentUser,
        upcoming_followups_only: this.upcomingFollowupsOnly,
      };
    },
    refreshHOTData() {
      const table = this.hotTable;
      const newSettings = {
        data: rewriteSettings(this.allActivists),
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
    formatInteractionDate(s: string) {
      return dayjs(s).format('YYYY-MM-DD');
    },
  },
  data() {
    return {
      root: 'activists-root',
      currentModalName: '',
      activistIndex: -1,
      currentActivist: {} as Activist,
      currentInteraction: {} as Interaction,
      disableConfirmButton: false,
      allActivists: [] as Activist[],
      lastRequestTimestamp: 0,
      height: 500,
      columns: getColumnsForChapter(this.chapterName, this.view),
      lastEventDateFrom:
        this.view === 'all_activists' || this.view === 'leaderboard'
          ? initialDateFromValue()
          : this.view === 'new_activists' || this.view === 'new_activists_pending_workshop'
            ? initialDateFromValue(6)
            : '',
      lastEventDateTo:
        this.view === 'all_activists' ||
        this.view === 'leaderboard' ||
        this.view === 'new_activists' ||
        this.view === 'new_activists_pending_workshop'
          ? initialDateToValue()
          : '',
      interestDateFrom: this.view === 'community_prospects' ? initialDateFromValue(6) : '',
      interestDateTo: this.view === 'community_prospects' ? initialDateToValue() : '',
      assignedToCurrentUser: this.view === 'community_prospects_followup',
      upcomingFollowupsOnly: false,
      showFilters: false,
      search: '',
      loading: false,
      rowCount: 0,
      mergeTarget: [] as Activist[],
      activistMergeOptions: [] as string[],
      filteredActivistMergeOptions: [] as string[],
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
          this.view === 'all_activists' ||
          this.view === 'new_activists' ||
          this.view === 'new_activists_pending_workshop',
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
  },
  watch: {
    lastEventDateFrom() {
      this.debounceLoadActivists();
    },
    lastEventDateTo() {
      this.debounceLoadActivists();
    },
    interestDateFrom() {
      this.debounceLoadActivists();
    },
    interestDateTo() {
      this.debounceLoadActivists();
    },
    search() {
      this.debounceLoadActivists();
    },
    assignedToCurrentUser() {
      this.loadActivists();
    },
    upcomingFollowupsOnly() {
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
.activist-options-btn:active {
  color: black;
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
