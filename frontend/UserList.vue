<template>
  <adb-page title="Users">
    <button class="btn btn-default" @click="showModal('edit-user-modal')">
      <span class="glyphicon glyphicon-plus"></span>&nbsp;&nbsp;Add New User
    </button>
    <table id="user-list" class="adb-table table table-hover table-striped">
      <thead>
        <tr>
          <th></th>
          <th>Email</th>
          <th>Name</th>
          <th>Roles</th>
          <th>Disabled</th>
        </tr>
      </thead>
      <tbody id="user-list-body">
        <tr v-for="(user, index) in users">
          <td>
            <button
              class="btn btn-default glyphicon glyphicon-pencil"
              @click="showModal('edit-user-modal', user, index)"
            ></button>
          </td>
          <td>{{ user.email }}</td>
          <td>{{ user.name }}</td>
          <td>{{ (user.roles || []).join(', ') }}</td>
          <!-- only disabled  if true to improve page readability -->
          <td><span v-if="user.disabled">Disabled</span></td>
        </tr>
      </tbody>
    </table>
    <modal
      name="edit-user-modal"
      :height="830"
      classes="no-background-color"
      @opened="modalOpened"
      @closed="modalClosed"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title" v-if="currentUser.id">Edit user</h2>
            <h2 class="modal-title" v-if="!currentUser.id">New user</h2>
          </div>
          <div class="modal-body">
            <form action="" id="editUserForm">
              <p>
                <label for="email">Email: </label
                ><input
                  class="form-control"
                  type="text"
                  v-model.trim="currentUser.email"
                  id="email"
                />
              </p>
              <p>
                <label for="name">Name: </label
                ><input
                  class="form-control"
                  type="text"
                  v-model.trim="currentUser.name"
                  id="name"
                />
              </p>
              <p>
                <label for="disabled">Disabled: </label
                ><input
                  class="form-control"
                  type="checkbox"
                  v-model="currentUser.disabled"
                  id="disabled"
                />
              </p>
            </form>

            <div v-if="currentUser.id">
              <p style="margin-top: 20px"></p>
              <h3 class="text-center">Roles</h3>
              <form action="" id="editUserRolesForm">
                <p>
                  <label for="admin_cb">Admin</label>
                  <input
                    class="form-control"
                    type="checkbox"
                    name="admin_cb"
                    id="admin_cb"
                    value="admin"
                    @click="updateUserRoleModal('admin')"
                    v-model="currentUserRoleSelections"
                  />
                </p>
                <p>
                  <label for="org_cb">Organizer</label>
                  <input
                    class="form-control"
                    type="checkbox"
                    name="org_cb"
                    id="org_cb"
                    value="organizer"
                    @click="updateUserRoleModal('organizer')"
                    v-model="currentUserRoleSelections"
                  />
                </p>
                <p>
                  <label for="att_cb">Attendance</label>
                  <input
                    class="form-control"
                    type="checkbox"
                    name="att_cb"
                    id="att_cb"
                    value="attendance"
                    @click="updateUserRoleModal('attendance')"
                    v-model="currentUserRoleSelections"
                  />
                </p>
              </form>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button
              type="button"
              v-bind:disabled="disableConfirmButton"
              class="btn btn-success"
              @click="confirmEditUserModal"
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
// Library from here: https://github.com/euvl/vue-js-modal
import vmodal from 'vue-js-modal';
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage } from './flash_message';
import { Dropdown } from 'uiv';

Vue.use(vmodal);

// Constants related to list ordering
// Corresponds to the constants DescOrder and AscOrder in model/activist.go
const DescOrder = 2;
const AscOrder = 1;

interface User {
  id: number;
  name: string;
  email: string;
  roles: string[];
}

export default Vue.extend({
  name: 'user-list',
  methods: {
    showModal(modalName: string, user: User, index: number) {
      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      // Make shallow copy of selected activist to prevent persisting unsaved
      // edits at the view layer when closing modal
      this.currentUser = { ...user };

      // Track current user index, or default to first in list
      this.userIndex = index === 0 ? 0 : index || -1;

      this.currentModalName = modalName;
      this.$modal.show(modalName);
    },
    hideModal() {
      // Make sure we update the main user list instance of
      // this current user to match the roles.
      // Only when we're not creating a new User.
      if (this.userIndex !== -1) {
        this.users[this.userIndex].roles = this.currentUser.roles;
      }

      if (this.currentModalName) {
        this.$modal.hide(this.currentModalName);
      }
      this.currentModalName = '';
      this.userIndex = -1;
      this.currentUser = {} as User;
      this.currentUserRoleSelections = [];

      // Sort user list by email
      this.sortUserListByEmail();
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

          var parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage('Error: ' + parsed.message, true);
            return;
          }

          flashMessage(this.currentUser.email + ' saved');

          if (this.userIndex === -1) {
            // We're getting a new user, insert them at the top.
            this.users = [parsed.user].concat(this.users);
          } else {
            // We edited an existing user, replace their row in
            // `users`.
            Vue.set(this.users, this.userIndex, parsed.user);
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
    modalOpened() {
      // Add noscroll to body tag so it doesn't scroll while the modal
      // is shown.
      $(document.body).addClass('noscroll');
      this.disableConfirmButton = false;

      // Track current user roles selections separate from currentUser.roles,
      // so we can compare selections when deciding if we need to remove or add a role
      // as we're making selections in the modal edit window.
      this.currentUserRoleSelections = $.extend([], this.currentUser.roles);
    },
    modalClosed() {
      // Allow body to scroll after modal is closed.
      $(document.body).removeClass('noscroll');
    },
    sortByEmail() {
      var order = this.pagingParameters.order;
      if (order === AscOrder) {
        order = DescOrder;
      } else {
        order = AscOrder;
      }
      this.reset();
      this.pagingParameters.order = order;
    },
    sortUserListByEmail() {
      if (!this.users) {
        return;
      }

      this.users.sort(function (a, b) {
        var emailA = a.email.toLowerCase();
        var emailB = b.email.toLowerCase();

        if (emailA < emailB) {
          return -1;
        }

        if (emailA > emailB) {
          return 1;
        }

        return 0;
      });
    },
    setUsers(users: User[]) {
      console.log('inside setUsers: ', users);
      this.users = users;
    },
    reset() {
      // reset data properties back to original values
      this.currentUser = {} as User;
      this.users = [];
      this.userIndex = -1;
      this.disableConfirmButton = false;
      this.currentModalName = '';
      this.pagingParameters = {
        email: '',
        order: AscOrder,
        limit: 40,
      };

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

          var parsed = JSON.parse(data);

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
      currentUser: {} as User,
      users: [] as User[],
      userIndex: -1,
      disableConfirmButton: false,
      currentModalName: '',
      pagingParameters: {
        email: '',
        order: AscOrder,
        limit: 40,
      },
      currentUserRoleSelections: [] as string[],
    };
  },
  created() {
    $.ajax({
      url: '/user/list',
      success: (data) => {
        var parsed = JSON.parse(data);
        if (parsed.status === 'error') {
          flashMessage('Error: ' + parsed.message, true);
          return;
        }
        // status === "success"
        this.setUsers(parsed);
      },
      error: () => {
        flashMessage('Error connecting to server.', true);
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
