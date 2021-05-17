<template>
  <adb-page title="Confirm your account" narrow>
    <div v-if="submitSuccess" class="content">
      <h2>Thank you!</h2>
      <p class="mb-5">Your account has been confirmed.</p>
      <b-button
        tag="a"
        href="https://discord.com/login"
        target="_blank"
        type="is-info"
        label="Return to Discord"
        icon-left="discord"
      ></b-button>
    </div>

    <div class="columns is-multiline" v-if="!submitSuccess">
      <div class="column is-full mb-4">
        Please fill out this form to confirm your account on DxE's Discord server.
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
        <b-button
          class="my-5"
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
  name: 'form-discord',
  props: {
    id: String,
    token: String,
  },
  methods: {
    validate: function() {
      type VueFormInput = Vue & { checkHtml5Validity: () => boolean };
      const refsToValidate = ['firstName', 'lastName'];
      const results = refsToValidate.map((ref) => {
        return (this.$refs[ref] as VueFormInput).checkHtml5Validity();
      });
      return results.indexOf(false) === -1;
    },
    citySelected: function(loc: locationData) {
      this.city = loc.locality;
      this.state = loc.administrative_area_level_1;
      this.country = loc.country;
      this.lat = loc.latitude;
      this.lng = loc.longitude;
      this.locationChosen = true;
    },
    submitForm: function() {
      if (!this.validate()) return;
      if (!this.locationChosen) {
        flashMessage('Please choose your city from the dropdown list.', true);
        return;
      }
      this.submitting = true;
      $.ajax({
        url: '/discord/confirm_new',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          id: this.id,
          token: this.token,
          name: this.firstName + ' ' + this.lastName,
          firstName: this.firstName,
          lastName: this.lastName,
          email: this.email,
          city: this.city,
          state: this.state,
          country: this.country,
          lat: this.lat,
          lng: this.lng,
        }),
        success: (data) => {
          this.submitting = false;
          let parsed = JSON.parse(data);
          if (parsed.status == 'success') {
            flashMessage('Submitted!', false);
            this.submitSuccess = true;
            return;
          }
          if (parsed.status == 'invalid token') {
            flashMessage('ERROR: Your token is invalid.', true);
            return;
          }
          flashMessage('An error occurred.', true);
          return;
        },
        error: () => {
          this.submitting = false;
          flashMessage('An error occurred.', true);
        },
      });
    },
  },
  data() {
    return {
      submitting: false,
      submitSuccess: false,
      locationChosen: false,
      firstName: '',
      lastName: '',
      email: '',
      city: '',
      state: '',
      country: '',
      lat: 0,
      lng: 0,
    };
  },
  created() {
    document.title = 'Confirm your account';
    initializeFlashMessage();
  },
  components: {
    AdbPage,
    VueGoogleAutocomplete,
  },
});
</script>
