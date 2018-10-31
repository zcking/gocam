<template>
  <div id="app" class="container-fluid">
    <!--<img alt="Vue logo" src="./assets/logo.png">-->
    <!--<HelloWorld msg="Welcome to Your Vue.js App"/>-->
    <DashboardComponent />
  </div>
</template>

<script>
// import HelloWorld from './components/HelloWorld.vue'
import DashboardComponent from './components/DashboardComponent'

export default {
  name: 'app',
  components: {
    // HelloWorld
    DashboardComponent
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
