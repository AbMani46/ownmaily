<template>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th v-for="col in columns" :key="col.key">{{ col.label }}</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="rows.length">
          <tr
            v-for="(row, i) in rows"
            :key="row.id || i"
            class="data-row"
            @click="$emit('row-click', row)"
          >
            <td
              v-for="col in columns"
              :key="col.key"
              :class="{ muted: col.muted }"
            >
              <slot :name="`cell-${col.key}`" :value="row[col.key]" :row="row">
                {{ row[col.key] ?? '—' }}
              </slot>
            </td>
          </tr>
        </template>
        <tr v-else>
          <td :colspan="columns.length" class="empty-cell">
            <slot name="empty">No results found.</slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
defineProps({
  columns: { type: Array, required: true },
  rows: { type: Array, default: () => [] },
})
defineEmits(['row-click'])
</script>

<style scoped>
.table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

thead tr {
  border-bottom: 1px solid #ece9e1;
}

th {
  text-align: left;
  padding: 9px 14px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  letter-spacing: 0.05em;
  text-transform: uppercase;
  white-space: nowrap;
}

.data-row {
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  transition: background 80ms;
}

.data-row:hover {
  background: #faf8f4;
}

td {
  padding: 11px 14px;
  color: #222;
}

td.muted {
  color: var(--text-muted);
}

.empty-cell {
  padding: 32px 14px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}
</style>
