<template>
  <adb-page title="Chapters" class="body-wrapper-wide">
    <div class="form-inline">
      <button class="btn btn-default" @click="showModal('edit-chapter-modal')">
        <span class="glyphicon glyphicon-plus"></span>&nbsp;&nbsp;Add New Chapter
      </button>

      &nbsp;&nbsp;

      <button
        class="btn btn-default glyphicon glyphicon-question-sign"
        onclick="alert('Help feature coming soon.')"
      ></button>

      &nbsp;&nbsp;

      <label for="region">Filter by mentor: </label>
      <select
        id="mentor-filter"
        class="form-control form-inline"
        style="width: 125px"
        v-model="mentorFilter"
      >
        <option v-for="mentor in mentors" v-bind:value="mentor">{{ mentor }}</option>
      </select>

      &nbsp;&nbsp;

      <span>Total chapters: {{ sortedChapters.length }}</span>
      &nbsp;&nbsp;
      <span
        >Active chapters:
        {{
          sortedChapters.filter((c) => {
            return dateInLastThreeMonths(c.LastAction);
          }).length
        }}</span
      >
    </div>

    <br />

    <table id="working-group-list" class="adb-table table table-hover table-striped">
      <thead>
        <tr>
          <th class="sticky"></th>
          <th class="sticky"></th>
          <th class="sticky"></th>
          <th @click="sort('Name')" class="cursor-pointer sticky">Name</th>
          <th @click="sort('Mentor')" class="cursor-pointer sticky">Mentor</th>
          <th @click="sort('LastContact')" class="cursor-pointer sticky">Last Contact</th>
          <th @click="sort('LastAction')" class="cursor-pointer sticky">Last Action</th>
          <th @click="sort('LastFBEvent')" class="cursor-pointer sticky">Last FB Event</th>
          <th @click="sort('LastFBSync')" class="cursor-pointer sticky">FB Sync Status</th>
        </tr>
      </thead>
      <tbody id="working-group-list-body">
        <tr v-for="chapter in sortedChapters">
          <td style="width: 1px; padding: 8px 2px 8px 5px;">
            <button
              class="btn btn-default glyphicon glyphicon-pencil"
              @click="showModal('edit-chapter-modal', chapter)"
            ></button>
          </td>
          <td style="width: 1px; padding: 8px 2px;">
            <button
              class="btn btn-default glyphicon glyphicon-envelope"
              @click="composeEmail(chapter)"
            ></button>
          </td>
          <td style="width: 1px; padding: 8px 5px 8px 2px">
            <dropdown>
              <button
                data-role="trigger"
                class="btn btn-default dropdown-toggle glyphicon glyphicon-option-horizontal"
                type="button"
              ></button>
              <template slot="dropdown">
                <li>
                  <a @click="showModal('delete-chapter-modal', chapter)">Delete Chapter</a>
                </li>
              </template>
            </dropdown>
          </td>
          <td>{{ chapter.Flag }} {{ chapter.Name }}</td>
          <td>{{ chapter.Mentor }}</td>
          <td class="cursor-pointer" @click="showModal('chapter-last-contact-modal', chapter)">
            <a style="text-decoration: none; color: black; font-weight: normal">
              <small>{{ chapter.LastContact || 'None' }}</small>
            </a>
          </td>
          <td
            class="cursor-pointer"
            @click="showModal('chapter-last-action-modal', chapter)"
            v-html="colorQuarterlyGoal(chapter.LastAction)"
          ></td>
          <td v-html="colorQuarterlyGoal(chapter.LastFBEvent)"></td>
          <td v-html="colorFBSyncStatus(chapter.LastFBSync)"></td>
        </tr>
      </tbody>
    </table>
    <modal
      name="delete-chapter-modal"
      height="auto"
      classes="no-background-color no-top"
      @opened="modalOpened"
      @closed="modalClosed"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header"><h2 class="modal-title">Delete chapter</h2></div>
          <div class="modal-body">
            <p>
              Are you sure you want to delete {{ currentChapter.Flag }}
              <strong>{{ currentChapter.Name }}</strong
              >?
            </p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Cancel</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-danger"
              @click="confirmDeleteChapterModal"
            >
              Delete chapter
            </button>
          </div>
        </div>
      </div>
    </modal>
    <modal
      name="edit-chapter-modal"
      height="auto"
      width="1000"
      classes="no-background-color no-top"
      @opened="modalOpened"
      @closed="modalClosed"
      :clickToClose="false"
    >
      <div class="modal-dialog" style="width: 1000px;">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title" v-if="currentChapter.ChapterID">
              {{ currentChapter.Flag }} {{ currentChapter.Name }}
            </h2>
            <h2 class="modal-title" v-if="!currentChapter.ChapterID">New chapter</h2>
          </div>
          <div class="modal-body">
            <form action="" id="editChapterForm">
              <div class="form-row">
                <div class="col-xs-2" style="margin-top: 10px;">
                  <label for="flag">Flag: </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.Flag"
                    id="flag"
                    maxlength="2"
                  />
                </div>

                <div class="col-xs-6" style="margin-top: 10px;">
                  <label for="flag">Name: </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.Name"
                    id="chapter-name"
                    maxlength="100"
                    :readonly="currentChapter.ChapterID"
                  />
                </div>

                <div class="col-xs-4" style="margin-top: 10px;">
                  <label for="mentor">Mentor: </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.Mentor"
                    id="mentor"
                    maxlength="100"
                  />
                </div>
              </div>

              <div class="form-row">
                <div class="col-xs-6" style="margin-top: 10px;">
                  <label for="facebook">Facebook: </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.FbURL"
                    id="facebook"
                    maxlength="100"
                  />
                </div>
                <div class="col-xs-6" style="margin-top: 10px;">
                  <label for="twitter">Twitter: </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.TwitterURL"
                    id="twitter"
                    maxlength="100"
                  />
                </div>
              </div>
              <div class="form-row">
                <div class="col-xs-6" style="margin-top: 10px;">
                  <label for="instagram">Instagram: </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.InstaURL"
                    id="instagram"
                    maxlength="100"
                  />
                </div>
                <div class="col-xs-6" style="margin-top: 10px;">
                  <label for="email">Email (Public): </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.Email"
                    id="email"
                    maxlength="100"
                  />
                </div>
              </div>
              <div class="form-row">
                <div class="col-xs-4" style="margin-top: 10px;">
                  <label for="region">Region: </label
                  ><select id="region" class="form-control" v-model="currentChapter.Region">
                    <option value="North America">North America</option>
                    <option value="Central & South America">Central & South America</option>
                    <option value="Europe">Europe</option>
                    <option value="Middle East & Africa">Middle East & Africa</option>
                    <option value="Asia-Pacific">Asia-Pacific</option>
                    <option value="Online">Online</option>
                  </select>
                </div>

                <div class="col-xs-4" style="margin-top: 10px;">
                  <label for="country">Country: </label
                  ><input
                    class="form-control"
                    type="text"
                    maxlength="128"
                    v-model.trim="currentChapter.Country"
                    id="country"
                  />
                </div>

                <div class="col-xs-2" style="margin-top: 10px;">
                  <label for="lat">Lat: </label
                  ><input
                    class="form-control"
                    type="number"
                    v-model.number="currentChapter.Lat"
                    id="lat"
                    placeholder="00.000000"
                    step="0.000001"
                    min="-90"
                    max="90"
                  />
                </div>
                <div class="col-xs-2" style="margin-top: 10px;">
                  <label for="lng">Lng: </label
                  ><input
                    class="form-control"
                    type="number"
                    v-model.number="currentChapter.Lng"
                    id="lng"
                    placeholder="000.000000"
                    step="0.000001"
                    min="-180"
                    max="180"
                  />
                </div>
              </div>

              <!-- TODO: decide whether to show this row or modal or remove it -->
              <div class="form-row">
                <div class="col-xs-6" style="margin-top: 10px;">
                  <label for="last-contact">Last Contact: </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.LastContact"
                    id="last-contact"
                  />
                </div>
                <div class="col-xs-6" style="margin-top: 10px;">
                  <label for="last-action">Last Action: </label
                  ><input
                    class="form-control"
                    type="text"
                    v-model.trim="currentChapter.LastAction"
                    id="last-action"
                  />
                </div>
              </div>
              <div class="form-row">
                <div class="col-xs-12" style="margin-top: 10px;">
                  <label for="notes">Notes: </label
                  ><textarea
                    class="form-control"
                    maxlength="512"
                    v-model.trim="currentChapter.Notes"
                    id="notes"
                  />
                </div>
              </div>

              <div class="form-row" v-if="currentChapter.ChapterID">
                <div class="col-xs-12" style="margin-top: 10px;">
                  <label>Organizers: </label>
                </div>

                <div class="form-row" v-for="(organizer, index) in currentChapter.Organizers">
                  <div class="col-xs-3" style="margin-top: 5px;">
                    <input
                      class="form-control"
                      type="text"
                      placeholder="Name"
                      v-model.trim="organizer.Name"
                    />
                  </div>

                  <div class="col-xs-3" style="margin-top: 5px;">
                    <input
                      class="form-control"
                      type="text"
                      placeholder="Email"
                      v-model.trim="organizer.Email"
                    />
                  </div>

                  <div class="col-xs-2" style="margin-top: 5px;">
                    <input
                      class="form-control"
                      type="text"
                      placeholder="Phone"
                      v-model.trim="organizer.Phone"
                    />
                  </div>

                  <div class="col-xs-3" style="margin-top: 5px;">
                    <input
                      class="form-control"
                      type="text"
                      placeholder="Facebook"
                      v-model.trim="organizer.Facebook"
                    />
                  </div>

                  <div class="col-xs-1" style="margin-top: 5px; padding: 0px;">
                    <a
                      class="form-control btn btn-danger cursor-pointer"
                      @click="deleteOrganizer(index)"
                      style="color: white"
                      ><span class="glyphicon glyphicon-trash"></span
                    ></a>
                  </div>
                </div>
                <div class="col-xs-12" style="margin-top: 10px;">
                  <a class="btn btn-primary btn-sm" @click="addOrganizer">Add</a>
                </div>
              </div>

              <div v-if="showMoreOptions">
                <div class="form-row">
                  <div class="col-xs-6" style="margin-top: 10px;">
                    <label for="id">Facebook ID: </label
                    ><input
                      class="form-control"
                      type="number"
                      maxlength="16"
                      v-model.number="currentChapter.ID"
                      id="id"
                    />
                  </div>
                  <div class="col-xs-6" style="margin-top: 10px;">
                    <label for="token">Facebook Token: </label
                    ><input
                      class="form-control"
                      type="text"
                      maxlength="200"
                      v-model.trim="currentChapter.Token"
                      id="token"
                    />
                  </div>
                </div>

                <div class="form-row">
                  <div class="col-xs-6" style="margin-top: 10px;">
                    <label for="eventbrite-id">Eventbrite ID: </label
                    ><input
                      class="form-control"
                      type="number"
                      maxlength="16"
                      v-model.trim="currentChapter.EventbriteID"
                      id="eventbrite-id"
                    />
                  </div>
                  <div class="col-xs-6" style="margin-top: 10px;">
                    <label for="eventbrite-token">Eventbrite Token: </label
                    ><input
                      class="form-control"
                      type="text"
                      maxlength="200"
                      v-model.trim="currentChapter.EventbriteToken"
                      id="eventbrite-token"
                    />
                  </div>
                </div>

                <div class="form-row">
                  <div class="col-xs-4" style="margin-top: 10px;">
                    <label for="ml-type">Mailing List Type: </label
                    ><select
                      id="ml-type"
                      class="form-control"
                      v-model="currentChapter.MailingListType"
                    >
                      <option value="">None</option>
                      <option value="Sendy">Sendy</option>
                      <option value="Google Groups">Google Groups</option>
                    </select>
                  </div>
                  <div class="col-xs-4" style="margin-top: 10px;">
                    <label for="ml-id">Mailing List ID: </label
                    ><input
                      class="form-control"
                      type="text"
                      maxlength="100"
                      v-model.trim="currentChapter.MailingListID"
                      id="ml-id"
                    />
                  </div>
                  <div class="col-xs-4" style="margin-top: 10px;">
                    <label for="ml-rad">Mailing List Radius: </label
                    ><input
                      class="form-control"
                      type="number"
                      min="0"
                      max="500"
                      v-model.number="currentChapter.MailingListRadius"
                      id="ml-rad"
                    />
                  </div>
                </div>
              </div>
              <!-- This p is needed to make the modal footer's top border display properly. -->
              <p>&nbsp;</p>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-warning" @click="toggleShowMoreOptions">
              Advanced options
            </button>
            &nbsp;&nbsp;
            <button type="button" class="btn btn-secondary" @click="hideModal">Cancel</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-success"
              @click="confirmEditChapterModal"
            >
              Save changes
            </button>
          </div>
        </div>
      </div>
    </modal>
    <modal
      name="chapter-last-contact-modal"
      height="auto"
      classes="no-background-color no-top"
      @opened="modalOpened"
      @closed="modalClosed"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Update Last Contact</h2>
            <h4>{{ currentChapter.Flag }} {{ currentChapter.Name }}</h4>
          </div>
          <div class="modal-body">
            <form action="" id="updateLastContactForm">
              <p>
                <input
                  class="form-control"
                  type="date"
                  v-model.trim="currentChapter.LastContact"
                  id="last-contact-picker"
                  v-focus
                />
              </p>
              <p>
                <button
                  class="btn btn-xs btn-primary"
                  style="margin: 0px 10px"
                  v-on:click.prevent="setDateToToday('LastContact')"
                >
                  today
                </button>
              </p>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Cancel</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-success"
              @click="confirmEditChapterModal"
            >
              Save changes
            </button>
          </div>
        </div>
      </div>
    </modal>
    <modal
      name="chapter-last-action-modal"
      height="auto"
      classes="no-background-color no-top"
      @opened="modalOpened"
      @closed="modalClosed"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Update Last Action</h2>
            <h4>{{ currentChapter.Flag }} {{ currentChapter.Name }}</h4>
          </div>
          <div class="modal-body">
            <form action="" id="updateLastActionForm">
              <p>
                <input
                  class="form-control"
                  type="date"
                  v-model.trim="currentChapter.LastAction"
                  id="last-action-picker"
                  v-focus
                />
              </p>
              <p>
                <button
                  class="btn btn-xs btn-primary"
                  style="margin: 0px 10px"
                  v-on:click.prevent="setDateToToday('LastAction')"
                >
                  today
                </button>
              </p>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Cancel</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-success"
              @click="confirmEditChapterModal"
            >
              Save changes
            </button>
          </div>
        </div>
      </div>
    </modal>
  </adb-page>
</template>

<script lang="ts">
import vmodal from 'vue-js-modal';
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage } from './flash_message';
//@ts-ignore
import { Dropdown } from 'uiv';
import { focus } from './directives/focus';
import moment from 'moment';

Vue.use(vmodal);

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
  LastFBSync: string; // TODO: use string or Date here?
  LastFBEvent: string; // TODO: use string or Date here?
  EventbriteID: string;
  EventbriteToken: string;
  Mentor: string;
  Country: string;
  Notes: string;
  LastContact: string; // TODO: use string or Date here?
  LastAction: string; // TODO: use string or Date here?
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
    sortedChapters: function(): Chapter[] {
      const sorted = this.chapters.sort((a: any, b: any) => {
        let modifier = 1;
        if (this.currentSortDir === 'desc') modifier = -1;
        if (a[this.currentSort] < b[this.currentSort]) return -1 * modifier;
        if (a[this.currentSort] > b[this.currentSort]) return modifier;
        return 0;
      });
      if (this.mentorFilter === 'All') {
        return sorted;
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
    showModal(modalName: string, chapter: Chapter) {
      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      this.currentChapter = { ...chapter };

      this.currentChapterIndex = this.chapters.findIndex((c) => {
        return c.ChapterID === this.currentChapter.ChapterID;
      });

      this.currentModalName = modalName;
      this.$modal.show(modalName);
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
      if (this.currentModalName) {
        this.$modal.hide(this.currentModalName);
      }
      this.currentModalName = '';
      this.currentChapter = {} as Chapter;
      this.currentChapterIndex = -1;
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
    modalOpened() {
      $(document.body).addClass('noscroll');
      this.disableConfirmButton = false;
    },
    modalClosed() {
      $(document.body).removeClass('noscroll');
    },
    addOrganizer() {
      this.currentChapter.Organizers = this.currentChapter.Organizers
        ? this.currentChapter.Organizers
        : [];
      this.currentChapter.Organizers.push({} as Organizer);
    },
    deleteOrganizer(index: number) {
      this.currentChapter.Organizers.splice(index, 1);
    },
    toggleShowMoreOptions() {
      this.showMoreOptions = !this.showMoreOptions;
    },
    sort(s: string) {
      if (s === this.currentSort) {
        this.currentSortDir = this.currentSortDir === 'asc' ? 'desc' : 'asc';
      }
      this.currentSort = s;
    },
    setDateToToday(field: string) {
      if (field === 'LastContact') {
        this.currentChapter.LastContact = moment()
          .local()
          .format('YYYY-MM-DD');
        return;
      }
      if (field === 'LastAction') {
        this.currentChapter.LastAction = moment()
          .local()
          .format('YYYY-MM-DD');
      }
    },
    colorFBSyncStatus(text: string) {
      const time = moment(text).add(8, 'hour'); // this converts our DB time for this field to UTC
      let color = 'grey';
      if (time.isValid()) {
        color = 'red';
      }
      if (time.isAfter(moment().add(-1, 'day'))) {
        color = 'yellow';
      }
      if (time.isAfter(moment().add(-1, 'hour'))) {
        color = 'green';
      }
      return `<div class="dot bg-${color}"><div>`;
    },
    colorQuarterlyGoal(text: string) {
      const time = moment(text);
      let color = 'grey';
      let timeStr = 'None';
      if (time.isValid()) {
        timeStr = time.format('YYYY-MM-DD');
        color = 'red';
      }
      if (time.isAfter(moment().add(-3, 'month'))) {
        color = 'green';
      }
      return `<div class="${color}"><small>${timeStr}</small></div>`;
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
      currentSort: 'Name',
      currentSortDir: 'asc',
      mentorFilter: 'All',
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
        this.chapters = parsed.chapters;
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
      },
    });
  },
  components: {
    AdbPage,
    Dropdown,
  },
  directives: {
    focus,
  },
});
</script>

<style>
.dot {
  height: 25px;
  width: 25px;
  border-radius: 50%;
  display: inline-block;
}
.bg-green {
  background-color: green;
}
.bg-yellow {
  background-color: yellow;
}
.bg-red {
  background-color: red;
}
.bg-grey {
  background-color: grey;
}
.green {
  color: green;
}
.red {
  color: red;
}
.cursor-pointer {
  cursor: pointer;
}
th.sticky {
  position: sticky;
  top: 50px;
  background-color: white;
  z-index: 1000;
}
</style>
