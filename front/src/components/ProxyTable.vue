<template>
  <el-table
    border
    :data="tableData"
    :row-key="(row) => row.id"
    :row-class-name="tableRowClassName"
    :expand-row-keys="expandRowKeys"
    @expand-change="expandChange"
  >
    <el-table-column type="expand">
      <template slot-scope="props">
        <filters-table
          :filters="props.row.filter_descriptions"
          @reload="updateProxies"
        />
      </template>
    </el-table-column>
    <el-table-column sortable prop="id" align="center" width="70" label="ID">
    </el-table-column>
    <el-table-column sortable prop="service.type" width="100" label="Type">
      <template v-slot="scope">
        <el-tag>{{ scope.row.service.type }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column sortable prop="service.name" label="Name" />
    <el-table-column sortable prop="service.listen" label="Listen" />
    <el-table-column sortable prop="service.target" label="Target" />
    <el-table-column align="center" label="Enabled">
      <template v-slot="scope">
        <el-switch
          v-model="scope.row.listening"
          active-color="#13ce66"
          inactive-color="#ff4949"
          @change="toggleListening(scope.row.id, scope.row.listening)"
        >
        </el-switch>
      </template>
    </el-table-column>
  </el-table>
</template>

<script>
import FiltersTable from "@/components/FiltersTable.vue";

export default {
  components: {
    FiltersTable,
  },
  methods: {
    async toggleListening(id, listening) {
      try {
        await this.$http.put(`/proxies/${id}/listening/`, {
          listening: listening,
        });
        await this.updateProxies();
      } catch {
        console.error("error!");
      }
    },
    tableRowClassName: function ({ row }) {
      if (!row.listening) {
        return "disabled-row";
      }
      return "";
    },
    updateProxies: async function () {
      try {
        const {
          data: { proxies },
        } = await this.$http.get("/proxies/");
        this.tableData = proxies;
      } catch {
        this.tableData = [];
      }
    },
    expandChange: function (row, expandedRows) {
      this.expandRowKeys = expandedRows.map((obj) => obj.id);
      console.log(this.expandRowKeys);
    },
  },
  created: async function () {
    await this.updateProxies();
  },
  data() {
    return {
      tableData: [],
      expandRowKeys: [],
    };
  },
};
</script>

<style>
.el-table .disabled-row {
  background: pink !important;
}

.el-table .disabled-row td {
  background: inherit !important;
}
</style>
