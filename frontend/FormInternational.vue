<template>
  <adb-page title="Sign up to join our International Network">
    <transition name="fade">
      <div v-if="submitSuccess" class="content">
        <h2>Thank you!</h2>
        <p>An organizer will reach out to you shortly.</p>
      </div>
    </transition>

    <div class="columns is-multiline" v-if="!submitSuccess">
      <div class="column is-full mb-4">
        Interested in getting involved with Direct Action Everywhere? Fill out this form and we'll
        contact you with opportunities!
      </div>

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
        <b-field label="Email" label-position="on-border">
          <b-input type="email" v-model.trim="email" required maxlength="60" ref="email"></b-input>
        </b-field>
      </div>

      <div class="column is-half">
        <b-field label="Phone" label-position="on-border">
          <b-input type="text" v-model.trim="phone" required maxlength="35" ref="phone"></b-input>
        </b-field>
      </div>

      <div class="column is-full">
        <b-field label="City" label-position="on-border">
          <vue-google-autocomplete
            id="map"
            name="citySearch"
            classname="input"
            placeholder="Enter your city & country"
            v-on:placechanged="this.citySelected"
            types="(cities)"
            :fields="['address_components', 'geometry']"
            required
          >
          </vue-google-autocomplete>
        </b-field>
      </div>

      <div class="column is-full">
        <b-field label="What are you most interested in?" class="mt-4">
          <b-radio v-model="interest" native-value="participate" type="is-primary" required>
            <span>I'd like to participate</span>
          </b-radio>
          <b-radio v-model="interest" native-value="organize" type="is-primary" required>
            <span>I'd like to organize or start a chapter</span>
          </b-radio>
        </b-field>
      </div>

      <div class="column is-full">
        <p class="mt-4">
          I am not law enforcement and my motive for expressing interest is a desire to help end
          animal exploitation. From this point forward, I commit to upholding DxE's
          <a href="http://dxe.io/values" target="_blank">values</a> and
          <a href="http://dxe.io/conduct" target="_blank">code of conduct</a> and understand that I
          may be removed if I fail to do so.
        </p>

        <b-field class="mt-3">
          <b-radio-button v-model="terms" native-value="agree" type="is-success" required>
            <b-icon icon="check"></b-icon>
            <span>Yes, I agree with the above statement.</span>
          </b-radio-button>
        </b-field>
      </div>

      <div class="column is-full mt-4">
        <b-field label="What special skills or talents do you have?" label-position="on-border">
          <b-input
            type="textarea"
            maxlength="500"
            v-model.trim="skills"
            placeholder="Let us know where we can get you plugged in by telling us what skills or talents you have."
          />
        </b-field>
      </div>

      <div class="column is-full">
        <b-field
          label="Is there anything specific you'd like to get involved with?"
          label-position="on-border"
        >
          <b-input type="textarea" maxlength="500" v-model.trim="involvement" />
        </b-field>
      </div>

      <div class="column is-full">
        <b-button
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
// Library from here: https://github.com/euvl/vue-js-modal
import Vue from 'vue';
//@ts-ignore
import VueGoogleAutocomplete from './external/vue-google-autocomplete';
import AdbPage from './AdbPage.vue';
import { flashMessage, initializeFlashMessage } from './flash_message';

interface locationData {
  locality: string;
  administrative_area_level_1: string;
  country: string;
  latitude: number;
  longitude: number;
}

export default Vue.extend({
  name: 'form-international',
  methods: {
    validate: function () {
      type VueFormInput = Vue & { checkHtml5Validity: () => boolean };
      const refsToValidate = ['firstName', 'lastName', 'email', 'phone'];
      const results = refsToValidate.map((ref) => {
        return (this.$refs[ref] as VueFormInput).checkHtml5Validity();
      });
      return results.indexOf(false) === -1;
    },
    citySelected: function (loc: locationData) {
      this.city = loc.locality;
      this.state = loc.administrative_area_level_1;
      this.country = loc.country;
      this.lat = loc.latitude;
      this.lng = loc.longitude;
      this.locationChosen = true;
    },
    submitForm: function () {
      if (!this.validate()) return;
      if (!this.locationChosen) {
        flashMessage('Please choose your city from the dropdown list.', true);
        return;
      }
      if (!this.interest) {
        flashMessage("Please choose whether you'd like to participate or organize.", true);
        return;
      }
      if (!this.terms) {
        flashMessage('You must agree to the terms.', true);
        return;
      }
      this.submitting = true;
      $.ajax({
        url: '/international',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          name: this.firstName + ' ' + this.lastName,
          firstName: this.firstName,
          lastName: this.lastName,
          email: this.email,
          phone: this.phone,
          interest: this.interest,
          terms: this.terms,
          skills: this.skills,
          involvement: this.involvement,
          city: this.city,
          state: this.state,
          country: this.country,
          lat: this.lat,
          lng: this.lng,
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
      submitting: false,
      submitSuccess: false,
      locationChosen: false,
      errorMessage: 'Sorry, there was an error submitting your form. Please try again.',
      firstName: '',
      lastName: '',
      email: '',
      phone: '',
      interest: '',
      terms: '',
      skills: '',
      involvement: '',
      city: '',
      state: '',
      country: '',
      lat: 0,
      lng: 0,
    };
  },
  created() {
    document.title = 'Join DxE';
    initializeFlashMessage();
  },
  components: {
    AdbPage,
    VueGoogleAutocomplete,
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
