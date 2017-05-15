<template>

  <div id="app" class="main">
    <table id="activist-list" class="adb-table table table-hover table-striped tablesorter">
      <thead>
        <tr>
          <th></th>
          <th>Name</th>
          <th>Email</th>
          <th>Phone</th>
          <th>First Event</th>
          <th>Last Event</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody id="activist-list-body">
        <tr v-for="activist in activists">
          <td><button class="btn btn-default glyphicon glyphicon-pencil"></button></td>
          <td>{{activist.name}}</td>
          <td>{{activist.email}}</td>
          <td>{{activist.phone}}</td>
          <td>{{activist.first_event}}</td>
          <td>{{activist.last_event}}</td>
          <td>{{activist.status}}</td>
        </tr>
      </tbody>
    </table>
  </div>

</template>

<script>
export default {
  name: 'activist-list',
  props: ['activists'],
}
</script>

<style>

  #app {
    background-color: yellow;
  }

</style>
