import Vue from 'vue';
import ActivistList from './ActivistList.vue';
import CirclesList from './CirclesList.vue';
import EventEdit from './EventEdit.vue';
import EventList from './EventList.vue';
import UserList from './UserList.vue';
import WorkingGroupList from './WorkingGroupList.vue';
import FormApply from './FormApply.vue';
import FormInterest from './FormInterest.vue';
import FormInternational from './FormInternational.vue';
import FormDiscord from './FormDiscord.vue';
import ChapterList from './ChapterList.vue';
import AdbNav from './AdbNav.vue';

new Vue({
  el: '#app',
  components: {
    ActivistList,
    CirclesList,
    EventEdit,
    EventList,
    UserList,
    WorkingGroupList,
    FormApply,
    FormInterest,
    FormInternational,
    FormDiscord,
    ChapterList,
    AdbNav,
  },
});
