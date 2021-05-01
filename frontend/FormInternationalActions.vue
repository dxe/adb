<template>
  <adb-page :title="'DxE ' + chapterName + ' Chapter Update'">
    <transition name="fade">
      <div v-if="submitSuccess" class="content">
        <h2>Thank you!</h2>
        <p>Your form submission has been received.</p>
      </div>
    </transition>

    <div class="columns is-multiline" v-if="!submitSuccess">
      <div class="column is-full">
        <b-field label="What is your name?">
          <b-input
            type="text"
            v-model.trim="organizerName"
            required
            maxlength="100"
            ref="organizerName"
          ></b-input>
        </b-field>
      </div>

      <div class="column is-full">
        <b-field label="Did your chapter organize an action last month?">
          <b-radio-button
            v-model="actionLastMonth"
            :native-value="false"
            type="is-danger is-outlined"
          >
            <b-icon icon="close"></b-icon>
            <span>No</span>
          </b-radio-button>

          <b-radio-button
            v-model="actionLastMonth"
            :native-value="true"
            type="is-success is-outlined"
          >
            <b-icon icon="check"></b-icon>
            <span>Yes</span>
          </b-radio-button>
        </b-field>
      </div>

      <div class="column is-full mt-4" v-if="actionLastMonth">
        <b-field label="What date was your last action?">
          <b-datepicker v-model="lastAction" inline :focused-date="month"></b-datepicker>
        </b-field>
      </div>

      <div class="column is-full mt-5">
        <b-field
          label="How can the International Coordination team be of any assistance to you? (Optional)"
        >
          <b-input type="textarea" maxlength="500" v-model.trim="needs" />
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
import Vue from 'vue';
import AdbPage from './AdbPage.vue';
import { flashMessage, initializeFlashMessage } from './flash_message';
import moment from 'moment';

export default Vue.extend({
  name: 'form-international-actions',
  props: {
    chapterId: Number,
    chapterName: String,
    token: String,
  },
  methods: {
    validate: function() {
      type VueFormInput = Vue & { checkHtml5Validity: () => boolean };
      const refsToValidate = ['organizerName'];
      const results = refsToValidate.map((ref) => {
        return (this.$refs[ref] as VueFormInput).checkHtml5Validity();
      });
      return results.indexOf(false) === -1;
    },
    submitForm: function() {
      if (!this.validate()) return;
      if (this.actionLastMonth === null) {
        flashMessage('Please choose "Yes" or "No" for the second question.', true);
        return;
      }
      if (this.actionLastMonth && !this.lastAction) {
        flashMessage('Please select the date of your last action from the calendar.', true);
        return;
      }
      const lastAction = moment(this.lastAction).isValid()
        ? moment(this.lastAction).format('YYYY-MM-DD')
        : '';
      this.submitting = true;
      const data = JSON.stringify({
        chapterID: this.chapterId,
        token: this.token,
        organizerName: this.organizerName,
        lastAction: lastAction,
        needs: this.needs,
      });
      console.log(data);
      $.ajax({
        url: '/international_actions',
        method: 'POST',
        contentType: 'application/json',
        data: data,
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
      errorMessage: 'Sorry, there was an error submitting your form. Please try again.',
      organizerName: '',
      lastAction: null,
      needs: '',
      month: moment()
        .subtract(1, 'month')
        .startOf('month')
        .toDate(),
      actionLastMonth: null,
    };
  },
  created() {
    document.title = 'DxE Chapter Update';
    initializeFlashMessage();
  },
  components: {
    AdbPage,
  },
});
</script>
