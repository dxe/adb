<template>
  <adb-page :title="formOptions.formTitle" :description="formOptions.formDescription">
    <transition name="fade">
      <div v-if="submitSuccess" class="content">
        <p v-if="formOptions.formName === 'Check-in'">
          Thank you, {{ firstName + ' ' + lastName }}!
        </p>
        <p v-else>Thank you for your submission!</p>
        <br />
        <b-button
          v-if="formOptions.formName === 'Check-in'"
          type="is-info"
          @click="reloadPage"
          label="Submit another form"
          :disabled="submitting"
        />
      </div>
    </transition>

    <div class="columns is-multiline" v-if="!submitSuccess">
      <div class="column is-half">
        <b-field label="First Name" label-position="on-border">
          <b-input
            type="text"
            v-model.trim="firstName"
            icon="alphabetical"
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
            icon="alphabetical"
            required
            maxlength="35"
            ref="lastName"
          ></b-input>
        </b-field>
      </div>

      <div class="column is-full">
        <b-field label="Email" label-position="on-border">
          <b-input
            type="email"
            v-model.trim="email"
            icon="email"
            required
            maxlength="80"
            ref="email"
          ></b-input>
        </b-field>
      </div>

      <div class="column is-half">
        <b-field label="Phone Number" label-position="on-border">
          <b-input
            type="text"
            v-model.trim="phone"
            icon="phone"
            required
            maxlength="20"
            ref="phone"
          ></b-input>
        </b-field>
      </div>

      <div class="column is-half">
        <b-field label="Zip Code" label-position="on-border">
          <b-input
            type="text"
            v-model.trim="zip"
            icon="city"
            required
            maxlength="5"
            ref="zip"
          ></b-input>
        </b-field>
      </div>

      <div class="column full">
        <p class="mb-3">What are your activism interests, if any?</p>

        <div class="block">
          <b-checkbox v-model="activismInterests" native-value="Sanctuary">
            <strong>Sanctuary Days:</strong>
            <small>Work with and spend time with rescued animals</small>
          </b-checkbox>
          <b-checkbox v-model="activismInterests" native-value="Community">
            <strong>Community Events:</strong>
            <small>Make friends and create connections in the animal rights community</small>
          </b-checkbox>
          <b-checkbox v-model="activismInterests" native-value="Outreach">
            <strong>Outreach:</strong>
            <small>Educate the public about animal cruelty</small>
          </b-checkbox>
          <b-checkbox v-model="activismInterests" native-value="Protest">
            <strong>Demonstrations:</strong>
            <small
              >Challenge corporations and other institutions to make change for animals via
              non-violent protests or marches</small
            >
          </b-checkbox>
          <b-checkbox v-model="activismInterests" native-value="Trainings">
            <strong>Trainings:</strong>
            <small
              >Learn how to talk to people effectively, stay legally safe as an activist, and
              organize protests</small
            >
          </b-checkbox>
        </div>
      </div>

      <div class="column is-full mt-5" v-if="formOptions.showReferralFriends">
        <b-field
          label="List any existing DxE activists who you are close friends with"
          label-position="on-border"
        >
          <b-input
            type="text"
            v-model.trim="referralFriends"
            icon="account-multiple"
            maxlength="200"
          ></b-input>
        </b-field>
      </div>

      <div class="column is-full" v-if="formOptions.showReferralApply">
        <b-field label="Who encouraged you to sign up?" label-position="on-border">
          <b-input
            type="text"
            v-model.trim="referralApply"
            icon="account"
            maxlength="200"
          ></b-input>
        </b-field>
      </div>

      <div class="column is-full" v-if="formOptions.showReferralOutlet">
        <p class="mb-3">Where did you hear about this opportunity to get involved in DxE?</p>

        <div class="block">
          <b-radio v-model="referralOutlet" native-value="Social Media">
            Social Media
          </b-radio>
          <b-radio v-model="referralOutlet" native-value="Email">
            Email
          </b-radio>
          <b-radio v-model="referralOutlet" native-value="Meetup">
            Saturday morning meetup
          </b-radio>
          <b-radio v-model="referralOutlet" native-value="In-person Invite">
            Someone invited me in person
          </b-radio>
        </div>
      </div>

      <div class="column is-full">
        <b-button
          class="my-4"
          type="is-primary"
          label="Submit"
          @click="submitForm"
          :disabled="submitting"
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
  name: 'form-interest',
  methods: {
    validate: function() {
      type VueFormInput = Vue & { checkHtml5Validity: () => boolean };
      const refsToValidate = ['firstName', 'lastName', 'email', 'phone', 'zip'];
      const results = refsToValidate.map((ref) => {
        return (this.$refs[ref] as VueFormInput).checkHtml5Validity();
      });
      return results.indexOf(false) === -1;
    },
    submitForm: function() {
      if (!this.validate()) return;
      this.submitting = true;
      $.ajax({
        url: '/interest',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          form: this.formOptions.formName + ' Form',
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
