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
  color: rgba(255, 255, 255, 0.87); /* Dark theme text color */
}

.app-header {
  background-color: #1a1a1a; /* Dark background */
  padding: 10px 20px;
  border-bottom: 1px solid #444; /* Dark theme border */
  text-align: center;
}

.app-header h1 {
    margin: 0;
    color: #646cff; /* Use link color for visibility */
}

.app-main {
  flex-grow: 1; /* Allows main content to fill available space */
  padding: 0; /* Views will add their own padding */
}

.app-footer {
  background-color: #1a1a1a; /* Dark background */
  padding: 10px 20px;
  border-top: 1px solid #444; /* Dark theme border */
  text-align: center;
  font-size: 0.9em;
  color: rgba(255, 255, 255, 0.6); /* Lighter text for dark background */
  margin-top: auto; /* Pushes footer to the bottom */
}

/* Reset default margin/padding */
body {
    margin: 0;
}

/* Import base styles if needed */
/* @import './style.css'; */

</style>
