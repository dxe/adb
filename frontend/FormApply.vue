<style>
.fade-enter-active {
  transition: opacity 0.5s;
}
.fade-enter, .fade-leave-to /* .fade-leave-active below version 2.1.8 */ {
  opacity: 0;
}
</style>

<template>
  <adb-page title="Join the SF Bay Area Chapter">
    <transition name="fade">
      <div v-if="submitSuccess">
        <h2>Thank you!</h2>
        <p>
          An organizer will reach out to you within the next week and let you know what next steps
          need to be taken for your application to be approved!
        </p>
      </div>
    </transition>

    <form id="form" @submit.prevent="submitForm" autocomplete="off" v-if="!submitSuccess">
      <div v-if="local === ''">
        <legend>Do you live within 100 miles of Berkeley, CA?</legend>
        <div class="radio">
          <label><input type="radio" name="location" v-model="local" value="yes" />Yes</label>
        </div>
        <div class="radio">
          <label><input type="radio" name="location" v-model="local" value="no" />No</label>
        </div>
      </div>

      <transition name="fade">
        <div v-if="local === 'no'">
          <p>
            <a href="https://www.directactioneverywhere.com/chapters">Click here</a> to find a
            chapter local to you.
          </p>
        </div>
      </transition>

      <transition name="fade">
        <div v-if="local === 'yes'">
          <div v-if="!showForm">
            <p>
              Decades of research into historic movements – like the Civil Rights Movement or Gay
              Rights Movement – tells us that the most important factor in creating social and
              political change is sustained participation in nonviolent direct action. We also know
              that sustained participation is more likely when activists are embedded in a strong,
              supportive community. When you become a chapter member, you're committing to being an
              integral part of this activist community and to coming to one direct action event
              (like a protest, outreach event, or sanctuary workday) every month.
            </p>

            <h2>How to become a Chapter Member</h2>
            <ul>
              <li>
                Attend DxE’s
                <a href="http://dxe.io/workshop" target="_blank">intro workshop</a> (held monthly)
                to learn about our community and theory of change
              </li>
              <li>Take the Movement Power Pledge below (attend one action each month)</li>
            </ul>

            <h3>Benefits of becoming a Chapter Member</h3>
            <ul>
              <li>Voting rights (after sustained participation in the chapter)</li>
              <li>Access to chapter member mailing lists, Discord server, and Facebook group</li>
              <li>
                Added to a small community group with other chapter members in your
                neighborhood/area
              </li>
              <li>Invites to special events and opportunities</li>
              <li>
                Your photo posted at the Berkeley Animal Rights Center and on the virtual chapter
                member wall
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
                <a href="https://www.directactioneverywhere.com/core-values" target="_blank"
                  >values</a
                >
              </li>
              <li>Uphold the Movement Power Pledge (attend one action every month)</li>
            </ul>

            <br />

            <button type="button" class="btn btn-success" @click="applyButtonClicked">
              Apply now
            </button>
          </div>

          <transition name="fade">
            <div v-if="showForm">
              <h2>Take The Movement Power Pledge</h2>

              <p>
                By entering your name below, you are pledging to come to at least one action each
                month.
              </p>

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

              <br />

              <div class="form-group">
                <div class="form-check">
                  <label class="form-check">
                    <input
                      type="checkbox"
                      class="form-check-input"
                      name="conduct"
                      value="true"
                      required
                    />
                    I will uphold DxE's
                    <a
                      href="https://docs.google.com/document/d/e/2PACX-1vSTt4SXGl-eARvOpTzOM8Nzcs9uo1ZVkCuqD0doRkE6zCpx6NgMbVqYwCh691GLKZ5-hklnsaPfHlfa/pub"
                      target="_blank"
                      >code of conduct</a
                    >
                  </label>
                </div>
              </div>

              <div class="form-group">
                <div class="form-check">
                  <label class="form-check">
                    <input
                      type="checkbox"
                      class="form-check-input"
                      name="mission"
                      value="true"
                      required
                    />
                    I support DxE’s
                    <a href="https://www.directactioneverywhere.com/core-values" target="_blank"
                      >mission and values</a
                    >
                  </label>
                </div>
              </div>

              <br />

              <h2>Contact Info</h2>

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

              <div class="form-group">
                <label>Street address</label>
                <input
                  type="text"
                  class="form-control"
                  v-model="address"
                  name="address"
                  maxlength="190"
                  required
                />
              </div>

              <div class="row">
                <div class="form-group col-sm-9">
                  <label>City</label>
                  <input
                    type="text"
                    class="form-control"
                    v-model="city"
                    name="city"
                    maxlength="90"
                    required
                  />
                </div>
                <div class="form-group col-sm-3">
                  <label>Zip code</label>
                  <input
                    type="text"
                    class="form-control"
                    v-model="zip"
                    name="zip"
                    maxlength="5"
                    required
                  />
                </div>
              </div>

              <div class="row">
                <div class="form-group col-sm-7">
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
                <div class="form-group col-sm-5">
                  <label>Birthday</label>
                  <input
                    type="date"
                    class="form-control"
                    v-model="birthday"
                    name="birthday"
                    maxlength="10"
                    required
                  />
                </div>
              </div>

              <div class="form-group">
                <label
                  >Who encouraged you to apply? <small><i>(Optional)</i></small></label
                >
                <input
                  type="text"
                  class="form-control"
                  v-model="referral"
                  name="referral"
                  maxlength="100"
                />
              </div>

              <br />

              <p>
                <strong
                  >Are you interested in further leveling up your activism by becoming an
                  Organizer?</strong
                >
                <br />
                <small>
                  Organizers take ownership over achieving the chapter’s objectives and make our
                  chapter function by organizing community events, editing videos, leading protests,
                  raising money, writing press releases, and more. By becoming an organizer, you
                  become a primary driver of the chapter’s objectives. They volunteer for 2-5 hours
                  per week.
                </small>
              </p>
              <div class="radio">
                <label
                  ><input
                    type="radio"
                    v-model="applicationType"
                    name="applicationType"
                    value="organizer"
                    required
                  />Yes</label
                >
              </div>
              <div class="radio">
                <label
                  ><input
                    type="radio"
                    v-model="applicationType"
                    name="applicationType"
                    value="chapter-member"
                    required
                  />No (or not sure)</label
                >
              </div>

              <br />

              <input type="submit" class="btn btn-primary" value="Submit" :disabled="submitting" />
            </div>
          </transition>
        </div>
      </transition>
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
  name: 'form-apply',
  methods: {
    applyButtonClicked: function () {
      this.showForm = true;
      window.scrollTo(0, 0);
    },
    submitForm: function () {
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
      local: '',
      submitting: false,
      submitSuccess: false,
      showForm: false,
      errorMessage: 'Sorry, there was an error submitting your form. Please try again.',
      firstName: '',
      lastName: '',
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
  },
  components: {
    AdbPage,
  },
});
</script>
