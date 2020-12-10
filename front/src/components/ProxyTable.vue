<template>
  <el-table stripe border :data="tableData">
    <el-table-column sortable prop="id" align="center" width="70" label="ID">
    </el-table-column>
    <el-table-column sortable prop="service.type" width="100" label="Type">
      <template v-slot="scope">
        <el-button size="small">{{ scope.row.service.type }}</el-button>
      </template>
    </el-table-column>
    <el-table-column sortable prop="service.name" label="Name" />
    <el-table-column sortable prop="service.listen" label="Listen" />
    <el-table-column sortable prop="service.target" label="Target" />
    <el-table-column align="center" label="Killswitch">
      <template v-slot="scope">
        <el-switch
          v-model="scope.row.listening"
          @change="toggleListening(scope.row.id, scope.row.listening)"
          active-color="#13ce66"
          inactive-color="#ff4949"
        >
        </el-switch>
      </template>
    </el-table-column>
  </el-table>
</template>

<script>
export default {
  methods: {
    async toggleListening(id, listening) {
      console.log("set listening", id, listening);
      try {
        await this.$http.put(`/proxies/${id}/`, { listening: listening });
      } catch {
        console.error("error!");
      }
    },
  },
  created: async function() {
    try {
      const {
        data: { proxies },
      } = await this.$http.get("/proxies/");
      console.log(proxies);
      this.tableData = proxies;
    } catch {
      this.tableData = [];
    }
  },
  data() {
    return {
      tableData: [],
    };
  },
};
</script>
