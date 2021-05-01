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
      </div>

      <div class="level-right">
        <div class="level-item px-1">
          <b-button
            label="Export CSV"
            type="is-info"
            icon-left="download"
            tag="a"
            href="/csv/international_organizers"
          ></b-button>
        </div>
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

    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-field label-position="on-border" label="Filter by name">
            <b-input v-model="filterName" type="text" icon="filter"></b-input>
          </b-field>
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
            <div class="column is-one-quarter" v-if="false">
              <b-field label="Flag" label-position="on-border">
                <b-input
                  type="text"
                  v-model.trim="currentChapter.Flag"
                  required
                  maxlength="4"
                  icon="flag"
                  readonly
                >
                </b-input>
              </b-field>
            </div>
            <div class="column is-three-quarters">
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
                <b-select v-model.trim="currentChapter.Country" @input="updateCountryFlag">
                  <option v-for="country in allCountries" :value="country.Name" :key="country.Name">
                    {{ country.Name }}
                  </option>
                </b-select>
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
      let chapters = this.chapters;
      if (this.mentorFilter != 'All') {
        chapters = chapters.filter((c) => {
          return c.Mentor === this.mentorFilter;
        });
      }
      if (this.filterName != '') {
        chapters = chapters.filter((c) => {
          return c.Name.toLowerCase().startsWith(this.filterName.toLowerCase());
        });
      }
      return chapters;
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
    updateCountryFlag(name: string) {
      console.log(name);
      this.currentChapter.Flag = this.allCountries.filter((c) => {
        return c.Name === name;
      })[0].Flag;
    },
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
      if (this.currentChapter.LastContact != undefined) {
        const m = moment(this.currentChapter.LastContact);
        this.currentChapter.LastContactParsed = m.isValid() ? m.toDate() : null;
      }
      if (this.currentChapter.LastAction != undefined) {
        const m = moment(this.currentChapter.LastAction);
        this.currentChapter.LastActionParsed = m.isValid() ? m.toDate() : null;
      }

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
        flashMessage(`There are no email addresses listed for ${chapter.Name}!`, true);
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
        flashMessage('Chapter name is required!', true);
        return;
      }
      if (!this.currentChapter.Region) {
        flashMessage('Region is required!', true);
        return;
      }
      if (!this.currentChapter.Country) {
        flashMessage('Country is required!', true);
        return;
      }
      if (!this.currentChapter.Lat) {
        flashMessage('Lat is required and must be a number!', true);
        return;
      }
      if (!this.currentChapter.Lng) {
        flashMessage('Lng is required and must be a number!', true);
        return;
      }

      // Format dates as strings.
      if (this.currentChapter.LastContactParsed) {
        console.log('last contact parsed TRUE');
        console.log(this.currentChapter.LastContactParsed);
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
        flashMessage('Last Contact date must be in YYYY-MM-DD format!', true);
        return;
      }
      if (
        this.currentChapter.LastAction &&
        !this.currentChapter.LastAction.match(/^\d{4}-\d{2}-\d{2}$/)
      ) {
        flashMessage('Last Action date must be in YYYY-MM-DD format!', true);
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
        flashMessage(
          'All organizers must have a name. Please delete any blank organizer rows.',
          true,
        );
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
      filterName: '',
      loading: true,
      allCountries: [
        { Name: 'N/A', Flag: '🌎' },
        { Name: 'Andorra', Flag: '🇦🇩' },
        { Name: 'United Arab Emirates', Flag: '🇦🇪' },
        { Name: 'Afghanistan', Flag: '🇦🇫' },
        { Name: 'Antigua and Barbuda', Flag: '🇦🇬' },
        { Name: 'Anguilla', Flag: '🇦🇮' },
        { Name: 'Albania', Flag: '🇦🇱' },
        { Name: 'Armenia', Flag: '🇦🇲' },
        { Name: 'Angola', Flag: '🇦🇴' },
        { Name: 'Antarctica', Flag: '🇦🇶' },
        { Name: 'Argentina', Flag: '🇦🇷' },
        { Name: 'American Samoa', Flag: '🇦🇸' },
        { Name: 'Austria', Flag: '🇦🇹' },
        { Name: 'Australia', Flag: '🇦🇺' },
        { Name: 'Aruba', Flag: '🇦🇼' },
        { Name: 'Åland Islands', Flag: '🇦🇽' },
        { Name: 'Azerbaijan', Flag: '🇦🇿' },
        { Name: 'Bosnia and Herzegovina', Flag: '🇧🇦' },
        { Name: 'Barbados', Flag: '🇧🇧' },
        { Name: 'Bangladesh', Flag: '🇧🇩' },
        { Name: 'Belgium', Flag: '🇧🇪' },
        { Name: 'Burkina Faso', Flag: '🇧🇫' },
        { Name: 'Bulgaria', Flag: '🇧🇬' },
        { Name: 'Bahrain', Flag: '🇧🇭' },
        { Name: 'Burundi', Flag: '🇧🇮' },
        { Name: 'Benin', Flag: '🇧🇯' },
        { Name: 'Saint Barthélemy', Flag: '🇧🇱' },
        { Name: 'Bermuda', Flag: '🇧🇲' },
        { Name: 'Brunei Darussalam', Flag: '🇧🇳' },
        { Name: 'Bolivia', Flag: '🇧🇴' },
        { Name: 'Bonaire, Sint Eustatius and Saba', Flag: '🇧🇶' },
        { Name: 'Brazil', Flag: '🇧🇷' },
        { Name: 'Bahamas', Flag: '🇧🇸' },
        { Name: 'Bhutan', Flag: '🇧🇹' },
        { Name: 'Bouvet Island', Flag: '🇧🇻' },
        { Name: 'Botswana', Flag: '🇧🇼' },
        { Name: 'Belarus', Flag: '🇧🇾' },
        { Name: 'Belize', Flag: '🇧🇿' },
        { Name: 'Canada', Flag: '🇨🇦' },
        { Name: 'Cocos (Keeling) Islands', Flag: '🇨🇨' },
        { Name: 'Congo', Flag: '🇨🇩' },
        { Name: 'Central African Republic', Flag: '🇨🇫' },
        { Name: 'Congo - Brazzaville', Flag: '🇨🇬' },
        { Name: 'Switzerland', Flag: '🇨🇭' },
        { Name: "Côte D'Ivoire", Flag: '🇨🇮' },
        { Name: 'Cook Islands', Flag: '🇨🇰' },
        { Name: 'Chile', Flag: '🇨🇱' },
        { Name: 'Cameroon', Flag: '🇨🇲' },
        { Name: 'China', Flag: '🇨🇳' },
        { Name: 'Colombia', Flag: '🇨🇴' },
        { Name: 'Costa Rica', Flag: '🇨🇷' },
        { Name: 'Cuba', Flag: '🇨🇺' },
        { Name: 'Cape Verde', Flag: '🇨🇻' },
        { Name: 'Curaçao', Flag: '🇨🇼' },
        { Name: 'Christmas Island', Flag: '🇨🇽' },
        { Name: 'Cyprus', Flag: '🇨🇾' },
        { Name: 'Czech Republic', Flag: '🇨🇿' },
        { Name: 'Germany', Flag: '🇩🇪' },
        { Name: 'Djibouti', Flag: '🇩🇯' },
        { Name: 'Denmark', Flag: '🇩🇰' },
        { Name: 'Dominica', Flag: '🇩🇲' },
        { Name: 'Dominican Republic', Flag: '🇩🇴' },
        { Name: 'Algeria', Flag: '🇩🇿' },
        { Name: 'Ecuador', Flag: '🇪🇨' },
        { Name: 'Estonia', Flag: '🇪🇪' },
        { Name: 'Egypt', Flag: '🇪🇬' },
        { Name: 'Western Sahara', Flag: '🇪🇭' },
        { Name: 'Eritrea', Flag: '🇪🇷' },
        { Name: 'Spain', Flag: '🇪🇸' },
        { Name: 'Ethiopia', Flag: '🇪🇹' },
        { Name: 'Finland', Flag: '🇫🇮' },
        { Name: 'Fiji', Flag: '🇫🇯' },
        { Name: 'Falkland Islands (Malvinas)', Flag: '🇫🇰' },
        { Name: 'Micronesia', Flag: '🇫🇲' },
        { Name: 'Faroe Islands', Flag: '🇫🇴' },
        { Name: 'France', Flag: '🇫🇷' },
        { Name: 'Gabon', Flag: '🇬🇦' },
        { Name: 'United Kingdom', Flag: '🇬🇧' },
        { Name: 'Grenada', Flag: '🇬🇩' },
        { Name: 'Georgia', Flag: '🇬🇪' },
        { Name: 'French Guiana', Flag: '🇬🇫' },
        { Name: 'Guernsey', Flag: '🇬🇬' },
        { Name: 'Ghana', Flag: '🇬🇭' },
        { Name: 'Gibraltar', Flag: '🇬🇮' },
        { Name: 'Greenland', Flag: '🇬🇱' },
        { Name: 'Gambia', Flag: '🇬🇲' },
        { Name: 'Guinea', Flag: '🇬🇳' },
        { Name: 'Guadeloupe', Flag: '🇬🇵' },
        { Name: 'Equatorial Guinea', Flag: '🇬🇶' },
        { Name: 'Greece', Flag: '🇬🇷' },
        { Name: 'South Georgia', Flag: '🇬🇸' },
        { Name: 'Guatemala', Flag: '🇬🇹' },
        { Name: 'Guam', Flag: '🇬🇺' },
        { Name: 'Guinea-Bissau', Flag: '🇬🇼' },
        { Name: 'Guyana', Flag: '🇬🇾' },
        { Name: 'Hong Kong', Flag: '🇭🇰' },
        { Name: 'Heard Island and Mcdonald Islands', Flag: '🇭🇲' },
        { Name: 'Honduras', Flag: '🇭🇳' },
        { Name: 'Croatia', Flag: '🇭🇷' },
        { Name: 'Haiti', Flag: '🇭🇹' },
        { Name: 'Hungary', Flag: '🇭🇺' },
        { Name: 'Indonesia', Flag: '🇮🇩' },
        { Name: 'Ireland', Flag: '🇮🇪' },
        { Name: 'Israel', Flag: '🇮🇱' },
        { Name: 'Isle of Man', Flag: '🇮🇲' },
        { Name: 'India', Flag: '🇮🇳' },
        { Name: 'British Indian Ocean Territory', Flag: '🇮🇴' },
        { Name: 'Iraq', Flag: '🇮🇶' },
        { Name: 'Iran', Flag: '🇮🇷' },
        { Name: 'Iceland', Flag: '🇮🇸' },
        { Name: 'Italy', Flag: '🇮🇹' },
        { Name: 'Jersey', Flag: '🇯🇪' },
        { Name: 'Jamaica', Flag: '🇯🇲' },
        { Name: 'Jordan', Flag: '🇯🇴' },
        { Name: 'Japan', Flag: '🇯🇵' },
        { Name: 'Kenya', Flag: '🇰🇪' },
        { Name: 'Kyrgyzstan', Flag: '🇰🇬' },
        { Name: 'Cambodia', Flag: '🇰🇭' },
        { Name: 'Kiribati', Flag: '🇰🇮' },
        { Name: 'Comoros', Flag: '🇰🇲' },
        { Name: 'Saint Kitts and Nevis', Flag: '🇰🇳' },
        { Name: 'North Korea', Flag: '🇰🇵' },
        { Name: 'South Korea', Flag: '🇰🇷' },
        { Name: 'Kuwait', Flag: '🇰🇼' },
        { Name: 'Cayman Islands', Flag: '🇰🇾' },
        { Name: 'Kazakhstan', Flag: '🇰🇿' },
        { Name: "Lao People's Democratic Republic", Flag: '🇱🇦' },
        { Name: 'Lebanon', Flag: '🇱🇧' },
        { Name: 'Saint Lucia', Flag: '🇱🇨' },
        { Name: 'Liechtenstein', Flag: '🇱🇮' },
        { Name: 'Sri Lanka', Flag: '🇱🇰' },
        { Name: 'Liberia', Flag: '🇱🇷' },
        { Name: 'Lesotho', Flag: '🇱🇸' },
        { Name: 'Lithuania', Flag: '🇱🇹' },
        { Name: 'Luxembourg', Flag: '🇱🇺' },
        { Name: 'Latvia', Flag: '🇱🇻' },
        { Name: 'Libya', Flag: '🇱🇾' },
        { Name: 'Morocco', Flag: '🇲🇦' },
        { Name: 'Monaco', Flag: '🇲🇨' },
        { Name: 'Moldova', Flag: '🇲🇩' },
        { Name: 'Montenegro', Flag: '🇲🇪' },
        { Name: 'Saint Martin (French Part)', Flag: '🇲🇫' },
        { Name: 'Madagascar', Flag: '🇲🇬' },
        { Name: 'Marshall Islands', Flag: '🇲🇭' },
        { Name: 'Macedonia', Flag: '🇲🇰' },
        { Name: 'Mali', Flag: '🇲🇱' },
        { Name: 'Myanmar', Flag: '🇲🇲' },
        { Name: 'Mongolia', Flag: '🇲🇳' },
        { Name: 'Macao', Flag: '🇲🇴' },
        { Name: 'Northern Mariana Islands', Flag: '🇲🇵' },
        { Name: 'Martinique', Flag: '🇲🇶' },
        { Name: 'Mauritania', Flag: '🇲🇷' },
        { Name: 'Montserrat', Flag: '🇲🇸' },
        { Name: 'Malta', Flag: '🇲🇹' },
        { Name: 'Mauritius', Flag: '🇲🇺' },
        { Name: 'Maldives', Flag: '🇲🇻' },
        { Name: 'Malawi', Flag: '🇲🇼' },
        { Name: 'Mexico', Flag: '🇲🇽' },
        { Name: 'Malaysia', Flag: '🇲🇾' },
        { Name: 'Mozambique', Flag: '🇲🇿' },
        { Name: 'Namibia', Flag: '🇳🇦' },
        { Name: 'New Caledonia', Flag: '🇳🇨' },
        { Name: 'Niger', Flag: '🇳🇪' },
        { Name: 'Norfolk Island', Flag: '🇳🇫' },
        { Name: 'Nigeria', Flag: '🇳🇬' },
        { Name: 'Nicaragua', Flag: '🇳🇮' },
        { Name: 'Netherlands', Flag: '🇳🇱' },
        { Name: 'Norway', Flag: '🇳🇴' },
        { Name: 'Nepal', Flag: '🇳🇵' },
        { Name: 'Nauru', Flag: '🇳🇷' },
        { Name: 'Niue', Flag: '🇳🇺' },
        { Name: 'New Zealand', Flag: '🇳🇿' },
        { Name: 'Oman', Flag: '🇴🇲' },
        { Name: 'Panama', Flag: '🇵🇦' },
        { Name: 'Peru', Flag: '🇵🇪' },
        { Name: 'French Polynesia', Flag: '🇵🇫' },
        { Name: 'Papua New Guinea', Flag: '🇵🇬' },
        { Name: 'Philippines', Flag: '🇵🇭' },
        { Name: 'Pakistan', Flag: '🇵🇰' },
        { Name: 'Poland', Flag: '🇵🇱' },
        { Name: 'Saint Pierre and Miquelon', Flag: '🇵🇲' },
        { Name: 'Pitcairn', Flag: '🇵🇳' },
        { Name: 'Puerto Rico', Flag: '🇵🇷' },
        { Name: 'Palestinian Territory', Flag: '🇵🇸' },
        { Name: 'Portugal', Flag: '🇵🇹' },
        { Name: 'Palau', Flag: '🇵🇼' },
        { Name: 'Paraguay', Flag: '🇵🇾' },
        { Name: 'Qatar', Flag: '🇶🇦' },
        { Name: 'Réunion', Flag: '🇷🇪' },
        { Name: 'Romania', Flag: '🇷🇴' },
        { Name: 'Serbia', Flag: '🇷🇸' },
        { Name: 'Russia', Flag: '🇷🇺' },
        { Name: 'Rwanda', Flag: '🇷🇼' },
        { Name: 'Saudi Arabia', Flag: '🇸🇦' },
        { Name: 'Solomon Islands', Flag: '🇸🇧' },
        { Name: 'Seychelles', Flag: '🇸🇨' },
        { Name: 'Sudan', Flag: '🇸🇩' },
        { Name: 'Sweden', Flag: '🇸🇪' },
        { Name: 'Singapore', Flag: '🇸🇬' },
        { Name: 'Saint Helena, Ascension and Tristan Da Cunha', Flag: '🇸🇭' },
        { Name: 'Slovenia', Flag: '🇸🇮' },
        { Name: 'Svalbard and Jan Mayen', Flag: '🇸🇯' },
        { Name: 'Slovakia', Flag: '🇸🇰' },
        { Name: 'Sierra Leone', Flag: '🇸🇱' },
        { Name: 'San Marino', Flag: '🇸🇲' },
        { Name: 'Senegal', Flag: '🇸🇳' },
        { Name: 'Somalia', Flag: '🇸🇴' },
        { Name: 'Suriname', Flag: '🇸🇷' },
        { Name: 'South Sudan', Flag: '🇸🇸' },
        { Name: 'Sao Tome and Principe', Flag: '🇸🇹' },
        { Name: 'El Salvador', Flag: '🇸🇻' },
        { Name: 'Sint Maarten (Dutch Part)', Flag: '🇸🇽' },
        { Name: 'Syrian Arab Republic', Flag: '🇸🇾' },
        { Name: 'Swaziland', Flag: '🇸🇿' },
        { Name: 'Turks and Caicos Islands', Flag: '🇹🇨' },
        { Name: 'Chad', Flag: '🇹🇩' },
        { Name: 'French Southern Territories', Flag: '🇹🇫' },
        { Name: 'Togo', Flag: '🇹🇬' },
        { Name: 'Thailand', Flag: '🇹🇭' },
        { Name: 'Tajikistan', Flag: '🇹🇯' },
        { Name: 'Tokelau', Flag: '🇹🇰' },
        { Name: 'Timor-Leste', Flag: '🇹🇱' },
        { Name: 'Turkmenistan', Flag: '🇹🇲' },
        { Name: 'Tunisia', Flag: '🇹🇳' },
        { Name: 'Tonga', Flag: '🇹🇴' },
        { Name: 'Turkey', Flag: '🇹🇷' },
        { Name: 'Trinidad and Tobago', Flag: '🇹🇹' },
        { Name: 'Tuvalu', Flag: '🇹🇻' },
        { Name: 'Taiwan', Flag: '🇹🇼' },
        { Name: 'Tanzania', Flag: '🇹🇿' },
        { Name: 'Ukraine', Flag: '🇺🇦' },
        { Name: 'Uganda', Flag: '🇺🇬' },
        { Name: 'United States Minor Outlying Islands', Flag: '🇺🇲' },
        { Name: 'United States', Flag: '🇺🇸' },
        { Name: 'Uruguay', Flag: '🇺🇾' },
        { Name: 'Uzbekistan', Flag: '🇺🇿' },
        { Name: 'Vatican City', Flag: '🇻🇦' },
        { Name: 'Saint Vincent and The Grenadines', Flag: '🇻🇨' },
        { Name: 'Venezuela', Flag: '🇻🇪' },
        { Name: 'Virgin Islands, British', Flag: '🇻🇬' },
        { Name: 'Virgin Islands, U.S.', Flag: '🇻🇮' },
        { Name: 'Viet Nam', Flag: '🇻🇳' },
        { Name: 'Vanuatu', Flag: '🇻🇺' },
        { Name: 'Wallis and Futuna', Flag: '🇼🇫' },
        { Name: 'Samoa', Flag: '🇼🇸' },
        { Name: 'Yemen', Flag: '🇾🇪' },
        { Name: 'Mayotte', Flag: '🇾🇹' },
        { Name: 'South Africa', Flag: '🇿🇦' },
        { Name: 'Zambia', Flag: '🇿🇲' },
        { Name: 'Zimbabwe', Flag: '🇿🇼' },
      ],
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
/* Don't display buefy table sorting options on mobile. */
.table-mobile-sort {
  display: none;
}
</style>
