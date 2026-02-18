<script>
  let { users = [] } = $props();
</script>

<section class="mt-4 space-y-3">
  <div>
    <h3 class="text-sm font-semibold text-gray-100">Users</h3>
    <p class="text-[11px] text-gray-400">Manage system users and their roles</p>
  </div>

  <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm overflow-hidden">
    <table class="min-w-full text-xs">
      <thead class="bg-gray-800">
        <tr>
          <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Username</th>
          <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Role</th>
          <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Created</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-800">
        {#if users.length === 0}
          <tr>
            <td colspan="3" class="px-6 py-12 text-center text-gray-400 text-sm">
              No users found
            </td>
          </tr>
        {:else}
          {#each users as u}
            <tr class="hover:bg-gray-800/70 transition-colors">
              <td class="px-4 py-3 text-gray-100 font-medium">{u.username}</td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium {u.role === 'admin'
                    ? 'bg-red-900 text-red-200 border border-red-700'
                    : u.role === 'editor'
                      ? 'bg-blue-900 text-blue-200 border border-blue-700'
                      : 'bg-gray-800 text-gray-200 border border-gray-700'}"
                >
                  {u.role}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-300">{u.created_at}</td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>
