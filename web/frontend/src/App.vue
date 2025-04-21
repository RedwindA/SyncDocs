<script setup lang="ts">
// No specific script needed for the main layout shell usually
</script>

<template>
  <div id="app-layout">
    <header class="app-header">
      <h1>SyncDocs</h1>
      <!-- Basic Navigation (optional) -->
      <!-- <nav>
        <router-link to="/">Home</router-link>
      </nav> -->
    </header>
    <main class="app-main">
      <router-view /> <!-- This is where the routed components will be rendered -->
    </main>
    <footer class="app-footer">
      <p>&copy; {{ new Date().getFullYear() }} SyncDocs</p>
    </footer>
  </div>
</template>

<style>
/* Global styles or styles for the main layout */
#app-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

.app-header {
  background-color: #f8f9fa;
  padding: 10px 20px;
  border-bottom: 1px solid #dee2e6;
  text-align: center;
}

.app-header h1 {
    margin: 0;
    color: #007bff; /* Example color */
}

.app-main {
  flex-grow: 1; /* Allows main content to fill available space */
  padding: 0; /* Views will add their own padding */
}

.app-footer {
  background-color: #f8f9fa;
  padding: 10px 20px;
  border-top: 1px solid #dee2e6;
  text-align: center;
  font-size: 0.9em;
  color: #6c757d;
  margin-top: auto; /* Pushes footer to the bottom */
}

/* Reset default margin/padding */
body {
    margin: 0;
}

/* Import base styles if needed */
/* @import './style.css'; */

</style>
