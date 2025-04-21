<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'; // Import onUnmounted
import apiService, { type RepositoryListItem } from '../services/api';
import { useRouter } from 'vue-router';

const repositories = ref<RepositoryListItem[]>([]);
const isLoading = ref(true); // For initial load
// const isPolling = ref(false); // Removed - Not currently used
const errorMessage = ref<string | null>(null);
const router = useRouter();
let pollingIntervalId: number | null = null; // Variable to hold interval ID

async function fetchRepositories(isInitial = false) {
  if (isInitial) {
      isLoading.value = true; // Only show big loading indicator on initial load
  } else {
      // Avoid showing the main loading indicator during background polls
      // isPolling.value = true; // Optionally track polling state separately
      console.log('Polling for repository updates...');
  }
  // Don't clear previous error message during polling, only on initial load or manual action
  // errorMessage.value = null;
  try {
    repositories.value = await apiService.listRepositories();
    // Clear error only on successful fetch
    errorMessage.value = null;
  } catch (error) {
    console.error('Failed to fetch repositories:', error);
    errorMessage.value = 'Failed to load repositories. Please try again later.';
    // Error handling is also done in apiService interceptor,
    // but we can add specific messages here.
    // Keep the existing error message if polling fails, maybe?
    if (isInitial) {
        errorMessage.value = 'Failed to load repositories. Please try again later.';
    } else {
        console.warn('Polling failed, keeping previous data/error state.');
    }
  } finally {
    isLoading.value = false; // Always turn off initial loading indicator
    // isPolling.value = false; // Turn off polling indicator if used
  }
}

// Update function signature and logic to handle the { Time, Valid } structure
function formatDateTime(timeObj: { Time: string; Valid: boolean; } | null | undefined): string {
  // Check if the object is null/undefined or if Valid is false
  if (!timeObj || !timeObj.Valid) {
    return 'N/A';
  }

  const value = timeObj.Time; // Extract the actual date string

  // Ensure the extracted Time is a string
  if (typeof value !== 'string') {
      console.warn("formatDateTime received non-string Time value:", value);
      return 'Invalid Input';
  }

  try {
    // Attempt to convert the string to a Date object
    // Replace space with 'T' and handle UTC offset for better compatibility
    const isoString = value.replace(' ', 'T').replace('+00', 'Z');
    const date = new Date(isoString);

    // Check if the resulting date is valid
    if (isNaN(date.getTime())) {
        // Fallback: Try parsing without modifications if the first attempt failed
        const fallbackDate = new Date(value);
        if (isNaN(fallbackDate.getTime())) {
            console.warn("Could not parse date string:", value);
            return 'Invalid Date';
        }
        return fallbackDate.toLocaleString(); // Use fallback if valid
    }
     return date.toLocaleString(); // Use browser's locale formatting
   } catch (e) {
     console.error("Error formatting date:", value, e); // Use 'value' instead of 'pgTimestamp'
     return 'Invalid Date';
   }
 }

function getStatusClass(status: string): string {
  switch (status.toLowerCase()) {
    case 'success': return 'status-success';
    case 'failed': return 'status-failed';
    case 'syncing': return 'status-syncing';
    case 'pending': return 'status-pending';
    default: return '';
  }
}

// Removed unused viewContent function
// function viewContent(id: number) {
//   router.push({ name: 'RepoDetail', params: { id } });
// }

function editRepo(id: number) {
  router.push({ name: 'RepoEdit', params: { id } });
}

async function triggerSync(id: number) {
    if (!confirm(`Are you sure you want to trigger a manual sync for repository ID ${id}?`)) {
        return;
    }
    try {
        const result = await apiService.triggerSync(id);
        alert(result.message);
        // Optionally refresh the list after a short delay to show 'syncing' status
        setTimeout(fetchRepositories, 2000);
    } catch (error) {
        console.error(`Failed to trigger sync for repo ${id}:`, error);
        // Error alert is handled by the interceptor in api.ts
    }
}

function downloadContent(id: number) {
    const url = apiService.getDownloadUrl(id);
    // Use window.location to trigger download, works with Basic Auth if browser prompts
    window.location.href = url;
}


async function deleteRepo(id: number) {
  if (!confirm(`Are you sure you want to delete repository ID ${id}? This action cannot be undone.`)) {
    return;
  }
  try {
    await apiService.deleteRepository(id);
    alert(`Repository ${id} deleted successfully.`);
    // Refresh the list
    fetchRepositories();
  } catch (error) {
    console.error(`Failed to delete repo ${id}:`, error);
     // Error alert is handled by the interceptor in api.ts
  }
}

onMounted(() => {
  fetchRepositories(true); // Perform initial fetch

  // Start polling every 15 seconds (adjust interval as needed)
  const pollIntervalMs = 15000;
  pollingIntervalId = setInterval(() => {
      fetchRepositories(false); // Fetch updates in the background
  }, pollIntervalMs);
  console.log(`Polling started with interval ${pollIntervalMs}ms.`);
});

onUnmounted(() => {
  // Clear the interval when the component is unmounted
  if (pollingIntervalId !== null) {
    clearInterval(pollingIntervalId);
    console.log('Polling stopped.');
  }
});
</script>

<template>
  <div class="home-view">
    <h1>Monitored Repositories</h1>
    <router-link :to="{ name: 'RepoAdd' }" class="btn btn-primary add-repo-btn">
      Add New Repository
    </router-link>

    <div v-if="isLoading" class="loading">Loading repositories...</div>
    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>

    <table v-if="!isLoading && !errorMessage && repositories.length > 0" class="repo-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>URL</th>
          <th>Docs Path</th>
          <th>Extensions</th>
          <th>Status</th>
          <th>Last Sync</th>
          <th>Last Error</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="repo in repositories" :key="repo.id">
          <td>{{ repo.id }}</td>
          <td><a :href="repo.url" target="_blank" rel="noopener noreferrer">{{ repo.url }}</a></td>
          <td>{{ repo.docs_path }}</td>
          <td>{{ repo.extensions }}</td>
          <td><span :class="['status-badge', getStatusClass(repo.last_sync_status)]">{{ repo.last_sync_status }}</span></td>
          <td>{{ formatDateTime(repo.last_sync_time) }}</td>
          <td :title="repo.last_sync_error" class="error-cell">{{ repo.last_sync_error || '-' }}</td>
          <td class="actions">
            <button @click="triggerSync(repo.id)" class="btn btn-sm btn-secondary" title="Sync Now">Sync</button>
            <button @click="downloadContent(repo.id)" class="btn btn-sm btn-success" title="Download Content">Download</button>
            <button @click="editRepo(repo.id)" class="btn btn-sm btn-warning" title="Edit Config">Edit</button>
            <button @click="deleteRepo(repo.id)" class="btn btn-sm btn-danger" title="Delete Repo">Delete</button>
          </td>
        </tr>
      </tbody>
    </table>
     <div v-if="!isLoading && repositories.length === 0 && !errorMessage" class="no-repos">
        No repositories configured yet. Add one to get started!
    </div>
  </div>
</template>

<style scoped>
.home-view {
  padding: 20px;
  font-family: sans-serif;
}

.add-repo-btn {
  margin-bottom: 20px;
  display: inline-block;
}

.loading, .error-message, .no-repos {
  margin-top: 20px;
  padding: 15px;
  border-radius: 4px;
}

.loading {
  background-color: #e0e0e0;
}

.error-message {
  background-color: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.no-repos {
    background-color: #e2e3e5;
    color: #383d41;
    border: 1px solid #d6d8db;
    text-align: center;
}

.repo-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.repo-table th, .repo-table td {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
  vertical-align: top; /* Align content top */
}

.repo-table th {
  background-color: #f2f2f2;
  font-weight: bold;
}

.repo-table tbody tr:nth-child(even) {
  background-color: #f9f9f9;
}

.repo-table tbody tr:hover {
  background-color: #f1f1f1;
}

.status-badge {
  padding: 3px 8px;
  border-radius: 12px;
  color: white;
  font-size: 0.85em;
  text-transform: capitalize;
}

.status-success { background-color: #28a745; }
.status-failed { background-color: #dc3545; }
.status-syncing { background-color: #007bff; }
.status-pending { background-color: #6c757d; }

.error-cell {
    max-width: 200px; /* Limit width */
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: help; /* Indicate that it's hoverable for full text */
}

.actions {
  white-space: nowrap; /* Prevent buttons from wrapping */
}

.actions .btn {
  margin-right: 5px;
  margin-bottom: 5px; /* Add space for wrapping on small screens */
}

/* Basic Button Styling */
.btn {
  display: inline-block;
  font-weight: 400;
  text-align: center;
  vertical-align: middle;
  user-select: none;
  border: 1px solid transparent;
  padding: 0.375rem 0.75rem;
  font-size: 1rem;
  line-height: 1.5;
  border-radius: 0.25rem;
  transition: color 0.15s ease-in-out, background-color 0.15s ease-in-out, border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
  cursor: pointer;
  text-decoration: none; /* For router-link */
}

.btn-sm {
  padding: 0.25rem 0.5rem;
  font-size: 0.875rem;
  line-height: 1.5;
  border-radius: 0.2rem;
}

.btn-primary { color: #fff; background-color: #007bff; border-color: #007bff; }
.btn-primary:hover { background-color: #0056b3; border-color: #0056b3; }
.btn-secondary { color: #fff; background-color: #6c757d; border-color: #6c757d; }
.btn-secondary:hover { background-color: #5a6268; border-color: #545b62; }
.btn-info { color: #fff; background-color: #17a2b8; border-color: #17a2b8; }
.btn-info:hover { background-color: #138496; border-color: #117a8b; }
.btn-success { color: #fff; background-color: #28a745; border-color: #28a745; }
.btn-success:hover { background-color: #218838; border-color: #1e7e34; }
.btn-warning { color: #212529; background-color: #ffc107; border-color: #ffc107; }
.btn-warning:hover { background-color: #e0a800; border-color: #d39e00; }
.btn-danger { color: #fff; background-color: #dc3545; border-color: #dc3545; }
.btn-danger:hover { background-color: #c82333; border-color: #bd2130; }

</style>
