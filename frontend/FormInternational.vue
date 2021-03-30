<style>
.fade-enter-active {
  transition: opacity 0.5s;
}
.fade-enter, .fade-leave-to /* .fade-leave-active below version 2.1.8 */ {
  opacity: 0;
}
legend {
  font-size: 1.25em;
}
</style>

<template>
  <adb-page title="Sign up to join our International Network">
    <transition name="fade">
      <div v-if="submitSuccess">
        <h2>Thank you!</h2>
        <p>An organizer will reach out to you within the next week.</p>
      </div>
    </transition>

    <form id="form" @submit.prevent="submitForm" autocomplete="off" v-if="!submitSuccess">
      <p>
        Interested in getting involved with Direct Action Everywhere? Fill out this form and we'll
        contact you with opportunities!
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
          <div class="form-group col-sm-6">
            <label>Email address</label>
            <input
              type="email"
              class="form-control"
              v-model="email"
              name="email"
              maxlength="60"
              required
            />
          </div>

          <div class="form-group col-sm-6">
            <label>Phone number</label>
            <input
              type="text"
              class="form-control"
              v-model="phone"
              name="phone"
              maxlength="35"
              required
            />
          </div>
        </div>
      </div>

      <!-- Note that the name of this element MUST contain "search" to prevent Safari autofill from getting in the way. -->
      <div class="form-group">
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

      <br />

      <div>
        <label>What are you most interested in?</label>
        <div class="radio">
          <label
            ><input
              type="radio"
              name="interest"
              v-model="interest"
              value="participate"
              required
            />I'd like to participate</label
          >
        </div>
        <div class="radio">
          <label
            ><input type="radio" name="interest" v-model="interest" value="organize" required />I'd
            like to organize or start a chapter</label
          >
        </div>
      </div>

      <br />

      <div>
        <label
          >I am not law enforcement and my motive for expressing interest is a desire to help end
          animal exploitation. From this point forward, I commit to upholding DxE's
          <a href="http://dxe.io/values" target="_blank">values</a> and
          <a href="http://dxe.io/conduct" target="_blank">code of conduct</a> and understand that I
          may be removed if I fail to do so.</label
        >
        <div class="radio">
          <label
            ><input type="radio" name="terms" v-model="terms" value="agree" required />Yes, I agree
            with the above statement.</label
          >
        </div>
      </div>

      <br />

      <div class="form-group">
        <label>What special skills or talents do you have?</label>
        <textarea
          class="form-control"
          v-model="skills"
          name="skills"
          maxlength="500"
          placeholder="Let us know where we can get you plugged in by telling us what skills or talents you have."
        />
      </div>

      <br />

      <div class="form-group">
        <label>Is there anything specific you'd like to get involved with?</label>
        <textarea class="form-control" v-model="involvement" name="involvement" maxlength="500" />
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
  name: 'form-international',
  methods: {
    citySelected: function (loc: locationData) {
      this.city = loc.locality;
      this.state = loc.administrative_area_level_1;
      this.country = loc.country;
      this.lat = loc.latitude;
      this.lng = loc.longitude;
      this.locationChosen = true;
    },
    submitForm: function () {
      if (!this.locationChosen) {
        alert('Please choose your city from the dropbox list.');
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
  },
  components: {
    AdbPage,
    VueGoogleAutocomplete,
  },
});
</script>
