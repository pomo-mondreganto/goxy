<template>
  <div id="app">
    <header class="grid-content header-color">
      <el-row>
        <a class="brand" href="#">Goxy</a>
      </el-row>
    </header>
    <el-container style="height: 500px; border: 1px solid #eee">
      <el-main>
        <proxy-table />
      </el-main>
    </el-container>
  </div>
</template>

<script>
import ProxyTable from "./components/ProxyTable.vue";

export default {
  name: "app",
  components: {
    ProxyTable,
  },
};
</script>

<style>
#app {
  font-family: "Avenir", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

body {
  margin: 0px;
}

header {
  width: 100%;
  height: 60px;
}

.header-color {
  background: #58b7ff;
}

#content {
  margin-top: 20px;
  padding-right: 40px;
}

.brand {
  color: #fff;
  background-color: transparent;
  margin-left: 20px;
  float: left;
  line-height: 25px;
  font-size: 25px;
  padding: 15px 15px;
  height: 30px;
  text-decoration: none;
}
</style>
