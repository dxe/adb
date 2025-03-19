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
            <p class="title">
              {{
                filteredChapters.filter((x) => {
                  return x.Region !== 'Online';
                }).length
              }}
            </p>
          </div>
        </div>
        <div class="level-item has-text-centered">
          <div>
            <p class="heading">Active Chapters</p>
            <p class="title">
              {{
                filteredChapters.filter((x) => {
                  return dateInLastThreeMonths(x.LastAction) && x.Region !== 'Online';
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
        <div class="level-item">
          <b-checkbox v-model="showFacebookColumns">Show FB columns</b-checkbox>
        </div>
      </div>
    </nav>

    <b-table
      :data="filteredChapters"
      striped
      hoverable
      default-sort="Name"
      :key="showFacebookColumns"
    >
      <b-table-column v-slot="props">
        <div style="width: 130px">
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
          >{{ props.row.LastContact || 'None' }}</span
        >
      </b-table-column>

      <b-table-column field="LastAction" label="Last Action" v-slot="props" centered sortable>
        <span
          @click="showModal('chapter-last-action-modal', props.row)"
          class="is-clickable tag"
          :class="colorLastAction(props.row.LastAction)"
          :title="lastActionTooltip(props.row.LastAction)"
          >{{ props.row.LastAction || 'None' }}</span
        >
      </b-table-column>

      <b-table-column
        field="LastFBEvent"
        label="Last FB Event"
        v-slot="props"
        centered
        sortable
        v-if="showFacebookColumns"
      >
        <span class="tag">{{ props.row.LastFBEvent || 'None' }}</span>
      </b-table-column>

      <b-table-column
        field="LastFBSync"
        label="FB Sync Status"
        v-slot="props"
        centered
        sortable
        v-if="showFacebookColumns"
      >
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
                  <option v-for="country in allCountries" :value="country.Code" :key="country.Code">
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
                <b-input type="text" maxlength="500" v-model.trim="currentChapter.Token" />
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
                      { value: 'SendGrid', text: 'SendGrid' },
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
import * as dayjs from 'dayjs';

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
    filteredChapters: function (): Chapter[] {
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
    mentors: function (): string[] {
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
    updateCountryFlag(code: string) {
      this.currentChapter.Flag = this.allCountries.filter((c) => {
        return c.Code === code;
      })[0].Flag;
    },
    isMobile() {
      return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
        navigator.userAgent,
      );
    },
    showModal(modalName: string, chapter?: Chapter) {
      // Hide the navbar so that the model doesn't go behind it.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'hidden';

      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      this.currentChapter = !chapter ? ({} as Chapter) : { ...chapter };

      // Parse strings to dates.
      if (this.currentChapter.LastContact != undefined) {
        const m = dayjs(this.currentChapter.LastContact);
        this.currentChapter.LastContactParsed = m.isValid() ? m.toDate() : null;
      }
      if (this.currentChapter.LastAction != undefined) {
        const m = dayjs(this.currentChapter.LastAction);
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
        this.currentChapter.LastContact = dayjs(this.currentChapter.LastContactParsed).format(
          'YYYY-MM-DD',
        );
      }
      if (this.currentChapter.LastActionParsed) {
        this.currentChapter.LastAction = dayjs(this.currentChapter.LastActionParsed).format(
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
      const data = JSON.stringify(this.currentChapter);
      console.log(data);
      $.ajax({
        url: '/chapter/save',
        method: 'POST',
        headers: { 'X-CSRF-Token': csrfToken },
        contentType: 'application/json',
        data,
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
        this.currentChapter.LastContactParsed = dayjs().toDate();
      } else if (field === 'LastAction') {
        this.currentChapter.LastActionParsed = dayjs().toDate();
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
      const time = dayjs(text).add(8, 'hour'); // this converts our DB time for this field to UTC
      let c = 'is-grey';
      if (time.isValid()) {
        c = 'is-danger';
      }
      if (time.isAfter(dayjs().add(-1, 'day'))) {
        c = 'is-warning';
      }
      if (time.isAfter(dayjs().add(-1, 'hour'))) {
        c = 'is-success';
      }
      return c;
    },
    colorLastAction(text: string) {
      const GREEN = 'is-success';
      const YELLOW = 'is-warning';
      const RED = 'is-danger';
      const GRAY = 'is-grey';
      const BLACK = 'is-black';

      const time = dayjs(text);

      if (!time.isValid()) {
        return GRAY;
      }

      if (time.isAfter(dayjs().subtract(30 * 2, 'day'))) {
        return GREEN;
      } else if (time.isAfter(dayjs().subtract(30 * 3.5, 'day'))) {
        return YELLOW;
      } else if (time.isAfter(dayjs().subtract(30 * 4 + 1, 'day'))) {
        // + 1 becauase "black" means the chapter should be offboarded, but
        // when there are "0" days remaining according to date subtraction,
        // there still may be a fraction of a day remaining in reality because
        // protests don't start at the 0th hour of the day.
        return RED;
      } else {
        return BLACK;
      }
    },
    lastActionTooltip(text: string) {
      const time = dayjs(text);

      if (!time.isValid()) {
        return undefined;
      }

      const daysSinceLastActionText = dayjs().diff(time, 'day') + ' days since last action';
      const daysRemainingToHostActionText =
        time.add(30 * 4, 'day').diff(dayjs(), 'day') + ' days remaining to host an action';
      return daysSinceLastActionText + '\n' + daysRemainingToHostActionText;
    },
    dateInLastThreeMonths(text: string): boolean {
      return dayjs(text).isAfter(dayjs().add(-3, 'month'));
    },
  },
  data() {
    return {
      currentChapter: {} as Chapter,
      currentChapterIndex: -1,
      chapters: [] as Chapter[],
      showFacebookColumns: false,
      disableConfirmButton: false,
      currentModalName: '',
      showMoreOptions: false,
      mentorFilter: 'All',
      filterName: '',
      loading: true,
      allCountries: [
        {
          Name: 'Ascension Island',
          Code: 'AC',
          Flag: '🇦🇨',
        },
        {
          Name: 'Andorra',
          Code: 'AD',
          Flag: '🇦🇩',
        },
        {
          Name: 'United Arab Emirates',
          Code: 'AE',
          Flag: '🇦🇪',
        },
        {
          Name: 'Afghanistan',
          Code: 'AF',
          Flag: '🇦🇫',
        },
        {
          Name: 'Antigua & Barbuda',
          Code: 'AG',
          Flag: '🇦🇬',
        },
        {
          Name: 'Anguilla',
          Code: 'AI',
          Flag: '🇦🇮',
        },
        {
          Name: 'Albania',
          Code: 'AL',
          Flag: '🇦🇱',
        },
        {
          Name: 'Armenia',
          Code: 'AM',
          Flag: '🇦🇲',
        },
        {
          Name: 'Angola',
          Code: 'AO',
          Flag: '🇦🇴',
        },
        {
          Name: 'Antarctica',
          Code: 'AQ',
          Flag: '🇦🇶',
        },
        {
          Name: 'Argentina',
          Code: 'AR',
          Flag: '🇦🇷',
        },
        {
          Name: 'American Samoa',
          Code: 'AS',
          Flag: '🇦🇸',
        },
        {
          Name: 'Austria',
          Code: 'AT',
          Flag: '🇦🇹',
        },
        {
          Name: 'Australia',
          Code: 'AU',
          Flag: '🇦🇺',
        },
        {
          Name: 'Aruba',
          Code: 'AW',
          Flag: '🇦🇼',
        },
        {
          Name: 'Åland Islands',
          Code: 'AX',
          Flag: '🇦🇽',
        },
        {
          Name: 'Azerbaijan',
          Code: 'AZ',
          Flag: '🇦🇿',
        },
        {
          Name: 'Bosnia & Herzegovina',
          Code: 'BA',
          Flag: '🇧🇦',
        },
        {
          Name: 'Barbados',
          Code: 'BB',
          Flag: '🇧🇧',
        },
        {
          Name: 'Bangladesh',
          Code: 'BD',
          Flag: '🇧🇩',
        },
        {
          Name: 'Belgium',
          Code: 'BE',
          Flag: '🇧🇪',
        },
        {
          Name: 'Burkina Faso',
          Code: 'BF',
          Flag: '🇧🇫',
        },
        {
          Name: 'Bulgaria',
          Code: 'BG',
          Flag: '🇧🇬',
        },
        {
          Name: 'Bahrain',
          Code: 'BH',
          Flag: '🇧🇭',
        },
        {
          Name: 'Burundi',
          Code: 'BI',
          Flag: '🇧🇮',
        },
        {
          Name: 'Benin',
          Code: 'BJ',
          Flag: '🇧🇯',
        },
        {
          Name: 'St. Barthélemy',
          Code: 'BL',
          Flag: '🇧🇱',
        },
        {
          Name: 'Bermuda',
          Code: 'BM',
          Flag: '🇧🇲',
        },
        {
          Name: 'Brunei',
          Code: 'BN',
          Flag: '🇧🇳',
        },
        {
          Name: 'Bolivia',
          Code: 'BO',
          Flag: '🇧🇴',
        },
        {
          Name: 'Caribbean Netherlands',
          Code: 'BQ',
          Flag: '🇧🇶',
        },
        {
          Name: 'Brazil',
          Code: 'BR',
          Flag: '🇧🇷',
        },
        {
          Name: 'Bahamas',
          Code: 'BS',
          Flag: '🇧🇸',
        },
        {
          Name: 'Bhutan',
          Code: 'BT',
          Flag: '🇧🇹',
        },
        {
          Name: 'Bouvet Island',
          Code: 'BV',
          Flag: '🇧🇻',
        },
        {
          Name: 'Botswana',
          Code: 'BW',
          Flag: '🇧🇼',
        },
        {
          Name: 'Belarus',
          Code: 'BY',
          Flag: '🇧🇾',
        },
        {
          Name: 'Belize',
          Code: 'BZ',
          Flag: '🇧🇿',
        },
        {
          Name: 'Canada',
          Code: 'CA',
          Flag: '🇨🇦',
        },
        {
          Name: 'Cocos (Keeling) Islands',
          Code: 'CC',
          Flag: '🇨🇨',
        },
        {
          Name: 'Congo - Kinshasa',
          Code: 'CD',
          Flag: '🇨🇩',
        },
        {
          Name: 'Central African Republic',
          Code: 'CF',
          Flag: '🇨🇫',
        },
        {
          Name: 'Congo - Brazzaville',
          Code: 'CG',
          Flag: '🇨🇬',
        },
        {
          Name: 'Switzerland',
          Code: 'CH',
          Flag: '🇨🇭',
        },
        {
          Name: 'Côte d’Ivoire',
          Code: 'CI',
          Flag: '🇨🇮',
        },
        {
          Name: 'Cook Islands',
          Code: 'CK',
          Flag: '🇨🇰',
        },
        {
          Name: 'Chile',
          Code: 'CL',
          Flag: '🇨🇱',
        },
        {
          Name: 'Cameroon',
          Code: 'CM',
          Flag: '🇨🇲',
        },
        {
          Name: 'China',
          Code: 'CN',
          Flag: '🇨🇳',
        },
        {
          Name: 'Colombia',
          Code: 'CO',
          Flag: '🇨🇴',
        },
        {
          Name: 'Clipperton Island',
          Code: 'CP',
          Flag: '🇨🇵',
        },
        {
          Name: 'Costa Rica',
          Code: 'CR',
          Flag: '🇨🇷',
        },
        {
          Name: 'Cuba',
          Code: 'CU',
          Flag: '🇨🇺',
        },
        {
          Name: 'Cape Verde',
          Code: 'CV',
          Flag: '🇨🇻',
        },
        {
          Name: 'Curaçao',
          Code: 'CW',
          Flag: '🇨🇼',
        },
        {
          Name: 'Christmas Island',
          Code: 'CX',
          Flag: '🇨🇽',
        },
        {
          Name: 'Cyprus',
          Code: 'CY',
          Flag: '🇨🇾',
        },
        {
          Name: 'Czechia',
          Code: 'CZ',
          Flag: '🇨🇿',
        },
        {
          Name: 'Germany',
          Code: 'DE',
          Flag: '🇩🇪',
        },
        {
          Name: 'Diego Garcia',
          Code: 'DG',
          Flag: '🇩🇬',
        },
        {
          Name: 'Djibouti',
          Code: 'DJ',
          Flag: '🇩🇯',
        },
        {
          Name: 'Denmark',
          Code: 'DK',
          Flag: '🇩🇰',
        },
        {
          Name: 'Dominica',
          Code: 'DM',
          Flag: '🇩🇲',
        },
        {
          Name: 'Dominican Republic',
          Code: 'DO',
          Flag: '🇩🇴',
        },
        {
          Name: 'Algeria',
          Code: 'DZ',
          Flag: '🇩🇿',
        },
        {
          Name: 'Ceuta & Melilla',
          Code: 'EA',
          Flag: '🇪🇦',
        },
        {
          Name: 'Ecuador',
          Code: 'EC',
          Flag: '🇪🇨',
        },
        {
          Name: 'Estonia',
          Code: 'EE',
          Flag: '🇪🇪',
        },
        {
          Name: 'Egypt',
          Code: 'EG',
          Flag: '🇪🇬',
        },
        {
          Name: 'Western Sahara',
          Code: 'EH',
          Flag: '🇪🇭',
        },
        {
          Name: 'Eritrea',
          Code: 'ER',
          Flag: '🇪🇷',
        },
        {
          Name: 'Spain',
          Code: 'ES',
          Flag: '🇪🇸',
        },
        {
          Name: 'Ethiopia',
          Code: 'ET',
          Flag: '🇪🇹',
        },
        {
          Name: 'European Union',
          Code: 'EU',
          Flag: '🇪🇺',
        },
        {
          Name: 'Finland',
          Code: 'FI',
          Flag: '🇫🇮',
        },
        {
          Name: 'Fiji',
          Code: 'FJ',
          Flag: '🇫🇯',
        },
        {
          Name: 'Falkland Islands',
          Code: 'FK',
          Flag: '🇫🇰',
        },
        {
          Name: 'Micronesia',
          Code: 'FM',
          Flag: '🇫🇲',
        },
        {
          Name: 'Faroe Islands',
          Code: 'FO',
          Flag: '🇫🇴',
        },
        {
          Name: 'France',
          Code: 'FR',
          Flag: '🇫🇷',
        },
        {
          Name: 'Gabon',
          Code: 'GA',
          Flag: '🇬🇦',
        },
        {
          Name: 'United Kingdom',
          Code: 'GB',
          Flag: '🇬🇧',
        },
        {
          Name: 'Grenada',
          Code: 'GD',
          Flag: '🇬🇩',
        },
        {
          Name: 'Georgia',
          Code: 'GE',
          Flag: '🇬🇪',
        },
        {
          Name: 'French Guiana',
          Code: 'GF',
          Flag: '🇬🇫',
        },
        {
          Name: 'Guernsey',
          Code: 'GG',
          Flag: '🇬🇬',
        },
        {
          Name: 'Ghana',
          Code: 'GH',
          Flag: '🇬🇭',
        },
        {
          Name: 'Gibraltar',
          Code: 'GI',
          Flag: '🇬🇮',
        },
        {
          Name: 'Greenland',
          Code: 'GL',
          Flag: '🇬🇱',
        },
        {
          Name: 'Gambia',
          Code: 'GM',
          Flag: '🇬🇲',
        },
        {
          Name: 'Guinea',
          Code: 'GN',
          Flag: '🇬🇳',
        },
        {
          Name: 'Guadeloupe',
          Code: 'GP',
          Flag: '🇬🇵',
        },
        {
          Name: 'Equatorial Guinea',
          Code: 'GQ',
          Flag: '🇬🇶',
        },
        {
          Name: 'Greece',
          Code: 'GR',
          Flag: '🇬🇷',
        },
        {
          Name: 'South Georgia & South Sandwich Islands',
          Code: 'GS',
          Flag: '🇬🇸',
        },
        {
          Name: 'Guatemala',
          Code: 'GT',
          Flag: '🇬🇹',
        },
        {
          Name: 'Guam',
          Code: 'GU',
          Flag: '🇬🇺',
        },
        {
          Name: 'Guinea-Bissau',
          Code: 'GW',
          Flag: '🇬🇼',
        },
        {
          Name: 'Guyana',
          Code: 'GY',
          Flag: '🇬🇾',
        },
        {
          Name: 'Hong Kong SAR China',
          Code: 'HK',
          Flag: '🇭🇰',
        },
        {
          Name: 'Heard & McDonald Islands',
          Code: 'HM',
          Flag: '🇭🇲',
        },
        {
          Name: 'Honduras',
          Code: 'HN',
          Flag: '🇭🇳',
        },
        {
          Name: 'Croatia',
          Code: 'HR',
          Flag: '🇭🇷',
        },
        {
          Name: 'Haiti',
          Code: 'HT',
          Flag: '🇭🇹',
        },
        {
          Name: 'Hungary',
          Code: 'HU',
          Flag: '🇭🇺',
        },
        {
          Name: 'Canary Islands',
          Code: 'IC',
          Flag: '🇮🇨',
        },
        {
          Name: 'Indonesia',
          Code: 'ID',
          Flag: '🇮🇩',
        },
        {
          Name: 'Ireland',
          Code: 'IE',
          Flag: '🇮🇪',
        },
        {
          Name: 'Israel',
          Code: 'IL',
          Flag: '🇮🇱',
        },
        {
          Name: 'Isle of Man',
          Code: 'IM',
          Flag: '🇮🇲',
        },
        {
          Name: 'India',
          Code: 'IN',
          Flag: '🇮🇳',
        },
        {
          Name: 'British Indian Ocean Territory',
          Code: 'IO',
          Flag: '🇮🇴',
        },
        {
          Name: 'Iraq',
          Code: 'IQ',
          Flag: '🇮🇶',
        },
        {
          Name: 'Iran',
          Code: 'IR',
          Flag: '🇮🇷',
        },
        {
          Name: 'Iceland',
          Code: 'IS',
          Flag: '🇮🇸',
        },
        {
          Name: 'Italy',
          Code: 'IT',
          Flag: '🇮🇹',
        },
        {
          Name: 'Jersey',
          Code: 'JE',
          Flag: '🇯🇪',
        },
        {
          Name: 'Jamaica',
          Code: 'JM',
          Flag: '🇯🇲',
        },
        {
          Name: 'Jordan',
          Code: 'JO',
          Flag: '🇯🇴',
        },
        {
          Name: 'Japan',
          Code: 'JP',
          Flag: '🇯🇵',
        },
        {
          Name: 'Kenya',
          Code: 'KE',
          Flag: '🇰🇪',
        },
        {
          Name: 'Kyrgyzstan',
          Code: 'KG',
          Flag: '🇰🇬',
        },
        {
          Name: 'Cambodia',
          Code: 'KH',
          Flag: '🇰🇭',
        },
        {
          Name: 'Kiribati',
          Code: 'KI',
          Flag: '🇰🇮',
        },
        {
          Name: 'Comoros',
          Code: 'KM',
          Flag: '🇰🇲',
        },
        {
          Name: 'St. Kitts & Nevis',
          Code: 'KN',
          Flag: '🇰🇳',
        },
        {
          Name: 'North Korea',
          Code: 'KP',
          Flag: '🇰🇵',
        },
        {
          Name: 'South Korea',
          Code: 'KR',
          Flag: '🇰🇷',
        },
        {
          Name: 'Kuwait',
          Code: 'KW',
          Flag: '🇰🇼',
        },
        {
          Name: 'Cayman Islands',
          Code: 'KY',
          Flag: '🇰🇾',
        },
        {
          Name: 'Kazakhstan',
          Code: 'KZ',
          Flag: '🇰🇿',
        },
        {
          Name: 'Laos',
          Code: 'LA',
          Flag: '🇱🇦',
        },
        {
          Name: 'Lebanon',
          Code: 'LB',
          Flag: '🇱🇧',
        },
        {
          Name: 'St. Lucia',
          Code: 'LC',
          Flag: '🇱🇨',
        },
        {
          Name: 'Liechtenstein',
          Code: 'LI',
          Flag: '🇱🇮',
        },
        {
          Name: 'Sri Lanka',
          Code: 'LK',
          Flag: '🇱🇰',
        },
        {
          Name: 'Liberia',
          Code: 'LR',
          Flag: '🇱🇷',
        },
        {
          Name: 'Lesotho',
          Code: 'LS',
          Flag: '🇱🇸',
        },
        {
          Name: 'Lithuania',
          Code: 'LT',
          Flag: '🇱🇹',
        },
        {
          Name: 'Luxembourg',
          Code: 'LU',
          Flag: '🇱🇺',
        },
        {
          Name: 'Latvia',
          Code: 'LV',
          Flag: '🇱🇻',
        },
        {
          Name: 'Libya',
          Code: 'LY',
          Flag: '🇱🇾',
        },
        {
          Name: 'Morocco',
          Code: 'MA',
          Flag: '🇲🇦',
        },
        {
          Name: 'Monaco',
          Code: 'MC',
          Flag: '🇲🇨',
        },
        {
          Name: 'Moldova',
          Code: 'MD',
          Flag: '🇲🇩',
        },
        {
          Name: 'Montenegro',
          Code: 'ME',
          Flag: '🇲🇪',
        },
        {
          Name: 'St. Martin',
          Code: 'MF',
          Flag: '🇲🇫',
        },
        {
          Name: 'Madagascar',
          Code: 'MG',
          Flag: '🇲🇬',
        },
        {
          Name: 'Marshall Islands',
          Code: 'MH',
          Flag: '🇲🇭',
        },
        {
          Name: 'North Macedonia',
          Code: 'MK',
          Flag: '🇲🇰',
        },
        {
          Name: 'Mali',
          Code: 'ML',
          Flag: '🇲🇱',
        },
        {
          Name: 'Myanmar (Burma)',
          Code: 'MM',
          Flag: '🇲🇲',
        },
        {
          Name: 'Mongolia',
          Code: 'MN',
          Flag: '🇲🇳',
        },
        {
          Name: 'Macao SAR China',
          Code: 'MO',
          Flag: '🇲🇴',
        },
        {
          Name: 'Northern Mariana Islands',
          Code: 'MP',
          Flag: '🇲🇵',
        },
        {
          Name: 'Martinique',
          Code: 'MQ',
          Flag: '🇲🇶',
        },
        {
          Name: 'Mauritania',
          Code: 'MR',
          Flag: '🇲🇷',
        },
        {
          Name: 'Montserrat',
          Code: 'MS',
          Flag: '🇲🇸',
        },
        {
          Name: 'Malta',
          Code: 'MT',
          Flag: '🇲🇹',
        },
        {
          Name: 'Mauritius',
          Code: 'MU',
          Flag: '🇲🇺',
        },
        {
          Name: 'Maldives',
          Code: 'MV',
          Flag: '🇲🇻',
        },
        {
          Name: 'Malawi',
          Code: 'MW',
          Flag: '🇲🇼',
        },
        {
          Name: 'Mexico',
          Code: 'MX',
          Flag: '🇲🇽',
        },
        {
          Name: 'Malaysia',
          Code: 'MY',
          Flag: '🇲🇾',
        },
        {
          Name: 'Mozambique',
          Code: 'MZ',
          Flag: '🇲🇿',
        },
        {
          Name: 'Namibia',
          Code: 'NA',
          Flag: '🇳🇦',
        },
        {
          Name: 'New Caledonia',
          Code: 'NC',
          Flag: '🇳🇨',
        },
        {
          Name: 'Niger',
          Code: 'NE',
          Flag: '🇳🇪',
        },
        {
          Name: 'Norfolk Island',
          Code: 'NF',
          Flag: '🇳🇫',
        },
        {
          Name: 'Nigeria',
          Code: 'NG',
          Flag: '🇳🇬',
        },
        {
          Name: 'Nicaragua',
          Code: 'NI',
          Flag: '🇳🇮',
        },
        {
          Name: 'Netherlands',
          Code: 'NL',
          Flag: '🇳🇱',
        },
        {
          Name: 'Norway',
          Code: 'NO',
          Flag: '🇳🇴',
        },
        {
          Name: 'Nepal',
          Code: 'NP',
          Flag: '🇳🇵',
        },
        {
          Name: 'Nauru',
          Code: 'NR',
          Flag: '🇳🇷',
        },
        {
          Name: 'Niue',
          Code: 'NU',
          Flag: '🇳🇺',
        },
        {
          Name: 'New Zealand',
          Code: 'NZ',
          Flag: '🇳🇿',
        },
        {
          Name: 'Oman',
          Code: 'OM',
          Flag: '🇴🇲',
        },
        {
          Name: 'Panama',
          Code: 'PA',
          Flag: '🇵🇦',
        },
        {
          Name: 'Peru',
          Code: 'PE',
          Flag: '🇵🇪',
        },
        {
          Name: 'French Polynesia',
          Code: 'PF',
          Flag: '🇵🇫',
        },
        {
          Name: 'Papua New Guinea',
          Code: 'PG',
          Flag: '🇵🇬',
        },
        {
          Name: 'Philippines',
          Code: 'PH',
          Flag: '🇵🇭',
        },
        {
          Name: 'Pakistan',
          Code: 'PK',
          Flag: '🇵🇰',
        },
        {
          Name: 'Poland',
          Code: 'PL',
          Flag: '🇵🇱',
        },
        {
          Name: 'St. Pierre & Miquelon',
          Code: 'PM',
          Flag: '🇵🇲',
        },
        {
          Name: 'Pitcairn Islands',
          Code: 'PN',
          Flag: '🇵🇳',
        },
        {
          Name: 'Puerto Rico',
          Code: 'PR',
          Flag: '🇵🇷',
        },
        {
          Name: 'Palestinian Territories',
          Code: 'PS',
          Flag: '🇵🇸',
        },
        {
          Name: 'Portugal',
          Code: 'PT',
          Flag: '🇵🇹',
        },
        {
          Name: 'Palau',
          Code: 'PW',
          Flag: '🇵🇼',
        },
        {
          Name: 'Paraguay',
          Code: 'PY',
          Flag: '🇵🇾',
        },
        {
          Name: 'Qatar',
          Code: 'QA',
          Flag: '🇶🇦',
        },
        {
          Name: 'Réunion',
          Code: 'RE',
          Flag: '🇷🇪',
        },
        {
          Name: 'Romania',
          Code: 'RO',
          Flag: '🇷🇴',
        },
        {
          Name: 'Serbia',
          Code: 'RS',
          Flag: '🇷🇸',
        },
        {
          Name: 'Russia',
          Code: 'RU',
          Flag: '🇷🇺',
        },
        {
          Name: 'Rwanda',
          Code: 'RW',
          Flag: '🇷🇼',
        },
        {
          Name: 'Saudi Arabia',
          Code: 'SA',
          Flag: '🇸🇦',
        },
        {
          Name: 'Solomon Islands',
          Code: 'SB',
          Flag: '🇸🇧',
        },
        {
          Name: 'Seychelles',
          Code: 'SC',
          Flag: '🇸🇨',
        },
        {
          Name: 'Sudan',
          Code: 'SD',
          Flag: '🇸🇩',
        },
        {
          Name: 'Sweden',
          Code: 'SE',
          Flag: '🇸🇪',
        },
        {
          Name: 'Singapore',
          Code: 'SG',
          Flag: '🇸🇬',
        },
        {
          Name: 'St. Helena',
          Code: 'SH',
          Flag: '🇸🇭',
        },
        {
          Name: 'Slovenia',
          Code: 'SI',
          Flag: '🇸🇮',
        },
        {
          Name: 'Svalbard & Jan Mayen',
          Code: 'SJ',
          Flag: '🇸🇯',
        },
        {
          Name: 'Slovakia',
          Code: 'SK',
          Flag: '🇸🇰',
        },
        {
          Name: 'Sierra Leone',
          Code: 'SL',
          Flag: '🇸🇱',
        },
        {
          Name: 'San Marino',
          Code: 'SM',
          Flag: '🇸🇲',
        },
        {
          Name: 'Senegal',
          Code: 'SN',
          Flag: '🇸🇳',
        },
        {
          Name: 'Somalia',
          Code: 'SO',
          Flag: '🇸🇴',
        },
        {
          Name: 'SuriName',
          Code: 'SR',
          Flag: '🇸🇷',
        },
        {
          Name: 'South Sudan',
          Code: 'SS',
          Flag: '🇸🇸',
        },
        {
          Name: 'São Tomé & Príncipe',
          Code: 'ST',
          Flag: '🇸🇹',
        },
        {
          Name: 'El Salvador',
          Code: 'SV',
          Flag: '🇸🇻',
        },
        {
          Name: 'Sint Maarten',
          Code: 'SX',
          Flag: '🇸🇽',
        },
        {
          Name: 'Syria',
          Code: 'SY',
          Flag: '🇸🇾',
        },
        {
          Name: 'Eswatini',
          Code: 'SZ',
          Flag: '🇸🇿',
        },
        {
          Name: 'Tristan da Cunha',
          Code: 'TA',
          Flag: '🇹🇦',
        },
        {
          Name: 'Turks & Caicos Islands',
          Code: 'TC',
          Flag: '🇹🇨',
        },
        {
          Name: 'Chad',
          Code: 'TD',
          Flag: '🇹🇩',
        },
        {
          Name: 'French Southern Territories',
          Code: 'TF',
          Flag: '🇹🇫',
        },
        {
          Name: 'Togo',
          Code: 'TG',
          Flag: '🇹🇬',
        },
        {
          Name: 'Thailand',
          Code: 'TH',
          Flag: '🇹🇭',
        },
        {
          Name: 'Tajikistan',
          Code: 'TJ',
          Flag: '🇹🇯',
        },
        {
          Name: 'Tokelau',
          Code: 'TK',
          Flag: '🇹🇰',
        },
        {
          Name: 'Timor-Leste',
          Code: 'TL',
          Flag: '🇹🇱',
        },
        {
          Name: 'Turkmenistan',
          Code: 'TM',
          Flag: '🇹🇲',
        },
        {
          Name: 'Tunisia',
          Code: 'TN',
          Flag: '🇹🇳',
        },
        {
          Name: 'Tonga',
          Code: 'TO',
          Flag: '🇹🇴',
        },
        {
          Name: 'Turkey',
          Code: 'TR',
          Flag: '🇹🇷',
        },
        {
          Name: 'Trinidad & Tobago',
          Code: 'TT',
          Flag: '🇹🇹',
        },
        {
          Name: 'Tuvalu',
          Code: 'TV',
          Flag: '🇹🇻',
        },
        {
          Name: 'Taiwan',
          Code: 'TW',
          Flag: '🇹🇼',
        },
        {
          Name: 'Tanzania',
          Code: 'TZ',
          Flag: '🇹🇿',
        },
        {
          Name: 'Ukraine',
          Code: 'UA',
          Flag: '🇺🇦',
        },
        {
          Name: 'Uganda',
          Code: 'UG',
          Flag: '🇺🇬',
        },
        {
          Name: 'U.S. Outlying Islands',
          Code: 'UM',
          Flag: '🇺🇲',
        },
        {
          Name: 'United Nations',
          Code: 'UN',
          Flag: '🇺🇳',
        },
        {
          Name: 'United States',
          Code: 'US',
          Flag: '🇺🇸',
        },
        {
          Name: 'Uruguay',
          Code: 'UY',
          Flag: '🇺🇾',
        },
        {
          Name: 'Uzbekistan',
          Code: 'UZ',
          Flag: '🇺🇿',
        },
        {
          Name: 'Vatican City',
          Code: 'VA',
          Flag: '🇻🇦',
        },
        {
          Name: 'St. Vincent & Grenadines',
          Code: 'VC',
          Flag: '🇻🇨',
        },
        {
          Name: 'Venezuela',
          Code: 'VE',
          Flag: '🇻🇪',
        },
        {
          Name: 'British Virgin Islands',
          Code: 'VG',
          Flag: '🇻🇬',
        },
        {
          Name: 'U.S. Virgin Islands',
          Code: 'VI',
          Flag: '🇻🇮',
        },
        {
          Name: 'Vietnam',
          Code: 'VN',
          Flag: '🇻🇳',
        },
        {
          Name: 'Vanuatu',
          Code: 'VU',
          Flag: '🇻🇺',
        },
        {
          Name: 'Wallis & Futuna',
          Code: 'WF',
          Flag: '🇼🇫',
        },
        {
          Name: 'Samoa',
          Code: 'WS',
          Flag: '🇼🇸',
        },
        {
          Name: 'Kosovo',
          Code: 'XK',
          Flag: '🇽🇰',
        },
        {
          Name: 'Yemen',
          Code: 'YE',
          Flag: '🇾🇪',
        },
        {
          Name: 'Mayotte',
          Code: 'YT',
          Flag: '🇾🇹',
        },
        {
          Name: 'South Africa',
          Code: 'ZA',
          Flag: '🇿🇦',
        },
        {
          Name: 'Zambia',
          Code: 'ZM',
          Flag: '🇿🇲',
        },
        {
          Name: 'Zimbabwe',
          Code: 'ZW',
          Flag: '🇿🇼',
        },
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
