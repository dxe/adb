<template>
  <adb-page title="Users">
    <b-loading :is-full-page="true" v-model="loading"></b-loading>
    <b-loading :is-full-page="true" v-model="loadingChapters"></b-loading>
    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <b-button
            icon-left="account-plus"
            type="is-primary"
            @click="showModal('edit-user-modal')"
          >
            New user
          </b-button>
        </div>
      </div>
    </nav>

    <b-table :data="users" striped hoverable default-sort="name">
      <b-table-column v-slot="props">
        <div style="width: 30px;">
          <b-button @click="showModal('edit-user-modal', props.row)">
            <b-icon icon="pencil" type="is-primary"></b-icon>
          </b-button>
        </div>
      </b-table-column>
      <b-table-column field="name" label="Name" v-slot="props" sortable>
        {{ props.row.name }}
      </b-table-column>
      <b-table-column field="email" label="Email" v-slot="props" sortable>
        {{ props.row.email }}
      </b-table-column>
      <b-table-column field="chapter_id" label="Chapter" v-slot="props" sortable>
        {{ chapterName(props.row.chapter_id) }}
      </b-table-column>
      <b-table-column
        field="
          chapterName(props.row.chapter_id) === 'SF Bay Area'
            ? (props.row.roles || []).join(', ')
            : 'user'"
        label="Roles"
        v-slot="props"
        sortable
      >
        {{
          chapterName(props.row.chapter_id) === 'SF Bay Area'
            ? (props.row.roles || []).join(', ')
            : 'user'
        }}
      </b-table-column>
      <b-table-column field="disabled" label="Disabled" v-slot="props" sortable>
        <b-icon icon="check" type="is-danger" v-if="props.row.disabled"></b-icon>
      </b-table-column>
    </b-table>

    <b-modal
      :active="currentModalName === 'edit-user-modal'"
      has-modal-card
      :destroy-on-hide="true"
      scroll="keep"
      :can-cancel="false"
      :full-screen="isMobile()"
    >
      <div class="modal-card" style="width: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ currentUser.id ? 'Edit' : 'New' }} user</p>
        </header>
        <section class="modal-card-body">
          <b-field label="Name" label-position="on-border">
            <b-input type="text" v-model.trim="currentUser.name" icon="account" required></b-input>
          </b-field>
          <b-field label="Email" label-position="on-border">
            <b-input type="email" v-model.trim="currentUser.email" icon="email" required></b-input>
          </b-field>
          <b-field label="Chapter" label-position="on-border">
            <b-select v-model.number="currentUser.chapter_id" expanded>
              <option v-for="chap in chapters" :value="chap.ID" :key="chap.ID">
                {{ chap.Name }}
              </option>
            </b-select>
          </b-field>
          <b-field>
            <b-switch v-model="currentUser.disabled" type="is-danger">Disabled</b-switch>
          </b-field>

          <div v-if="chapterName(currentUser.chapter_id) === 'SF Bay Area'">
            <b-field label="Roles" custom-class="has-text-primary"></b-field>
            <template v-if="!currentUser.id">
              Please save the new user before adding roles.
            </template>
            <template v-if="currentUser.id">
              <b-field>
                <b-checkbox
                  v-model="currentUserRoleSelections"
                  native-value="admin"
                  @click.native="updateUserRoleModal('admin')"
                  >Admin</b-checkbox
                >
              </b-field>

              <b-field>
                <b-checkbox
                  v-model="currentUserRoleSelections"
                  native-value="organizer"
                  @click.native="updateUserRoleModal('organizer')"
                  >Organizer</b-checkbox
                >
              </b-field>

              <b-field>
                <b-checkbox
                  v-model="currentUserRoleSelections"
                  native-value="attendance"
                  @click.native="updateUserRoleModal('attendance')"
                  >Attendance</b-checkbox
                >
              </b-field>
            </template>
          </div>
        </section>
        <footer class="modal-card-foot">
          <b-button label="Cancel" icon-left="cancel" @click="hideModal" />
          <b-button
            label="Save"
            icon-left="floppy"
            type="is-primary"
            :disabled="disableConfirmButton"
            @click="confirmEditUserModal"
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

interface User {
  id: number;
  name: string;
  email: string;
  roles: string[];
  chapter_id: number;
}

interface Chapter {
  ID: number;
  Name: string;
  Region?: string;
}

export default Vue.extend({
  name: 'user-list',
  methods: {
    isMobile() {
      return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
        navigator.userAgent,
      );
    },
    chapterName(id: number) {
      const chapters = this.chapters.filter((c) => {
        return c.ID === id;
      });
      if (chapters.length > 0) {
        return chapters[0].Name;
      }
      return '';
    },
    showModal(modalName: string, user: User) {
      // Hide the navbar so that the model doesn't go behind it.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'hidden';

      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      // Make shallow copy of selected activist to prevent persisting unsaved
      // edits at the view layer when closing modal
      this.currentUser = { ...user };

      // Get the index for updating the view w/o refreshing the whole page.
      this.userIndex = this.users.findIndex((u) => {
        return u.id === this.currentUser.id;
      });

      this.currentModalName = modalName;

      this.disableConfirmButton = false;

      // Track current user roles selections separate from currentUser.roles,
      // so we can compare selections when deciding if we need to remove or add a role
      // as we're making selections in the modal edit window.
      this.currentUserRoleSelections = $.extend([], this.currentUser.roles);
    },
    hideModal() {
      // Show the navbar.
      const mainNav = document.getElementById('mainNav');
      if (mainNav) mainNav.style.visibility = 'visible';

      // Make sure we update the main user list instance of
      // this current user to match the roles.
      // Only when we're not creating a new User.
      if (this.userIndex !== -1) {
        this.users[this.userIndex].roles = this.currentUser.roles;
      }

      this.currentModalName = '';
      this.userIndex = -1;
      this.currentUser = {} as User;
      this.currentUserRoleSelections = [];
    },
    confirmEditUserModal() {
      // Disable the save button when the user clicks it so they don't
      // try to save twice. Re-enable it when we get any response back
      // from the server (even an error).
      this.disableConfirmButton = true;
      const csrfToken = $('meta[name="csrf-token"]').attr('content');

      $.ajax({
        url: '/user/save',
        method: 'POST',
        headers: { 'X-CSRF-Token': csrfToken },
        contentType: 'application/json',
        data: JSON.stringify(this.currentUser),
        success: (data) => {
          this.disableConfirmButton = false;

          const parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }

          flashMessage(this.currentUser.email + ' saved');

          if (this.userIndex === -1) {
            // We're saving a new user. Insert them at the top,
            // then open the modal again for adding roles if SF Bay.
            this.users = [parsed.user].concat(this.users);
            this.hideModal();
            if (this.chapterName(parsed.user.chapter_id) === 'SF Bay Area') {
              this.showModal('edit-user-modal', parsed.user);
            }
          } else {
            // We edited an existing user, replace their row in
            // `users` & close the modal.
            Vue.set(this.users, this.userIndex, parsed.user);
            this.hideModal();
          }
        },
        error: (err) => {
          this.disableConfirmButton = false;

          console.warn(err.responseText);
          flashMessage('Server error: ' + err.responseText, true);
        },
      });
    },
    setUsers(users: User[]) {
      this.users = users;
    },
    reset() {
      // reset data properties back to original values
      this.currentUser = {} as User;
      this.users = [];
      this.userIndex = -1;
      this.disableConfirmButton = false;
      this.currentModalName = '';
      this.currentUserRoleSelections = [];
    },
    updateUserRoleModal(role: string) {
      if (this.disableConfirmButton) {
        return;
      }

      if (!role) {
        return;
      }

      if (!this.currentUser.roles) {
        this.currentUser.roles = [];
      }

      this.disableConfirmButton = true;

      // If the specified Role already exists in the Current User's role list,
      // then we assume the role should be removed.
      const existingRole = this.currentUser.roles.indexOf(role) >= 0;
      const csrfToken = $('meta[name="csrf-token"]').attr('content');

      $.ajax({
        url: existingRole ? '/users-roles/remove' : '/users-roles/add',
        method: 'POST',
        headers: { 'X-CSRF-Token': csrfToken },
        contentType: 'application/json',
        data: JSON.stringify({
          user_id: this.currentUser.id,
          role: role,
        }),
        success: (data) => {
          this.disableConfirmButton = false;

          const parsed = JSON.parse(data);

          if (parsed.status === 'error') {
            this.currentUserRoleSelections = $.extend([], this.currentUser.roles);
            flashMessage(
              "And error occurred while updating this User's Role. Reverting Role Selections back to original: " +
                parsed.message,
              true,
            );
            return;
          }

          flashMessage(
            (existingRole ? 'Removed' : 'Added') +
              ' the ' +
              role +
              ' role for ' +
              this.currentUser.email,
          );

          // Sync the currentUser.roles with the current modal selections.
          this.currentUser.roles = this.currentUserRoleSelections;
        },
        error: (err) => {
          this.disableConfirmButton = false;
          this.currentUserRoleSelections = $.extend([], this.currentUser.roles);
          console.warn(err.responseText);
          flashMessage(
            'Server error. Reverting Role Selections back to original: ' + err.responseText,
            true,
          );
        },
      });
    },
  },
  data() {
    return {
      loading: false,
      loadingChapters: false,
      currentUser: {} as User,
      users: [] as User[],
      userIndex: -1,
      disableConfirmButton: false,
      currentModalName: '',
      currentUserRoleSelections: [] as string[],
      chapters: [] as Chapter[],
    };
  },
  created() {
    this.loading = true;
    $.ajax({
      url: '/user/list',
      success: (data) => {
        const parsed = JSON.parse(data);
        if (parsed.status === 'error') {
          flashMessage('Error: ' + parsed.message, true);
          return;
        }
        // status === "success"
        this.setUsers(parsed);
        this.loading = false;
      },
      error: () => {
        flashMessage('Error connecting to server.', true);
        this.loading = false;
      },
    });
    this.loadingChapters = true;
    $.ajax({
      url: '/chapters',
      method: 'GET',
      dataType: 'json',
      success: (data) => {
        this.chapters = data.filter((c: Chapter) => {
          return c.Region !== 'Online';
        });
        this.chapters.unshift({ ID: 0, Name: 'None' });
        this.loadingChapters = false;
      },
      error: () => {
        flashMessage('Error: could not load chapters', true);
        this.loadingChapters = false;
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
