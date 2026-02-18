<script>
  export let users = [];
</script>

<h3>Users</h3>
<table style="width:100%;border-collapse:collapse;">
  <thead>
    <tr>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Username</th>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Role</th>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Created</th>
    </tr>
  </thead>
  <tbody>
    {#each users as u}
      <tr>
        <td style="border-bottom:1px solid #eee;padding:8px;">{u.username}</td>
        <td style="border-bottom:1px solid #eee;padding:8px;">{u.role}</td>
        <td style="border-bottom:1px solid #eee;padding:8px;">{u.created_at}</td>
      </tr>
    {/each}
  </tbody>
</table>

