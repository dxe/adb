import Vue from 'vue';
import ActivistList from './ActivistList.vue';
import CirclesList from './CirclesList.vue';
import EventEdit from './EventEdit.vue';
import EventList from './EventList.vue';
import WorkingGroupList from './WorkingGroupList.vue';
import FormApply from './FormApply.vue';
import FormInterest from './FormInterest.vue';
import FormInternational from './FormInternational.vue';
import ChapterList from './ChapterList.vue';
import AdbNav from './AdbNav.vue';
import FacebookEvents from './FacebookEvents.vue';

new Vue({
  el: '#app',
  components: {
    ActivistList,
    CirclesList,
    EventEdit,
    EventList,
    WorkingGroupList,
    FormApply,
    FormInterest,
    FormInternational,
    ChapterList,
    AdbNav,
    FacebookEvents,
  },
});
