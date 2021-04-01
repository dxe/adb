<template>
  <adb-page title="Confirm your account">
    <div v-if="submitSuccess">
      <h2>Thank you!</h2>
      <p>Your account has been confirmed.</p>
      <a class="btn btn-primary" style="color: white" type="submit" href="https://discord.com/login"
        >Return to Discord</a
      >
    </div>

    <form id="form" @submit.prevent="submitForm" autocomplete="off" v-if="!submitSuccess">
      <p>
        Please fill out this form to confirm your account on DxE's Discord server.
      </p>
      <br />

      <div class="form-group">
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
      </div>

      <div class="form-group">
        <div class="row">
          <div class="form-group col-sm-12">
            <label>City</label>
            <vue-google-autocomplete
              id="map"
              name="citySearch"
              classname="form-control"
              placeholder="Enter your city & country"
              v-on:placechanged="this.citySelected"
              types="(cities)"
              :fields="['address_components', 'geometry']"
              required
            >
            </vue-google-autocomplete>
          </div>
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
//@ts-ignore
import VueGoogleAutocomplete from 'vue-google-autocomplete';
import AdbPage from './AdbPage.vue';
import { flashMessage } from './flash_message';

Vue.use(vmodal);

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
    citySelected: function(loc: locationData) {
      this.city = loc.locality;
      this.state = loc.administrative_area_level_1;
      this.country = loc.country;
      this.lat = loc.latitude;
      this.lng = loc.longitude;
      this.locationChosen = true;
    },
    submitForm: function() {
      if (!this.locationChosen) {
        alert('Please choose your city from the dropbox list.');
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
          console.log(data);
          let parsed = JSON.parse(data);
          console.log(parsed);
          if (parsed.status != 'success') {
            flashMessage(this.errorMessage, true);
            return;
          }
          flashMessage('Submitted!', false);
          this.submitSuccess = true;
          return;
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
      city: '',
      state: '',
      country: '',
      lat: 0,
      lng: 0,
    };
  },
  created() {
    document.title = 'Confirm your account';
  },
  components: {
    AdbPage,
    VueGoogleAutocomplete,
  },
});
</script>
