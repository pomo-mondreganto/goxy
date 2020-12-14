<template>
  <el-table
    border
    :style="{ width: '50%' }"
    :row-class-name="tableRowClassName"
    :data="filters"
  >
    <el-table-column sortable prop="id" align="center" width="70" label="ID" />
    <el-table-column prop="rule" label="Rule" />
    <el-table-column prop="verdict" width="100" label="Verdict" />
    <el-table-column align="center" width="100" label="Enabled">
      <template v-slot="scope">
        <el-switch
          v-model="scope.row.enabled"
          @change="
            toggleFilter(scope.row.proxy_id, scope.row.id, scope.row.enabled)
          "
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
  props: {
    filters: Array,
  },
  methods: {
    async toggleFilter(proxyId, filterId, enabled) {
      try {
        await this.$http.put(`/proxies/${proxyId}/filter_enabled/`, {
          id: filterId,
          enabled: enabled,
        });
        this.$emit("reload");
      } catch {
        console.error("error!");
      }
    },
    tableRowClassName: function({ row }) {
      if (!row.enabled) {
        return "disabled-row";
      }
      return "";
    },
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
