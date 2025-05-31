<script>
  import svelteLogo from './assets/svelte.svg'
  import viteLogo from '/vite.svg'
  import Counter from './lib/Counter.svelte'
  import Sidebar from './lib/Sidebar.svelte'

  let sidebar = false;
  function toggleSidebar(){
    sidebar = !sidebar;
  } 
</script>

<main>
  <div>
    <a href="https://svelte.dev" target="_blank" rel="noreferrer">
      <img src={svelteLogo} class="logo" alt="Svelte Logo" />
    </a>
  </div>
  <h1>Svelte</h1>
  <div class="card">
    <button class="toggle-btn" on:click={toggleSidebar}>☰</button>
    {#if sidebar }
      <Sidebar />
    {/if}
  </div>
  <div class="card">
    <Counter />
  </div>
</main>

<style>
  .logo {
    top: 0;
    right: 0;
    height: 6em;
    padding: 1.5em;
    will-change: filter;
    transition: filter 300ms;
  }
  .logo:hover {
    filter: drop-shadow(0 0 2em #ff3e00aa);
  }
</style>
