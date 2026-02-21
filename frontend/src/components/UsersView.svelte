<script>
  let {
    users = [],
    isAdmin = false,
    onUpdateUser = null,
    onDeleteUser = null,
  } = $props();

  let editingUser = $state(null);
  let showEditModal = $state(false);
  let showDeleteModal = $state(false);
  let userToDelete = $state(null);
  let editForm = $state({
    password: "",
    role: "",
  });

  function openEditModal(user) {
    editingUser = user;
    editForm = {
      password: "",
      role: user.role,
    };
    showEditModal = true;
  }

  function closeEditModal() {
    editingUser = null;
    editForm = { password: "", role: "" };
    showEditModal = false;
  }

  function openDeleteModal(user) {
    userToDelete = user;
    showDeleteModal = true;
  }

  function closeDeleteModal() {
    userToDelete = null;
    showDeleteModal = false;
  }

  function handleUpdate() {
    if (!editingUser) return;
    onUpdateUser?.(editingUser.username, editForm);
    closeEditModal();
  }

  function handleDelete() {
    if (!userToDelete) return;
    onDeleteUser?.(userToDelete.username);
    closeDeleteModal();
  }
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
          {#if isAdmin}
            <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Actions</th>
          {/if}
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-800">
        {#if users.length === 0}
          <tr>
            <td colspan={isAdmin ? 4 : 3} class="px-6 py-12 text-center text-gray-400 text-sm">
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
                    : u.role === 'engineer'
                      ? 'bg-blue-900 text-blue-200 border border-blue-700'
                      : u.role === 'operator'
                        ? 'bg-green-900 text-green-200 border border-green-700'
                        : u.role === 'automation'
                          ? 'bg-purple-900 text-purple-200 border border-purple-700'
                          : 'bg-gray-800 text-gray-200 border border-gray-700'}"
                >
                  {u.role}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-300">{u.created_at}</td>
              {#if isAdmin}
                <td class="px-4 py-3">
                  <div class="flex gap-2">
                    <button
                      onclick={() => openEditModal(u)}
                      class="px-2.5 py-1 rounded-md border border-gray-700 bg-gray-900 text-[11px] text-gray-200 hover:bg-gray-800 transition-colors"
                      title="Edit user"
                    >
                      Edit
                    </button>
                    <button
                      onclick={() => openDeleteModal(u)}
                      class="px-2.5 py-1 rounded-md border border-red-800 bg-red-900/60 text-[11px] text-red-200 hover:bg-red-900 transition-colors"
                      title="Delete user"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              {/if}
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>

<!-- Edit User Modal -->
{#if showEditModal && editingUser}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-fade-in"
    onclick={closeEditModal}
    role="dialog"
    aria-modal="true"
  >
    <div
      class="bg-gray-900 border border-gray-800 rounded-xl shadow-2xl w-full max-w-md animate-slide-in"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-800">
        <h2 class="text-xl font-bold text-gray-100">Edit User: {editingUser.username}</h2>
        <button
          onclick={closeEditModal}
          class="p-1.5 rounded-md text-gray-400 hover:text-gray-200 hover:bg-gray-800 transition-colors"
          aria-label="Close"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-6 space-y-4">
        <div class="space-y-2">
          <label for="edit-password" class="block text-sm font-medium text-gray-300">
            New Password (leave empty to keep current)
          </label>
          <input
            id="edit-password"
            type="password"
            bind:value={editForm.password}
            placeholder="Enter new password..."
            class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
        </div>

        <div class="space-y-2">
          <label for="edit-role" class="block text-sm font-medium text-gray-300">
            Role
          </label>
          <select
            id="edit-role"
            bind:value={editForm.role}
            class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
          >
            <option value="viewer">viewer</option>
            <option value="operator">operator</option>
            <option value="engineer">engineer</option>
            <option value="admin">admin</option>
            <option value="automation">automation</option>
          </select>
        </div>
      </div>

      <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-800">
        <button
          onclick={closeEditModal}
          class="px-4 py-2 rounded-md border border-gray-700 bg-gray-800 text-gray-200 text-sm font-medium hover:bg-gray-700 transition-colors"
        >
          Cancel
        </button>
        <button
          onclick={handleUpdate}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors"
        >
          Update User
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete User Modal -->
{#if showDeleteModal && userToDelete}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-fade-in"
    onclick={closeDeleteModal}
    role="dialog"
    aria-modal="true"
  >
    <div
      class="bg-gray-900 border border-gray-800 rounded-xl shadow-2xl w-full max-w-md animate-slide-in"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-800">
        <h2 class="text-xl font-bold text-gray-100">Delete User</h2>
        <button
          onclick={closeDeleteModal}
          class="p-1.5 rounded-md text-gray-400 hover:text-gray-200 hover:bg-gray-800 transition-colors"
          aria-label="Close"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-6">
        <p class="text-gray-200 mb-4">
          Are you sure you want to delete user <span class="font-semibold">{userToDelete.username}</span>?
        </p>
        <p class="text-sm text-gray-400">
          This action cannot be undone. The user will lose access to the system immediately.
        </p>
      </div>

      <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-800">
        <button
          onclick={closeDeleteModal}
          class="px-4 py-2 rounded-md border border-gray-700 bg-gray-800 text-gray-200 text-sm font-medium hover:bg-gray-700 transition-colors"
        >
          Cancel
        </button>
        <button
          onclick={handleDelete}
          class="px-4 py-2 rounded-md bg-red-600 hover:bg-red-500 text-white text-sm font-medium transition-colors"
        >
          Delete User
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateY(-20px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  .animate-fade-in {
    animation: fade-in 0.2s ease-out;
  }
  .animate-slide-in {
    animation: slide-in 0.3s ease-out;
  }
</style>
