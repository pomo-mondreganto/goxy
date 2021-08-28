<template>
    <el-table
        border
        :style="{ width: '80%' }"
        :row-class-name="tableRowClassName"
        :data="filters"
    >
        <el-table-column
            sortable
            prop="id"
            align="center"
            width="70"
            label="ID"
        />
        <el-table-column label="Rule">
            <template v-slot="scope">
                <highlightjs language="goxy" :code="scope.row.rule" />
            </template>
        </el-table-column>
        <el-table-column width="200" label="Verdict">
            <template v-slot="scope">
                <highlightjs language="goxy" :code="scope.row.verdict" />
            </template>
        </el-table-column>
        <el-table-column align="center" width="100" label="Enabled">
            <template v-slot="scope">
                <el-switch
                    v-model="scope.row.enabled"
                    @change="updateState(scope.row)"
                    active-color="#13ce66"
                    inactive-color="#ff4949"
                >
                </el-switch>
            </template>
        </el-table-column>
        <el-table-column align="center" width="100" label="Alert">
            <template v-slot="scope">
                <el-switch
                    v-model="scope.row.alert"
                    @change="updateState(scope.row)"
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
        async updateState(filter) {
            try {
                await this.$http.put(
                    `/proxies/${filter.proxy_id}/filters/${filter.id}/`,
                    {
                        enabled: filter.enabled,
                        alert: filter.alert,
                    }
                );
                this.$emit('reload');
            } catch {
                console.error('error!');
            }
        },
        tableRowClassName: function({ row }) {
            if (!row.enabled) {
                return 'disabled-row';
            }
            return '';
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

.hljs {
    background-color: rgba(0, 0, 0, 0) !important;
}
</style>
