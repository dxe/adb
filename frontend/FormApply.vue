<template>
  <adb-page title="Join the SF Bay Area Chapter">
    <div v-if="!local">
      <b-field label="Do you live within 100 miles of Berkeley, CA?">
        <b-radio-button v-model="local" native-value="true" type="is-primary" required>
          <span>Yes</span>
        </b-radio-button>
        <b-radio-button
          v-model="local"
          native-value="false"
          type="is-primary"
          required
          @click.native="notLocal"
        >
          <span>No</span>
        </b-radio-button>
      </b-field>
    </div>

    <transition name="fade">
      <div v-if="submitSuccess" class="content">
        <h2>Thank you!</h2>
        <p>
          An organizer will reach out to you within the next week and let you know what next steps
          need to be taken for your application to be approved!
        </p>
      </div>
    </transition>

    <div v-if="local">
      <div v-if="!showForm" class="content">
        <p>
          Decades of research into historic movements – like the Civil Rights Movement or Gay Rights
          Movement – tells us that the most important factor in creating social and political change
          is sustained participation in nonviolent direct action. We also know that sustained
          participation is more likely when activists are embedded in a strong, supportive
          community. When you become a chapter member, you're committing to being an integral part
          of this activist community and to coming to one direct action event (like a protest,
          outreach event, or sanctuary workday) every month.
        </p>

        <h2>How to become a Chapter Member</h2>
        <ul>
          <li>
            Attend DxE’s
            <a href="http://dxe.io/workshop" target="_blank">intro workshop</a> (held monthly) to
            learn about our community and theory of change
          </li>
          <li>Take the Movement Power Pledge below (attend one action each month)</li>
        </ul>

        <h3>Benefits of becoming a Chapter Member</h3>
        <ul>
          <li>Voting rights (after sustained participation in the chapter)</li>
          <li>Access to chapter member mailing lists, Discord server, and Facebook group</li>
          <li>Discounts at local and online businesses</li>
          <li>
            Added to a small community group with other chapter members in your neighborhood/area
          </li>
          <li>Invites to special events and opportunities</li>
          <li>
            Your photo posted at the Berkeley Animal Rights Center and on the virtual chapter member
            wall
          </li>
        </ul>

        <h3>Expectations of Chapter Members</h3>
        <ul>
          <li>
            Uphold our
            <a
              href="https://docs.google.com/document/d/e/2PACX-1vSTt4SXGl-eARvOpTzOM8Nzcs9uo1ZVkCuqD0doRkE6zCpx6NgMbVqYwCh691GLKZ5-hklnsaPfHlfa/pub"
              target="_blank"
              >code of conduct</a
            >
            and
            <a href="https://www.directactioneverywhere.com/core-values" target="_blank">values</a>
          </li>
          <li>Uphold the Movement Power Pledge (attend one action every month)</li>
        </ul>

        <br />

        <b-button class="is-primary" @click="applyButtonClicked"> Apply now </b-button>
      </div>

      <div v-if="showForm && !submitSuccess" class="content">
        <h2 class="mt-6">Take The Movement Power Pledge</h2>

        <p>
          By entering your name below, you are pledging to come to at least one action each month.
        </p>

        <div class="columns is-multiline">
          <div class="column is-half">
            <b-field label="First Name" label-position="on-border">
              <b-input
                type="text"
                v-model.trim="firstName"
                required
                maxlength="35"
                ref="firstName"
              ></b-input>
            </b-field>
          </div>

          <div class="column is-half">
            <b-field label="Last Name" label-position="on-border">
              <b-input
                type="text"
                v-model.trim="lastName"
                required
                maxlength="35"
                ref="lastName"
              ></b-input>
            </b-field>
          </div>

          <div class="column is-half">
            <p>
              I will uphold DxE's
              <a
                href="https://docs.google.com/document/d/e/2PACX-1vSTt4SXGl-eARvOpTzOM8Nzcs9uo1ZVkCuqD0doRkE6zCpx6NgMbVqYwCh691GLKZ5-hklnsaPfHlfa/pub"
                target="_blank"
                >code of conduct</a
              >.
            </p>

            <b-field class="mt-3">
              <b-radio-button v-model="conduct" native-value="true" type="is-success" required>
                <b-icon icon="check"></b-icon>
                <span>Yes, I agree with the above statement.</span>
              </b-radio-button>
            </b-field>
          </div>

          <div class="column is-half">
            <p>
              I support DxE’s
              <a href="https://www.directactioneverywhere.com/core-values" target="_blank"
                >mission and values</a
              >.
            </p>

            <b-field class="mt-3">
              <b-radio-button v-model="mission" native-value="true" type="is-success" required>
                <b-icon icon="check"></b-icon>
                <span>Yes, I agree with the above statement.</span>
              </b-radio-button>
            </b-field>
          </div>

          <div class="column is-full mt-3">
            <h2>Contact Info</h2>
          </div>

          <div class="column is-full">
            <b-field label="Email" label-position="on-border">
              <b-input
                type="email"
                v-model.trim="email"
                required
                maxlength="60"
                ref="email"
              ></b-input>
            </b-field>
          </div>

          <div class="column is-full">
            <b-field label="Street Address" label-position="on-border">
              <b-input
                type="text"
                v-model.trim="address"
                required
                maxlength="60"
                ref="address"
              ></b-input>
            </b-field>
          </div>

          <div class="column is-half">
            <b-field label="City" label-position="on-border">
              <b-input type="text" v-model.trim="city" required maxlength="90" ref="city"></b-input>
            </b-field>
          </div>

          <div class="column is-half">
            <b-field label="Zip Code" label-position="on-border">
              <b-input type="number" v-model.trim="zip" required maxlength="5" ref="zip"></b-input>
            </b-field>
          </div>

          <div class="column is-half">
            <b-field label="Phone" label-position="on-border">
              <b-input
                type="text"
                v-model.trim="phone"
                required
                maxlength="20"
                ref="phone"
              ></b-input>
            </b-field>
          </div>

          <div class="column is-half">
            <b-field label="Birthday" label-position="on-border">
              <b-input
                type="date"
                v-model.trim="birthday"
                required
                maxlength="10"
                ref="birthday"
              ></b-input>
            </b-field>
          </div>

          <div class="column is-full">
            <b-field label="Who encouraged you to apply? (optional)" label-position="on-border">
              <b-input type="text" v-model.trim="referral" maxlength="100"></b-input>
            </b-field>
          </div>

          <div class="column is-full">
            <b-field
              label="Are you interested in further leveling up your activism by becoming an
                  Organizer?"
            >
              <small>
                Organizers take ownership over achieving the chapter’s objectives and make our
                chapter function by organizing community events, editing videos, leading protests,
                raising money, writing press releases, and more. By becoming an organizer, you
                become a primary driver of the chapter’s objectives. They volunteer for 2-5 hours
                per week.
              </small>
              <br />
              <b-radio-button
                v-model="applicationType"
                native-value="organizer"
                type="is-info"
                required
              >
                <span>Yes</span>
              </b-radio-button>
              <b-radio-button
                v-model="applicationType"
                native-value="chapter-member"
                type="is-info"
                required
              >
                <span>No (or not sure)</span>
              </b-radio-button>
            </b-field>
          </div>
        </div>

        <b-button
          type="is-primary"
          label="Submit"
          @click="submitForm"
          :disabled="submitting"
          class="mt-3"
        ></b-button>
      </div>
    </div>
  </adb-page>
</template>

<script lang="ts">
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage, initializeFlashMessage } from './flash_message';

export default Vue.extend({
  name: 'form-apply',
  methods: {
    validate: function() {
      type VueFormInput = Vue & { checkHtml5Validity: () => boolean };
      const refsToValidate = [
        'firstName',
        'lastName',
        'email',
        'address',
        'city',
        'zip',
        'phone',
        'birthday',
      ];
      const results = refsToValidate.map((ref) => {
        return (this.$refs[ref] as VueFormInput).checkHtml5Validity();
      });
      return results.indexOf(false) === -1;
    },
    applyButtonClicked: function() {
      this.showForm = true;
      window.scrollTo(0, 0);
    },
    notLocal: function() {
      window.location.href = '/international';
    },
    submitForm: function() {
      if (!this.validate()) return;
      if (!this.conduct) {
        flashMessage('You must agree to the code of conduct.', true);
        return;
      }
      if (!this.mission) {
        flashMessage('Please must agree with our mission & values.', true);
        return;
      }
      if (!this.applicationType) {
        flashMessage(
          'You must choose whether or not you are interested in becoming an organizer.',
          true,
        );
        return;
      }
      this.submitting = true;
      $.ajax({
        url: '/apply',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          name: this.firstName + ' ' + this.lastName,
          firstName: this.firstName,
          lastName: this.lastName,
          email: this.email,
          address: this.address,
          city: this.city,
          zip: this.zip,
          phone: this.phone,
          birthday: this.birthday,
          referral: this.referral,
          applicationType: this.applicationType,
        }),
        success: (data) => {
          this.submitting = false;
          let parsed = JSON.parse(data);
          if (parsed.status === 'error') {
            flashMessage(this.errorMessage, true);
            return;
          }
          flashMessage('Submitted!', false);
          this.submitSuccess = true;
        },
        error: () => {
          this.submitting = false;
          flashMessage(this.errorMessage, true);
        },
      });
    },
  },
  data() {
    return {
      local: false,
      submitting: false,
      submitSuccess: false,
      showForm: false,
      errorMessage: 'Sorry, there was an error submitting your form. Please try again.',
      firstName: '',
      lastName: '',
      conduct: false,
      mission: false,
      email: '',
      address: '',
      city: '',
      zip: '',
      phone: '',
      birthday: '',
      referral: '',
      applicationType: '',
    };
  },
  created() {
    document.title = 'Join DxE SF Bay';
    initializeFlashMessage();
  },
  components: {
    AdbPage,
  },
});
</script>

<style>
.fade-enter-active {
  transition: opacity 0.5s;
}
.fade-enter, .fade-leave-to /* .fade-leave-active below version 2.1.8 */ {
  opacity: 0;
}
</style>
