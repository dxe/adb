<style>
.fade-enter-active {
  transition: opacity 0.5s;
}
.fade-enter, .fade-leave-to /* .fade-leave-active below version 2.1.8 */ {
  opacity: 0;
}
</style>

<template>
  <adb-page v-bind:title="formOptions.formTitle" v-bind:description="formOptions.formDescription">
    <transition name="fade">
      <div v-if="submitSuccess">
        <h2>Thank you!</h2>
        <p v-if="formOptions.formName === 'Circle Interest'">
          Thank you for applying. An organizer will reach out to you soon to help you pick a Circle.
        </p>
        <p v-else-if="formOptions.formName === 'Check-in'">
          Thank you, {{ firstName + ' ' + lastName }}.
        </p>
        <p v-else>Thank you for your submission.</p>
        <br />
        <input
          v-if="formOptions.formName === 'Check-in'"
          type="button"
          @click="reloadPage"
          class="btn"
          value="Submit another form"
          :disabled="submitting"
        />
      </div>
    </transition>

    <form id="form" @submit.prevent="submitForm" autocomplete="off" v-if="!submitSuccess">
      <div class="form-group">
        <label>Email address</label>
        <input
          type="email"
          class="form-control"
          v-model="email"
          name="email"
          maxlength="80"
          required
        />
      </div>

      <div class="row">
        <div class="form-group col-sm-6">
          <label>First Name</label>
          <input
            type="text"
            class="form-control"
            v-model="firstName"
            name="firstName"
            maxlength="35"
            required
          />
        </div>

        <div class="form-group col-sm-6">
          <label>Last Name</label>
          <input
            type="text"
            class="form-control"
            v-model="lastName"
            name="lastName"
            maxlength="35"
            required
          />
        </div>
      </div>

      <div class="row">
        <div class="form-group col-sm-8">
          <label>Phone number</label>
          <input
            type="tel"
            class="form-control"
            v-model="phone"
            name="phone"
            maxlength="20"
            required
          />
        </div>
        <div class="form-group col-sm-4">
          <label>Zip code</label>
          <input type="text" class="form-control" v-model="zip" name="zip" maxlength="5" required />
        </div>
      </div>

      <br />

      <div class="form-group" v-if="formOptions.formName !== 'Circle Interest'">
        <label>What are your activism interests, if any?</label>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="checkbox"
              class="form-check-input"
              name="activismInterests"
              value="Sanctuary"
              v-model="activismInterests"
            />
            <strong>Sanctuary Days:</strong>
            <small>Work with and spend time with rescued animals</small>
          </label>
        </div>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="checkbox"
              class="form-check-input"
              name="activismInterests"
              value="Community"
              v-model="activismInterests"
            />
            <strong>Community Events:</strong>
            <small>Make friends and create connections in the animal rights community</small>
          </label>
        </div>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="checkbox"
              class="form-check-input"
              name="activismInterests"
              value="Outreach"
              v-model="activismInterests"
            />
            <strong>Outreach:</strong> <small>Educate the public about animal cruelty</small>
          </label>
        </div>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="checkbox"
              class="form-check-input"
              name="activismInterests"
              value="Protest"
              v-model="activismInterests"
            />
            <strong>Demonstrations:</strong>
            <small
              >Challenge corporations and other institutions to make change for animals via
              non-violent protests or marches</small
            >
          </label>
        </div>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="checkbox"
              class="form-check-input"
              name="activismInterests"
              value="Trainings"
              v-model="activismInterests"
            />
            <strong>Trainings:</strong>
            <small
              >Learn how to talk to people effectively, stay legally safe as an activist, and
              organize protests</small
            >
          </label>
        </div>
      </div>

      <div class="form-group" v-if="formOptions.formName === 'Circle Interest'">
        <label>Which Circle(s) would you like to visit?</label>
        <div v-for="circle in circleOptions" :key="circle.id" class="form-check">
          <label class="form-check-label">
            <input
              type="checkbox"
              class="form-check-input"
              name="circleInterests"
              v-model="circleInterests"
              v-bind:value="circle.name"
            />
            <strong>{{ circle.name }}<span v-if="circle.description">:</span></strong>
            <small>{{ circle.description }}</small>
          </label>
        </div>
      </div>

      <br />

      <div class="form-group" v-if="formOptions.showReferralFriends">
        <label>List any existing DxE activists who you are close friends with:</label>
        <input
          type="text"
          class="form-control"
          v-model="referralFriends"
          name="referralFriends"
          maxlength="200"
        />
        <br />
      </div>

      <div class="form-group" v-if="formOptions.showReferralApply">
        <label>Who encouraged you to sign up?</label>
        <input
          type="text"
          class="form-control"
          v-model="referralApply"
          name="referralApply"
          maxlength="200"
        />
        <br />
      </div>

      <div class="form-group" v-if="formOptions.showReferralOutlet">
        <label>Where did you hear about this opportunity to get involved in DxE?</label>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="radio"
              class="form-check-input"
              name="referralOutlet"
              value="Social Media"
              v-model="referralOutlet"
            />
            Social Media
          </label>
        </div>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="radio"
              class="form-check-input"
              name="referralOutlet"
              value="Email"
              v-model="referralOutlet"
            />
            Email
          </label>
        </div>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="radio"
              class="form-check-input"
              name="referralOutlet"
              value="Meetup"
              v-model="referralOutlet"
            />
            Saturday morning Meetup
          </label>
        </div>
        <div class="form-check">
          <label class="form-check-label">
            <input
              type="radio"
              class="form-check-input"
              name="referralOutlet"
              value="In-person Invite"
              v-model="referralOutlet"
            />
            Someone invited me in person
          </label>
        </div>
      </div>

      <br />

      <input type="submit" class="btn btn-primary" value="Submit" :disabled="submitting" />
    </form>
  </adb-page>
</template>

<script lang="ts">
// Library from here: https://github.com/euvl/vue-js-modal
import vmodal from 'vue-js-modal';
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage } from './flash_message';

Vue.use(vmodal);

export default Vue.extend({
  name: 'form-interest',
  methods: {
    submitForm: function() {
      this.submitting = true;
      $.ajax({
        url: '/interest',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          form: this.formOptions.formName + " Form",
          name: this.firstName + ' ' + this.lastName,
          firstName: this.firstName,
          lastName: this.lastName,
          email: this.email,
          zip: this.zip,
          phone: this.phone,
          referralFriends: this.referralFriends,
          referralApply: this.referralApply,
          referralOutlet: this.referralOutlet,
          interests:
            this.formOptions.formName === 'Circle Interest'
              ? this.circleInterests.join(', ')
              : this.activismInterests.join(', '),
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
    reloadPage: function() {
      if (this.$options.data) {
        location.reload();
      }
    },
  },
  data() {
    return {
      formOptions: {
        formName: '',
        formTitle: '',
        formDescription: '',
        showReferralFriends: false,
        showReferralApply: false,
        showReferralOutlet: false,
      },
      submitting: false,
      submitSuccess: false,
      errorMessage: 'Sorry, there was an error submitting your form. Please try again.',
      circleOptions: [],
      firstName: '',
      lastName: '',
      email: '',
      zip: '',
      phone: '',
      referralFriends: '',
      referralApply: '',
      referralOutlet: '',
      activismInterests: [],
      circleInterests: [],
    };
  },
  created() {
    // Get form options from URL query params
    const urlParams = new URLSearchParams(window.location.search);
    this.formOptions.formName = urlParams.get('name') || 'Interest Form';
    this.formOptions.formTitle = urlParams.get('title') || 'DxE SF Bay - Get Involved';
    this.formOptions.formDescription = urlParams.get('description') || '';
    this.formOptions.showReferralFriends = urlParams.get('showReferralFriends') === 'true';
    this.formOptions.showReferralApply = urlParams.get('showReferralApply') === 'true';
    this.formOptions.showReferralOutlet = urlParams.get('showReferralOutlet') === 'true';
    const referralApply = urlParams.get('referralApply') || '';
    if (referralApply != 'null') this.referralApply = referralApply;
    document.title = this.formOptions.formTitle;
    // Get circles
    interface circle {
      id: number;
      name: string;
      description: string;
      visible: boolean;
    }
    $.ajax({
      url: '/circles',
      method: 'GET',
      success: (data) => {
        const parsed = JSON.parse(data);
        if (parsed.status === 'error') {
          flashMessage('Error: ' + parsed.message, true);
          return;
        }
        // status === "success"
        this.circleOptions = parsed.circle_groups.filter((circle: circle) => {
          return circle.visible;
        });
      },
      error: (err) => {
        console.warn(err.responseText);
        flashMessage('Server error: ' + err.responseText, true);
      },
    });
  },
  components: {
    AdbPage,
  },
});
</script>
