import Vue from 'vue';
import ActivistList from './ActivistList.vue';
import CirclesList from './CirclesList.vue';
import EventEdit from './EventEdit.vue';
import EventList from './EventList.vue';
import UserList from './UserList.vue';
import WorkingGroupList from './WorkingGroupList.vue';

new Vue({
  el: '#app',
  components: {
    ActivistList, // TODO: Add view prop to template.
    CirclesList,
    EventEdit,
    EventList,
    UserList,
    WorkingGroupList,
  },
});
