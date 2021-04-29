<template>
  <adb-page title="Chapters" class="body-wrapper-extra-wide">
    <b-loading :is-full-page="true" v-model="loading"></b-loading>

    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-button icon-left="plus" type="is-primary" @click="showModal('edit-chapter-modal')">
            New chapter
          </b-button>
        </div>
        <div class="level-item">
          <b-field label-position="on-border" label="Mentor">
            <b-select v-model="mentorFilter">
              <option v-for="mentor in mentors" :value="mentor" :key="mentor">
                {{ mentor }}
              </option>
            </b-select>
          </b-field>
        </div>
        <div class="level-item">
          <b-button
            icon-left="help"
            class="is-hidden-mobile"
            onclick="alert('Help feature coming soon.')"
          >
          </b-button>
        </div>
      </div>

      <div class="level-right">
        <div class="level-item has-text-centered">
          <div>
            <p class="heading">Total Chapters</p>
            <p class="title">{{ filteredChapters.length }}</p>
          </div>
        </div>
        <div class="level-item has-text-centered">
          <div>
            <p class="heading">Active Chapters</p>
            <p class="title">
              {{
                filteredChapters.filter((c) => {
                  return dateInLastThreeMonths(c.LastAction);
                }).length
              }}
            </p>
          </div>
        </div>
      </div>
    </nav>

    <b-table :data="filteredChapters" striped hoverable default-sort="Name">
      <b-table-column v-slot="props">
        <div style="width: 130px;">
          <b-button @click="showModal('edit-chapter-modal', props.row)">
            <b-icon icon="pencil" type="is-primary"></b-icon>
          </b-button>
          <b-button @click="composeEmail(props.row)">
            <b-icon icon="email" type="is-info"></b-icon>
          </b-button>
          <b-button @click="showModal('delete-chapter-modal', props.row)">
            <b-icon icon="delete" type="is-danger"></b-icon>
          </b-button>
        </div>
      </b-table-column>

      <b-table-column field="Name" label="Name" v-slot="props" sortable>
        {{ props.row.Flag }} {{ props.row.Name }}
      </b-table-column>

      <b-table-column field="Mentor" label="Mentor" v-slot="props" sortable>
        {{ props.row.Mentor }}
      </b-table-column>

      <b-table-column field="LastContact" label="Last Contact" v-slot="props" centered sortable>
        <span
          @click="showModal('chapter-last-contact-modal', props.row)"
          class="is-clickable tag"
          :class="colorQuarterlyGoal(props.row.LastContact)"
          >{{ props.row.LastContact || 'None' }}</span
        >
      </b-table-column>

      <b-table-column field="LastAction" label="Last Action" v-slot="props" centered sortable>
        <span
          @click="showModal('chapter-last-action-modal', props.row)"
          class="is-clickable tag"
          :class="colorQuarterlyGoal(props.row.LastAction)"
          >{{ props.row.LastAction || 'None' }}</span
        >
      </b-table-column>

      <b-table-column field="LastFBEvent" label="Last FB Event" v-slot="props" centered sortable>
        <span class="tag" :class="colorQuarterlyGoal(props.row.LastFBEvent)">{{
          props.row.LastFBEvent || 'None'
        }}</span>
      </b-table-column>

      <b-table-column field="FBSyncStatus" label="FB Sync Status" v-slot="props" centered sortable>
        <b-icon icon="circle" :type="colorFBSyncStatus(props.row.LastFBSync)"></b-icon>
      </b-table-column>
    </b-table>

    <b-modal
      :active="currentModalName === 'delete-chapter-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">Delete chapter</p>
        </header>
        <section class="modal-card-body">
          Are you sure you want to delete {{ currentChapter.Flag }}
          <strong>{{ currentChapter.Name }}</strong
          >?
        </section>
        <footer class="modal-card-foot">
          <b-button label="Cancel" @click="hideModal" />
          <b-button
            label="Delete"
            type="is-danger"
            :disabled="disableConfirmButton"
            @click="confirmDeleteChapterModal"
          />
        </footer>
      </div>
    </b-modal>

    <b-modal
      :active="currentModalName === 'edit-chapter-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="false"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">
            {{
              currentChapter.ChapterID
                ? `${currentChapter.Flag} ${currentChapter.Name}`
                : 'New chapter'
            }}
          </p>
        </header>
        <section class="modal-card-body">
          <div class="columns is-multiline">
            <div class="column is-full">
              <b-field label="Basic Info" custom-class="has-text-primary"></b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Flag" label-position="on-border">
                <b-input
                  type="text"
                  v-model.trim="currentChapter.Flag"
                  required
                  maxlength="4"
                  icon="flag"
                >
                </b-input>
              </b-field>
            </div>
            <div class="column is-half">
              <b-field label="Name" label-position="on-border">
                <b-input
                  type="text"
                  v-model.trim="currentChapter.Name"
                  required
                  maxlength="100"
                  icon="city"
                  :disabled="currentChapter.ChapterID"
                >
                </b-input>
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Mentor" label-position="on-border">
                <b-input
                  type="text"
                  v-model.trim="currentChapter.Mentor"
                  maxlength="100"
                  icon="school"
                >
                </b-input>
              </b-field>
            </div>
            <div class="column">
              <b-field label="Notes" label-position="on-border">
                <b-input type="textarea" maxlength="512" v-model.trim="currentChapter.Notes" />
              </b-field>
            </div>
            <div class="column is-full">
              <b-field label="Location" custom-class="has-text-primary"></b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Region" label-position="on-border">
                <b-select v-model="currentChapter.Region" icon="earth" expanded required>
                  <option
                    v-for="region in [
                      'North America',
                      'Central & South America',
                      'Europe',
                      'Middle East & Africa',
                      'Asia-Pacific',
                      'Online',
                    ]"
                    :value="region"
                    :key="region"
                  >
                    {{ region }}
                  </option>
                </b-select>
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Country" label-position="on-border">
                <b-input
                  type="text"
                  maxlength="128"
                  v-model.trim="currentChapter.Country"
                  required
                  icon="map"
                />
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Lat" label-position="on-border">
                <b-input
                  type="number"
                  v-model.number="currentChapter.Lat"
                  placeholder="00.000000"
                  step="0.000001"
                  min="-90"
                  max="90"
                  required
                  icon="ruler"
                />
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Lng" label-position="on-border">
                <b-input
                  type="number"
                  v-model.number="currentChapter.Lng"
                  placeholder="000.000000"
                  step="0.000001"
                  min="-180"
                  max="180"
                  required
                  icon="ruler"
                />
              </b-field>
            </div>
            <div class="column is-full">
              <b-field label="Social Links" custom-class="has-text-primary"></b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Facebook" label-position="on-border">
                <b-input
                  type="text"
                  maxlength="100"
                  v-model.trim="currentChapter.FbURL"
                  icon="facebook"
                />
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Twitter" label-position="on-border">
                <b-input
                  type="text"
                  maxlength="100"
                  v-model.trim="currentChapter.TwitterURL"
                  icon="twitter"
                />
              </b-field>
            </div>
            <div class="column is is-one-quarter">
              <b-field label="Instagram" label-position="on-border">
                <b-input
                  type="text"
                  maxlength="100"
                  v-model.trim="currentChapter.InstaURL"
                  icon="instagram"
                />
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Email (Public)" label-position="on-border">
                <b-input
                  type="email"
                  maxlength="100"
                  v-model.trim="currentChapter.Email"
                  icon="email"
                />
              </b-field>
            </div>
          </div>

          <div class="columns is-multiline" v-if="showMoreOptions">
            <div class="column is-full">
              <b-field label="Advanced Options" custom-class="has-text-primary"></b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Facebook ID" label-position="on-border">
                <b-input type="number" maxlength="16" v-model.number="currentChapter.ID" />
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Facebook Token" label-position="on-border">
                <b-input type="text" maxlength="200" v-model.trim="currentChapter.Token" />
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Eventbrite ID" label-position="on-border">
                <b-input type="number" maxlength="16" v-model.trim="currentChapter.EventbriteID" />
              </b-field>
            </div>
            <div class="column is-one-quarter">
              <b-field label="Eventbrite Token" label-position="on-border">
                <b-input
                  type="text"
                  maxlength="200"
                  v-model.trim="currentChapter.EventbriteToken"
                />
              </b-field>
            </div>
            <div class="column is-one-third">
              <b-field label="Mailing List Type" label-position="on-border">
                <b-select v-model="currentChapter.MailingListType" expanded>
                  <option
                    v-for="t in [
                      { value: '', text: 'None' },
                      { value: 'Sendy', text: 'Sendy' },
                      { value: 'Google Groups', text: 'Google Groups' },
                    ]"
                    :value="t.value"
                    :key="t.value"
                  >
                    {{ t.text }}
                  </option>
                </b-select>
              </b-field>
            </div>
            <div class="column is-one-third">
              <b-field label="Mailing List ID" label-position="on-border">
                <b-input type="text" maxlength="100" v-model.trim="currentChapter.MailingListID" />
              </b-field>
            </div>
            <div class="column is-one-third">
              <b-field label="Mailing List Radius" label-position="on-border">
                <b-input
                  type="number"
                  min="0"
                  max="300"
                  v-model.number="currentChapter.MailingListRadius"
                />
              </b-field>
            </div>
          </div>

          <b-field label="Organizers" custom-class="has-text-primary">
            <p v-if="!currentChapter.ChapterID">
              Please save the new chapter before adding organizers.
            </p>
            <b-table :data="currentChapter.Organizers" v-if="currentChapter.ChapterID">
              <template #empty>
                <div class="has-text-centered">No organizers found. Add one below.</div>
              </template>
              <b-table-column field="Name" label="Name" v-slot="props">
                <b-input
                  type="text"
                  v-model="props.row.Name"
                  placeholder="Name"
                  icon="human"
                ></b-input>
              </b-table-column>
              <b-table-column field="Email" label="Email" v-slot="props">
                <b-input
                  type="email"
                  v-model="props.row.Email"
                  placeholder="Email"
                  icon="email"
                ></b-input>
              </b-table-column>
              <b-table-column field="Phone" label="Phone" v-slot="props">
                <b-input
                  type="text"
                  v-model="props.row.Phone"
                  placeholder="Phone"
                  icon="phone"
                ></b-input>
              </b-table-column>
              <b-table-column field="Facebook" label="Facebook" v-slot="props">
                <b-input
                  type="text"
                  v-model="props.row.Facebook"
                  placeholder="Facebook"
                  icon="facebook"
                ></b-input>
              </b-table-column>
              <b-table-column v-slot="props">
                <b-button @click="deleteOrganizer(props.row)">
                  <b-icon icon="delete" type="is-danger"></b-icon>
                </b-button>
              </b-table-column>
            </b-table>
          </b-field>
          <b-button
            label="Add new organizer"
            icon-left="plus"
            @click="addOrganizer"
            v-if="currentChapter.ChapterID"
          ></b-button>
        </section>
        <footer class="modal-card-foot is-flex is-justify-content-space-between">
          <div>
            <b-button label="Cancel" @click="hideModal" icon-left="cancel" class="mb-2" />
            <b-button
              label="Save"
              type="is-primary"
              native-type="submit"
              :disabled="disableConfirmButton"
              @click="confirmEditChapterModal"
              icon-left="floppy"
            />
          </div>
          <div>
            <b-button
              label="Advanced options"
              type="is-warning"
              @click="toggleShowMoreOptions"
              class="right"
              icon-left="wrench"
            />
          </div>
        </footer>
      </div>
    </b-modal>

    <b-modal
      :active="currentModalName === 'chapter-last-contact-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">Last Contact</p>
          <h4>{{ currentChapter.Flag }} {{ currentChapter.Name }}</h4>
        </header>
        <section class="modal-card-body">
          <b-datepicker v-model="currentChapter.LastContactParsed" inline></b-datepicker>
          <p class="mt-2 is-flex is-justify-content-center">
            <b-button class="is-success is-small mx-1" @click="setDateToToday('LastContact')">
              Today
            </b-button>
            <b-button class="is-danger is-small mx-1" @click="resetDate('LastContact')">
              Reset
            </b-button>
          </p>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Cancel" icon-left="cancel" @click="hideModal" />
          <b-button
            label="Save"
            icon-left="floppy"
            type="is-primary"
            :disabled="disableConfirmButton"
            @click="confirmEditChapterModal"
          />
        </footer>
      </div>
    </b-modal>

    <b-modal
      :active="currentModalName === 'chapter-last-action-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="true"
      :on-cancel="hideModal"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">Last Action</p>
          <h4>{{ currentChapter.Flag }} {{ currentChapter.Name }}</h4>
        </header>
        <section class="modal-card-body">
          <b-datepicker v-model="currentChapter.LastActionParsed" inline></b-datepicker>
          <p class="mt-2 is-flex is-justify-content-center">
            <b-button class="is-success is-small mx-1" @click="setDateToToday('LastAction')">
              Today
            </b-button>
            <b-button class="is-danger is-small mx-1" @click="resetDate('LastAction')">
              Reset
            </b-button>
          </p>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Cancel" icon-left="cancel" @click="hideModal" />
          <b-button
            label="Save"
            icon-left="floppy"
            type="is-primary"
            :disabled="disableConfirmButton"
            @click="confirmEditChapterModal"
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
import moment from 'moment';

interface Chapter {
  ChapterID: number;
  Name: string;
  Flag: string;
  FbURL: string;
  TwitterURL: string;
  InstaURL: string;
  Email: string;
  Region: string;
  Lat: number;
  Lng: number;
  MailingListType: string;
  MailingListRadius: number;
  MailingListID: string;
  ID: number; // Facebook ID
  Token: string;
  LastFBSync: string;
  LastFBEvent: string;
  EventbriteID: string;
  EventbriteToken: string;
  Mentor: string;
  Country: string;
  Notes: string;
  LastContact: string; // TODO: just get Dates from the backend instead of parsing strings.
  LastContactParsed: Date | null;
  LastAction: string;
  LastActionParsed: Date | null;
  Organizers: Organizer[];
}

interface Organizer {
  Name: string;
  Email: string;
  Phone: string;
  Facebook: string;
}

export default Vue.extend({
  name: 'chapter-list',
  computed: {
    filteredChapters: function(): Chapter[] {
      if (this.mentorFilter === 'All') {
        return this.chapters;
      }
      return this.chapters.filter((c) => {
        return c.Mentor === this.mentorFilter;
      });
    },
    mentors: function(): string[] {
      let mentors: string[];
      mentors = ['All'];
      this.chapters.forEach((c) => {
        if (c.Mentor && mentors.indexOf(c.Mentor) === -1) {
          mentors.push(c.Mentor);
        }
      });
      return mentors.sort();
    },
  },
  methods: {
    isMobile() {
      return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
        navigator.userAgent,
      );
    },
    showModal(modalName: string, chapter: Chapter) {
      // Hide the navbar so that the model doesn't go behind it.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'hidden';

      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      this.currentChapter = { ...chapter };

      // Parse strings to dates.
      const c = moment(this.currentChapter.LastContact);
      this.currentChapter.LastContactParsed = c.isValid() ? c.toDate() : null;
      const a = moment(this.currentChapter.LastAction);
      this.currentChapter.LastActionParsed = a.isValid() ? a.toDate() : null;

      // Get the index for updating the view w/o refreshing the whole page.
      this.currentChapterIndex = this.chapters.findIndex((c) => {
        return c.ChapterID === this.currentChapter.ChapterID;
      });

      this.currentModalName = modalName;
    },
    composeEmail(chapter: Chapter) {
      let orgEmails = [] as string[];

      if (chapter.Email) {
        orgEmails.push(chapter.Email);
      }

      chapter.Organizers.forEach((o) => {
        if (o.Email) {
          orgEmails.push(o.Email);
        }
      });

      if (!orgEmails.length) {
        alert(`There are no email addresses listed for ${chapter.Name}!`);
        return;
      }

      const emailLink = `https://mail.google.com/mail/?view=cm&fs=1&su=${
        chapter.Name
      }&to=${orgEmails.join(',')}`;

      window.open(emailLink);
    },
    hideModal() {
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'visible';

      this.currentModalName = '';
      this.currentChapter = {} as Chapter;
      this.currentChapterIndex = -1;
      this.showMoreOptions = false;
    },
    confirmEditChapterModal() {
      if (!this.currentChapter.Name) {
        alert('Chapter name is required!');
        return;
      }
      if (!this.currentChapter.Flag) {
        alert('Flag emoji is required!');
        return;
      }
      if (!this.currentChapter.Region) {
        alert('Region is required!');
        return;
      }
      if (!this.currentChapter.Country) {
        alert('Country is required!');
        return;
      }
      if (!this.currentChapter.Lat) {
        alert('Lat is required and must be a number!');
        return;
      }
      if (!this.currentChapter.Lng) {
        alert('Lng is required and must be a number!');
        return;
      }

      // Format dates as strings.
      if (this.currentChapter.LastContactParsed) {
        this.currentChapter.LastContact = moment(this.currentChapter.LastContactParsed).format(
          'YYYY-MM-DD',
        );
      }
      if (this.currentChapter.LastActionParsed) {
        this.currentChapter.LastAction = moment(this.currentChapter.LastActionParsed).format(
          'YYYY-MM-DD',
        );
      }

      if (
        this.currentChapter.LastContact &&
        !this.currentChapter.LastContact.match(/^\d{4}-\d{2}-\d{2}$/)
      ) {
        alert('Last Contact date must be in YYYY-MM-DD format!');
        return;
      }
      if (
        this.currentChapter.LastAction &&
        !this.currentChapter.LastAction.match(/^\d{4}-\d{2}-\d{2}$/)
      ) {
        alert('Last Action date must be in YYYY-MM-DD format!');
        return;
      }

      if (!this.currentChapter.Organizers) this.currentChapter.Organizers = [];

      let organizersUnclean = false;
      this.currentChapter.Organizers.forEach((o) => {
        if (!o.Name) {
          organizersUnclean = true;
        }
      });
      if (organizersUnclean) {
        alert('All organizers must have a name. Please delete any blank organizer rows.');
        return;
      }

      const csrfToken = $('meta[name="csrf-token"]').attr('content');
      this.disableConfirmButton = true;
      $.ajax({
        url: '/chapter/save',
        method: 'POST',
        headers: { 'X-CSRF-Token': csrfToken },
        contentType: 'application/json',
        data: JSON.stringify(this.currentChapter),
        success: (data) => {
          this.disableConfirmButton = false;

          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          // status === "success"
          flashMessage(this.currentChapter.Name + ' saved');

          if (!this.currentChapter.ChapterID) {
            // New chapter, insert at the top.
            this.chapters = [parsed.chapter].concat(this.chapters);
          } else {
            // Existing chapter, update it.
            Vue.set(this.chapters, this.currentChapterIndex, parsed.chapter);
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
    confirmDeleteChapterModal() {
      const csrfToken = $('meta[name="csrf-token"]').attr('content');
      this.disableConfirmButton = true;
      $.ajax({
        url: '/chapter/delete',
        method: 'POST',
        headers: { 'X-CSRF-Token': csrfToken },
        contentType: 'application/json',
        data: JSON.stringify({
          chapter_id: this.currentChapter.ChapterID,
        }),
        success: (data) => {
          this.disableConfirmButton = false;

          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }
          flashMessage(this.currentChapter.Name + ' deleted');

          this.chapters.splice(this.currentChapterIndex, 1);

          this.hideModal();
        },
        error: (err) => {
          this.disableConfirmButton = false;
          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    addOrganizer() {
      this.currentChapter.Organizers = this.currentChapter.Organizers
        ? this.currentChapter.Organizers
        : [];
      this.currentChapter.Organizers.push({} as Organizer);
    },
    deleteOrganizer(o: Organizer) {
      const index = this.currentChapter.Organizers.indexOf(o);
      this.currentChapter.Organizers.splice(index, 1);
    },
    toggleShowMoreOptions() {
      this.showMoreOptions = !this.showMoreOptions;
    },
    setDateToToday(field: string) {
      if (field === 'LastContact') {
        this.currentChapter.LastContactParsed = moment()
          .local()
          .toDate();
      } else if (field === 'LastAction') {
        this.currentChapter.LastActionParsed = moment()
          .local()
          .toDate();
      }
      this.confirmEditChapterModal();
    },
    resetDate(field: string) {
      if (field === 'LastContact') {
        this.currentChapter.LastContactParsed = null;
        this.currentChapter.LastContact = '';
      } else if (field === 'LastAction') {
        this.currentChapter.LastActionParsed = null;
        this.currentChapter.LastAction = '';
      }
      this.confirmEditChapterModal();
    },
    colorFBSyncStatus(text: string) {
      const time = moment(text).add(8, 'hour'); // this converts our DB time for this field to UTC
      let c = 'is-grey';
      if (time.isValid()) {
        c = 'is-danger';
      }
      if (time.isAfter(moment().add(-1, 'day'))) {
        c = 'is-warning';
      }
      if (time.isAfter(moment().add(-1, 'hour'))) {
        c = 'is-success';
      }
      return c;
    },
    colorQuarterlyGoal(text: string) {
      const time = moment(text);
      let c = '';
      if (time.isValid()) {
        c = 'is-danger';
      }
      if (time.isAfter(moment().add(-3, 'month'))) {
        c = 'is-success';
      }
      return c;
    },
    dateInLastThreeMonths(text: string): boolean {
      return moment(text).isAfter(moment().add(-3, 'month'));
    },
  },
  data() {
    return {
      currentChapter: {} as Chapter,
      currentChapterIndex: -1,
      chapters: [] as Chapter[],
      disableConfirmButton: false,
      currentModalName: '',
      showMoreOptions: false,
      mentorFilter: 'All',
      loading: true,
    };
  },

  created() {
    const csrfToken = $('meta[name="csrf-token"]').attr('content');
    // Get chapters
    $.ajax({
      url: '/chapter/list',
      headers: { 'X-CSRF-Token': csrfToken },
      method: 'POST',
      success: (data) => {
        const parsed = JSON.parse(data);
        if (parsed.status === 'error') {
          flashMessage('Error: ' + parsed.message, true);
          return;
        }
        // status === "success"
        this.loading = false;
        this.chapters = parsed.chapters;
      },
      error: (err) => {
        this.loading = false;
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
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

<style>
/* Don't display buefy tbale sorting options on mobile. */
.table-mobile-sort {
  display: none;
}
</style>
