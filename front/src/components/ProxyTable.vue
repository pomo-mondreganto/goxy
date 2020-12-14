<template>
  <el-table
    border
    :data="tableData"
    :row-key="getProxyKey"
    :row-class-name="tableRowClassName"
    :expand-row-keys="expandRowKeys"
    @expand-change="expandChange"
  >
    <el-table-column type="expand">
      <template slot-scope="props">
        <el-table
          border
          :style="{ width: '50%' }"
          :data="props.row.filter_descriptions"
        >
          <el-table-column
            sortable
            prop="id"
            align="center"
            width="70"
            label="ID"
          />
          <el-table-column prop="rule" label="Rule" />
          <el-table-column prop="verdict" width="100" label="Verdict" />
          <el-table-column align="center" width="100" label="Enabled">
            <template v-slot="scope">
              <el-switch
                v-model="scope.row.enabled"
                @change="
                  toggleFilter(
                    scope.row.proxy_id,
                    scope.row.id,
                    scope.row.enabled
                  )
                "
                active-color="#13ce66"
                inactive-color="#ff4949"
              >
              </el-switch>
            </template>
          </el-table-column>
        </el-table>
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
      try {
        await this.$http.put(`/proxies/${id}/listening/`, {
          listening: listening,
        });
        await this.updateProxies();
      } catch {
        console.error("error!");
      }
    },
    async toggleFilter(proxyId, filterId, enabled) {
      try {
        await this.$http.put(`/proxies/${proxyId}/filter_enabled/`, {
          id: filterId,
          enabled: enabled,
        });
        await this.updateProxies();
      } catch {
        console.error("error!");
      }
    },
    tableRowClassName: function ({ row }) {
      console.log("called table row", row);
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
    getProxyKey: (row) => row.id,
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
