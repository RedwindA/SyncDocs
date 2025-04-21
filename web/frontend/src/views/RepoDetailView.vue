<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import apiService, { type Repository } from '../services/api';

const route = useRoute();
const router = useRouter();
const repo = ref<Repository | null>(null);
const isLoading = ref(true);
const errorMessage = ref<string | null>(null);

// Get ID from route params
const repoId = computed(() => Number(route.params.id));

async function fetchRepositoryDetails() {
  if (isNaN(repoId.value)) {
    errorMessage.value = 'Invalid Repository ID.';
    isLoading.value = false;
    return;
  }
  isLoading.value = true;
  errorMessage.value = null;
  try {
    repo.value = await apiService.getRepository(repoId.value);
  } catch (error: any) {
    console.error(`Failed to fetch repository details for ID ${repoId.value}:`, error);
     if (error.response && error.response.status === 404) {
        errorMessage.value = `Repository with ID ${repoId.value} not found.`;
    } else {
        errorMessage.value = 'Failed to load repository details. Please try again later.';
    }
    // More specific error handling can be added based on API responses
  } finally {
    isLoading.value = false;
  }
}

async function copyToClipboard() {
  if (!repo.value?.aggregated_content) {
    alert('No content to copy.');
    return;
  }
  try {
    await navigator.clipboard.writeText(repo.value.aggregated_content);
    alert('Content copied to clipboard!');
  } catch (err) {
    console.error('Failed to copy content: ', err);
    alert('Failed to copy content to clipboard.');
  }
}

function goBack() {
  router.push({ name: 'Home' });
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
     console.error("Error formatting date:", value, e); // Use 'value'
     return 'Invalid Date';
   }
 }


onMounted(fetchRepositoryDetails);
</script>

<template>
  <div class="repo-detail-view">
    <button @click="goBack" class="btn btn-secondary back-btn">&larr; Back to List</button>

    <div v-if="isLoading" class="loading">Loading repository details...</div>
    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>

    <div v-if="repo && !isLoading && !errorMessage" class="repo-content">
      <h2>Aggregated Content for: {{ repo.url }}</h2>
      <p><strong>Docs Path:</strong> {{ repo.docs_path }}</p>
      <p><strong>Extensions:</strong> {{ repo.extensions }}</p>
      <p><strong>Last Synced:</strong> {{ formatDateTime(repo.last_sync_time) }}</p> {/* Use formatDateTime here */}

      <button @click="copyToClipboard" class="btn btn-primary copy-btn" :disabled="!repo.aggregated_content">
        Copy Content to Clipboard
      </button>

      <pre v-if="repo.aggregated_content" class="content-box">{{ repo.aggregated_content }}</pre>
      <div v-else class="no-content">
        No aggregated content available for this repository. It might need to be synced.
      </div>
    </div>
  </div>
</template>

<style scoped>
.repo-detail-view {
  padding: 20px;
  font-family: sans-serif;
}

.back-btn {
  margin-bottom: 20px;
}

.loading, .error-message, .no-content {
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

.no-content {
    background-color: #fff3cd;
    color: #856404;
    border: 1px solid #ffeeba;
}


.repo-content h2 {
  margin-bottom: 15px;
  word-break: break-all; /* Break long URLs */
}

.repo-content p {
    margin-bottom: 8px;
}

.copy-btn {
  margin-top: 10px;
  margin-bottom: 20px;
}

.content-box {
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  padding: 15px;
  border-radius: 4px;
  white-space: pre-wrap; /* Allow wrapping */
  word-wrap: break-word; /* Break long lines */
  max-height: 60vh; /* Limit height and make scrollable */
  overflow-y: auto;
  font-family: monospace; /* Use monospace for code-like content */
}

/* Basic Button Styling (copied from HomeView for consistency) */
.btn {
  display: inline-block; font-weight: 400; text-align: center; vertical-align: middle; user-select: none; border: 1px solid transparent; padding: 0.375rem 0.75rem; font-size: 1rem; line-height: 1.5; border-radius: 0.25rem; transition: color 0.15s ease-in-out, background-color 0.15s ease-in-out, border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out; cursor: pointer; text-decoration: none;
}
.btn-primary { color: #fff; background-color: #007bff; border-color: #007bff; }
.btn-primary:hover { background-color: #0056b3; border-color: #0056b3; }
.btn-primary:disabled { background-color: #007bff; border-color: #007bff; opacity: 0.65; cursor: not-allowed; }
.btn-secondary { color: #fff; background-color: #6c757d; border-color: #6c757d; }
.btn-secondary:hover { background-color: #5a6268; border-color: #545b62; }
</style>
